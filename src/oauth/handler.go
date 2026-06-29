package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/nocodeleaks/quepasa/qplog"
)

var globalProvider OAuthProvider

// InitializeOAuthProvider sets up the OAuth provider at startup. Call after
// LoadOAuthConfig. If OAuth is disabled, this is a no-op.
func InitializeOAuthProvider(jwtAuth JWTAuthEncoder) {
	cfg := GetOAuthConfig()
	if cfg == nil || !cfg.Enabled {
		return
	}

	globalProvider = NewGenericOIDCProvider(cfg)
	globalJWTAuth = jwtAuth
	log.Infof("oauth: initialized with provider %s", cfg.ProviderURL)
}

// JWTAuthEncoder abstracts the JWT encoding interface used by api/form so oauth
// can issue tokens without importing jwtauth directly.
type JWTAuthEncoder interface {
	Encode(jwt.Claims) (*jwt.Token, string, error)
}

var globalJWTAuth JWTAuthEncoder

const oauthReturnURLCookieName = "oauth_return_url"

// OAuthLoginHandler initiates the OAuth authorization flow by redirecting the
// user to the external provider's login page.
//
//	GET /oauth/login
func OAuthLoginHandler(w http.ResponseWriter, r *http.Request) {
	if !IsEnabled() {
		http.Error(w, "OAuth not enabled", http.StatusNotFound)
		return
	}

	state, err := generateState()
	if err != nil {
		log.Errorf("oauth: generate state: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Generate PKCE code verifier and challenge (required by most OIDC providers).
	codeVerifier, codeChallenge, err := generatePKCE()
	if err != nil {
		log.Errorf("oauth: generate PKCE: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Store state + code_verifier in secure cookies to validate the callback.
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600, // 10 minutes
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_code_verifier",
		Value:    codeVerifier,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})
	if returnURL := getOAuthReturnURL(r); returnURL != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     oauthReturnURLCookieName,
			Value:    returnURL,
			Path:     "/",
			MaxAge:   600,
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteLaxMode,
		})
	}

	authURL := globalProvider.GetAuthURL(state, codeChallenge)
	http.Redirect(w, r, authURL, http.StatusFound)
}

// OAuthCallbackHandler handles the OAuth provider's redirect after user authentication.
// It exchanges the authorization code for an access token, fetches the user profile,
// creates or links the local account, and issues a QuePasa JWT.
//
//	GET /oauth/callback?code=...&state=...
func OAuthCallbackHandler(w http.ResponseWriter, r *http.Request) {
	if !IsEnabled() {
		http.Error(w, "OAuth not enabled", http.StatusNotFound)
		return
	}

	// Validate state (CSRF protection).
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value == "" {
		log.Warnf("oauth: callback missing state cookie")
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}
	stateQuery := r.URL.Query().Get("state")
	if stateQuery != stateCookie.Value {
		log.Warnf("oauth: state mismatch (cookie=%s query=%s)", stateCookie.Value, stateQuery)
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	// Clear state cookie.
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Retrieve code_verifier from cookie (PKCE).
	verifierCookie, err := r.Cookie("oauth_code_verifier")
	if err != nil || verifierCookie.Value == "" {
		log.Warnf("oauth: callback missing code_verifier cookie")
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}
	codeVerifier := verifierCookie.Value

	// Clear code_verifier cookie.
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_code_verifier",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	returnURL := "/"
	if returnCookie, err := r.Cookie(oauthReturnURLCookieName); err == nil && returnCookie.Value != "" {
		if normalized := normalizeOAuthReturnURL(r, returnCookie.Value); normalized != "" {
			returnURL = normalized
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:   oauthReturnURLCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	// Exchange authorization code for access token (with PKCE code_verifier).
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing code parameter", http.StatusBadRequest)
		return
	}

	accessToken, err := globalProvider.Exchange(code, codeVerifier)
	if err != nil {
		log.Errorf("oauth: token exchange failed: %v", err)
		http.Error(w, "Token exchange failed", http.StatusInternalServerError)
		return
	}

	// Fetch user info from the provider.
	userInfo, err := globalProvider.GetUserInfo(accessToken)
	if err != nil {
		log.Errorf("oauth: fetch user info failed: %v", err)
		http.Error(w, "Failed to fetch user info", http.StatusInternalServerError)
		return
	}

	// Create or link the local user account.
	user, err := FindOrCreateUser(userInfo)
	if err != nil {
		log.Errorf("oauth: link user failed: %v", err)
		http.Error(w, "Failed to create/link user", http.StatusInternalServerError)
		return
	}

	// Issue the same JWT that the form login flow uses (reuses existing session logic).
	if globalJWTAuth == nil {
		log.Errorf("oauth: JWT auth not initialized")
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	_, tokenString, err := globalJWTAuth.Encode(jwt.MapClaims{
		"user_id": user.Username,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})
	if err != nil {
		log.Errorf("oauth: encode JWT: %v", err)
		http.Error(w, "Failed to issue token", http.StatusInternalServerError)
		return
	}

	// Set JWT as an HTTP-only cookie (same pattern as form login).
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

	log.Infof("oauth: successful login for user %s", user.Username)

	// Redirect to the frontend that initiated the login. The frontend will pick up the JWT cookie.
	http.Redirect(w, r, returnURL, http.StatusFound)
}

func getOAuthReturnURL(r *http.Request) string {
	for _, key := range []string{"return_url", "redirect", "next"} {
		if value := normalizeOAuthReturnURL(r, r.URL.Query().Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func normalizeOAuthReturnURL(r *http.Request, raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	if parsed.IsAbs() {
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return ""
		}
		if sameHost(parsed.Host, r.Host) {
			return parsed.RequestURI()
		}
		if isLoopbackHost(parsed.Hostname()) && isLoopbackHost(hostnameOnly(r.Host)) {
			return parsed.String()
		}
		return ""
	}

	if parsed.Host != "" || strings.HasPrefix(raw, "//") || !strings.HasPrefix(parsed.Path, "/") {
		return ""
	}
	return parsed.RequestURI()
}

func sameHost(a, b string) bool {
	return strings.EqualFold(a, b)
}

func hostnameOnly(host string) string {
	if name, _, err := net.SplitHostPort(host); err == nil {
		return name
	}
	return host
}

func isLoopbackHost(host string) bool {
	host = strings.Trim(strings.ToLower(host), "[]")
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// generatePKCE creates a PKCE code_verifier and code_challenge (S256 method).
// Returns: (code_verifier, code_challenge, error)
func generatePKCE() (string, string, error) {
	// code_verifier: 43-128 chars of [A-Z][a-z][0-9]-._~ (RFC 7636).
	// We generate 32 random bytes → 43 base64url chars (no padding).
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	verifier := base64.RawURLEncoding.EncodeToString(b)

	// code_challenge = BASE64URL(SHA256(code_verifier))
	h := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(h[:])

	return verifier, challenge, nil
}

// BuildOAuthLoginURL returns the full URL for initiating OAuth login, useful for
// rendering a "Sign in with SSO" button in the UI.
func BuildOAuthLoginURL() string {
	if !IsEnabled() {
		return ""
	}
	return GetBaseURL() + "/oauth/login"
}

// GetOAuthStatus returns OAuth availability info for the login page configuration.
func GetOAuthStatus() map[string]interface{} {
	enabled := IsEnabled()
	status := map[string]interface{}{
		"enabled": enabled,
	}
	if enabled {
		cfg := GetOAuthConfig()
		status["login_url"] = BuildOAuthLoginURL()
		status["provider"] = cfg.ProviderURL
	}
	return status
}
