//go:build integration

package model

import (
	"context"
	"testing"
	"time"

	"github.com/korprulu/interview-homework-b/internal/pkg"
	"github.com/korprulu/interview-homework-b/internal/config"
)

func TestBlockModelSave(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}

	dbClient, err := pkg.NewDBClient(pkg.DBClientConfig{
		Host:     cfg.Postgres.Host,
		Port:     cfg.Postgres.Port,
		User:     cfg.Postgres.User,
		Password: cfg.Postgres.Password,
		Database: cfg.Postgres.DB,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer dbClient.Close()

	ctx := context.Background()
	block := &Block{
		Number:     uint64(1),
		Hash:       "0x123",
		ParentHash: "0x122",
		Timestamp:  uint64(time.Now().Unix()),
	}

	err = block.Save(ctx, dbClient)
	if err != nil {
		t.Error(err)
	}
	defer dbClient.ExecContext(ctx, "DELETE FROM blocks WHERE number = $1 AND hash = $2", block.Number, block.Hash)

	row := dbClient.QueryRowContext(ctx, "SELECT * FROM blocks WHERE number = $1 AND hash = $2", block.Number, block.Hash)

	var actual Block
	err = row.Scan(&actual.Number, &actual.Hash, &actual.ParentHash, &actual.Timestamp)
	if err != nil {
		t.Error(err)
	}

	if actual.Number != block.Number {
		t.Errorf("Expected number %d, got %d", block.Number, actual.Number)
	}
	if actual.Hash != block.Hash {
		t.Errorf("Expected hash %s, got %s", block.Hash, actual.Hash)
	}
	if actual.ParentHash != block.ParentHash {
		t.Errorf("Expected parent hash %s, got %s", block.ParentHash, actual.ParentHash)
	}
	if actual.Timestamp != block.Timestamp {
		t.Errorf("Expected timestamp %d, got %d", block.Timestamp, actual.Timestamp)
	}
}
