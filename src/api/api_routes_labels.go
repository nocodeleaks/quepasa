package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func registerCanonicalLabelRoutes(r chi.Router) {
	r.Get("/labels", CanonicalLabelsController)
	r.Post("/labels", CanonicalLabelsController)
	r.With(canonicalMethodOverride(http.MethodPut)).Patch("/labels", CanonicalLabelsController)
	r.Delete("/labels", CanonicalLabelsController)
	r.Post("/labels/search", CanonicalLabelSearchController)
}

func CanonicalLabelsController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedConversationLabelController(w, r)
}
