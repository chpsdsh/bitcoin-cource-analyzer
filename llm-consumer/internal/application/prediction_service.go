package application

import (
	"context"
	"errors"

	"llm-consumer/internal/domain"
)

var (
	ErrNotValidCategory = errors.New("invalid category")
	ErrNoCategoryFound  = errors.New("no category found")
	ErrPredicting       = errors.New("predicting error")
)

const (
	environmentCategory = "environment"
	politicsCategory    = "politics"
	economyCategory     = "economy"
	technologyCategory  = "technology"
	cryptoCategory      = "crypto"
)

type PredictionStorage interface {
	GetPrediction(ctx context.Context, category string) (domain.LLMResponse, error)
}

type PredictionService struct {
	storage PredictionStorage
}

func NewPredictionService(storage PredictionStorage) *PredictionService {
	return &PredictionService{storage: storage}
}

func (s PredictionService) ValidateCategories(categories domain.Categories) error {
	if len(categories.CategoriesList) == 0 {
		return ErrNoCategoryFound
	}

	for _, category := range categories.CategoriesList {
		switch category {
		case politicsCategory, environmentCategory, economyCategory, technologyCategory, cryptoCategory:
			continue
		default:
			return ErrNotValidCategory
		}
	}
	return nil
}

func (s PredictionService) DoPrediction(ctx context.Context, categories domain.Categories) (domain.Prediction, error) {
	var scores float64
	for _, category := range categories.CategoriesList {
		prediction, err := s.storage.GetPrediction(ctx, category)
		if err != nil {
			return domain.Prediction{}, errors.Join(ErrPredicting, err)
		}
		scores += calculateScores(prediction)
	}
	return domain.Prediction{Target: scores}, nil
}

func calculateScores(prediction domain.LLMResponse) float64 {
	scores := prediction.Features.EntityDensity + prediction.Features.NumbersDensity //TODO: make a formula
	return scores
}
