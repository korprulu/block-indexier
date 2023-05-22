// Package validator ...
package validator

import (
	"context"
	"database/sql"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/korprulu/interview-homework-b/internal/model"
	"github.com/korprulu/interview-homework-b/internal/pkg"
	"github.com/rs/zerolog"
)

// Validator checks if the block has become an uncle block
type Validator struct {
	dbClient    *pkg.DBClient
	redisClient *pkg.RedisClient
	ethClient   *pkg.EthClient
	logger      *zerolog.Logger

	blockProducer pkg.StreamProducer

	blockStreamName string
	reorgCheckCount int
	watchInterval   time.Duration

	cancelFunc context.CancelFunc
}

// Config is the configuration for the validator
type Config struct {
	BlockStreamName   string
	ReorgCheckCount   int
	DBClient          *pkg.DBClient
	RedisClient       *pkg.RedisClient
	EthClient         *pkg.EthClient
	Logger            *zerolog.Logger
	WatchIntervalSecs int
}

// NewValidator creates a new validator
func NewValidator(ctx context.Context, cfg Config) (*Validator, error) {
	producer, err := pkg.NewRedisStream(ctx, pkg.RedisStreamConfig{
		Client:     cfg.RedisClient,
		StreamName: cfg.BlockStreamName,
	})
	if err != nil {
		return nil, err
	}

	return &Validator{
		dbClient:        cfg.DBClient,
		redisClient:     cfg.RedisClient,
		ethClient:       cfg.EthClient,
		logger:          cfg.Logger,
		blockStreamName: cfg.BlockStreamName,
		reorgCheckCount: cfg.ReorgCheckCount,
		blockProducer:   producer,
		watchInterval:   time.Duration(cfg.WatchIntervalSecs) * time.Second,
	}, nil
}

// Start starts the validator
func (v *Validator) Start(ctx context.Context) {
	newCtx, cancelFunc := context.WithCancel(ctx)
	v.cancelFunc = cancelFunc

	for {
		err := v.process(newCtx)
		if err != nil {
			v.logger.Error().Err(err).Msg("Failed to process")
		}
		time.Sleep(v.watchInterval)
	}
}

// Close closes the validator
func (v *Validator) Close() {
	if v.cancelFunc != nil {
		v.cancelFunc()
	}
	v.dbClient.Close()
	v.redisClient.Close()
	v.ethClient.Close()
}

func (v *Validator) process(ctx context.Context) error {
	unfinalizedBlocks, err := v.queryUnfinalizedBlocks(ctx)
	if err != nil {
		return err
	}

	finalizedBlocks, uncleBlocks, err := v.validateBlock(ctx, unfinalizedBlocks)
	if err != nil {
		return err
	}

	err = v.updateFinalizedBlocks(ctx, finalizedBlocks)
	if err != nil {
		return err
	}

	err = v.reorgUncleBlocks(ctx, uncleBlocks)
	if err != nil {
		return err
	}

	return nil
}

func (v *Validator) queryUnfinalizedBlocks(ctx context.Context) ([]*model.Block, error) {
	rows, err := v.dbClient.QueryContext(ctx, "SELECT number, hash FROM blocks WHERE status = 'unfinalized' ORDER BY timestamp DESC OFFSET $1", v.reorgCheckCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []*model.Block
	for rows.Next() {
		var block model.Block
		err = rows.Scan(&block.Number, &block.Hash)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, &block)
	}

	return blocks, nil
}

// validateBlock checks if the block has become an uncle block and returns
// block hashes that finalized and block numbers that become uncle blocks
func (v *Validator) validateBlock(ctx context.Context, blocks []*model.Block) (finalizedBlocks []*model.Block, uncleBlocks []*model.Block, err error) {
	checkingNumbers := make([]uint64, len(blocks))
	for i, block := range blocks {
		checkingNumbers[i] = block.Number
	}

	headers, err := v.ethClient.BatchHeaderByNumbers(ctx, checkingNumbers...)
	if err != nil {
		return nil, nil, err
	}

	for i, header := range headers {
		if header.Header.Hash().Hex() == blocks[i].Hash {
			finalizedBlocks = append(finalizedBlocks, blocks[i])
		} else {
			uncleBlocks = append(uncleBlocks, blocks[i])
		}
	}

	return finalizedBlocks, uncleBlocks, nil
}

func (v *Validator) updateFinalizedBlocks(ctx context.Context, blocks []*model.Block) error {
	stmt, err := v.dbClient.PrepareContext(ctx, "UPDATE blocks SET status = 'finalized' WHERE number = $1 AND hash = $2 AND status = 'unfinalized'")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, block := range blocks {
		_, err = stmt.ExecContext(ctx, block.Number, block.Hash)
		if err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) reorgUncleBlocks(ctx context.Context, uncleBlocks []*model.Block) error {
	for _, model := range uncleBlocks {
		tx, err := v.dbClient.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, "UPDATE blocks SET status = 'finalized', is_uncle = true WHERE number = $1 AND hash = $2 AND status = 'unfinalized'", model.Number, model.Hash)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				v.logger.Error().Err(rollbackErr).Msg("failed to rollback")
			}
			return err
		}

		_, err = v.blockProducer.Add(ctx, pkg.StreamValue{
			"number": hexutil.EncodeUint64(model.Number),
			"status": "finalized",
		})
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				v.logger.Error().Err(rollbackErr).Msg("failed to rollback")
			}
			return err
		}

		if err := tx.Commit(); err != nil {
			v.logger.Error().Err(err).Msg("failed to commit")
		}
	}

	return nil
}
