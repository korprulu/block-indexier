// Package api ...
package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/korprulu/interview-homework-b/internal/pkg"
	"github.com/rs/zerolog"
)

// Server is the handler for the API
type Server struct {
	dbClient *pkg.DBClient
	server   *http.Server
	logger   *zerolog.Logger
}

// Config is the config for the handler
type Config struct {
	Logger   *zerolog.Logger
	DBClient *pkg.DBClient
}

// NewServer creates a new handler
func NewServer(cfg Config) *Server {
	return &Server{
		dbClient: cfg.DBClient,
		logger:   cfg.Logger,
	}
}

// Run runs the API
func (s *Server) Run(port string) {
	router := gin.Default()

	router.GET("/blocks", s.GetBlocks)
	router.GET("/blocks/:id", s.GetBlockByID)
	router.GET("/transaction/:txHash", s.GetTransactionByHash)

	s.server = &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatal().Err(err).Msg("failed to listen and serve")
	}
}

// Close closes the API
func (s *Server) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Fatal().Err(err).Msg("failed to shutdown server")
	}
	// wait for context timeout
	select {
	case <-ctx.Done():
	}
	s.dbClient.Close()
}
