package main

import (
    "bytes"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
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