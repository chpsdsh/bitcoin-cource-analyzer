package domain

type NewsArticles struct {
	Articles []GdeltApiDto `json:"articles"`
}
type GdeltApiDto struct {
	Category string `json:"category"`
	Url      string `json:"url"`
	Title    string `json:"title"`
}
