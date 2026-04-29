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
	SPAServersController(w, r)
}
func CanonicalSessionCreateController(w http.ResponseWriter, r *http.Request) {
	SPAServerCreateController(w, r)
}
func CanonicalSessionSearchController(w http.ResponseWriter, r *http.Request) {
	SPAServersSearchController(w, r)
}
func CanonicalSessionGetController(w http.ResponseWriter, r *http.Request) {
	SPAServerInfoController(w, r)
}
func CanonicalSessionUpdateController(w http.ResponseWriter, r *http.Request) {
	SPAServerUpdateController(w, r)
}
func CanonicalSessionDeleteController(w http.ResponseWriter, r *http.Request) {
	SPAServerDeleteController(w, r)
}
func CanonicalSessionQRCodeController(w http.ResponseWriter, r *http.Request) {
	SPAServerQRCodeController(w, r)
}
func CanonicalSessionPairCodeController(w http.ResponseWriter, r *http.Request) {
	SPAServerPairCodeController(w, r)
}
func CanonicalSessionEnableController(w http.ResponseWriter, r *http.Request) {
	SPAServerEnableController(w, r)
}
func CanonicalSessionDisableController(w http.ResponseWriter, r *http.Request) {
	SPAServerDisableController(w, r)
}
func CanonicalSessionDebugController(w http.ResponseWriter, r *http.Request) {
	SPAServerDebugToggleController(w, r)
}
func CanonicalSessionOptionController(w http.ResponseWriter, r *http.Request) {
	SPAServerOptionToggleController(w, r)
}
