# Technical Specification: Oracle to PostgreSQL Data Migration CLI

## 1. Introduction
This document defines the technical design and implementation details for a Go-based CLI tool that migrates data from Oracle databases to PostgreSQL-compatible SQL scripts.

## 2. Technical Stack
- **Language:** Go 1.21+
- **Oracle Driver:** `github.com/sijms/go-ora/v2` (Pure Go driver, no Oracle Instant Client required)
- **Concurrency:** Go standard library `sync` (WaitGroups, Mutex) and `channels`

## 3. Architecture & Design
The tool operates as a single-binary CLI. It follows a streaming architecture to ensure low memory footprint regardless of table size.

### 3.1 Data Flow
1. **Connection:** Establish a connection to the source Oracle DB using the provided DSN.
2. **Metadata Discovery:** For each table, query column names and types.
3. **Streaming Extraction:** Execute `SELECT * FROM <table>` and iterate over rows using `sql.Rows`.
4. **Transformation:** Convert Oracle types to PostgreSQL-compatible literals.
5. **Batching:** Accumulate rows into memory until the `batch` size is reached.
6. **Writing:** Format and write `INSERT INTO` statements to the target file(s).

## 4. Data Type Mapping
The tool must handle the following conversions:

| Oracle Type | PostgreSQL Type | Transformation Logic |
| --- | --- | --- |
| `VARCHAR2`, `CHAR`, `NVARCHAR2` | `text` / `varchar` | Escape single quotes (`'`) by doubling them (`''`). |
| `NUMBER` | `numeric` / `int` / `float` | Direct string representation. |
| `DATE`, `TIMESTAMP` | `timestamp` | Format as `YYYY-MM-DD HH24:MI:SS.FF`. |
| `CLOB` | `text` | Treat as large string, handle escaping. |
| `BLOB`, `RAW` | `bytea` | Convert to hex format `\x...`. |
| `NULL` | `NULL` | Explicitly write `NULL`. |

## 5. CLI Interface
The application will use the standard `flag` package or a library like `cobra`.

| Flag | Type | Description |
| --- | --- | --- |
| `-url` | string | Oracle DSN (e.g., `oracle://user:pass@host:port/service`) |
| `-user` | string | DB Username |
| `-password` | string | DB Password |
| `-tables` | string | Comma-separated list of tables (e.g., `USERS,ORDERS`) |
| `-out` | string | Output filename (Default: `migration.sql`) |
| `-batch` | int | Rows per `INSERT` statement (Default: `1000`) |
| `-per-table` | bool | Create separate files: `<TABLE>_migration.sql` |
| `-parallel` | bool | Process multiple tables concurrently |

## 6. Implementation Details

### 6.1 Parallel Processing (`--parallel`)
- Use a `sync.WaitGroup` to track completion of table processing routines.
- Limit concurrency if necessary (though the PRD doesn't specify a limit, a worker pool could be a future enhancement).

### 6.2 File Handling & Synchronization
- **Per-Table Mode:** Each goroutine opens and writes to its own file. No synchronization required between table workers.
- **Single-File Mode + Parallel:** A shared `io.Writer` or `*os.File` must be protected by a `sync.Mutex` to ensure that bulk insert blocks from different tables do not interleave.

### 6.3 Performance Optimizations
- **Buffered I/O:** Use `bufio.Writer` for all file operations to minimize system calls.
- **Memory Management:** Rows are processed one-by-one; only the current batch of rows is held in memory before flushing to disk.

## 7. Security & Safety
- **SQL Injection:** Since this tool generates scripts for manual execution, string values MUST be escaped. Oracle driver should return values as `interface{}`, which are then safely cast and formatted.
- **Credential Handling:** Password should be accepted via flag, but future versions should support environment variables for better security.
