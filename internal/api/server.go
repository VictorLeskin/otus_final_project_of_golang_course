package api

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Server представляет HTTP сервер API.
type Server struct {
	srv *http.Server
}

// NewServer создает новый HTTP сервер.
// Timeout'ы:
// ReadTimeout — время на чтение запроса.
// WriteTimeout — время на отправку ответа.
// IdleTimeout — время удержания keep-alive соединения.
func NewServer(addr string, handler http.Handler) *Server {
	return &Server{
		srv: &http.Server{
			Addr:         addr,
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// Start запускает сервер.
func (s *Server) Start() error {
	fmt.Printf("Starting API server on %s\n", s.srv.Addr)
	// Ошибка http.ErrServerClosed:
	// Нормальная ошибка при остановке, не нужно на нее ругаться
	if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}
	return nil
}

// Stop gracefully останавливает сервер.
// Graceful shutdown:
// Shutdown(ctx) — ждет завершения текущих запросов.
// Не обрывает активные соединения.
func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("Shutting down API server...")
	return s.srv.Shutdown(ctx)
}
