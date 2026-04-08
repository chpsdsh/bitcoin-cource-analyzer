package application

import (
	"context"
	"errors"
	"math"
	"strconv"

	"llm-consumer/internal/domain"
)

var (
	ErrNotValidCategory = errors.New("invalid category")
	ErrNoCategoryFound  = errors.New("no category found")
	ErrPredicting       = errors.New("predicting error")
)

const (
	environmentCategory   = "environment"
	politicsCategory      = "politics"
	economyCategory       = "economy"
	technologyCategory    = "technology"
	cryptoCategory        = "crypto"
	upDirection           = "up"
	downDirection         = "down"
	neutralDirection      = "neutral"
	defaultCategoryWeight = 1.0
	rMax                  = 0.03
)

type CryptoClient interface {
	RequestBTCPrice() (domain.PriceResponse, error)
}

type PredictionStorage interface {
	GetPrediction(ctx context.Context, category string) (domain.LLMResponse, error)
}

type PredictionService struct {
	storage PredictionStorage
	client  CryptoClient
}

func NewPredictionService(storage PredictionStorage, client CryptoClient) *PredictionService {
	return &PredictionService{storage: storage, client: client}
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
	priceDto, err := s.client.RequestBTCPrice()
	if err != nil {
		return domain.Prediction{}, err
	}
	btcPrice, err := strconv.ParseFloat(priceDto.Price, 64)
	if err != nil {
		return domain.Prediction{}, errors.Join(ErrPredicting, err)
	}

	categoriesPredictions := make([]domain.LLMResponse, len(categories.CategoriesList))

	for i, category := range categories.CategoriesList {
		prediction, predictionErr := s.storage.GetPrediction(ctx, category)
		if predictionErr != nil {
			return domain.Prediction{}, errors.Join(ErrPredicting, predictionErr)
		}
		categoriesPredictions[i] = prediction
	}
	rPred := calcRPred(categoriesPredictions)

	return domain.Prediction{Target: btcPrice * math.Exp(rPred)}, nil
}

func directionToNumber(dir string) float64 {
	switch dir {
	case upDirection:
		return 1.0
	case downDirection:
		return -1.0
	case neutralDirection:
		return 0.0
	default:
		return 0.0
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func infoWeight(numbersDensity, entityDensity float64) float64 {
	n := clamp01(numbersDensity)
	e := clamp01(entityDensity)
	return 0.7 + 0.15*n + 0.15*e
}

func categoryScore(pred domain.LLMResponse) float64 {
	d := directionToNumber(pred.Features.SignalDirection)
	s := clamp01(pred.Features.SignalStrength)
	u := clamp01(pred.Features.Uncertainty)
	n := clamp01(pred.Features.NumbersDensity)
	e := clamp01(pred.Features.EntityDensity)

	c := 1.0 - u
	wInfo := infoWeight(n, e)

	return d * s * c * wInfo
}

func calcRPred(predictions []domain.LLMResponse) float64 {
	var weightedScoreSum float64
	var weightsSum float64

	for _, pred := range predictions {
		wCat := defaultCategoryWeight
		score := categoryScore(pred)

		weightedScoreSum += wCat * score
		weightsSum += wCat
	}

	if weightsSum == 0 {
		return 0
	}

	return rMax * (weightedScoreSum / weightsSum)
}
