package migration

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestOracleCopySource(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT \\* FROM test").WillReturnRows(
		sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "Alice").AddRow(2, "Bob"),
	)

	rows, err := db.Query("SELECT * FROM test")
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	defer rows.Close()

	src := &oracleCopySource{
		rows: rows,
		cols: []string{"id", "name"},
	}

	if !src.Next() {
		t.Fatal("expected true")
	}

	vals, err := src.Values()
	if err != nil {
		t.Fatalf("failed to get values: %v", err)
	}
	if len(vals) != 2 {
		t.Fatalf("expected 2 values, got %d", len(vals))
	}

	if !src.Next() {
		t.Fatal("expected true")
	}

	if src.Next() {
		t.Fatal("expected false")
	}

	if src.Err() != nil {
		t.Fatalf("expected nil error, got %v", src.Err())
	}
}

func TestSplitNames(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"A", []string{"A"}},
		{" A , B ", []string{"A", "B"}},
	}
	for _, tt := range tests {
		got := splitNames(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitNames(%q) = %v, want %v", tt.input, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitNames(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	}
}

func TestPublishRetryEvent_NilSafe(t *testing.T) {
	publishRetryEvent(nil, RetryEvent{})
}
