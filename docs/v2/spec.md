# Technical Specification: Oracle to PostgreSQL Migration Tool v2

## 1. Introduction
This specification defines the technical architecture and design for v2 of the migration tool, focusing on direct database migration, schema discovery, and optimized resource management.

## 2. Updated Technical Stack
- **Go 1.22+**
- **Oracle Driver:** `github.com/sijms/go-ora/v2`
- **PostgreSQL Driver:** `github.com/jackc/pgx/v5`
- **Logging:** `log/slog` (Structured Logging)

## 3. Architecture Overview

### 3.1 Migration Modes
1. **File-based (v1 Legacy):** Oracle -> Memory -> SQL File.
2. **Direct Migration (v2):** Oracle -> Memory -> PostgreSQL (`COPY` or Batch `INSERT`).

### 3.2 Component Design
- **Dispatcher:** Reads the list of tables, creates jobs, and manages a pool of workers.
- **Worker:** Consumes jobs (table names) from a channel, performs extraction from Oracle, and handles writing (either to file or PostgreSQL).
- **DDL Generator:** Queries Oracle metadata to construct compatible PostgreSQL `CREATE TABLE` statements.

## 4. Implementation Details

### 4.1 Worker Pool
- Implement a worker pool using a job channel and `sync.WaitGroup`.
- The `--workers` flag determines the number of concurrent table processors.
- Each worker maintains its own Oracle and (optional) PostgreSQL connection to avoid contention, or uses a thread-safe pool.

### 4.2 Direct PostgreSQL Migration
- Use `pgx.Conn` or `pgxpool.Pool`.
- **Preferred Method:** `COPY` command via `pgx.Conn.CopyFrom` for high-performance bulk loading.
- **Fallback:** Batch `INSERT` statements with parameterized queries.

### 4.3 DDL Mapping (Oracle to PostgreSQL)
| Oracle Type | PostgreSQL Type | Notes |
| --- | --- | --- |
| `NUMBER(*, 0)` | `bigint` / `integer` | Based on precision. |
| `NUMBER(*, >0)` | `numeric` | |
| `VARCHAR2(n)`, `NVARCHAR2(n)` | `text` or `varchar(n)` | |
| `DATE`, `TIMESTAMP` | `timestamp` | |
| `CLOB` | `text` | |
| `BLOB`, `RAW` | `bytea` | |

### 4.4 Dry Run Logic
- When `--dry-run` is active:
  - Establish connections to Oracle.
  - Query `SELECT COUNT(*) FROM table` for each table.
  - Report: "Table X: ~Y rows will be migrated."
  - Do NOT open any output files or execute insertions on the target.

### 4.5 Structured Logging (`slog`)
- Initialize a global logger with `slog.New(slog.NewJSONHandler(os.Stdout, nil))` if a JSON flag is set, otherwise use TextHandler.
- Log context: `slog.Info("processing table", "table", tableName, "status", "started")`.

## 5. Security & Safety
- **PostgreSQL DSN:** Handle via `--pg-url` flag or `PG_URL` environment variable.
- **Transaction Safety:** For direct insertion, consider wrapping each table migration in a transaction to ensure atomic results per table.
