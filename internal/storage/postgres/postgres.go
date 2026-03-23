package postgresstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/models"

	"github.com/lib/pq" // драйвер PostgreSQL
)

type Config struct {
	Host     string
	Port     int
	Database string
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
		"host=%s port=%d dbname=%s sslmode=%s",
		c.Host, c.Port, c.Database, c.SSLMode,
	)
}

func (s *PostgresStorage) Connect(ctx context.Context) error {
	db, err := sql.Open("postgres", s.DSN())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// check the connection
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

// Вспомогательная функция для определения duplicate key.
func isDuplicateError(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code == "23505" { // PostgrePostgres error code for duplicating "23505".
			return true
		}
	}

	return false
}

func (s *PostgresStorage) Add(ctx context.Context, l models.IPList) error {
	//checking
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

/*
func (s *PostgresStorage) Remove(ctx context.Context, listType models.ListType, subnet string) error {
	if subnet == "" {
		return ErrEmptySubnet
	}

	_, err := s.db.ExecContext(ctx,
		`DELETE FROM ip_lists WHERE list_type = $1 AND subnet = $2`,
		listType, subnet,
	)
	return err
}

func (s *PostgresStorage) Contains(ctx context.Context, listType models.ListType, ip string) (bool, error) {
	if ip == "" {
		return false, ErrEmptyIP
	}

	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(
            SELECT 1 FROM ip_lists
            WHERE list_type = $1 AND $2::inet <<= subnet
        )`,
		listType, ip,
	).Scan(&exists)

	return exists, err
}

func (s *PostgresStorage) GetAll(ctx context.Context, listType models.ListType) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT subnet FROM ip_lists WHERE list_type = $1 ORDER BY subnet`,
		listType,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subnets []string
	for rows.Next() {
		var subnet string
		if err := rows.Scan(&subnet); err != nil {
			return nil, err
		}
		subnets = append(subnets, subnet)
	}
	return subnets, rows.Err()
}

func (s *PostgresStorage) Clear(ctx context.Context, listType models.ListType) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM ip_lists WHERE list_type = $1`,
		listType,
	)
	return err
}

func (s *PostgresStorage) Close() error {
	return s.db.Close()
}
*/
