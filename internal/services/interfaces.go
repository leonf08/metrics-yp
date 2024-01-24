package services

import (
	"context"

	"github.com/leonf08/metrics-yp.git/internal/models"
)

//go:generate mockery --name Repository --output ./mocks --filename repo_mock.go
//go:generate mockery --name FileStore --output ./mocks --filename filestore_mock.go
type (
	Agent interface {
		GatherMetrics(context.Context) error
		ReportMetrics(context.Context) ([]string, error)
	}

	Repository interface {
		ReadAll(context.Context) (map[string]models.Metric, error)
		Update(context.Context, any) error
		SetVal(context.Context, string, models.Metric) error
		GetVal(context.Context, string) (models.Metric, error)
	}

	FileStore interface {
		Save(Repository) error
		Load(Repository) error
	}

	Pinger interface {
		Ping() error
	}
)
