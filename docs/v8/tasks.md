# Implementation Tasks - v8 (Scalability & Enterprise Features)

## 1. Web UI Multi-Session Support
- [x] **Session Manager Implementation**
  - [x] Create `SessionManager` struct in `ws/tracker.go` to handle multiple `WebSocketTracker` instances.
  - [x] Remove global `tracker` singleton.
- [x] **Session ID Generation & Handling**
  - [x] Generate UUID-based `SessionID` for connecting clients in `web/server.go`.
  - [x] Update `/api/migrate` endpoint to accept and parse `SessionID` from the request payload.
  - [x] Inject the correct session-specific tracker into `migration.Run()`.
- [x] **Session Cleanup (Garbage Collection)**
  - [x] Implement logic to securely close and remove disconnected or stale sessions from memory.
- [x] **UI Updates**
  - [x] Display the connected `SessionID` on the Web UI (`index.html`) for debugging and tracking.

## 2. Constraints Migration Support
- [x] **Default Values**
  - [x] Update `GetTableMetadata` in `db/oracle.go` to extract `DATA_DEFAULT`.
  - [x] Update `CREATE TABLE` DDL generation to include extracted default values.
- [x] **Foreign Key & Check Constraints**
  - [x] Implement `GetConstraintMetadata` function in `db/oracle.go` to extract FK and CHECK constraints.
  - [x] Add `--with-constraints` flag to the CLI configuration.
  - [x] Add "Include Constraints" checkbox to the Web UI (`index.html`).
  - [x] Define `CreateConstraintDDL` interface in the dialect package.
  - [x] Implement `CreateConstraintDDL` for all supported dialects (PostgreSQL, MySQL, MariaDB, SQLite, MSSQL).
- [x] **Post-processing Execution**
  - [x] Modify `migration/migration.go` to defer FK constraints execution.
  - [x] Implement the final step to execute `ALTER TABLE ADD CONSTRAINT` for all extracted FKs after data insertion is fully completed.

## 3. Connection Pool Fine-Tuning
- [x] **Configuration Updates**
  - [x] Add `--db-max-open`, `--db-max-idle`, and `--db-max-life` parameters to CLI flags and internal Config struct.
- [x] **Connection Application**
  - [x] Apply pool settings to the Oracle source `sql.DB` instance.
  - [x] Apply pool settings to the Postgres target `pgxpool.Config`.
  - [x] Apply pool settings to all other target `sql.DB` instances in `db/connect.go`.
- [x] **UI Updates**
  - [x] Add "DB Connection Pool Tuning" inputs within the "Advanced Settings" section of `index.html`.

## 4. Resumable Migration (Checkpoints)
- [x] **State Management**
  - [x] Create `migration/state.go` module for state handling.
  - [x] Implement functionality to save/persist progress (offset or max PK value) to `.migration_state/{job_id}.json`.
- [x] **Resume Functionality**
  - [x] Add `--resume {job_id}` CLI flag and integrate it into the initial startup logic.
- [x] **Data Extraction Refactoring**
  - [x] Implement PK-based chunking/pagination queries for Oracle data extraction.
  - [x] Ensure extraction can dynamically start from the last recorded checkpoint without re-fetching existing data.
