package pkg

import (
	"context"
	"testing"
)

func redisClient() *RedisClient {
	return NewRedisClient(RedisClientConfig{
		Addr: "localhost:6379",
	})
}

func redisStream(ctx context.Context, t *testing.T, redisClient *RedisClient, streamName, groupName, consumerName string) *RedisStream {
	stream, err := NewRedisStream(ctx, RedisStreamConfig{
		Client:       redisClient,
		StreamName:   streamName,
		GroupName:    groupName,
		ConsumerName: consumerName,
	})
	if err != nil {
		t.Fatal(err)
	}

	return stream
}
