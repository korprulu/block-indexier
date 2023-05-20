package pkg

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

type (
	// RedisStream is a stream implementation using Redis
	RedisStream struct {
		client *RedisClient

		streamName   string
		groupName    string
		consumerName string
	}

	// RedisStreamConfig is the config for a RedisStream
	RedisStreamConfig struct {
		StreamName   string
		GroupName    string
		ConsumerName string
		Client       *RedisClient
	}
)

// NewRedisStream creates a new RedisStream
func NewRedisStream(ctx context.Context, cfg RedisStreamConfig) (*RedisStream, error) {
	stream := &RedisStream{
		client:       cfg.Client,
		streamName:   cfg.StreamName,
		groupName:    cfg.GroupName,
		consumerName: cfg.ConsumerName,
	}

	if err := stream.registerConsumer(ctx); err != nil {
		return nil, err
	}

	return stream, nil
}

func (r *RedisStream) registerConsumer(ctx context.Context) error {
	redisClient := r.client
	if err := redisClient.XGroupCreateMkStream(ctx, r.streamName, r.groupName, "$").Err(); err != nil {
		return err
	}

	return redisClient.XGroupCreateConsumer(ctx, r.streamName, r.groupName, r.consumerName).Err()
}

func (r *RedisStream) deregisterConsumer(ctx context.Context) error {
	return r.client.XGroupDelConsumer(ctx, r.streamName, r.groupName, r.consumerName).Err()
}

// Read reads messages from the stream
func (r *RedisStream) Read(ctx context.Context, id string, count int) ([]StreamMessage, error) {
	streams, err := r.client.XReadGroup(ctx, &goredis.XReadGroupArgs{
		Group:    r.groupName,
		Consumer: r.consumerName,
		Streams:  []string{r.streamName, id},
		Count:    int64(count),
		Block:    0,
	}).Result()
	if err != nil {
		return nil, err
	}

	result := make([]StreamMessage, 0)

	for _, stream := range streams {
		for _, message := range stream.Messages {
			result = append(result, StreamMessage{
				ID:    message.ID,
				Values: message.Values,
			})
		}
	}

	return result, nil
}

// Ack acknowledges a message
func (r *RedisStream) Ack(ctx context.Context, id string) error {
	return r.client.XAck(ctx, r.streamName, r.groupName, id).Err()
}

// Add adds a message to the stream
func (r *RedisStream) Add(ctx context.Context, values StreamValue) (string, error) {
	return r.client.XAdd(ctx, &goredis.XAddArgs{
		Stream: r.streamName,
		ID:     "*",
		Values: values,
	}).Result()
}

// Close closes the stream consumer
func (r *RedisStream) Close(ctx context.Context) error {
	return r.deregisterConsumer(ctx)
}

var (
	_ StreamProducer = (*RedisStream)(nil)
	_ StreamConsumer = (*RedisStream)(nil)
	_ Stream         = (*RedisStream)(nil)
)
