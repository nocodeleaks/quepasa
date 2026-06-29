package oauth

import (
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
	cfg       *OAuthConfig
	discovery *oidcDiscovery
	mu        sync.RWMutex
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
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", err
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("empty access_token in response")
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

	var claims struct {
		Sub               string `json:"sub"`
		Email             string `json:"email"`
		PreferredUsername string `json:"preferred_username"`
		Name              string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
		return nil, err
	}

	if claims.Email == "" {
		return nil, fmt.Errorf("userinfo missing email claim")
	}

	return &OAuthUserInfo{
		Email:    claims.Email,
		Username: claims.PreferredUsername,
		Name:     claims.Name,
	}, nil
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
