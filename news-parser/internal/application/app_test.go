package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"news-parser/internal/application/mocks"
	"news-parser/internal/domain"
	"news-parser/internal/observability"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplicationParseNews(t *testing.T) {
	const traceID = "test-trace-id"

	tests := []struct {
		name              string
		setupHandler      func(*gomock.Controller) RequestHandler
		expectedRequests  []domain.GdeltAPIDto
		expectedNewsItems []domain.NewsDto
	}{
		{
			name: "sends articles to both channels with normalized categories",
			setupHandler: func(ctrl *gomock.Controller) RequestHandler {
				handler := mocks.NewMockRequestHandler(ctrl)
				for _, category := range domain.AllCategories {
					currentCategory := category
					handler.EXPECT().
						DoNewsRequest(gomock.Any(), currentCategory).
						Return(domain.Articles{Articles: []domain.GdeltAPIDto{{
							Title:       "title-" + domain.CategoryToString(currentCategory),
							URL:         "https://example.com/" + domain.CategoryToString(currentCategory),
							SocialImage: "https://example.com/image-" + domain.CategoryToString(currentCategory),
						}}}, nil)
				}
				return handler
			},
			expectedRequests: []domain.GdeltAPIDto{
				{TraceID: traceID, Category: "politics", Title: "title-politics", URL: "https://example.com/politics", SocialImage: "https://example.com/image-politics"},
				{TraceID: traceID, Category: "environment", Title: "title-environment", URL: "https://example.com/environment", SocialImage: "https://example.com/image-environment"},
				{TraceID: traceID, Category: "economy", Title: "title-economy", URL: "https://example.com/economy", SocialImage: "https://example.com/image-economy"},
				{TraceID: traceID, Category: "technology", Title: "title-technology", URL: "https://example.com/technology", SocialImage: "https://example.com/image-technology"},
				{TraceID: traceID, Category: "crypto", Title: "title-crypto", URL: "https://example.com/crypto", SocialImage: "https://example.com/image-crypto"},
			},
			expectedNewsItems: []domain.NewsDto{
				{TraceID: traceID, Category: "politics", Title: "title-politics", URL: "https://example.com/politics", SocialImage: "https://example.com/image-politics"},
				{TraceID: traceID, Category: "environment", Title: "title-environment", URL: "https://example.com/environment", SocialImage: "https://example.com/image-environment"},
				{TraceID: traceID, Category: "economy", Title: "title-economy", URL: "https://example.com/economy", SocialImage: "https://example.com/image-economy"},
				{TraceID: traceID, Category: "technology", Title: "title-technology", URL: "https://example.com/technology", SocialImage: "https://example.com/image-technology"},
				{TraceID: traceID, Category: "crypto", Title: "title-crypto", URL: "https://example.com/crypto", SocialImage: "https://example.com/image-crypto"},
			},
		},
		{
			name: "skips category when news request fails",
			setupHandler: func(ctrl *gomock.Controller) RequestHandler {
				handler := mocks.NewMockRequestHandler(ctrl)
				handler.EXPECT().
					DoNewsRequest(gomock.Any(), domain.PoliticsCategory).
					Return(domain.Articles{}, errors.New("request failed"))
				handler.EXPECT().
					DoNewsRequest(gomock.Any(), domain.EnvironmentCategory).
					Return(domain.Articles{Articles: []domain.GdeltAPIDto{{
						Title:       "title-environment",
						URL:         "https://example.com/environment",
						SocialImage: "https://example.com/image-environment",
					}}}, nil)
				handler.EXPECT().
					DoNewsRequest(gomock.Any(), domain.EconomyCategory).
					Return(domain.Articles{Articles: []domain.GdeltAPIDto{{
						Title:       "title-economy",
						URL:         "https://example.com/economy",
						SocialImage: "https://example.com/image-economy",
					}}}, nil)
				handler.EXPECT().
					DoNewsRequest(gomock.Any(), domain.TechnologyCategory).
					Return(domain.Articles{}, nil)
				handler.EXPECT().
					DoNewsRequest(gomock.Any(), domain.CryptoCategory).
					Return(domain.Articles{}, nil)
				return handler
			},
			expectedRequests: []domain.GdeltAPIDto{
				{TraceID: traceID, Category: "environment", Title: "title-environment", URL: "https://example.com/environment", SocialImage: "https://example.com/image-environment"},
				{TraceID: traceID, Category: "economy", Title: "title-economy", URL: "https://example.com/economy", SocialImage: "https://example.com/image-economy"},
			},
			expectedNewsItems: []domain.NewsDto{
				{TraceID: traceID, Category: "environment", Title: "title-environment", URL: "https://example.com/environment", SocialImage: "https://example.com/image-environment"},
				{TraceID: traceID, Category: "economy", Title: "title-economy", URL: "https://example.com/economy", SocialImage: "https://example.com/image-economy"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			requestChan := make(chan domain.GdeltAPIDto, 16)
			newsChan := make(chan domain.NewsDto, 16)

			app := Application{
				RequestHandler: tt.setupHandler(ctrl),
				RequestChan:    requestChan,
				NewsChan:       newsChan,
			}

			app.parseNews(observability.ContextWithTraceID(context.Background(), traceID))

			require.Eventually(t, func() bool {
				return len(requestChan) == len(tt.expectedRequests) && len(newsChan) == len(tt.expectedNewsItems)
			}, time.Second, 10*time.Millisecond)

			assert.ElementsMatch(t, tt.expectedRequests, drainRequestChan(requestChan, len(tt.expectedRequests)))
			assert.ElementsMatch(t, tt.expectedNewsItems, drainNewsChan(newsChan, len(tt.expectedNewsItems)))
		})
	}
}

func TestApplicationStartParsingNews(t *testing.T) {
	tests := []struct {
		name          string
		notifierError error
	}{
		{
			name:          "starts parsing on tick and closes request channel on cancel",
			notifierError: nil,
		},
		{
			name:          "continues shutdown when notifier returns error",
			notifierError: errors.New("llm unavailable"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			ctx, cancel := context.WithCancel(context.Background())
			t.Cleanup(cancel)

			handler := mocks.NewMockRequestHandler(ctrl)
			for _, category := range domain.AllCategories {
				handler.EXPECT().
					DoNewsRequest(gomock.Any(), category).
					Return(domain.Articles{}, nil)
			}

			notifier := mocks.NewMockLLMNotifier(ctrl)
			notifier.EXPECT().
				StartLLMPrediction(gomock.Any()).
				DoAndReturn(func(context.Context) error {
					cancel()
					return tt.notifierError
				})

			requestChan := make(chan domain.GdeltAPIDto, 1)

			app := Application{
				Ticker:         time.NewTicker(10 * time.Millisecond),
				RequestHandler: handler,
				LLMNotifier:    notifier,
				RequestChan:    requestChan,
				NewsChan:       make(chan domain.NewsDto, 1),
			}

			done := make(chan struct{})
			go func() {
				defer close(done)
				app.StartParsingNews(ctx)
			}()

			require.Eventually(t, func() bool {
				select {
				case <-done:
					return true
				default:
					return false
				}
			}, time.Second, 10*time.Millisecond)

			select {
			case _, ok := <-requestChan:
				assert.False(t, ok)
			default:
				t.Fatal("request channel was not closed")
			}
		})
	}
}

func drainRequestChan(ch <-chan domain.GdeltAPIDto, count int) []domain.GdeltAPIDto {
	result := make([]domain.GdeltAPIDto, 0, count)
	for range count {
		result = append(result, <-ch)
	}
	return result
}

func drainNewsChan(ch <-chan domain.NewsDto, count int) []domain.NewsDto {
	result := make([]domain.NewsDto, 0, count)
	for range count {
		result = append(result, <-ch)
	}
	return result
}
