package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func registerCanonicalStatusRoutes(r chi.Router) {
	r.With(withCanonicalParams(canonicalTokenParam), requireOwnedServerToken()).Post("/status/publish", CanonicalStatusPublishController)
}

func CanonicalStatusPublishController(w http.ResponseWriter, r *http.Request) {
	PublishStatusController(w, r)
}
