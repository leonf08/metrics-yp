package client

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/ratelimit"
	"testing"
)

func Test_workerPool_run(t *testing.T) {
	type fields struct {
		tasks   []task
		workers int
		jobs    chan task
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
				tasks: []task{
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
				jobs:    make(chan task, 30),
				result:  make(chan error, 30),
				limiter: ratelimit.New(1),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &workerPool{
				tasks:   tt.fields.tasks,
				workers: tt.fields.workers,
				jobs:    tt.fields.jobs,
				result:  tt.fields.result,
				limiter: tt.fields.limiter,
			}
			got := w.run()
			for c := range got {
				assert.ErrorIs(t, c, tt.want)
			}
		})
	}
}
