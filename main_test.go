package main

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"testing"

	"dbmigrator/internal/config"
	"dbmigrator/internal/dialect"
	"dbmigrator/internal/migration"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMigrateTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql connection: %v", err)
	}
	defer db.Close()

	tableName := "MOCK_TABLE"

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM MOCK_TABLE").WillReturnRows(sqlmock.NewRows([]string{"COUNT"}).AddRow(2))

	rows := sqlmock.NewRows([]string{"ID", "NAME"}).
		AddRow(1, "Alice").
		AddRow(2, "Bob")

	mock.ExpectQuery("SELECT \\* FROM " + tableName).WillReturnRows(rows)

	tmpFile, err := os.CreateTemp("", "migrate_test_*.sql")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	mainBuf := bufio.NewWriter(tmpFile)
	var outMutex sync.Mutex
	cfg := &config.Config{BatchSize: 1000}
	dia := &dialect.PostgresDialect{}

	err = migration.MigrateTable(db, nil, nil, dia, "MOCK_TABLE", mainBuf, cfg, &outMutex, nil, migration.NewMigrationState("test"))
	if err != nil {
		t.Errorf("MigrateTable returned error: %v", err)
	}
	mainBuf.Flush()

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	if !strings.Contains(strings.ToLower(string(content)), "insert into \"mock_table\"") {
		t.Errorf("Output missing expected INSERT statement. Got:\n%s", string(content))
	}
	if !strings.Contains(string(content), "1") || !strings.Contains(string(content), "Bob") {
		t.Errorf("Output missing expected row data. Got:\n%s", string(content))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
