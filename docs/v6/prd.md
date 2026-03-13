# PRD (Product Requirements Document) - 출력 대상 DB 선택 지원 (v6)

## 1. 개요 (Overview)

현재 `dbmigrator`는 Oracle → **PostgreSQL** 고정 경로만 지원합니다.
이 버전에서는 `--target-db` 옵션을 추가하여 출력 대상 DB를 **PostgreSQL·MySQL·MariaDB·SQLite·MSSQL** 중 선택할 수 있도록 합니다.
SQL 파일 출력 모드와 Direct 마이그레이션 모드 모두 선택된 DB의 방언(dialect)에 맞게 DDL·DML을 생성합니다.

---

## 2. 배경 (Background)

### 2.1. 현재 한계

| 출력 대상 | 현재 지원 | 비고 |
|-----------|:---------:|------|
| PostgreSQL | O | 기본 출력 대상 |
| MySQL / MariaDB | X | 문법·타입 차이로 그대로 사용 불가 |
| SQLite | X | 로컬 테스트 및 임베디드 용도 |
| MSSQL (SQL Server) | X | 기업 환경 이전 수요 |

### 2.2. 방언별 주요 차이

| 항목 | PostgreSQL | MySQL/MariaDB | SQLite | MSSQL |
|------|-----------|---------------|--------|-------|
| 자동 증가 | `SERIAL` / `BIGSERIAL` / `nextval()` | `AUTO_INCREMENT` | `AUTOINCREMENT` | `IDENTITY(1,1)` |
| 문자열 타입 | `TEXT`, `VARCHAR(n)` | `TEXT`, `VARCHAR(n)` | `TEXT` | `NVARCHAR(n)`, `NTEXT` |
| 날짜/시간 | `TIMESTAMP` | `DATETIME` | `TEXT` (ISO 8601) | `DATETIME2` |
| CLOB/BLOB | `TEXT`, `BYTEA` | `LONGTEXT`, `LONGBLOB` | `TEXT`, `BLOB` | `NVARCHAR(MAX)`, `VARBINARY(MAX)` |
| FLOAT | `NUMERIC(p,s)` | `DECIMAL(p,s)` | `REAL` | `DECIMAL(p,s)` |
| DDL 가드 | `IF NOT EXISTS` | `IF NOT EXISTS` | `IF NOT EXISTS` | 별도 존재 체크 필요 |
| 식별자 인용 | `"name"` | `` `name` `` | `"name"` | `[name]` |
| INSERT 다중 행 | `VALUES (),()`  | `VALUES (),()`  | `VALUES (),()` | `VALUES (),()` (2008+) |
| Sequence | 네이티브 지원 | 미지원 (`AUTO_INCREMENT`로 대체) | 미지원 | `SEQUENCE` (2012+) |
| Index 문법 | `CREATE INDEX IF NOT EXISTS` | `CREATE INDEX` (IF NOT EXISTS MySQL 8+) | `CREATE INDEX IF NOT EXISTS` | `CREATE INDEX` |

---

## 3. 목표 (Goals)

- `--target-db` 하나의 플래그로 출력 방언 전환이 가능하도록 합니다.
- Oracle 타입 → 대상 DB 타입 매핑 테이블을 방언별로 분리·관리합니다.
- DDL(CREATE TABLE, Sequence, Index), DML(INSERT) 모두 대상 DB에 맞는 문법으로 생성합니다.
- Direct 마이그레이션 모드에서 대상 DB에 맞는 드라이버로 연결합니다.
- Web UI의 연결 설정이 선택된 대상 DB에 맞게 동적으로 변경됩니다.
- PostgreSQL 기본값 유지 — 기존 사용자에게 아무 변화 없음.

---

## 4. 기능 요구사항 (Functional Requirements)

### 4.1. `--target-db` 플래그

| 값 | 설명 | Direct 드라이버 | 기본 포트 |
|----|------|----------------|---------|
| `postgres` | PostgreSQL (기본값) | `pgx/v5` | 5432 |
| `mysql` | MySQL 8.x | `go-sql-driver/mysql` | 3306 |
| `mariadb` | MariaDB 10.x+ | `go-sql-driver/mysql` | 3306 |
| `sqlite` | SQLite3 (파일 경로) | `mattn/go-sqlite3` | N/A |
| `mssql` | SQL Server 2019+ | `microsoft/go-mssqldb` | 1433 |

- 미지정 시 `postgres`로 동작 (완전 하위 호환)
- SQL 파일 출력 모드에서도 방언에 맞는 DDL·DML을 생성

### 4.2. Oracle 타입 매핑

방언별 타입 변환 테이블. `TypeMapper` 인터페이스를 구현하는 방언별 struct로 분리합니다.

| Oracle 타입 | PostgreSQL | MySQL/MariaDB | SQLite | MSSQL |
|------------|-----------|---------------|--------|-------|
| `NUMBER(p,0)` p≤4 | `SMALLINT` | `SMALLINT` | `INTEGER` | `SMALLINT` |
| `NUMBER(p,0)` p≤9 | `INTEGER` | `INT` | `INTEGER` | `INT` |
| `NUMBER(p,0)` p≤18 | `BIGINT` | `BIGINT` | `INTEGER` | `BIGINT` |
| `NUMBER(p,s)` s>0 | `NUMERIC(p,s)` | `DECIMAL(p,s)` | `REAL` | `DECIMAL(p,s)` |
| `VARCHAR2(n)` | `VARCHAR(n)` | `VARCHAR(n)` | `TEXT` | `NVARCHAR(n)` |
| `CHAR(n)` | `CHAR(n)` | `CHAR(n)` | `TEXT` | `NCHAR(n)` |
| `CLOB` | `TEXT` | `LONGTEXT` | `TEXT` | `NVARCHAR(MAX)` |
| `BLOB` | `BYTEA` | `LONGBLOB` | `BLOB` | `VARBINARY(MAX)` |
| `DATE` | `TIMESTAMP` | `DATETIME` | `TEXT` | `DATETIME2` |
| `TIMESTAMP(n)` | `TIMESTAMP(n)` | `DATETIME(n)` | `TEXT` | `DATETIME2(n)` |
| `FLOAT` | `DOUBLE PRECISION` | `DOUBLE` | `REAL` | `FLOAT` |

### 4.3. DDL 방언 차이 처리

#### 4.3.1. CREATE TABLE

```sql
-- PostgreSQL (기존)
CREATE TABLE IF NOT EXISTS "schema"."USERS" ( ... );

-- MySQL/MariaDB
CREATE TABLE IF NOT EXISTS `schema`.`USERS` ( ... );

-- SQLite
CREATE TABLE IF NOT EXISTS "USERS" ( ... );  -- 스키마 미지원

-- MSSQL
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'USERS')
CREATE TABLE [schema].[USERS] ( ... );
```

#### 4.3.2. Sequence DDL (`--with-sequences`)

| 대상 DB | 처리 방식 |
|---------|---------|
| PostgreSQL | `CREATE SEQUENCE IF NOT EXISTS ...` (기존) |
| MSSQL (2012+) | `CREATE SEQUENCE ... START WITH ... INCREMENT BY ...` |
| MySQL/MariaDB | Sequence 미지원 → 해당 컬럼을 `AUTO_INCREMENT`로 변환, 경고 로그 출력 |
| SQLite | Sequence 미지원 → `AUTOINCREMENT` 키워드로 대체, 경고 로그 출력 |

#### 4.3.3. Index DDL (`--with-indexes`)

```sql
-- PostgreSQL (기존)
CREATE INDEX IF NOT EXISTS idx_name ON "table" (col);

-- MySQL 5.7 이하 / MariaDB 10.4 이하
CREATE INDEX idx_name ON `table` (col);

-- MySQL 8.0+ / MariaDB 10.5+
CREATE INDEX IF NOT EXISTS idx_name ON `table` (col);

-- MSSQL
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_name')
    CREATE INDEX idx_name ON [schema].[table] (col);
```

#### 4.3.4. INSERT 문

- 기존 배치 INSERT(`VALUES (), ()`) 방식은 PostgreSQL·MySQL·SQLite·MSSQL(2008+) 모두 지원
- MSSQL: 최대 1000행 제한 → BatchSize 초과 시 자동으로 분할 출력

### 4.4. Direct 마이그레이션 연결

`--target-db` 값에 따라 대상 연결 URL 형식이 달라집니다.

| 대상 | URL 형식 | 플래그 |
|------|---------|--------|
| PostgreSQL | `postgres://user:pass@host:port/db` | `--pg-url` (기존 유지) |
| MySQL/MariaDB | `user:pass@tcp(host:port)/db` | `--target-url` (신규) |
| SQLite | `/path/to/file.db` | `--target-url` (신규) |
| MSSQL | `sqlserver://user:pass@host:port?database=db` | `--target-url` (신규) |

- 기존 `--pg-url`은 `--target-db postgres` 시 그대로 사용 가능 (하위 호환)
- `--target-db`가 postgres 이외이면서 `--pg-url`이 지정된 경우 경고 로그 출력

### 4.5. CLI 플래그 요약

| 플래그 | 타입 | 기본값 | 설명 |
|--------|------|--------|------|
| `--target-db` | string | `postgres` | 출력 대상 DB 종류 (`postgres`/`mysql`/`mariadb`/`sqlite`/`mssql`) |
| `--target-url` | string | `""` | 대상 DB 연결 URL (PostgreSQL 외 Direct 마이그레이션 시) |

#### CLI 조합 예시

```bash
# MySQL로 SQL 파일 출력 (DDL 포함)
dbmigrator -url localhost:1521/ORCL -user scott -password tiger \
  -tables USERS,ORDERS -with-ddl \
  --target-db mysql

# MariaDB 직접 마이그레이션
dbmigrator -url localhost:1521/ORCL -user scott -password tiger \
  -tables USERS \
  --target-db mariadb \
  --target-url "scott:tiger@tcp(localhost:3306)/mydb"

# SQLite 파일 생성 (테스트 환경)
dbmigrator -url localhost:1521/ORCL -user scott -password tiger \
  -tables USERS -with-ddl \
  --target-db sqlite

# MSSQL 직접 마이그레이션
dbmigrator -url localhost:1521/ORCL -user scott -password tiger \
  -tables USERS -with-ddl -with-indexes \
  --target-db mssql \
  --target-url "sqlserver://sa:pass@localhost:1433?database=mydb"
```

### 4.6. Web UI 확장

#### 4.6.1. 대상 DB 선택 드롭다운

고급 설정 섹션 상단에 대상 DB 선택 드롭다운을 추가합니다.

```
[고급 설정] ▾
┌──────────────────────────────────────────────────────┐
│  출력 대상 DB: [PostgreSQL ▾]          ← NEW          │
│                                                      │
│  Direct 마이그레이션                                   │
│  ☐ Direct Migration                                  │
│  대상 URL: [postgres://... or mysql://...]  ← 동적 변경│
│                                                      │
│  ...                                                 │
└──────────────────────────────────────────────────────┘
```

#### 4.6.2. 동적 UI 변화

| 선택 | `target-url` placeholder | `schema` 입력 | Sequence 체크박스 |
|------|--------------------------|-------------|-----------------|
| PostgreSQL | `postgres://user:pass@host:5432/db` | 표시 | 활성 |
| MySQL/MariaDB | `user:pass@tcp(host:3306)/db` | 표시 | 비활성 + 경고 툴팁 |
| SQLite | `/path/to/file.db` | 숨김 | 비활성 + 경고 툴팁 |
| MSSQL | `sqlserver://user:pass@host:1433?database=db` | 표시 | 활성 (2012+) |

- MySQL/MariaDB/SQLite 선택 시 `--with-sequences`가 체크되어 있으면 체크 해제 및 경고 표시
- SQLite 선택 시 `schema` 입력 필드 숨김 (SQLite는 스키마 미지원)

#### 4.6.3. API 요청 필드 추가

```json
{
  "targetDb": "mysql",
  "targetUrl": "user:pass@tcp(localhost:3306)/mydb"
}
```

#### 4.6.4. WebSocket 경고 메시지

```json
{ "type": "warning", "message": "MySQL은 Sequence를 지원하지 않습니다. AUTO_INCREMENT로 대체됩니다." }
```

---

## 5. 아키텍처 설계

### 5.1. Dialect 인터페이스

```go
// internal/dialect/dialect.go
type Dialect interface {
    Name() string
    QuoteIdentifier(name string) string
    MapOracleType(oracleType string, precision, scale int) string
    CreateTableDDL(tableName, schema string, cols []ColumnDef) string
    CreateSequenceDDL(seq SequenceMetadata, schema string) (string, bool) // (ddl, supported)
    CreateIndexDDL(idx IndexMetadata, tableName, schema string) string
    InsertStatement(tableName, schema string, cols []string, rows [][]any, batchSize int) []string
    DriverName() string
    NormalizeURL(url string) string
}
```

### 5.2. 방언별 구현체

```
internal/dialect/
├── dialect.go          # Dialect 인터페이스 + 공통 유틸
├── postgres.go         # PostgresDialect (기존 로직 이관)
├── mysql.go            # MySQLDialect
├── mariadb.go          # MariaDBDialect (MySQLDialect 임베딩 + 미세 차이)
├── sqlite.go           # SQLiteDialect
└── mssql.go            # MSSQLDialect
```

### 5.3. 기존 코드 영향

| 파일 | 변경 내용 |
|------|----------|
| `internal/config/config.go` | `TargetDB`, `TargetURL` 필드 및 플래그 추가 |
| `internal/dialect/` | 신규 패키지 (Dialect 인터페이스 + 5개 구현체) |
| `internal/migration/ddl.go` | `GenerateSequenceDDL`, `GenerateIndexDDL` → Dialect 위임 |
| `internal/migration/migration.go` | Dialect 선택 후 주입, Direct 모드 연결 URL 분기 |
| `internal/web/server.go` | `targetDb`, `targetUrl` 필드 추가, Config 매핑 |
| `internal/web/templates/index.html` | 대상 DB 드롭다운, URL placeholder 동적 변경 로직 |

---

## 6. 비기능 요구사항 (Non-Functional Requirements)

- **하위 호환성**: `--target-db` 미지정 시 완전히 기존 동작과 동일
- **Dialect 확장성**: 새 DB 추가 시 `Dialect` 인터페이스 구현체만 추가하면 됨
- **경고 투명성**: 방언 제약(Sequence 미지원 등) 발생 시 사용자에게 명확한 경고 출력
- **테스트 용이성**: Dialect 인터페이스로 단위 테스트에서 mock 교체 가능

---

## 7. 마일스톤 (Milestones)

1. **PRD 확정**: `docs/v6/prd.md` 작성 ✅
2. **Dialect 인터페이스 설계**: `internal/dialect/dialect.go` 정의
3. **PostgresDialect 구현**: 기존 로직 이관 (동작 동일 보장)
4. **MySQLDialect / MariaDBDialect 구현**: 타입 매핑 + DDL 생성
5. **SQLiteDialect 구현**: 타입 매핑 + DDL 생성
6. **MSSQLDialect 구현**: 타입 매핑 + DDL 생성 + MSSQL 배치 제한 처리
7. **Config 확장**: `--target-db`, `--target-url` 플래그 추가
8. **migration.go 연동**: Dialect 주입 및 Direct 모드 드라이버 분기
9. **Web API/UI 확장**: 드롭다운 + placeholder 동적 변경
10. **테스트**: 방언별 DDL·DML 출력 단위 테스트 + 통합 테스트
11. **최종 리뷰 및 문서 정리**

---

## 8. 향후 확장 고려사항 (Future Considerations)

- **CockroachDB** 지원: PostgreSQL 호환 방언으로 비교적 적은 변경
- **TiDB** 지원: MySQL 호환 방언
- **Oracle → Oracle** 지원: 동일 Oracle 버전 간 테이블 복제
- **소스 DB 다양화**: MySQL → PostgreSQL 등 Oracle 이외 소스 지원 (별도 PRD)
