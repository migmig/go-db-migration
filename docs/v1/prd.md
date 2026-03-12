# Product Requirements Document (PRD): Oracle to PostgreSQL Data Migration CLI

## 1. Overview
This project is a Command Line Interface (CLI) application built in Go, designed to extract data from an Oracle database and generate PostgreSQL-compatible `INSERT` SQL scripts. The tool emphasizes performance and flexibility, offering bulk insert capabilities, correct data type mapping (e.g., CLOB, BLOB, Timestamp), and advanced execution options like parallel processing and per-table output file generation.

## 2. Objectives
- **Data Export:** Query specified tables from an Oracle database.
- **Data Transformation:** Convert Oracle-specific data types (BLOB, RAW, CLOB, DATE, TIMESTAMP) to appropriate PostgreSQL literal syntax (e.g., `\x...` hex encoding for bytea, properly escaped strings, precise timestamp strings).
- **Efficiency:** Generate `BULK INSERT` statements rather than individual row inserts to reduce PostgreSQL import overhead.
- **Parallelism & Organization:** Provide options to process tables concurrently and output results into separate `.sql` files per table.

## 3. Scope and Features
### Core Features
- Connect to an Oracle database using a pure Go driver (no CGO/Oracle Instant Client required).
- Accept connection credentials, URL, and a comma-separated list of tables via CLI flags.
- Handle `NULL` values seamlessly.
- Extract `SELECT * FROM <table>` iteratively to avoid excessive memory consumption.
- Batch rows up to a configurable size (default 1000) before writing an `INSERT INTO` statement.

### Advanced Output Options (New)
- **`--per-table` Flag:**
  - When enabled, instead of writing all SQL commands to a single output file (e.g., `migration.sql`), the tool will generate a separate file for each table in the format: `<table>_migration.sql`.
- **`--parallel` Flag:**
  - Enables concurrent extraction and file writing for multiple tables using Go routines (`sync.WaitGroup`).
  - Greatly speeds up the extraction of large schemas when multiple tables are specified.
  - If both `--parallel` and a single output file (no `--per-table`) are used, the tool must safely synchronize file writes using a Mutex to prevent data interleaving/corruption.

## 4. CLI Arguments
| Flag | Description | Default | Required |
| --- | --- | --- | --- |
| `-url` | Oracle DB Connection URL / DSN | None | Yes |
| `-user` | Database Username | None | Yes |
| `-password`| Database Password | None | Yes |
| `-tables` | Comma-separated list of tables | None | Yes |
| `-out` | Output SQL file name (if `-per-table` is false) | `migration.sql` | No |
| `-batch` | Number of rows per bulk insert statement | `1000` | No |
| `-per-table`| Output to separate files named `<tablename>.sql` | `false` | No |
| `-parallel`| Process tables concurrently | `false` | No |

## 5. Non-Functional Requirements
- **Language:** Go 1.21+
- **Driver:** `github.com/sijms/go-ora/v2`
- **Performance:** Stream results efficiently; keep memory usage low even for large tables. Use Goroutines for parallel operations.
- **Safety:** Prevent SQL injection risks in the generated file by strictly escaping single quotes in string columns. Write concurrently to a shared file using a Mutex when necessary.

## 6. Future Enhancements
- DDL Generation (CREATE TABLE schemas).
- Direct PostgreSQL insertion without intermediate files.
- Support for more complex data types like Spatial data.
