# Oracle to PostgreSQL Data Migration CLI (v2)

A high-performance Go-based CLI application designed to migrate data from Oracle databases to PostgreSQL.

## Features

- **Pure Go Drivers:** No Oracle Instant Client or CGO required.
- **Direct Migration (v2):** Stream data directly from Oracle to PostgreSQL using the high-performance `COPY` protocol.
- **Bulk SQL Generation:** Generate PostgreSQL-compatible `INSERT` SQL scripts as an alternative.
- **DDL Generation (v2):** Automatically generate and execute `CREATE TABLE` statements on the target PostgreSQL database based on Oracle metadata.
- **Worker Pool Parallelism (v2):** Configurable worker pool for efficient concurrent table processing.
- **Dry Run Mode (v2):** Verify connectivity and estimate data volumes without performing actual migration.
- **Structured Logging (v2):** JSON or Text-based structured logging using `log/slog`.
- **Data Type Mapping:** Handles VARCHAR2, CLOB, BLOB, RAW, DATE, TIMESTAMP, and NUMBER (with precision/scale).

## Installation

```bash
go build -o dbmigrator main.go
```

## Usage

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

### File-based Migration (Legacy)

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
| `-url` | Oracle DB Connection URL (host:port/service) | None | Yes |
| `-user` | Oracle Database Username | None | Yes |
| `-password` | Oracle Database Password | None | Yes |
| `-tables` | Comma-separated list of tables | None | Yes |
| `-pg-url` | PostgreSQL Connection URL (v2) | None | No |
| `-workers` | Number of concurrent workers (v2) | `4` | No |
| `-with-ddl` | Generate/Execute CREATE TABLE DDLs (v2) | `false` | No |
| `-dry-run` | connectivity check and estimation (v2) | `false` | No |
| `-log-json` | Enable JSON structured logging (v2) | `false` | No |
| `-out` | Output SQL file name (Legacy) | `migration.sql` | No |
| `-batch` | Rows per bulk insert statement (Legacy) | `1000` | No |
| `-schema` | PostgreSQL target schema name | None | No |
| `-per-table` | Output to separate files per table | `false` | No |
| `-parallel` | Process tables concurrently | `false` | No |

## Development

```bash
go test -v ./...
```
