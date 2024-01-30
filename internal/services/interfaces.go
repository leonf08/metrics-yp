package services

import (
	"context"

	"github.com/leonf08/metrics-yp.git/internal/models"
)

//go:generate mockery --name Repository --output ./mocks --filename repo_mock.go
//go:generate mockery --name FileStore --output ./mocks --filename filestore_mock.go
type (
	// Agent is an interface for gathering and reporting metrics.
	Agent interface {
		GatherMetrics(context.Context) error
		ReportMetrics(context.Context) ([]string, error)
	}

	// Repository is an interface for metrics storage.
	Repository interface {
		ReadAll(context.Context) (map[string]models.Metric, error)
		Update(context.Context, any) error
		SetVal(context.Context, string, models.Metric) error
		GetVal(context.Context, string) (models.Metric, error)
	}

	// FileStore is an interface for file storage.
	FileStore interface {
		Save(Repository) error
		Load(Repository) error
	}

	// Pinger is an interface for checking connection to the database.
	Pinger interface {
		Ping() error
	}
)
