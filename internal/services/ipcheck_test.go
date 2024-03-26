package services

import (
	"fmt"
	"net/netip"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIPCheckService_IsTrusted(t *testing.T) {
	type fields struct {
		trustedSubnet string
	}
	type args struct {
		ip string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "IsTrusted, true",
			fields: fields{
				trustedSubnet: "192.168.1.0/24",
			},
			args: args{
				ip: "192.168.1.4",
			},
			want:    true,
			wantErr: assert.NoError,
		},
		{
			name: "IsTrusted, false",
			fields: fields{
				trustedSubnet: "192.168.1.0/24",
			},
			args: args{
				ip: "192.168.0.0",
			},
			want:    false,
			wantErr: assert.NoError,
		},
		{
			name: "IsTrusted, error",
			fields: fields{
				trustedSubnet: "192.168.1.0/24",
			},
			args: args{
				ip: "",
			},
			want:    false,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, err := netip.ParsePrefix(tt.fields.trustedSubnet)
			require.NoError(t, err, "ParsePrefix")

			ip := NewIPChecker(prefix)

			got, err := ip.IsTrusted(tt.args.ip)
			if !tt.wantErr(t, err, fmt.Sprintf("IsTrusted(%v)", tt.args.ip)) {
				return
			}
			assert.Equalf(t, tt.want, got, "IsTrusted(%v)", tt.args.ip)
		})
	}
}

func TestNewIPChecker(t *testing.T) {
	type args struct {
		trustedSubnet string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "NewIPChecker",
			args: args{
				trustedSubnet: "192.168.0.1/24",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, err := netip.ParsePrefix(tt.args.trustedSubnet)
			require.NoError(t, err, "ParsePrefix")

			ip := NewIPChecker(prefix)
			if !reflect.DeepEqual(ip, &IPCheckService{prefix: prefix}) {
				t.Errorf("NewIPChecker() = %v, want %v", ip, &IPCheckService{prefix: prefix})
			}
		})
	}
}
