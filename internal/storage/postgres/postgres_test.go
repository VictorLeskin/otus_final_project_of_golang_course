package postgresstorage

import (
	"context"
	"testing"

	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Подсети для теста they are address of google.com
var subnet30 string = "173.194.221.138/30"
var subnet31 string = "173.194.221.138/31"

func createStore(t *testing.T) *PostgresStorage {
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

	// Подключаемся к БД
	err := store.Connect(context.Background())
	require.NoError(t, err)

	return store
}

func createDB(t *testing.T, ctx context.Context, store *PostgresStorage) {
	// Создаем БД ip_lists
	_, err := store.db.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS ip_lists (
            id BIGSERIAL PRIMARY KEY,
            subnet TEXT NOT NULL,
            list_type BOOLEAN NOT NULL,
            created_at TIMESTAMP DEFAULT NOW(),
            UNIQUE(subnet, list_type)
        )
    `)
	require.NoError(t, err, "Failed to create table")
}

// convert to subnets mask for comparing
func getSubnetsAsMap(items []models.IPList) map[string]bool {
	subnets := make(map[string]bool)
	for _, item := range items {
		subnets[item.Subnet] = bool(item.IsWhite)
	}
	return subnets
}

func (s *PostgresStorage) createIPList(s0 string, c models.ListType) models.IPList {
	return models.IPList{
		Subnet:  s0,
		IsWhite: c,
	}
}

func TestPostgresStorage_Add(t *testing.T) {
	// Пропускаем если тесты запускаются в коротком режиме
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем хранилище и подключаемся к БД
	store := createStore(t)
	defer store.Close(ctx)

	// Создаем БД ip_lists
	createDB(t, ctx, store)

	// Очищаем тестовые данные перед тестом
	cleanupTestData(t, store)

	t.Run("Add and get one subnet", func(t *testing.T) {
		// Подсети для теста
		// Добавляем первую подсеть в белый список
		err := store.Add(ctx, store.createIPList(subnet30, models.White))
		require.NoError(t, err, "Failed to add subnet %s", subnet30)

		// Получаем все записи
		items, err := store.GetAll(ctx)
		require.NoError(t, err)

		assert.Equal(t, getSubnetsAsMap(items), map[string]bool{
			subnet30: bool(models.White),
		})

		// try to add an invalid subnet
		err = store.Add(ctx, store.createIPList("invalid", models.White))
		require.Error(t, err, "Failed to add subnet %s", subnet30)

		// Получаем все записи
		items, err = store.GetAll(ctx)
		require.NoError(t, err)
		assert.Equal(t, getSubnetsAsMap(items), map[string]bool{
			subnet30: bool(models.White),
		})
	})
}

func TestPostgresStorage_AddAndGetAll(t *testing.T) {
	// Пропускаем если тесты запускаются в коротком режиме
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем хранилище и подключаемся к БД
	store := createStore(t)
	defer store.Close(ctx)

	// Создаем БД ip_lists
	createDB(t, ctx, store)

	// Очищаем тестовые данные перед тестом
	cleanupTestData(t, store)

	items, err := store.GetAll(ctx)
	require.NoError(t, err)
	require.Len(t, items, 0)

	t.Run("Add and GetAll", func(t *testing.T) {
		// Подсети для теста
		err := store.Add(ctx, store.createIPList(subnet30, models.White))
		require.NoError(t, err, "Failed to add subnet %s", subnet30)

		// Добавляем вторую подсеть в белый список
		err = store.Add(ctx, store.createIPList(subnet31, models.Black))
		require.NoError(t, err, "Failed to add subnet %s", subnet31)

		// Получаем все записи
		items, err := store.GetAll(ctx)
		require.NoError(t, err)

		// Проверяем количество
		require.Len(t, items, 2, "Expected 2 items")
		assert.Equal(t, getSubnetsAsMap(items), map[string]bool{
			subnet30: bool(models.White),
			subnet31: bool(models.Black),
		})
	})
}

// Вспомогательная функция для очистки тестовых данных
func cleanupTestData(t *testing.T, store *PostgresStorage) {
	t.Helper()
	ctx := context.Background()

	// Удаляем тестовые подсети
	_, err := store.db.ExecContext(ctx,
		"DELETE FROM ip_lists WHERE subnet LIKE '%'")
	require.NoError(t, err, "Failed to clean test data")
}

func TestPostgresStorage_Remove(t *testing.T) {
	// Пропускаем если тесты запускаются в коротком режиме
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем хранилище и подключаемся к БД
	store := createStore(t)
	defer store.Close(ctx)

	// Создаем БД ip_lists
	createDB(t, ctx, store)

	// Очищаем тестовые данные перед тестом
	cleanupTestData(t, store)

	t.Run("Add two subnets and remove both", func(t *testing.T) {
		// Добавляем первую подсеть в белый список
		err := store.Add(ctx, store.createIPList(subnet30, models.White))
		require.NoError(t, err, "Failed to add subnet %s", subnet30)
		err = store.Add(ctx, store.createIPList(subnet31, models.Black))
		require.NoError(t, err, "Failed to add subnet %s", subnet31)

		items, err := store.GetAll(ctx)
		require.NoError(t, err)
		assert.Equal(t, getSubnetsAsMap(items), map[string]bool{
			subnet30: bool(models.White),
			subnet31: bool(models.Black),
		})

		err = store.Remove(ctx, store.createIPList("Invalid", models.White))
		assert.ErrorContains(t, err, "invalid IP address or subnet")

		err = store.Remove(ctx, store.createIPList(subnet30, models.White))
		require.NoError(t, err)
		items, err = store.GetAll(ctx)
		require.NoError(t, err)
		assert.Equal(t, getSubnetsAsMap(items), map[string]bool{
			subnet31: bool(models.Black),
		})

		err = store.Remove(ctx, store.createIPList(subnet31, models.Black))
		require.NoError(t, err)
		items, err = store.GetAll(ctx)
		require.NoError(t, err)
		assert.Equal(t, getSubnetsAsMap(items), map[string]bool{})
	})
}

func TestPostgresStorage_GetIpList(t *testing.T) {
	// Пропускаем если тесты запускаются в коротком режиме
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем хранилище и подключаемся к БД
	store := createStore(t)
	defer store.Close(ctx)

	// Создаем БД ip_lists
	createDB(t, ctx, store)

	// Очищаем тестовые данные перед тестом
	cleanupTestData(t, store)

	t.Run("Add 3 subnets and get white/black lists", func(t *testing.T) {
		// Добавляем первую подсеть в белый список
		_ = store.Add(ctx, store.createIPList(subnet30, models.White))
		_ = store.Add(ctx, store.createIPList(subnet31, models.Black))
		_ = store.Add(ctx, store.createIPList("173.194.221.138/31", models.White))

		items, err := store.GetIpList(ctx, models.White)
		require.NoError(t, err)
		assert.Equal(t, getSubnetsAsMap(items), map[string]bool{
			subnet30:             bool(models.White),
			"173.194.221.138/31": bool(models.White),
		})

		items, err = store.GetIpList(ctx, models.Black)
		require.NoError(t, err)
		assert.Equal(t, getSubnetsAsMap(items), map[string]bool{
			subnet31:             bool(models.Black),
			"173.194.221.138/31": bool(models.White),
		})
	})
}

func TestPostgresStorage_Clear(t *testing.T) {
	// Пропускаем если тесты запускаются в коротком режиме
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем хранилище и подключаемся к БД
	store := createStore(t)
	defer store.Close(ctx)

	// Создаем БД ip_lists
	createDB(t, ctx, store)

	// Очищаем тестовые данные перед тестом
	cleanupTestData(t, store)

	t.Run("Add 3 subnets and clear white/black lists", func(t *testing.T) {
		// Добавляем первую подсеть в белый список
		_ = store.Add(ctx, store.createIPList(subnet30, models.White))
		_ = store.Add(ctx, store.createIPList(subnet31, models.Black))
		_ = store.Add(ctx, store.createIPList("173.194.221.138/31", models.White))

		err := store.Clear(ctx, models.White)
		require.NoError(t, err)

		items, err := store.GetAll(ctx)
		require.NoError(t, err)
		assert.Equal(t, getSubnetsAsMap(items), map[string]bool{
			subnet31: bool(models.Black),
		})
	})
}

func TestPostgresStorage_ClearAll(t *testing.T) {
	// Пропускаем если тесты запускаются в коротком режиме
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем хранилище и подключаемся к БД
	store := createStore(t)
	defer store.Close(ctx)

	// Создаем БД ip_lists
	createDB(t, ctx, store)

	// Очищаем тестовые данные перед тестом
	cleanupTestData(t, store)

	t.Run("Add 3 subnets and clear both lists", func(t *testing.T) {
		// Добавляем первую подсеть в белый список
		_ = store.Add(ctx, store.createIPList(subnet30, models.White))
		_ = store.Add(ctx, store.createIPList(subnet31, models.Black))
		_ = store.Add(ctx, store.createIPList("173.194.221.138/31", models.White))

		err := store.ClearAll(ctx)
		require.NoError(t, err)

		items, err := store.GetAll(ctx)
		require.NoError(t, err)
		assert.Equal(t, getSubnetsAsMap(items), map[string]bool{})
	})
}

func TestPostgresStorage_Contains(t *testing.T) {
	// Пропускаем если тесты запускаются в коротком режиме
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем хранилище и подключаемся к БД
	store := createStore(t)
	defer store.Close(ctx)

	// Создаем БД ip_lists
	createDB(t, ctx, store)

	// Очищаем тестовые данные перед тестом
	cleanupTestData(t, store)

	t.Run("Add 3 subnets and check is a address in a list ", func(t *testing.T) {
		// Добавляем первую подсеть в белый список
		_ = store.Add(ctx, store.createIPList("173.194.221.138/24", models.White))
		_ = store.Add(ctx, store.createIPList("174.194.221.138/24", models.Black))
		_ = store.Add(ctx, store.createIPList("173.194.221.138/16", models.White))

		res, err := store.Contains(ctx, models.White, "Invalid")
		assert.ErrorContains(t, err, "invalid IP address")
		assert.False(t, res)

		res, err = store.Contains(ctx, models.White, "173.194.221.138")
		require.NoError(t, err)
		assert.True(t, res)
		res, err = store.Contains(ctx, models.Black, "173.194.221.138")
		require.NoError(t, err)
		assert.False(t, res)
		res, err = store.Contains(ctx, models.Black, "174.194.221.255")
		require.NoError(t, err)
		assert.True(t, res)
	})
}

func TestPostgresStorage_IsIPAuthorized(t *testing.T) {
	// Пропускаем если тесты запускаются в коротком режиме
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем хранилище и подключаемся к БД
	store := createStore(t)
	defer store.Close(ctx)

	// Создаем БД ip_lists
	createDB(t, ctx, store)

	// Очищаем тестовые данные перед тестом
	cleanupTestData(t, store)

	t.Run("Add 2 subnets and authorise ip-address ", func(t *testing.T) {
		// Добавляем первую подсеть в белый список
		_ = store.Add(ctx, store.createIPList("173.194.221.0/24", models.White))
		_ = store.Add(ctx, store.createIPList("174.194.221.0/24", models.Black))

		res, err := store.IsIPAuthorized(ctx, "Invalid")
		assert.ErrorContains(t, err, "invalid IP address")
		assert.False(t, res)

		res, err = store.IsIPAuthorized(ctx, "173.194.221.1")
		require.NoError(t, err)
		assert.True(t, res)
		res, err = store.IsIPAuthorized(ctx, "174.194.221.1")
		require.NoError(t, err)
		assert.False(t, res)

		res, err = store.IsIPAuthorized(ctx, "200.200.200.200")
		require.NoError(t, err)
		assert.False(t, res)

		store.Clear(ctx, models.White)
		res, err = store.IsIPAuthorized(ctx, "200.200.200.200")
		require.NoError(t, err)
		assert.True(t, res)
	})
}
