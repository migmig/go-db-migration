# v6 Implementation Tasks

## 1. Dialect Interface
- [ ] Create `internal/dialect/dialect.go` and define the `Dialect` interface.

## 2. Dialect Implementations
- [ ] Create `internal/dialect/postgres.go` for PostgreSQL (migrate existing logic).
- [ ] Create `internal/dialect/mysql.go` for MySQL.
- [ ] Create `internal/dialect/mariadb.go` for MariaDB.
- [ ] Create `internal/dialect/sqlite.go` for SQLite.
- [ ] Create `internal/dialect/mssql.go` for MSSQL.

## 3. Config Expansion
- [ ] Update `internal/config/config.go` to add `--target-db` and `--target-url` flags.
- [ ] Ensure backward compatibility for `--pg-url`.

## 4. DDL Refactoring
- [ ] Update `internal/migration/ddl.go` to use `Dialect` interface for DDL generation (CreateTable, Sequence, Index).

## 5. Migration Logic Refactoring
- [ ] Update `internal/migration/migration.go` to inject `Dialect`.
- [ ] Update `MigrateTableDirect` to use dynamic driver initialization based on `Dialect`.
- [ ] Update `ProcessRow` and `WriteBatch` to use `Dialect.InsertStatement`.

## 6. Web UI & API
- [ ] Update `internal/web/server.go` to handle `targetDb` and `targetUrl`.
- [ ] Update `internal/web/templates/index.html` to add Target DB dropdown and dynamic UI changes.

## 7. Testing
- [ ] Write unit tests for `Dialect` implementations (`internal/dialect/*`).
- [ ] Update integration tests for new targets.
