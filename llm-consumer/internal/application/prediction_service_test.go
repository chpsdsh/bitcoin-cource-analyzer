package application

import (
	"context"
	"errors"
	"math"
	"testing"

	"llm-consumer/internal/application/mocks"
	"llm-consumer/internal/domain"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPredictionServiceValidateCategories(t *testing.T) {
	tests := []struct {
		name          string
		categories    domain.Categories
		expectedError error
	}{
		{
			name: "accepts valid categories",
			categories: domain.Categories{
				CategoriesList: []string{politicsCategory, environmentCategory, economyCategory, technologyCategory, cryptoCategory},
			},
		},
		{
			name:          "returns error on empty category list",
			categories:    domain.Categories{},
			expectedError: ErrNoCategoryFound,
		},
		{
			name: "returns error on unknown category",
			categories: domain.Categories{
				CategoriesList: []string{politicsCategory, "sports"},
			},
			expectedError: ErrNotValidCategory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := PredictionService{}
			err := service.ValidateCategories(tt.categories)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedError)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestPredictionServiceDoPrediction(t *testing.T) {
	tests := []struct {
		name           string
		categories     domain.Categories
		setupClient    func(*gomock.Controller) CryptoClient
		setupStorage   func(*gomock.Controller) PredictionStorage
		expectedTarget float64
		expectedCurr   float64
		expectedError  error
	}{
		{
			name: "returns prediction for valid inputs",
			categories: domain.Categories{
				CategoriesList: []string{politicsCategory, cryptoCategory},
			},
			setupClient: func(ctrl *gomock.Controller) CryptoClient {
				client := mocks.NewMockCryptoClient(ctrl)
				client.EXPECT().
					RequestBTCPrice().
					Return(domain.PriceResponse{Price: "100000"}, nil)
				return client
			},
			setupStorage: func(ctrl *gomock.Controller) PredictionStorage {
				storage := mocks.NewMockPredictionStorage(ctrl)
				storage.EXPECT().
					GetPrediction(gomock.Any(), politicsCategory).
					Return(makePrediction(upDirection, 0.8, 0.1, 0.9, 0.8), nil)
				storage.EXPECT().
					GetPrediction(gomock.Any(), cryptoCategory).
					Return(makePrediction(downDirection, 0.2, 0.1, 0.2, 0.1), nil)
				return storage
			},
			expectedCurr:   100000,
			expectedTarget: 100000 * math.Exp(calcRPred([]domain.LLMResponse{makePrediction(upDirection, 0.8, 0.1, 0.9, 0.8), makePrediction(downDirection, 0.2, 0.1, 0.2, 0.1)})),
		},
		{
			name: "returns wrapped error when price request fails",
			categories: domain.Categories{
				CategoriesList: []string{politicsCategory},
			},
			setupClient: func(ctrl *gomock.Controller) CryptoClient {
				client := mocks.NewMockCryptoClient(ctrl)
				client.EXPECT().
					RequestBTCPrice().
					Return(domain.PriceResponse{}, errors.New("binance unavailable"))
				return client
			},
			setupStorage: func(ctrl *gomock.Controller) PredictionStorage {
				return mocks.NewMockPredictionStorage(ctrl)
			},
			expectedError: errors.New("request btc price"),
		},
		{
			name: "returns predicting error when price is invalid",
			categories: domain.Categories{
				CategoriesList: []string{politicsCategory},
			},
			setupClient: func(ctrl *gomock.Controller) CryptoClient {
				client := mocks.NewMockCryptoClient(ctrl)
				client.EXPECT().
					RequestBTCPrice().
					Return(domain.PriceResponse{Price: "not-a-number"}, nil)
				return client
			},
			setupStorage: func(ctrl *gomock.Controller) PredictionStorage {
				return mocks.NewMockPredictionStorage(ctrl)
			},
			expectedError: ErrPredicting,
		},
		{
			name: "returns predicting error when storage fails",
			categories: domain.Categories{
				CategoriesList: []string{technologyCategory},
			},
			setupClient: func(ctrl *gomock.Controller) CryptoClient {
				client := mocks.NewMockCryptoClient(ctrl)
				client.EXPECT().
					RequestBTCPrice().
					Return(domain.PriceResponse{Price: "100000"}, nil)
				return client
			},
			setupStorage: func(ctrl *gomock.Controller) PredictionStorage {
				storage := mocks.NewMockPredictionStorage(ctrl)
				storage.EXPECT().
					GetPrediction(gomock.Any(), technologyCategory).
					Return(domain.LLMResponse{}, errors.New("redis unavailable"))
				return storage
			},
			expectedError: ErrPredicting,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := NewPredictionService(tt.setupStorage(ctrl), tt.setupClient(ctrl))
			result, err := service.DoPrediction(context.Background(), tt.categories)

			if tt.expectedError != nil {
				require.Error(t, err)
				require.ErrorContains(t, err, tt.expectedError.Error())
				assert.Equal(t, domain.Prediction{}, result)
				return
			}

			require.NoError(t, err)
			assert.InDelta(t, tt.expectedCurr, result.Current, 1e-9)
			assert.InDelta(t, tt.expectedTarget, result.Target, 1e-9)
			assert.Contains(t, []int{slowReaction, mediumReaction, fastReaction}, result.PredHorizon)
		})
	}
}

func TestPredictionMathHelpers(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "direction to number maps all directions",
			run: func(t *testing.T) {
				assert.InDelta(t, 1.0, directionToNumber(upDirection), 1e-12)
				assert.InDelta(t, -1.0, directionToNumber(downDirection), 1e-12)
				assert.InDelta(t, 0.0, directionToNumber(neutralDirection), 1e-12)
				assert.InDelta(t, 0.0, directionToNumber("unknown"), 1e-12)
			},
		},
		{
			name: "clamp01 limits values to range",
			run: func(t *testing.T) {
				assert.InDelta(t, 0.0, clamp01(-1), 1e-12)
				assert.InDelta(t, 0.4, clamp01(0.4), 1e-12)
				assert.InDelta(t, 1.0, clamp01(2), 1e-12)
			},
		},
		{
			name: "calc r pred returns zero for empty list",
			run: func(t *testing.T) {
				assert.InDelta(t, 0.0, calcRPred(nil), 1e-12)
			},
		},
		{
			name: "category score respects uncertainty and densities",
			run: func(t *testing.T) {
				pred := makePrediction(upDirection, 0.5, 0.25, 0.5, 1.0)
				expected := 1.0 * 0.5 * (1.0 - 0.25) * (0.7 + 0.15*0.5 + 0.15*1.0)
				assert.InDelta(t, expected, categoryScore(pred), 1e-12)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func makePrediction(direction string, strength, uncertainty, numbersDensity, entityDensity float64) domain.LLMResponse {
	var pred domain.LLMResponse
	pred.Features.SignalDirection = direction
	pred.Features.SignalStrength = strength
	pred.Features.Uncertainty = uncertainty
	pred.Features.NumbersDensity = numbersDensity
	pred.Features.EntityDensity = entityDensity
	return pred
}
