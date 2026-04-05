package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetMissedToDefault(t *testing.T) {
	t.Run("пустой конфиг - заполняется значениями по умолчанию", func(t *testing.T) {
		cfg := &Config{}

		setMissedToDefault(cfg)

		expected := GetDefault()
		assert.Equal(t, expected.Port, cfg.Port)
		assert.Equal(t, expected.StorageType, cfg.StorageType)
		assert.Equal(t, expected.LoginRate, cfg.LoginRate)
		assert.Equal(t, expected.PasswordRate, cfg.PasswordRate)
		assert.Equal(t, expected.IPRate, cfg.IPRate)
		assert.Equal(t, expected.CleanupInterval, cfg.CleanupInterval)
		assert.Equal(t, expected.MaxIdleTime, cfg.MaxIdleTime)
	})

	t.Run("полностью заполненный конфиг - значения не меняются", func(t *testing.T) {
		cfg := &Config{
			Port:            "9090",
			StorageType:     "postgres",
			LoginRate:       20,
			PasswordRate:    200,
			IPRate:          2000,
			CleanupInterval: 600,
			MaxIdleTime:     1200,
		}

		original := *cfg
		setMissedToDefault(cfg)

		assert.Equal(t, original, *cfg)
	})
}
