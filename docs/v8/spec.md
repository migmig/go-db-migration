# Technical Specification - v8 (Scalability & Enterprise Features)

## 1. Web UI Multi-Session Support
- **Session Management**: Introduce a `SessionManager` in `web/server.go` and `ws/tracker.go` to handle multiple `WebSocketTracker` instances, replacing the global singleton pattern.
- **Session ID**: Generate a UUID-based `SessionID` for each connecting client upon accessing the webpage.
- **API Integration**: Modify the `/api/migrate` endpoint to accept `SessionID` in the payload, ensuring the correct session-specific tracker is injected into `migration.Run()`.
- **Garbage Collection**: Implement cleanup logic to securely remove old or disconnected sessions from memory.
- **UI Tracking**: Display the current connected Session ID in the UI for debugging purposes.

## 2. Constraints Migration Support
### 2.1. Default Values
- Extract `DATA_DEFAULT` from Oracle's metadata in `GetTableMetadata` (`db/oracle.go`).
- Integrate the default values into the generated `CREATE TABLE` DDL statements.

### 2.2. Foreign Key & Check Constraints
- **Metadata Extraction**: Create a `GetConstraintMetadata` function to extract Foreign Key (FK) and CHECK constraints from `ALL_CONSTRAINTS`.
- **CLI/UI Integration**: Add `--with-constraints` flag (Web UI: "Include Constraints" checkbox).
- **Dialect DDL**: Define a `CreateConstraintDDL` interface and implement constraint syntax for each supported dialect.
- **Post-processing Execution**: To guarantee referential integrity order, execute `ALTER TABLE ADD CONSTRAINT` for FKs only as a final step after all table structures are created and all data is fully inserted/copied.

## 3. Connection Pool Fine-Tuning
- **Configuration Properties**: Add advanced connection pool parameters:
  - `--db-max-open` (Default: 0, unlimited)
  - `--db-max-idle` (Default: 2)
  - `--db-max-life` (Default: 0, unlimited)
- **Application**: Apply these settings directly to the Oracle source `sql.DB` instance, the Postgres target `pgxpool.Config`, and all other target `sql.DB` instances in `db/connect.go`.
- **Web UI**: Expose these settings under a new "DB Connection Pool Tuning" section within the "Advanced Settings".

## 4. Resumable Migration (Checkpoints)
- **State Management**: Create a new module `migration/state.go` to manage migration states.
- **Checkpointing**: Persist progress to the local filesystem at `.migration_state/{job_id}.json`, logging the completed offset or the maximum Primary Key value for each processed table.
- **Resume Command**: Introduce a `--resume {job_id}` CLI flag to resume interrupted migrations from their last successful checkpoint.
- **Chunking Queries**: Implement PK-based chunking/pagination queries for Oracle data extraction to correctly continue fetching data from where it left off, avoiding full re-fetches.
