package migration

import (
	"testing"

	"dbmigrator/internal/config"
	"dbmigrator/internal/dialect"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestValidateTable_Pass(t *testing.T) {
	srcDB, srcMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer srcDB.Close()

	tgtDB, tgtMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer tgtDB.Close()

	srcMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"USERS\"").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))
	tgtMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))

	cfg := &config.Config{Schema: "public"}
	dia := &dialect.PostgresDialect{}
	result := validateTable(srcDB, tgtDB, nil, dia, "USERS", cfg)

	if result.Status != "pass" {
		t.Errorf("expected pass, got %s (detail: %s)", result.Status, result.Detail)
	}
	if result.SourceCount != 100 || result.TargetCount != 100 {
		t.Errorf("counts: source=%d target=%d", result.SourceCount, result.TargetCount)
	}
}

func TestValidateTable_Mismatch(t *testing.T) {
	srcDB, srcMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer srcDB.Close()

	tgtDB, tgtMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer tgtDB.Close()

	srcMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"ORDERS\"").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(200))
	tgtMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(195))

	cfg := &config.Config{}
	dia := &dialect.PostgresDialect{}
	result := validateTable(srcDB, tgtDB, nil, dia, "ORDERS", cfg)

	if result.Status != "mismatch" {
		t.Errorf("expected mismatch, got %s", result.Status)
	}
	if result.Detail == "" {
		t.Error("expected non-empty detail for mismatch")
	}
}

func TestValidateTable_SourceError(t *testing.T) {
	srcDB, srcMock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer srcDB.Close()

	srcMock.ExpectQuery("SELECT COUNT").
		WillReturnError(sqlmock.ErrCancelled)

	cfg := &config.Config{}
	dia := &dialect.PostgresDialect{}
	result := validateTable(srcDB, nil, nil, dia, "BAD_TABLE", cfg)

	if result.Status != "error" {
		t.Errorf("expected error, got %s", result.Status)
	}
}
