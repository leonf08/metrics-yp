package storage

import (
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type PostGresDB struct {
	db *sqlx.DB
}

func NewDb(sourceName string) (*PostGresDB, error) {
	if sourceName == "" {
		return nil, nil
	}

	db, err := sqlx.Open("pgx", sourceName)
	if err != nil {
		return nil, err
	}

	return &PostGresDB{
		db: db,
	}, nil
}

func (db *PostGresDB) CheckConn() error {
	return db.db.Ping()
}
