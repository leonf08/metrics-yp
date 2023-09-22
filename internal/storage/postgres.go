package storage

import (
	"context"
	"errors"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type PostgresDB struct {
	db *sqlx.DB
}

func NewDB(sourceName string) (*PostgresDB, error) {
	if sourceName == "" {
		return nil, nil
	}

	db, err := sqlx.Open("pgx", sourceName)
	if err != nil {
		return nil, err
	}

	return &PostgresDB{
		db: db,
	}, nil
}

func (db *PostgresDB) Ping() error {
	return db.db.Ping()
}

func (db *PostgresDB) CreateTable(ctx context.Context) error {
	queryStr := `
		CREATE TABLE IF NOT EXISTS metrics(
			name TEXT PRIMARY KEY,
			type TEXT,
			value DOUBLE PRECISION
		)
	`
	if _, err := db.db.ExecContext(ctx, queryStr); err != nil {
		return err
	}

	return nil
}

func (db *PostgresDB) Update(ctx context.Context, v any) error {
	return nil
}

func (db *PostgresDB) ReadAll(ctx context.Context) (map[string]any, error) {
	queryStr := `SELECT * FROM metrics`

	rows, err := db.db.QueryxContext(ctx, queryStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	metrics := make(map[string]any)
	for rows.Next() {
		var m dbEntry
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

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return metrics, nil
}

func (db *PostgresDB) SetVal(ctx context.Context, k string, v any) error {
	queryCheck := `SELECT EXISTS(SELECT 1 FROM metrics WHERE NAME = $1)`

	row := db.db.QueryRowxContext(ctx, queryCheck, k)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return err
	}

	var queryStr string
	if exists {
		queryStr = `UPDATE metrics SET VALUE = $1 WHERE NAME = $2`
		_, err := db.db.ExecContext(ctx, queryStr, k, v)
		if err != nil {
			return err
		}
	} else {
		var t string
		_, ok := v.(float64)
		if ok {
			t = "gauge"
		} else {
			t = "counter"
		}

		queryStr = `INSERT INTO metrics (NAME,TYPE,VALUE) VALUES ($1, $2, $3)`
		_, err := db.db.ExecContext(ctx, queryStr, k, t, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *PostgresDB) GetVal(ctx context.Context, k string) (any, error) {
	queryStr := `SELECT TYPE, VALUE FROM metrics WHERE NAME = $1`

	row := db.db.QueryRowxContext(ctx, queryStr, k)

	var m Metric
	if err := row.StructScan(&m); err != nil {
		return nil, err
	}

	return m, nil
}
