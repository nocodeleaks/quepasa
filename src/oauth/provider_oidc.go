package oauth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// GenericOIDCProvider implements OAuthProvider for any OpenID Connect compliant
// identity provider. It performs discovery via .well-known/openid-configuration
// and follows the standard authorization code flow.
type GenericOIDCProvider struct {
	cfg         *OAuthConfig
	discovery   *oidcDiscovery
	mu          sync.RWMutex
	tokenClaims sync.Map
}

type oidcDiscovery struct {
	Issuer                string `json:"issuer"`
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserInfoEndpoint      string `json:"userinfo_endpoint"`
	fetched               time.Time
}

// NewGenericOIDCProvider creates an OIDC provider client. It performs discovery
// lazily on first use.
func NewGenericOIDCProvider(cfg *OAuthConfig) *GenericOIDCProvider {
	return &GenericOIDCProvider{cfg: cfg}
}

func (p *GenericOIDCProvider) GetAuthURL(state string, codeChallenge string) string {
	disc, err := p.getDiscovery()
	if err != nil {
		// Fallback: construct standard OIDC path if discovery fails.
		return p.cfg.ProviderURL + "/authorize?" + p.buildAuthQuery(state, codeChallenge).Encode()
	}
	return disc.AuthorizationEndpoint + "?" + p.buildAuthQuery(state, codeChallenge).Encode()
}

func (p *GenericOIDCProvider) buildAuthQuery(state string, codeChallenge string) url.Values {
	q := url.Values{}
	q.Set("client_id", p.cfg.ClientID)
	q.Set("redirect_uri", p.cfg.RedirectURI)
	q.Set("response_type", "code")
	q.Set("scope", strings.Join(p.cfg.Scopes, " "))
	q.Set("state", state)
	if codeChallenge != "" {
		q.Set("code_challenge", codeChallenge)
		q.Set("code_challenge_method", "S256")
	}
	return q
}

func (p *GenericOIDCProvider) Exchange(code string, codeVerifier string) (string, error) {
	disc, err := p.getDiscovery()
	if err != nil {
		return "", fmt.Errorf("oidc discovery: %w", err)
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", p.cfg.RedirectURI)
	data.Set("client_id", p.cfg.ClientID)
	data.Set("client_secret", p.cfg.ClientSecret)
	if codeVerifier != "" {
		data.Set("code_verifier", codeVerifier)
	}

	req, err := http.NewRequest(http.MethodPost, disc.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed: %d %s", resp.StatusCode, string(body))
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
		IDToken     string `json:"id_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access_token in response")
	}

	if claims := mergeOIDCClaims(parseJWTClaims(tokenResp.AccessToken), parseJWTClaims(tokenResp.IDToken)); len(claims) > 0 {
		p.tokenClaims.Store(tokenResp.AccessToken, claims)
	}

	return tokenResp.AccessToken, nil
}

func (p *GenericOIDCProvider) GetUserInfo(accessToken string) (*OAuthUserInfo, error) {
	disc, err := p.getDiscovery()
	if err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, disc.UserInfoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo failed: %d %s", resp.StatusCode, string(body))
	}

	var claims map[string]interface{}
	decoder := json.NewDecoder(resp.Body)
	decoder.UseNumber()
	if err := decoder.Decode(&claims); err != nil {
		return nil, err
	}

	if cachedClaims, ok := p.tokenClaims.Load(accessToken); ok {
		if tokenClaims, ok := cachedClaims.(map[string]interface{}); ok {
			claims = mergeOIDCClaims(claims, tokenClaims)
		}
	}

	email := oidcStringClaim(claims["email"])
	if email == "" {
		return nil, fmt.Errorf("userinfo missing email claim")
	}

	return &OAuthUserInfo{
		Subject:  oidcStringClaim(claims["sub"]),
		Email:    email,
		Username: oidcStringClaim(claims["preferred_username"]),
		Name:     oidcStringClaim(claims["name"]),
		Claims:   claims,
	}, nil
}

func oidcStringClaim(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case json.Number:
		return typed.String()
	}
	return ""
}

func parseJWTClaims(token string) map[string]interface{} {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return nil
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil
	}

	var claims map[string]interface{}
	decoder := json.NewDecoder(strings.NewReader(string(payload)))
	decoder.UseNumber()
	if err := decoder.Decode(&claims); err != nil {
		return nil
	}

	return claims
}

func mergeOIDCClaims(primary map[string]interface{}, fallback map[string]interface{}) map[string]interface{} {
	if len(primary) == 0 {
		return fallback
	}
	if len(fallback) == 0 {
		return primary
	}

	merged := make(map[string]interface{}, len(primary)+len(fallback))
	for key, value := range fallback {
		merged[key] = value
	}
	for key, value := range primary {
		merged[key] = value
	}
	return merged
}

func (p *GenericOIDCProvider) getDiscovery() (*oidcDiscovery, error) {
	p.mu.RLock()
	if p.discovery != nil && time.Since(p.discovery.fetched) < 1*time.Hour {
		defer p.mu.RUnlock()
		return p.discovery, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock.
	if p.discovery != nil && time.Since(p.discovery.fetched) < 1*time.Hour {
		return p.discovery, nil
	}

	discoveryURL := strings.TrimRight(p.cfg.ProviderURL, "/") + "/.well-known/openid-configuration"
	resp, err := http.Get(discoveryURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery request failed: %d", resp.StatusCode)
	}

	var disc oidcDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&disc); err != nil {
		return nil, err
	}

	if disc.AuthorizationEndpoint == "" || disc.TokenEndpoint == "" || disc.UserInfoEndpoint == "" {
		return nil, fmt.Errorf("incomplete discovery document")
	}

	disc.fetched = time.Now()
	p.discovery = &disc
	return p.discovery, nil
}
