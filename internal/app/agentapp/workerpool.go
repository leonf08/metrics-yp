package agentapp

import (
	"context"
	"go.uber.org/ratelimit"
	"sync"
)

type (
	task struct {
		err error
		fn  func(ctx context.Context) error
	}

	workerPool struct {
		tasks   []*task
		workers int
		jobs    chan *task
		limiter ratelimit.Limiter
		sync.WaitGroup
	}
)

func newWorkerPool(ts []*task, ws int, l ratelimit.Limiter) *workerPool {
	return &workerPool{
		tasks:   ts,
		workers: ws,
		jobs:    make(chan *task, len(ts)),
		limiter: l,
	}
}

func (w *workerPool) run(ctx context.Context) {
	for i := 0; i < w.workers; i++ {
		go w.work(ctx)
	}

	w.Add(len(w.tasks))
	for _, task := range w.tasks {
		w.jobs <- task
	}

	close(w.jobs)

	w.Wait()
}

func (w *workerPool) work(ctx context.Context) {
	for t := range w.jobs {
		w.limiter.Take()
		t.err = t.fn(ctx)
		w.Done()
	}
}
