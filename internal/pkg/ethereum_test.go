//go:build integration

package pkg

import (
	"context"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

func TestBatchTransactionReceipts(t *testing.T) {
	t.Parallel()

	rpcUrl := "https://eth.llamarpc.com"

	ethClient, err := NewEthClient(EthClientConfig{
		URL: rpcUrl,
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	receipts, err := ethClient.BatchTransactionReceipts(ctx,
		common.HexToHash("0x6576804cb20d1bab7898d22eaf4fed6fec75ddaf43ef43b97f2c8011e449deef"),
		common.HexToHash("0x33708516e0c2d4aca31ecd2856e90192f8f2301c7d87285862927249ec072e6a"),
		common.HexToHash("0x454ca753b3e954b56cfcd3d56aa41d9a1f2a36eda441453a4120e3836b3f7d56"),
	)
	if err != nil {
		t.Errorf("BatchTransactionReceipts() error = %v", err)
	}

	if len(receipts) != 3 {
		t.Errorf("BatchTransactionReceipts() got = %v, want %v", len(receipts), 3)
	}
}
