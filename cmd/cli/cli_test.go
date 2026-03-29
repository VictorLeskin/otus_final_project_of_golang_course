package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_runServerCommand(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedCode     int
		expectedOutput   string
		expectedError    string
		ServerStatusCode int
	}{
		{
			name:             "run reset command",
			args:             []string{"reset", "login", "google_admin"},
			expectedCode:     0,
			expectedOutput:   "Reset successful for login: google_admin\n",
			expectedError:    "",
			ServerStatusCode: http.StatusOK,
		},
		{
			name:             "run whitelist command",
			args:             []string{"whitelist", "add", "192.168.1.0/24"},
			expectedCode:     0,
			expectedOutput:   "Added 192.168.1.0/24 to whitelist\n",
			expectedError:    "",
			ServerStatusCode: http.StatusOK,
		},
		{
			name:             "run blacklist command",
			args:             []string{"blacklist", "add", "192.168.1.0/24"},
			expectedCode:     0,
			expectedOutput:   "Added 192.168.1.0/24 to blacklist\n",
			expectedError:    "",
			ServerStatusCode: http.StatusOK,
		},
		{
			name:             "run check command",
			args:             []string{"check", "--login", "me", "--password", "qwerty", "--ip", "200.201.202.203"},
			expectedCode:     0,
			expectedOutput:   "OK: allowed\n",
			expectedError:    "",
			ServerStatusCode: http.StatusOK,
		},
		{
			name:             "bad request at run reset command",
			args:             []string{"reset", "login", "google_admin"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Server error: 400 Bad Request\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "bad request at run whitelist command",
			args:             []string{"whitelist", "add", "192.168.1.0/24"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Server error: 400 Bad Request\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "bad request at run blacklist command",
			args:             []string{"blacklist", "add", "192.168.1.0/24"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Server error: 400 Bad Request\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "bad request at run check command",
			args:             []string{"check", "--login", "me", "--password", "qwerty", "--ip", "ABCD.201.202.203"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "invalid ip address\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "no parameters",
			args:             []string{},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Wrong command line parameters\n",
			ServerStatusCode: http.StatusOK,
		},
		{
			name:             "bad command",
			args:             []string{"a_reset", "login", "google_admin"},
			expectedCode:     1,
			expectedOutput:   "Anti-BruteForce CLI\n\nUsage:\n",
			expectedError:    "",
			ServerStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				// Проверяем тело запроса
				var req map[string]string
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)

				if tt.name == "run check command" {
					response := map[string]bool{
						"result": true,
					}
					err := json.NewEncoder(w).Encode(response)
					require.NoError(t, err)
				}
				if tt.name == "bad request at run check command" {
					response := CheckResult{
						Result: false,
						Error:  "invalid ip address",
					}
					err := json.NewEncoder(w).Encode(response)
					require.NoError(t, err)
				}

				w.WriteHeader(tt.ServerStatusCode)

			}))
			defer server.Close()

			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)
			cli.server = server.URL
			stdout := &bytes.Buffer{}
			cli.stdout = stdout
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.runServerCommand()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			assert.Contains(t, stdout.String(), tt.expectedOutput)
			assert.Contains(t, stderr.String(), tt.expectedError)
		})
	}
}

func TestCLI_runReset(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		expectedCode       int
		expectedOutput     string
		expectedError      string
		ServerStatusCode   int
		emulateServerError bool
	}{
		{
			name:               "reset login",
			args:               []string{"login", "google_admin"},
			expectedCode:       0,
			expectedOutput:     "Reset successful for login: google_admin\n",
			expectedError:      "",
			ServerStatusCode:   http.StatusOK,
			emulateServerError: false,
		},
		{
			name:               "reset ip",
			args:               []string{"ip", "10.11.12.13"},
			expectedCode:       0,
			expectedOutput:     "Reset successful for ip: 10.11.12.13\n",
			expectedError:      "",
			ServerStatusCode:   http.StatusOK,
			emulateServerError: false,
		},
		{
			name:             "failed reset login",
			args:             []string{"login", "google_admin"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Server error: 400 Bad Request\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "failed reset ip",
			args:             []string{"ip", "10.11.12.13"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Server error: 400 Bad Request\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "missed command line parameters",
			args:             []string{},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Wrong 'cli reset' command line parameters\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "wrong command",
			args:             []string{"klogin", "google_admin"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Unknown reset command: klogin\n",
			ServerStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Проверяем путь и метод
				assert.Equal(t, "/reset", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				// Проверяем тело запроса
				var req map[string]string
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)

				w.WriteHeader(tt.ServerStatusCode)
			}))
			defer server.Close()

			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)
			cli.server = server.URL
			if tt.emulateServerError {
				server.Close()
			}
			stdout := &bytes.Buffer{}
			cli.stdout = stdout
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.runReset()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			assert.Contains(t, stdout.String(), tt.expectedOutput)
			assert.Contains(t, stderr.String(), tt.expectedError)
		})
	}
}

func TestCLI_Check_runWhitelist(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedCode     int
		expectedOutput   string
		expectedError    string
		ServerStatusCode int
		listOnServer     []string
	}{
		{
			name:             "add subnet to whitelist",
			args:             []string{"add", "192.168.1.0/24"},
			expectedCode:     0,
			expectedOutput:   "Added 192.168.1.0/24 to whitelist\n",
			expectedError:    "",
			ServerStatusCode: http.StatusOK,
		},
		{
			name:             "remove subnet from whitelist",
			args:             []string{"remove", "192.168.1.0/24"},
			expectedCode:     0,
			expectedOutput:   "Removed 192.168.1.0/24 from whitelist\n",
			expectedError:    "",
			ServerStatusCode: http.StatusOK,
		},
		{
			name:             "get whitelist",
			listOnServer:     []string{"192.168.1.0/24", "10.0.0.0/8"},
			ServerStatusCode: http.StatusOK,
			args:             []string{"list"},
			expectedCode:     0,
			expectedOutput:   "Whitelist: [192.168.1.0/24 10.0.0.0/8]\n",
			expectedError:    "",
		},
		{
			name:             "failed adding subnet to whitelist",
			args:             []string{"add", "192.168.1.0/24"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Server error: 400 Bad Request\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "failed removing subnet to whitelist",
			args:             []string{"remove", "192.168.1.0/24"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Server error: 400 Bad Request\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "failed get list request",
			args:             []string{"list"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Server error: 400 Bad Request\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "missed command line parameters",
			args:             []string{},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Wrong 'cli whitelist' command line parameters\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "wrong command",
			args:             []string{"kadd", "192.168.1.0/24"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Unknown whitelist command: kadd\n",
			ServerStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.ServerStatusCode)
				if tt.ServerStatusCode == http.StatusBadRequest {
					return
				}
				if len(tt.listOnServer) != 0 {
					response := map[string][]string{
						"whitelist": tt.listOnServer,
					}
					err := json.NewEncoder(w).Encode(response)
					require.NoError(t, err)
				} else {
					// Проверяем тело запроса
					var req map[string]string
					err := json.NewDecoder(r.Body).Decode(&req)
					require.NoError(t, err)
				}
			}))
			defer server.Close()

			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)
			cli.server = server.URL
			stdout := &bytes.Buffer{}
			cli.stdout = stdout
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.runWhitelist()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			assert.Contains(t, stdout.String(), tt.expectedOutput)
			assert.Contains(t, stderr.String(), tt.expectedError)
		})
	}
}

func TestCLI_Check_runBlacklist(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedCode     int
		expectedOutput   string
		expectedError    string
		ServerStatusCode int
		listOnServer     []string
	}{
		{
			name:             "add subnet to blacklist",
			args:             []string{"add", "192.168.1.0/24"},
			expectedCode:     0,
			expectedOutput:   "Added 192.168.1.0/24 to blacklist\n",
			expectedError:    "",
			ServerStatusCode: http.StatusOK,
		},
		{
			name:             "remove subnet from blacklist",
			args:             []string{"remove", "192.168.1.0/24"},
			expectedCode:     0,
			expectedOutput:   "Removed 192.168.1.0/24 from blacklist\n",
			expectedError:    "",
			ServerStatusCode: http.StatusOK,
		},
		{
			name:             "get blacklist",
			listOnServer:     []string{"192.168.1.0/24", "10.0.0.0/8"},
			ServerStatusCode: http.StatusOK,
			args:             []string{"list"},
			expectedCode:     0,
			expectedOutput:   "Whitelist: [192.168.1.0/24 10.0.0.0/8]\n",
			expectedError:    "",
		},
		{
			name:             "failed adding subnet to blacklist",
			args:             []string{"add", "192.168.1.0/24"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Server error: 400 Bad Request\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "failed removing subnet to blacklist",
			args:             []string{"remove", "192.168.1.0/24"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Server error: 400 Bad Request\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "failed get list request",
			args:             []string{"list"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Server error: 400 Bad Request\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "missed command line parameters",
			args:             []string{},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Wrong 'cli blacklist' command line parameters\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:             "wrong command",
			args:             []string{"kadd", "192.168.1.0/24"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Unknown blacklist command: kadd\n",
			ServerStatusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.ServerStatusCode)
				if tt.ServerStatusCode == http.StatusBadRequest {
					return
				}
				if len(tt.listOnServer) != 0 {
					response := map[string][]string{
						"blacklist": tt.listOnServer,
					}
					err := json.NewEncoder(w).Encode(response)
					require.NoError(t, err)
				} else {
					// Проверяем тело запроса
					var req map[string]string
					err := json.NewDecoder(r.Body).Decode(&req)
					require.NoError(t, err)
				}
			}))
			defer server.Close()

			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)
			cli.server = server.URL
			stdout := &bytes.Buffer{}
			cli.stdout = stdout
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.runBlacklist()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			assert.Contains(t, stdout.String(), tt.expectedOutput)
			assert.Contains(t, stderr.String(), tt.expectedError)
		})
	}
}

func TestCLI_parseSubnetCommand(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantSubnet string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:       "успешный парсинг с subnet",
			args:       []string{"cli", "whitelist", "add", "192.168.1.0/24"},
			wantSubnet: "192.168.1.0/24",
			wantErr:    false,
		},
		{
			name:       "отсутствует subnet",
			args:       []string{"cli", "whitelist", "add"},
			wantErr:    true,
			wantErrMsg: "Usage: cli whitelist add <subnet>",
		},
		{
			name:       "неизвестный флаг",
			args:       []string{"cli", "whitelist", "add", "--unknown", "value", "192.168.1.0/24"},
			wantErrMsg: "flag provided but not defined: -unknown",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := NewCLI(tt.args)
			errOut := &bytes.Buffer{}
			cli.args = tt.args[3:]
			cli.stderr = errOut

			fs := cli.parseSubnetCommand("whitelist add", "subnet")

			if tt.wantErr {
				assert.Nil(t, fs)
				assert.Contains(t, errOut.String(), tt.wantErrMsg)
			} else {
				assert.NotNil(t, fs)
				args := fs.Args()
				assert.Len(t, args, 1)
				assert.Equal(t, tt.wantSubnet, args[0])
			}
		})
	}
}

func TestCLI_processServerStatus(t *testing.T) {
	{
		cli := NewCLI([]string{})
		stderr := &bytes.Buffer{}
		cli.stderr = stderr

		resp := &http.Response{
			Status:     "200 OK",
			StatusCode: 200,
		}
		i := 0
		f := func() {
			i++
		}

		ret := cli.processServerStatus(resp, f)
		assert.Equal(t, 0, ret)
		assert.Equal(t, 1, i)
		assert.Equal(t, stderr.String(), "")
	}

	{
		cli := NewCLI([]string{})
		stderr := &bytes.Buffer{}
		cli.stderr = stderr

		resp := &http.Response{
			Status:     "NOK",
			StatusCode: 429,
		}
		i := 0
		f := func() {
			i++
		}

		ret := cli.processServerStatus(resp, f)
		assert.Equal(t, 1, ret)
		assert.Equal(t, 0, i)
		assert.Equal(t, stderr.String(), "Server error: NOK\n")
	}
}

func TestCLI_whitelistAdd(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		expectedCode       int
		expectedOutput     string
		expectedError      string
		ServerStatusCode   int
		emulateServerError bool
	}{
		{
			name:               "successful",
			args:               []string{"192.168.1.0/24"},
			expectedCode:       0,
			expectedOutput:     "Added 192.168.1.0/24 to whitelist\n",
			expectedError:      "",
			ServerStatusCode:   http.StatusOK,
			emulateServerError: false,
		},
		{
			name:               "missed subnet",
			args:               []string{},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Usage: cli whitelist add <subnet>",
			ServerStatusCode:   http.StatusOK,
			emulateServerError: false,
		},
		{
			name:               "server return bad request",
			args:               []string{"192.168.1.0/24"},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Server error: 400 Bad Request",
			ServerStatusCode:   http.StatusBadRequest,
			emulateServerError: false,
		},
		{
			name:               "internal server error",
			args:               []string{"192.168.1.0/24"},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Error: Post",
			ServerStatusCode:   http.StatusInternalServerError,
			emulateServerError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Проверяем путь и метод
				assert.Equal(t, "/whitelist/add", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				// Проверяем тело запроса
				var req map[string]string
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, "192.168.1.0/24", req["subnet"])

				w.WriteHeader(tt.ServerStatusCode)
			}))
			defer server.Close()

			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)
			cli.server = server.URL
			if tt.emulateServerError {
				server.Close()
			}
			stdout := &bytes.Buffer{}
			cli.stdout = stdout
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.whitelistAdd()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			assert.Contains(t, stdout.String(), tt.expectedOutput)
			assert.Contains(t, stderr.String(), tt.expectedError)
		})
	}
}

func TestCLI_whitelistRemove(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		expectedCode       int
		expectedOutput     string
		expectedError      string
		ServerStatusCode   int
		emulateServerError bool
	}{
		{
			name:               "successful",
			args:               []string{"192.168.1.0/24"},
			expectedCode:       0,
			expectedOutput:     "Removed 192.168.1.0/24 from whitelist\n",
			expectedError:      "",
			ServerStatusCode:   http.StatusOK,
			emulateServerError: false,
		},
		{
			name:               "missed subnet",
			args:               []string{},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Usage: cli whitelist remove <subnet>",
			ServerStatusCode:   http.StatusOK,
			emulateServerError: false,
		},
		{
			name:               "server return bad request",
			args:               []string{"192.168.1.0/24"},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Server error: 400 Bad Request",
			ServerStatusCode:   http.StatusBadRequest,
			emulateServerError: false,
		},
		{
			name:               "internal server error",
			args:               []string{"192.168.1.0/24"},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Error: Post",
			ServerStatusCode:   http.StatusInternalServerError,
			emulateServerError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Проверяем путь и метод
				assert.Equal(t, "/whitelist/remove", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				// Проверяем тело запроса
				var req map[string]string
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, "192.168.1.0/24", req["subnet"])

				w.WriteHeader(tt.ServerStatusCode)
			}))
			defer server.Close()

			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)
			cli.server = server.URL
			if tt.emulateServerError {
				server.Close()
			}
			stdout := &bytes.Buffer{}
			cli.stdout = stdout
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.whitelistRemove()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			assert.Contains(t, stdout.String(), tt.expectedOutput)
			assert.Contains(t, stderr.String(), tt.expectedError)
		})
	}
}

func TestCLI_whitelistList(t *testing.T) {
	tests := []struct {
		name                  string
		listOnServer          []string
		generatedJsonOnServer string
		ServerStatusCode      int
		args                  []string
		expectedCode          int
		expectedOutput        string
		expectedError         string
		emulateServerError    bool
	}{
		{
			name:               "get two subnets",
			listOnServer:       []string{"192.168.1.0/24", "10.0.0.0/8"},
			ServerStatusCode:   http.StatusOK,
			args:               []string{},
			expectedCode:       0,
			expectedOutput:     "Whitelist: [192.168.1.0/24 10.0.0.0/8]\n",
			expectedError:      "",
			emulateServerError: false,
		},
		{
			name:               "get empty subnets",
			listOnServer:       []string{},
			ServerStatusCode:   http.StatusOK,
			args:               []string{},
			expectedCode:       0,
			expectedOutput:     "Whitelist: []\n",
			expectedError:      "",
			emulateServerError: false,
		},
		{
			name:               "server error",
			listOnServer:       []string{"192.168.1.0/24", "10.0.0.0/8"},
			ServerStatusCode:   http.StatusOK,
			args:               []string{},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Error: Get",
			emulateServerError: true,
		},
		{
			name:                  "get send wrong JSON",
			listOnServer:          []string{},
			ServerStatusCode:      http.StatusOK,
			generatedJsonOnServer: "{invalid json}",
			args:                  []string{},
			expectedCode:          1,
			expectedOutput:        "",
			expectedError:         "Error parsing response:",
			emulateServerError:    false,
		},
		{
			name:               "internal server error",
			listOnServer:       []string{"192.168.1.0/24", "10.0.0.0/8"},
			ServerStatusCode:   http.StatusInternalServerError,
			args:               []string{},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Server error: 500 Internal Server Error\n",
			emulateServerError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.ServerStatusCode)
				if tt.generatedJsonOnServer == "" {
					response := map[string][]string{
						"whitelist": tt.listOnServer,
					}
					err := json.NewEncoder(w).Encode(response)
					require.NoError(t, err)
				} else {
					w.Write([]byte(tt.generatedJsonOnServer))
				}
			}))
			defer server.Close()

			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)
			cli.server = server.URL
			if tt.emulateServerError {
				server.Close()
			}
			stdout := &bytes.Buffer{}
			cli.stdout = stdout
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.whitelistList()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			assert.Contains(t, stdout.String(), tt.expectedOutput)
			assert.Contains(t, stderr.String(), tt.expectedError)
		})
	}
}

func TestCLI_resetLogin(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		expectedCode       int
		expectedOutput     string
		expectedError      string
		ServerStatusCode   int
		emulateServerError bool
	}{
		{
			name:               "successful",
			args:               []string{"google_admin"},
			expectedCode:       0,
			expectedOutput:     "Reset successful for login: google_admin\n",
			expectedError:      "",
			ServerStatusCode:   http.StatusOK,
			emulateServerError: false,
		},
		{
			name:               "missed login",
			args:               []string{},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Usage: cli reset login <login>\n",
			ServerStatusCode:   http.StatusOK,
			emulateServerError: false,
		},
		{
			name:               "server return bad request",
			args:               []string{"google_admin"},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Server error: 400 Bad Request",
			ServerStatusCode:   http.StatusBadRequest,
			emulateServerError: false,
		},
		{
			name:               "server error",
			args:               []string{"google_admin"},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Error: Post",
			ServerStatusCode:   http.StatusBadRequest,
			emulateServerError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Проверяем путь и метод
				assert.Equal(t, "/reset", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				// Проверяем тело запроса
				var req map[string]string
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, "google_admin", req["login"])

				w.WriteHeader(tt.ServerStatusCode)
			}))
			defer server.Close()

			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)
			cli.server = server.URL
			if tt.emulateServerError {
				server.Close()
			}
			stdout := &bytes.Buffer{}
			cli.stdout = stdout
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.resetLogin()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			if tt.expectedOutput != "" {
				assert.Equal(t, stdout.String(), tt.expectedOutput)
			}
			assert.Contains(t, stderr.String(), tt.expectedError)
		})
	}
}

func TestCLI_resetIP(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		expectedCode       int
		expectedOutput     string
		expectedError      string
		ServerStatusCode   int
		emulateServerError bool
	}{
		{
			name:               "successful",
			args:               []string{"192.201.202.203"},
			expectedCode:       0,
			expectedOutput:     "Reset successful for ip: 192.201.202.203\n",
			expectedError:      "",
			ServerStatusCode:   http.StatusOK,
			emulateServerError: false,
		},
		{
			name:               "missed ip",
			args:               []string{},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Usage: cli reset ip <ip>\n",
			ServerStatusCode:   http.StatusOK,
			emulateServerError: false,
		},
		{
			name:               "server return bad request",
			args:               []string{"192.201.202.203"},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Server error: 400 Bad Request",
			ServerStatusCode:   http.StatusBadRequest,
			emulateServerError: false,
		},
		{
			name:               "server error",
			args:               []string{"192.201.202.203"},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Error: Post",
			ServerStatusCode:   http.StatusBadRequest,
			emulateServerError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Проверяем путь и метод
				assert.Equal(t, "/reset", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				// Проверяем тело запроса
				var req map[string]string
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, "192.201.202.203", req["ip"])

				w.WriteHeader(tt.ServerStatusCode)
			}))
			defer server.Close()

			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)
			cli.server = server.URL
			if tt.emulateServerError {
				server.Close()
			}
			stdout := &bytes.Buffer{}
			cli.stdout = stdout
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.resetIP()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			if tt.expectedOutput != "" {
				assert.Equal(t, stdout.String(), tt.expectedOutput)
			}
			assert.Contains(t, stderr.String(), tt.expectedError)
		})
	}
}

func TestCLI_parseCheckCommand(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedCode   int
		expectedResult []string
		expectedError  string
	}{
		{
			name:           "successful",
			args:           []string{"--login", "me", "--password", "qwerty", "--ip", "200.201.202.203"},
			expectedCode:   0,
			expectedResult: []string{"me", "qwerty", "200.201.202.203"},
			expectedError:  "",
		},
		{
			name:           "successful_shifted",
			args:           []string{"--password", "qwerty", "--ip", "200.201.202.203", "--login", "me"},
			expectedCode:   0,
			expectedResult: []string{"me", "qwerty", "200.201.202.203"},
			expectedError:  "",
		},
		{
			name:           "missed login",
			args:           []string{"--password", "qwerty", "--ip", "200.201.202.203"},
			expectedCode:   1,
			expectedResult: []string{},
			expectedError:  "Usage: cli check --login <login> --password <password> --ip <ip>\n",
		},
		{
			name:           "wrong login flag",
			args:           []string{"--login1", "me", "--password", "qwerty", "--ip", "200.201.202.203"},
			expectedCode:   1,
			expectedResult: []string{},
			expectedError:  "flag provided but not defined: -login1\n",
		},
		{
			name:           "empty login",
			args:           []string{"--login", "", "--password", "qwerty", "--ip", "200.201.202.203"},
			expectedCode:   1,
			expectedResult: []string{},
			expectedError:  "Error: --login, --password, --ip are required\n",
		},
		{
			name:           "empty password",
			args:           []string{"--login", "me", "--password", "", "--ip", "200.201.202.203"},
			expectedCode:   1,
			expectedResult: []string{},
			expectedError:  "Error: --login, --password, --ip are required\n",
		},
		{
			name:           "empty ip",
			args:           []string{"--login", "me", "--password", "qwerty", "--ip", ""},
			expectedCode:   1,
			expectedResult: []string{},
			expectedError:  "Error: --login, --password, --ip are required\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code, login, password, ip := cli.parseCheckCommand()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			if code == 0 {
				assert.Equal(t, tt.expectedResult[0], *login)
				assert.Equal(t, tt.expectedResult[1], *password)
				assert.Equal(t, tt.expectedResult[2], *ip)
				assert.Equal(t, tt.expectedError, "")
			} else {
				assert.Nil(t, login)
				assert.Nil(t, password)
				assert.Nil(t, ip)
				assert.Contains(t, stderr.String(), tt.expectedError)
			}
		})
	}
}

func TestCLI_runCheck(t *testing.T) {
	tests := []struct {
		name                  string
		listOnServer          []string
		checkResult           bool
		checkError            string
		generatedJsonOnServer string
		ServerStatusCode      int
		args                  []string
		expectedCode          int
		expectedOutput        string
		expectedError         string
		emulateServerError    bool
	}{
		{
			name:               "successful",
			args:               []string{"--login", "me", "--password", "qwerty", "--ip", "200.201.202.203"},
			checkResult:        true,
			checkError:         "",
			ServerStatusCode:   http.StatusOK,
			expectedCode:       0,
			expectedOutput:     "OK: allowed\n",
			expectedError:      "",
			emulateServerError: false,
		},
		{
			name:               "disabled",
			args:               []string{"--login", "me", "--password", "qwerty", "--ip", "200.201.202.203"},
			checkResult:        false,
			checkError:         "",
			ServerStatusCode:   http.StatusOK,
			expectedCode:       1,
			expectedOutput:     "DENIED: brute-force detected\n",
			expectedError:      "",
			emulateServerError: false,
		},
		{
			name:               "error in intput parameters",
			args:               []string{"--login", "me", "--password", "qwerty", "--ip", "ABCD.201.202.203"},
			checkResult:        false,
			checkError:         "invalid ip addresss",
			ServerStatusCode:   http.StatusOK,
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "invalid ip addresss",
			emulateServerError: false,
		},
		{
			name:                  "error in intput parameters",
			args:                  []string{"--login", "me", "--password", "qwerty", "--ip", "ABCD.201.202.203"},
			checkResult:           false,
			checkError:            "invalid ip addresss",
			generatedJsonOnServer: "{invalid json}",
			ServerStatusCode:      http.StatusOK,
			expectedCode:          1,
			expectedOutput:        "",
			expectedError:         "Error parsing response:",
			emulateServerError:    false,
		},
		{
			name:               "server error",
			args:               []string{"--login", "me", "--password", "qwerty", "--ip", "200.201.202.203"},
			checkResult:        true,
			checkError:         "",
			ServerStatusCode:   http.StatusInternalServerError,
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Error: Post",
			emulateServerError: true,
		},
		{
			name:               "wrong input",
			args:               []string{"--login1"},
			checkResult:        true,
			checkError:         "",
			ServerStatusCode:   http.StatusOK,
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Usage: cli check --login <login> --password <password> --ip <ip>\n",
			emulateServerError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.ServerStatusCode)
				if tt.generatedJsonOnServer == "" {
					if tt.checkError != "" {
						response := CheckResult{
							Result: tt.checkResult,
							Error:  tt.checkError,
						}
						err := json.NewEncoder(w).Encode(response)
						require.NoError(t, err)
					} else {
						response := map[string]bool{
							"result": tt.checkResult,
						}
						err := json.NewEncoder(w).Encode(response)
						require.NoError(t, err)
					}
				} else {
					w.Write([]byte(tt.generatedJsonOnServer))
				}
			}))
			defer server.Close()

			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)
			cli.server = server.URL
			if tt.emulateServerError {
				server.Close()
			}
			stdout := &bytes.Buffer{}
			cli.stdout = stdout
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.runCheck()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			assert.Contains(t, stdout.String(), tt.expectedOutput)
			assert.Contains(t, stderr.String(), tt.expectedError)
		})
	}
}

func TestCLI_blacklistAdd(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		expectedCode       int
		expectedOutput     string
		expectedError      string
		ServerStatusCode   int
		emulateServerError bool
	}{
		{
			name:               "successful",
			args:               []string{"192.168.1.0/24"},
			expectedCode:       0,
			expectedOutput:     "Added 192.168.1.0/24 to blacklist\n",
			expectedError:      "",
			ServerStatusCode:   http.StatusOK,
			emulateServerError: false,
		},
		{
			name:               "missed subnet",
			args:               []string{},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Usage: cli blacklist add <subnet>",
			ServerStatusCode:   http.StatusOK,
			emulateServerError: false,
		},
		{
			name:               "server return bad request",
			args:               []string{"192.168.1.0/24"},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Server error: 400 Bad Request",
			ServerStatusCode:   http.StatusBadRequest,
			emulateServerError: false,
		},
		{
			name:               "internal server error",
			args:               []string{"192.168.1.0/24"},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Error: Post",
			ServerStatusCode:   http.StatusInternalServerError,
			emulateServerError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Проверяем путь и метод
				assert.Equal(t, "/blacklist/add", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				// Проверяем тело запроса
				var req map[string]string
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, "192.168.1.0/24", req["subnet"])

				w.WriteHeader(tt.ServerStatusCode)
			}))
			defer server.Close()

			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)
			cli.server = server.URL
			if tt.emulateServerError {
				server.Close()
			}
			stdout := &bytes.Buffer{}
			cli.stdout = stdout
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.blacklistAdd()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			assert.Contains(t, stdout.String(), tt.expectedOutput)
			assert.Contains(t, stderr.String(), tt.expectedError)
		})
	}
}

func TestCLI_blacklistRemove(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		expectedCode       int
		expectedOutput     string
		expectedError      string
		ServerStatusCode   int
		emulateServerError bool
	}{
		{
			name:               "successful",
			args:               []string{"192.168.1.0/24"},
			expectedCode:       0,
			expectedOutput:     "Removed 192.168.1.0/24 from blacklist\n",
			expectedError:      "",
			ServerStatusCode:   http.StatusOK,
			emulateServerError: false,
		},
		{
			name:               "missed subnet",
			args:               []string{},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Usage: cli blacklist remove <subnet>",
			ServerStatusCode:   http.StatusOK,
			emulateServerError: false,
		},
		{
			name:               "server return bad request",
			args:               []string{"192.168.1.0/24"},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Server error: 400 Bad Request",
			ServerStatusCode:   http.StatusBadRequest,
			emulateServerError: false,
		},
		{
			name:               "internal server error",
			args:               []string{"192.168.1.0/24"},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Error: Post",
			ServerStatusCode:   http.StatusInternalServerError,
			emulateServerError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Проверяем путь и метод
				assert.Equal(t, "/blacklist/remove", r.URL.Path)
				assert.Equal(t, "POST", r.Method)

				// Проверяем тело запроса
				var req map[string]string
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				assert.Equal(t, "192.168.1.0/24", req["subnet"])

				w.WriteHeader(tt.ServerStatusCode)
			}))
			defer server.Close()

			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)
			cli.server = server.URL
			if tt.emulateServerError {
				server.Close()
			}
			stdout := &bytes.Buffer{}
			cli.stdout = stdout
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.blacklistRemove()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			assert.Contains(t, stdout.String(), tt.expectedOutput)
			assert.Contains(t, stderr.String(), tt.expectedError)
		})
	}
}

func TestCLI_blacklistList(t *testing.T) {
	tests := []struct {
		name                  string
		listOnServer          []string
		generatedJsonOnServer string
		ServerStatusCode      int
		args                  []string
		expectedCode          int
		expectedOutput        string
		expectedError         string
		emulateServerError    bool
	}{
		{
			name:               "get two subnets",
			listOnServer:       []string{"192.168.1.0/24", "10.0.0.0/8"},
			ServerStatusCode:   http.StatusOK,
			args:               []string{},
			expectedCode:       0,
			expectedOutput:     "Whitelist: [192.168.1.0/24 10.0.0.0/8]\n",
			expectedError:      "",
			emulateServerError: false,
		},
		{
			name:               "get empty subnets",
			listOnServer:       []string{},
			ServerStatusCode:   http.StatusOK,
			args:               []string{},
			expectedCode:       0,
			expectedOutput:     "Whitelist: []\n",
			expectedError:      "",
			emulateServerError: false,
		},
		{
			name:               "server error",
			listOnServer:       []string{"192.168.1.0/24", "10.0.0.0/8"},
			ServerStatusCode:   http.StatusOK,
			args:               []string{},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Error: Get",
			emulateServerError: true,
		},
		{
			name:                  "get send wrong JSON",
			listOnServer:          []string{},
			ServerStatusCode:      http.StatusOK,
			generatedJsonOnServer: "{invalid json}",
			args:                  []string{},
			expectedCode:          1,
			expectedOutput:        "",
			expectedError:         "Error parsing response:",
			emulateServerError:    false,
		},
		{
			name:               "internal server error",
			listOnServer:       []string{"192.168.1.0/24", "10.0.0.0/8"},
			ServerStatusCode:   http.StatusInternalServerError,
			args:               []string{},
			expectedCode:       1,
			expectedOutput:     "",
			expectedError:      "Server error: 500 Internal Server Error\n",
			emulateServerError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.ServerStatusCode)
				if tt.generatedJsonOnServer == "" {
					response := map[string][]string{
						"blacklist": tt.listOnServer,
					}
					err := json.NewEncoder(w).Encode(response)
					require.NoError(t, err)
				} else {
					w.Write([]byte(tt.generatedJsonOnServer))
				}
			}))
			defer server.Close()

			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)
			cli.server = server.URL
			if tt.emulateServerError {
				server.Close()
			}
			stdout := &bytes.Buffer{}
			cli.stdout = stdout
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.blacklistList()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			assert.Contains(t, stdout.String(), tt.expectedOutput)
			assert.Contains(t, stderr.String(), tt.expectedError)
		})
	}
}

func TestCLI_removeServerFlag(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		expectedCode int
		after        []string
	}{
		{
			name:         "remove from end",
			args:         []string{"--server", "https://192.168.1.0", "A", "B"},
			expectedCode: 0,
			after:        []string{"A", "B"},
		},
		{
			name:         "remove from middle",
			args:         []string{"A", "--server", "https://192.168.1.0", "B"},
			expectedCode: 0,
			after:        []string{"A", "B"},
		},
		{
			name:         "remove from end",
			args:         []string{"A", "B", "--server", "https://192.168.1.0"},
			expectedCode: 0,
			after:        []string{"A", "B"},
		},
		{
			name:         "there is not --server",
			args:         []string{"https://192.168.1.0", "A", "B"},
			expectedCode: 1,
			after:        []string{"https://192.168.1.0", "A", "B"},
		},
		{
			name:         "there is not ip address",
			args:         []string{"A", "B", "--server"},
			expectedCode: 1,
			after:        []string{"A", "B", "--server"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)

			// Вызываем функцию
			code := cli.removeServerFlag()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			assert.Equal(t, tt.after, cli.args)
		})
	}
}

func TestCLI_initServer(t *testing.T) {
	tests := []struct {
		name          string
		args          []string
		expectedCode  int
		server        string
		expectedError string
	}{
		{
			name:          "remove from end",
			args:          []string{"--server", "https://192.168.1.0", "A", "B"},
			expectedCode:  0,
			server:        "https://192.168.1.0",
			expectedError: "",
		},
		{
			name:          "empty server",
			args:          []string{"--server", "", "A", "B"},
			expectedCode:  1,
			server:        "",
			expectedError: "--server is required",
		},
		{
			name:          "missed --server",
			args:          []string{"https://192.168.1.0", "A", "B"},
			expectedCode:  1,
			server:        "",
			expectedError: "failed to parse --server flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)

			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.initServer()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			assert.Equal(t, tt.server, cli.server)
		})
	}
}

func TestCLI_Run(t *testing.T) {
	tests := []struct {
		name             string
		args             []string
		expectedCode     int
		expectedOutput   string
		expectedError    string
		ServerStatusCode int
	}{
		{
			name:             "run reset command",
			args:             []string{"cli", "--server", "Insert server address", "reset", "login", "google_admin"},
			expectedCode:     0,
			expectedOutput:   "Reset successful for login: google_admin\n",
			expectedError:    "",
			ServerStatusCode: http.StatusOK,
		},
		{
			name:             "bad request at run reset command",
			args:             []string{"cli", "--server", "Insert server address", "reset", "login", "google_admin"},
			expectedCode:     1,
			expectedOutput:   "",
			expectedError:    "Server error: 400 Bad Request\n",
			ServerStatusCode: http.StatusBadRequest,
		},
		{
			name:           "bad server name",
			args:           []string{"cli", "reset", "login", "google_admin", "--server"},
			expectedCode:   1,
			expectedOutput: "",
			expectedError:  "--server is required\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем тестовый сервер
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

				// Проверяем тело запроса
				var req map[string]string
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)

				if tt.name == "run check command" {
					response := map[string]bool{
						"result": true,
					}
					err := json.NewEncoder(w).Encode(response)
					require.NoError(t, err)
				}
				if tt.name == "bad request at run check command" {
					response := CheckResult{
						Result: false,
						Error:  "invalid ip address",
					}
					err := json.NewEncoder(w).Encode(response)
					require.NoError(t, err)
				}

				w.WriteHeader(tt.ServerStatusCode)

			}))
			defer server.Close()

			for i := range tt.args {
				if tt.args[i] == "Insert server address" {
					tt.args[i] = server.URL
					break
				}
			}

			// Создаем CLI с аргументами
			cli := NewCLI(tt.args)

			stdout := &bytes.Buffer{}
			cli.stdout = stdout
			stderr := &bytes.Buffer{}
			cli.stderr = stderr

			// Вызываем функцию
			code := cli.Run()

			// Проверяем результат
			assert.Equal(t, tt.expectedCode, code)
			assert.Contains(t, stdout.String(), tt.expectedOutput)
			assert.Contains(t, stderr.String(), tt.expectedError)
		})
	}
}
