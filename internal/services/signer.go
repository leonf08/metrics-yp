package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

var _ http.ResponseWriter = (*HashSigner)(nil)

// HashSigner service is used to hash calculation and verification.
type HashSigner struct {
	http.ResponseWriter
	key string
}

// NewHashSigner creates instance of the signer. It returns nil if key is empty.
func NewHashSigner(key string) *HashSigner {
	if key == "" {
		return nil
	}

	return &HashSigner{
		key: key,
	}
}

// Write overrides http.ResponseWriter.Write method.
// It calculates hash of the data and writes it to the header.
func (h *HashSigner) Write(b []byte) (int, error) {
	hash, err := h.CalcHash(b)
	if err != nil {
		return 0, err
	}

	h.Header().Set("HashSHA256", hex.EncodeToString(hash))
	return h.ResponseWriter.Write(b)
}

// CalcHash calculates SHA256 hash of the data.
func (h *HashSigner) CalcHash(src []byte) ([]byte, error) {
	hash := hmac.New(sha256.New, []byte(h.key))
	_, err := hash.Write(src)
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

// CheckHash checks if the hash is equal to the hash of the data.
// It returns true if hashes are equal and false otherwise.
func (h *HashSigner) CheckHash(h1, h2 []byte) bool {
	return hmac.Equal(h1, h2)
}
