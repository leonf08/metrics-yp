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
	prep.ExpectExec().
		WithArgs("name", "counter", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
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

func TestPGStorage_UpdateExecErr(t *testing.T) {
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

func TestPGStorage_UpdateBeginErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin().WillReturnError(assert.AnError)

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

func TestPGStorage_UpdatePrepareErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	mock.ExpectPrepare("INSERT INTO metrics").WillReturnError(assert.AnError)
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

func TestPGStorage_UpdateCommitErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	prep := mock.ExpectPrepare("INSERT INTO metrics")
	prep.ExpectExec().
		WithArgs("name", "counter", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit().WillReturnError(assert.AnError)

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

func TestPGStorage_Ping(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectPing()

	st := &PGStorage{db: sqlxDB}

	err = st.Ping()
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPGStorage_PingErr(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectPing().WillReturnError(assert.AnError)

	st := &PGStorage{db: sqlxDB}

	err = st.Ping()
	assert.Error(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPGStorage_SetVal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectExec("INSERT INTO metrics").
		WithArgs("name", "counter", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	st := &PGStorage{db: sqlxDB}

	err = st.SetVal(context.Background(), "name", models.Metric{Type: "counter", Val: 1})
	assert.NoError(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPGStorage_SetValExecErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectExec("INSERT INTO metrics").
		WithArgs("name", "counter", 1).
		WillReturnError(assert.AnError)

	st := &PGStorage{db: sqlxDB}

	err = st.SetVal(context.Background(), "name", models.Metric{Type: "counter", Val: 1})
	assert.Error(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPGStorage_GetVal(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	rows := sqlmock.NewRows([]string{"type", "value"}).AddRow("counter", 1)
	mock.ExpectQuery("SELECT TYPE, VALUE FROM metrics").
		WithArgs("name").
		WillReturnRows(rows)

	st := &PGStorage{db: sqlxDB}

	val, err := st.GetVal(context.Background(), "name")
	assert.NoError(t, err)
	assert.Equal(t, models.Metric{Type: "counter", Val: int64(1)}, val)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPGStorage_GetValQueryErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectQuery("SELECT TYPE, VALUE FROM metrics").
		WithArgs("name").
		WillReturnError(assert.AnError)

	st := &PGStorage{db: sqlxDB}

	_, err = st.GetVal(context.Background(), "name")
	assert.Error(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPGStorage_ReadAll(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	rows := sqlmock.NewRows([]string{"name", "type", "value"}).
		AddRow("name1", "counter", float64(1)).
		AddRow("name2", "gauge", 2.5)
	mock.ExpectQuery("SELECT \\* FROM metrics").
		WillReturnRows(rows)

	st := &PGStorage{db: sqlxDB}

	m, err := st.ReadAll(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, map[string]models.Metric{
		"name1": {Type: "counter", Val: int64(1)},
		"name2": {Type: "gauge", Val: 2.5},
	}, m)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPGStorage_ReadAllQueryErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectQuery("SELECT \\* FROM metrics").
		WillReturnError(assert.AnError)

	st := &PGStorage{db: sqlxDB}

	_, err = st.ReadAll(context.Background())
	assert.Error(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func TestPGStorage_ReadAllInvalidCounterValueErr(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	rows := sqlmock.NewRows([]string{"name", "type", "value"}).
		AddRow("name1", "counter", 1)
	mock.ExpectQuery("SELECT \\* FROM metrics").
		WillReturnRows(rows)

	st := &PGStorage{db: sqlxDB}

	_, err = st.ReadAll(context.Background())
	assert.Error(t, err)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
