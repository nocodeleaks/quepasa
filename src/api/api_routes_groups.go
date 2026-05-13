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
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Put("/groups/name", CanonicalGroupNameController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Put("/groups/description", CanonicalGroupDescriptionController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Put("/groups/participants", CanonicalGroupParticipantsController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Put("/groups/photo", CanonicalGroupPhotoController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Get("/groups/requests", CanonicalGroupRequestsController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Post("/groups/requests", CanonicalGroupRequestsController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Get("/groups/invite", CanonicalGroupInviteController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Post("/groups/invite", CanonicalGroupInviteController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalGroupIDParam)).Delete("/groups/invite", CanonicalGroupRevokeInviteController)
}

func CanonicalGroupsListController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedServerGroupsController(w, r)
}
func CanonicalGroupCreateController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedGroupCreateController(w, r)
}
func CanonicalGroupGetController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedGroupInfoController(w, r)
}
func CanonicalGroupLeaveController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedGroupLeaveController(w, r)
}
func CanonicalGroupNameController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedGroupNameController(w, r)
}
func CanonicalGroupDescriptionController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedGroupDescriptionController(w, r)
}
func CanonicalGroupParticipantsController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedGroupParticipantsController(w, r)
}
func CanonicalGroupPhotoController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedGroupPhotoController(w, r)
}
func CanonicalGroupRequestsController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedGroupRequestsController(w, r)
}
func CanonicalGroupInviteController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedGroupInviteController(w, r)
}
func CanonicalGroupRevokeInviteController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedGroupRevokeInviteController(w, r)
}
