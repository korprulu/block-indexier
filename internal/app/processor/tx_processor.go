// Package processor ...
package processor

import (
	"context"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/korprulu/interview-homework-b/internal/model"
	"github.com/korprulu/interview-homework-b/internal/pkg"
	"github.com/rs/zerolog"
)

type (
	// TxProcessor processes blocks
	TxProcessor struct {
		redisClient     *pkg.RedisClient
		ethClient       *pkg.EthClient
		dbClient        *pkg.DBClient
		logger          *zerolog.Logger
		concurrentCount int
		batchTxSize     int

		pool       *pkg.Pool[[]txRecordDTO]
		txConsumer pkg.StreamConsumer

		txConsumerStreamName string
		txConsumerGroupName  string
		txConsumerName       string

		cancelFunc context.CancelFunc
	}

	// TxProcessorConfig contains the configuration for the processor
	TxProcessorConfig struct {
		RedisClient     *pkg.RedisClient
		EthClient       *pkg.EthClient
		DBClient        *pkg.DBClient
		Logger          *zerolog.Logger
		ConcurrentCount int
		BatchTxSize     int

		TxConsumerSteamName string
		TxConsumerGroupName string
	}
)

type (
	txRecordDTO struct {
		id    string
		model *model.Transaction
	}
)

// NewTxProcessor creates a new processor
func NewTxProcessor(ctx context.Context, config TxProcessorConfig) (*TxProcessor, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	consumerName := hostname

	consumer, err := pkg.NewRedisStream(ctx, pkg.RedisStreamConfig{
		Client:       config.RedisClient,
		StreamName:   config.TxConsumerSteamName,
		GroupName:    config.TxConsumerGroupName,
		ConsumerName: consumerName,
	})
	if err != nil {
		return nil, err
	}

	processor := &TxProcessor{
		redisClient:          config.RedisClient,
		ethClient:            config.EthClient,
		dbClient:             config.DBClient,
		logger:               config.Logger,
		txConsumerStreamName: config.TxConsumerSteamName,
		txConsumerGroupName:  config.TxConsumerGroupName,
		txConsumerName:       consumerName,
		concurrentCount:      config.ConcurrentCount,
		batchTxSize:          config.BatchTxSize,
		txConsumer:           consumer,
	}

	return processor, nil
}

// FetchRecords fetches records from Redis
func (p *TxProcessor) fetchRecords(ctx context.Context) <-chan []txRecordDTO {
	ch := make(chan []txRecordDTO, p.concurrentCount)
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
				messages, err := p.txConsumer.Read(ctx, ">", p.batchTxSize)
				if err != nil {
					p.logger.Error().Err(err).Msg("failed to read stream")
					continue
				}

				txRecordDTOs := make([]txRecordDTO, 0, len(messages))
				for _, message := range messages {
					nounce, err := hexutil.DecodeUint64(message.Values["nonce"].(string))
					if err != nil {
						p.logger.Error().Err(err).Msgf("failed to decode nonce %v", message.Values["nonce"])
					}
					blockNumber, err := hexutil.DecodeUint64(message.Values["block_number"].(string))
					if err != nil {
						p.logger.Error().Err(err).Msgf("failed to decode block number %v", message.Values["block_number"])
					}
					txRecordDTOs = append(txRecordDTOs, txRecordDTO{
						id: message.ID,
						model: &model.Transaction{
							Hash:        message.Values["tx_hash"].(string),
							From:        message.Values["from"].(string),
							To:          message.Values["to"].(string),
							Nonce:       nounce,
							Data:        message.Values["data"].(string),
							Value:       message.Values["value"].(string),
							BlockHash:   message.Values["block_hash"].(string),
							BlockNumber: blockNumber,
						},
					})
				}
				ch <- txRecordDTOs
			}
		}
	}()
	return ch
}

func (p *TxProcessor) getTxReceipts(ctx context.Context, hash ...string) ([]pkg.BatchTransctionReceiptsResult, error) {
	txHashes := make([]common.Hash, len(hash))
	for i, h := range hash {
		txHashes[i] = common.HexToHash(h)
	}

	return p.ethClient.BatchTransactionReceipts(ctx, txHashes...)
}

// storeData stores block data in the database
func (p *TxProcessor) storeData(ctx context.Context, data model.Transactions) error {
	return data.Save(ctx, p.dbClient)
}

// acknowledge acknowledges the successful processing of a block
func (p *TxProcessor) acknowledge(ctx context.Context, id string) error {
	return p.redisClient.XAck(ctx, p.txConsumerStreamName, p.txConsumerGroupName, id).Err()
}

// Close closes the processor
func (p *TxProcessor) Close() {
	if p.pool != nil {
		p.pool.Stop()
	}
	p.redisClient.Close()
	p.ethClient.Close()
	p.dbClient.Close()
	p.txConsumer.Close()
}

// Start starts the processor
func (p *TxProcessor) Start(ctx context.Context) {
	newCtx, cancel := context.WithCancel(ctx)
	p.cancelFunc = cancel

	p.pool = pkg.NewPool(p.concurrentCount, func(r []txRecordDTO) {
		p.process(newCtx, r)
	})

	go func() {
		for r := range p.fetchRecords(newCtx) {
			p.pool.Add(r)
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

func (p *TxProcessor) process(ctx context.Context, records []txRecordDTO) {
	hashes := make([]string, len(records))
	models := make(model.Transactions, len(records))
	for i, r := range records {
		hashes[i] = r.model.Hash
		models[i] = r.model
	}

	receipts, err := p.getTxReceipts(ctx, hashes...)
	if err != nil {
		p.logger.Error().Err(err).Msg("failed to batch get transaction receipts")
		return
	}

	for i, r := range receipts {
		if r.Err != nil {
			p.logger.Error().Err(r.Err).Msgf("failed to get receipt in transaction %s", hashes[i])
			continue
		}

		logs := make([]model.TransactionLog, len(r.Receipt.Logs))
		for j, l := range r.Receipt.Logs {
			logs[j] = model.TransactionLog{
				Index: l.Index,
				Data:  hexutil.Encode(l.Data),
			}
		}
		models[i].Logs = logs
	}

	err = p.storeData(ctx, models)
	if err != nil {
		p.logger.Error().Err(err).Msg("failed to store data")
		return
	}

	for i, r := range records {
		// TODO handles different types of errors, some errors, we may need to
		// retry, some errors may not.
		if receipts[i].Err == nil {
			p.acknowledge(ctx, r.id)
		}
	}
}
