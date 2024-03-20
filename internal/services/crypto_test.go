package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCryptoService_Decrypt(t *testing.T) {
	type args struct {
		src      []byte
		filename string
		t        string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Decrypt, success",
			args: args{
				src:      []byte("test"),
				filename: "private.pem",
				t:        "RSA PRIVATE KEY",
			},
			wantErr: false,
		},
		{
			name: "Test Decrypt, invalid type",
			args: args{
				src:      []byte("test"),
				filename: "private.pem",
				t:        "PRIVATE KEY",
			},
			wantErr: true,
		},
		{
			name: "Test Decrypt, key file not found",
			args: args{
				src:      []byte("test"),
				filename: "test.pem",
				t:        "RSA PUBLIC KEY",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CryptoService{
				cryptoKeyFile: tt.args.filename,
			}

			err := generateKeyPair("private.pem", "public.pem", tt.args.t, "RSA PUBLIC KEY")
			require.NoError(t, err)

			src, err := CryptoService{"public.pem"}.Encrypt(tt.args.src)
			require.NoError(t, err)

			_, err = c.Decrypt(src)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			err = os.Remove("public.pem")
			require.NoError(t, err)

			err = os.Remove("private.pem")
			require.NoError(t, err)
		})
	}
}

func TestCryptoService_Encrypt(t *testing.T) {
	type args struct {
		src      []byte
		filename string
		t        string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Test Encrypt, success",
			args: args{
				src:      []byte("test"),
				filename: "public.pem",
				t:        "RSA PUBLIC KEY",
			},
			wantErr: false,
		},
		{
			name: "Test Encrypt, invalid type",
			args: args{
				src:      []byte("test"),
				filename: "public.pem",
				t:        "PUBLIC KEY",
			},
			wantErr: true,
		},
		{
			name: "Test Encrypt, key file not found",
			args: args{
				src:      []byte("test"),
				filename: "test.pem",
				t:        "RSA PUBLIC KEY",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CryptoService{
				cryptoKeyFile: tt.args.filename,
			}

			err := generateKeyPair("private.pem", "public.pem", "RSA PRIVATE KEY", tt.args.t)
			require.NoError(t, err)

			s, err := c.Encrypt(tt.args.src)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Emptyf(t, s, "expected empty string, got %s", s)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, s)
			}

			err = os.Remove("public.pem")
			require.NoError(t, err)

			err = os.Remove("private.pem")
			require.NoError(t, err)
		})
	}
}

func generateKeyPair(prFile, pubFile, prType, pubType string) error {
	prF, err := os.Create(prFile)
	if err != nil {
		return err
	}
	defer prF.Close()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  prType,
		Bytes: privateKeyBytes,
	}

	err = pem.Encode(prF, block)
	if err != nil {
		return err
	}

	pubF, err := os.Create(pubFile)
	if err != nil {
		return err
	}

	publicKeyBytes := x509.MarshalPKCS1PublicKey(&privateKey.PublicKey)
	block = &pem.Block{
		Type:  pubType,
		Bytes: publicKeyBytes,
	}

	return pem.Encode(pubF, block)
}

func TestNewCryptoService(t *testing.T) {
	type args struct {
		cryptoKeyFile string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test NewCryptoService",
			args: args{
				cryptoKeyFile: "test.pem",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCryptoService(tt.args.cryptoKeyFile)

			assert.Equal(t, tt.args.cryptoKeyFile, c.cryptoKeyFile)
		})
	}
}
