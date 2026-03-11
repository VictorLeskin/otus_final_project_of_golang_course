package bucket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTick_Conversion(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 30, 500000000, time.UTC)
	tick := NewTick(now)

	// Проверяем, что округлилось до секунд (отбросили наносекунды)
	assert.Equal(t, Tick(1704110430500), tick)

	// Обратное преобразование
	back := tick.ToTime()
	assert.Equal(t, int64(1704110430), back.Unix())
	assert.Equal(t, 500_000_000, back.Nanosecond()) // наносекунды
}

func TestTick_Arithmetic(t *testing.T) {
	tick := Tick(100)

	assert.Equal(t, Tick(105), tick.Add(5))
	assert.Equal(t, int64(10), tick.Sub(Tick(90)))
	assert.Equal(t, int64(-10), tick.Sub(Tick(110)))
}

func TestTick_String(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 30, 500000000, time.UTC)
	tick := NewTick(now)

	// Проверяем, что округлилось до секунд (отбросили наносекунды)
	assert.Equal(t, "15:00:30.500", tick.String())
}
