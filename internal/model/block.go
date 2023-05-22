package model

import (
	"context"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/korprulu/interview-homework-b/internal/pkg"
)

// Block is a struct that represents a block in the Ethereum blockchain
type Block struct {
	Number     uint64 `json:"number"`
	Hash       string `json:"hash"`
	ParentHash string `json:"parent_hash"`
	Timestamp  uint64 `json:"timestamp"`
	Status     string `json:"status"`
	IsUncle    bool   `json:"is_uncle"`

	Transactions Transactions `json:"transactions,omitempty"`
}

// ToBlockModel converts an Ethereum block to a Block
func ToBlockModel(ethBlock *types.Block) *Block {
	return &Block{
		Number:     ethBlock.NumberU64(),
		Hash:       ethBlock.Hash().Hex(),
		ParentHash: ethBlock.ParentHash().Hex(),
		Timestamp:  ethBlock.Time(),
	}
}

// Save saves a block to the database
func (b *Block) Save(ctx context.Context, db *pkg.DBClient) error {
	_, err := db.ExecContext(ctx, "INSERT INTO blocks (number, hash, parent_hash, timestamp, status) VALUES ($1, $2, $3, $4, $5)", b.Number, b.Hash, b.ParentHash, b.Timestamp, b.Status)
	return err
}
