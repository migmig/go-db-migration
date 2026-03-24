package main

import (
	"bufio"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"dbmigrator/internal/config"
	"dbmigrator/internal/db"
	"dbmigrator/internal/dialect"
	"dbmigrator/internal/migration"
	"dbmigrator/internal/security"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/mattn/go-sqlite3"
)

func newCLIUserStore(t *testing.T) *db.UserStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "auth.db")
	store, err := db.OpenUserStore(path)
	if err != nil {
		t.Fatalf("open user store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestMigrateTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql connection: %v", err)
	}
	defer db.Close()

	tableName := "MOCK_TABLE"

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"MOCK_TABLE\"").WillReturnRows(sqlmock.NewRows([]string{"COUNT"}).AddRow(2))

	rows := sqlmock.NewRows([]string{"ID", "NAME"}).
		AddRow(1, "Alice").
		AddRow(2, "Bob")

	mock.ExpectQuery("SELECT \\* FROM \"" + tableName + "\"").WillReturnRows(rows)

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

	_, err = migration.MigrateTable(db, nil, nil, dia, "MOCK_TABLE", mainBuf, cfg, &outMutex, nil, migration.NewMigrationState("test"))
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

func TestExecuteUserCommand_UserLifecycle(t *testing.T) {
	store := newCLIUserStore(t)

	exitCode, stdout, stderr := runUserCommandForTest(store, []string{"add", "alice", "password123", "--admin"})
	if exitCode != 0 {
		t.Fatalf("add exit code = %d, stderr = %s", exitCode, stderr)
	}
	if !strings.Contains(stdout, `user "alice" created`) {
		t.Fatalf("unexpected add stdout: %s", stdout)
	}

	user, err := store.GetUserByUsername("alice")
	if err != nil {
		t.Fatalf("get user alice: %v", err)
	}
	if !user.IsAdmin {
		t.Fatalf("expected alice admin=true")
	}

	exitCode, stdout, stderr = runUserCommandForTest(store, []string{"list"})
	if exitCode != 0 {
		t.Fatalf("list exit code = %d, stderr = %s", exitCode, stderr)
	}
	if !strings.Contains(stdout, "alice") {
		t.Fatalf("expected alice in list output, got: %s", stdout)
	}

	exitCode, stdout, stderr = runUserCommandForTest(store, []string{"reset-password", "alice", "newpass123"})
	if exitCode != 0 {
		t.Fatalf("reset-password exit code = %d, stderr = %s", exitCode, stderr)
	}
	if !strings.Contains(stdout, `password reset for "alice"`) {
		t.Fatalf("unexpected reset stdout: %s", stdout)
	}

	user, err = store.GetUserByUsername("alice")
	if err != nil {
		t.Fatalf("get user after reset: %v", err)
	}
	if !security.VerifyPassword(user.PasswordHash, "newpass123") {
		t.Fatalf("expected password hash to match new password")
	}

	exitCode, stdout, stderr = runUserCommandForTest(store, []string{"delete", "alice"})
	if exitCode != 0 {
		t.Fatalf("delete exit code = %d, stderr = %s", exitCode, stderr)
	}
	if !strings.Contains(stdout, `user "alice" deleted`) {
		t.Fatalf("unexpected delete stdout: %s", stdout)
	}
}

func TestExecuteUserCommand_ErrorCases(t *testing.T) {
	store := newCLIUserStore(t)

	exitCode, _, stderr := runUserCommandForTest(store, []string{"add", "short", "123"})
	if exitCode == 0 {
		t.Fatalf("expected non-zero for weak password")
	}
	if !strings.Contains(stderr, "invalid password") {
		t.Fatalf("unexpected weak password stderr: %s", stderr)
	}

	exitCode, _, stderr = runUserCommandForTest(store, []string{"reset-password", "unknown", "password123"})
	if exitCode == 0 {
		t.Fatalf("expected non-zero for unknown user reset")
	}
	if !strings.Contains(stderr, `user "unknown" not found`) {
		t.Fatalf("unexpected unknown user stderr: %s", stderr)
	}

	exitCode, _, stderr = runUserCommandForTest(store, []string{"unknown"})
	if exitCode == 0 {
		t.Fatalf("expected non-zero for invalid subcommand")
	}
	if !strings.Contains(stderr, "usage:") {
		t.Fatalf("expected usage output, got: %s", stderr)
	}

	// Extra cases
	if exitCode, _, _ = runUserCommandForTest(store, []string{}); exitCode == 0 {
		t.Error("expected failure for empty args")
	}
	if exitCode, _, _ = runUserCommandForTest(store, []string{"add", "u"}); exitCode == 0 {
		t.Error("expected failure for add with 1 arg")
	}
	if exitCode, _, _ = runUserCommandForTest(store, []string{"reset-password", "u"}); exitCode == 0 {
		t.Error("expected failure for reset-password with 1 arg")
	}
	if exitCode, _, _ = runUserCommandForTest(store, []string{"delete"}); exitCode == 0 {
		t.Error("expected failure for delete with 0 args")
	}
	if exitCode, _, stderr = runUserCommandForTest(store, []string{"delete", "unknown"}); exitCode == 0 {
		t.Error("expected failure for delete unknown user")
	}
	if !strings.Contains(stderr, "not found") {
		t.Errorf("expected not found error, got %s", stderr)
	}
}

func TestHandleUserCommand_NotUsers(t *testing.T) {
	if handleUserCommand([]string{}) {
		t.Error("expected false for empty args")
	}
	if handleUserCommand([]string{"other"}) {
		t.Error("expected false for non-users args")
	}
}

func TestRunMain_Completion(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	
	os.Args = []string{"cmd", "-completion", "bash"}
	
	exitCode := runMain()
	if exitCode != 0 {
		t.Errorf("expected 0, got %d", exitCode)
	}
}

func TestHandleUserCommand_StoreError(t *testing.T) {
	// Provide a path that is a directory to cause error on OpenUserStore (SQLite)
	t.Setenv("DBM_AUTH_DB_PATH", t.TempDir())
	
	oldExit := userCommandExit
	defer func() { userCommandExit = oldExit }()
	
	var exitCode int
	userCommandExit = func(code int) {
		exitCode = code
	}

	handled := handleUserCommand([]string{"users", "list"})
	if !handled {
		t.Error("expected true for users args")
	}
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}

func TestHandleUserCommand_Users(t *testing.T) {
	t.Setenv("DBM_AUTH_DB_PATH", filepath.Join(t.TempDir(), "auth.db"))
	
	oldExit := userCommandExit
	defer func() { userCommandExit = oldExit }()
	
	var exitCode int
	userCommandExit = func(code int) {
		exitCode = code
	}

	handled := handleUserCommand([]string{"users", "unknown"})
	if !handled {
		t.Error("expected true for users args")
	}
	if exitCode == 0 {
		t.Error("expected non-zero exit code for unknown subcommand")
	}
}

func TestRunMain_OracleError(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	
	os.Args = []string{"cmd", "-url", "invalid", "-user", "u", "-password", "p", "-tables", "T1"}
	
	exitCode := runMain()
	if exitCode != 1 {
		t.Errorf("expected 1, got %d", exitCode)
	}
}

func TestRunMain_TargetDBError(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	
	os.Args = []string{"cmd", "-target-db", "mysql"}
	
	exitCode := runMain()
	if exitCode != 1 {
		t.Errorf("expected 1, got %d", exitCode)
	}
}
