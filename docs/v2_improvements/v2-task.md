# Implementation Tasks: Oracle to PostgreSQL Migration Tool v2

## Phase 1: Dependency Management & Infrastructure
- [x] Upgrade Go to 1.22+ in `go.mod`.
- [x] Install PostgreSQL driver: `github.com/jackc/pgx/v5`.
- [x] Implement `log/slog` for structured logging across the application.
- [x] Implement a configuration struct to handle new flags (`--pg-url`, `--workers`, `--with-ddl`, `--dry-run`).
- [x] Modularize codebase into internal packages.

## Phase 2: Worker Pool & Concurrency
- [x] Implement a `Dispatcher` (in `Run`) to manage a pool of `n` workers.
- [x] Implement a `Job` struct and a thread-safe worker mechanism.
- [x] Replace the simple `sync.WaitGroup` loop with the worker pool for processing tables.
- [x] Ensure proper graceful shutdown (all workers finish before Run returns).

## Phase 3: Direct Migration Implementation
- [x] Implement PostgreSQL connection pool management using `pgxpool`.
- [x] Implement a `DirectWriter` (in `MigrateTableDirect`) that uses `pgx.Conn.CopyFrom` for high-speed data transfer.
- [ ] Implement a fallback batch `INSERT` mechanism using parameterized queries for compatibility.
- [x] Add transaction support per table migration.

## Phase 4: Schema & DDL Generation
- [x] Implement Oracle metadata discovery (in `GetTableMetadata`) for precision, scale, and constraints.
- [x] Implement a mapping function (`MapOracleToPostgres`) from Oracle types to PostgreSQL types.
- [x] Implement `CREATE TABLE` script generation logic (`GenerateCreateTableDDL`).
- [x] Add the `--with-ddl` execution flow to run DDLs before data insertion.

## Phase 5: Dry Run & Validation
- [x] Implement `--dry-run` logic (in `Run`) to verify connectivity and report estimated row counts.
- [x] Implement validation to check if target tables exist before starting migration (in `MigrateTableDirect`).
- [x] Add pre-flight checks (connectivity verified during pool creation and dry-run).

## Phase 6: Testing & Quality Assurance
- [x] Update unit tests to use `slog` (implicitly via package refactoring).
- [x] Add new unit tests for the worker pool and job dispatching (`worker_test.go`).
- [x] Add integration tests using `pgx` and `sqlmock` to simulate both Oracle and PostgreSQL (`direct_test.go`).
- [ ] Perform performance benchmarking comparing file-based vs. direct migration.
- [x] Update documentation and examples in `README.md`.

