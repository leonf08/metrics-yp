package services

import (
	"context"

	"github.com/leonf08/metrics-yp.git/internal/services/repo"
)

//go:generate mockery --name FileStore --output ./mocks --filename filestore_mock.go
//go:generate mockery --name Crypto --output ./mocks --filename crypto_mock.go
//go:generate mockery --name Pinger --output ./mocks --filename pinger_mock.go
type (
	// Agent is an interface for gathering and reporting metrics.
	Agent interface {
		GatherMetrics(context.Context) error
		ReportMetrics(context.Context) ([]string, error)
	}

	// FileStore is an interface for file storage.
	FileStore interface {
		Save(repo.Repository) error
		Load(repo.Repository) error
		Close()
	}

	// Crypto is an interface for encryption and decryption.
	Crypto interface {
		Decrypt(src []byte) ([]byte, error)
		Encrypt(src []byte) ([]byte, error)
	}

	// Pinger is an interface for checking connection to the database.
	Pinger interface {
		Ping() error
	}
)
