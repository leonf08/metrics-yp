package agentconf

import (
	"reflect"
	"testing"
)

func TestMustLoadConfig(t *testing.T) {
	tests := []struct {
		name string
		want Config
	}{
		{
			name: "Test MustLoadConfig",
			want: Config{
				Addr:      defaultAddress,
				Mode:      defaultMode,
				SignKey:   "",
				CryptoKey: "",
				ReportInt: defaultReportInt,
				PollInt:   defaultPollInt,
				RateLim:   int(defaultRateLimit),
				GRPCAddr:  defaultGRPCAddr,
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

func Test_modeEnum_Set(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		m       modeEnum
		args    args
		wantErr bool
	}{
		{
			name:    "Test modeEnum Set, valid",
			m:       modeEnum("json"),
			args:    args{s: "json"},
			wantErr: false,
		},
		{
			name:    "Test modeEnum Set, invalid",
			m:       modeEnum("json"),
			args:    args{s: "invalid"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.m.Set(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_modeEnum_String(t *testing.T) {
	tests := []struct {
		name string
		m    modeEnum
		want string
	}{
		{
			name: "Test modeEnum String",
			m:    modeEnum("json"),
			want: "json",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_modeEnum_Type(t *testing.T) {
	tests := []struct {
		name string
		m    modeEnum
		want string
	}{
		{
			name: "Test modeEnum Type",
			m:    modeEnum("json"),
			want: "modeEnum",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.m.Type(); got != tt.want {
				t.Errorf("Type() = %v, want %v", got, tt.want)
			}
		})
	}
}
