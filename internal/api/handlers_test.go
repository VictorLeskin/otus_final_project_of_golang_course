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

// equalIPLists сравнивает два списка IPList по Subnet и IsWhite (игнорирует ID и CreatedAt).
func equalIPLists(t *testing.T, got, want []models.IPList) {
	t.Helper()
	if len(got) != len(want) {
		return
	}

	// Создаем map для быстрого поиска.
	gotMap := make(map[string]bool)
	for _, item := range got {
		key := fmt.Sprintf("%s:%t", item.Subnet, item.IsWhite)
		gotMap[key] = true
	}

	// Проверяем, что все ожидаемые элементы есть.
	wantMap := make(map[string]bool)
	for _, item := range want {
		key := fmt.Sprintf("%s:%t", item.Subnet, item.IsWhite)
		wantMap[key] = true
	}
	assert.Equal(t, gotMap, wantMap)
}

func CreateTestBucketManager() *bucket.BucketManager {
	return bucket.NewBucketManager(&bucket.Config{
		LoginRate:    10,
		PasswordRate: 100,
		IPRate:       1000,
	})
}

func RequestBody(t *testing.T, requestBody map[string]string) []byte {
	t.Helper()
	var body []byte
	if requestBody != nil {
		var err error
		body, err = json.Marshal(requestBody)
		require.NoError(t, err)
	}
	return body
}

func checkErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedError string) {
	t.Helper()
	var resp ErrorResponse
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.Contains(t, resp.Error, expectedError)
}

func TestAPI_checkRequest_validate(t *testing.T) {
	{
		t0 := CheckRequest{
			Login:    "login",
			Password: "password",
			IP:       "IP",
		}
		assert.True(t, t0.validate())
	}
	{
		t0 := CheckRequest{
			Login:    "",
			Password: "password",
			IP:       "IP",
		}
		assert.False(t, t0.validate())
	}
	{
		t0 := CheckRequest{
			Login:    "login",
			Password: "",
			IP:       "IP",
		}
		assert.False(t, t0.validate())
	}
	{
		t0 := CheckRequest{
			Login:    "login",
			Password: "password",
			IP:       "",
		}
		assert.False(t, t0.validate())
	}
}

// TestBlacklistAddHandler тестирует добавление подсети в черный список.
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
			// Создаем мок storage.
			memStorage := memorystorage.New()

			// Создаем API с моком.
			api := &API{
				bucketManager: CreateTestBucketManager(),
				storage:       memStorage,
			}

			// Формируем запрос.
			body := RequestBody(t, tt.requestBody)
			req := httptest.NewRequest("POST", "/blacklist/add", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Вызываем handler.
			api.blacklistAddHandler(w, req)

			// Проверяем статус.
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				// Проверяем тело ответа.
				checkErrorResponse(t, w, tt.expectedError)
			} else {
				var resp SuccessfulResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, "ok", resp.Status) // предполагаем что sendJSON отправляет map[string]string{"status": "ok"}.

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
			// Создаем мок storage.
			memStorage := memorystorage.New()
			memStorage.Add(context.Background(), models.IPList{Subnet: "100.101.102.103/31", IsWhite: false})
			memStorage.Add(context.Background(), models.IPList{Subnet: "192.168.1.0/24", IsWhite: false})

			// Создаем API с моком.
			api := &API{
				bucketManager: CreateTestBucketManager(),
				storage:       memStorage,
			}

			// Формируем запрос.
			body := RequestBody(t, tt.requestBody)
			req := httptest.NewRequest("POST", "/blacklist/remove", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Вызываем handler.
			api.blacklistRemoveHandler(w, req)

			// Проверяем статус.
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				// Проверяем тело ответа.
				checkErrorResponse(t, w, tt.expectedError)
			} else {
				var resp SuccessfulResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, "ok", resp.Status) // предполагаем что sendJSON отправляет map[string]string{"status": "ok"}.

				subnets, _ := memStorage.GetAll(context.Background())
				equalIPLists(t, subnets, []models.IPList{
					models.IPList{Subnet: "100.101.102.103/31", IsWhite: false},
				})
			}
		})
	}
}

func TestAPI_blacklistHandler(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectedError  string
		cancel         bool
	}{
		{
			name:           "succsessful",
			expectedStatus: http.StatusOK,
			expectedError:  "",
			cancel:         false,
		},
		{
			name:           "server send error",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get blacklist",
			cancel:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок storage.
			memStorage := memorystorage.New()
			memStorage.Add(context.Background(), models.IPList{Subnet: "100.101.102.103/31", IsWhite: false})
			memStorage.Add(context.Background(), models.IPList{Subnet: "110.111.112.113/31", IsWhite: true})
			memStorage.Add(context.Background(), models.IPList{Subnet: "192.168.1.0/24", IsWhite: false})

			// Создаем API с моком.
			api := &API{
				bucketManager: CreateTestBucketManager(),
				storage:       memStorage,
			}

			// Формируем запрос.
			ctx1, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.cancel {
				cancel() // run cancel to cause server error.
			}

			req := httptest.NewRequestWithContext(ctx1, "GET", "/blacklist", nil)
			w := httptest.NewRecorder()

			// Вызываем handler.
			api.blacklistHandler(w, req)

			// Проверяем статус.
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				// Проверяем тело ответа.
				checkErrorResponse(t, w, tt.expectedError)
			} else {
				var resp SuccessfulResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)

				subnets, _ := memStorage.GetAll(context.Background())
				equalIPLists(t, subnets, []models.IPList{
					models.IPList{Subnet: "100.101.102.103/31", IsWhite: false},
					models.IPList{Subnet: "192.168.1.0/24", IsWhite: false},
				})
			}
		})
	}
}

func TestAPI_whitelistAddHandler(t *testing.T) {
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
			expectedError:  "failed to add to whitelist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок storage.
			memStorage := memorystorage.New()

			// Создаем API с моком.
			api := &API{
				bucketManager: CreateTestBucketManager(),
				storage:       memStorage,
			}

			// Формируем запрос.
			body := RequestBody(t, tt.requestBody)
			req := httptest.NewRequest("POST", "/whitelist/add", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Вызываем handler.
			api.whitelistAddHandler(w, req)

			// Проверяем статус.
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				// Проверяем тело ответа.
				checkErrorResponse(t, w, tt.expectedError)
			} else {
				var resp SuccessfulResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, "ok", resp.Status) // предполагаем что sendJSON отправляет map[string]string{"status": "ok"}.

				subnets, _ := memStorage.GetAll(context.Background())
				equalIPLists(t, subnets, []models.IPList{
					models.IPList{Subnet: "192.168.1.0/24", IsWhite: true},
				})
			}
		})
	}
}

func TestAPI_whitelistRemoveHandler(t *testing.T) {
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
			expectedError:  "failed to remove from whitelist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок storage.
			memStorage := memorystorage.New()
			memStorage.Add(context.Background(), models.IPList{Subnet: "100.101.102.103/31", IsWhite: true})
			memStorage.Add(context.Background(), models.IPList{Subnet: "192.168.1.0/24", IsWhite: true})

			// Создаем API с моком.
			api := &API{
				bucketManager: CreateTestBucketManager(),
				storage:       memStorage,
			}

			// Формируем запрос.
			body := RequestBody(t, tt.requestBody)
			req := httptest.NewRequest("POST", "/whitelist/remove", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Вызываем handler.
			api.whitelistRemoveHandler(w, req)

			// Проверяем статус.
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				// Проверяем тело ответа.
				checkErrorResponse(t, w, tt.expectedError)
			} else {
				var resp SuccessfulResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, "ok", resp.Status) // предполагаем что sendJSON отправляет map[string]string{"status": "ok"}.

				subnets, _ := memStorage.GetAll(context.Background())
				equalIPLists(t, subnets, []models.IPList{
					models.IPList{Subnet: "100.101.102.103/31", IsWhite: false},
				})
			}
		})
	}
}

func TestAPI_whitelistHandler(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectedError  string
		cancel         bool
	}{
		{
			name:           "succsessful",
			expectedStatus: http.StatusOK,
			expectedError:  "",
			cancel:         false,
		},
		{
			name:           "server send error",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "failed to get whitelist",
			cancel:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок storage.
			memStorage := memorystorage.New()
			memStorage.Add(context.Background(), models.IPList{Subnet: "100.101.102.103/31", IsWhite: true})
			memStorage.Add(context.Background(), models.IPList{Subnet: "110.111.112.113/31", IsWhite: false})
			memStorage.Add(context.Background(), models.IPList{Subnet: "192.168.1.0/24", IsWhite: true})

			// Создаем API с моком.
			api := &API{
				bucketManager: CreateTestBucketManager(),
				storage:       memStorage,
			}

			// Формируем запрос.
			ctx1, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.cancel {
				cancel() // run cancel to cause server error.
			}

			req := httptest.NewRequestWithContext(ctx1, "GET", "/whitelist", nil)
			w := httptest.NewRecorder()

			req.Context()

			// Вызываем handler.
			api.whitelistHandler(w, req)

			// Проверяем статус.
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				// Проверяем тело ответа.
				checkErrorResponse(t, w, tt.expectedError)
			} else {
				var resp SuccessfulResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)

				subnets, _ := memStorage.GetAll(context.Background())
				equalIPLists(t, subnets, []models.IPList{
					models.IPList{Subnet: "100.101.102.103/31", IsWhite: true},
					models.IPList{Subnet: "192.168.1.0/24", IsWhite: true},
				})
			}
		})
	}
}

func TestAPI_checkHandler(t *testing.T) {
	tests := []struct {
		name           string
		cancel         bool
		requestBody    map[string]string
		expectedStatus int
		expectedOK     bool
		expectedError  string
	}{
		{
			name:   "успешная проверка - IP не в списках, лимиты не превышены",
			cancel: false,
			requestBody: map[string]string{
				"login":    "user0",
				"password": "password0",
				"ip":       "100.101.102.103",
			},
			expectedStatus: http.StatusOK,
			expectedOK:     true,
			expectedError:  "",
		},
		{
			name:   "not authorized from black list",
			cancel: false,
			requestBody: map[string]string{
				"login":    "user1",
				"password": "password1",
				"ip":       "110.111.112.103",
			},
			expectedStatus: http.StatusOK,
			expectedOK:     false,
			expectedError:  "",
		},
		{
			name:   "failed because context ",
			cancel: true,
			requestBody: map[string]string{
				"login":    "user0",
				"password": "password0",
				"ip":       "100.101.102.103",
			},
			expectedStatus: http.StatusInternalServerError,
			expectedOK:     false,
			expectedError:  "failed to check IP authorization",
		},
		{
			name:   "failed missed login",
			cancel: false,
			requestBody: map[string]string{
				"login":    "",
				"password": "password0",
				"ip":       "100.101.102.103",
			},
			expectedStatus: http.StatusBadRequest,
			expectedOK:     false,
			expectedError:  "login, password and ip are required",
		},
		{
			name:           "missed request",
			cancel:         false,
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedOK:     false,
			expectedError:  "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок storage.
			memStorage := memorystorage.New()
			memStorage.Add(context.Background(), models.IPList{Subnet: "100.101.102.103/24", IsWhite: true})
			memStorage.Add(context.Background(), models.IPList{Subnet: "110.111.112.113/24", IsWhite: false})
			memStorage.Add(context.Background(), models.IPList{Subnet: "192.168.1.0/24", IsWhite: true})

			bm := bucket.NewBucketManager(&bucket.Config{
				LoginRate:    10,
				PasswordRate: 100,
				IPRate:       1000,
			})

			// Создаем API с моком.
			api := &API{
				bucketManager: bm,
				storage:       memStorage,
			}

			ctx1, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.cancel {
				cancel() // run cancel to cause server error.
			}

			// Формируем запрос.
			body, _ := json.Marshal(tt.requestBody)
			if tt.requestBody == nil {
				body = nil
			}
			req := httptest.NewRequestWithContext(ctx1, "POST", "/check", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Вызываем handler.
			api.checkHandler(w, req)

			// Проверяем статус.
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Проверяем ответ.
			if tt.expectedError != "" {
				checkErrorResponse(t, w, tt.expectedError)
			} else {
				var resp CheckResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedOK, resp.OK)
			}
		})
	}
}

func TestAPI_checkHandler_login_bucket_is_full(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        map[string]string
		expectedRunResults []bool
	}{
		{
			name: "последняя проверка не проходит потому что лимиты превышены",
			requestBody: map[string]string{
				"login":    "user0",
				"password": "password0",
				"ip":       "100.101.102.103",
			},
			expectedRunResults: []bool{true, true, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок storage.
			memStorage := memorystorage.New()

			bm := bucket.NewBucketManager(&bucket.Config{
				LoginRate:    2,
				PasswordRate: 100,
				IPRate:       1000,
			})

			// Создаем API с моком.
			api := &API{
				bucketManager: bm,
				storage:       memStorage,
			}

			// Формируем запрос.
			for _, res := range tt.expectedRunResults {
				body, _ := json.Marshal(tt.requestBody)
				req := httptest.NewRequest("POST", "/check", bytes.NewReader(body))
				w := httptest.NewRecorder()

				// Вызываем handler.
				api.checkHandler(w, req)

				// Проверяем статус.
				assert.Equal(t, http.StatusOK, w.Code)

				// Проверяем ответ.
				var resp CheckResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)
				assert.Equal(t, res, resp.OK)
			}
		})
	}
}

func TestAPI_resetHandler(t *testing.T) {
	tests := []struct {
		name               string
		requestBody        map[string]string
		expectedStatus     int
		expectedOK         bool
		expectedError      string
		expectedLoginCount int
		expectedIPCount    int
	}{
		{
			name: "successful reset of login bucket",
			requestBody: map[string]string{
				"login": "user0",
			},
			expectedStatus:     http.StatusOK,
			expectedOK:         true,
			expectedError:      "",
			expectedLoginCount: 1,
			expectedIPCount:    3,
		},
		{
			name: "successful reset of ip bucket",
			requestBody: map[string]string{
				"ip": "150.151.152.153",
			},
			expectedStatus:     http.StatusOK,
			expectedOK:         true,
			expectedError:      "",
			expectedLoginCount: 2,
			expectedIPCount:    2,
		},
		{
			name: "wrong request",
			requestBody: map[string]string{
				"ip0": "150.151.152.153",
			},
			expectedStatus: http.StatusBadRequest,
			expectedOK:     true,
			expectedError:  "login or ip required",
		},
		{
			name:           "missed request",
			requestBody:    nil,
			expectedStatus: http.StatusBadRequest,
			expectedOK:     false,
			expectedError:  "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок storage.
			memStorage := memorystorage.New()

			bm := bucket.NewBucketManager(&bucket.Config{
				LoginRate:    10,
				PasswordRate: 100,
				IPRate:       1000,
			})

			// Создаем API с моком.
			api := &API{
				bucketManager: bm,
				storage:       memStorage,
			}

			// add tokens to all buckets.
			_ = bm.CheckAuth("user0", "password0", "100.101.102.103")
			_ = bm.CheckAuth("user0", "password0", "200.201.202.203")
			_ = bm.CheckAuth("user1", "password1", "150.151.152.153")

			bucketStats := bm.Stats()

			require.Equal(t, 2, bucketStats["login"])
			require.Equal(t, 2, bucketStats["password"])
			require.Equal(t, 3, bucketStats["ip"])

			// Формируем запрос.
			body, _ := json.Marshal(tt.requestBody)
			if tt.requestBody == nil {
				body = nil
			}
			req := httptest.NewRequest("POST", "/reset", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Вызываем handler.
			api.resetHandler(w, req)

			// Проверяем статус.
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Проверяем ответ.
			if tt.expectedError != "" {
				checkErrorResponse(t, w, tt.expectedError)
			} else {
				var resp SuccessfulResponse
				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)

				bucketStats := bm.Stats()
				assert.Equal(t, "ok", resp.Status) // предполагаем что sendJSON отправляет map[string]string{"status": "ok"}.
				assert.Equal(t, tt.expectedLoginCount, bucketStats["login"])
				assert.Equal(t, 2, bucketStats["password"])
				assert.Equal(t, tt.expectedIPCount, bucketStats["ip"])
			}
		})
	}
}
