package memorystorage

import (
	"context"
	"sync"

	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/models"
	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/storage"
)

type MemoryStorage struct {
	mu     sync.RWMutex
	ipList []*models.IPList
}

func New() *MemoryStorage {
	return &MemoryStorage{
		ipList: make([]*models.IPList, 0),
	}
}

// true=white, false=black

func (ms *MemoryStorage) find(subnet string, isWhite models.ListType) *models.IPList {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	for _, ip := range ms.ipList {
		if ip.AreSameS(subnet, isWhite) {
			return ip
		}
	}

	return nil
}

func (ms *MemoryStorage) Add(ctx context.Context, l *models.IPList) error {
	//checking
	if err := l.Validate(); err != nil {
		return storage.ErrInvalidSubnetDetected
	}

	// subnet exist
	if ms.find(l.Subnet, l.IsWhite) != nil {
		return nil
	}

	ms.mu.Lock()
	defer ms.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ms.ipList = append(ms.ipList, l)
	return nil
}
