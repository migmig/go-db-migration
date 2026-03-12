package main

import (
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestWriteBatch_SchemaMapping(t *testing.T) {
	// Setup
	tableName := "MY_TABLE"
	schemaPrefix := "public."
	cols := []string{"ID", "NAME"}
	batch := []string{"(1, 'John')", "(2, 'Jane')"}
	perTable := false
	var outMutex sync.Mutex

	// We need to modify writeBatch or its caller to handle schema mapping.
	// For now, let's assume we add a schema parameter to writeBatch.

	// Create a temporary file to capture output
	tmpFile, err := os.CreateTemp("", "test_migration_*.sql")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// ACTION: Call writeBatch with schema mapping
	writeBatch(tmpFile, tableName, cols, batch, perTable, "public", &outMutex)

	// VERIFY
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
	// Setup
	tableName := "MY_TABLE"
	cols := []string{"ID", "NAME"}
	batch := []string{"(1, 'John')"}
	perTable := false
	var outMutex sync.Mutex

	tmpFile, err := os.CreateTemp("", "test_migration_noschema_*.sql")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// ACTION: Call writeBatch without schema mapping
	writeBatch(tmpFile, tableName, cols, batch, perTable, "", &outMutex)

	// VERIFY
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	expectedPrefix := "INSERT INTO " + tableName
	if !strings.Contains(string(content), expectedPrefix) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedPrefix, string(content))
	}

	// Ensure no dot is present before table name if no schema
	if strings.Contains(string(content), "INSERT INTO .") {
		t.Errorf("Output should not contain '.' before table name if schema is empty")
	}
}

func TestWriteBatch_PerTable(t *testing.T) {
	// Setup
	tableName := "MY_TABLE"
	cols := []string{"ID"}
	batch := []string{"(1)"}
	perTable := true
	var outMutex sync.Mutex // Mutex should not be used when perTable is true

	tmpFile, err := os.CreateTemp("", "test_migration_pertable_*.sql")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// ACTION: Call writeBatch with perTable=true
	writeBatch(tmpFile, tableName, cols, batch, perTable, "", &outMutex)

	// VERIFY
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
		{
			name:      "BLOB/RAW mapping",
			values:    []interface{}{[]byte{0xDE, 0xAD, 0xBE, 0xEF}},
			typeNames: []string{"BLOB"},
			expected:  "'\\xdeadbeef'",
		},
		{
			name:      "RAW mapping",
			values:    []interface{}{[]byte{0x01, 0x02}},
			typeNames: []string{"RAW"},
			expected:  "'\\x0102'",
		},
		{
			name:      "Byte slice as string",
			values:    []interface{}{[]byte("data")},
			typeNames: []string{"VARCHAR2"},
			expected:  "'data'",
		},
		{
			name:      "Timestamp formatting",
			values:    []interface{}{time.Date(2023, 10, 27, 10, 30, 0, 123456789, time.UTC)},
			typeNames: []string{"TIMESTAMP"},
			expected:  "'2023-10-27 10:30:00.123456789'",
		},
		{
			name:      "Boolean TRUE",
			values:    []interface{}{true},
			typeNames: []string{"BOOL"},
			expected:  "TRUE",
		},
		{
			name:      "Boolean FALSE",
			values:    []interface{}{false},
			typeNames: []string{"BOOL"},
			expected:  "FALSE",
		},
		{
			name:      "Float number",
			values:    []interface{}{123.456},
			typeNames: []string{"NUMBER"},
			expected:  "123.456",
		},
		{
			name:      "Unknown type fallback",
			values:    []interface{}{struct{ A int }{1}},
			typeNames: []string{"OTHER"},
			expected:  "'{1}'",
		},
		{
			name:      "Unknown type with quote",
			values:    []interface{}{struct{ A string }{"o'clock"}},
			typeNames: []string{"OTHER"},
			expected:  "'{o''clock}'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := processRow(tt.values, tt.typeNames)
			if result != tt.expected {
				t.Errorf("processRow(%v, %v) = %v; expected %v", tt.values, tt.typeNames, result, tt.expected)
			}
		})
	}
}
