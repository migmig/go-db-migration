# Implementation Tasks: Oracle to PostgreSQL Migration Tool v2

## Phase 1: Dependency Management & Infrastructure
- [ ] Upgrade Go to 1.22+ in `go.mod`.
- [ ] Install PostgreSQL driver: `github.com/jackc/pgx/v5`.
- [ ] Implement `log/slog` for structured logging across the application.
- [ ] Implement a configuration struct to handle new flags (`--pg-url`, `--workers`, `--with-ddl`, `--dry-run`).

## Phase 2: Worker Pool & Concurrency
- [ ] Implement a `Dispatcher` to manage a pool of `n` workers.
- [ ] Implement a `Job` struct and a thread-safe worker mechanism.
- [ ] Replace the simple `sync.WaitGroup` loop with the worker pool for processing tables.
- [ ] Ensure proper graceful shutdown and signal handling.

## Phase 3: Direct Migration Implementation
- [ ] Implement PostgreSQL connection pool management using `pgxpool`.
- [ ] Implement a `DirectWriter` that uses `pgx.Conn.CopyFrom` for high-speed data transfer.
- [ ] Implement a fallback batch `INSERT` mechanism using parameterized queries for compatibility.
- [ ] Add transaction support per table migration.

## Phase 4: Schema & DDL Generation
- [ ] Implement Oracle metadata discovery for precision, scale, and constraints.
- [ ] Implement a mapping function from Oracle types to PostgreSQL types.
- [ ] Implement `CREATE TABLE` script generation logic.
- [ ] Add the `--with-ddl` execution flow to run DDLs before data insertion.

## Phase 5: Dry Run & Validation
- [ ] Implement `--dry-run` logic to verify connectivity and report estimated row counts.
- [ ] Implement validation to check if target tables exist before starting migration.
- [ ] Add pre-flight checks for permissions on both Oracle and PostgreSQL.

## Phase 6: Testing & Quality Assurance
- [ ] Update unit tests to use `slog`.
- [ ] Add new unit tests for the worker pool and job dispatching.
- [ ] Add integration tests using `pgx` and `sqlmock` to simulate both Oracle and PostgreSQL.
- [ ] Perform performance benchmarking comparing file-based vs. direct migration.
- [ ] Update documentation and examples in `README.md`.
