package application

import (
	"context"
	"errors"
	"testing"

	"news-gateway/internal/application/mocks"
	"news-gateway/internal/domain"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewsServiceRequestNews(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		setupStorage  func(*gomock.Controller) Storage
		expected      []domain.NewsDto
		expectedError error
	}{
		{
			name: "returns news for valid category",
			key:  politicsCategory,
			setupStorage: func(ctrl *gomock.Controller) Storage {
				storage := mocks.NewMockStorage(ctrl)
				storage.EXPECT().
					GetNews(gomock.Any(), int64(newsLimit), politicsCategory).
					Return([]domain.NewsDto{{
						Title:       "Market update",
						URL:         "https://example.com/news",
						SocialImage: "https://example.com/news.png",
						Category:    politicsCategory,
					}}, nil)
				return storage
			},
			expected: []domain.NewsDto{{
				Title:       "Market update",
				URL:         "https://example.com/news",
				SocialImage: "https://example.com/news.png",
				Category:    politicsCategory,
			}},
		},
		{
			name: "returns joined internal error when storage fails",
			key:  cryptoCategory,
			setupStorage: func(ctrl *gomock.Controller) Storage {
				storage := mocks.NewMockStorage(ctrl)
				storage.EXPECT().
					GetNews(gomock.Any(), int64(newsLimit), cryptoCategory).
					Return(nil, errors.New("redis unavailable"))
				return storage
			},
			expectedError: ErrInternalError,
		},
		{
			name: "returns validation error for unknown category",
			key:  "sports",
			setupStorage: func(ctrl *gomock.Controller) Storage {
				return mocks.NewMockStorage(ctrl)
			},
			expectedError: ErrNotValidCategory,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			service := NewsService{
				Storage: tt.setupStorage(ctrl),
			}

			result, err := service.RequestNews(context.Background(), tt.key)

			if tt.expectedError != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.expectedError)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
