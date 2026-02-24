package webserver

import (
	"net/http"
	"net/url"
	"strings"
)

func MiddlewareForNormalizePaths(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r == nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		if r.URL == nil {
			r.URL = &url.URL{}
		}
		if r.URL.Path != "" {
			r.URL.Path = strings.ToLower(r.URL.Path)
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
