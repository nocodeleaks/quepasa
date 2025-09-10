package api

// WebhookQueueStats represents statistics about the webhook queue system
type WebhookQueueStats struct {
	Enabled         bool    `json:"enabled"`
	CurrentSize     int     `json:"current_size"`
	MaxSize         int     `json:"max_size"`
	Utilization     float64 `json:"utilization_percentage"`
	ProcessingDelay string  `json:"processing_delay"`
	Workers         int     `json:"workers"`
	ProcessedTotal  float64 `json:"processed_total"`
	DiscardedTotal  float64 `json:"discarded_total"`
	RetriesTotal    float64 `json:"retries_total"`
	CompletedTotal  float64 `json:"completed_total"`
	FailedTotal     float64 `json:"failed_total"`
}
