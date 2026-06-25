package api

// HealthStats represents statistics about the health of all servers
type HealthStats struct {
	Total      int     `json:"total"`
	Healthy    int     `json:"healthy"`
	Unhealthy  int     `json:"unhealthy"`
	Percentage float64 `json:"percentage"`
}
