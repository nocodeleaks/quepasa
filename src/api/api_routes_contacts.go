package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func registerCanonicalContactRoutes(r chi.Router) {
	r.With(withCanonicalParams(canonicalTokenParam)).Get("/contacts", CanonicalContactsListController)
	r.With(withCanonicalParams(canonicalTokenParam), requireOwnedServerToken()).Get("/contacts/identifier", CanonicalContactIdentifierController)
	r.With(withCanonicalParams(canonicalTokenParam), requireOwnedServerToken()).Post("/contacts/search", CanonicalContactSearchController)
	r.With(withCanonicalParams(canonicalTokenParam), requireOwnedServerToken()).Post("/contacts/get", CanonicalContactGetController)
	r.With(withCanonicalParams(canonicalTokenParam), requireOwnedServerToken()).Post("/contacts/availability", CanonicalContactAvailabilityController)
	r.With(withCanonicalParams(canonicalTokenParam), requireOwnedServerToken()).Post("/contacts/block", CanonicalContactBlockController)
	r.With(withCanonicalParams(canonicalTokenParam), requireOwnedServerToken()).Delete("/contacts/block", CanonicalContactUnblockController)
}

func CanonicalContactsListController(w http.ResponseWriter, r *http.Request) {
	SPAServerContactsController(w, r)
}
func CanonicalContactIdentifierController(w http.ResponseWriter, r *http.Request) {
	GetUserIdentifierController(w, r)
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
func CanonicalContactBlockController(w http.ResponseWriter, r *http.Request) {
	BlockContactController(w, r)
}
func CanonicalContactUnblockController(w http.ResponseWriter, r *http.Request) {
	UnblockContactController(w, r)
}
