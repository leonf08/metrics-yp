package repo

import (
	"context"
	"reflect"
	"runtime"
	"testing"

	"github.com/leonf08/metrics-yp.git/internal/models"
)

func TestMemStorage_GetVal(t *testing.T) {
	type args struct {
		k string
	}
	tests := []struct {
		name    string
		args    args
		want    models.Metric
		wantErr bool
	}{
		{
			name: "test 1, valid data",
			args: args{
				k: "test",
			},
			want: models.Metric{
				Type: "gauge",
				Val:  1,
			},
			wantErr: false,
		},
		{
			name: "test 2, invalid data",
			args: args{
				k: "test2",
			},
			want:    models.Metric{},
			wantErr: true,
		},
	}

	st := &MemStorage{
		Storage: make(map[string]models.Metric, 1),
	}

	st.Storage["test"] = models.Metric{
		Type: "gauge",
		Val:  1,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := st.GetVal(context.Background(), tt.args.k)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetVal() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_ReadAll(t *testing.T) {
	type args struct {
		k string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]models.Metric
		wantErr bool
	}{
		{
			name: "test 1, valid data",
			args: args{
				k: "test",
			},
			want: map[string]models.Metric{
				"test": {
					Type: "gauge",
					Val:  1,
				},
			},
			wantErr: false,
		},
	}

	st := &MemStorage{
		Storage: make(map[string]models.Metric, 1),
	}

	st.Storage["test"] = models.Metric{
		Type: "gauge",
		Val:  1,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := st.ReadAll(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReadAll() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMemStorage_SetVal(t *testing.T) {
	type args struct {
		k string
		m models.Metric
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test 1, gauge data",
			args: args{
				k: "test",
				m: models.Metric{
					Type: "gauge",
					Val:  1,
				},
			},
			wantErr: false,
		},
		{
			name: "test 2, counter data, existing key",
			args: args{
				k: "test2",
				m: models.Metric{
					Type: "counter",
					Val:  int64(1),
				},
			},
			wantErr: false,
		},
		{
			name: "test 3, counter data, new key",
			args: args{
				k: "test3",
				m: models.Metric{
					Type: "counter",
					Val:  1,
				},
			},
			wantErr: false,
		},
		{
			name: "test 4, invalid type of data",
			args: args{
				k: "test3",
				m: models.Metric{
					Type: "invalid",
					Val:  1,
				},
			},
			wantErr: true,
		},
	}

	st := &MemStorage{
		Storage: make(map[string]models.Metric, 1),
	}

	st.Storage["test2"] = models.Metric{
		Type: "counter",
		Val:  int64(1),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := st.SetVal(context.Background(), tt.args.k, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("SetVal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemStorage_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test 1, valid data",
			args: args{
				data: []byte(`{"metrics":{"test":{"type":"gauge","value":1}, "test2":{"type":"counter","value":1}}}`),
			},
			wantErr: false,
		},
		{
			name: "test 2, invalid json",
			args: args{
				data: []byte(`{"metrics":{"test":{"type":"gauge","value":1}`),
			},
			wantErr: true,
		},
		{
			name: "test 3, invalid value type",
			args: args{
				data: []byte(`{"metrics":{"test":{"type":"counter","value":"invalid"}}}`),
			},
			wantErr: true,
		},
	}

	st := &MemStorage{
		Storage: make(map[string]models.Metric, 1),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := st.UnmarshalJSON(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemStorage_Update(t *testing.T) {
	type args struct {
		v any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test 1, valid data",
			args: args{
				v: runtime.MemStats{},
			},
			wantErr: false,
		},
		{
			name: "test 2, invalid data",
			args: args{
				v: "invalid",
			},
			wantErr: true,
		},
	}

	st := &MemStorage{
		Storage: make(map[string]models.Metric, 1),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := st.Update(context.Background(), tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
