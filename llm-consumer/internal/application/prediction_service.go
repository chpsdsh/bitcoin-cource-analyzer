package application

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

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
	slowReaction          = 0
	mediumReaction        = 1
	fastReaction          = 2
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
		return domain.Prediction{}, fmt.Errorf("request btc price: %w", err)
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

	return domain.Prediction{
		Target:      btcPrice * math.Exp(rPred),
		Current:     btcPrice,
		PredHorizon: calculatePredHorizon(),
	}, nil
}

func calculatePredHorizon() int {
	/*
	   PredictHorizon computes a discrete market reaction horizon (3 levels) using simple calendar/time
	   proxies of BTC/USD trading activity.

	   Output levels:
	     0 = slow reaction
	     1 = medium reaction
	     2 = fast reaction

	     predHorizon = 0 if volumeLevel == 0
	                 = 1 if volumeLevel == 1
	                 = 2 if volumeLevel >= 2

	   Rationale:
	     The selected windows were derived from analysis of historical volatility/activity patterns and are
	     consistent with documented time-of-day/day-of-week/month-of-year effects in Bitcoin trading activity,
	     e.g., Baur, Cahill, Godfrey & Liu (2019), "Bitcoin time-of-day, day-of-week and month-of-year effects
	     in returns and trading volume".
	*/

	now := time.Now().UTC()

	dayOfWeek := 0
	if now.Weekday() == time.Thursday || now.Weekday() == time.Friday {
		dayOfWeek = 1
	}

	var month int

	switch now.Month() {
	case time.October, time.November, time.December, time.January, time.February:
		month = 1
	case time.March, time.April, time.May, time.June, time.July, time.August, time.September:
		month = 0
	}

	timeUTC := 0
	if now.Hour() >= 13 && now.Hour() < 21 {
		timeUTC = 1
	}

	volumeLevel := dayOfWeek + month + timeUTC

	if volumeLevel == 0 {
		return 0
	}
	if volumeLevel == 1 {
		return mediumReaction
	}
	return fastReaction
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
