package migration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMaskPassword(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"postgres://user:secret@localhost:5432/db", "postgres://user:***@localhost:5432/db"},
		{"oracle://scott:tiger@host:1521/ORCL", "oracle://scott:***@host:1521/ORCL"},
		{"mysql://admin:p@ssw0rd@host/db", "mysql://admin:***@host/db"},
		{"no-password-url", "no-password-url"},
		{"", ""},
	}
	for _, tc := range cases {
		got := maskPassword(tc.input)
		if got != tc.expected {
			t.Errorf("maskPassword(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Millisecond, "500ms"},
		{12300 * time.Millisecond, "12.3s"},
		{5*time.Minute + 23*time.Second, "5m23s"},
	}
	for _, tc := range cases {
		got := formatDuration(tc.d)
		if got != tc.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

func TestFormatCount(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1,000"},
		{9999, "9,999"},
		{10000, "10K"},
		{50000, "50K"},
		{1500000, "1.5M"},
		{2100000, "2.1M"},
	}
	for _, tc := range cases {
		got := formatCount(tc.n)
		if got != tc.want {
			t.Errorf("formatCount(%d) = %q, want %q", tc.n, got, tc.want)
		}
	}
}

func TestMigrationReport_StartTable(t *testing.T) {
	r := NewMigrationReport("test-job", "oracle://user:pass@host/ORCL", "postgres", "postgres://user:pass@host/db")

	// 소스 URL 비밀번호 마스킹 확인
	if r.SourceURL != "oracle://user:***@host/ORCL" {
		t.Errorf("SourceURL not masked: %q", r.SourceURL)
	}

	// 성공 케이스
	finish1 := r.StartTable("USERS", true)
	finish1(50000, nil)

	// 실패 케이스
	finish2 := r.StartTable("ORDERS", false)
	finish2(100, fmt.Errorf("connection lost"))

	if r.SuccessCount != 1 {
		t.Errorf("SuccessCount = %d, want 1", r.SuccessCount)
	}
	if r.ErrorCount != 1 {
		t.Errorf("ErrorCount = %d, want 1", r.ErrorCount)
	}
	if r.TotalRows != 50100 {
		t.Errorf("TotalRows = %d, want 50100", r.TotalRows)
	}

	if len(r.Tables) != 2 {
		t.Fatalf("Tables count = %d, want 2", len(r.Tables))
	}
	if r.Tables[0].Status != "ok" {
		t.Errorf("Tables[0].Status = %q, want ok", r.Tables[0].Status)
	}
	if r.Tables[1].Status != "error" {
		t.Errorf("Tables[1].Status = %q, want error", r.Tables[1].Status)
	}
	if len(r.Tables[1].Errors) == 0 {
		t.Error("Tables[1].Errors should not be empty")
	}
}

func TestMigrationReport_Finalize(t *testing.T) {
	// 임시 디렉토리에서 테스트 (실제 .migration_state 폴더 오염 방지)
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmp)
	defer os.Chdir(origDir)

	r := NewMigrationReport("job-finalize-test", "oracle://u:p@h/S", "postgres", "postgres://u:p@h/db")
	r.UserID = 42
	finish := r.StartTable("USERS", true)
	finish(100, nil)

	if err := r.Finalize(); err != nil {
		t.Fatalf("Finalize() error: %v", err)
	}

	reportPath := filepath.Join(".migration_state", "job-finalize-test_report.json")
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("Report file not found: %v", err)
	}

	var decoded MigrationReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("JSON unmarshal error: %v", err)
	}
	if decoded.JobID != "job-finalize-test" {
		t.Errorf("JobID = %q, want job-finalize-test", decoded.JobID)
	}
	if decoded.UserID != 42 {
		t.Errorf("UserID = %d, want 42", decoded.UserID)
	}
	if decoded.SuccessCount != 1 {
		t.Errorf("SuccessCount = %d, want 1", decoded.SuccessCount)
	}
	if decoded.DurationHuman == "" {
		t.Error("DurationHuman should not be empty after Finalize")
	}
}
