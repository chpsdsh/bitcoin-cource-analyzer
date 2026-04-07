package domain

type Articles struct {
	Articles []GdeltApiDto `json:"articles"`
}
type GdeltApiDto struct {
	Category    string `json:"category"`
	SocialImage string `json:"socialimage"`
	URL         string `json:"url"`
	Title       string `json:"title"`
}
