# Oracle to PostgreSQL Data Migration CLI (v3)

A high-performance Go-based CLI application designed to migrate data from Oracle databases to PostgreSQL. Now features a Web UI for easier table selection and progress tracking.

## Features

- **Pure Go Drivers:** No Oracle Instant Client or CGO required.
- **Web UI Mode (v3):** Interactive web interface for table browsing, selection, and migration monitoring.
- **Direct Migration (v2):** Stream data directly from Oracle to PostgreSQL using the high-performance `COPY` protocol.
- **Bulk SQL Generation:** Generate PostgreSQL-compatible `INSERT` SQL scripts as an alternative.
- **DDL Generation (v2):** Automatically generate and execute `CREATE TABLE` statements on the target PostgreSQL database based on Oracle metadata.
- **Worker Pool Parallelism (v2):** Configurable worker pool for efficient concurrent table processing.
- **Dry Run Mode (v2):** Verify connectivity and estimate data volumes without performing actual migration.
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

### Web UI Mode (v3)

Run the migrator in web mode to use the browser-based interface:

```bash
./dbmigrator -web
```
- Default URL: `http://localhost:8080`
- Features: Table discovery (LIKE search), real-time progress tracking, and ZIP download of generated SQL files.

### Direct Migration (v2)

```bash
./dbmigrator -url "host:port/service" \
             -user "oracle_user" \
             -password "oracle_pass" \
             -pg-url "postgres://user:pass@host:port/dbname" \
             -tables "USERS,ORDERS" \
             -with-ddl \
             -parallel -workers 4
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
| `-web` | Run in Web UI mode (v3) | `false` | No |
| `-url` | Oracle DB Connection URL (host:port/service) | None | Yes* |
| `-user` | Oracle Database Username | None | Yes* |
| `-password` | Oracle Database Password | None | Yes* |
| `-tables` | Comma-separated list of tables | None | Yes* |
| `-pg-url` | PostgreSQL Connection URL (v2) | None | No |
| `-workers` | Number of concurrent workers | `4` | No |
| `-with-ddl` | Generate/Execute CREATE TABLE DDLs | `false` | No |
| `-dry-run` | connectivity check and estimation | `false` | No |
| `-log-json` | Enable JSON structured logging | `false` | No |
| `-out` | Output SQL file name | `migration.sql` | No |
| `-batch` | Rows per bulk insert statement | `1000` | No |
| `-schema` | PostgreSQL target schema name | None | No |
| `-per-table` | Output to separate files per table | `false` | No |
| `-parallel` | Process tables concurrently | `false` | No |

\* *Required for CLI mode only. In Web mode, these are provided through the UI.*

## Development

```bash
go test -v ./...
```
