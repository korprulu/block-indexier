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

// Redis ...
type Redis struct {
	Address string `env:"REDIS_ADDRESS" env-default:"localhost:6379"`
	DB      int    `env:"REDIS_DB" env-default:"0"`
}

// Ethereum ...
type Ethereum struct {
	URL string `env:"ETHEREUM_RPC_URL" env-default:"http://localhost:8545"`
}

// BlockProcessor ...
type BlockProcessor struct {
	BlockStreamName       string `env:"BLOCK_STREAM_NAME" env-default:"blocks"`
	ConsumerGroup         string `env:"BLOCK_PROCESSOR_CONSUMER_GROUP" env-default:"block-processors"`
	TransactionStreamName string `env:"TRANSACTION_STREAM_NAME" env-default:"transactions"`
	ConcurrentCount       int    `env:"BLOCK_PROCESSOR_CONCURRENT_COUNT" env-default:"10"`
}

// TransactionProcessor ...
type TransactionProcessor struct {
	TransactionStreamName string `env:"TRANSACTION_STREAM_NAME" env-default:"transactions"`
	ConsumerGroup         string `env:"TRANSACTION_PROCESSOR_CONSUMER_GROUP" env-default:"transaction-processors"`
	ConcurrentCount       int    `env:"TRANSACTION_PROCESSOR_CONCURRENT_COUNT" env-default:"10"`
	BatchTransactionCount int    `env:"TRANSACTION_PROCESSOR_BATCH_TRANSACTION_COUNT" env-default:"100"`
}

// Scanner ...
type Scanner struct {
	BlockStreamName   string `env:"BLOCK_STREAM_NAME" env-default:"blocks"`
	ReorgCheckCount   int    `env:"BLOCK_REORG_CHECK_COUNT" env-default:"50"`
	StartBlockNumber  uint64 `env:"SCANNER_START_BLOCK_NUMBER" env-default:"0"`
	WatchIntervalSecs int    `env:"SCANNER_WATCH_INTERVAL_SECONDS" env-default:"300"`
}

// Validator ...
type Validator struct {
	BlockStreamName   string `env:"BLOCK_STREAM_NAME" env-default:"blocks"`
	ReorgCheckCount   int    `env:"BLOCK_REORG_CHECK_COUNT" env-default:"50"`
	WatchIntervalSecs int    `env:"VALIDATOR_WATCH_INTERVAL_SECONDS" env-default:"300"`
}

// API ...
type API struct {
	Port string `env:"API_PORT" env-default:"8080"`
}

// Config ...
type Config struct {
	Postgres             Postgres
	Redis                Redis
	Ethereum             Ethereum
	BlockProcessor       BlockProcessor
	TransactionProcessor TransactionProcessor
	Scanner              Scanner
	Validator            Validator
	API                  API
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
