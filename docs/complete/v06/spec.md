# Technical Specification (Spec) - Multi-Target DB Migration (v6)

## 1. 개요 (Overview)

본 문서는 `dbmigrator` v6의 핵심 기능인 **출력 대상 DB(Target DB) 다변화**에 대한 기술적 명세를 정의합니다.
기존 PostgreSQL 단일 지원 구조에서 벗어나, MySQL, MariaDB, SQLite, MSSQL 등 다양한 데이터베이스 방언(Dialect)을 지원하기 위한 아키텍처, 인터페이스 설계, 타입 매핑, DDL/DML 생성 전략, CLI 플래그 및 Web UI 확장에 대한 세부 사항을 다룹니다.

## 2. 아키텍처 설계 (Architecture Design)

다양한 대상 데이터베이스를 지원하기 위해 방언을 추상화하는 `Dialect` 인터페이스를 도입합니다.
`dbmigrator`의 핵심 로직(`migration`, `ddl` 등)은 특정 DB에 종속되지 않고 `Dialect` 인터페이스에 의존하여 유연성을 확보합니다.

### 2.1. Dialect 인터페이스 (`internal/dialect/dialect.go`)

각 대상 DB별로 고유한 동작(식별자 인용, 타입 매핑, DDL 생성, INSERT 문법 차이 등)을 캡슐화합니다.

```go
package dialect

// Dialect defines the interface for different target database dialects.
type Dialect interface {
	// Name returns the dialect name (e.g., "postgres", "mysql").
	Name() string

	// QuoteIdentifier quotes an identifier (e.g., table or column name) according to the dialect.
	QuoteIdentifier(name string) string

	// MapOracleType maps an Oracle data type to the target database type.
	MapOracleType(oracleType string, precision, scale int) string

	// CreateTableDDL generates the CREATE TABLE DDL.
	CreateTableDDL(tableName, schema string, cols []ColumnDef) string

	// CreateSequenceDDL generates the CREATE SEQUENCE DDL.
	// Returns a boolean indicating whether the target DB supports sequences.
	CreateSequenceDDL(seq SequenceMetadata, schema string) (string, bool)

	// CreateIndexDDL generates the CREATE INDEX DDL.
	CreateIndexDDL(idx IndexMetadata, tableName, schema string) string

	// InsertStatement generates batch INSERT statements.
	InsertStatement(tableName, schema string, cols []string, rows [][]any, batchSize int) []string

	// DriverName returns the Go SQL driver name (e.g., "pgx", "mysql", "sqlite3", "sqlserver").
	DriverName() string

	// NormalizeURL standardizes the connection URL for the target driver.
	NormalizeURL(url string) string
}
```

### 2.2. 방언별 구현체 (Dialect Implementations)

- **PostgreSQL (`internal/dialect/postgres.go`)**: 기존 동작(v1~v5)을 유지 (`pgx/v5` 드라이버 사용)
- **MySQL (`internal/dialect/mysql.go`)**: MySQL 8.x 대상 (`go-sql-driver/mysql` 드라이버 사용)
- **MariaDB (`internal/dialect/mariadb.go`)**: MySQL 구현체를 확장하거나 임베딩하여 미세 조정 (`go-sql-driver/mysql` 드라이버 사용)
- **SQLite (`internal/dialect/sqlite.go`)**: SQLite3 대상. 로컬 파일 시스템을 사용하며 스키마 개념 무시 (`mattn/go-sqlite3` 드라이버 사용)
- **MSSQL (`internal/dialect/mssql.go`)**: SQL Server 2019+ 대상. 배치 제한 등 고려 (`microsoft/go-mssqldb` 드라이버 사용)

## 3. 기능 상세 명세 (Functional Specifications)

### 3.1. CLI 플래그 및 Config 구조체 확장

`internal/config/config.go`에 다중 DB 지원을 위한 플래그를 추가합니다.

- `--target-db` (string, default: `"postgres"`): 대상 DB 선택. 유효값: `"postgres"`, `"mysql"`, `"mariadb"`, `"sqlite"`, `"mssql"`
- `--target-url` (string, default: `""`): PostgreSQL 외 DB로 Direct 마이그레이션 시 사용하는 URL 형식.

#### 하위 호환성 (Backward Compatibility)
- 기존 `--pg-url`은 `--target-db postgres` 일 때 `--target-url`처럼 처리됩니다.
- `--target-db`를 명시하지 않으면 기본적으로 `"postgres"`로 동작하여 기존 사용자 환경에 영향을 주지 않습니다.

### 3.2. Oracle 타입 매핑 (Type Mapping)

각 방언 구현체(`MapOracleType`)는 아래 테이블을 기준으로 매핑을 수행합니다.

| Oracle 타입 | PostgreSQL | MySQL/MariaDB | SQLite | MSSQL |
|------------|-----------|---------------|--------|-------|
| `NUMBER(p,0)` p≤4 | `SMALLINT` | `SMALLINT` | `INTEGER` | `SMALLINT` |
| `NUMBER(p,0)` p≤9 | `INTEGER` | `INT` | `INTEGER` | `INT` |
| `NUMBER(p,0)` p≤18| `BIGINT` | `BIGINT` | `INTEGER` | `BIGINT` |
| `NUMBER(p,s)` s>0 | `NUMERIC(p,s)` | `DECIMAL(p,s)` | `REAL` | `DECIMAL(p,s)` |
| `VARCHAR2(n)` | `VARCHAR(n)` | `VARCHAR(n)` | `TEXT` | `NVARCHAR(n)` |
| `CHAR(n)` | `CHAR(n)` | `CHAR(n)` | `TEXT` | `NCHAR(n)` |
| `CLOB` | `TEXT` | `LONGTEXT` | `TEXT` | `NVARCHAR(MAX)` |
| `BLOB` | `BYTEA` | `LONGBLOB` | `BLOB` | `VARBINARY(MAX)` |
| `DATE` | `TIMESTAMP` | `DATETIME` | `TEXT` | `DATETIME2` |
| `TIMESTAMP(n)`| `TIMESTAMP(n)` | `DATETIME(n)` | `TEXT` | `DATETIME2(n)` |
| `FLOAT` | `DOUBLE PRECISION`| `DOUBLE` | `REAL` | `FLOAT` |

### 3.3. DDL (CREATE TABLE, Sequence, Index) 생성 차이 처리

#### 3.3.1. 테이블 생성 (`CreateTableDDL`)
- **PostgreSQL**: `CREATE TABLE IF NOT EXISTS "schema"."TABLE" (...)`
- **MySQL/MariaDB**: `CREATE TABLE IF NOT EXISTS \`schema\`.\`TABLE\` (...)` (백틱 사용)
- **SQLite**: `CREATE TABLE IF NOT EXISTS "TABLE" (...)` (스키마 무시)
- **MSSQL**:
  ```sql
  IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'TABLE')
  CREATE TABLE [schema].[TABLE] (...)
  ```

#### 3.3.2. Sequence 생성 (`CreateSequenceDDL`)
- **PostgreSQL**: 지원. (기존 v5 유지)
- **MSSQL (2012+)**: 지원. `CREATE SEQUENCE [schema].[SEQ_NAME] START WITH ... INCREMENT BY ...`
- **MySQL / MariaDB / SQLite**:
  - `CreateSequenceDDL`은 빈 문자열과 `false`(미지원)를 반환.
  - 마이그레이션 엔진에서 해당 컬럼을 `AUTO_INCREMENT` (MySQL) 또는 `AUTOINCREMENT` (SQLite)로 자동 대체 처리.
  - 사용자에게 Sequence 미지원에 대한 경고(Warning)를 출력.

#### 3.3.3. 인덱스 생성 (`CreateIndexDDL`)
- **PostgreSQL**: `CREATE INDEX IF NOT EXISTS idx_name ON "schema"."table" (col)`
- **MySQL/MariaDB**: MySQL 8.0+에서는 `IF NOT EXISTS`를 지원하므로 활용. (이전 버전 호환성 필요 시 분기 고려)
- **SQLite**: `CREATE INDEX IF NOT EXISTS "idx_name" ON "table" ("col")`
- **MSSQL**:
  ```sql
  IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_name')
  CREATE INDEX idx_name ON [schema].[table] (col)
  ```

### 3.4. DML (INSERT) 생성 및 배치 처리 차이

- 다중 행 INSERT(`VALUES (), ()`)는 모든 타겟 DB에서 기본적으로 지원됨.
- **MSSQL 배치 제한 (`InsertStatement` 특징)**:
  - MSSQL은 단일 `INSERT` 문에서 `VALUES` 절 뒤에 올 수 있는 최대 행의 수가 1000행으로 제한됨.
  - `MSSQLDialect` 구현체의 `InsertStatement` 메서드는 전달된 `rows` 배열의 길이가 1000을 초과할 경우(예: `-batch 2000`), 자동으로 1000행 단위로 청크(chunk)를 분할하여 여러 개의 `INSERT` 문 배열을 반환해야 함.

### 3.5. Web UI 및 API 변경사항

#### 3.5.1. UI/UX 요구사항
- "고급 설정" 섹션 상단에 "출력 대상 DB" (Target DB) 드롭다운 (PostgreSQL, MySQL, MariaDB, SQLite, MSSQL) 추가.
- 선택된 대상에 따라 대상 URL 입력창의 PlaceHolder가 동적으로 변경됨.
  - PG: `postgres://user:pass@host:5432/db`
  - MySQL: `user:pass@tcp(host:3306)/db`
  - SQLite: `/path/to/file.db`
  - MSSQL: `sqlserver://user:pass@host:1433?database=db`
- SQLite 선택 시, `schema` 입력창 비활성화 또는 숨김 (스키마 미지원).
- MySQL/SQLite 선택 시, `--with-sequences`가 체크되어 있다면 체크 해제 후 비활성화, 툴팁으로 미지원 사유 안내.

#### 3.5.2. API 확장
마이그레이션 시작 요청 페이로드(`internal/web/server.go`)에 필드 추가:
```json
{
  ...
  "targetDb": "mysql",
  "targetUrl": "user:pass@tcp(localhost:3306)/mydb",
  ...
}
```

#### 3.5.3. WebSocket 메시지 확장
방언 제약사항에 따른 경고를 프론트엔드로 전달하기 위해 `warning` 타입의 메시지를 추가 활용.

## 4. 모듈 간 의존성 변경 요약 (Scope of Changes)

1. **`internal/dialect/*`**: 신규 패키지 및 5개의 방언 구현체 추가.
2. **`internal/config/config.go`**: `--target-db`, `--target-url` 처리.
3. **`internal/migration/ddl.go`**: DDL 템플릿 로직이 `Dialect` 인터페이스의 메서드 호출로 변경.
4. **`internal/migration/migration.go`**: Direct 모드에서 `Dialect` 드라이버 정보를 바탕으로 `sql.Open` 인자 동적 결정. `Insert` 구문 생성 시 `Dialect.InsertStatement` 위임.
5. **`internal/web/*`**: `targetDb`, `targetUrl` 파라미터 핸들링 추가 및 Web UI 템플릿(JS 포함) 업데이트.

## 5. 테스트 전략 (Testing Strategy)

- **Unit Test**: `internal/dialect/` 아래 각 구현체들의 `MapOracleType`, `CreateTableDDL`, `InsertStatement`(특히 MSSQL 1000 row 분할 검증)에 대한 테스트 작성.
- **Integration Test**: Direct 마이그레이션 시 `Dialect` 선택에 따라 올바른 드라이버명과 URL 포맷으로 `database/sql` 초기화가 이뤄지는지 검증.
- **E2E Test / File Output**: SQL 파일 출력 시 각 방언별 특성이 쿼리에 올바르게 묻어나는지 텍스트 파일 내용 검증.
