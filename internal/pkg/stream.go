package pkg

import (
	"context"
)

type (
	// StreamValue is the type for a stream value
	StreamValue map[string]any

	// StreamMessage is the type for a stream message
	StreamMessage struct {
		ID     string
		Values StreamValue
	}
)

type (
	// StreamProducer is the interface for a stream producer
	StreamProducer interface {
		Add(ctx context.Context, value StreamValue) (string, error)
	}

	// StreamConsumer is the interface for a stream consumer
	StreamConsumer interface {
		Read(ctx context.Context, id string, count int) ([]StreamMessage, error)
		Ack(ctx context.Context, id string) error
		Close(ctx context.Context) error
	}

	// Stream is the interface for a stream
	Stream interface {
		StreamProducer
		StreamConsumer
	}
)
