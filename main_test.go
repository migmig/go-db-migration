package main

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"testing"

	"dbmigrator/internal/config"
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

	err = migration.MigrateTable(db, nil, tableName, mainBuf, cfg, &outMutex)
	if err != nil {
		t.Errorf("MigrateTable returned error: %v", err)
	}
	mainBuf.Flush()

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	if !strings.Contains(string(content), "INSERT INTO MOCK_TABLE (ID, NAME) VALUES") {
		t.Errorf("Output missing expected INSERT statement. Got:\n%s", string(content))
	}
	if !strings.Contains(string(content), "(1, 'Alice')") || !strings.Contains(string(content), "(2, 'Bob')") {
		t.Errorf("Output missing expected row data. Got:\n%s", string(content))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestWriteBatch_SchemaMapping(t *testing.T) {
	tableName := "MY_TABLE"
	schemaPrefix := "public."
	cols := []string{"ID", "NAME"}
	batch := []string{"(1, 'John')", "(2, 'Jane')"}
	var outMutex sync.Mutex
	cfg := &config.Config{Schema: "public", PerTable: false}

	tmpFile, err := os.CreateTemp("", "test_migration_*.sql")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	buf := bufio.NewWriter(tmpFile)
	migration.WriteBatch(buf, tableName, cols, batch, cfg, &outMutex)
	buf.Flush()

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	expectedPrefix := "INSERT INTO " + schemaPrefix + tableName
	if !strings.Contains(string(content), expectedPrefix) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedPrefix, string(content))
	}
}

func TestWriteBatch_NoSchemaMapping(t *testing.T) {
	tableName := "MY_TABLE"
	cols := []string{"ID", "NAME"}
	batch := []string{"(1, 'John')"}
	var outMutex sync.Mutex
	cfg := &config.Config{Schema: "", PerTable: false}

	tmpFile, err := os.CreateTemp("", "test_migration_noschema_*.sql")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	buf := bufio.NewWriter(tmpFile)
	migration.WriteBatch(buf, tableName, cols, batch, cfg, &outMutex)
	buf.Flush()

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	expectedPrefix := "INSERT INTO " + tableName
	if !strings.Contains(string(content), expectedPrefix) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedPrefix, string(content))
	}

	if strings.Contains(string(content), "INSERT INTO .") {
		t.Errorf("Output should not contain '.' before table name if schema is empty")
	}
}

func TestWriteBatch_PerTable(t *testing.T) {
	tableName := "MY_TABLE"
	cols := []string{"ID"}
	batch := []string{"(1)"}
	var outMutex sync.Mutex
	cfg := &config.Config{Schema: "", PerTable: true}

	tmpFile, err := os.CreateTemp("", "test_migration_pertable_*.sql")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	buf := bufio.NewWriter(tmpFile)
	migration.WriteBatch(buf, tableName, cols, batch, cfg, &outMutex)
	buf.Flush()

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	if !strings.Contains(string(content), "INSERT INTO MY_TABLE") {
		t.Errorf("Expected output to contain 'INSERT INTO MY_TABLE'")
	}
}

func TestProcessRow(t *testing.T) {
	tests := []struct {
		name      string
		values    []interface{}
		typeNames []string
		expected  string
	}{
		{
			name:      "Simple types",
			values:    []interface{}{1, "John"},
			typeNames: []string{"NUMBER", "VARCHAR2"},
			expected:  "1, 'John'",
		},
		{
			name:      "NULL value",
			values:    []interface{}{nil},
			typeNames: []string{"VARCHAR2"},
			expected:  "NULL",
		},
		{
			name:      "Quote escaping",
			values:    []interface{}{"O'Reilly"},
			typeNames: []string{"VARCHAR2"},
			expected:  "'O''Reilly'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := migration.ProcessRow(tt.values, tt.typeNames)
			if result != tt.expected {
				t.Errorf("ProcessRow(%v, %v) = %v; expected %v", tt.values, tt.typeNames, result, tt.expected)
			}
		})
	}
}
