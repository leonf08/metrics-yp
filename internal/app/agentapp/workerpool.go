package agentapp

import (
	"context"
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
		wg      *sync.WaitGroup
	}
)

func newWorkerPool(tasks []*task, workers int) *workerPool {
	return &workerPool{
		tasks:   tasks,
		workers: workers,
		jobs:    make(chan *task, len(tasks)),
		wg:      new(sync.WaitGroup),
	}
}

func (w *workerPool) run(ctx context.Context) {
	for i := 0; i < w.workers; i++ {
		go w.work(ctx)
	}

	w.wg.Add(len(w.tasks))
	for _, task := range w.tasks {
		w.jobs <- task
	}

	close(w.jobs)

	w.wg.Wait()
}

func (w *workerPool) work(ctx context.Context) {
	for t := range w.jobs {
		t.err = t.fn(ctx)
		w.wg.Done()
	}
}
