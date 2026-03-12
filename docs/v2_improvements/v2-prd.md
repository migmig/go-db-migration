# Product Requirements Document (PRD): Oracle to PostgreSQL Migration Tool v2

## 1. Overview
Building on the success of the initial CLI tool, v2 aims to enhance robustness, performance, and features to support production-scale migrations. This version focuses on direct database insertion, schema creation, and better resource management.

## 2. Objectives
- **Direct Migration:** Support direct data transfer from Oracle to PostgreSQL without intermediate SQL files.
- **Schema Autodiscovery:** Generate `CREATE TABLE` DDLs based on Oracle table structures.
- **Improved Performance:** Use a worker pool to manage concurrent table processing more efficiently.
- **Enhanced Validation:** Provide a "dry run" mode to validate connectivity and estimate data volumes.

## 3. New Features

### 3.1 Direct PostgreSQL Insertion
- **New Flag:** `--pg-url`
- When specified, the tool connects to the target PostgreSQL database and executes `COPY` or `INSERT` commands directly.
- Uses `github.com/lib/pq` or `github.com/jackc/pgx/v5`.

### 3.2 DDL Generation
- **New Flag:** `--with-ddl`
- Generates `CREATE TABLE` statements before the `INSERT` statements.
- Automatically maps Oracle types to the most appropriate PostgreSQL types (e.g., `VARCHAR2` -> `text`, `NUMBER(10,0)` -> `integer`).

### 3.3 Worker Pool for Parallelism
- **New Flag:** `--workers` (Default: 4)
- Instead of spawning one goroutine per table, the tool uses a fixed number of workers to process the table queue, preventing resource exhaustion on the databases.

### 3.4 Dry Run Mode
- **New Flag:** `--dry-run`
- Connects to Oracle, verifies permissions, counts rows in specified tables, and reports the estimated migration plan without writing any files or inserting data.

### 3.5 Structured Logging
- Replace standard `log` with `log/slog` for structured, leveled logging (JSON or Text).

## 4. Technical Requirements
- **Language:** Go 1.22+
- **Drivers:** 
  - Oracle: `github.com/sijms/go-ora/v2`
  - PostgreSQL: `github.com/jackc/pgx/v5`
- **Concurrency:** Worker pool pattern using channels and `sync.WaitGroup`.

## 5. Scope & Constraints
- Migration remains focused on data and basic schema; complex objects like procedures, triggers, or views are out of scope for v2.
- Large BLOB/CLOB handling must be optimized to avoid memory spikes (streaming).
