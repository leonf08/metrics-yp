package serverconf

import (
	"reflect"
	"testing"
)

func TestConfig_IsFileStorage(t *testing.T) {
	type args struct {
		fileStore string
		dsn       string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test IsFileStorage 1, true",
			args: args{
				fileStore: "test",
				dsn:       "",
			},
			want: true,
		},
		{
			name: "Test IsFileStorage 2, false",
			args: args{
				fileStore: "",
				dsn:       "test",
			},
			want: false,
		},
		{
			name: "Test IsFileStorage 3, false",
			args: args{
				fileStore: "test",
				dsn:       "test",
			},
			want: false,
		},
		{
			name: "Test IsFileStorage 4, false",
			args: args{
				fileStore: "",
				dsn:       "",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				FileStoragePath: tt.args.fileStore,
				DatabaseAddr:    tt.args.dsn,
			}
			if got := cfg.IsFileStorage(); got != tt.want {
				t.Errorf("IsFileStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_IsInMemStorage(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want bool
	}{
		{
			name: "Test IsInMemStorage, true",
			arg:  "",
			want: true,
		},
		{
			name: "Test IsInMemStorage, false",
			arg:  "postgresql://localhost:5432/postgres",
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				DatabaseAddr: tt.arg,
			}
			if got := cfg.IsInMemStorage(); got != tt.want {
				t.Errorf("IsInMemStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustLoadConfig(t *testing.T) {
	tests := []struct {
		name string
		want Config
	}{
		{
			name: "Test MustLoadConfig",
			want: Config{
				Addr:            defaultAddress,
				StoreInt:        defaultStoreInterval,
				FileStoragePath: "",
				Restore:         defaultRestore,
				DatabaseAddr:    "",
				SignKey:         "",
				CryptoKey:       "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MustLoadConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MustLoadConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
