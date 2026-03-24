package migration

import (
	"os"
	"sync"
	"testing"

	"dbmigrator/internal/config"
	"dbmigrator/internal/dialect"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestWorker_ErrorHandling(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql connection: %v", err)
	}
	defer db.Close()

	// Mocking query failure for T1
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"T1\"").WillReturnRows(sqlmock.NewRows([]string{"COUNT"}).AddRow(1))
	mock.ExpectQuery("SELECT \\* FROM \"T1\"").WillReturnError(os.ErrPermission)

	cfg := &config.Config{
		Tables:    []string{"T1"},
		BatchSize: 100,
	}

	jobs := make(chan job, 1)
	var wg sync.WaitGroup
	var outMutex sync.Mutex
	
	dia := &dialect.PostgresDialect{}
	report := NewMigrationReport("job", "s", "t", "tu", "all")
	state := NewMigrationState("t")

	wg.Add(1)
	go worker(1, db, nil, nil, dia, jobs, &wg, nil, cfg, &outMutex, nil, state, report)

	jobs <- job{tableName: "T1"}
	close(jobs)
	wg.Wait()

	// Check if report recorded the error
	summary := report.ToSummary()
	if summary.ErrorCount != 1 {
		t.Errorf("expected 1 error, got %d", summary.ErrorCount)
	}
}
