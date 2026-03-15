package bucket

import (
	"time"
)

// Tick представляет собой дискретный момент времени в миллисекундах.
type Tick int64

// NewTick создает Tick из time.Time (округляет до миллисекунд).
func NewTick(t time.Time) Tick {
	return Tick(t.UnixNano() / int64(time.Millisecond))
}

// NowTick возвращает текущий Tick.
func NowTick() Tick {
	return Tick(time.Now().UnixNano() / int64(time.Millisecond))
}

// ToTime преобразует Tick обратно в time.Time .
func (t Tick) ToTime() time.Time {
	seconds := int64(t) / 1000
	nanoseconds := (int64(t) % 1000) * int64(time.Millisecond)
	return time.Unix(seconds, nanoseconds)
}

// Add возвращает новый Tick, увеличенный на указанное количество миллисекунд.
func (t Tick) Add(milliseconds int64) Tick {
	return t + Tick(milliseconds)
}

// Sub возвращает разницу в миллисекундах между двумя Tick.
func (t Tick) Sub(other Tick) int64 {
	return int64(t - other)
}

// String возвращает строковое представление Tick (для отладки).
func (t Tick) String() string {
	return t.ToTime().UTC().Format("15:04:05.000")
}
