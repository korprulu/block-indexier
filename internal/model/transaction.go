package model

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/korprulu/interview-homework-b/internal/pkg"
)

// Transaction is a struct that represents a transaction in the Ethereum blockchain
type Transaction struct {
	Index       uint64          `json:"index"`
	Hash        string          `json:"tx_hash"`
	From        string          `json:"from"`
	To          string          `json:"to"`
	Nonce       uint64          `json:"nonce"`
	Data        string          `json:"data"`
	Value       string          `json:"value"`
	Logs        TransactionLogs `json:"logs"` // get from receipt
	BlockHash   string          `json:"block_hash"`
	BlockNumber uint64          `json:"block_number"`
}

// StreamValue returns the value of the transaction as a StreamValue
func (tx *Transaction) StreamValue() pkg.StreamValue {
	return pkg.StreamValue{
		"index":        tx.Index,
		"tx_hash":      tx.Hash,
		"from":         tx.From,
		"to":           tx.To,
		"nonce":        hexutil.EncodeUint64(tx.Nonce),
		"data":         tx.Data,
		"value":        tx.Value,
		"block_hash":   tx.BlockHash,
		"block_number": hexutil.EncodeUint64(tx.BlockNumber),
	}
}

// TransactionLog is a struct that represents a transaction log in the Ethereum blockchain
type TransactionLog struct {
	Index uint   `json:"index"`
	Data  string `json:"data"`
}

// TransactionLogs is a slice of TransactionLog
type TransactionLogs []TransactionLog

// Value returns the value of the transaction log as a driver.Value
func (t TransactionLogs) Value() (driver.Value, error) {
	return json.Marshal(t)
}

// Scan scans the value into a TransactionLog
func (t *TransactionLogs) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return json.Unmarshal(src, t)
	case string:
		return json.Unmarshal([]byte(src), t)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
}

// Transactions is a slice of Transaction
type Transactions []*Transaction

// ToTransaction converts an Ethereum transaction to a Transaction
func ToTransaction(ctx context.Context, ethClient *pkg.EthClient, tx *types.Transaction, block *types.Block, index int) (*Transaction, error) {
	from, err := ethClient.TransactionSender(ctx, tx, block.Hash(), uint(index))
	if err != nil {
		return nil, err
	}
	toAddr := tx.To()
	var to string
	if toAddr != nil {
		to = toAddr.Hex()
	}
	valueAddr := tx.Value()
	var value string
	if valueAddr != nil {
		value = valueAddr.String()
	}
	return &Transaction{
		Index:       uint64(index),
		Hash:        tx.Hash().Hex(),
		From:        from.Hex(),
		To:          to,
		Nonce:       tx.Nonce(),
		Data:        hexutil.Encode(tx.Data()),
		Value:       value,
		BlockHash:   block.Hash().Hex(),
		BlockNumber: block.Number().Uint64(),
	}, nil
}

// Save saves a slice of Transaction to the database
func (txs Transactions) Save(ctx context.Context, db *pkg.DBClient) error {
	if len(txs) == 0 {
		return nil
	}
	statement := "INSERT INTO transactions (hash, index, from_address, to_address, nonce, data, value, logs, block_hash, block_number) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)"
	args := []any{txs[0].Hash, txs[0].Index, txs[0].From, txs[0].To, txs[0].Nonce, txs[0].Data, txs[0].Value, txs[0].Logs, txs[0].BlockHash, txs[0].BlockNumber}
	for i := 1; i < len(txs); i++ {
		statement += fmt.Sprintf(", ($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", (i*10)+1, (i*10)+2, (i*10)+3, (i*10)+4, (i*10)+5, (i*10)+6, (i*10)+7, (i*10)+8, (i*10)+9, (i*10)+10)
		tx := txs[i]
		args = append(args, tx.Hash, tx.Index, tx.From, tx.To, tx.Nonce, tx.Data, tx.Value, tx.Logs, tx.BlockHash, tx.BlockNumber)
	}
	_, err := db.ExecContext(ctx, statement, args...)
	return err
}
