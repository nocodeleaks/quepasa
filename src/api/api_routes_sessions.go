package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func registerCanonicalSessionRoutes(r chi.Router) {
	r.Get("/sessions", CanonicalSessionsListController)
	r.Post("/sessions", CanonicalSessionCreateController)
	r.Post("/sessions/search", CanonicalSessionSearchController)
	r.With(withCanonicalParams(canonicalTokenParam)).Post("/sessions/get", CanonicalSessionGetController)
	r.With(withCanonicalParams(canonicalTokenParam)).Patch("/sessions", CanonicalSessionUpdateController)
	r.With(withCanonicalParams(canonicalTokenParam)).Delete("/sessions", CanonicalSessionDeleteController)

	r.With(withCanonicalParams(canonicalTokenParam)).Get("/session/qrcode", CanonicalSessionQRCodeController)
	r.With(withCanonicalParams(canonicalTokenParam)).Get("/session/paircode", CanonicalSessionPairCodeController)
	r.With(withCanonicalParams(canonicalTokenParam)).Post("/session/enable", CanonicalSessionEnableController)
	r.With(withCanonicalParams(canonicalTokenParam)).Post("/session/disable", CanonicalSessionDisableController)
	r.With(withCanonicalParams(canonicalTokenParam)).Post("/session/debug", CanonicalSessionDebugController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalOptionParam)).Post("/session/option", CanonicalSessionOptionController)
}

func CanonicalSessionsListController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedSessionsController(w, r)
}
func CanonicalSessionCreateController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedSessionCreateController(w, r)
}
func CanonicalSessionSearchController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedSessionsSearchController(w, r)
}
func CanonicalSessionGetController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedSessionGetController(w, r)
}
func CanonicalSessionUpdateController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedSessionUpdateController(w, r)
}
func CanonicalSessionDeleteController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedSessionDeleteController(w, r)
}
func CanonicalSessionQRCodeController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedSessionQRCodeController(w, r)
}
func CanonicalSessionPairCodeController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedSessionPairCodeController(w, r)
}
func CanonicalSessionEnableController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedSessionEnableController(w, r)
}
func CanonicalSessionDisableController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedSessionDisableController(w, r)
}
func CanonicalSessionDebugController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedSessionDebugToggleController(w, r)
}
func CanonicalSessionOptionController(w http.ResponseWriter, r *http.Request) {
	AuthenticatedSessionOptionToggleController(w, r)
}
