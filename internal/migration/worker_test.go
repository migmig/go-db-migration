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

func TestWorkerPool(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(origDir)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql connection: %v", err)
	}
	defer db.Close()

	mock.MatchExpectationsInOrder(false)

	// Setup expectations for 2 tables
	tables := []string{"T1", "T2"}
	for _, table := range tables {
		rows := sqlmock.NewRows([]string{"ID"}).AddRow(1)
		mock.ExpectQuery("SELECT \\* FROM \"" + table + "\"").WillReturnRows(rows)
	}

	cfg := &config.Config{
		Parallel:  true,
		Workers:   2,
		Tables:    tables,
		BatchSize: 100,
		PerTable:  true,
		OutputDir: tmp,
	}

	// Setup row counts for progress tracker queries
	for _, table := range tables {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"" + table + "\"").WillReturnRows(sqlmock.NewRows([]string{"COUNT"}).AddRow(1))
	}

	dia := &dialect.PostgresDialect{}
	_, err = Run(db, nil, nil, dia, cfg, nil)
	if err != nil {
		t.Errorf("Run failed: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}

func TestWorkerSingleFile(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql connection: %v", err)
	}
	defer db.Close()

	mock.MatchExpectationsInOrder(false)

	tables := []string{"T1", "T2"}
	for _, table := range tables {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"" + table + "\"").WillReturnRows(sqlmock.NewRows([]string{"COUNT"}).AddRow(1))
		rows := sqlmock.NewRows([]string{"ID"}).AddRow(1)
		mock.ExpectQuery("SELECT \\* FROM \"" + table + "\"").WillReturnRows(rows)
	}

	tmpFile, err := os.CreateTemp("", "worker_test_*.sql")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	cfg := &config.Config{
		Parallel:  true,
		Workers:   2,
		Tables:    tables,
		BatchSize: 100,
		PerTable:  false,
		OutFile:   tmpFile.Name(),
	}

	// Run logic
	jobs := make(chan job, len(cfg.Tables))
	var wg sync.WaitGroup
	var outMutex sync.Mutex
	mainBuf := bufio.NewWriter(tmpFile)

	dia := &dialect.PostgresDialect{}
	report := NewMigrationReport("test", "oracle://user:pass@host/SID", "postgres", "")
	for w := 1; w <= cfg.Workers; w++ {
		wg.Add(1)
		go worker(w, db, nil, nil, dia, jobs, &wg, mainBuf, cfg, &outMutex, nil, NewMigrationState("test"), report)
	}

	for _, table := range cfg.Tables {
		jobs <- job{tableName: table}
	}
	close(jobs)
	wg.Wait()
	mainBuf.Flush()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("There were unfulfilled expectations: %s", err)
	}
}
