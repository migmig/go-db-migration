# v7 Implementation Tasks

## 1. MySQL 타입 매핑 precision 반영

> **파일**: `internal/dialect/mysql.go`
> **Spec 참조**: §2

- [x] `MapOracleType`에서 `VARCHAR2`와 `CHAR`의 합쳐진 `case` 분기를 분리한다.
- [x] `VARCHAR2` 분기: precision > 0이면 `VARCHAR(n)` 반환, n > 16383이면 `LONGTEXT` 반환.
- [x] `CHAR` 분기: precision > 0이면 `CHAR(n)` 반환, 기본값 `CHAR(255)` 유지.
- [x] 기존 `NUMBER`, `DATE`, `CLOB`, `BLOB`, `FLOAT` 분기는 변경하지 않는다.

## 2. MSSQL 타입 매핑 precision 반영

> **파일**: `internal/dialect/mssql.go`
> **Spec 참조**: §3

- [x] `MapOracleType`에서 `VARCHAR2` 분기: precision > 0 && ≤ 4000이면 `NVARCHAR(n)`, 초과이면 `NVARCHAR(MAX)`.
- [x] `CHAR` 분기: precision > 0이면 `NCHAR(n)` (최대 4000 클램핑), 기본값 `NCHAR(255)` 유지.
- [x] `NUMBER` (precision 없음) 기본 반환값을 `FLOAT` → `NUMERIC`으로 변경한다.

## 3. MSSQL `CreateTableDDL` — TABLE_SCHEMA 조건 추가

> **파일**: `internal/dialect/mssql.go`
> **Spec 참조**: §4.1

- [x] `IF NOT EXISTS` 쿼리에 `TABLE_SCHEMA` 조건을 추가한다.
- [x] schema가 빈 문자열이면 `dbo`를 기본값으로 사용한다.
- [x] 기존 `TABLE_NAME` 조건은 유지한다.

## 4. MSSQL `CreateIndexDDL` — object_id 필터 추가

> **파일**: `internal/dialect/mssql.go`
> **Spec 참조**: §4.2

- [x] `IF NOT EXISTS` 쿼리를 `sys.indexes i JOIN sys.objects o ON i.object_id = o.object_id` 형태로 변경한다.
- [x] `WHERE` 절에 `i.name`, `o.name` (테이블명), `SCHEMA_NAME(o.schema_id)` (스키마명) 조건을 추가한다.
- [x] schema가 빈 문자열이면 `dbo`를 기본값으로 사용한다.
- [x] `IsPK == true`인 경우의 `ALTER TABLE ADD PRIMARY KEY` 로직은 변경하지 않는다.

## 5. WebSocket `warning` 메시지 타입 — tracker 확장

> **파일**: `internal/web/ws/tracker.go`
> **Spec 참조**: §5.1

- [x] `MsgWarning MsgType = "warning"` 상수를 추가한다.
- [x] `ProgressMsg` 구조체에 `Message string \`json:"message,omitempty"\`` 필드를 추가한다.
- [x] `WebSocketTracker`에 `Warning(message string)` 메서드를 추가한다. (`broadcast`로 `MsgWarning` + `Message` 전송)

## 6. `WarningTracker` 인터페이스 및 호출 지점

> **파일**: `internal/migration/migration.go`
> **Spec 참조**: §5.2, §5.3

- [x] `WarningTracker` 인터페이스를 정의한다: `Warning(message string)`.
- [x] `MigrateTableDirect` 내 `cfg.WithSequences` 블록에서 `supported == false`일 때 `WarningTracker.Warning()` 호출을 추가한다.
- [x] `MigrateTableToFile` 내 동일 위치에 같은 패턴을 적용한다.

## 7. Dry-Run 대상 DB 연결 검증

> **파일**: `internal/migration/migration.go`
> **Spec 참조**: §6

- [x] `tryConnectTarget(dia dialect.Dialect, targetURL string) bool` 헬퍼 함수를 추가한다.
  - PostgreSQL: `db.ConnectPostgres` → 즉시 Close.
  - 기타: `sql.Open` + `Ping()` → 즉시 Close.
- [x] `Run()` 함수의 DryRun 블록에서 `cfg.TargetURL != ""`이면 `tryConnectTarget`를 호출한다.
- [x] 결과를 `connOk` 변수에 저장하고, `DryRunResult(table, count, connOk)`에 전달한다.

## 8. Web UI — 타이틀 및 레이블 수정

> **파일**: `internal/web/server.go`, `internal/web/templates/index.html`
> **Spec 참조**: §7.1, §7.2

- [x] `server.go`에서 title 값을 `"Oracle to PostgreSQL Migrator"` → `"Oracle DB Migrator"`로 변경한다.
- [x] `index.html`에서 Schema 입력 필드 레이블을 `"PG Schema"` → `"Schema"`로 변경한다.

## 9. Web UI — DDL 옵션 파일 출력 모드 노출

> **파일**: `internal/web/templates/index.html`
> **Spec 참조**: §7.3

- [x] DDL 관련 옵션(`withDdl`, `withSequences`, `withIndexes`, `oracleOwner`)을 Direct Migration 토글 내부에서 **공통 DDL 설정 섹션**으로 분리한다.
- [x] DDL 설정 섹션은 Direct Migration 체크 여부와 무관하게 항상 표시되도록 한다.
- [x] `withDdl` 체크박스 변경 이벤트: 해제 시 하위 옵션(sequences, indexes, oracleOwner)을 숨기고 값을 초기화한다.
- [x] Direct Migration 토글 내부에는 대상 URL 입력만 남긴다.

## 10. Web UI — warning 배너 및 Dry-Run 연결 실패 표시

> **파일**: `internal/web/templates/index.html`
> **Spec 참조**: §7.4, §7.5

- [x] `handleProgressMessage`에 `msg.type === 'warning'` 분기를 추가하여 `showWarningBanner(msg.message)`를 호출한다.
- [x] `showWarningBanner` 함수를 구현한다: `Set`으로 중복 방지, 진행 컨테이너 상단에 노란색 배너 삽입.
- [x] `.warning-banner` CSS 스타일을 추가한다 (`#fff3cd` 배경, `#ffc107` 테두리, `#856404` 텍스트).
- [x] `dry_run_result` 처리에서 `connection_ok === false`이면 대상 DB 연결 실패 경고 배너를 표시한다.

## 11. 단위 테스트 — MySQL

> **파일**: `internal/dialect/mysql_test.go` (신규)
> **Spec 참조**: §8

- [x] `TestMapOracleType_MySQL` — VARCHAR2 precision 반영 (정상, 초과→LONGTEXT, 미지정→255), CHAR precision 반영, NUMBER, DATE, CLOB, BLOB, FLOAT 매핑.
- [x] `TestCreateTableDDL_MySQL` — 스키마 포함/미포함, NOT NULL 처리, `IF NOT EXISTS` 포함 검증.
- [x] `TestCreateIndexDDL_MySQL` — 일반 인덱스, UNIQUE 인덱스, PRIMARY KEY (ALTER TABLE).
- [x] `TestInsertStatement_MySQL` — 단일 배치, 다중 배치 분할, 값 포맷팅 (문자열 이스케이프, NULL, 시간).

## 12. 단위 테스트 — MariaDB

> **파일**: `internal/dialect/mariadb_test.go` (신규)
> **Spec 참조**: §8.3

- [x] `TestMariaDB_Name` — `Name()` 반환값이 `"mariadb"`인지 검증.
- [x] `TestMariaDB_InheritsMySQL` — MySQL과 동일한 타입 매핑 (VARCHAR2 precision 포함) 확인.
- [x] `TestCreateTableDDL_MariaDB` — MySQL과 동일한 DDL 생성 확인.
- [x] `TestInsertStatement_MariaDB` — MySQL과 동일한 INSERT 문 생성 확인.

## 13. 단위 테스트 — SQLite

> **파일**: `internal/dialect/sqlite_test.go` (신규)
> **Spec 참조**: §8

- [x] `TestMapOracleType_SQLite` — VARCHAR2→TEXT, NUMBER→INTEGER/REAL, DATE→TEXT, CLOB→TEXT, BLOB→BLOB, FLOAT→REAL.
- [x] `TestCreateTableDDL_SQLite` — 스키마 무시 (schema 전달해도 DDL에 미포함), NOT NULL, `IF NOT EXISTS`.
- [x] `TestCreateIndexDDL_SQLite` — 일반 인덱스 `IF NOT EXISTS`, UNIQUE, PK 건너뛰기 (빈 문자열 반환).
- [x] `TestInsertStatement_SQLite` — 단일 배치, 다중 배치 분할, 값 포맷팅.

## 14. 단위 테스트 — MSSQL

> **파일**: `internal/dialect/mssql_test.go` (신규)
> **Spec 참조**: §8

- [x] `TestMapOracleType_MSSQL` — VARCHAR2 precision ≤4000 → NVARCHAR(n), >4000 → NVARCHAR(MAX), CHAR precision → NCHAR(n), NUMBER 무precision → NUMERIC (v7 수정).
- [x] `TestCreateTableDDL_MSSQL` — TABLE_SCHEMA 조건 포함 (schema 지정/미지정→dbo), NOT NULL, 컬럼 타입 (v7 수정).
- [x] `TestCreateIndexDDL_MSSQL` — sys.objects JOIN + object_id 필터 (v7 수정), 스키마 포함/미포함→dbo, UNIQUE, PK (ALTER TABLE).
- [x] `TestInsertStatement_MSSQL` — 1000행 배치 제한 (batchSize > 1000 → 자동 분할), 값 포맷팅 (N'' 유니코드, 0x 바이너리).

## 15. WebSocket tracker 테스트 확장

> **파일**: `internal/web/ws/tracker_test.go`
> **Spec 참조**: §10.2

- [x] `TestWarning` — `Warning()` 호출 시 `MsgWarning` 타입 + `Message` 필드가 올바르게 브로드캐스트되는지 검증.

## 16. 전체 테스트 통과 확인

- [x] `go test ./...` 실행하여 전체 테스트 통과를 확인한다.
- [x] `--target-db postgres` 출력이 v7 수정 전과 동일한지 확인한다 (PostgreSQL 하위 호환성).
