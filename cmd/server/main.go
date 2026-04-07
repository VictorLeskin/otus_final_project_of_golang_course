package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/api"
	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/bucket"
	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/config"
	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/storage"
	memorystorage "github.com/VictorLeskin/otus_final_project_of_golang_course/internal/storage/memory"
	postgresstorage "github.com/VictorLeskin/otus_final_project_of_golang_course/internal/storage/postgres"
)

func CreatePostgresStorage(cfg config.Config) storage.IPListStorage {
	pgCfg := postgresstorage.Config{
		Host:     cfg.PostgresHost,
		Port:     cfg.PostgresPort,
		User:     cfg.PostgresUser,
		Password: cfg.PostgresPassword,
		Database: cfg.PostgresDB,
		SSLMode:  cfg.PostgresSSLMode,
	}
	store := postgresstorage.New(pgCfg)
	if err := store.Connect(context.Background()); err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}
	defer store.Close(context.Background())
	return store
}

func main() {
	// Загружаем конфиг
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем bucket manager с правильным преобразованием секунд в time.Duration
	bucketConfig := &bucket.Config{
		LoginRate:       cfg.LoginRate,
		PasswordRate:    cfg.PasswordRate,
		IPRate:          cfg.IPRate,
		CleanupInterval: cfg.GetCleanupInterval(),
	}
	bucketManager := bucket.NewBucketManager(bucketConfig)
	defer bucketManager.Stop()

	// Выбираем storage
	var store storage.IPListStorage
	switch cfg.StorageType {
	case "postgres":
		store = CreatePostgresStorage(*cfg)
		log.Println("Using PostgreSQL storage")
	default:
		store = memorystorage.New()
		log.Println("Using Memory storage")
	}

	// Создаем API
	apiServer := api.NewAPI(bucketManager, store)

	// HTTP сервер
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           apiServer.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// Запуск сервера в горутине
	go func() {
		log.Printf("Starting server on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
