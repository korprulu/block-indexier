package pkg

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// EthClient is a wrapper around the go-ethereum client
type EthClient struct {
	*ethclient.Client

	rpc *rpc.Client
}

// EthClientConfig is the configuration for the Ethereum client
type EthClientConfig struct {
	URL string
}

// BatchTransctionReceiptsResult is the result of a batch transaction receipts call
type BatchTransctionReceiptsResult struct {
	Receipt *types.Receipt
	Err     error
}

// NewEthClient creates a new Ethereum client
func NewEthClient(config EthClientConfig) (*EthClient, error) {
	rpcClient, err := rpc.Dial(config.URL)
	if err != nil {
		return nil, err
	}
	client := ethclient.NewClient(rpcClient)
	return &EthClient{Client: client, rpc: rpcClient}, nil
}

// BatchTransactionReceipts returns the transaction receipts for the given transaction hashes
func (c *EthClient) BatchTransactionReceipts(ctx context.Context, hash ...common.Hash) ([]BatchTransctionReceiptsResult, error) {
	batchElem := make([]rpc.BatchElem, len(hash))
	for i, h := range hash {
		batchElem[i] = rpc.BatchElem{
			Method: "eth_getTransactionReceipt",
			Args:   []any{h},
			Result: &types.Receipt{},
		}
	}

	err := c.rpc.BatchCallContext(ctx, batchElem)
	if err != nil {
		return nil, err
	}

	result := make([]BatchTransctionReceiptsResult, len(batchElem))
	for i, elem := range batchElem {
		result[i] = BatchTransctionReceiptsResult{
			Receipt: elem.Result.(*types.Receipt),
			Err:     elem.Error,
		}
	}

	return result, err
}
