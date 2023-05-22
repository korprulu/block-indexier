// Package scanner ...
package scanner

import (
	"context"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/korprulu/interview-homework-b/internal/pkg"
	"github.com/rs/zerolog"
)

// Scanner scan blocks from startBlockNumber to latest block number
type Scanner struct {
	startBlockNumber uint64
	ethClient        *pkg.EthClient
	redisClient      *pkg.RedisClient
	logger           *zerolog.Logger

	blockProducer pkg.StreamProducer

	blockStreamName string
	reorgCheckCount int
	watchInterval   time.Duration

	cancelFunc context.CancelFunc
}

// Config is the config for scanner
type Config struct {
	StartBlockNumber  uint64
	EthClient         *pkg.EthClient
	RedisClient       *pkg.RedisClient
	BlockStreamName   string
	ReorgCheckCount   int
	Logger            *zerolog.Logger
	WatchIntervalSecs int
}

const latestBlockNumberKey = "latest_block_number"

// NewScanner create a new scanner
func NewScanner(ctx context.Context, cfg Config) (*Scanner, error) {
	producer, err := pkg.NewRedisStream(ctx, pkg.RedisStreamConfig{
		Client:     cfg.RedisClient,
		StreamName: cfg.BlockStreamName,
	})
	if err != nil {
		return nil, err
	}

	return &Scanner{
		startBlockNumber: cfg.StartBlockNumber,
		ethClient:        cfg.EthClient,
		redisClient:      cfg.RedisClient,
		logger:           cfg.Logger,
		blockStreamName:  cfg.BlockStreamName,
		reorgCheckCount:  cfg.ReorgCheckCount,
		blockProducer:    producer,
		watchInterval:    time.Duration(cfg.WatchIntervalSecs) * time.Second,
	}, nil
}

// Start start the scanner
func (s *Scanner) Start(ctx context.Context) error {
	newCtx, cancelFunc := context.WithCancel(ctx)
	s.cancelFunc = cancelFunc

	lastNumber, err := s.blockNumber(newCtx)
	if err != nil {
		return err
	}

	startNumber := s.startBlockNumber

	latestNumber, err := s.getLatestBlockNumber(newCtx)
	if err == nil && latestNumber > 0 {
		startNumber = latestNumber + 1
	}

	s.produce(newCtx, startNumber, lastNumber)
	if err := s.setLatestBlockNumber(newCtx, lastNumber); err != nil {
		s.logger.Error().Err(err).Msgf("failed to set latest block number %d", lastNumber)
	}
	s.watch(newCtx, lastNumber)
	return nil
}

// Close close the scanner
func (s *Scanner) Close() {
	if s.cancelFunc != nil {
		s.cancelFunc()
	}
	s.redisClient.Close()
	s.ethClient.Close()
}

func (s *Scanner) blockNumber(ctx context.Context) (uint64, error) {
	return s.ethClient.BlockNumber(ctx)
}

func (s *Scanner) produce(ctx context.Context, startNumber, lastNumber uint64) {
	for i := startNumber; i <= lastNumber; i++ {
		select {
		case <-ctx.Done():
			if err := s.setLatestBlockNumber(ctx, i-1); err != nil {
				s.logger.Error().Err(err).Msgf("failed to set latest block number %d", i)
			}
			if err := ctx.Err(); err != nil {
				s.logger.Error().Err(err).Msg("scanner context done")
			}
			return
		default:
			num := hexutil.EncodeUint64(i)
			status := "finalized"
			if i+uint64(s.reorgCheckCount) >= lastNumber {
				status = "unfinalized"
			}
			_, err := s.blockProducer.Add(ctx, pkg.StreamValue{
				"number": num,
				"status": status,
			})
			if err != nil {
				s.logger.Error().Err(err).Msgf("failed to add block %s to stream", num)
			}
		}
	}
}

func (s *Scanner) watch(ctx context.Context, lastNumber uint64) {
	for {
		select {
		case <-ctx.Done():
			if err := s.setLatestBlockNumber(ctx, lastNumber); err != nil {
				s.logger.Error().Err(err).Msgf("failed to set latest block number %d", lastNumber)
			}
			if err := ctx.Err(); err != nil {
				s.logger.Error().Err(err).Msg("scanner context done")
			}
			return
		default:
			num, err := s.blockNumber(ctx)
			if err != nil {
				s.logger.Error().Err(err).Msg("failed to get block number")
				continue
			}
			if num > lastNumber {
				s.produce(ctx, lastNumber+1, num)
				lastNumber = num
			}
			time.Sleep(s.watchInterval)
		}
	}
}

func (s *Scanner) setLatestBlockNumber(ctx context.Context, latestNumber uint64) error {
	return s.redisClient.Set(ctx, latestBlockNumberKey, latestNumber, 0).Err()
}

func (s *Scanner) getLatestBlockNumber(ctx context.Context) (uint64, error) {
	return s.redisClient.Get(ctx, latestBlockNumberKey).Uint64()
}
