package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/bucket"
	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/models"
	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/storage"
)

// mockStorage для тестов
type mockStorage struct {
	storage.IPListStorage
	addFunc    func(ctx context.Context, l models.IPList) error
	removeFunc func(ctx context.Context, l models.IPList) error
	getAllFunc func(ctx context.Context, listType models.ListType) ([]string, error)
}

func (m *mockStorage) Add(ctx context.Context, l models.IPList) error {
	if m.addFunc != nil {
		return m.addFunc(ctx, l)
	}
	return nil
}

func (m *mockStorage) Remove(ctx context.Context, l models.IPList) error {
	if m.removeFunc != nil {
		return m.removeFunc(ctx, l)
	}
	return nil
}

func (m *mockStorage) Close() error {
	return nil
}

// TestBlacklistAddHandler тестирует добавление подсети в черный список
func TestBlacklistAddHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		mockAddFunc    func(ctx context.Context, l models.IPList) error
		expectedStatus int
		expectedError  string
	}{
		{
			name: "succsessful",
			requestBody: map[string]string{
				"subnet": "192.168.1.0/24",
			},
			mockAddFunc: func(ctx context.Context, l models.IPList) error {
				assert.Equal(t, "192.168.1.0/24", l.Subnet)
				assert.Equal(t, models.Black, l.IsWhite)
				return nil
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		/*
			{
				name: "пустой subnet",
				requestBody: map[string]string{
					"subnet": "",
				},
				mockAddFunc:    nil,
				expectedStatus: http.StatusBadRequest,
				expectedError:  "subnet is required",
			},
			{
				name: "отсутствует поле subnet",
				requestBody: map[string]string{
					"wrong": "value",
				},
				mockAddFunc:    nil,
				expectedStatus: http.StatusBadRequest,
				expectedError:  "subnet is required",
			},
			{
				name:           "пустое тело запроса",
				requestBody:    nil,
				mockAddFunc:    nil,
				expectedStatus: http.StatusBadRequest,
				expectedError:  "invalid request body",
			},
			{
				name: "ошибка при добавлении в storage",
				requestBody: map[string]string{
					"subnet": "invalid-subnet",
				},
				mockAddFunc: func(ctx context.Context, listType models.ListType, subnet string) error {
					return assert.AnError
				},
				expectedStatus: http.StatusInternalServerError,
				expectedError:  "failed to add to blacklist",
			},
		*/
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок storage
			mockSt := &mockStorage{
				addFunc: tt.mockAddFunc,
			}

			// Создаем API с моком
			api := &API{
				bucketManager: bucket.NewBucketManager(&bucket.Config{
					LoginRate:    10,
					PasswordRate: 100,
					IPRate:       1000,
				}),
				storage: mockSt,
			}

			// Формируем запрос
			var body []byte
			if tt.requestBody != nil {
				var err error
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/blacklist/add", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Вызываем handler
			api.blacklistAddHandler(w, req)

			// Проверяем статус
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				// Проверяем тело ответа
				var resp ErrorResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Contains(t, resp.Error, tt.expectedError)
			} else {
				var resp SuccessfulResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, "ok", resp.Status) // предполагаем что sendJSON отправляет map[string]string{"status": "ok"}
			}
		})
	}
}
