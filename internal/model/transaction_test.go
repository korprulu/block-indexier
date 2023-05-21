//go:build integration

package model

import (
	"context"
	"testing"

	"github.com/korprulu/interview-homework-b/internal/config"
	"github.com/korprulu/interview-homework-b/internal/pkg"
)

func TestModelTransactionSave(t *testing.T) {
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
	models := Transactions{
		&Transaction{
			Index: uint64(1),
			Hash:  "0xabc",
			From:  "0x123",
			To:    "0x124",
			Nonce: uint64(1),
			Data:  "0x123",
			Value: "12345678",
			Logs: TransactionLogs{
				{
					Index: uint(1),
					Data:  "0x123",
				},
			},
			BlockHash:   "0x123123",
			BlockNumber: uint64(1),
		},
		&Transaction{
			Index: uint64(2),
			Hash:  "0xdea",
			From:  "0x123",
			To:    "0x124",
			Nonce: uint64(2),
			Data:  "0x123",
			Value: "12345678",
			Logs: TransactionLogs{
				{
					Index: uint(1),
					Data:  "0x123",
				},
			},
			BlockHash:   "0x123123",
			BlockNumber: uint64(1),
		},
	}

	err = models.Save(ctx, dbClient)
	if err != nil {
		t.Error(err)
	}
	defer dbClient.ExecContext(ctx, "DELETE FROM transactions WHERE hash IN ($1, $2)", models[0].Hash, models[1].Hash)

	rows, err := dbClient.QueryContext(ctx, "SELECT * FROM transactions WHERE hash IN ($1, $2)", models[0].Hash, models[1].Hash)
	if err != nil {
		t.Error(err)
	}
	defer rows.Close()

	actualModels := make(Transactions, 0)
	for rows.Next() {
		var transaction Transaction
		err = rows.Scan(&transaction.Hash, &transaction.Index, &transaction.BlockHash, &transaction.BlockNumber, &transaction.From, &transaction.To, &transaction.Nonce, &transaction.Data, &transaction.Value, &transaction.Logs)
		if err != nil {
			t.Error(err)
		}
		actualModels = append(actualModels, &transaction)
	}

	if len(actualModels) != len(models) {
		t.Errorf("Expected %d transactions, got %d", len(models), len(actualModels))
	}

	for i, actual := range actualModels {
		expected := models[i]
		if actual.Hash != expected.Hash {
			t.Errorf("Expected hash %s, got %s", expected.Hash, actual.Hash)
		}
		if actual.Index != expected.Index {
			t.Errorf("Expected index %d, got %d", expected.Index, actual.Index)
		}
		if actual.BlockHash != expected.BlockHash {
			t.Errorf("Expected block hash %s, got %s", expected.BlockHash, actual.BlockHash)
		}
		if actual.BlockNumber != expected.BlockNumber {
			t.Errorf("Expected block number %d, got %d", expected.BlockNumber, actual.BlockNumber)
		}
		if actual.From != expected.From {
			t.Errorf("Expected from %s, got %s", expected.From, actual.From)
		}
		if actual.To != expected.To {
			t.Errorf("Expected to %s, got %s", expected.To, actual.To)
		}
		if actual.Nonce != expected.Nonce {
			t.Errorf("Expected nonce %d, got %d", expected.Nonce, actual.Nonce)
		}
		if actual.Data != expected.Data {
			t.Errorf("Expected data %s, got %s", expected.Data, actual.Data)
		}
		if actual.Value != expected.Value {
			t.Errorf("Expected value %s, got %s", expected.Value, actual.Value)
		}
		if len(actual.Logs) != len(expected.Logs) {
			t.Errorf("Expected %d logs, got %d", len(expected.Logs), len(actual.Logs))
		}
	}
}
