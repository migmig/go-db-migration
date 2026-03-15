package config

import (
	"flag"
	"io"
	"os"
	"strings"
	"testing"
)

// resetFlags resets the global flag.CommandLine so ParseFlags can be called
// multiple times across tests without "flag redefined" panics.
func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestParseFlags_WebMode(t *testing.T) {
	resetFlags()
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"cmd", "-web"}

	cfg, err := ParseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.WebMode {
		t.Error("expected WebMode=true")
	}
}

func TestParseFlags_WebMode_SkipsRequiredFieldValidation(t *testing.T) {
	resetFlags()
	old := os.Args
	defer func() { os.Args = old }()
	// No url/user/password/tables — should still succeed because WebMode is set
	os.Args = []string{"cmd", "-web"}

	_, err := ParseFlags()
	if err != nil {
		t.Errorf("expected no error in web mode without connection flags, got: %v", err)
	}
}

func TestParseFlags_MissingRequiredFlags(t *testing.T) {
	resetFlags()
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"cmd"}

	_, err := ParseFlags()
	if err == nil {
		t.Error("expected error when required flags are missing, got nil")
	}
}

func TestParseFlags_TablesSplitAndTrimmed(t *testing.T) {
	resetFlags()
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"cmd", "-url=host/svc", "-user=u", "-password=p", "-tables=FOO , BAR , BAZ"}

	cfg, err := ParseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Tables) != 3 {
		t.Fatalf("expected 3 tables, got %d: %v", len(cfg.Tables), cfg.Tables)
	}
	expected := []string{"FOO", "BAR", "BAZ"}
	for i, want := range expected {
		if cfg.Tables[i] != want {
			t.Errorf("Tables[%d] = %q, want %q", i, cfg.Tables[i], want)
		}
	}
}

func TestParseFlags_SingleTable(t *testing.T) {
	resetFlags()
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"cmd", "-url=host/svc", "-user=u", "-password=p", "-tables=ORDERS"}

	cfg, err := ParseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Tables) != 1 || cfg.Tables[0] != "ORDERS" {
		t.Errorf("unexpected tables: %v", cfg.Tables)
	}
}

func TestParseFlags_Defaults(t *testing.T) {
	resetFlags()
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"cmd", "-url=host/svc", "-user=u", "-password=p", "-tables=T"}

	cfg, err := ParseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutFile != "migration.sql" {
		t.Errorf("OutFile default = %q, want %q", cfg.OutFile, "migration.sql")
	}
	if cfg.BatchSize != 1000 {
		t.Errorf("BatchSize default = %d, want 1000", cfg.BatchSize)
	}
	if cfg.Workers != 4 {
		t.Errorf("Workers default = %d, want 4", cfg.Workers)
	}
	if cfg.PerTable {
		t.Error("PerTable should default to false")
	}
	if cfg.Parallel {
		t.Error("Parallel should default to false")
	}
	if cfg.WithDDL {
		t.Error("WithDDL should default to false")
	}
	if cfg.DryRun {
		t.Error("DryRun should default to false")
	}
	if cfg.LogJSON {
		t.Error("LogJSON should default to false")
	}
	if cfg.AuthEnabled {
		t.Error("AuthEnabled should default to false")
	}
	if cfg.MasterKey != "" {
		t.Errorf("MasterKey default = %q, want empty", cfg.MasterKey)
	}
}

func TestParseFlags_ExplicitFlags(t *testing.T) {
	resetFlags()
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{
		"cmd",
		"-url=myhost:1522/myservice",
		"-user=scott",
		"-password=tiger",
		"-tables=T1",
		"-out=output.sql",
		"-batch=500",
		"-schema=myschema",
		"-per-table",
		"-parallel",
		"-pg-url=postgres://u:p@localhost/db",
		"-workers=8",
		"-with-ddl",
		"-dry-run",
		"-log-json",
		"-auth-enabled",
	}

	cfg, err := ParseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OracleURL != "myhost:1522/myservice" {
		t.Errorf("OracleURL = %q, want %q", cfg.OracleURL, "myhost:1522/myservice")
	}
	if cfg.User != "scott" {
		t.Errorf("User = %q, want %q", cfg.User, "scott")
	}
	if cfg.Password != "tiger" {
		t.Errorf("Password = %q, want %q", cfg.Password, "tiger")
	}
	if cfg.OutFile != "output.sql" {
		t.Errorf("OutFile = %q, want %q", cfg.OutFile, "output.sql")
	}
	if cfg.BatchSize != 500 {
		t.Errorf("BatchSize = %d, want 500", cfg.BatchSize)
	}
	if cfg.Schema != "myschema" {
		t.Errorf("Schema = %q, want %q", cfg.Schema, "myschema")
	}
	if !cfg.PerTable {
		t.Error("expected PerTable=true")
	}
	if !cfg.Parallel {
		t.Error("expected Parallel=true")
	}
	if cfg.PGURL != "postgres://u:p@localhost/db" {
		t.Errorf("PGURL = %q, want %q", cfg.PGURL, "postgres://u:p@localhost/db")
	}
	if cfg.Workers != 8 {
		t.Errorf("Workers = %d, want 8", cfg.Workers)
	}
	if !cfg.WithDDL {
		t.Error("expected WithDDL=true")
	}
	if !cfg.DryRun {
		t.Error("expected DryRun=true")
	}
	if !cfg.LogJSON {
		t.Error("expected LogJSON=true")
	}
	if !cfg.AuthEnabled {
		t.Error("expected AuthEnabled=true")
	}
}

func TestParseFlags_CompletionMode_SkipsRequiredFieldValidation(t *testing.T) {
	resetFlags()
	old := os.Args
	defer func() { os.Args = old }()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	os.Args = []string{"cmd", "-completion=bash"}

	cfg, err := ParseFlags()
	if err != nil {
		t.Fatalf("expected no error in completion mode, got: %v", err)
	}
	if cfg.CompletionShell != "bash" {
		t.Fatalf("expected CompletionShell=bash, got %q", cfg.CompletionShell)
	}

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read completion output: %v", err)
	}
	if !strings.Contains(string(out), "complete -F _dbmigrator_completions dbmigrator") {
		t.Fatalf("expected bash completion output, got: %s", string(out))
	}
}

func TestParseFlags_CompletionMode_UnsupportedShell(t *testing.T) {
	resetFlags()
	old := os.Args
	defer func() { os.Args = old }()
	os.Args = []string{"cmd", "-completion=tcsh"}

	_, err := ParseFlags()
	if err == nil {
		t.Fatal("expected error for unsupported completion shell")
	}
	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseFlags_CompletionNoArgs_DetectsShell_Success(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	oldEnv := os.Getenv("SHELL")
	oldExit := osExit
	defer func() {
		os.Args = oldArgs
		os.Setenv("SHELL", oldEnv)
		osExit = oldExit
	}()

	os.Setenv("SHELL", "/bin/bash")
	os.Args = []string{"cmd", "-completion"}

	exitCode := -1
	osExit = func(code int) {
		exitCode = code
	}

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	_, _ = ParseFlags()

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read completion output: %v", err)
	}

	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d", exitCode)
	}
	if !strings.Contains(string(out), "complete -F _dbmigrator_completions dbmigrator") {
		t.Errorf("expected bash completion output, got: %s", string(out))
	}
}

func TestParseFlags_CompletionNoArgs_UnsupportedShell_Fails(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	oldEnv := os.Getenv("SHELL")
	oldExit := osExit
	defer func() {
		os.Args = oldArgs
		os.Setenv("SHELL", oldEnv)
		osExit = oldExit
	}()

	os.Setenv("SHELL", "/usr/bin/tcsh")
	os.Args = []string{"cmd", "-completion"}

	exitCode := -1
	osExit = func(code int) {
		exitCode = code
	}

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	_, _ = ParseFlags()

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read usage output: %v", err)
	}

	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(string(out), "자동 감지된 쉘이 지원되지 않거나 알 수 없습니다") {
		t.Errorf("expected usage output, got: %s", string(out))
	}
}


func TestParseFlags_LoadsMasterKeyFromEnv(t *testing.T) {
	resetFlags()
	oldArgs := os.Args
	oldMaster := os.Getenv("DBM_MASTER_KEY")
	defer func() {
		os.Args = oldArgs
		_ = os.Setenv("DBM_MASTER_KEY", oldMaster)
	}()

	_ = os.Setenv("DBM_MASTER_KEY", "test-master-key")
	os.Args = []string{"cmd", "-url=host/svc", "-user=u", "-password=p", "-tables=T"}

	cfg, err := ParseFlags()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.MasterKey != "test-master-key" {
		t.Fatalf("MasterKey = %q, want %q", cfg.MasterKey, "test-master-key")
	}
}
