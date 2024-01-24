package client

import (
	"sync"

	"go.uber.org/ratelimit"
)

type (
	task func() error

	workerPool struct {
		tasks   []task
		workers int
		jobs    chan task
		result  chan error
		limiter ratelimit.Limiter
		sync.WaitGroup
	}
)

func newWorkerPool(ts []task, ws int, l ratelimit.Limiter) *workerPool {
	return &workerPool{
		tasks:   ts,
		workers: ws,
		jobs:    make(chan task, len(ts)),
		result:  make(chan error, len(ts)),
		limiter: l,
	}
}

func (w *workerPool) run() <-chan error {
	for i := 0; i < w.workers; i++ {
		go w.work()
	}

	w.Add(len(w.tasks))
	for _, task := range w.tasks {
		w.jobs <- task
	}

	close(w.jobs)

	go func() {
		w.Wait()
		close(w.result)
	}()

	return w.result
}

func (w *workerPool) work() {
	for t := range w.jobs {
		w.limiter.Take()
		w.result <- t()
		w.Done()
	}
}
