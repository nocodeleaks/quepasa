package oauth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenericOIDCProviderUsesTokenSubjectWhenUserInfoOmitsSub(t *testing.T) {
	const subject = "11111111-2222-4333-8444-555555555555"
	const email = "owner@example.com"

	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			writeJSON(t, w, map[string]string{
				"issuer":                 serverURL,
				"authorization_endpoint": serverURL + "/authorize",
				"token_endpoint":         serverURL + "/token",
				"userinfo_endpoint":      serverURL + "/userinfo",
			})
		case "/token":
			writeJSON(t, w, map[string]interface{}{
				"access_token": unsignedJWT(map[string]interface{}{
					"sub": subject,
				}),
				"token_type": "Bearer",
				"expires_in": 3600,
			})
		case "/userinfo":
			writeJSON(t, w, map[string]interface{}{
				"email": email,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	serverURL = server.URL

	provider := NewGenericOIDCProvider(&OAuthConfig{
		Enabled:      true,
		ProviderURL:  server.URL,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURI:  "http://localhost/oauth/callback",
		Scopes:       []string{"openid", "email", "profile"},
	})

	accessToken, err := provider.Exchange("code", "verifier")
	if err != nil {
		t.Fatalf("exchange token: %v", err)
	}

	userInfo, err := provider.GetUserInfo(accessToken)
	if err != nil {
		t.Fatalf("get userinfo: %v", err)
	}

	if userInfo.Subject != subject {
		t.Fatalf("expected subject from token claims, got %q", userInfo.Subject)
	}
	if userInfo.Email != email {
		t.Fatalf("expected email from userinfo, got %q", userInfo.Email)
	}
	if userInfo.Claims["sub"] != subject {
		t.Fatalf("expected merged sub claim, got %#v", userInfo.Claims)
	}
}

func unsignedJWT(claims map[string]interface{}) string {
	return strings.Join([]string{
		base64.RawURLEncoding.EncodeToString(mustJSON(map[string]string{"alg": "none", "typ": "JWT"})),
		base64.RawURLEncoding.EncodeToString(mustJSON(claims)),
		"",
	}, ".")
}

func mustJSON(value interface{}) []byte {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return data
}

func writeJSON(t *testing.T, w http.ResponseWriter, value interface{}) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		t.Fatalf("write json: %v", err)
	}
}
