//go:build integration

package pkg

import (
	"context"
	"reflect"
	"testing"

	goredis "github.com/redis/go-redis/v9"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestRedisStreamAdd(t *testing.T) {
	t.Parallel()

	redisClient := redisClient()

	ctx := context.Background()
	streamName := t.Name()
	groupName := t.Name()
	consumerName := t.Name()

	stream := redisStream(ctx, t, redisClient, streamName, groupName, consumerName)

	_, err := stream.Add(ctx, StreamValue{
		"hello": "world",
	})
	if err != nil {
		t.Fatal(err)
	}

	respStream, err := redisClient.XRead(ctx, &goredis.XReadArgs{
		Streams: []string{streamName, "0"},
		Count:   10,
		Block:   0,
	}).Result()
	if err != nil {
		t.Fatal(err)
	}

	if len(respStream) != 1 {
		t.Fatalf("expected 1 stream, got %d", len(respStream))
	}

	if len(respStream[0].Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(respStream[0].Messages))
	}

	if respStream[0].Messages[0].Values["hello"] != "world" {
		t.Fatalf("expected hello world, got %s", respStream[0].Messages[0].Values["hello"])
	}

	redisClient.Del(ctx, streamName)
}

func TestRedisStreamRead(t *testing.T) {
	t.Parallel()

	redisClient := redisClient()

	ctx := context.Background()
	streamName := t.Name()
	groupName := t.Name()
	consumerName := t.Name()

	stream := redisStream(ctx, t, redisClient, streamName, groupName, consumerName)
	defer redisClient.Del(ctx, streamName)

	type testCase struct {
		name       string
		add        []StreamValue
		startID    string
		expCount   int
		expMessage []StreamValue
	}

	testCases := []testCase{
		{"add 1", []StreamValue{{"hello": "world"}}, ">", 1, []StreamValue{{"hello": "world"}}},
		{"add 2", []StreamValue{{"hello": "world"}, {"hello": "world2"}}, ">", 2, []StreamValue{{"hello": "world"}, {"hello": "world2"}}},
		{"get all", []StreamValue{}, "0-0", 3, []StreamValue{{"hello": "world"}, {"hello": "world"}, {"hello": "world2"}}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, msg := range tc.add {
				_, err := stream.Add(ctx, msg)
				if err != nil {
					t.Fatal(err)
				}
			}

			messages, err := stream.Read(ctx, tc.startID, 10)
			if err != nil {
				t.Fatal(err)
			}

			if len(messages) != tc.expCount {
				t.Fatalf("expected %d message, got %d", tc.expCount, len(messages))
			}

			for i, msg := range messages {
				if !reflect.DeepEqual(msg.Values, tc.expMessage[i]) {
					t.Errorf("expected %v, got %v", tc.expMessage[i], msg.Values)
				}
			}
		})
	}
}
