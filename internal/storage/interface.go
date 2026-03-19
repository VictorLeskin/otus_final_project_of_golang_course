// internal/storage/interface.go
// interface class to the storage 
package storage


import (
	"context"
	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/models"
)

type IPListStorage interface {
	// Add добавляет подсеть в указанный список
	Add(ctx context.Context, listType models.ListType, subnet string) error

	// Remove удаляет подсеть из указанного списка
	Remove(ctx context.Context, listType models.ListType, subnet string) error

	// Contains проверяет, содержится ли IP в указанном списке
	Contains(ctx context.Context, listType models.ListType, ip string) (bool, error)

	// GetAll возвращает все подсети из указанного списка
	GetAll(ctx context.Context, listType models.ListType) ([]string, error)

	// Clear очищает указанный список (опционально)
	Clear(ctx context.Context, listType models.ListType) error

	// Close закрывает соединение
	Close() error
}
