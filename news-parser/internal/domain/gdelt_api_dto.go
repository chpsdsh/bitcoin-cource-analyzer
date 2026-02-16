package domain

type NewsArticles struct {
	Articles []GdeltApiDto `json:"articles"`
}
type GdeltApiDto struct {
	Url   string `json:"url"`
	Title string `json:"title"`
}
