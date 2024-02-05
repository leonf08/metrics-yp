package services

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashSigner_CalcHash(t *testing.T) {
	type args struct {
		src []byte
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test 1, valid data",
			args: args{
				src: []byte("test"),
			},
			want:    "88cd2108b5347d973cf39cdf9053d7dd42704876d8c9a9bd8e2d168259d3ddf7",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HashSigner{
				key: "test",
			}
			got, err := h.CalcHash(tt.args.src)

			assert.Equal(t, tt.wantErr, err != nil, "CalcHash(%v)", tt.args.src)
			assert.Equalf(t, tt.want, hex.EncodeToString(got), "CalcHash(%v)", tt.args.src)
		})
	}
}

func TestHashSigner_CheckHash(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test 1, valid hashes",
			args: args{
				s: "88cd2108b5347d973cf39cdf9053d7dd42704876d8c9a9bd8e2d168259d3ddf7",
			},
			want: true,
		},
		{
			name: "test 2, invalid hashes",
			args: args{
				s: "88cd2108b5347d973cf39cdf9053d7dd42704876d8c9a9bd8e2d168259d3ddf8",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &HashSigner{
				key: "test",
			}

			h1, _ := h.CalcHash([]byte("test"))
			h2, _ := hex.DecodeString(tt.args.s)
			assert.Equal(t, tt.want, h.CheckHash(h1, h2), "CheckHash()")
		})
	}
}

func TestNewHashSigner(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want *HashSigner
	}{
		{
			name: "test 1, key is valid",
			args: args{
				key: "test",
			},
			want: &HashSigner{
				key: "test",
			},
		},
		{
			name: "test 2, key is empty",
			args: args{
				key: "",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, NewHashSigner(tt.args.key), "NewHashSigner(%v)", tt.args.key)
		})
	}
}
