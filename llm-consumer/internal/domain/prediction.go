package domain

type Prediction struct {
	Target      float64 `json:"target"`
	Current     float64 `json:"current"`
	PredHorizon int     `json:"pred_horizon"`
}
