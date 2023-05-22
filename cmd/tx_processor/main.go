// Package main ...
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/korprulu/interview-homework-b/internal/app/processor"
	"github.com/korprulu/interview-homework-b/internal/config"
	"github.com/korprulu/interview-homework-b/internal/pkg"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.
		New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().Timestamp().
		Logger()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	redisClient := pkg.NewRedisClient(pkg.RedisClientConfig{
		Addr: cfg.Redis.Address,
		DB:   cfg.Redis.DB,
	})

	ethClient, err := pkg.NewEthClient(pkg.EthClientConfig{
		URL: cfg.Ethereum.URL,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create eth client")
	}

	dbClient, err := pkg.NewDBClient(pkg.DBClientConfig{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		Database: cfg.Postgres.DB,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create db client")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	txProcessor, err := processor.NewTxProcessor(ctx, processor.TxProcessorConfig{
		RedisClient:         redisClient,
		EthClient:           ethClient,
		DBClient:            dbClient,
		Logger:              &logger,
		ConcurrentCount:     cfg.TransactionProcessor.ConcurrentCount,
		BatchTxSize:         cfg.TransactionProcessor.BatchTransactionCount,
		TxConsumerSteamName: cfg.TransactionProcessor.TransactionStreamName,
		TxConsumerGroupName: cfg.TransactionProcessor.ConsumerGroup,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create transaction processor")
	}

	go func() {
		logger.Info().Msg("starting transaction processor")
		txProcessor.Start(ctx)
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
	logger.Info().Msg("shutting down")

	txProcessor.Close()

	logger.Info().Msg("shutdown complete")
}
