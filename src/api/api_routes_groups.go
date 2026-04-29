package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func registerCanonicalGroupRoutes(r chi.Router) {
	r.With(withCanonicalParams(canonicalTokenParam)).Get("/groups", CanonicalGroupsListController)
	r.With(withCanonicalParams(canonicalTokenParam)).Post("/groups", CanonicalGroupCreateController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Post("/groups/get", CanonicalGroupGetController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Post("/groups/leave", CanonicalGroupLeaveController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Patch("/groups", CanonicalGroupsPatchController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Put("/groups/participants", CanonicalGroupParticipantsController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Put("/groups/photo", CanonicalGroupPhotoController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Post("/groups/invite", CanonicalGroupInviteController)
}

func CanonicalGroupsListController(w http.ResponseWriter, r *http.Request) {
	SPAServerGroupsController(w, r)
}
func CanonicalGroupCreateController(w http.ResponseWriter, r *http.Request) {
	SPAGroupsCreateController(w, r)
}
func CanonicalGroupGetController(w http.ResponseWriter, r *http.Request) {
	SPAGroupInfoController(w, r)
}
func CanonicalGroupLeaveController(w http.ResponseWriter, r *http.Request) {
	SPAGroupLeaveController(w, r)
}
func CanonicalGroupParticipantsController(w http.ResponseWriter, r *http.Request) {
	SPAGroupParticipantsController(w, r)
}
func CanonicalGroupPhotoController(w http.ResponseWriter, r *http.Request) {
	SPAGroupPhotoController(w, r)
}
func CanonicalGroupInviteController(w http.ResponseWriter, r *http.Request) {
	SPAGroupInviteController(w, r)
}
