package repo

import (
	"context"

	"github.com/leonf08/metrics-yp.git/internal/models"
)

// Repository is an interface for metrics storage.
//
//go:generate mockery --name Repository --output ../mocks --filename repo_mock.go
type Repository interface {
	ReadAll(context.Context) (map[string]models.Metric, error)
	Update(context.Context, any) error
	SetVal(context.Context, string, models.Metric) error
	GetVal(context.Context, string) (models.Metric, error)
}
