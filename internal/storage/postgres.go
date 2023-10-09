package storage

import (
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/leonf08/metrics-yp.git/internal/database/migrations/postgres"
	"github.com/leonf08/metrics-yp.git/internal/errorhandling"
)

type PGStorage struct {
	db *sqlx.DB
}

func NewDB(sourceName string) (*PGStorage, error) {
	if sourceName == "" {
		return nil, nil
	}

	db, err := sqlx.Open("pgx", sourceName)
	if err != nil {
		return nil, err
	}

	m, err := postgres.Migrate(db.DB)
	if err != nil {
		return nil, err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	return &PGStorage{
		db: db,
	}, nil
}

func (st *PGStorage) Ping() error {
	return st.db.Ping()
}

func (st *PGStorage) Update(ctx context.Context, v any) error {
	metrics, ok := v.([]MetricDB)
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
		WHERE metrics.NAME = $1;`

	fn := func() error {
		tx, err := st.db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}

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

func (st *PGStorage) ReadAll(ctx context.Context) (map[string]any, error) {
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

	metrics := make(map[string]any, 30)
	for rows.Next() {
		var m MetricDB
		if err := rows.StructScan(&m); err != nil {
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

func (st *PGStorage) SetVal(ctx context.Context, k string, v any) error {
	const queryStr = `
		INSERT INTO metrics (NAME, TYPE, VALUE)
		VALUES ($1, $2, $3)
		ON CONFLICT (NAME) 
		DO UPDATE SET
		VALUE = CASE
			WHEN $2 = 'counter' THEN metrics.VALUE + $3
			ELSE $3
		END
		WHERE metrics.NAME = $1;`

	var t string
	_, ok := v.(float64)
	if ok {
		t = "gauge"
	} else {
		t = "counter"
	}

	err := errorhandling.Retry(ctx, func() error {
		_, err := st.db.ExecContext(ctx, queryStr, k, t, v)
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

func (st *PGStorage) GetVal(ctx context.Context, k string) (any, error) {
	queryStr := `SELECT TYPE, VALUE FROM metrics WHERE NAME = $1`

	var m Metric

	err := errorhandling.Retry(ctx, func() error {
		row := st.db.QueryRowxContext(ctx, queryStr, k)
		err := row.StructScan(&m)
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

func (st *PGStorage) Close() {
	st.db.Close()
}
