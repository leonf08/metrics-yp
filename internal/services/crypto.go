package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
)

type CryptoService struct {
	cryptoKeyFile string
}

func NewCryptoService(cryptoKeyFile string) *CryptoService {
	return &CryptoService{
		cryptoKeyFile: cryptoKeyFile,
	}
}

func (c CryptoService) Decrypt(src []byte) ([]byte, error) {
	b, err := os.ReadFile(c.cryptoKeyFile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(b)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	message, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, src, nil)
	if err != nil {
		return nil, err
	}

	return message, nil
}

func (c CryptoService) Encrypt(src []byte) ([]byte, error) {
	b, err := os.ReadFile(c.cryptoKeyFile)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(b)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	message, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, src, nil)
	if err != nil {
		return nil, err
	}

	return message, nil
}
