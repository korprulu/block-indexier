package pkg

import (
	"context"
	"strings"
	"time"

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

	err := cfg.Client.XGroupCreateMkStream(ctx, cfg.StreamName, cfg.GroupName, "$").Err()
	if err != nil {
		if !strings.HasPrefix(err.Error(), "BUSYGROUP") {
			return nil, err
		}
	}

	err = cfg.Client.XAutoClaim(ctx, &goredis.XAutoClaimArgs{
		Stream:   cfg.StreamName,
		Group:    cfg.GroupName,
		Consumer: cfg.ConsumerName,
		MinIdle:  time.Minute * 10,
		Start:    "0-0",
	}).Err()
	if err != nil {
		return nil, err
	}

	return stream, nil
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
				ID:     message.ID,
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
		Values: map[string]any(values),
	}).Result()
}

// Close closes the stream
func (r *RedisStream) Close() {
	// noop
	return
}

var (
	_ StreamProducer = (*RedisStream)(nil)
	_ StreamConsumer = (*RedisStream)(nil)
	_ Stream         = (*RedisStream)(nil)
)
