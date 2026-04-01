package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/bucket"
	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/models"
	memorystorage "github.com/VictorLeskin/otus_final_project_of_golang_course/internal/storage/memory"
)

// equalIPLists сравнивает два списка IPList по Subnet и IsWhite (игнорирует ID и CreatedAt)
func equalIPLists(got, want []models.IPList) bool {
	if len(got) != len(want) {
		return false
	}

	// Создаем map для быстрого поиска
	gotMap := make(map[string]bool)
	for _, item := range got {
		key := fmt.Sprintf("%s:%t", item.Subnet, item.IsWhite)
		gotMap[key] = true
	}

	// Проверяем, что все ожидаемые элементы есть
	for _, item := range want {
		key := fmt.Sprintf("%s:%t", item.Subnet, item.IsWhite)
		if !gotMap[key] {
			return false
		}
	}

	return true
}

// TestBlacklistAddHandler тестирует добавление подсети в черный список
func TestAPI_blacklistAddHandler(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]string
		expectedStatus int
		expectedError  string
	}{
		{
			name: "succsessful",
			requestBody: map[string]string{
				"subnet": "192.168.1.0/24",
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
		},
		{
			name: "empty subnet",
			requestBody: map[string]string{
				"subnet": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "subnet is required",
		},
		{
			name: "wrong JSON",
			requestBody: map[string]string{
				"wrong": "value",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "subnet is required",
		},
		{
			name:           "missed request",
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid request body",
		},
		{
			name: "ошибка при добавлении в storage",
			requestBody: map[string]string{
				"subnet": "invalid-subnet",
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to add to blacklist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок storage
			memStorage := memorystorage.New()

			// Создаем API с моком
			api := &API{
				bucketManager: bucket.NewBucketManager(&bucket.Config{
					LoginRate:    10,
					PasswordRate: 100,
					IPRate:       1000,
				}),
				storage: memStorage,
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

				subnets, _ := memStorage.GetAll(context.Background())
				assert.True(t, equalIPLists(subnets, []models.IPList{
					models.IPList{Subnet: "192.168.1.0/24", IsWhite: false},
				}))
			}
		})
	}
}
