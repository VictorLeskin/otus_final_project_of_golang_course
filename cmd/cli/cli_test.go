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

func TestCLI_Check(t *testing.T) {
	// Подменяем сервер
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/check", r.URL.Path)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	cli := NewCLI([]string{"cli", "check", "--login", "user", "--password", "pass", "--ip", "1.1.1.1"})
	cli.getenv = func(key string) string {
		if key == "ANTIBRUTEFORCE_SERVER" {
			return server.URL
		}
		return ""
	}

	out := &bytes.Buffer{}
	cli.stdout = out

	code := cli.Run()

	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "OK")
}

func TestCLI_Check_InvalidArgs(t *testing.T) {
	cli := NewCLI([]string{"cli", "check"}) // missing args
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	cli.stdout = out
	cli.stderr = errOut

	code := cli.Run()

	assert.Equal(t, 1, code)
	assert.Contains(t, errOut.String(), "required")
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

			fs := cli.parseSubnetCommand("whitelist add")

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

func TestCLI_whitelistAdd_Success(t *testing.T) {
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

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем CLI с аргументами
	cli := NewCLI([]string{"192.168.1.0/24"})
	cli.server = server.URL
	out := &bytes.Buffer{}
	cli.stdout = out

	// Вызываем функцию
	code := cli.whitelistAdd()

	// Проверяем результат
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "Added 192.168.1.0/24 to whitelist")
}

func TestCLI_whitelistAdd_MissedSubnet(t *testing.T) {
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

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем CLI с аргументами
	cli := NewCLI([]string{})
	cli.server = server.URL
	out := &bytes.Buffer{}
	cli.stderr = out

	// Вызываем функцию
	code := cli.whitelistAdd()

	// Проверяем результат
	assert.Equal(t, 1, code)
	assert.Contains(t, out.String(), "Usage: cli whitelist add <subnet>")
}

func TestCLI_whitelistAdd_NoSuccess(t *testing.T) {
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

		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	// Создаем CLI с аргументами
	cli := NewCLI([]string{"192.168.1.0/24"})
	cli.server = server.URL
	out := &bytes.Buffer{}
	cli.stderr = out

	// Вызываем функцию
	code := cli.whitelistAdd()

	// Проверяем результат
	assert.Equal(t, 1, code)
	assert.Contains(t, out.String(), "Server error: 400 Bad Request")
}

func TestCLI_whitelistAdd_ServerError(t *testing.T) {
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

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Создаем CLI с аргументами
	cli := NewCLI([]string{"192.168.1.0/24"})
	cli.server = server.URL
	server.Close()
	out := &bytes.Buffer{}
	cli.stderr = out

	// Вызываем функцию
	code := cli.whitelistAdd()

	// Проверяем результат
	assert.Equal(t, 1, code)
	assert.Contains(t, out.String(), "Error: Post ")
}

func TestCLI_whitelistAdd_WithServerFlag(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		assert.Equal(t, "/whitelist/add", r.URL.Path)

		var req map[string]string
		json.NewDecoder(r.Body).Decode(&req)
		assert.Equal(t, "192.168.1.0/24", req["subnet"])

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cli := NewCLI([]string{"cli", "whitelist", "add", "--server", server.URL, "192.168.1.0/24"})
	out := &bytes.Buffer{}
	cli.stdout = out

	code := cli.Run()

	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "Added")
	assert.Equal(t, 1, callCount)
}
