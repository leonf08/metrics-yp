package services

import (
	"fmt"
	"testing"

	"github.com/leonf08/metrics-yp.git/internal/services/mocks"
	"github.com/leonf08/metrics-yp.git/internal/services/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStorage_Load(t *testing.T) {
	type args struct {
		r repo.Repository
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test 1, save, no error",
			args: args{
				r: repo.NewStorage(),
			},
			wantErr: false,
		},
		{
			name: "test 2, save, error",
			args: args{
				r: mocks.NewRepository(t),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := NewFileStorage("test.json")
			require.NoError(t, err, "NewFileStorage()")

			assert.Equal(t, tt.wantErr, fs.Load(tt.args.r) != nil, fmt.Sprintf("Load(%v)", tt.args.r))
		})
	}
}

func TestFileStorage_Save(t *testing.T) {
	type args struct {
		r repo.Repository
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test 1, save, no error",
			args: args{
				r: repo.NewStorage(),
			},
			wantErr: false,
		},
		{
			name: "test 2, save, error",
			args: args{
				r: mocks.NewRepository(t),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs, err := NewFileStorage("test.json")
			require.NoError(t, err, "NewFileStorage()")

			assert.Equal(t, tt.wantErr, fs.Save(tt.args.r) != nil, fmt.Sprintf("Save(%v)", tt.args.r))
		})
	}
}

func TestNewFileStorage(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test 1, new file storage, no error",
			args: args{
				path: "test.json",
			},
			wantErr: false,
		},
		{
			name: "test 2, new file storage, error",
			args: args{
				path: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFileStorage(tt.args.path)

			assert.Equal(t, tt.wantErr, err != nil, fmt.Sprintf("NewFileStorage(%v)", tt.args.path))
		})
	}
}
