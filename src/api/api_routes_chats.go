package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func registerCanonicalChatRoutes(r chi.Router) {
	r.With(withCanonicalParams(canonicalTokenParam)).Post("/chats/archive", CanonicalChatArchiveController)
	r.With(withCanonicalParams(canonicalTokenParam), requireOwnedServerToken()).Post("/chats/read", CanonicalChatReadController)
	r.With(withCanonicalParams(canonicalTokenParam), requireOwnedServerToken()).Post("/chats/unread", CanonicalChatUnreadController)
	r.With(withCanonicalParams(canonicalTokenParam)).Post("/chats/presence", CanonicalChatPresenceController)
	r.With(withCanonicalParams(canonicalTokenParam), canonicalMethodOverride(http.MethodGet)).Post("/chats/labels/get", CanonicalChatLabelsGetController)
	r.With(withCanonicalParams(canonicalTokenParam)).Post("/chats/labels", CanonicalChatLabelsUpsertController)
	r.With(withCanonicalParams(canonicalTokenParam)).Delete("/chats/labels", CanonicalChatLabelsDeleteController)
}

func CanonicalChatArchiveController(w http.ResponseWriter, r *http.Request) {
	SPAServerArchiveChatController(w, r)
}
func CanonicalChatReadController(w http.ResponseWriter, r *http.Request) {
	MarkChatAsReadController(w, r)
}
func CanonicalChatUnreadController(w http.ResponseWriter, r *http.Request) {
	MarkChatAsUnreadController(w, r)
}
func CanonicalChatPresenceController(w http.ResponseWriter, r *http.Request) {
	SPAServerPresenceController(w, r)
}
func CanonicalChatLabelsGetController(w http.ResponseWriter, r *http.Request) {
	SPAServerConversationLabelController(w, r)
}
func CanonicalChatLabelsUpsertController(w http.ResponseWriter, r *http.Request) {
	SPAServerConversationLabelController(w, r)
}
func CanonicalChatLabelsDeleteController(w http.ResponseWriter, r *http.Request) {
	SPAServerConversationLabelController(w, r)
}
