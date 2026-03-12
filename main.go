package main

import (
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	go_ora "github.com/sijms/go-ora/v2"
)

func main() {
	dbURL := flag.String("url", "", "Oracle Database URL (e.g., host:port/service_name)")
	user := flag.String("user", "", "Database username")
	password := flag.String("password", "", "Database password")
	tablesFlag := flag.String("tables", "", "Comma-separated list of tables to migrate")
	outFile := flag.String("out", "migration.sql", "Output SQL file name")
	batchSize := flag.Int("batch", 1000, "Number of rows per bulk insert")

	flag.Parse()

	if *dbURL == "" || *user == "" || *password == "" || *tablesFlag == "" {
		fmt.Println("Usage: dbmigrator -url <dburl> -user <username> -password <password> -tables <table1,table2> [-out <outfile>] [-batch <size>]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	tables := strings.Split(*tablesFlag, ",")
	for i := range tables {
		tables[i] = strings.TrimSpace(tables[i])
	}

	log.Printf("Connecting to Oracle DB at %s as %s...", *dbURL, *user)
	log.Printf("Tables to migrate: %v", tables)
	log.Printf("Output file: %s", *outFile)
	log.Printf("Batch size: %d", *batchSize)

	// Create Output File
	out, err := os.Create(*outFile)
	if err != nil {
		log.Fatalf("Error creating output file: %v\n", err)
	}
	defer out.Close()

	// Parse Oracle DSN
	// e.g., oracle://user:password@host:port/service_name
	// if url doesn't contain oracle://, let's format it.
	dsn := *dbURL
	if !strings.HasPrefix(dsn, "oracle://") {
		// Attempt to use go-ora connection string builder
		serverParts := strings.Split(dsn, "/")
		hostPort := serverParts[0]
		serviceName := ""
		if len(serverParts) > 1 {
			serviceName = serverParts[1]
		}

		host := hostPort
		port := 1521
		if strings.Contains(hostPort, ":") {
			parts := strings.Split(hostPort, ":")
			host = parts[0]
			fmt.Sscanf(parts[1], "%d", &port)
		}

		dsn = go_ora.BuildUrl(host, port, serviceName, *user, *password, nil)
	}

	db, err := sql.Open("oracle", dsn)
	if err != nil {
		log.Fatalf("Error opening connection: %v\n", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to database: %v\n", err)
	}
	log.Println("Connected to Oracle Database successfully!")

	for _, table := range tables {
		err = migrateTable(db, table, out, *batchSize)
		if err != nil {
			log.Printf("Error migrating table %s: %v\n", table, err)
		}
	}
	log.Println("Migration completed successfully!")
}

func migrateTable(db *sql.DB, tableName string, out *os.File, batchSize int) error {
	log.Printf("Processing table: %s...\n", tableName)

	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to execute query on %s: %v", tableName, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns for %s: %v", tableName, err)
	}

	// Fetch column types to identify CLOB/BLOB/DATE
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return fmt.Errorf("failed to get column types for %s: %v", tableName, err)
	}

	values := make([]interface{}, len(cols))
	valuePtrs := make([]interface{}, len(cols))
	for i := range cols {
		valuePtrs[i] = &values[i]
	}

	rowCount := 0
	var batch []string

	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}

		// Process row
		rowValues := processRow(values, colTypes)
		batch = append(batch, fmt.Sprintf("(%s)", rowValues))
		rowCount++

		if len(batch) >= batchSize {
			writeBatch(out, tableName, cols, batch)
			// keep the underlying array
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		writeBatch(out, tableName, cols, batch)
	}

	log.Printf("Finished processing table: %s (%d rows)\n", tableName, rowCount)
	return nil
}

func processRow(values []interface{}, colTypes []*sql.ColumnType) string {
	var row []string

	for i, val := range values {
		if val == nil {
			row = append(row, "NULL")
			continue
		}

		// Handle specific type conversions
		dbTypeName := colTypes[i].DatabaseTypeName()

		switch v := val.(type) {
		case []byte:
			if strings.Contains(strings.ToUpper(dbTypeName), "BLOB") || strings.Contains(strings.ToUpper(dbTypeName), "RAW") {
				// Likely BLOB or RAW, convert to bytea for postgres
				row = append(row, fmt.Sprintf("'\\x%s'", hex.EncodeToString(v)))
			} else {
				// Sometimes strings come through as []byte
				str := string(v)
				escaped := strings.ReplaceAll(str, "'", "''")
				row = append(row, fmt.Sprintf("'%s'", escaped))
			}
		case string:
			// Strings, likely CLOB or VARCHAR, escape quotes
			escaped := strings.ReplaceAll(v, "'", "''")
			row = append(row, fmt.Sprintf("'%s'", escaped))
		case time.Time:
			// Date or Timestamp, preserve precision
			row = append(row, fmt.Sprintf("'%s'", v.Format("2006-01-02 15:04:05.999999999")))
		case int, int64, float64:
			row = append(row, fmt.Sprintf("%v", v))
		case bool:
			if v {
				row = append(row, "TRUE")
			} else {
				row = append(row, "FALSE")
			}
		default:
			// Fallback to string representation for other types
			str := fmt.Sprintf("%v", v)
			escaped := strings.ReplaceAll(str, "'", "''")
			row = append(row, fmt.Sprintf("'%s'", escaped))
		}
	}

	return strings.Join(row, ", ")
}

func writeBatch(out *os.File, tableName string, cols []string, batch []string) {
	// Build insert statement
	colStr := strings.Join(cols, ", ")
	valStr := strings.Join(batch, ",\n    ")
	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES\n    %s;\n\n", tableName, colStr, valStr)
	_, err := out.WriteString(stmt)
	if err != nil {
		log.Printf("Warning: failed to write batch for table %s: %v\n", tableName, err)
	}
}
