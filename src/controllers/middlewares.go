package controllers

import (
	"net/http"
	"strings"
)

func MiddlewareForNormalizePaths(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "" {
			r.URL.Path = strings.ToLower(r.URL.Path)
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
