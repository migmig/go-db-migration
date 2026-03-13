package migration

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"

	"dbmigrator/internal/config"
	"dbmigrator/internal/db"
	"dbmigrator/internal/dialect"

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

// WarningTracker extends ProgressTracker with warning broadcasting.
type WarningTracker interface {
	Warning(message string)
}

type job struct {
	tableName string
}

// tryConnectTarget attempts to open and ping the target database.
// Returns true if the connection succeeds, false otherwise.
func tryConnectTarget(dia dialect.Dialect, targetURL string) bool {
	if dia.Name() == "postgres" {
		pool, err := db.ConnectPostgres(targetURL, 0, 0, 0)
		if err != nil {
			return false
		}
		pool.Close()
		return true
	}

	conn, err := sql.Open(dia.DriverName(), dia.NormalizeURL(targetURL))
	if err != nil {
		return false
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		return false
	}
	return true
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

func Run(dbConn *sql.DB, targetDB *sql.DB, pgPool db.PGPool, dia dialect.Dialect, cfg *config.Config, tracker ProgressTracker) error {
	if cfg.OutputDir != "" {
		if err := os.MkdirAll(cfg.OutputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}

	jobID := cfg.ResumeJobID
	if jobID == "" {
		jobID = fmt.Sprintf("job_%s", os.Getenv("SESSION_ID")) // or just a timestamp from config? We don't have it in config. Let's just use a fixed or passed one.
		// Actually config doesn't have JobID. We can just use "default" if empty.
		if jobID == "" {
			jobID = "default"
		}
	}
	mState, err := LoadState(jobID)
	if err != nil {
		slog.Warn("Failed to load state, starting fresh", "error", err)
		mState = NewMigrationState(jobID)
	}

	if len(cfg.Tables) == 0 && cfg.ResumeJobID != "" {
		for tableName := range mState.Tables {
			cfg.Tables = append(cfg.Tables, tableName)
		}
		slog.Info("Loaded tables from state for resume", "tables", cfg.Tables)
	}

	if cfg.DryRun {
		slog.Info("Dry run mode enabled. Verifying connectivity and estimating row counts.")
		dryTracker, hasDryTracker := tracker.(DryRunTracker)

		connOk := true
		if cfg.TargetURL != "" {
			connOk = tryConnectTarget(dia, cfg.TargetURL)
			if !connOk {
				slog.Warn("Target DB connection failed during dry-run",
					"targetDB", dia.Name(), "url", cfg.TargetURL)
			} else {
				slog.Info("Target DB connection verified",
					"targetDB", dia.Name())
			}
		}

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
				dryTracker.DryRunResult(table, count, connOk)
			}
		}
		slog.Info("Dry run completed successfully.")
		return nil
	}

	var mainOut *os.File
	var mainBuf *bufio.Writer

	// If not direct migration, setup output file
	if pgPool == nil && targetDB == nil && !cfg.PerTable {
		outFile := cfg.OutFile
		if cfg.OutputDir != "" {
			outFile = cfg.OutputDir + "/" + outFile
		}
		// If resuming, append to the file
		flag := os.O_CREATE | os.O_WRONLY
		if cfg.ResumeJobID != "" {
			flag |= os.O_APPEND
		} else {
			flag |= os.O_TRUNC
		}
		mainOut, err = os.OpenFile(outFile, flag, 0644)
		if err != nil {
			return fmt.Errorf("error creating/opening output file: %v", err)
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
		go worker(w, dbConn, targetDB, pgPool, dia, jobs, &wg, mainBuf, cfg, &outMutex, tracker, mState)
	}

	// Send jobs
	for _, table := range cfg.Tables {
		jobs <- job{tableName: table}
	}
	close(jobs)

	wg.Wait()

	// Post-processing execution for Constraints
	if cfg.WithDDL && cfg.WithConstraints {
		slog.Info("Starting constraint post-processing")
		owner := resolveOwner(cfg)
		ddlTracker, hasDDLTracker := tracker.(DDLProgressTracker)

		for _, tableName := range cfg.Tables {
			constraints, err := GetConstraintMetadata(dbConn, tableName, owner)
			if err != nil {
				slog.Warn("failed to get constraint metadata", "table", tableName, "error", err)
				continue
			}

			for _, c := range constraints {
				ddl := GenerateConstraintDDL(c, cfg.Schema, dia)
				if ddl == "" || strings.HasPrefix(ddl, "--") {
					if ddl != "" && tracker != nil {
						if wt, ok := tracker.(WarningTracker); ok {
							wt.Warning(strings.TrimSpace(strings.TrimPrefix(ddl, "--")))
						}
					}
					continue
				}

				var execErr error
				if pgPool != nil {
					_, execErr = pgPool.Exec(context.Background(), ddl)
				} else if targetDB != nil {
					_, execErr = targetDB.Exec(ddl)
				} else {
					// File mode
					if cfg.PerTable {
						fileName := fmt.Sprintf("%s.sql", tableName)
						if cfg.OutputDir != "" {
							fileName = cfg.OutputDir + "/" + fileName
						}
						f, err := os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, 0644)
						if err == nil {
							f.WriteString(ddl)
							f.Close()
						}
					} else {
						writeToBuf(mainBuf, ddl, cfg.PerTable, &outMutex)
					}
				}

				if execErr != nil {
					slog.Warn("failed to execute constraint DDL", "constraint", c.Name, "error", execErr)
					if hasDDLTracker {
						ddlTracker.DDLProgress("constraint", c.Name, "error", execErr)
					}
				} else {
					slog.Info("constraint DDL applied", "constraint", c.Name)
					if hasDDLTracker {
						ddlTracker.DDLProgress("constraint", c.Name, "ok", nil)
					}
				}
			}
		}
	}

	return nil
}

func worker(id int, dbConn *sql.DB, targetDB *sql.DB, pgPool db.PGPool, dia dialect.Dialect, jobs <-chan job, wg *sync.WaitGroup, mainBuf *bufio.Writer, cfg *config.Config, outMutex *sync.Mutex, tracker ProgressTracker, mState *MigrationState) {
	defer wg.Done()
	for j := range jobs {
		slog.Info("worker processing table", "worker_id", id, "table", j.tableName)
		err := MigrateTable(dbConn, targetDB, pgPool, dia, j.tableName, mainBuf, cfg, outMutex, tracker, mState)
		if err != nil {
			slog.Error("error migrating table", "table", j.tableName, "error", err)
			if tracker != nil {
				tracker.Error(j.tableName, err)
			}
		}
	}
}

func MigrateTable(dbConn *sql.DB, targetDB *sql.DB, pgPool db.PGPool, dia dialect.Dialect, tableName string, mainBuf *bufio.Writer, cfg *config.Config, outMutex *sync.Mutex, tracker ProgressTracker, mState *MigrationState) error {
	tState := mState.GetState(tableName)
	if tState.Completed {
		slog.Info("table already completed, skipping", "table", tableName)
		if tracker != nil {
			// Fake init and done
			var totalRows int
			_ = dbConn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&totalRows)
			tracker.Init(tableName, totalRows)
			tracker.Update(tableName, totalRows)
			tracker.Done(tableName)
		}
		return nil
	}

	var totalRows int
	_ = dbConn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&totalRows)
	if tracker != nil {
		tracker.Init(tableName, totalRows)
	}

	var err error
	if pgPool != nil || targetDB != nil {
		err = MigrateTableDirect(dbConn, targetDB, pgPool, dia, tableName, cfg, tracker, mState)
	} else {
		err = MigrateTableToFile(dbConn, dia, tableName, mainBuf, cfg, outMutex, tracker, mState)
	}

	if err == nil {
		mState.MarkCompleted(tableName)
		if tracker != nil {
			tracker.Done(tableName)
		}
	}
	return err
}

func MigrateTableDirect(dbConn *sql.DB, targetDB *sql.DB, pgPool db.PGPool, dia dialect.Dialect, tableName string, cfg *config.Config, tracker ProgressTracker, mState *MigrationState) error {
	slog.Info("direct migration started", "table", tableName)

	ddlTracker, hasDDLTracker := tracker.(DDLProgressTracker)
	owner := resolveOwner(cfg)
	tState := mState.GetState(tableName)

	if cfg.WithDDL && tState.Offset == 0 {
		// Sequence DDLs — executed before CREATE TABLE
		if cfg.WithSequences {
			seqs, err := GetSequenceMetadata(dbConn, tableName, owner, splitNames(cfg.Sequences))
			if err != nil {
				slog.Warn("failed to get sequence metadata", "table", tableName, "error", err)
			} else {
				for _, seq := range seqs {
					ddl, supported := GenerateSequenceDDL(seq, cfg.Schema, dia)
					if !supported || ddl == "" {
						slog.Warn("Sequence not supported by dialect", "dialect", dia.Name(), "sequence", seq.Name)
						if wt, ok := tracker.(WarningTracker); ok {
							wt.Warning(fmt.Sprintf("%s은(는) Sequence를 지원하지 않습니다. --with-sequences 옵션은 무시됩니다.", dia.Name()))
						}
						continue
					}
					var err error
					if pgPool != nil {
						_, err = pgPool.Exec(context.Background(), ddl)
					} else if targetDB != nil {
						_, err = targetDB.Exec(ddl)
					}
					if err != nil {
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
			ddl := GenerateCreateTableDDL(tableName, cfg.Schema, colsMeta, dia)
			var err error
			if pgPool != nil {
				_, err = pgPool.Exec(context.Background(), ddl)
			} else if targetDB != nil {
				_, err = targetDB.Exec(ddl)
			}
			if err != nil {
				return fmt.Errorf("failed to execute DDL for %s: %v", tableName, err)
			}
			slog.Info("DDL executed successfully", "table", tableName)
		}
	} else if tState.Offset == 0 {
		// Validation check if not using DDL
		if pgPool != nil {
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
		} else if targetDB != nil {
			// For non-postgres, we can try to query 1 row to check existence
			// This is a naive check but works for most DBs
			qTableName := dia.QuoteIdentifier(strings.ToLower(tableName))
			if cfg.Schema != "" {
				qTableName = fmt.Sprintf("%s.%s", dia.QuoteIdentifier(strings.ToLower(cfg.Schema)), qTableName)
			}
			rows, err := targetDB.Query(fmt.Sprintf("SELECT 1 FROM %s WHERE 1=0", qTableName))
			if err != nil {
				return fmt.Errorf("target table %s does not exist or cannot be accessed. Use --with-ddl to create it automatically. err: %v", qTableName, err)
			}
			rows.Close()
		}
	}

	query := fmt.Sprintf("SELECT * FROM %s", tableName)
	if tState.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d ROWS", tState.Offset)
	}
	rows, err := dbConn.Query(query)
	if err != nil {
		return fmt.Errorf("failed to execute query on Oracle table %s: %v", tableName, err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	rowCount := tState.Offset

	if pgPool != nil {
		// Use COPY for high performance (Postgres only)
		ctx := context.Background()

		// Transaction per table
		tx, err := pgPool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to start pg transaction: %v", err)
		}
		defer tx.Rollback(ctx)

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

		rowCount += int(n)
		mState.UpdateOffset(tableName, rowCount)
	} else if targetDB != nil {
		// Non-postgres direct migration via batch INSERT
		tx, err := targetDB.Begin()
		if err != nil {
			return fmt.Errorf("failed to start target transaction: %v", err)
		}
		defer tx.Rollback()

		colTypes, err := rows.ColumnTypes()
		if err != nil {
			return err
		}

		values := make([]interface{}, len(cols))
		valuePtrs := make([]interface{}, len(cols))
		for i := range cols {
			valuePtrs[i] = &values[i]
		}

		var currentBatch [][]any

		for rows.Next() {
			err := rows.Scan(valuePtrs...)
			if err != nil {
				return fmt.Errorf("failed to scan row: %v", err)
			}

			// We need to copy values because valuePtrs points to the same underlying slice memory
			rowCopy := make([]any, len(cols))
			for i, v := range values {
				// Handle some basic driver specific parsing if needed, but mostly pass through
				if b, ok := v.([]byte); ok && strings.Contains(strings.ToUpper(colTypes[i].DatabaseTypeName()), "BLOB") {
					rowCopy[i] = append([]byte(nil), b...)
				} else if b, ok := v.([]byte); ok && strings.Contains(strings.ToUpper(colTypes[i].DatabaseTypeName()), "RAW") {
					rowCopy[i] = append([]byte(nil), b...)
				} else {
					rowCopy[i] = v
				}
			}

			currentBatch = append(currentBatch, rowCopy)
			rowCount++

			if len(currentBatch) >= cfg.BatchSize {
				stmts := dia.InsertStatement(tableName, cfg.Schema, cols, currentBatch, cfg.BatchSize)
				for _, stmt := range stmts {
					if _, err := tx.Exec(stmt); err != nil {
						return fmt.Errorf("failed to execute batch insert: %v\nstmt: %s", err, stmt)
					}
				}
				currentBatch = currentBatch[:0]
				mState.UpdateOffset(tableName, rowCount)
				if tracker != nil {
					tracker.Update(tableName, rowCount)
				}
			}
		}

		if len(currentBatch) > 0 {
			stmts := dia.InsertStatement(tableName, cfg.Schema, cols, currentBatch, cfg.BatchSize)
			for _, stmt := range stmts {
				if _, err := tx.Exec(stmt); err != nil {
					return fmt.Errorf("failed to execute batch insert: %v\nstmt: %s", err, stmt)
				}
			}
			mState.UpdateOffset(tableName, rowCount)
			if tracker != nil {
				tracker.Update(tableName, rowCount)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction for %s: %v", tableName, err)
		}
	}

	// Index DDLs — executed after COPY/INSERT for performance
	if cfg.WithDDL && cfg.WithIndexes && tState.Offset == 0 {
		indexes, err := GetIndexMetadata(dbConn, tableName, owner)
		if err != nil {
			slog.Warn("failed to get index metadata", "table", tableName, "error", err)
		} else {
			for _, idx := range indexes {
				ddl := GenerateIndexDDL(idx, tableName, cfg.Schema, dia)
				if ddl == "" {
					continue
				}
				var err error
				if pgPool != nil {
					_, err = pgPool.Exec(context.Background(), ddl)
				} else if targetDB != nil {
					_, err = targetDB.Exec(ddl)
				}
				if err != nil {
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

func MigrateTableToFile(dbConn *sql.DB, dia dialect.Dialect, tableName string, mainBuf *bufio.Writer, cfg *config.Config, outMutex *sync.Mutex, tracker ProgressTracker, mState *MigrationState) error {
	slog.Info("file-based migration started", "table", tableName)

	tState := mState.GetState(tableName)

	var tableBuf *bufio.Writer
	if cfg.PerTable {
		fileName := fmt.Sprintf("%s.sql", tableName)
		if cfg.OutputDir != "" {
			fileName = cfg.OutputDir + "/" + fileName
		}

		flag := os.O_CREATE | os.O_WRONLY
		if cfg.ResumeJobID != "" {
			flag |= os.O_APPEND
		} else {
			flag |= os.O_TRUNC
		}

		f, err := os.OpenFile(fileName, flag, 0644)
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

	if cfg.WithDDL && tState.Offset == 0 {
		owner := resolveOwner(cfg)

		// Sequence DDLs — written before CREATE TABLE
		if cfg.WithSequences {
			seqs, err := GetSequenceMetadata(dbConn, tableName, owner, splitNames(cfg.Sequences))
			if err != nil {
				slog.Warn("failed to get sequence metadata", "table", tableName, "error", err)
			} else {
				for _, seq := range seqs {
					ddl, supported := GenerateSequenceDDL(seq, cfg.Schema, dia)
					if !supported || ddl == "" {
						slog.Warn("Sequence not supported by dialect", "dialect", dia.Name(), "sequence", seq.Name)
						if wt, ok := tracker.(WarningTracker); ok {
							wt.Warning(fmt.Sprintf("%s은(는) Sequence를 지원하지 않습니다. --with-sequences 옵션은 무시됩니다.", dia.Name()))
						}
						continue
					}
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
			ddl := GenerateCreateTableDDL(tableName, cfg.Schema, colsMeta, dia)
			writeToBuf(tableBuf, ddl+"\n", cfg.PerTable, outMutex)
		}

		// Index DDLs — written after CREATE TABLE, before INSERT
		if cfg.WithIndexes {
			indexes, err := GetIndexMetadata(dbConn, tableName, owner)
			if err != nil {
				slog.Warn("failed to get index metadata", "table", tableName, "error", err)
			} else {
				for _, idx := range indexes {
					ddl := GenerateIndexDDL(idx, tableName, cfg.Schema, dia)
					if ddl == "" {
						continue
					}
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
	if tState.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d ROWS", tState.Offset)
	}
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

	rowCount := tState.Offset
	var currentBatch [][]any

	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return fmt.Errorf("failed to scan row: %v", err)
		}

		// We need to copy values because valuePtrs points to the same underlying slice memory
		rowCopy := make([]any, len(cols))
		for i, v := range values {
			if b, ok := v.([]byte); ok && strings.Contains(strings.ToUpper(colTypes[i].DatabaseTypeName()), "BLOB") {
				rowCopy[i] = append([]byte(nil), b...)
			} else if b, ok := v.([]byte); ok && strings.Contains(strings.ToUpper(colTypes[i].DatabaseTypeName()), "RAW") {
				rowCopy[i] = append([]byte(nil), b...)
			} else {
				rowCopy[i] = v
			}
		}
		currentBatch = append(currentBatch, rowCopy)
		rowCount++

		if len(currentBatch) >= cfg.BatchSize {
			stmts := dia.InsertStatement(tableName, cfg.Schema, cols, currentBatch, cfg.BatchSize)
			for _, stmt := range stmts {
				writeToBuf(tableBuf, stmt, cfg.PerTable, outMutex)
			}
			currentBatch = currentBatch[:0]
			mState.UpdateOffset(tableName, rowCount)
			if tracker != nil {
				tracker.Update(tableName, rowCount)
			}
		}
	}

	if len(currentBatch) > 0 {
		stmts := dia.InsertStatement(tableName, cfg.Schema, cols, currentBatch, cfg.BatchSize)
		for _, stmt := range stmts {
			writeToBuf(tableBuf, stmt, cfg.PerTable, outMutex)
		}
		mState.UpdateOffset(tableName, rowCount)
		if tracker != nil {
			tracker.Update(tableName, rowCount)
		}
	}

	slog.Info("file-based migration finished", "table", tableName, "rows", rowCount)
	return nil
}

func writeToBuf(buf *bufio.Writer, s string, perTable bool, outMutex *sync.Mutex) {
	if !perTable {
		outMutex.Lock()
		defer outMutex.Unlock()
	}
	buf.WriteString(s)
}
