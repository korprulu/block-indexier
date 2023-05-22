package pkg

import (
	"database/sql"
	"fmt"

	// postgres driver
	_ "github.com/lib/pq"
)

// DBClient is a wrapper around sql.DBClient
type DBClient struct {
	*sql.DB
}

// DBClientConfig is a configuration for Postgres
type DBClientConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

// NewDBClient creates a new Postgres instance
func NewDBClient(cfg DBClientConfig) (*DBClient, error) {
	connStr := fmt.Sprintf("user=%s password='%s' host=%s port=%s dbname=%s sslmode=disable", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return &DBClient{db}, nil
}
