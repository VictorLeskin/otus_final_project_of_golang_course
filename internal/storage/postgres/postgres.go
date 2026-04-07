package postgresstorage

import (
	"context"
	"database/sql"
	"fmt"
	"net"

	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/models"

	_ "github.com/lib/pq" // драйвер PostgreSQL...
)

type Config struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

type PostgresStorage struct {
	cfg Config
	db  *sql.DB
}

func New(cfg Config) *PostgresStorage {
	return &PostgresStorage{
		cfg: cfg,
	}
}

func (s *PostgresStorage) DSN() string {
	c := s.cfg
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

func (s *PostgresStorage) Connect(ctx context.Context) error {
	db, err := sql.Open("postgres", s.DSN())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// check the connection.
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	s.db = db

	return nil
}

func (s *PostgresStorage) Close(_ context.Context) error {
	return s.db.Close()
}

func (s *PostgresStorage) Add(ctx context.Context, l models.IPList) error {
	// checking.
	if err := l.Validate(); err != nil {
		return fmt.Errorf("invalid IP list entry: %w", err)
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO ip_lists (subnet, list_type) VALUES ($1, $2)
         ON CONFLICT (subnet, list_type) DO NOTHING`,
		l.Subnet, l.IsWhite,
	)
	return err
}

// Remove удаляет подсеть из указанного списка.
func (s *PostgresStorage) Remove(ctx context.Context, l models.IPList) error {
	if err := l.Validate(); err != nil {
		return fmt.Errorf("invalid IP list entry: %w", err)
	}

	_, err := s.db.ExecContext(ctx,
		`DELETE FROM ip_lists WHERE subnet = $1 AND list_type = $2`,
		l.Subnet, l.IsWhite,
	)
	return err
}

// GetIPList возвращает все подсети из указанного списка.
func (s *PostgresStorage) GetIPList(ctx context.Context, listType models.ListType) ([]models.IPList, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT subnet, list_type
         FROM ip_lists
         WHERE list_type = $1
         ORDER BY subnet`,
		listType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.IPList
	for rows.Next() {
		var ip models.IPList
		if err := rows.Scan(&ip.Subnet, &ip.IsWhite); err != nil {
			return nil, err
		}
		result = append(result, ip)
	}
	return result, nil
}

// GetAll возвращает все подсети из обоих списков.
func (s *PostgresStorage) GetAll(ctx context.Context) ([]models.IPList, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, subnet, list_type, created_at
         FROM ip_lists
         ORDER BY list_type, subnet`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.IPList
	for rows.Next() {
		var ip models.IPList
		if err := rows.Scan(&ip.ID, &ip.Subnet, &ip.IsWhite, &ip.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, ip)
	}
	return result, rows.Err()
}

// Clear очищает указанный список.
func (s *PostgresStorage) Clear(ctx context.Context, listType models.ListType) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM ip_lists WHERE list_type = $1`,
		listType,
	)
	return err
}

// ClearAll очищает оба списка.
func (s *PostgresStorage) ClearAll(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM ip_lists`,
	)
	return err
}

// Contains проверяет, принадлежит ли IP-адрес указанному списку.
func (s *PostgresStorage) Contains(ctx context.Context, listType models.ListType, address string) (bool, error) {
	// Проверяем, что address валидный IP
	ipAddr := net.ParseIP(address)
	if ipAddr == nil {
		return false, fmt.Errorf("invalid IP address: %s", address)
	}

	ipList, err := s.GetIPList(ctx, listType)
	if err != nil {
		return false, err
	}
	for _, ip := range ipList {
		if ip.Contains(ipAddr) {
			return true, nil
		}
	}
	return false, nil
}

// IsIPAuthorized проверяет IP по white/black спискам
// Возвращает true, если IP разрешен:
//   - не в blacklist И (whitelist пуст ИЛИ IP в whitelist).
func (s *PostgresStorage) IsIPAuthorized(ctx context.Context, ip string) (bool, error) {
	// Проверяем валидность IP.
	if net.ParseIP(ip) == nil {
		return false, fmt.Errorf("invalid IP address: %s", ip)
	}

	// 1. Проверяем blacklist (высший приоритет).
	inBlack, err := s.Contains(ctx, models.Black, ip)
	if err != nil {
		return false, err
	}
	if inBlack {
		return false, nil
	}

	// 2. Получаем whitelist.
	whiteList, err := s.GetIPList(ctx, models.White)
	if err != nil {
		return false, err
	}

	// 3. Если whitelist пуст — разрешаем.
	if len(whiteList) == 0 {
		return true, nil
	}

	// 4. Иначе проверяем, есть ли IP в whitelist.
	return s.Contains(ctx, models.White, ip)
}
