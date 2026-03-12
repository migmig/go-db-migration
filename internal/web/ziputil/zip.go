package ziputil

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ZipDirectory compresses the given directory into a zip file.
func ZipDirectory(sourceDir string, targetFile string) error {
	zipfile, err := os.Create(targetFile)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %v", err)
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Set the header name to be relative to the source directory
		header.Name, err = filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		header.Method = zip.Deflate

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})

	return nil
}
