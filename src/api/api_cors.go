package api

import (
	"net/http"

	environment "github.com/nocodeleaks/quepasa/environment"
)

// corsDefaultHeaders lists the request headers the API accepts cross-origin when
// the browser does not send an explicit Access-Control-Request-Headers preflight.
const corsDefaultHeaders = "Accept, Authorization, Content-Type, X-QUEPASA-TOKEN"

// corsAllowMethods is the method set advertised to CORS preflight requests.
const corsAllowMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"

// APICORSMiddleware applies an explicit, allow-list-based CORS policy driven by
// the CORS_ALLOWED_ORIGINS environment variable.
//
//   - empty list (default): no CORS headers are emitted and OPTIONS is left to
//     the router — the API stays same-origin only, which is the safe default for
//     multi-tenant deployments.
//   - a literal "*" entry: allow any origin, WITHOUT credentials (the browser
//     forbids credentialed wildcard responses).
//   - explicit origins: only an exact match is reflected, and credentials are
//     permitted so a trusted SPA origin can use the cookie/session flow.
//
// This replaces the permissive, commented-out "allow everything" block that used
// to live in api.go.
func APICORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origins := environment.Settings.API.AllowedOrigins

		// CORS disabled → preserve original behaviour exactly (no headers, no
		// preflight short-circuit).
		if len(origins) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		allowAll := false
		matched := false
		for _, o := range origins {
			switch o {
			case "*":
				allowAll = true
			case origin:
				matched = true
			}
		}

		if origin != "" && (matched || allowAll) {
			if matched {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Add("Vary", "Origin")
			} else {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			w.Header().Set("Access-Control-Allow-Methods", corsAllowMethods)
			reqHeaders := r.Header.Get("Access-Control-Request-Headers")
			if reqHeaders == "" {
				reqHeaders = corsDefaultHeaders
			}
			w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
			w.Header().Set("Access-Control-Max-Age", "300")
		}

		// Answer CORS preflight requests directly.
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
