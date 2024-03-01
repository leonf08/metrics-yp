package repo

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/leonf08/metrics-yp.git/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestPGStorage_Update(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	prep := mock.ExpectPrepare("INSERT INTO metrics")
	prep.ExpectExec().WithArgs("name", "counter", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	st := &PGStorage{db: sqlxDB}

	m := []models.MetricDB{
		{
			Name: "name",
			Metric: models.Metric{
				Type: "counter",
				Val:  1,
			},
		},
	}

	err = st.Update(context.Background(), m)
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPGStorage_UpdateErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	prep := mock.ExpectPrepare("INSERT INTO metrics")
	prep.ExpectExec().WithArgs("name", "counter", 1).WillReturnError(assert.AnError)
	mock.ExpectRollback()

	st := &PGStorage{db: sqlxDB}

	m := []models.MetricDB{
		{
			Name: "name",
			Metric: models.Metric{
				Type: "counter",
				Val:  1,
			},
		},
	}

	err = st.Update(context.Background(), m)
	assert.Error(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
