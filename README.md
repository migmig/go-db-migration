# Oracle to Multi-Target Data Migration CLI (v11)

A high-performance Go-based CLI application designed to migrate data from Oracle databases to multiple target databases (PostgreSQL, MySQL, MariaDB, SQLite, MSSQL). Features an advanced Web UI for real-time monitoring, auto-healing, and chunk-based large table migration.

## Features

- **Pure Go Drivers:** No Oracle Instant Client or CGO required.
- **Multi-Target Database Support:** Migrate data directly to PostgreSQL, MySQL, MariaDB, SQLite, or MSSQL.
- **Advanced Web UI (v11):** Interactive web interface with WebSocket-based real-time progress tracking, dashboard monitoring, and topology viewing.
- **Table Chunking (v11):** Automatic partitioning of large tables for intra-table parallel migration, significantly boosting performance for massive datasets.
- **Auto-Healing (v11):** Smart auto-retry mechanisms for network timeouts and temporary failures to ensure seamless and resilient migration.
- **Direct Migration:** Stream data directly from Oracle to the target database. Uses high-performance `COPY` protocol for PostgreSQL.
- **Bulk SQL Generation:** Generate target-compatible `INSERT` SQL scripts as an alternative.
- **DDL Generation:** Automatically generate and execute `CREATE TABLE` statements on the target database based on Oracle metadata.
- **Worker Pool Parallelism (v2):** Configurable worker pool for efficient concurrent table processing.
- **Dry Run Mode:** Verify connectivity and estimate data volumes without performing actual migration.
- **Migration History (v11):** Local SQLite-based persistence of migration sessions and logs for easy auditing and resuming.
- **Structured Logging:** JSON or Text-based structured logging using `log/slog`.
- **Data Type Mapping:** Handles VARCHAR2, CLOB, BLOB, RAW, DATE, TIMESTAMP, and NUMBER (with precision/scale).

## Installation

```bash
go build -o dbmigrator main.go
```

### Cross-Platform Build (OS별 빌드)

Go의 크로스 컴파일 기능을 사용하여 다른 OS용 실행 파일을 빌드할 수 있습니다:

**Linux:**
```bash
GOOS=linux GOARCH=amd64 go build -o dbmigrator-linux main.go
```

**Windows:**
```bash
GOOS=windows GOARCH=amd64 go build -o dbmigrator.exe main.go
```

**macOS (Apple Silicon):**
```bash
GOOS=darwin GOARCH=arm64 go build -o dbmigrator-mac main.go
```

## Usage

### Web UI Mode

Run the migrator in web mode to use the browser-based interface:

```bash
./dbmigrator -web
```
- Default URL: `http://localhost:8080`
- Features: Table discovery (LIKE search), real-time progress tracking, and ZIP download of generated SQL files.

![Web UI Screenshot](docs/web-ui.png)

### Direct Migration

**PostgreSQL:**
```bash
./dbmigrator -url "host:port/service" \
             -user "oracle_user" \
             -password "oracle_pass" \
             -pg-url "postgres://user:pass@host:port/dbname" \
             -tables "USERS,ORDERS" \
             -with-ddl \
             -parallel -workers 4
```

**MySQL/MariaDB:**
```bash
./dbmigrator -url "host:port/service" \
             -user "oracle_user" \
             -password "oracle_pass" \
             -target-db "mysql" \
             -target-url "user:pass@tcp(host:port)/dbname" \
             -tables "USERS,ORDERS" \
             -with-ddl \
             -parallel -workers 4
```

**SQLite:**
```bash
./dbmigrator -url "host:port/service" \
             -user "oracle_user" \
             -password "oracle_pass" \
             -target-db "sqlite" \
             -target-url "file:./mydb.sqlite3?_foreign_keys=on" \
             -tables "USERS,ORDERS" \
             -with-ddl
```

### File-based Migration

```bash
./dbmigrator -url "host:port/service" \
             -user "oracle_user" \
             -password "oracle_pass" \
             -tables "USERS,ORDERS" \
             -out "migration.sql" \
             -batch 1000
```

### Dry Run

```bash
./dbmigrator -url "host:port/service" -user "u" -password "p" -tables "T1" -dry-run
```

## Flags

| Flag | Description | Default | Required |
| --- | --- | --- | --- |
| `-web` | Run in Web UI mode | `false` | No |
| `-url` | Oracle DB Connection URL (host:port/service) | None | Yes* |
| `-user` | Oracle Database Username | None | Yes* |
| `-password` | Oracle Database Password | None | Yes* |
| `-tables` | Comma-separated list of tables | None | Yes* |
| `-pg-url` | PostgreSQL Connection URL (Legacy) | None | No |
| `-target-db` | Target DB Type (`postgres`, `mysql`, `mariadb`, `sqlite`, `mssql`) | `postgres` | No |
| `-target-url` | Target Database Connection URL | None | No |
| `-workers` | Number of concurrent workers | `4` | No |
| `-with-ddl` | Generate/Execute CREATE TABLE DDLs | `false` | No |
| `-dry-run` | connectivity check and estimation | `false` | No |
| `-log-json` | Enable JSON structured logging | `false` | No |
| `-out` | Output SQL file name | `migration.sql` | No |
| `-batch` | Rows per bulk insert statement | `1000` | No |
| `-schema` | PostgreSQL target schema name | None | No |
| `-per-table` | Output to separate files per table | `false` | No |
| `-parallel` | Process tables concurrently | `false` | No |
| `-chunk` | (v11) Enable auto-chunking for large tables | `false` | No |
| `-chunk-size`| (v11) Chunk size limit in rows for partitioning | `1000000` | No |
| `-auto-retry`| (v11) Enable automatic retry on failure | `true` | No |

\* *Required for CLI mode only. In Web mode, these are provided through the UI.*

## Development

```bash
go test -v ./...
```
