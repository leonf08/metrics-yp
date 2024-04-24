package workerpool

import (
	"sync"

	"go.uber.org/ratelimit"
)

type (
	Task func() error

	WorkerPool struct {
		tasks   []Task
		workers int
		jobs    chan Task
		result  chan error
		limiter ratelimit.Limiter
		sync.WaitGroup
	}
)

func NewWorkerPool(ts []Task, ws int, l ratelimit.Limiter) *WorkerPool {
	return &WorkerPool{
		tasks:   ts,
		workers: ws,
		jobs:    make(chan Task, len(ts)),
		result:  make(chan error, len(ts)),
		limiter: l,
	}
}

func (w *WorkerPool) Run() <-chan error {
	for i := 0; i < w.workers; i++ {
		go w.work()
	}

	w.Add(len(w.tasks))
	for _, t := range w.tasks {
		w.jobs <- t
	}

	close(w.jobs)

	go func() {
		w.Wait()
		close(w.result)
	}()

	return w.result
}

func (w *WorkerPool) work() {
	for t := range w.jobs {
		w.limiter.Take()
		w.result <- t()
		w.Done()
	}
}
