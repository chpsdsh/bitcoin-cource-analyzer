package domain

type ArticleDto struct {
	Category string `json:"category"`
	Title    string `json:"title"`
	Text     string `json:"text"`
}
