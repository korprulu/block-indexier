package pkg

import (
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

// NewEthClient creates a new Ethereum client
func NewEthClient(config EthClientConfig) (*EthClient, error) {
	rpcClient, err := rpc.Dial(config.URL)
	if err != nil {
		return nil, err
	}
	client := ethclient.NewClient(rpcClient)
	return &EthClient{Client: client, rpc: rpcClient}, nil
}
