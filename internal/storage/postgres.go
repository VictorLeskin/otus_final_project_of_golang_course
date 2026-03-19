package storage

import (
    "context"
    "database/sql"
    "fmt"
    
    "github.com/VictorLeskin/otus_final_project_of_golang_course/internal/models"
)

type PostgresStorage struct {
    db *sql.DB
}

func (s *PostgresStorage) Add(ctx context.Context, listType models.ListType, subnet string) error {
    // Валидация CIDR
    if subnet == "" {
        return ErrEmptySubnet
    }
    if _, _, err := net.ParseCIDR(subnet); err != nil {
        return fmt.Errorf("%w: %v", ErrInvalidSubnet, err)
    }
    
    _, err := s.db.ExecContext(ctx,
        `INSERT INTO ip_lists (list_type, subnet) VALUES ($1, $2)
         ON CONFLICT (list_type, subnet) DO NOTHING`,
        listType, subnet,
    )
    return err
}

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

// Ошибки
var (
    ErrEmptySubnet   = errors.New("subnet cannot be empty")
    ErrEmptyIP       = errors.New("IP cannot be empty")
    ErrInvalidSubnet = errors.New("invalid subnet format")
)