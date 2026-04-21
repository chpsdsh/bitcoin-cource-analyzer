package domain

type Articles struct {
	Articles []GdeltAPIDto `json:"articles"`
}
type GdeltAPIDto struct {
	TraceID     string `json:"trace_id,omitempty"`
	Category    string `json:"category"`
	SocialImage string `json:"socialimage"`
	URL         string `json:"url"`
	Title       string `json:"title"`
}
