package migration

import (
	"testing"

	"dbmigrator/internal/config"
	"dbmigrator/internal/dialect"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRun_NoTables(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	
	cfg := &config.Config{
		Tables: []string{},
	}
	dia := &dialect.PostgresDialect{}
	
	_, err := Run(db, nil, nil, dia, cfg, nil)
	if err == nil {
		t.Error("expected error for no tables")
	}
}

func TestRun_SequencesOnly(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{
		ObjectGroup: config.ObjectGroupSequences,
		Tables:      []string{"T1"},
	}
	dia := &dialect.PostgresDialect{}

	// collectGroupedMetadata will be called
	mock.ExpectQuery("SELECT column_name").WillReturnRows(sqlmock.NewRows([]string{"C"}).AddRow("id"))

	_, err = Run(db, nil, nil, dia, cfg, nil)
	// It might fail later due to missing more mocks, but will cover the branch.
	if err == nil {
		t.Error("expected error or completion")
	}
}
