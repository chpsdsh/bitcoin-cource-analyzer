package domain

type ArticleDto struct {
	TraceID  string `json:"trace_id,omitempty"`
	Category string `json:"category"`
	Title    string `json:"title"`
	Text     string `json:"text"`
}
