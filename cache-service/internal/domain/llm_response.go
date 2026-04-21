package domain

type LLMResponse struct {
	TraceID       string `json:"trace_id,omitempty"`
	Category      string `json:"category"`
	Summarization string `json:"summarization"`
	Features      struct {
		SignalDirection   string  `json:"signal_direction"`
		SignalStrength    float64 `json:"signal_strength"`
		Uncertainty       float64 `json:"uncertainty"`
		EventUrgencyHours int     `json:"event_urgency_hours"`
		NumbersDensity    float64 `json:"numbers_density"`
		EntityDensity     float64 `json:"entity_density"`
	} `json:"features"`
}
