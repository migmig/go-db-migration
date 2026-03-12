# Implementation Tasks: Oracle to PostgreSQL Data Migration CLI

## Phase 1: Project Setup & CLI Skeleton
- [x] Initialize Go module (`go mod init`).
- [x] Install Oracle driver dependency: `github.com/sijms/go-ora/v2`.
- [x] Implement CLI flag parsing (URL, User, Password, Tables, Out, Batch, PerTable, Parallel).
- [x] Validate required flags and provide basic help/usage output.

## Phase 2: Database Connectivity & Metadata
- [x] Implement Oracle connection logic using `sql.Open`.
- [x] Implement a function to fetch column names and types for a given table name.
- [x] Create a robust connection test to verify credentials and reachability.

## Phase 3: Core Data Extraction & Transformation
- [x] Implement streaming row extraction using `sql.Rows.Next()`.
- [x] Implement data type mapping logic:
    - [x] String escaping (VARCHAR2, CLOB).
    - [x] Numeric formatting (NUMBER).
    - [x] Timestamp formatting (DATE, TIMESTAMP).
    - [x] Hex encoding for binary data (BLOB, RAW).
    - [x] NULL handling.
- [x] Implement basic batching logic to group rows into `INSERT INTO` statements.

## Phase 4: Output Management
- [x] Implement single-file output writer (default mode).
- [x] Implement `--per-table` logic: generate individual files per table.
- [x] Implement buffered writing using `bufio` to improve performance.

## Phase 5: Concurrency & Parallelism
- [x] Implement `--parallel` processing using `sync.WaitGroup`.
- [x] Implement thread-safe writing for single-file mode using `sync.Mutex`.
- [x] Ensure proper error handling and propagation from goroutines back to the main process.

## Phase 6: Validation & Refinement
- [x] Add unit tests for type transformation logic.
- [x] Add integration tests (mocking or using a test Oracle instance if available).
- [x] Perform manual verification of generated SQL against a PostgreSQL target.
- [x] Documentation update and final code cleanup.
