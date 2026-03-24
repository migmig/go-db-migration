package migration

import (
	"bufio"
	"os"
	"sync"
	"testing"

	"dbmigrator/internal/config"
	"dbmigrator/internal/dialect"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMigrateTableToFile_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	defer db.Close()

	cfg := &config.Config{
		Schema: "public",
	}
	dia := &dialect.PostgresDialect{}

	// 1 row
	mock.ExpectQuery("SELECT \\* FROM \"T1\"").WillReturnRows(
		sqlmock.NewRows([]string{"id", "val"}).AddRow(1, "A"),
	)

	state := NewMigrationState("t")
	
	f, _ := os.CreateTemp("", "test*.sql")
	defer os.Remove(f.Name())
	writer := bufio.NewWriter(f)

	count, err := MigrateTableToFile(db, dia, "T1", writer, cfg, &sync.Mutex{}, nil, state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
	writer.Flush()
	f.Close()
}
