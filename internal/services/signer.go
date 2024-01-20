package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

var _ http.ResponseWriter = (*HashSigner)(nil)

type HashSigner struct {
	http.ResponseWriter
	key string
}

func NewHashSigner(key string) *HashSigner {
	if key == "" {
		return nil
	}

	return &HashSigner{
		key: key,
	}
}

func (h *HashSigner) Write(b []byte) (int, error) {
	hash, err := h.CalcHash(b)
	if err != nil {
		return 0, err
	}

	h.Header().Set("HashSHA256", hex.EncodeToString(hash))
	return h.ResponseWriter.Write(b)
}

func (h *HashSigner) CalcHash(src []byte) ([]byte, error) {
	hash := hmac.New(sha256.New, []byte(h.key))
	_, err := hash.Write(src)
	if err != nil {
		return nil, err
	}

	return hash.Sum(nil), nil
}

func (h *HashSigner) CheckHash(h1, h2 []byte) bool {
	return hmac.Equal(h1, h2)
}
