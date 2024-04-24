package workerpool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/ratelimit"
)

func Test_workerPool_run(t *testing.T) {
	type fields struct {
		tasks   []Task
		workers int
		jobs    chan Task
		result  chan error
		limiter ratelimit.Limiter
	}
	tests := []struct {
		name   string
		fields fields
		want   error
	}{
		{
			name: "Test_workerPool_run",
			fields: fields{
				tasks: []Task{
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
				},
				workers: 3,
				jobs:    make(chan Task, 30),
				result:  make(chan error, 30),
				limiter: ratelimit.New(1),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &WorkerPool{
				tasks:   tt.fields.tasks,
				workers: tt.fields.workers,
				jobs:    tt.fields.jobs,
				result:  tt.fields.result,
				limiter: tt.fields.limiter,
			}
			got := w.Run()
			for c := range got {
				assert.ErrorIs(t, c, tt.want)
			}
		})
	}
}

func Test_newWorkerPool(t *testing.T) {
	type args struct {
		ts []Task
		ws int
		l  ratelimit.Limiter
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test_newWorkerPool",
			args: args{
				ts: []Task{
					func() error { return nil },
					func() error { return nil },
					func() error { return nil },
				},
				ws: 3,
				l:  ratelimit.New(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewWorkerPool(tt.args.ts, tt.args.ws, tt.args.l)
			assert.NotNil(t, got.tasks)
			assert.Equal(t, tt.args.ws, got.workers)
			assert.NotNil(t, got.jobs)
			assert.NotNil(t, got.result)
		})
	}
}
