package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func registerCanonicalMessageRoutes(r chi.Router) {
	r.With(withCanonicalParams(canonicalTokenParam)).Get("/messages", CanonicalMessagesListController)
	r.With(withCanonicalParams(canonicalTokenParam)).Post("/messages", CanonicalMessageCreateController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalMessageIDParam), requireOwnedServerToken()).Post("/messages/get", CanonicalMessageGetController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalMessageIDParam), canonicalMethodOverride(http.MethodPut)).Patch("/messages", CanonicalMessageEditController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalMessageIDParam), requireOwnedServerToken()).Delete("/messages", CanonicalMessageDeleteController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalMessageIDParam), requireOwnedServerToken()).Post("/messages/retry", CanonicalMessageRetryController)
}

func CanonicalMessagesListController(w http.ResponseWriter, r *http.Request) {
	SPAServerMessagesController(w, r)
}
func CanonicalMessageCreateController(w http.ResponseWriter, r *http.Request) {
	SPAServerSendController(w, r)
}
func CanonicalMessageGetController(w http.ResponseWriter, r *http.Request) {
	GetMessageController(w, r)
}
func CanonicalMessageEditController(w http.ResponseWriter, r *http.Request) {
	SPAServerEditMessageController(w, r)
}
func CanonicalMessageDeleteController(w http.ResponseWriter, r *http.Request) { RevokeController(w, r) }
func CanonicalMessageRetryController(w http.ResponseWriter, r *http.Request) {
	RedispatchAPIHandler(w, r)
}
