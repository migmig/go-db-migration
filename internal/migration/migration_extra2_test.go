package migration

import (
	"fmt"
	"sync"
	"testing"

	"dbmigrator/internal/config"
	"dbmigrator/internal/dialect"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMigrateTable_QueryFail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{
		Schema: "public",
	}
	dia := &dialect.PostgresDialect{}

	// Mocking query failure
	mock.ExpectQuery("SELECT \\* FROM \"T1\"").WillReturnError(fmt.Errorf("query error"))

	state := NewMigrationState("t")

	_, err = MigrateTable(db, nil, nil, dia, "T1", nil, cfg, &sync.Mutex{}, nil, state)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMigrateTable_NoDialect(t *testing.T) {
	db, _, _ := sqlmock.New()
	defer db.Close()
	
	state := NewMigrationState("t")
	_, err := MigrateTable(db, nil, nil, nil, "T1", nil, &config.Config{}, &sync.Mutex{}, nil, state)
	if err == nil {
		t.Error("expected error for nil dialect")
	}
}
