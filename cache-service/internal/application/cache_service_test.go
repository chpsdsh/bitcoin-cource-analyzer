package application

import (
	"context"
	"errors"
	"testing"

	"data-cleaner/internal/application/mocks"
	"data-cleaner/internal/domain"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestCacheServiceHandleArticle(t *testing.T) {
	tests := []struct {
		name         string
		article      domain.ArticleDto
		setupStorage func(*gomock.Controller) Storage
	}{
		{
			name: "stores article successfully",
			article: domain.ArticleDto{
				Category: "crypto",
				Title:    "BTC update",
				Text:     "Bitcoin moved higher",
			},
			setupStorage: func(ctrl *gomock.Controller) Storage {
				storage := mocks.NewMockStorage(ctrl)
				storage.EXPECT().AddArticle(gomock.Any(), domain.ArticleDto{
					Category: "crypto",
					Title:    "BTC update",
					Text:     "Bitcoin moved higher",
				}).Return(nil)
				return storage
			},
		},
		{
			name: "logs and continues when storing article fails",
			article: domain.ArticleDto{
				Category: "economy",
				Title:    "Inflation cools",
				Text:     "Inflation slowed this month",
			},
			setupStorage: func(ctrl *gomock.Controller) Storage {
				storage := mocks.NewMockStorage(ctrl)
				storage.EXPECT().AddArticle(gomock.Any(), domain.ArticleDto{
					Category: "economy",
					Title:    "Inflation cools",
					Text:     "Inflation slowed this month",
				}).Return(errors.New("redis unavailable"))
				return storage
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := CacheService{Storage: tt.setupStorage(ctrl)}
			assert.NotPanics(t, func() {
				service.HandleArticle(context.Background(), tt.article)
			})
		})
	}
}

func TestCacheServiceHandleNews(t *testing.T) {
	tests := []struct {
		name         string
		news         domain.NewsDto
		setupStorage func(*gomock.Controller) Storage
	}{
		{
			name: "stores news successfully",
			news: domain.NewsDto{
				Category:    "politics",
				Title:       "Election update",
				URL:         "https://example.com/election",
				SocialImage: "https://example.com/election.png",
			},
			setupStorage: func(ctrl *gomock.Controller) Storage {
				storage := mocks.NewMockStorage(ctrl)
				storage.EXPECT().AddNews(gomock.Any(), domain.NewsDto{
					Category:    "politics",
					Title:       "Election update",
					URL:         "https://example.com/election",
					SocialImage: "https://example.com/election.png",
				}).Return(nil)
				return storage
			},
		},
		{
			name: "logs and continues when storing news fails",
			news: domain.NewsDto{
				Category:    "technology",
				Title:       "GPU launch",
				URL:         "https://example.com/gpu",
				SocialImage: "https://example.com/gpu.png",
			},
			setupStorage: func(ctrl *gomock.Controller) Storage {
				storage := mocks.NewMockStorage(ctrl)
				storage.EXPECT().AddNews(gomock.Any(), domain.NewsDto{
					Category:    "technology",
					Title:       "GPU launch",
					URL:         "https://example.com/gpu",
					SocialImage: "https://example.com/gpu.png",
				}).Return(errors.New("redis unavailable"))
				return storage
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := CacheService{Storage: tt.setupStorage(ctrl)}
			assert.NotPanics(t, func() {
				service.HandleNews(context.Background(), tt.news)
			})
		})
	}
}

func TestCacheServiceHandleLLMResponse(t *testing.T) {
	tests := []struct {
		name         string
		response     domain.LLMResponse
		setupStorage func(*gomock.Controller) Storage
	}{
		{
			name:     "stores llm response successfully",
			response: makeLLMResponse("crypto", "up", 0.8),
			setupStorage: func(ctrl *gomock.Controller) Storage {
				storage := mocks.NewMockStorage(ctrl)
				storage.EXPECT().AddLLMResponse(gomock.Any(), makeLLMResponse("crypto", "up", 0.8)).Return(nil)
				return storage
			},
		},
		{
			name:     "logs and continues when storing llm response fails",
			response: makeLLMResponse("economy", "down", 0.4),
			setupStorage: func(ctrl *gomock.Controller) Storage {
				storage := mocks.NewMockStorage(ctrl)
				storage.EXPECT().AddLLMResponse(gomock.Any(), makeLLMResponse("economy", "down", 0.4)).Return(errors.New("redis unavailable"))
				return storage
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := CacheService{Storage: tt.setupStorage(ctrl)}
			assert.NotPanics(t, func() {
				service.HandleLLMResponse(context.Background(), tt.response)
			})
		})
	}
}

func makeLLMResponse(category, direction string, strength float64) domain.LLMResponse {
	var response domain.LLMResponse
	response.Category = category
	response.Summarization = "summary"
	response.Features.SignalDirection = direction
	response.Features.SignalStrength = strength
	response.Features.Uncertainty = 0.2
	response.Features.EventUrgencyHours = 6
	response.Features.NumbersDensity = 0.5
	response.Features.EntityDensity = 0.4
	return response
}
