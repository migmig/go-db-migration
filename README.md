# Oracle to PostgreSQL Data Migration CLI

A Go-based Command Line Interface (CLI) application designed to extract data from an Oracle database and generate PostgreSQL-compatible `INSERT` SQL scripts.

## Features

- **Pure Go Driver:** Uses `github.com/sijms/go-ora/v2`, requiring no Oracle Instant Client or CGO.
- **Bulk Inserts:** Generates `INSERT INTO ... VALUES (...)` statements with configurable batch sizes to optimize PostgreSQL import performance.
- **Data Type Mapping:** Correctly handles Oracle types including VARCHAR2, CLOB, BLOB, RAW, DATE, and TIMESTAMP, converting them to PostgreSQL-compatible literals.
- **Parallel Processing:** Supports concurrent extraction from multiple tables.
- **Flexible Output:** Option to write all tables to a single file or generate separate SQL files per table.
- **Thread-Safe Writing:** Uses Mutex synchronization for safe concurrent writes to a shared output file.
- **Buffered I/O:** Utilizes buffered writing for high-performance file output.

## Installation

```bash
go build -o dbmigrator main.go
```

## Usage

```bash
./dbmigrator -url "host:port/service_name" \
             -user "your_user" \
             -password "your_password" \
             -tables "TABLE1,TABLE2" \
             -out "migration.sql" \
             -batch 1000
```

### Flags

| Flag | Description | Default | Required |
| --- | --- | --- | --- |
| `-url` | Oracle DB Connection URL (host:port/service) | None | Yes |
| `-user` | Database Username | None | Yes |
| `-password` | Database Password | None | Yes |
| `-tables` | Comma-separated list of tables | None | Yes |
| `-out` | Output SQL file name (for single-file mode) | `migration.sql` | No |
| `-batch` | Number of rows per bulk insert statement | `1000` | No |
| `-schema` | PostgreSQL target schema name | None | No |
| `-per-table` | Output to separate files named `<TABLE>.sql` | `false` | No |
| `-parallel` | Process tables concurrently | `false` | No |

## Development

### Running Tests

```bash
go test -v .
```

The test suite includes:
- Unit tests for data type transformation logic.
- Integration tests using `sqlmock` to verify the full migration flow.
- Schema mapping and output formatting tests.
