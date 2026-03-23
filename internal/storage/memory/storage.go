package memorystorage

import (
	"context"
	"fmt"
	"net"
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

func (ms *MemoryStorage) Add(ctx context.Context, l models.IPList) error {
	//checking
	if err := l.Validate(); err != nil {
		return fmt.Errorf("invalid IP list entry: %w", err)
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

	tmp := new(models.IPList)
	*tmp = l
	ms.ipList = append(ms.ipList, tmp)
	return nil
}

func (ms *MemoryStorage) Remove(ctx context.Context, l models.IPList) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	for i, ip := range ms.ipList {
		if ip.AreSame(&l) {
			ms.ipList = append(ms.ipList[:i], ms.ipList[i+1:]...)
			return nil
		}
	}

	// no such subnet in list
	return nil
}

func (ms *MemoryStorage) GetIpList(ctx context.Context, listType models.ListType) ([]models.IPList, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	res := make([]models.IPList, 0)

	for _, ip := range ms.ipList {
		if ip.IsWhite == listType {
			res = append(res, *ip)
		}
	}

	return res, nil
}

func (ms *MemoryStorage) GetAll(ctx context.Context) ([]models.IPList, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	res := make([]models.IPList, 0)

	for _, ip := range ms.ipList {
		res = append(res, *ip)
	}

	return res, nil
}

func (ms *MemoryStorage) Clear(ctx context.Context, listType models.ListType) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	n := 0
	for _, it := range ms.ipList {
		if it.IsWhite != listType {
			ms.ipList[n] = it
			n++
		}
	}
	ms.ipList = ms.ipList[:n]
	return nil
}

//

func (ms *MemoryStorage) ClearAll(ctx context.Context) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	ms.ipList = ms.ipList[:0]
	return nil
}

// Contains проверяет принадлежность IP-адреса к указанному списку.
// Параметры:
//   - ctx: контекст для отмены операции
//   - listType: тип списка (white/black)
//   - address: IP-адрес или CIDR подсеть для проверки
//
// Возвращает:
//   - bool: true если адрес найден в списке, false если нет
//   - error: ошибка при парсинге address или проблемах с БД
func (ms *MemoryStorage) Contains(ctx context.Context, listType models.ListType, address string) (bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
	}

	// Проверяем IP на валидность
	ipAddr := net.ParseIP(address)
	if ipAddr == nil {
		return false, storage.ErrInvalidAddressDetected
	}

	for _, ip := range ms.ipList {
		if ip.IsWhite == listType {
			if ip.Contains(ipAddr) {
				return true, nil
			}
		}
	}
	return false, nil
}

// count возвращает количестов subnet даного типа
// Параметры:
//   - listType: тип списка (white/black)
func (ms *MemoryStorage) count(listType models.ListType) int {
	n := 0
	for _, ip := range ms.ipList {
		if ip.IsWhite == listType {
			n++
		}
	}
	return n
}

// IsIPAuthorized проверяет IP по white/black спискам.
// Возвращает true, если IP разрешен:
//   - не в blacklist и whitelist пуст
//   - не в blacklist и IP в whitelist если white list не пуст)
func (ms *MemoryStorage) IsIPAuthorized(ctx context.Context, ip string) (bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	resB, errB := ms.Contains(ctx, models.Black, ip)
	if errB != nil {
		return false, errB
	}
	if resB {
		return false, nil
	}

	if ms.count(models.Black) == len(ms.ipList) {
		return true, nil
	}

	resW, errW := ms.Contains(ctx, models.White, ip)
	if errW != nil {
		return false, errW
	}
	return resW, nil
}
