package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func registerCanonicalPublicUserRoutes(r chi.Router) {
	r.Post("/users", CanonicalUserCreateController)
}

func registerCanonicalProtectedUserRoutes(r chi.Router) {
	r.Get("/users", CanonicalUsersListController)
	r.With(withCanonicalParams(canonicalUsernameParam)).Delete("/users", CanonicalUserDeleteController)
}

func CanonicalUserCreateController(w http.ResponseWriter, r *http.Request) {
	PublicUserCreateController(w, r)
}
func CanonicalUsersListController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedUsersListController(w, r)
}
func CanonicalUserDeleteController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedUserDeleteController(w, r)
}
