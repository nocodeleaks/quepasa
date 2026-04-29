package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func registerCanonicalContactRoutes(r chi.Router) {
	r.With(withCanonicalParams(canonicalTokenParam)).Get("/contacts", CanonicalContactsListController)
	r.With(withCanonicalParams(canonicalTokenParam), requireOwnedServerToken()).Post("/contacts/search", CanonicalContactSearchController)
	r.With(withCanonicalParams(canonicalTokenParam), requireOwnedServerToken()).Post("/contacts/get", CanonicalContactGetController)
	r.With(withCanonicalParams(canonicalTokenParam), requireOwnedServerToken()).Post("/contacts/availability", CanonicalContactAvailabilityController)
}

func CanonicalContactsListController(w http.ResponseWriter, r *http.Request) {
	SPAServerContactsController(w, r)
}
func CanonicalContactSearchController(w http.ResponseWriter, r *http.Request) {
	ContactSearchController(w, r)
}
func CanonicalContactGetController(w http.ResponseWriter, r *http.Request) {
	ContactSearchController(w, r)
}
func CanonicalContactAvailabilityController(w http.ResponseWriter, r *http.Request) {
	IsOnWhatsappController(w, r)
}
