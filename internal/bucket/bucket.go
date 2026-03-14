package bucket

import (
	"sync"
)

// Bucket представляет собой ведро для rate limiting'а (алгоритм leaky bucket).
type Bucket struct {
	mu         sync.RWMutex
	key        string // идентификатор (логин/пароль/IP)
	capacity   int    // размер ведра (сколько запросов может накопиться)
	drops      int    // сколько сейчас воды (запросов) в ведре
	lastLeak   Tick   // время последней утечки (в миллисекундах)
	leakRate   int    // сколько МИЛЛИСЕКУНД на утечку одной единицы воды
	lastAccess Tick   // время последнего обращения (для очистки)
}

// NewBucket создает новое ведро с указанным leakRate в миллисекундах.
func NewBucket(key string, capacity int, leakRateMillis int) *Bucket {
	return &Bucket{
		key:        key,
		capacity:   capacity,
		drops:      0,
		lastLeak:   0,
		leakRate:   leakRateMillis,
		lastAccess: 0,
	}
}

// NewBucketFromRPM создает ведро из количества попыток в минуту.
func NewBucketFromRPM(key string, attemptsPerMinute int) *Bucket {
	// В минуте 60000 миллисекунд
	leakRateMillis := 60000 / attemptsPerMinute
	return NewBucket(key, attemptsPerMinute, leakRateMillis)
}

// Allow добавляет запрос (воду) в ведро в указанное время.
func (b *Bucket) Allow(tick Tick) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.lastAccess = tick
	b.leak(tick)

	if b.drops < b.capacity {
		b.drops++
		return true
	}
	return false
}

// leak выливает воду из ведра пропорционально прошедшему времени.
func (b *Bucket) leak(tick Tick) {
	elapsed := tick.Sub(b.lastLeak) // разница в миллисекундах

	// Сколько воды утекло за это время
	if leaked := int(elapsed / int64(b.leakRate)); leaked > 0 {
		b.drops -= leaked
		if b.drops < 0 {
			b.drops = 0
		}
		b.lastLeak = tick
	}
}

// Stats возвращает текущую статистику ведра.
func (b *Bucket) Stats() (drops, capacity int, lastLeak Tick) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.drops, b.capacity, b.lastLeak
}

// IsExpired проверяет, не было ли обращений к ведру дольше чем milliseconds.
func (b *Bucket) IsExpired(now Tick, milliseconds int64) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return now.Sub(b.lastAccess) > milliseconds
}

// Key возвращает ключ ведра.
func (b *Bucket) Key() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.key
}

// WaterLevel возвращает текущий уровень воды.
func (b *Bucket) WaterLevel() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.drops
}

// WaterLevel возвращает true если ведро пустое
func (b *Bucket) IsEmpty() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.drops == 0
}

// Remaining возвращает оставшееся место в ведре.
func (b *Bucket) Remaining() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.capacity - b.drops
}
