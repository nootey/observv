package chaos

type Settings struct {
	LatencyMS int     `json:"latency_ms"`
	ErrorRate float64 `json:"error_rate"`
}

type UpdateSettingsRequest struct {
	LatencyMS *int     `json:"latency_ms"`
	ErrorRate *float64 `json:"error_rate"`
}
