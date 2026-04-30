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
	SPASessionsController(w, r)
}
func CanonicalSessionCreateController(w http.ResponseWriter, r *http.Request) {
	SPASessionCreateController(w, r)
}
func CanonicalSessionSearchController(w http.ResponseWriter, r *http.Request) {
	SPASessionsSearchController(w, r)
}
func CanonicalSessionGetController(w http.ResponseWriter, r *http.Request) {
	SPASessionGetController(w, r)
}
func CanonicalSessionUpdateController(w http.ResponseWriter, r *http.Request) {
	SPASessionUpdateController(w, r)
}
func CanonicalSessionDeleteController(w http.ResponseWriter, r *http.Request) {
	SPASessionDeleteController(w, r)
}
func CanonicalSessionQRCodeController(w http.ResponseWriter, r *http.Request) {
	SPASessionQRCodeController(w, r)
}
func CanonicalSessionPairCodeController(w http.ResponseWriter, r *http.Request) {
	SPASessionPairCodeController(w, r)
}
func CanonicalSessionEnableController(w http.ResponseWriter, r *http.Request) {
	SPASessionEnableController(w, r)
}
func CanonicalSessionDisableController(w http.ResponseWriter, r *http.Request) {
	SPASessionDisableController(w, r)
}
func CanonicalSessionDebugController(w http.ResponseWriter, r *http.Request) {
	SPASessionDebugToggleController(w, r)
}
func CanonicalSessionOptionController(w http.ResponseWriter, r *http.Request) {
	SPASessionOptionToggleController(w, r)
}
