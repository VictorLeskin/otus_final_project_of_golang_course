package bucket

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBucket_NewBucketCollection(t *testing.T) {
	t0 := NewBucketCollection(150)

	assert.Equal(t, 150, t0.capacity)
	assert.Equal(t, 400, t0.leakRateMillis)
	assert.NotNil(t, t0.buckets)
}

func TestBucket_DefaultConfig(t *testing.T) {
	t0 := DefaultConfig()

	assert.Equal(t, 10, t0.LoginRate)
	assert.Equal(t, 100, t0.PasswordRate)
	assert.NotNil(t, 1000, t0.IPRate)
	assert.NotNil(t, 60, t0.CleanupInterval)
}

func TestNewBucketManager_WithConfig(t *testing.T) {
	// Arrange
	config := &Config{
		LoginRate:       20,
		PasswordRate:    200,
		IPRate:          2000,
		CleanupInterval: 30 * time.Second,
	}

	// Act
	manager := NewBucketManager(config)
	defer manager.Stop()

	// Assert
	assert.NotNil(t, manager)
	assert.Equal(t, config, manager.config) // тот же объект, не копия

	// Проверяем, что коллекции созданы с правильными параметрами
	assert.Equal(t, config.LoginRate, manager.loginBuckets.capacity)
	assert.Equal(t, config.PasswordRate, manager.passwordBuckets.capacity)
	assert.Equal(t, config.IPRate, manager.ipBuckets.capacity)

	// Проверяем leakRate (60000 / rate)
	assert.Equal(t, 3000, manager.loginBuckets.leakRateMillis)   // 60000/20
	assert.Equal(t, 300, manager.passwordBuckets.leakRateMillis) // 60000/200
	assert.Equal(t, 30, manager.ipBuckets.leakRateMillis)        // 60000/2000

	// Проверяем, что канал создан
	assert.NotNil(t, manager.stopCleanup)
}

func TestNewBucketManager_WithNilConfig(t *testing.T) {
	// Act
	manager := NewBucketManager(nil)
	defer manager.Stop()

	// Assert
	assert.NotNil(t, manager)

	// Проверяем, что использованы значения по умолчанию
	defaultConfig := DefaultConfig()
	assert.Equal(t, defaultConfig.LoginRate, manager.config.LoginRate)
	assert.Equal(t, defaultConfig.PasswordRate, manager.config.PasswordRate)
	assert.Equal(t, defaultConfig.IPRate, manager.config.IPRate)
	assert.Equal(t, defaultConfig.CleanupInterval, manager.config.CleanupInterval)
}

func TestNewBucketManager_WithZeroCleanupInterval(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	config.CleanupInterval = 0

	// Act
	manager := NewBucketManager(config)
	defer manager.Stop()

	// Assert
	assert.NotNil(t, manager)
	// Должен быть установлен интервал по умолчанию
	assert.Equal(t, DefaultConfig().CleanupInterval, manager.config.CleanupInterval)
}

func TestNewBucketManager_StartedCleanup(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	config.CleanupInterval = 50 * time.Millisecond // маленький интервал для теста

	// Act
	manager := NewBucketManager(config)

	// Assert - cleanup должен запуститься
	// Создадим bucket и сразу сделаем его пустым
	tick := Tick(1000)
	manager.Check(manager.loginBuckets, "test", tick) // water = 1

	// count buckets in loginBuckets
	assert.Equal(t, 1, dropsInBucket(manager.loginBuckets.buckets))

	// Ждем утечки (leakRate=6000ms для LoginRate=10)
	time.Sleep(1 * time.Second)        // вода утекла
	time.Sleep(100 * time.Millisecond) // cleanup

	// Проверяем что bucket удалился
	assert.Equal(t, 0, dropsInBucket(manager.loginBuckets.buckets))

	manager.Stop()
}

func TestBucketManager_Stats(t *testing.T) {
	manager := NewBucketManager(DefaultConfig())
	defer manager.Stop()

	// Создаем несколько bucket'ов
	tick := NowTick()
	manager.Check(manager.loginBuckets, "user1", tick)
	manager.Check(manager.loginBuckets, "user2", tick)
	manager.Check(manager.passwordBuckets, "pass1", tick)
	manager.Check(manager.ipBuckets, "192.168.1.1", tick)
	manager.Check(manager.ipBuckets, "10.0.0.1", tick)

	// Получаем статистику
	stats := manager.Stats()

	// Проверяем
	assert.Equal(t, 2, stats["login"])    // user1, user2
	assert.Equal(t, 1, stats["password"]) // pass1
	assert.Equal(t, 2, stats["ip"])       // 192.168.1.1, 10.0.0.1
}

func TestBucketManager_BucketStats(t *testing.T) {
	manager := NewBucketManager(DefaultConfig())
	defer manager.Stop()

	// Создаем несколько bucket'ов
	tick := NowTick()
	manager.Check(manager.loginBuckets, "user1", tick)
	manager.Check(manager.loginBuckets, "user1", tick+1)
	manager.Check(manager.loginBuckets, "user2", tick)

	// Получаем статистику
	stats := manager.BucketStats(manager.loginBuckets.buckets)

	// Проверяем
	assert.Equal(t, 2, stats["user1"])
	assert.Equal(t, 1, stats["user2"])
}

func TestNewBucketManager_cleanupCollection(t *testing.T) {
	config := &Config{
		LoginRate:       60,
		PasswordRate:    60,
		IPRate:          60,
		CleanupInterval: 2 * time.Second,
	}

	t0 := NewBucketManager(config)
	initTickMs := Tick(32000) // 32 sec

	t0.Check(t0.loginBuckets, "user0", initTickMs)
	t0.cleanupCollection(t0.loginBuckets, initTickMs+500)

	val, ok := t0.loginBuckets.buckets.Load("user0")
	assert.Equal(t, 1, val.(*Bucket).drops)
	assert.Equal(t, true, ok)
	assert.Equal(t, 1, t0.Stats()["login"])

	t0.Check(t0.loginBuckets, "user0", initTickMs+1000)
	t0.cleanupCollection(t0.loginBuckets, initTickMs+1500)

	val, ok = t0.loginBuckets.buckets.Load("user0")
	assert.Equal(t, 1, val.(*Bucket).drops)
	assert.Equal(t, true, ok)
	assert.Equal(t, 1, t0.Stats()["login"])

	t0.Check(t0.loginBuckets, "user0", initTickMs+2000)
	t0.cleanupCollection(t0.loginBuckets, initTickMs+2500)

	val, ok = t0.loginBuckets.buckets.Load("user0")
	assert.Equal(t, 1, val.(*Bucket).drops)
	assert.Equal(t, true, ok)
	assert.Equal(t, 1, t0.Stats()["login"])

	t0.Check(t0.loginBuckets, "user1", initTickMs+3000)
	t0.cleanupCollection(t0.loginBuckets, initTickMs+3500)

	val, ok = t0.loginBuckets.buckets.Load("user1")
	assert.Equal(t, 1, val.(*Bucket).drops)
	assert.Equal(t, true, ok)
	assert.Equal(t, 1, t0.Stats()["login"])

	t0.Check(t0.loginBuckets, "user1", initTickMs+4000)
	t0.cleanupCollection(t0.loginBuckets, initTickMs+4500)

	val, ok = t0.loginBuckets.buckets.Load("user1")
	assert.Equal(t, 1, val.(*Bucket).drops)
	assert.Equal(t, true, ok)
	assert.Equal(t, 1, t0.Stats()["login"])
}

func TestNewBucketManager_cleanup(t *testing.T) {
	config := &Config{
		LoginRate:       200,
		PasswordRate:    200,
		IPRate:          200,
		CleanupInterval: 20 * time.Millisecond,
	}

	t0 := NewBucketManager(config)
	initTickMs := NowTick() // 32 sec

	t0.Check(t0.loginBuckets, "user0", initTickMs)
	t0.Check(t0.passwordBuckets, "pass0", initTickMs)
	t0.Check(t0.ipBuckets, "ip0", initTickMs)

	assert.Equal(t, 1, t0.Stats()["login"])
	assert.Equal(t, 1, t0.Stats()["password"])
	assert.Equal(t, 1, t0.Stats()["password"])

	time.Sleep(2 * time.Second) // вода утекла
	t0.cleanup()

	assert.Equal(t, 0, t0.Stats()["login"])
	assert.Equal(t, 0, t0.Stats()["password"])
	assert.Equal(t, 0, t0.Stats()["password"])
}

func TestNewBucketManager_getBucket(t *testing.T) {
	t0 := NewBucketManager(DefaultConfig())

	user0 := t0.getBucket(t0.loginBuckets, "user0")
	user1 := t0.getBucket(t0.loginBuckets, "user0")
	assert.Equal(t, user0, user1)

	user2 := t0.getBucket(t0.loginBuckets, "user1")
	assert.NotEqual(t, user0, user2)
}

func TestNewBucketManager_Check(t *testing.T) {
	t0 := NewBucketManager(DefaultConfig())

	t0.loginBuckets.capacity = 3
	res0 := t0.Check(t0.loginBuckets, "user0", Tick(32000))
	res1 := t0.Check(t0.loginBuckets, "user0", Tick(32000+1))
	res2 := t0.Check(t0.loginBuckets, "user0", Tick(32000+2))
	res3 := t0.Check(t0.loginBuckets, "user0", Tick(32000+3))

	assert.True(t, res0)
	assert.True(t, res1)
	assert.True(t, res2)
	assert.False(t, res3)

	assert.Equal(t, 3, t0.BucketStats(t0.loginBuckets.buckets)["user0"])
}

func TestNewBucketManager_CheckAuth(t *testing.T) {
	t0 := NewBucketManager(DefaultConfig())

	t0.loginBuckets.capacity = 2
	t0.passwordBuckets.capacity = 2
	t0.ipBuckets.capacity = 2

	// all empty add one token to all buckets
	res0 := t0.CheckAuth("user0", "password0", "ip0")
	assert.True(t, res0)
	assert.Equal(t, 1, t0.BucketStats(t0.loginBuckets.buckets)["user0"])
	assert.Equal(t, 1, t0.BucketStats(t0.passwordBuckets.buckets)["password0"])
	assert.Equal(t, 1, t0.BucketStats(t0.ipBuckets.buckets)["ip0"])

	// add one more token to all buckets
	res0 = t0.CheckAuth("user0", "password0", "ip0")
	assert.True(t, res0)
	assert.Equal(t, 2, t0.BucketStats(t0.loginBuckets.buckets)["user0"])
	assert.Equal(t, 2, t0.BucketStats(t0.passwordBuckets.buckets)["password0"])
	assert.Equal(t, 2, t0.BucketStats(t0.ipBuckets.buckets)["ip0"])

	// add one more token to all buckets. bucket of logins is full
	res0 = t0.CheckAuth("user1", "password1", "ip0")
	assert.False(t, res0)
	assert.Equal(t, 2, t0.BucketStats(t0.loginBuckets.buckets)["user0"])
	assert.Equal(t, 2, t0.BucketStats(t0.passwordBuckets.buckets)["password0"])
	assert.Equal(t, 2, t0.BucketStats(t0.ipBuckets.buckets)["ip0"])

	// add one more token to all buckets. bucket of logins is full
	res0 = t0.CheckAuth("user0", "password1", "ip1")
	assert.False(t, res0)
	assert.Equal(t, 2, t0.BucketStats(t0.loginBuckets.buckets)["user0"])
	assert.Equal(t, 2, t0.BucketStats(t0.passwordBuckets.buckets)["password0"])
	assert.Equal(t, 2, t0.BucketStats(t0.ipBuckets.buckets)["ip0"])

}

func TestNewBucketManager_Stop(t *testing.T) {
	// Arrange
	config := DefaultConfig()
	manager := NewBucketManager(config)

	// Act
	manager.Stop()

	// Assert - после Stop канал должен быть закрыт
	_, ok := <-manager.stopCleanup
	assert.False(t, ok, "канал должен быть закрыт")

	// Проверяем что повторный вызов Stop вызовет панику
	assert.Panics(t, func() {
		manager.Stop()
	})
}

func TestNewBucketManager_ConcurrentAccess(t *testing.T) {
	// Arrange
	config := DefaultConfig()

	manager := NewBucketManager(config)
	defer manager.Stop()

	// Act - конкурентный доступ из многих горутин
	const goroutines = 50
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(_ int) {
			defer wg.Done()
			tick := NowTick()
			for j := 0; j < iterations; j++ {
				// Проверяем разные ключи
				manager.Check(manager.loginBuckets, "user", tick)
				manager.Check(manager.passwordBuckets, "pass", tick)
				manager.Check(manager.ipBuckets, "192.168.1.1", tick)
			}
		}(i)
	}

	wg.Wait()

	// Assert - все должно работать без паник и data race
	// Проверяем что bucket'ы создались
	count := 0
	manager.loginBuckets.buckets.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	assert.Equal(t, 1, count) // один пользователь
}

func TestNewBucketManager_ResetLogin(t *testing.T) {
	t0 := NewBucketManager(DefaultConfig())

	// all empty add one token to all buckets
	_ = t0.CheckAuth("user0", "password0", "ip0")
	_ = t0.CheckAuth("user1", "password1", "ip1")

	assert.Equal(t, 2, len(t0.BucketStats(t0.loginBuckets.buckets)))
	t0.ResetLogin("user0")
	assert.Equal(t, 1, len(t0.BucketStats(t0.loginBuckets.buckets)))
	assert.Equal(t, 1, t0.BucketStats(t0.loginBuckets.buckets)["user1"])
}

func TestNewBucketManager_ResetIP(t *testing.T) {
	t0 := NewBucketManager(DefaultConfig())

	// all empty add one token to all buckets
	_ = t0.CheckAuth("user0", "password0", "ip0")
	_ = t0.CheckAuth("user1", "password1", "ip1")

	assert.Equal(t, 2, len(t0.BucketStats(t0.ipBuckets.buckets)))
	t0.ResetIP("ip1")
	assert.Equal(t, 1, len(t0.BucketStats(t0.ipBuckets.buckets)))
	assert.Equal(t, 1, t0.BucketStats(t0.ipBuckets.buckets)["ip0"])
}

func TestNewBucketManager_ResetAll(t *testing.T) {
	t0 := NewBucketManager(DefaultConfig())

	// all empty add one token to all buckets
	_ = t0.CheckAuth("user0", "password0", "ip0")
	_ = t0.CheckAuth("user1", "password1", "ip1")

	assert.Equal(t, 2, len(t0.BucketStats(t0.ipBuckets.buckets)))
	t0.ResetAll("user0", "ip1")
	assert.Equal(t, 1, len(t0.BucketStats(t0.loginBuckets.buckets)))
	assert.Equal(t, 1, t0.BucketStats(t0.loginBuckets.buckets)["user1"])
	assert.Equal(t, 1, len(t0.BucketStats(t0.ipBuckets.buckets)))
	assert.Equal(t, 1, t0.BucketStats(t0.ipBuckets.buckets)["ip0"])
}
