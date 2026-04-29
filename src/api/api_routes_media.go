package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func registerCanonicalMediaRoutes(r chi.Router) {
	r.With(withCanonicalParams(canonicalTokenParam, canonicalMessageIDParam), requireOwnedServerToken()).Get("/media/messages", CanonicalMediaMessageController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalMessageIDParam), requireOwnedServerToken()).Post("/media/messages/get", CanonicalMediaMessageController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalChatIDParam, canonicalPictureIDParam)).Post("/media/pictures/get", CanonicalMediaPictureInfoController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalChatIDParam, canonicalPictureIDParam)).Post("/media/pictures/info", CanonicalMediaPictureInfoController)
	r.With(withCanonicalParams(canonicalTokenParam, canonicalMessageIDParam), requireOwnedServerToken()).Post("/media/download", CanonicalMediaDownloadController)
}

func CanonicalMediaMessageController(w http.ResponseWriter, r *http.Request) {
	DownloadController(w, r)
}
func CanonicalMediaPictureInfoController(w http.ResponseWriter, r *http.Request) {
	SPAPictureInfoController(w, r)
}
func CanonicalMediaDownloadController(w http.ResponseWriter, r *http.Request) {
	SPAServerHistoryDownloadController(w, r)
}
