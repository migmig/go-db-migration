package ziputil

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestZipDirectory_BasicRoundtrip(t *testing.T) {
	srcDir := t.TempDir()

	// Create a couple of SQL files to compress
	files := map[string]string{
		"users.sql":  "INSERT INTO users VALUES (1, 'Alice');",
		"orders.sql": "INSERT INTO orders VALUES (100, 1);",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(srcDir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	zipPath := filepath.Join(t.TempDir(), "out.zip")
	if err := ZipDirectory(srcDir, zipPath); err != nil {
		t.Fatalf("ZipDirectory failed: %v", err)
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 2 {
		t.Errorf("expected 2 files in zip, got %d", len(r.File))
	}

	found := map[string]bool{}
	for _, f := range r.File {
		found[f.Name] = true
	}
	for name := range files {
		if !found[name] {
			t.Errorf("expected %q in zip, got files: %v", name, found)
		}
	}
}

func TestZipDirectory_FileContentsPreserved(t *testing.T) {
	srcDir := t.TempDir()
	content := "SELECT * FROM employees;"
	if err := os.WriteFile(filepath.Join(srcDir, "q.sql"), []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	zipPath := filepath.Join(t.TempDir(), "out.zip")
	if err := ZipDirectory(srcDir, zipPath); err != nil {
		t.Fatalf("ZipDirectory failed: %v", err)
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 1 {
		t.Fatalf("expected 1 file, got %d", len(r.File))
	}

	rc, err := r.File[0].Open()
	if err != nil {
		t.Fatalf("failed to open file in zip: %v", err)
	}
	defer rc.Close()

	buf := make([]byte, len(content)+10)
	n, _ := rc.Read(buf)
	if string(buf[:n]) != content {
		t.Errorf("content = %q, want %q", string(buf[:n]), content)
	}
}

func TestZipDirectory_EmptyDir(t *testing.T) {
	srcDir := t.TempDir() // empty directory
	zipPath := filepath.Join(t.TempDir(), "empty.zip")

	if err := ZipDirectory(srcDir, zipPath); err != nil {
		t.Fatalf("ZipDirectory should not fail on empty dir, got: %v", err)
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open empty zip: %v", err)
	}
	defer r.Close()

	if len(r.File) != 0 {
		t.Errorf("expected 0 files in zip for empty dir, got %d", len(r.File))
	}
}

func TestZipDirectory_RelativeNamesInZip(t *testing.T) {
	srcDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(srcDir, "table.sql"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	zipPath := filepath.Join(t.TempDir(), "out.zip")
	if err := ZipDirectory(srcDir, zipPath); err != nil {
		t.Fatalf("ZipDirectory failed: %v", err)
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open zip: %v", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if filepath.IsAbs(f.Name) {
			t.Errorf("zip entry name should be relative, got absolute path: %s", f.Name)
		}
	}
}

func TestZipDirectory_TargetFileCreationError(t *testing.T) {
	srcDir := t.TempDir()
	// Use an invalid target path (directory as target file) to trigger os.Create error
	err := ZipDirectory(srcDir, t.TempDir()) // a directory, not a file path
	if err == nil {
		t.Error("expected error when target is a directory, got nil")
	}
}
