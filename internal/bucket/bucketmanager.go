package bucket

import (
	"sync"
	"time"
)

// BucketCollection представляет коллекцию bucket'ов одного типа (login/password/IP).
type BucketCollection struct {
	buckets        *sync.Map // map[string]*Bucket ....
	capacity       int
	leakRateMillis int
}

// NewBucketCollection создает новую коллекцию bucket'ов.
func NewBucketCollection(maxAttemptsPM int) *BucketCollection {
	leakRateMillis := 60000 / maxAttemptsPM
	return &BucketCollection{
		buckets:        &sync.Map{},
		capacity:       maxAttemptsPM,
		leakRateMillis: leakRateMillis,
	}
}

// BucketManager управляет всеми bucket'ами для rate limiting'а.
type BucketManager struct {
	loginBuckets    *BucketCollection
	passwordBuckets *BucketCollection
	ipBuckets       *BucketCollection

	config *Config

	stopCleanup chan struct{}
	wg          sync.WaitGroup
}

// Config содержит настройки rate limiting'а.
type Config struct {
	loginRate    int // max attempts per minute for login
	passwordRate int // max attempts per minute for password
	ipRate       int // max attempts per minute for IP address

	CleanupInterval time.Duration // как часто чистить пустые bucket'ы
}

func DefaultConfig() *Config {
	return &Config{
		loginRate:       10, // no more 10 times in  1 minute
		passwordRate:    100,
		ipRate:          1000,
		CleanupInterval: 1 * time.Minute, // once a minute
	}
}

// NewBucketManager создает новый менеджер bucket'ов.
func NewBucketManager(config *Config) *BucketManager {
	if config == nil {
		config = DefaultConfig()
	}

	if config.CleanupInterval == 0 {
		config.CleanupInterval = DefaultConfig().CleanupInterval
	}

	m := &BucketManager{
		loginBuckets:    NewBucketCollection(config.loginRate),
		passwordBuckets: NewBucketCollection(config.passwordRate),
		ipBuckets:       NewBucketCollection(config.ipRate),
		config:          config,
		stopCleanup:     make(chan struct{}),
	}

	m.startCleanup()
	return m
}

// Stop останавливает фоновую очистку.
func (m *BucketManager) Stop() {
	close(m.stopCleanup)
	m.wg.Wait()
}

func (m *BucketManager) startCleanup() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		ticker := time.NewTicker(m.config.CleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.cleanup()
			case <-m.stopCleanup:
				return
			}
		}
	}()
}

func (m *BucketManager) cleanupCollection(collection *BucketCollection, tick Tick) {
	var keysToDelete []string
	// обновить состояние ведра для текущего времени и
	// если оно пусто удалить его.

	collection.buckets.Range(func(key, value interface{}) bool {
		if !value.(*Bucket).TimeUpdate(tick) {
			keysToDelete = append(keysToDelete, key.(string))
		}
		return true
	})

	for _, key := range keysToDelete {
		collection.buckets.Delete(key)
	}
}

func (m *BucketManager) cleanup() {
	now := NowTick()

	m.cleanupCollection(m.loginBuckets, now)
	m.cleanupCollection(m.passwordBuckets, now)
	m.cleanupCollection(m.ipBuckets, now)
}

func (m *BucketManager) getBucket(collection *BucketCollection, key string) *Bucket {
	if val, ok := collection.buckets.Load(key); ok {
		return val.(*Bucket)
	}

	bucket := NewBucket(key, collection.capacity, collection.leakRateMillis)
	val, _ := collection.buckets.LoadOrStore(key, bucket)
	return val.(*Bucket)
}

// Check проверяет лимит для ключа в указанной коллекции.
func (m *BucketManager) Check(collection *BucketCollection, key string, tick Tick) bool {
	bucket := m.getBucket(collection, key)
	return bucket.Allow(tick)
}

// CheckAuth проверяет лимиты для логина, пароля и IP.
func (m *BucketManager) CheckAuth(login, password, ip string) bool {
	now := NowTick()

	return m.Check(m.loginBuckets, login, now) &&
		m.Check(m.passwordBuckets, password, now) &&
		m.Check(m.ipBuckets, ip, now)
}

// ResetLogin сбрасывает bucket для логина.
func (m *BucketManager) ResetLogin(login string) {
	m.loginBuckets.buckets.Delete(login)
}

// ResetIP сбрасывает bucket для IP.
func (m *BucketManager) ResetIP(ip string) {
	m.ipBuckets.buckets.Delete(ip)
}

// ResetAll сбрасывает bucket'ы для логина и IP.
func (m *BucketManager) ResetAll(login, ip string) {
	m.ResetLogin(login)
	m.ResetIP(ip)
}

func dropsInBucket(buckets *sync.Map) int {
	count0 := 0
	buckets.Range(func(_, _ interface{}) bool {
		count0++
		return true
	})
	return count0
}

// Stats возвращает количество активных bucket'ов по типам.
func (m *BucketManager) Stats() map[string]int {
	stats := make(map[string]int)

	stats["login"] = dropsInBucket(m.loginBuckets.buckets)
	stats["password"] = dropsInBucket(m.passwordBuckets.buckets)
	stats["ip"] = dropsInBucket(m.ipBuckets.buckets)

	return stats
}

func (m *BucketManager) BucketStats(buckets *sync.Map) map[string]int {
	stats := make(map[string]int)
	buckets.Range(func(key, val interface{}) bool {
		stats[key.(string)] = val.(*Bucket).drops
		return true
	})
	return stats
}
