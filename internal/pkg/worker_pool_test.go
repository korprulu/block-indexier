package pkg

import (
	"sync/atomic"
	"testing"
)

func TestWorkerPool(t *testing.T) {
	t.Parallel()

	var count int32

	pool := NewPool(10, func(e int) {
		atomic.AddInt32(&count, 1)
	})

	pool.Add(1)
	pool.Add(2)
	pool.Add(3)

	pool.Stop()

	if count != 3 {
		t.Errorf("result length is not 3, got %d", count)
	}
}
