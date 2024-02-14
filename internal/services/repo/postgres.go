package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/leonf08/metrics-yp.git/internal/database/migrations/postgres"
	"github.com/leonf08/metrics-yp.git/internal/errorhandling"
	"github.com/leonf08/metrics-yp.git/internal/models"
)

// PGStorage is database implementation of metrics storage.
type PGStorage struct {
	db *sqlx.DB
}

// NewDB creates a new database connection.
func NewDB(sourceName string) (*PGStorage, error) {
	if sourceName == "" {
		return nil, nil
	}

	db, err := postgres.NewConnection(sourceName)
	if err != nil {
		return nil, err
	}

	return &PGStorage{
		db: db,
	}, nil
}

// Ping checks connection to the database.
func (st *PGStorage) Ping() error {
	return st.db.Ping()
}

// Update updates metrics in the storage.
func (st *PGStorage) Update(ctx context.Context, v any) error {
	metrics, ok := v.([]models.MetricDB)
	if !ok {
		return errors.New("invalid type assertion")
	}

	const queryStr = `
		INSERT INTO metrics (NAME, TYPE, VALUE)
		VALUES ($1, $2, $3)
		ON CONFLICT (NAME) 
		DO UPDATE SET
		VALUE = CASE
			WHEN $2 = 'counter' THEN metrics.VALUE + $3
			ELSE $3
		END
		WHERE metrics.NAME = $1`

	fn := func() error {
		tx, err := st.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}

		//lint:ignore errcheck
		defer tx.Rollback()

		stmt, err := tx.PreparexContext(ctx, queryStr)
		if err != nil {
			return err
		}

		defer stmt.Close()

		for _, m := range metrics {
			_, err := stmt.ExecContext(ctx, m.Name, m.Type, m.Val)
			if err != nil {
				return err
			}
		}

		return tx.Commit()
	}

	err := errorhandling.Retry(ctx, func() error {
		err := fn()
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) &&
				(pgerrcode.IsInsufficientResources(pgErr.Code) ||
					pgerrcode.IsConnectionException(pgErr.Code)) {
				err = errorhandling.ErrRetriable
			}
		}

		return err
	})

	return err
}

// ReadAll returns all metrics.
func (st *PGStorage) ReadAll(ctx context.Context) (map[string]models.Metric, error) {
	const queryStr = `SELECT * FROM metrics`

	var rows *sqlx.Rows
	err := errorhandling.Retry(ctx, func() error {
		var err error
		r, err := st.db.QueryxContext(ctx, queryStr)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) &&
				(pgerrcode.IsInsufficientResources(pgErr.Code) ||
					pgerrcode.IsConnectionException(pgErr.Code)) {
				err = errorhandling.ErrRetriable
				return err
			}

			return err
		}

		rows = r

		return r.Err()
	})

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	metrics := make(map[string]models.Metric, 30)
	for rows.Next() {
		var m models.MetricDB
		if err = rows.StructScan(&m); err != nil {
			return nil, err
		}

		if m.Type == "counter" {
			v, ok := m.Val.(float64)
			if !ok {
				return nil, errors.New("invalid type assertion")
			}

			m.Val = int64(v)
		}
		metrics[m.Name] = m.Metric
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return metrics, nil
}

// SetVal sets a value for a metric.
func (st *PGStorage) SetVal(ctx context.Context, k string, m models.Metric) error {
	const queryStr = `
		INSERT INTO metrics (NAME, TYPE, VALUE)
		VALUES ($1, $2, $3)
		ON CONFLICT (NAME) 
		DO UPDATE SET
		VALUE = CASE
			WHEN $2 = 'counter' THEN metrics.VALUE + $3
			ELSE $3
		END
		WHERE metrics.NAME = $1`

	err := errorhandling.Retry(ctx, func() error {
		_, err := st.db.ExecContext(ctx, queryStr, k, m.Type, m.Val)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) &&
				(pgerrcode.IsInsufficientResources(pgErr.Code) ||
					pgerrcode.IsConnectionException(pgErr.Code)) {
				err = errorhandling.ErrRetriable
			}
		}

		return err
	})

	return err
}

// GetVal returns a value for a metric.
func (st *PGStorage) GetVal(ctx context.Context, k string) (models.Metric, error) {
	queryStr := `SELECT TYPE, VALUE FROM metrics WHERE NAME = $1`

	var m models.Metric

	err := errorhandling.Retry(ctx, func() error {
		err := st.db.GetContext(ctx, &m, queryStr, k)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) &&
				(pgerrcode.IsInsufficientResources(pgErr.Code) ||
					pgerrcode.IsConnectionException(pgErr.Code)) {
				err = errorhandling.ErrRetriable
			}
		}

		return err
	})

	return m, err
}
