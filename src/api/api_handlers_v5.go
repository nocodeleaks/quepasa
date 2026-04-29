package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	v5 "github.com/nocodeleaks/quepasa/api/v5"
	events "github.com/nocodeleaks/quepasa/events"
	models "github.com/nocodeleaks/quepasa/models"
)

// CurrentCanonicalAPIVersion identifies the latest family-based canonical API.
const CurrentCanonicalAPIVersion = "v5"

type canonicalParamSpec struct {
	Name       string
	BodyKeys   []string
	QueryKeys  []string
	HeaderKeys []string
}

var (
	canonicalTokenParam = canonicalParamSpec{
		Name:       "token",
		BodyKeys:   []string{"token", "sessionToken"},
		QueryKeys:  []string{"token", "sessionToken"},
		HeaderKeys: []string{"X-QUEPASA-TOKEN"},
	}
	canonicalMessageIDParam = canonicalParamSpec{
		Name:       "messageid",
		BodyKeys:   []string{"messageid", "messageId", "id"},
		QueryKeys:  []string{"messageid", "id"},
		HeaderKeys: []string{"X-QUEPASA-MESSAGEID"},
	}
	canonicalGroupIDParam = canonicalParamSpec{
		Name:       "groupid",
		BodyKeys:   []string{"groupid", "groupId", "group_jid"},
		QueryKeys:  []string{"groupid", "groupId", "group_jid"},
		HeaderKeys: []string{"X-QUEPASA-GROUPID"},
	}
	canonicalChatIDParam = canonicalParamSpec{
		Name:       "chatid",
		BodyKeys:   []string{"chatid", "chatId"},
		QueryKeys:  []string{"chatid", "chatId"},
		HeaderKeys: []string{"X-QUEPASA-CHATID"},
	}
	canonicalPictureIDParam = canonicalParamSpec{
		Name:       "pictureid",
		BodyKeys:   []string{"pictureid", "pictureId"},
		QueryKeys:  []string{"pictureid", "pictureId"},
		HeaderKeys: []string{"X-QUEPASA-PICTUREID"},
	}
	canonicalOptionParam = canonicalParamSpec{
		Name:       "option",
		BodyKeys:   []string{"option"},
		QueryKeys:  []string{"option"},
		HeaderKeys: []string{"X-QUEPASA-OPTION"},
	}
	canonicalUsernameParam = canonicalParamSpec{
		Name:       "username",
		BodyKeys:   []string{"username"},
		QueryKeys:  []string{"username"},
		HeaderKeys: []string{"X-QUEPASA-USER"},
	}
)

// RegisterAPIV5Controllers mounts the canonical family-based API under both /api and /api/v5.
// Legacy routes remain registered separately for compatibility.
func RegisterAPIV5Controllers(r chi.Router) {
	v5.RegisterControllers(r, CurrentCanonicalAPIVersion, v5.Groups{
		Public: registerCanonicalPublicRoutes,
		Protected: func(protected chi.Router) {
			tokenAuth := GetSPATokenAuth()
			protected.Use(jwtauth.Verifier(tokenAuth))
			protected.Use(SPAAuthenticatorHandler)
			registerCanonicalProtectedRoutes(protected)
		},
	})
}

func registerCanonicalPublicRoutes(r chi.Router) {
	registerCanonicalSystemRoutes(r)
	registerCanonicalPublicAuthRoutes(r)
	registerCanonicalPublicUserRoutes(r)
}

func registerCanonicalProtectedRoutes(r chi.Router) {
	registerCanonicalProtectedAuthRoutes(r)
	registerCanonicalProtectedUserRoutes(r)
	registerCanonicalSessionRoutes(r)
	registerCanonicalDispatchRoutes(r)
	registerCanonicalContactRoutes(r)
	registerCanonicalMessageRoutes(r)
	registerCanonicalChatRoutes(r)
	registerCanonicalGroupRoutes(r)
	registerCanonicalMediaRoutes(r)
	registerCanonicalLabelRoutes(r)
}

// VersionController exposes the current backend version in the canonical system family.
func VersionController(w http.ResponseWriter, r *http.Request) {
	RespondSuccess(w, map[string]interface{}{
		"version": models.QpVersion,
	})
}

// SPAMasterKeyStatusController exposes only master key status, never the secret value itself.
func SPAMasterKeyStatusController(w http.ResponseWriter, r *http.Request) {
	if _, err := GetSPAUser(r); err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	RespondSuccess(w, buildMasterKeyStatusResponse(strings.TrimSpace(models.ENV.MasterKey())))
}

// CanonicalDispatchesController keeps one family entry point while still delegating
// to the current transport-specific controllers.
func CanonicalDispatchesController(w http.ResponseWriter, r *http.Request) {
	dispatchType := strings.ToLower(strings.TrimSpace(resolveCanonicalDispatchType(r)))
	switch dispatchType {
	case "webhook", "webhooks":
		SPAWebHooksController(w, r)
	case "rabbitmq":
		SPARabbitMQController(w, r)
	default:
		RespondErrorCode(w, fmt.Errorf("dispatch type is required and must be webhook or rabbitmq"), http.StatusBadRequest)
	}
}

// CanonicalGroupsPatchController normalizes the canonical PATCH contract to the
// existing group name/topic handlers without duplicating business logic.
func CanonicalGroupsPatchController(w http.ResponseWriter, r *http.Request) {
	body, payload := readCanonicalRequestBody(r)
	r.Body = io.NopCloser(bytes.NewReader(body))

	if value := lookupBodyString(payload, "description"); value != "" && lookupBodyString(payload, "topic") == "" {
		payload["topic"] = mustMarshalCanonicalValue(value)
		body = rebuildCanonicalBody(payload)
		r.Body = io.NopCloser(bytes.NewReader(body))
	}

	if lookupBodyString(payload, "name") != "" {
		SPAGroupNameController(w, r)
		return
	}

	if lookupBodyString(payload, "topic") != "" {
		SPAGroupDescriptionController(w, r)
		return
	}

	RespondErrorCode(w, fmt.Errorf("groups patch requires name or topic/description"), http.StatusBadRequest)
}

// CanonicalLabelSearchController exposes a body-based search contract for labels.
func CanonicalLabelSearchController(w http.ResponseWriter, r *http.Request) {
	user, err := GetSPAUser(r)
	if err != nil {
		RespondErrorCode(w, err, http.StatusUnauthorized)
		return
	}

	store, err := getConversationLabelStoreOrError()
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	var request struct {
		Query        string   `json:"query"`
		Active       *bool    `json:"active"`
		Token        string   `json:"token"`
		ChatIDs      []string `json:"chatIds"`
		IncludeChats bool     `json:"includeChats"`
	}
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil && err.Error() != "EOF" {
			RespondErrorCode(w, err, http.StatusBadRequest)
			return
		}
	}

	labels, err := store.FindAllForUser(strings.TrimSpace(user.Username), request.Active)
	if err != nil {
		RespondErrorCode(w, err, http.StatusInternalServerError)
		return
	}

	query := strings.ToLower(strings.TrimSpace(request.Query))
	filtered := make([]*models.QpConversationLabel, 0, len(labels))
	for _, label := range labels {
		if label == nil {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(label.Name), query) && !strings.Contains(strings.ToLower(label.Color), query) {
			continue
		}
		filtered = append(filtered, label)
	}

	response := map[string]interface{}{
		"labels": filtered,
		"total":  len(filtered),
		"query":  request.Query,
	}

	if request.IncludeChats && strings.TrimSpace(request.Token) != "" && len(request.ChatIDs) > 0 {
		if _, err := GetSPAOwnedServerRecord(user, request.Token); err == nil {
			if loaded, err := store.FindConversationLabelsMap(strings.TrimSpace(request.Token), strings.TrimSpace(user.Username), request.ChatIDs); err == nil {
				response["assignedChats"] = invertConversationLabelMap(loaded)
			}
		}
	}

	RespondSuccess(w, response)
}

func invertConversationLabelMap(source map[string][]*models.QpConversationLabel) map[string][]string {
	response := map[string][]string{}
	for chatID, labels := range source {
		for _, label := range labels {
			if label == nil {
				continue
			}
			key := strconv.FormatInt(label.ID, 10)
			response[key] = append(response[key], chatID)
		}
	}
	return response
}

func requireOwnedServerToken() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetSPAUser(r)
			if err != nil {
				RespondErrorCode(w, err, http.StatusUnauthorized)
				return
			}

			token := strings.TrimSpace(GetToken(r))
			if token == "" {
				RespondErrorCode(w, fmt.Errorf("missing token parameter"), http.StatusBadRequest)
				return
			}

			if _, err := GetSPAOwnedServerRecord(user, token); err != nil {
				respondSPAServerLookupError(w, err)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func withCanonicalParams(specs ...canonicalParamSpec) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, payload := readCanonicalRequestBody(r)
			r.Body = io.NopCloser(bytes.NewReader(body))

			request := r
			for _, spec := range specs {
				if strings.TrimSpace(chi.URLParam(request, spec.Name)) != "" {
					continue
				}

				value := lookupCanonicalParam(request, payload, spec)
				if value == "" {
					continue
				}

				request = setCanonicalURLParam(request, spec.Name, value)
			}

			next.ServeHTTP(w, request)
		})
	}
}

func canonicalMethodOverride(method string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cloned := r.Clone(r.Context())
			cloned.Method = method
			next.ServeHTTP(w, cloned)
		})
	}
}

func setCanonicalURLParam(r *http.Request, name string, value string) *http.Request {
	if r == nil || strings.TrimSpace(value) == "" {
		return r
	}

	ctx := chi.RouteContext(r.Context())
	if ctx == nil {
		ctx = chi.NewRouteContext()
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))
	}

	ctx.URLParams.Add(name, value)
	return r
}

func readCanonicalRequestBody(r *http.Request) ([]byte, map[string]json.RawMessage) {
	if r == nil || r.Body == nil {
		return nil, nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		return body, nil
	}

	payload := map[string]json.RawMessage{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return body, nil
	}

	return body, payload
}

func lookupCanonicalParam(r *http.Request, payload map[string]json.RawMessage, spec canonicalParamSpec) string {
	for _, key := range spec.HeaderKeys {
		if value := strings.TrimSpace(r.Header.Get(key)); value != "" {
			return value
		}
	}

	for _, key := range spec.QueryKeys {
		if value := strings.TrimSpace(r.URL.Query().Get(key)); value != "" {
			return value
		}
	}

	for _, key := range spec.BodyKeys {
		if value := lookupBodyString(payload, key); value != "" {
			return value
		}
	}

	return ""
}

func lookupBodyString(payload map[string]json.RawMessage, key string) string {
	if payload == nil {
		return ""
	}

	raw, ok := payload[key]
	if !ok {
		return ""
	}

	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		return strings.TrimSpace(asString)
	}

	var asNumber json.Number
	if err := json.Unmarshal(raw, &asNumber); err == nil {
		return strings.TrimSpace(asNumber.String())
	}

	var asBool bool
	if err := json.Unmarshal(raw, &asBool); err == nil {
		if asBool {
			return "true"
		}
		return "false"
	}

	return ""
}

func mustMarshalCanonicalValue(value string) json.RawMessage {
	encoded, _ := json.Marshal(value)
	return encoded
}

func rebuildCanonicalBody(payload map[string]json.RawMessage) []byte {
	if payload == nil {
		return nil
	}
	body, _ := json.Marshal(payload)
	return body
}

func resolveCanonicalDispatchType(r *http.Request) string {
	body, payload := readCanonicalRequestBody(r)
	r.Body = io.NopCloser(bytes.NewReader(body))

	for _, key := range []string{"type", "dispatchType"} {
		if value := lookupBodyString(payload, key); value != "" {
			return value
		}
		if value := strings.TrimSpace(r.URL.Query().Get(key)); value != "" {
			return value
		}
	}

	return strings.TrimSpace(r.Header.Get("X-QUEPASA-DISPATCH-TYPE"))
}

func init() {
	// Emit one event when the canonical API surface is initialized so observers can
	// distinguish legacy-only requests from canonical family-route requests over time.
	events.Publish(events.Event{
		Name:   "api.v5.routes.initialized",
		Source: "api.v5",
		Status: "ready",
	})
}
