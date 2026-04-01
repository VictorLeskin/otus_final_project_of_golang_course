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
func equalIPLists(t *testing.T, got, want []models.IPList) bool {
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
	wantMap := make(map[string]bool)
	for _, item := range want {
		key := fmt.Sprintf("%s:%t", item.Subnet, item.IsWhite)
		wantMap[key] = true
	}
	assert.Equal(t, gotMap, wantMap)

	return true
}

func CreateTestBucketManager() *bucket.BucketManager {
	return bucket.NewBucketManager(&bucket.Config{
		LoginRate:    10,
		PasswordRate: 100,
		IPRate:       1000,
	})
}

func RequestBody(t *testing.T, requestBody map[string]string) []byte {
	var body []byte
	if requestBody != nil {
		var err error
		body, err = json.Marshal(requestBody)
		require.NoError(t, err)
	}
	return body
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
				bucketManager: CreateTestBucketManager(),
				storage:       memStorage,
			}

			// Формируем запрос
			body := RequestBody(t, tt.requestBody)
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
				equalIPLists(t, subnets, []models.IPList{
					models.IPList{Subnet: "192.168.1.0/24", IsWhite: false},
				})
			}
		})
	}
}

func TestAPI_blacklistRemoveHandler(t *testing.T) {
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
			expectedError:  "failed to remove from blacklist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок storage
			memStorage := memorystorage.New()
			memStorage.Add(context.Background(), models.IPList{Subnet: "100.101.102.103/31", IsWhite: false})
			memStorage.Add(context.Background(), models.IPList{Subnet: "192.168.1.0/24", IsWhite: false})

			// Создаем API с моком
			api := &API{
				bucketManager: CreateTestBucketManager(),
				storage:       memStorage,
			}

			// Формируем запрос
			body := RequestBody(t, tt.requestBody)
			req := httptest.NewRequest("POST", "/blacklist/remove", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Вызываем handler
			api.blacklistRemoveHandler(w, req)

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
				equalIPLists(t, subnets, []models.IPList{
					models.IPList{Subnet: "100.101.102.103/31", IsWhite: false},
				})
			}
		})
	}
}
