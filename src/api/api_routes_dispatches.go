package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func registerCanonicalDispatchRoutes(r chi.Router) {
	r.With(withCanonicalParams(canonicalTokenParam)).Get("/dispatches", CanonicalDispatchesController)
	r.With(withCanonicalParams(canonicalTokenParam)).Post("/dispatches", CanonicalDispatchesController)
	r.With(withCanonicalParams(canonicalTokenParam)).Delete("/dispatches", CanonicalDispatchesController)
	r.With(withCanonicalParams(canonicalTokenParam)).Get("/dispatches/webhooks", CanonicalDispatchWebhooksController)
	r.With(withCanonicalParams(canonicalTokenParam)).Post("/dispatches/webhooks", CanonicalDispatchWebhooksController)
	r.With(withCanonicalParams(canonicalTokenParam)).Delete("/dispatches/webhooks", CanonicalDispatchWebhooksController)
	r.With(withCanonicalParams(canonicalTokenParam)).Get("/dispatches/rabbitmq", CanonicalDispatchRabbitMQController)
	r.With(withCanonicalParams(canonicalTokenParam)).Post("/dispatches/rabbitmq", CanonicalDispatchRabbitMQController)
	r.With(withCanonicalParams(canonicalTokenParam)).Delete("/dispatches/rabbitmq", CanonicalDispatchRabbitMQController)
}

func CanonicalDispatchWebhooksController(w http.ResponseWriter, r *http.Request) {
	SPAWebHooksController(w, r)
}
func CanonicalDispatchRabbitMQController(w http.ResponseWriter, r *http.Request) {
	SPARabbitMQController(w, r)
}
