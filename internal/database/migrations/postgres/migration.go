package postgres

import (
	"database/sql"
	"embed"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed schema/*.sql
var fs embed.FS

func Migrate(db *sql.DB) (*migrate.Migrate, error) {
	d, err := iofs.New(fs, "schema")
	if err != nil {
		return nil, err
	}

	inst, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, err
	}

	m, err := migrate.NewWithInstance("iofs", d, "metrics", inst)
	if err != nil {
		return nil, err
	}

	return m, nil
}
