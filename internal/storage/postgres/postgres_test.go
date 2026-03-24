package postgresstorage

import (
	"context"
	"testing"

	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresStorage_AddAndGetAll(t *testing.T) {
	// Пропускаем если тесты запускаются в коротком режиме
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Конфигурация для тестов
	cfg := Config{
		Host:     "localhost",
		Port:     5432,
		Database: "ip_list_test",
		User:     "postgres",
		Password: "123456",
		SSLMode:  "disable",
	}

	// Создаем хранилище
	store := New(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Подключаемся к БД
	err := store.Connect(ctx)
	require.NoError(t, err)
	defer store.Close(ctx)

	// Создаем БД ip_lists
	_, err = store.db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS ip_lists (
            id BIGSERIAL PRIMARY KEY,
            subnet TEXT NOT NULL,
            list_type BOOLEAN NOT NULL,
            created_at TIMESTAMP DEFAULT NOW(),
            UNIQUE(subnet, list_type)
        )
    `)
	require.NoError(t, err, "Failed to create table")

	// Очищаем тестовые данные перед тестом
	cleanupTestData(t, store)

	t.Run("Add and GetAll", func(t *testing.T) {
		// Подсети для теста
		subnet30 := "173.194.221.138/30"
		subnet31 := "173.194.221.138/31"

		// Добавляем первую подсеть в белый список
		ipList1 := models.IPList{
			Subnet:  subnet30,
			IsWhite: models.White,
		}
		err := store.Add(ctx, ipList1)
		require.NoError(t, err, "Failed to add subnet %s", subnet30)

		// Добавляем вторую подсеть в белый список
		ipList2 := models.IPList{
			Subnet:  subnet31,
			IsWhite: models.Black,
		}
		err = store.Add(ctx, ipList2)
		require.NoError(t, err, "Failed to add subnet %s", subnet31)

		// Получаем все записи
		items, err := store.GetAll(ctx)
		require.NoError(t, err)

		// Проверяем количество
		require.Len(t, items, 2, "Expected 2 items")

		// Проверяем содержимое (порядок не важен)
		subnets := make(map[string]bool)
		for _, item := range items {
			subnets[item.Subnet] = bool(item.IsWhite)
		}

		shouldBe := map[string]bool{
			subnet30: bool(models.White),
			subnet31: bool(models.Black),
		}

		assert.Equal(t, shouldBe, subnets, "List should be equal")
	})
}

// Вспомогательная функция для очистки тестовых данных
func cleanupTestData(t *testing.T, store *PostgresStorage) {
	t.Helper()
	ctx := context.Background()

	// Удаляем тестовые подсети
	_, err := store.db.ExecContext(ctx,
		"DELETE FROM ip_lists WHERE subnet LIKE '173.194.221.%'")
	require.NoError(t, err, "Failed to clean test data")
}
