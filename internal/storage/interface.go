// internal/storage/interface.go
// interface class to the storage
package storage

import (
	"context"
	"fmt"

	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/models"
)

type Config struct {
	Type string // memory or SQL
}

type IPListStorage interface {
	// Connect открывает соединение
	Connect(ctx context.Context) error

	// Add добавляет подсеть в указанный список
	Add(ctx context.Context, l models.IPList) error

	// Remove удаляет подсеть из указанного списка
	Remove(ctx context.Context, l models.IPList) error

	// Contains проверяет, содержится ли IP в указанном списке
	Contains(ctx context.Context, listType models.ListType, address string) (bool, error)

	GetIpList(ctx context.Context, listType models.ListType) ([]models.IPList, error)
	GetAll(ctx context.Context) ([]models.IPList, error)

	// Clear очищает указанный список (опционально)
	Clear(ctx context.Context, listType models.ListType) error
	ClearAll(ctx context.Context) error

	IsIPAuthorized(ctx context.Context, ip string) (bool, error)

	// Close закрывает соединение
	Close(ctx context.Context) error
}

var (
	ErrInvalidSubnetDetected  = fmt.Errorf("invalid subnet")
	ErrInvalidAddressDetected = fmt.Errorf("invalid address")
	ErrEventExists            = fmt.Errorf("this event exists")
	ErrEmptySubnet            = fmt.Errorf("subnet cannot be empty")
	ErrEmptyIP                = fmt.Errorf("IP cannot be empty")
	ErrInvalidSubnet          = fmt.Errorf("invalid subnet format")
)
