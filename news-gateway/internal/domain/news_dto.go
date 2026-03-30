package domain

type NewsDto struct {
	Category    string `json:"category"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	SocialImage string `json:"socialimage"`
}
