package postgresstorage

import (
	"context"
	"testing"

	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/models"

	"github.com/stretchr/testify/require"
)

func TestPostgresStorage_Integration(t *testing.T) {
	// Пропускаем если тесты запускаются в коротком режиме.
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Конфигурация для тестов.
	cfg := Config{
		Host:     "localhost",
		Port:     5432,
		Database: "ip_lists",
		SSLMode:  "disable",
	}

	// Создаем хранилище.
	store := New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store.Connect(ctx)
	defer store.Close(ctx)

	// Очищаем тестовые данные перед тестом.
	cleanupTestData(t, store)

	t.Run("Complete event lifecycle test", func(t *testing.T) {
		// Создаем события.
		ipList := models.IPList{
			Subnet:  "209.85.233.139/24",
			IsWhite: models.Black,
		}

		// 1. Добавляем первое subnet.
		err := store.Add(ctx, ipList)
		require.NoError(t, err, "Failed to create IP List")
	})
}

// Вспомогательная функция для очистки тестовых данных.
func cleanupTestData(t *testing.T, store *PostgresStorage) {
	t.Helper()
	ctx := context.Background()
	_, err := store.db.ExecContext(ctx, "DELETE FROM ip_lists WHERE subnet LIKE '209.85.233.%'")
	require.NoError(t, err, "Failed to clean test data")
}
