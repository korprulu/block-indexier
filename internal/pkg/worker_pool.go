package pkg

import (
	"sync"
)

type (
	// A Pool is a worker pool, which allows concurrent execution of multiple
	// jobs by a fixed number of worker goroutines.
	Pool[J any] struct {
		wg   *sync.WaitGroup
		jobs chan J
	}

	// Worker is a user-defined callback function that is executed by a worker
	// goroutine
	Worker[J any] func(J)
)

// NewPool creates a new worker pool
func NewPool[J any](size int, worker Worker[J]) *Pool[J] {
	wg := &sync.WaitGroup{}
	jobs := make(chan J, size)

	for i := 0; i < size; i++ {
		go func() {
			for job := range jobs {
				worker(job)
				wg.Done()
			}
		}()
	}

	return &Pool[J]{
		wg:   wg,
		jobs: jobs,
	}
}

// Add adds a job to the worker pool
func (p *Pool[J]) Add(job J) {
	p.wg.Add(1)
	p.jobs <- job
}

// Stop stops the worker pool
func (p *Pool[J]) Stop() {
	p.wg.Wait()
	close(p.jobs)
}
