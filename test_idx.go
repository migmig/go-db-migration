//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/sijms/go-ora/v2"

	"dbmigrator/internal/migration"
)

func main() {
	db, err := sql.Open("oracle", "oracle://MY_USER:MY_PASSWORD@mini.local:1521/XE")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Test with the exact logic from migration.go
	owner := "MY_USER"
	indexes, err := migration.GetIndexMetadata(db, "SAMPLE_DATA", owner)

	fmt.Printf("Indexes found: %d\n", len(indexes))
}
