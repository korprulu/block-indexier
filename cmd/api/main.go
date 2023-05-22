// Package main ...
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/korprulu/interview-homework-b/internal/app/api"
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

	server := api.NewServer(api.Config{
		DBClient: dbClient,
		Logger:   &logger,
	})

	go func() {
		logger.Info().Msg("starting api server")
		server.Run(cfg.API.Port)
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
	logger.Info().Msg("shutting down")

	server.Close()

	logger.Info().Msg("shutdown complete")
}
