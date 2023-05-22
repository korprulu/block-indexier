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

	blockProcessor, err := processor.NewBlockProcessor(ctx, processor.BlockProcessorConfig{
		RedisClient:            redisClient,
		EthClient:              ethClient,
		DBClient:               dbClient,
		Logger:                 &logger,
		ConcurrentCount:        cfg.BlockProcessor.ConcurrentCount,
		BlockConsumerSteamName: cfg.BlockProcessor.BlockStreamName,
		BlockConsumerGroupName: cfg.BlockProcessor.ConsumerGroup,
		TxProducerStreamName:   cfg.BlockProcessor.TransactionStreamName,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create block processor")
	}

	go func() {
		logger.Info().Msg("starting block processor")
		blockProcessor.Start(ctx)
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
	logger.Info().Msg("shutting down")

	blockProcessor.Close()

	logger.Info().Msg("shutdown complete")
}
