# v6 구현 작업

## 1. 방언(Dialect) 인터페이스
- [ ] `internal/dialect/dialect.go`를 생성하고 `Dialect` 인터페이스를 정의.

## 2. 방언 구현
- [ ] PostgreSQL용 `internal/dialect/postgres.go` 생성 (기존 로직 마이그레이션).
- [ ] MySQL용 `internal/dialect/mysql.go` 생성.
- [ ] MariaDB용 `internal/dialect/mariadb.go` 생성.
- [ ] SQLite용 `internal/dialect/sqlite.go` 생성.
- [ ] MSSQL용 `internal/dialect/mssql.go` 생성.

## 3. 구성(Config) 확장
- [ ] `--target-db` 및 `--target-url` 플래그를 추가하기 위해 `internal/config/config.go` 업데이트.
- [ ] `--pg-url`에 대한 하위 호환성 보장.

## 4. DDL 리팩토링
- [ ] DDL 생성(CreateTable, Sequence, Index)에 `Dialect` 인터페이스를 사용하도록 `internal/migration/ddl.go` 업데이트.

## 5. 마이그레이션 로직 리팩토링
- [ ] `Dialect`를 주입하기 위해 `internal/migration/migration.go` 업데이트.
- [ ] `Dialect`를 기반으로 동적 드라이버 초기화를 사용하도록 `MigrateTableDirect` 업데이트.
- [ ] `Dialect.InsertStatement`를 사용하도록 `ProcessRow` 및 `WriteBatch` 업데이트.

## 6. 웹 UI 및 API
- [ ] `targetDb` 및 `targetUrl`을 처리하기 위해 `internal/web/server.go` 업데이트.
- [ ] 대상 DB 드롭다운 및 동적 UI 변경 사항을 추가하기 위해 `internal/web/templates/index.html` 업데이트.

## 7. 테스트
- [ ] `Dialect` 구현체(`internal/dialect/*`)에 대한 단위 테스트 작성.
- [ ] 새로운 대상에 대한 통합 테스트 업데이트.
