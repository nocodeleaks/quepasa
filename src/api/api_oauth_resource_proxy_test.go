package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/nocodeleaks/quepasa/oauth"
)

func TestOAuthResourceProxyForwardsWithStoredAccessToken(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/Telephony/Destination/Search" {
			t.Fatalf("upstream path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("filter") != "6001" {
			t.Fatalf("filter query = %q", r.URL.Query().Get("filter"))
		}
		if r.Header.Get("Authorization") != "Bearer provider-token" {
			t.Fatalf("authorization header = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]string{{"Title": "Destino 6001"}})
	}))
	defer upstream.Close()

	t.Setenv("OAUTH_ENABLED", "true")
	t.Setenv("OAUTH_PROVIDER_URL", "https://identity.example.test")
	t.Setenv("OAUTH_CLIENT_ID", "client")
	t.Setenv("OAUTH_RESOURCE_BASE_URL", upstream.URL)
	oauth.LoadOAuthConfig()

	tokenString := encodeOAuthProxyTestToken(t)
	oauth.StoreAccessTokenForSession(tokenString, "provider-token", time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/api/v5/auth/oauth/resource/Telephony/Destination/Search?contextid=ctx&filter=6001&limit=5", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()

	newCanonicalTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected proxy status 200, got %d with body %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("expected upstream content type to be preserved, got %q", rec.Header().Get("Content-Type"))
	}
}

func TestOAuthResourceProxyRequiresConfiguredResourceBaseURL(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	t.Setenv("OAUTH_ENABLED", "true")
	t.Setenv("OAUTH_PROVIDER_URL", "https://identity.example.test")
	t.Setenv("OAUTH_CLIENT_ID", "client")
	t.Setenv("OAUTH_RESOURCE_BASE_URL", "")
	oauth.LoadOAuthConfig()

	tokenString := encodeOAuthProxyTestToken(t)
	oauth.StoreAccessTokenForSession(tokenString, "provider-token", time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/api/v5/auth/oauth/resource/Telephony/Destination/Search?filter=6001", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()

	newCanonicalTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d with body %s", rec.Code, rec.Body.String())
	}
}

func TestOAuthResourceProxyRequiresStoredAccessToken(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("upstream should not be called without stored access token")
	}))
	defer upstream.Close()

	t.Setenv("OAUTH_ENABLED", "true")
	t.Setenv("OAUTH_PROVIDER_URL", "https://identity.example.test")
	t.Setenv("OAUTH_CLIENT_ID", "client")
	t.Setenv("OAUTH_RESOURCE_BASE_URL", upstream.URL)
	oauth.LoadOAuthConfig()

	tokenString := encodeOAuthProxyTestToken(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v5/auth/oauth/resource/Telephony/Destination/Search?filter=6001", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()

	newCanonicalTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d with body %s", rec.Code, rec.Body.String())
	}
}

func TestAuthSessionReportsOAuthResourceAuthentication(t *testing.T) {
	SetupTestService(t)
	defer CleanupTestDatabase(t)
	CreateTestUser(t, "owner@example.com", "Password123!")

	t.Setenv("OAUTH_ENABLED", "true")
	t.Setenv("OAUTH_PROVIDER_URL", "https://identity.example.test")
	t.Setenv("OAUTH_CLIENT_ID", "client")
	t.Setenv("OAUTH_RESOURCE_BASE_URL", "https://resource.example.test")
	oauth.LoadOAuthConfig()

	tokenString := encodeOAuthProxyTestToken(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v5/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	newCanonicalTestRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %s", rec.Code, rec.Body.String())
	}

	var missingPayload struct {
		OAuthResourceAuthenticated bool `json:"oauthResourceAuthenticated"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &missingPayload); err != nil {
		t.Fatalf("decode missing token session response: %v", err)
	}
	if missingPayload.OAuthResourceAuthenticated {
		t.Fatalf("expected oauth resource authentication to be false before storing access token")
	}

	oauth.StoreAccessTokenForSession(tokenString, "provider-token", time.Hour)

	req = httptest.NewRequest(http.MethodGet, "/api/v5/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec = httptest.NewRecorder()
	newCanonicalTestRouter().ServeHTTP(rec, req)

	var presentPayload struct {
		OAuthResourceAuthenticated bool `json:"oauthResourceAuthenticated"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &presentPayload); err != nil {
		t.Fatalf("decode present token session response: %v", err)
	}
	if !presentPayload.OAuthResourceAuthenticated {
		t.Fatalf("expected oauth resource authentication to be true after storing access token")
	}
}

func encodeOAuthProxyTestToken(t *testing.T) string {
	t.Helper()

	_, tokenString, err := GetAuthenticatedTokenAuth().Encode(jwt.MapClaims{
		"user_id": "owner@example.com",
		"nonce":   time.Now().UnixNano(),
	})
	if err != nil {
		t.Fatalf("encode authenticated token: %v", err)
	}
	return tokenString
}
