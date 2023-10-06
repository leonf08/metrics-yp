package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

type hashWriter struct {
	http.ResponseWriter
	key []byte
}

func NewAuthentificator(w http.ResponseWriter, key []byte) *hashWriter {
	return &hashWriter{
		ResponseWriter: w,
		key:            key,
	}
}

func (h *hashWriter) Write(b []byte) (int, error) {
	hash, err := CalcHash(b, h.key)
	if err != nil {
		return 0, err
	}

	h.Header().Set("HashSHA256", hex.EncodeToString(hash))
	return h.ResponseWriter.Write(b)
}

func CalcHash(src, key []byte) ([]byte, error) {
	h := hmac.New(sha256.New, key)
	_, err := h.Write(src)
	if err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

func CheckHash(h1, h2 []byte) bool {
	return hmac.Equal(h1, h2)
}
