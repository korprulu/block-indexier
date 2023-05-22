package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Block is the block DTO
type Block struct {
	BlockNum   uint64 `json:"block_num"`
	BlockHash  string `json:"block_hash"`
	BlockTime  uint64 `json:"block_time"`
	ParentHash string `json:"parent_hash"`
	IsUncle    bool   `json:"is_uncle"`
}

// BlockByID is the block DTO with transactions
type BlockByID struct {
	Block
	Transactions []string `json:"transactions"`
}

// Blocks is a slice of Block
type Blocks []Block

// GetBlocks returns all blocks
func (h *Server) GetBlocks(c *gin.Context) {
	ctx := c.Request.Context()

	limit := c.DefaultQuery("limit", "10")
	rows, err := h.dbClient.QueryContext(ctx, "SELECT number, hash, timestamp, parent_hash, is_uncle FROM blocks ORDER BY number DESC LIMIT $1", limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var blocks Blocks

	for rows.Next() {
		var block Block
		err := rows.Scan(&block.BlockNum, &block.BlockHash, &block.BlockTime, &block.ParentHash, &block.IsUncle)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		blocks = append(blocks, block)
	}

	c.JSON(http.StatusOK, blocks)
}

// GetBlockByID returns a block by id
func (h *Server) GetBlockByID(c *gin.Context) {
	ctx := c.Request.Context()

	number := c.Param("id")
	row := h.dbClient.QueryRowContext(ctx, "SELECT number, hash, timestamp, parent_hash, is_uncle FROM blocks WHERE number = $1 AND is_uncle = false", number)

	var block BlockByID
	err := row.Scan(&block.BlockNum, &block.BlockHash, &block.BlockTime, &block.ParentHash, &block.IsUncle)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "block not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rows, err := h.dbClient.QueryContext(ctx, "SELECT hash FROM transactions WHERE block_number = $1 AND block_hash = $2", number, block.BlockHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var transactions []string

	for rows.Next() {
		var transaction string
		err := rows.Scan(&transaction)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		transactions = append(transactions, transaction)
	}

	block.Transactions = transactions

	c.JSON(http.StatusOK, block)
}
