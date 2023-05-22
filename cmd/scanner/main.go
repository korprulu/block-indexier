// Package main ...
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/korprulu/interview-homework-b/internal/app/scanner"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scannerInstance, err := scanner.NewScanner(ctx, scanner.Config{
		StartBlockNumber:  cfg.Scanner.StartBlockNumber,
		EthClient:         ethClient,
		RedisClient:       redisClient,
		BlockStreamName:   cfg.Scanner.BlockStreamName,
		ReorgCheckCount:   cfg.Scanner.ReorgCheckCount,
		Logger:            &logger,
		WatchIntervalSecs: cfg.Scanner.WatchIntervalSecs,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to create scanner")
	}

	go func() {
		logger.Info().Msg("starting scanner")
		scannerInstance.Start(ctx)
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
	logger.Info().Msg("shutting down")

	scannerInstance.Close()

	logger.Info().Msg("shutdown complete")
}
