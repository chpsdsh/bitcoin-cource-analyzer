package domain

type NewsDto struct {
	TraceID     string `json:"trace_id,omitempty"`
	Category    string `json:"category"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	SocialImage string `json:"socialimage"`
}
