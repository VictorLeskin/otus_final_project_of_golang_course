package bucket

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucket_Allow(t *testing.T) {
	initTickMs := Tick(32000) // 32 sec
	// Ведро на 3 запроса, утечка 1 запрос в секунду
	t0 := NewBucket("test", 3, 1000)

	tick := initTickMs // начальное время

	// Наполняем ведро
	assert.True(t, t0.Allow(tick)) // water=1
	assert.Equal(t, Tick(32_000), t0.lastAccess)
	assert.Equal(t, Tick(32_000), t0.lastLeak)
	assert.Equal(t, 1, t0.drops)

	assert.True(t, t0.Allow(tick)) // water=2
	assert.Equal(t, Tick(32_000), t0.lastAccess)
	assert.Equal(t, Tick(32_000), t0.lastLeak)
	assert.Equal(t, 2, t0.drops)

	assert.True(t, t0.Allow(tick)) // water=3
	assert.Equal(t, Tick(32_000), t0.lastAccess)
	assert.Equal(t, Tick(32_000), t0.lastLeak)
	assert.Equal(t, 3, t0.drops)

	assert.False(t, t0.Allow(tick+1)) // water=3, переполнено
	assert.Equal(t, Tick(32_001), t0.lastAccess)
	assert.Equal(t, Tick(32_000), t0.lastLeak)
	assert.Equal(t, 3, t0.drops)

	// Прошла 1 секунда - утекла 1 единица
	tick = tick.Add(1000)
	assert.True(t, t0.Allow(tick)) // water=2 -> +1 = water=3
	assert.Equal(t, Tick(33_000), t0.lastAccess)
	assert.Equal(t, Tick(33_000), t0.lastLeak)
	assert.Equal(t, 3, t0.drops)

	// Прошло 3 секунды - утекло 3 единицы
	tick = tick.Add(3 * 1000)
	assert.True(t, t0.Allow(tick)) // water=0 -> +1 = water=1
	assert.Equal(t, Tick(36_000), t0.lastAccess)
	assert.Equal(t, Tick(36_000), t0.lastLeak)
	assert.Equal(t, 1, t0.drops)

	assert.True(t, t0.Allow(tick))  // water=2
	assert.True(t, t0.Allow(tick))  // water=3
	assert.False(t, t0.Allow(tick)) // water=3
}

func TestBucket_AllowRealTime(t *testing.T) {
	initTickMs := Tick(1773483538422)
	// Ведро на 3 запроса, утечка 1 запрос в секунду
	t0 := NewBucket("test", 3, 1000)

	tick := initTickMs // начальное время

	// Наполняем ведро
	assert.True(t, t0.Allow(tick)) // water=1
	assert.Equal(t, 1, t0.drops)
	assert.True(t, t0.Allow(tick+1000)) // water=1
	assert.Equal(t, 1, t0.drops)
}

func TestBucket_leak(t *testing.T) {
	var t0 Bucket

	t0.leakRate = 100 // allowed 1 drop per 100 ms

	t0.lastLeak = Tick(32_000)
	t0.drops = 3

	t0.leak(Tick(32_099))
	assert.Equal(t, Tick(32_000), t0.lastLeak)
	assert.Equal(t, 3, t0.drops)

	t0.leak(Tick(32_101))
	assert.Equal(t, Tick(32_101), t0.lastLeak)
	assert.Equal(t, 2, t0.drops)

	t0.leak(Tick(32_201))
	assert.Equal(t, Tick(32_201), t0.lastLeak)
	assert.Equal(t, 1, t0.drops)

	// dropped out all
	t0.leak(Tick(32_401))
	assert.Equal(t, Tick(32_401), t0.lastLeak)
	assert.Equal(t, 0, t0.drops)

	// dropped out from empty bucket
	t0.leak(Tick(32_501))
	assert.Equal(t, Tick(32_501), t0.lastLeak)
	assert.Equal(t, 0, t0.drops)
}

func TestBucket_NewBucketFromRPM(t *testing.T) {
	t0 := NewBucketFromRPM("K0", 100)

	assert.Equal(t, "K0", t0.Key())
	assert.Equal(t, 0, t0.WaterLevel())
	assert.Equal(t, 100, t0.Remaining())
	assert.Equal(t, 600, t0.leakRate) // 100 tockens per minuter = 600 ms per tocken
}
