package models

// RealtimePresenceChecker abstracts realtime connection lookup away from models.
type RealtimePresenceChecker interface {
	HasActiveConnections(token string) bool
}

// GlobalRealtimePresenceChecker can be wired at startup by transport modules.
var GlobalRealtimePresenceChecker RealtimePresenceChecker

func HasActiveRealtimeConnections(token string) bool {
	transportServicesMu.RLock()
	checker := GlobalRealtimePresenceChecker
	transportServicesMu.RUnlock()

	if checker == nil {
		return false
	}

	return checker.HasActiveConnections(token)
}

// GlobalRabbitMQClientResolver allows transport-layer wiring without importing
// rabbitmq directly in the models package.
var GlobalRabbitMQClientResolver = func(connectionString string) bool {
	return false
}

func ResolveRabbitMQClient(connectionString string) bool {
	transportServicesMu.RLock()
	resolver := GlobalRabbitMQClientResolver
	transportServicesMu.RUnlock()
	return resolver(connectionString)
}
