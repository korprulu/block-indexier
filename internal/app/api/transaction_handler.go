package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/korprulu/interview-homework-b/internal/model"
)

// Transaction is a DTO for a transaction
type Transaction struct {
	TxHash string           `json:"tx_hash"`
	From   string           `json:"from"`
	To     string           `json:"to"`
	Nonce  uint64           `json:"nonce"`
	Data   string           `json:"data"`
	Value  string           `json:"value"`
	Logs   []TransactionLog `json:"logs"`
}

// TransactionLog is a DTO for a transaction log
type TransactionLog struct {
	Index uint   `json:"index"`
	Data  string `json:"data"`
}

// GetTransactionByHash returns a transaction by hash
func (h *Server) GetTransactionByHash(c *gin.Context) {
	ctx := c.Request.Context()

	txHash := c.Param("txHash")
	row := h.dbClient.QueryRowContext(ctx, "SELECT hash, from_address, to_address, nonce, data, value, logs FROM transactions WHERE hash = $1", txHash)

	var tx model.Transaction
	err := row.Scan(&tx.Hash, &tx.From, &tx.To, &tx.Nonce, &tx.Data, &tx.Value, &tx.Logs)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "transaction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	txDTO := Transaction{
		TxHash: tx.Hash,
		From:   tx.From,
		To:     tx.To,
		Nonce:  tx.Nonce,
		Data:   tx.Data,
		Value:  tx.Value,
		Logs:   make([]TransactionLog, len(tx.Logs)),
	}
	for i, log := range tx.Logs {
		txDTO.Logs[i] = TransactionLog{
			Index: log.Index,
			Data:  log.Data,
		}
	}

	c.JSON(http.StatusOK, txDTO)
}
