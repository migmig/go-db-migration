package migration

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"dbmigrator/internal/config"
	"dbmigrator/internal/db"

	"github.com/jackc/pgx/v5"
)

// ProgressTracker is used to send progress updates to the caller (e.g., Web UI)
type ProgressTracker interface {
	Init(table string, totalRows int)
	Update(table string, processedRows int)
	Done(table string)
	Error(table string, err error)
}

// DryRunTracker extends ProgressTracker with dry-run result reporting
type DryRunTracker interface {
	DryRunResult(table string, totalRows int, connectionOk bool)
}

// DDLProgressTracker extends ProgressTracker with DDL object progress reporting
type DDLProgressTracker interface {
	DDLProgress(object, name, status string, err error)
}

type job struct {
	tableName string
}

// resolveOwner returns the effective Oracle schema owner.
// Falls back to uppercase User when OracleOwner is not set.
func resolveOwner(cfg *config.Config) string {
	if cfg.OracleOwner != "" {
		return strings.ToUpper(cfg.OracleOwner)
	}
	return strings.ToUpper(cfg.User)
}

// splitNames splits a comma-separated string into a trimmed slice.
func splitNames(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func Run(dbConn *sql.DB, pgPool db.PGPool, cfg *config.Config, tracker ProgressTracker) error {
	if cfg.OutputDir != "" {
		if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}

	if cfg.DryRun {
		slog.Info("Dry run mode enabled. Verifying connectivity and estimating row counts.")
		dryTracker, hasDryTracker := tracker.(DryRunTracker)
		for _, table := range cfg.Tables {
			var count int
			err := dbConn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
			if err != nil {
				slog.Error("failed to get row count for table", "table", table, "error", err)
				if tracker != nil {
					tracker.Error(table, err)
				}
				continue
			}
			slog.Info("table estimation", "table", table, "estimated_rows", count)
			if hasDryTracker {
				dryTracker.DryRunResult(table, count, true)
			}
		}
		slog.Info("Dry run completed successfully.")
		return nil
	}

	var mainOut *os.File
	var mainBuf *bufio.Writer
	var err error

	// If not direct migration, setup output file
	if pgPool == nil && !cfg.PerTable {
		outFile := cfg.OutFile
		if cfg.OutputDir != "" {
			outFile = cfg.OutputDir + "/" + outFile
		}
		mainOut, err = os.Create(outFile)
		if err != nil {
			return fmt.Errorf("error creating output file: %v", err)
		}
		defer mainOut.Close()
		mainBuf = bufio.NewWriter(mainOut)
		defer mainBuf.Flush()
	}

	var outMutex sync.Mutex
	numWorkers := 1
	if cfg.Parallel {
		numWorkers = cfg.Workers
	}

	jobs := make(chan job, len(cfg.Tables))
	var wg sync.WaitGroup

	// Start workers
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(w, dbConn, pgPool, jobs, &wg, mainBuf, cfg, &outMutex, tracker)
	}

	// Send jobs
	for _, table := range cfg.Tables {
		jobs <- job{tableName: table}
	}
	close(jobs)

	wg.Wait()
	return nil
}

func worker(id int, dbConn *sql.DB, pgPool db.PGPool, jobs <-chan job, wg *sync.WaitGroup, mainBuf *bufio.Writer, cfg *config.Config, outMutex *sync.Mutex, tracker ProgressTracker) {
	defer wg.Done()
	for j := range jobs {
		slog.Info("worker processing table", "worker_id", id, "table", j.tableName)
		err := MigrateTable(dbConn, pgPool, j.tableName, mainBuf, cfg, outMutex, tracker)
		if err != nil {
			slog.Error("error migrating table", "table", j.tableName, "error", err)
			if tracker != nil {
				tracker.Error(j.tableName, err)
			}
		}
	}
}

func MigrateTable(dbConn *sql.DB, pgPool db.PGPool, tableName string, mainBuf *bufio.Writer, cfg *config.Config, outMutex *sync.Mutex, tracker ProgressTracker) error {
	var totalRows int
	_ = dbConn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&totalRows)
	if tracker != nil {
		tracker.Init(tableName, totalRows)
	}

	var err error
	if pgPool != nil {
		err = MigrateTableDirect(dbConn, pgPool, tableName, cfg, tracker)
	} else {
		err = MigrateTableToFile(dbConn, tableName, mainBuf, cfg, outMutex, tracker)
	}

	if err == nil && tracker != nil {
		tracker.Done(tableName)
	}
	return err
}

func MigrateTableDirect(dbConn *sql.DB, pgPool db.PGPool, tableName string, cfg *config.Config, tracker ProgressTracker) error {
	slog.Info("direct migration started", "table", tableName)

	ddlTracker, hasDDLTracker := tracker.(DDLProgressTracker)
	owner := resolveOwner(cfg)

	if cfg.WithDDL {
		// Sequence DDLs — executed before CREATE TABLE
		if cfg.WithSequences {
			seqs, err := GetSequenceMetadata(dbConn, tableName, owner, splitNames(cfg.Sequences))
			if err != nil {
				slog.Warn("failed to get sequence metadata", "table", tableName, "error", err)
			} else {
				for _, seq := range seqs {
					ddl := GenerateSequenceDDL(seq, cfg.Schema)
					if _, err := pgPool.Exec(context.Background(), ddl); err != nil {
						slog.Warn("failed to execute sequence DDL", "sequence", seq.Name, "error", err)
						if hasDDLTracker {
							ddlTracker.DDLProgress("sequence", seq.Name, "error", err)
						}
					} else {
						slog.Info("sequence DDL executed", "sequence", seq.Name)
						if hasDDLTracker {
							ddlTracker.DDLProgress("sequence", seq.Name, "ok", nil)
						}
					}
				}
			}
		}

		colsMeta, err := GetTableMetadata(dbConn, tableName)
		if err != nil {
			slog.Warn("failed to get table metadata for DDL", "table", tableName, "error", err)
		} else {
			ddl := GenerateCreateTableDDL(tableName, cfg.Schema, colsMeta)
			_, err = pgPool.Exec(context.Background(), ddl)
			if err != nil {
				return fmt.Errorf("failed to execute DDL for %s: %v", tableName, err)
			}
			slog.Info("DDL executed successfully", "table", tableName)
		}
	} else {
		// Validation check if not using DDL
		schema := cfg.Schema
		if schema == "" {
			schema = "public"
		}
		exists, err := db.TableExists(context.Background(), pgPool, schema, tableName)
		if err != nil {
			return fmt.Errorf("failed to check table existence for %s: %v", tableName, err)
		}
		if !exists {
			return fmt.Errorf("target table %s.%s does not exist. Use --with-ddl to create it automatically", schema, tableName)
		}
	}

	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	rows, err := dbConn.Query(query)
	if err != nil {
		return fmt.Errorf("failed to execute query on Oracle table %s: %v", tableName, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	// Use COPY for high performance
	ctx := context.Background()

	// Transaction per table
	tx, err := pgPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start pg transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	rowCount := 0
	// We need to implement pgx.CopyFromSource
	source := &oracleCopySource{
		rows: rows,
		cols: cols,
		err:  nil,
	}

	n, err := tx.CopyFrom(ctx, pgx.Identifier{cfg.Schema, tableName}, cols, source)
	if err != nil {
		return fmt.Errorf("COPY failed for %s: %v", tableName, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction for %s: %v", tableName, err)
	}

	rowCount = int(n)

	// Index DDLs — executed after COPY for performance
	if cfg.WithDDL && cfg.WithIndexes {
		indexes, err := GetIndexMetadata(dbConn, tableName, owner)
		if err != nil {
			slog.Warn("failed to get index metadata", "table", tableName, "error", err)
		} else {
			for _, idx := range indexes {
				ddl := GenerateIndexDDL(idx, tableName, cfg.Schema)
				if _, err := pgPool.Exec(context.Background(), ddl); err != nil {
					slog.Warn("failed to execute index DDL", "index", idx.Name, "error", err)
					if hasDDLTracker {
						ddlTracker.DDLProgress("index", idx.Name, "error", err)
					}
				} else {
					slog.Info("index DDL executed", "index", idx.Name)
					if hasDDLTracker {
						ddlTracker.DDLProgress("index", idx.Name, "ok", nil)
					}
				}
			}
		}
	}

	slog.Info("direct migration finished", "table", tableName, "rows", rowCount)
	return nil
}

type oracleCopySource struct {
	rows *sql.Rows
	cols []string
	err  error
}

func (s *oracleCopySource) Next() bool {
	return s.rows.Next()
}

func (s *oracleCopySource) Values() ([]interface{}, error) {
	values := make([]interface{}, len(s.cols))
	valuePtrs := make([]interface{}, len(s.cols))
	for i := range s.cols {
		valuePtrs[i] = &values[i]
	}

	if err := s.rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}

	// Post-process values for Postgres (convert time.Time, handle types if needed)
	// pgx handles time.Time and basic types well, so we mostly pass as is.
	// For raw binary, Oracle driver gives []byte which pgx also handles for bytea.
	return values, nil
}

func (s *oracleCopySource) Err() error {
	return s.rows.Err()
}

func MigrateTableToFile(dbConn *sql.DB, tableName string, mainBuf *bufio.Writer, cfg *config.Config, outMutex *sync.Mutex, tracker ProgressTracker) error {
	slog.Info("file-based migration started", "table", tableName)

	var tableBuf *bufio.Writer
	if cfg.PerTable {
		fileName := fmt.Sprintf("%s.sql", tableName)
		if cfg.OutputDir != "" {
			fileName = cfg.OutputDir + "/" + fileName
		}
		f, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("failed to create output file for %s: %v", tableName, err)
		}
		defer f.Close()
		tableBuf = bufio.NewWriter(f)
		defer tableBuf.Flush()
	} else {
		tableBuf = mainBuf
	}

	ddlTracker, hasDDLTracker := tracker.(DDLProgressTracker)

	if cfg.WithDDL {
		owner := resolveOwner(cfg)

		// Sequence DDLs — written before CREATE TABLE
		if cfg.WithSequences {
			seqs, err := GetSequenceMetadata(dbConn, tableName, owner, splitNames(cfg.Sequences))
			if err != nil {
				slog.Warn("failed to get sequence metadata", "table", tableName, "error", err)
			} else {
				for _, seq := range seqs {
					ddl := GenerateSequenceDDL(seq, cfg.Schema)
					writeToBuf(tableBuf, ddl, cfg.PerTable, outMutex)
					if hasDDLTracker {
						ddlTracker.DDLProgress("sequence", seq.Name, "ok", nil)
					}
					slog.Info("sequence DDL written", "sequence", seq.Name)
				}
			}
		}

		// CREATE TABLE DDL
		colsMeta, err := GetTableMetadata(dbConn, tableName)
		if err != nil {
			slog.Warn("failed to get table metadata for DDL", "table", tableName, "error", err)
		} else {
			ddl := GenerateCreateTableDDL(tableName, cfg.Schema, colsMeta)
			writeToBuf(tableBuf, ddl+"\n", cfg.PerTable, outMutex)
		}

		// Index DDLs — written after CREATE TABLE, before INSERT
		if cfg.WithIndexes {
			indexes, err := GetIndexMetadata(dbConn, tableName, owner)
			if err != nil {
				slog.Warn("failed to get index metadata", "table", tableName, "error", err)
			} else {
				for _, idx := range indexes {
					ddl := GenerateIndexDDL(idx, tableName, cfg.Schema)
					writeToBuf(tableBuf, ddl, cfg.PerTable, outMutex)
					if hasDDLTracker {
						ddlTracker.DDLProgress("index", idx.Name, "ok", nil)
					}
					slog.Info("index DDL written", "index", idx.Name)
				}
			}
		}
	}

	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	rows, err := dbConn.Query(query)
	if err != nil {
		return fmt.Errorf("failed to execute query on %s: %v", tableName, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns for %s: %v", tableName, err)
	}

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

		typeNames := make([]string, len(colTypes))
		for i, ct := range colTypes {
			typeNames[i] = ct.DatabaseTypeName()
		}
		rowValues := ProcessRow(values, typeNames)
		batch = append(batch, fmt.Sprintf("(%s)", rowValues))
		rowCount++

		if len(batch) >= cfg.BatchSize {
			WriteBatch(tableBuf, tableName, cols, batch, cfg, outMutex)
			batch = batch[:0]
			if tracker != nil {
				tracker.Update(tableName, rowCount)
			}
		}
	}

	if len(batch) > 0 {
		WriteBatch(tableBuf, tableName, cols, batch, cfg, outMutex)
		if tracker != nil {
			tracker.Update(tableName, rowCount)
		}
	}

	slog.Info("file-based migration finished", "table", tableName, "rows", rowCount)
	return nil
}

func ProcessRow(values []interface{}, typeNames []string) string {
	var row []string

	for i, val := range values {
		if val == nil {
			row = append(row, "NULL")
			continue
		}

		dbTypeName := typeNames[i]

		switch v := val.(type) {
		case []byte:
			if strings.Contains(strings.ToUpper(dbTypeName), "BLOB") || strings.Contains(strings.ToUpper(dbTypeName), "RAW") {
				row = append(row, fmt.Sprintf("'\\x%s'", hex.EncodeToString(v)))
			} else {
				str := string(v)
				escaped := strings.ReplaceAll(str, "'", "''")
				row = append(row, fmt.Sprintf("'%s'", escaped))
			}
		case string:
			escaped := strings.ReplaceAll(v, "'", "''")
			row = append(row, fmt.Sprintf("'%s'", escaped))
		case time.Time:
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
			str := fmt.Sprintf("%v", v)
			escaped := strings.ReplaceAll(str, "'", "''")
			row = append(row, fmt.Sprintf("'%s'", escaped))
		}
	}

	return strings.Join(row, ", ")
}

func writeToBuf(buf *bufio.Writer, s string, perTable bool, outMutex *sync.Mutex) {
	if !perTable {
		outMutex.Lock()
		defer outMutex.Unlock()
	}
	buf.WriteString(s)
}

func WriteBatch(out *bufio.Writer, tableName string, cols []string, batch []string, cfg *config.Config, outMutex *sync.Mutex) {
	colStr := strings.Join(cols, ", ")
	valStr := strings.Join(batch, ",\n    ")

	fullTableName := tableName
	if cfg.Schema != "" {
		fullTableName = fmt.Sprintf("%s.%s", cfg.Schema, tableName)
	}

	stmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES\n    %s;\n\n", fullTableName, colStr, valStr)

	if !cfg.PerTable {
		outMutex.Lock()
		defer outMutex.Unlock()
	}

	_, err := out.WriteString(stmt)
	if err != nil {
		slog.Warn("failed to write batch", "table", tableName, "error", err)
	}
}
