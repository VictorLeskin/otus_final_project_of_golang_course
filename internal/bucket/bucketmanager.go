package bucket

import (
	"sync"
	"time"
)

// BucketCollection представляет коллекцию bucket'ов одного типа (login/password/IP)
type BucketCollection struct {
	buckets        *sync.Map // map[string]*Bucket
	capacity       int
	leakRateMillis int
}

// NewBucketCollection создает новую коллекцию bucket'ов
func NewBucketCollection(maxAttemptsPM int) *BucketCollection {
	leakRateMillis := 60000 / maxAttemptsPM
	return &BucketCollection{
		buckets:        &sync.Map{},
		capacity:       maxAttemptsPM,
		leakRateMillis: leakRateMillis,
	}
}

// BucketManager управляет всеми bucket'ами для rate limiting'а
type BucketManager struct {
	loginBuckets    *BucketCollection
	passwordBuckets *BucketCollection
	ipBuckets       *BucketCollection

	config *Config

	stopCleanup chan struct{}
	wg          sync.WaitGroup
}

// Config содержит настройки rate limiting'а
type Config struct {
	N int // max attempts per minute for login
	M int // max attempts per minute for password
	K int // max attempts per minute for IP

	CleanupInterval time.Duration // как часто чистить пустые bucket'ы
}

// NewBucketManager создает новый менеджер bucket'ов
func NewBucketManager(config *Config) *BucketManager {
	if config == nil {
		config = &Config{
			N:               10,
			M:               100,
			K:               1000,
			CleanupInterval: 5 * time.Minute,
		}
	}

	if config.CleanupInterval == 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	m := &BucketManager{
		loginBuckets:    NewBucketCollection(config.N),
		passwordBuckets: NewBucketCollection(config.M),
		ipBuckets:       NewBucketCollection(config.K),
		config:          config,
		stopCleanup:     make(chan struct{}),
	}

	m.startCleanup()
	return m
}

// Stop останавливает фоновую очистку
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

func (m *BucketManager) cleanupCollection(collection *BucketCollection) {
	var keysToDelete []string

	collection.buckets.Range(func(key, value interface{}) bool {
		if value.(*Bucket).IsEmpty() {
			keysToDelete = append(keysToDelete, key.(string))
		}
		return true
	})

	for _, key := range keysToDelete {
		collection.buckets.Delete(key)
	}
}

func (m *BucketManager) cleanup() {
	m.cleanupCollection(m.loginBuckets)
	m.cleanupCollection(m.passwordBuckets)
	m.cleanupCollection(m.ipBuckets)
}

func (m *BucketManager) getBucket(collection *BucketCollection, key string) *Bucket {
	if val, ok := collection.buckets.Load(key); ok {
		return val.(*Bucket)
	}

	bucket := NewBucket(key, collection.capacity, collection.leakRateMillis)
	val, _ := collection.buckets.LoadOrStore(key, bucket)
	return val.(*Bucket)
}

// Check проверяет лимит для ключа в указанной коллекции
func (m *BucketManager) Check(collection *BucketCollection, key string, tick Tick) bool {
	bucket := m.getBucket(collection, key)
	return bucket.Allow(tick)
}

// CheckAuth проверяет лимиты для логина, пароля и IP
func (m *BucketManager) CheckAuth(login, password, ip string) bool {
	now := NowTick()

	return m.Check(m.loginBuckets, login, now) &&
		m.Check(m.passwordBuckets, password, now) &&
		m.Check(m.ipBuckets, ip, now)
}

// ResetLogin сбрасывает bucket для логина
func (m *BucketManager) ResetLogin(login string) {
	m.loginBuckets.buckets.Delete(login)
}

// ResetIP сбрасывает bucket для IP
func (m *BucketManager) ResetIP(ip string) {
	m.ipBuckets.buckets.Delete(ip)
}

// ResetAll сбрасывает bucket'ы для логина и IP
func (m *BucketManager) ResetAll(login, ip string) {
	m.ResetLogin(login)
	m.ResetIP(ip)
}

// Stats возвращает количество активных bucket'ов по типам
func (m *BucketManager) Stats() map[string]int {
	stats := make(map[string]int)

	count := 0
	m.loginBuckets.buckets.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	stats["login"] = count

	count = 0
	m.passwordBuckets.buckets.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	stats["password"] = count

	count = 0
	m.ipBuckets.buckets.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	stats["ip"] = count

	return stats
}
