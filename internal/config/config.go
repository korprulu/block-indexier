// Package config ...
package config

import (
	"github.com/ilyakaznacheev/cleanenv"

	// autoload .env file
	_ "github.com/joho/godotenv/autoload"
)

// Postgres ...
type Postgres struct {
	Host     string `env:"POSTGRES_HOST" env-default:"localhost"`
	Port     string `env:"POSTGRES_PORT" env-default:"5432"`
	User     string `env:"POSTGRES_USER" env-default:"postgres"`
	Password string `env:"POSTGRES_PASSWORD" env-default:"postgres"`
	DB       string `env:"POSTGRES_DB" env-default:"postgres"`
}

// Config ...
type Config struct {
	Postgres Postgres
}

var config *Config

// Load loads the config
func Load() (*Config, error) {
	if config != nil {
		return config, nil
	}

	cfg := Config{}
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	config = &cfg

	return config, nil
}
