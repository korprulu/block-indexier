// Package processor ...
package processor

import (
	"context"
	"os"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/korprulu/interview-homework-b/internal/model"
	"github.com/korprulu/interview-homework-b/internal/pkg"
	"github.com/rs/zerolog"
)

type (
	// BlockProcessor processes blocks
	BlockProcessor struct {
		redisClient     *pkg.RedisClient
		ethClient       *pkg.EthClient
		dbClient        *pkg.DBClient
		logger          *zerolog.Logger
		concurrentCount int

		pool          *pkg.Pool[blockRecordDTO]
		blockConsumer pkg.StreamConsumer
		txProducer    pkg.StreamProducer

		blockConsumerStreamName string
		blockConsumerGroupName  string
		blockConsumerName       string
		txProducerStreamName    string

		cancelFunc context.CancelFunc
	}

	// BlockProcessorConfig contains the configuration for the processor
	BlockProcessorConfig struct {
		RedisClient     *pkg.RedisClient
		EthClient       *pkg.EthClient
		DBClient        *pkg.DBClient
		Logger          *zerolog.Logger
		ConcurrentCount int

		BlockConsumerSteamName string
		BlockConsumerGroupName string
		TxProducerStreamName   string
	}
)

type (
	blockNumber string

	blockRecordDTO struct {
		number blockNumber
		id     string
		status string
	}
)

// NewBlockProcessor creates a new processor
func NewBlockProcessor(ctx context.Context, config BlockProcessorConfig) (*BlockProcessor, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	consumerName := hostname

	consumer, err := pkg.NewRedisStream(ctx, pkg.RedisStreamConfig{
		Client:       config.RedisClient,
		StreamName:   config.BlockConsumerSteamName,
		GroupName:    config.BlockConsumerGroupName,
		ConsumerName: consumerName,
	})
	if err != nil {
		return nil, err
	}

	producer, err := pkg.NewRedisStream(ctx, pkg.RedisStreamConfig{
		Client:     config.RedisClient,
		StreamName: config.TxProducerStreamName,
	})
	if err != nil {
		return nil, err
	}

	processor := &BlockProcessor{
		redisClient:             config.RedisClient,
		ethClient:               config.EthClient,
		dbClient:                config.DBClient,
		logger:                  config.Logger,
		blockConsumerStreamName: config.BlockConsumerSteamName,
		blockConsumerGroupName:  config.BlockConsumerGroupName,
		blockConsumerName:       consumerName,
		concurrentCount:         config.ConcurrentCount,
		blockConsumer:           consumer,
		txProducer:              producer,
	}

	return processor, nil
}

// FetchRecords fetches records from Redis
func (p *BlockProcessor) fetchRecords(ctx context.Context) <-chan blockRecordDTO {
	ch := make(chan blockRecordDTO, p.concurrentCount)
	go func() {
		defer close(ch)
		for {
			select {
			case <-ctx.Done():
				if err := ctx.Err(); err != nil {
					p.logger.Error().Err(err).Msg("context canceled")
				}
				return
			default:
				messages, err := p.blockConsumer.Read(ctx, ">", p.concurrentCount)
				if err != nil {
					p.logger.Error().Err(err).Msg("failed to read stream")
					continue
				}

				for _, message := range messages {
					ch <- blockRecordDTO{
						id:     message.ID,
						number: blockNumber(message.Values["number"].(string)),
						status: message.Values["status"].(string),
					}
				}
			}
		}
	}()
	return ch
}

func (p *BlockProcessor) getBlockByNumber(ctx context.Context, number blockNumber) (*model.Block, error) {
	convertedNum, err := hexutil.DecodeBig(string(number))
	if err != nil {
		return nil, err
	}

	block, err := p.ethClient.BlockByNumber(ctx, convertedNum)
	if err != nil {
		return nil, err
	}

	blockModel := model.ToBlockModel(block)
	transactions := make(model.Transactions, len(block.Transactions()))
	blockModel.Transactions = transactions

	for i, tx := range block.Transactions() {
		transactions[i], err = model.ToTransaction(ctx, p.ethClient, tx, block, i)
		if err != nil {
			p.logger.Error().Err(err).Msg("failed to convert transaction")
		}
	}

	return blockModel, nil
}

// storeData stores block data in the database
func (p *BlockProcessor) storeData(ctx context.Context, data *model.Block) error {
	return data.Save(ctx, p.dbClient)
}

// acknowledge acknowledges the successful processing of a block
func (p *BlockProcessor) acknowledge(ctx context.Context, id string) error {
	return p.redisClient.XAck(ctx, p.blockConsumerStreamName, p.blockConsumerGroupName, id).Err()
}

func (p *BlockProcessor) sendTransactions(ctx context.Context, block *model.Block) error {
	for _, tx := range block.Transactions {
		_, err := p.txProducer.Add(ctx, tx.StreamValue())
		if err != nil {
			return err
		}
	}
	return nil
}

// Close closes the processor
func (p *BlockProcessor) Close() {
	if p.cancelFunc != nil {
		p.cancelFunc()
	}
	if p.pool != nil {
		p.pool.Stop()
	}
	p.redisClient.Close()
	p.ethClient.Close()
	p.dbClient.Close()
	p.blockConsumer.Close()
}

// Start fetches records, gets block info, stores data, and acknowledges
func (p *BlockProcessor) Start(ctx context.Context) {
	newCtx, cancelFunc := context.WithCancel(ctx)
	p.cancelFunc = cancelFunc

	p.pool = pkg.NewPool(p.concurrentCount, func(record blockRecordDTO) {
		p.process(newCtx, record)
	})

	go func() {
		for record := range p.fetchRecords(newCtx) {
			p.pool.Add(record)
		}
	}()

	select {
	case <-newCtx.Done():
		if err := newCtx.Err(); err != nil {
			p.logger.Error().Err(err).Msg("context canceled")
		}
		p.Close()
		return
	}
}

func (p *BlockProcessor) process(ctx context.Context, record blockRecordDTO) {
	block, err := p.getBlockByNumber(ctx, record.number)
	if err != nil {
		p.logger.Error().Err(err).Msg("failed to get block by number")
		return
	}

	block.Status = record.status

	err = p.storeData(ctx, block)
	if err != nil {
		p.logger.Error().Err(err).Msg("failed to store data")
		return
	}

	err = p.sendTransactions(ctx, block)
	if err != nil {
		p.logger.Error().Err(err).Msg("failed to send transactions")
		return
	}

	p.acknowledge(ctx, record.id)
}
