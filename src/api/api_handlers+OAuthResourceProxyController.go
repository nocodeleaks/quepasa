package api

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nocodeleaks/quepasa/oauth"
)

var oauthResourceProxyClient = &http.Client{Timeout: 20 * time.Second}

// AuthenticatedOAuthResourceProxyController forwards authenticated requests to
// the configured OAuth resource server using the provider access token captured
// during the OAuth login flow.
func AuthenticatedOAuthResourceProxyController(w http.ResponseWriter, r *http.Request) {
	targetURL, err := buildOAuthResourceProxyURL(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusServiceUnavailable)
		return
	}

	sessionToken := authenticatedOAuthProxySessionToken(r)
	accessToken, ok := oauth.AccessTokenForSession(sessionToken)
	if !ok {
		RespondErrorCode(w, fmt.Errorf("oauth access token not available; login again"), http.StatusUnauthorized)
		return
	}

	outReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadRequest)
		return
	}
	copyOAuthProxyRequestHeaders(outReq.Header, r.Header)
	outReq.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := oauthResourceProxyClient.Do(outReq)
	if err != nil {
		RespondErrorCode(w, err, http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	copyOAuthProxyResponseHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func buildOAuthResourceProxyURL(r *http.Request) (string, error) {
	cfg := oauth.GetOAuthConfig()
	if cfg == nil || strings.TrimSpace(cfg.ResourceURL) == "" {
		return "", fmt.Errorf("oauth resource base url not configured")
	}

	resourcePath := strings.TrimSpace(chi.URLParam(r, "*"))
	resourcePath = strings.TrimLeft(resourcePath, "/")
	if resourcePath == "" {
		return "", fmt.Errorf("missing resource path")
	}
	if strings.Contains(resourcePath, "://") || strings.HasPrefix(resourcePath, "//") {
		return "", fmt.Errorf("invalid resource path")
	}

	baseURL, err := url.Parse(cfg.ResourceURL)
	if err != nil || baseURL.Scheme == "" || baseURL.Host == "" {
		return "", fmt.Errorf("invalid oauth resource base url")
	}

	target := *baseURL
	target.Path = strings.TrimRight(baseURL.Path, "/") + "/" + resourcePath
	target.RawQuery = r.URL.RawQuery
	return target.String(), nil
}

func authenticatedOAuthProxySessionToken(r *http.Request) string {
	authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(authHeader) > 7 && strings.EqualFold(authHeader[:7], "Bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}

	if cookie, err := r.Cookie("jwt"); err == nil {
		return strings.TrimSpace(cookie.Value)
	}
	return ""
}

func copyOAuthProxyRequestHeaders(dst http.Header, src http.Header) {
	for _, key := range []string{"Accept", "Content-Type"} {
		if value := src.Values(key); len(value) > 0 {
			dst[key] = append([]string(nil), value...)
		}
	}
}

func copyOAuthProxyResponseHeaders(dst http.Header, src http.Header) {
	for key, values := range src {
		if isHopByHopHeader(key) {
			continue
		}
		dst[key] = append([]string(nil), values...)
	}
}

func isHopByHopHeader(key string) bool {
	switch strings.ToLower(key) {
	case "connection", "keep-alive", "proxy-authenticate", "proxy-authorization", "te", "trailer", "transfer-encoding", "upgrade":
		return true
	default:
		return false
	}
}
