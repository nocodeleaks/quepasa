package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	events "github.com/nocodeleaks/quepasa/events"
)

// APIEventMiddleware emits one non-blocking internal event for each HTTP request.
func APIEventMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		wrapped := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		next.ServeHTTP(wrapped, r)

		events.Publish(events.Event{
			Name:     "api.request",
			Source:   "api.middleware",
			Status:   strconv.Itoa(wrapped.Status()),
			Duration: time.Since(startedAt),
			Attributes: map[string]string{
				"method": r.Method,
				"route":  resolveRoutePattern(r),
			},
		})
	})
}

func resolveRoutePattern(r *http.Request) string {
	if r == nil {
		return "unknown"
	}

	routeContext := chi.RouteContext(r.Context())
	if routeContext == nil {
		return r.URL.Path
	}

	pattern := routeContext.RoutePattern()
	if pattern == "" {
		return r.URL.Path
	}

	return pattern
}
