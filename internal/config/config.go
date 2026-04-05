package config

import (
	"encoding/json"
	"flag"
	"os"
	"time"
)

type Config struct {
	// Server settings.
	Port string `json:"port"`

	// Storage settings.
	StorageType string `json:"storage_type"` // "memory" or "postgres"

	// PostgreSQL settings
	PostgresHost     string `json:"postgres_host"`
	PostgresPort     int    `json:"postgres_port"`
	PostgresUser     string `json:"postgres_user"`
	PostgresPassword string `json:"postgres_password"`
	PostgresDB       string `json:"postgres_db"`
	PostgresSSLMode  string `json:"postgres_sslmode"`

	// Rate limiter settings
	LoginRate       int `json:"login_rate"`       // N
	PasswordRate    int `json:"password_rate"`    // M
	IPRate          int `json:"ip_rate"`          // K
	CleanupInterval int `json:"cleanup_interval"` // seconds
}

func GetDefault() *Config {
	return &Config{
		// Server settings.
		Port: "8080",

		StorageType: "memory",

		// Rate limiter settings
		LoginRate:       10,
		PasswordRate:    100,
		IPRate:          1000,
		CleanupInterval: 300, // 5 minutes
	}
}

func setMissedToDefault(cfg *Config) {
	def := GetDefault()
	if cfg.Port == "" {
		cfg.Port = def.Port
	}
	if cfg.StorageType == "" {
		cfg.StorageType = def.StorageType
	}
	if cfg.LoginRate == 0 {
		cfg.LoginRate = def.LoginRate
	}
	if cfg.PasswordRate == 0 {
		cfg.PasswordRate = def.PasswordRate
	}
	if cfg.IPRate == 0 {
		cfg.IPRate = def.IPRate
	}
	if cfg.CleanupInterval == 0 {
		cfg.CleanupInterval = def.CleanupInterval
	}
}

func Load() (*Config, error) {
	// Флаги командной строки.
	configPath := flag.String("config", "config.json", "path to config file")
	flag.Parse()

	// Загружаем из файла.
	file, err := os.Open(*configPath)
	if err != nil {
		// Если файл не найден, возвращаем default конфиг.
		if os.IsNotExist(err) {
			cfg := GetDefault()
			return cfg, nil
		}
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err := json.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}

	// Значения по умолчанию.
	setMissedToDefault(&cfg)

	return &cfg, nil
}

// Duration helpers
func (c *Config) GetCleanupInterval() time.Duration {
	return time.Duration(c.CleanupInterval) * time.Second
}
