# Technical Specification - v7 (Quality Improvement & Bug Fixes)

## 1. Type Mapping Precision Improvement
### 1.1. MySQL `MapOracleType`
- `VARCHAR2(n)`: Map to `VARCHAR(n)`. If `n > 16383`, map to `LONGTEXT`.
- `CHAR(n)`: Map to `CHAR(n)` (incorporating precision).

### 1.2. MSSQL `MapOracleType`
- `VARCHAR2(n)`: If `n <= 4000`, map to `NVARCHAR(n)`. If `n > 4000`, map to `NVARCHAR(MAX)`.
- `CHAR(n)`: Map to `NCHAR(n)` (incorporating precision, max 4000).
- `NUMBER` (no precision): Map to `NUMERIC`.

## 2. MSSQL DDL Condition Checks
### 2.1. `CreateTableDDL`
- Add `TABLE_SCHEMA` condition in the `IF NOT EXISTS` check to avoid false positives when a table with the same name exists in a different schema.
- Default to `dbo` if the schema is not specified.

### 2.2. `CreateIndexDDL`
- Use `sys.objects` join with `sys.indexes` on `object_id` to strictly filter by table name and schema, preventing duplicate index name conflicts across different tables.

## 3. Web UI Enhancements
### 3.1. DDL Options Visibility
- Extract DDL-related options (`--with-ddl`, `--with-sequences`, `--with-indexes`, `oracleOwner`) out from the "Direct Migration" toggle section.
- Display them in a common "DDL Settings" section under "Advanced Settings", making them applicable and visible for both File Output and Direct Migration modes.

### 3.2. Label & Title Fixes (PostgreSQL Dependency Removal)
- Update `server.go` HTML title from `"Oracle to PostgreSQL Migrator"` to `"Oracle DB Migrator"`.
- Update the Schema input label in `index.html` from `"PG Schema"` to `"Schema"`.

## 4. WebSocket Warning Message Implementation
- Add `MsgWarning MsgType = "warning"` to `ws/tracker.go`.
- Implement `Warning(message string)` method in `WebSocketTracker`.
- Extend `ProgressTracker` interface to `WarningTracker` interface.
- Emit a warning message during migration when a specific dialect does not support sequence DDL (e.g., MySQL).
- Update Web UI (`index.html`) `handleProgressMessage` to display a warning banner (yellow) at the top of the progress container when receiving a `warning` event.

## 5. Dry-Run Target DB Connection Validation
- In Dry-Run mode (`cfg.DryRun == true`), if `--target-url` is specified, attempt to connect to the target database.
- Send the connection success/failure result via `DryRunResult.ConnectionOk`.
- Reflect the target DB connection status in the Web UI.

## 6. Unit Tests Addition
- Create test files in `internal/dialect/`:
  - `mysql_test.go`
  - `mariadb_test.go`
  - `sqlite_test.go`
  - `mssql_test.go`
- Test cases must cover:
  1. `TestMapOracleType_*`
  2. `TestCreateTableDDL_*`
  3. `TestCreateIndexDDL_*`
  4. `TestInsertStatement_*`
