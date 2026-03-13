# PRD (Product Requirements Document) - v6 품질 개선 및 버그 수정 (v7)

## 1. 개요 (Overview)

v6에서 멀티 타겟 DB 지원(`--target-db`)이 도입되었습니다.
이 버전(v7)은 v6 코드베이스를 면밀히 검토하여 발견된 **버그 수정**, **기능 누락**, **Web UI 불일치**, **테스트 공백**을 체계적으로 해소하는 **품질 개선 릴리스**입니다.
신규 기능 추가보다 기존 기능의 완성도와 안정성을 높이는 데 집중합니다.

---

## 2. 배경 및 문제 분석 (Background & Issues Found)

코드베이스 분석을 통해 다음 7가지 카테고리의 개선점을 발견했습니다.

### 2.1. 타입 매핑 정밀도 부족

| 방언 | 대상 타입 | 현재 동작 | 문제 |
|------|-----------|-----------|------|
| MySQL | `VARCHAR2(n)` | `VARCHAR(255)` 하드코딩 | precision 무시 → 데이터 손실 가능 (`VARCHAR2(4000)` 등) |
| MySQL | `CHAR(n)` | `CHAR(255)` 하드코딩 | precision 무시 |
| MSSQL | `VARCHAR2(n)` | `NVARCHAR(MAX)` 전체 적용 | precision 있을 경우 `NVARCHAR(n)` 사용이 성능상 유리 |
| MSSQL | `CHAR(n)` | `NCHAR(255)` 하드코딩 | precision 무시 |
| MSSQL | `NUMBER` 精度 없음 | `FLOAT` | PRD 표에서는 `DECIMAL` 계열로 명시됨 |

### 2.2. MSSQL DDL 조건 검사 미흡

#### 2.2.1. `CreateTableDDL` — `TABLE_SCHEMA` 조건 누락

현재 코드(`mssql.go`):
```sql
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'tablename')
```
동일한 이름의 테이블이 다른 스키마에 있을 경우 `IF NOT EXISTS`가 잘못된 결과를 반환하여 테이블 생성을 건너뜁니다.

#### 2.2.2. `CreateIndexDDL` — `object_id` 필터 누락

현재 코드(`mssql.go`):
```sql
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_name')
```
인덱스 이름이 다른 테이블에 같은 이름으로 존재할 경우 중복으로 오인하여 인덱스를 생성하지 않습니다.

### 2.3. Web UI — DDL 옵션 파일 출력 모드 미노출

`index.html`에서 `withDdl`, `withSequences`, `withIndexes`, `oracleOwner` 옵션이 **Direct Migration 토글(`directMigration`) 영역 내부**에만 위치합니다.
CLI에서는 `--with-ddl --with-sequences --with-indexes` 플래그가 파일 출력 모드에서도 완전히 동작하지만, Web UI에서는 파일 출력 모드일 때 이 옵션들에 접근할 방법이 없습니다.

```
현재 UI 구조:
[Direct Migration 체크박스] ← 해제 상태
  └─ (숨겨짐) pgUrl
  └─ (숨겨짐) withDdl ← 파일 모드에서 접근 불가
  └─ (숨겨짐) withSequences
  └─ (숨겨짐) withIndexes
  └─ (숨겨짐) oracleOwner
```

### 2.4. Web UI — 레이블/제목 PostgreSQL 종속 잔재

| 위치 | 현재 값 | 문제 |
|------|---------|------|
| `server.go` title | `"Oracle to PostgreSQL Migrator"` | 멀티 타겟 DB 지원 후 부적절 |
| `index.html` Schema 레이블 | `"PG Schema"` | PostgreSQL 전용 레이블 |
| `index.html` `<html lang>` | `"en"` | 한국어 혼용 UI와 불일치 |

### 2.5. WebSocket `warning` 메시지 타입 미구현

v6 PRD §4.6.4에서 다음을 명시했습니다:
```json
{ "type": "warning", "message": "MySQL은 Sequence를 지원하지 않습니다. AUTO_INCREMENT로 대체됩니다." }
```
그러나 `ws/tracker.go`에 `MsgType = "warning"`이 없으며, 마이그레이션 엔진에서도 `tracker.Warning(...)` 호출이 없습니다.
결과적으로 Sequence 미지원 방언 사용 시 사용자는 WebSocket 메시지가 아닌 서버 로그로만 경고를 확인할 수 있습니다.

### 2.6. Dry-Run — 대상 DB 연결 검증 미지원

`migration.go`의 Dry-Run 모드는 Oracle DB에 `SELECT COUNT(*)` 쿼리만 수행합니다.
`DryRunResult`의 `connectionOk`가 항상 `true`로 전달되므로, Direct 마이그레이션 대상 URL이 잘못된 경우에도 Dry-Run이 성공으로 표시됩니다.

### 2.7. 테스트 커버리지 부족

`internal/dialect/` 패키지의 MySQL, MariaDB, SQLite, MSSQL 구현체에 대한 단위 테스트가 없습니다.
(v6 `tasks.md` §7 항목이 TODO 상태로 남아 있음)

---

## 3. 목표 (Goals)

1. MySQL / MSSQL 타입 매핑 시 precision 값을 올바르게 반영한다.
2. MSSQL `CreateTableDDL` · `CreateIndexDDL`에 스키마/테이블 범위 조건을 추가한다.
3. Web UI 파일 출력 모드에서 DDL 관련 옵션(with-ddl, sequences, indexes)을 노출한다.
4. UI의 PostgreSQL 종속 레이블을 방언 중립적으로 수정한다.
5. WebSocket `warning` 메시지 타입을 구현하고 Sequence 미지원 경고를 프론트엔드로 전달한다.
6. Dry-Run 시 대상 DB 연결도 함께 검증하고 결과를 UI에 표시한다.
7. 4개 방언(MySQL, MariaDB, SQLite, MSSQL) 단위 테스트를 추가한다.

---

## 4. 기능 요구사항 (Functional Requirements)

### 4.1. 타입 매핑 정확도 개선

#### 4.1.1. MySQL `MapOracleType` 수정

| Oracle 타입 | 현재 | 수정 후 |
|------------|------|---------|
| `VARCHAR2(n)` | `VARCHAR(255)` | `VARCHAR(n)` (n ≤ 16383이면 그대로, 초과 시 `LONGTEXT`) |
| `CHAR(n)` | `CHAR(255)` 하드코딩 | `CHAR(n)` (precision 반영) |

- MySQL utf8mb4 환경에서 최대 Row 크기(65535 bytes)를 고려하여, 단일 컬럼 `VARCHAR(n)`의 n이 16383(≈ 65535÷4)을 초과하면 자동으로 `LONGTEXT`로 매핑한다.

#### 4.1.2. MSSQL `MapOracleType` 수정

| Oracle 타입 | 현재 | 수정 후 |
|------------|------|---------|
| `VARCHAR2(n)` | `NVARCHAR(MAX)` | precision ≤ 4000이면 `NVARCHAR(n)`, 초과이면 `NVARCHAR(MAX)` |
| `CHAR(n)` | `NCHAR(255)` | `NCHAR(n)` (precision 반영, 최대 4000) |
| `NUMBER` (precision 없음) | `FLOAT` | `NUMERIC` (Oracle NUMBER 기본 매핑과 일치) |

### 4.2. MSSQL DDL 조건 검사 수정

#### 4.2.1. `CreateTableDDL` — `TABLE_SCHEMA` 조건 추가

```sql
-- 수정 전
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES
               WHERE TABLE_NAME = 'tablename')

-- 수정 후 (schema 지정 시)
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES
               WHERE TABLE_SCHEMA = 'schemaname' AND TABLE_NAME = 'tablename')

-- 수정 후 (schema 미지정 시, dbo 기본값 사용)
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES
               WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = 'tablename')
```

#### 4.2.2. `CreateIndexDDL` — `object_id` 필터 추가

```sql
-- 수정 전
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_name')

-- 수정 후
IF NOT EXISTS (
    SELECT 1 FROM sys.indexes i
    JOIN sys.objects o ON i.object_id = o.object_id
    WHERE i.name = 'idx_name'
      AND o.name = 'tablename'
      AND SCHEMA_NAME(o.schema_id) = 'schemaname'
)
```

### 4.3. Web UI — DDL 옵션 파일 출력 모드 노출

현재 Direct Migration 토글 하위에만 있는 DDL 관련 옵션을 **공통 DDL 설정 섹션**으로 분리한다.

```
[목표 UI 구조]
Advanced Settings ▾
  ├─ 출력 대상 DB: [드롭다운]
  ├─ Batch Size / Workers
  └─ DDL 설정 (항상 표시, 파일 출력·Direct 공통)
       ├─ ☑ CREATE TABLE DDL 포함 (--with-ddl)
       ├─   ├─ ☐ Sequence DDL 포함 (--with-sequences)  ← withDdl 체크 시 활성
       ├─   ├─ ☐ Index DDL 포함 (--with-indexes)
       └─   └─ Oracle 소유자 입력 (oracle-owner)

[Direct Migration 체크박스]
  └─ (표시 시) 대상 URL 입력
```

- DDL 체크박스 비활성 상태에서 하위 옵션(sequences, indexes, oracleOwner)은 비활성화된다.
- Direct Migration 영역과 파일 출력 영역 모두 동일한 DDL 설정 섹션을 참조한다.

### 4.4. Web UI — 레이블/제목 수정

| 위치 | 현재 | 수정 후 |
|------|------|---------|
| `server.go` title 값 | `"Oracle to PostgreSQL Migrator"` | `"Oracle DB Migrator"` |
| Schema 입력 레이블 | `"PG Schema"` | `"Schema"` |

### 4.5. WebSocket `warning` 메시지 타입 구현

#### 4.5.1. `ws/tracker.go` 확장

```go
const (
    // 기존
    MsgInit         MsgType = "init"
    MsgUpdate       MsgType = "update"
    MsgDone         MsgType = "done"
    MsgError        MsgType = "error"
    MsgAllDone      MsgType = "all_done"
    MsgDryRunResult MsgType = "dry_run_result"
    MsgDDLProgress  MsgType = "ddl_progress"
    // 신규
    MsgWarning      MsgType = "warning"  // ← 추가
)
```

`WebSocketTracker`에 `Warning(message string)` 메서드를 추가한다.

#### 4.5.2. `ProgressTracker` 인터페이스 확장

```go
// WarningTracker extends ProgressTracker with warning broadcasting.
type WarningTracker interface {
    Warning(message string)
}
```

#### 4.5.3. 호출 위치

`MigrateTableToFile` / `MigrateTableDirect` 내부에서 Sequence DDL이 미지원(`supported == false`)인 경우:

```go
if wt, ok := tracker.(WarningTracker); ok {
    wt.Warning(fmt.Sprintf(
        "%s은(는) Sequence를 지원하지 않습니다. --with-sequences 옵션은 무시됩니다.",
        dia.Name(),
    ))
}
```

#### 4.5.4. Web UI `handleProgressMessage` 확장

```js
if (msg.type === 'warning') {
    // 경고 배너 또는 토스트 표시
    showWarningBanner(msg.message);
    return;
}
```

- 경고 메시지는 진행 컨테이너 상단에 노란색 배너로 표시된다.
- 동일한 경고가 여러 번 오면 중복 표시하지 않는다.

### 4.6. Dry-Run 대상 DB 연결 검증

Dry-Run 모드(`cfg.DryRun == true`)에서 `--target-url`이 지정된 경우, 대상 DB 연결을 시도하고 결과를 `DryRunResult.ConnectionOk`에 반영한다.

```go
// migration.go Run() 내 DryRun 블록
if cfg.TargetURL != "" {
    // 연결 시도 후 즉시 Close
    connOk = tryConnectTarget(dia, cfg.TargetURL)
}
for _, table := range cfg.Tables {
    // ...
    dryTracker.DryRunResult(table, count, connOk)
}
```

- 연결 실패 시 `connectionOk = false`로 전송하고, UI에 대상 DB 연결 실패 표시를 추가한다.

### 4.7. 방언별 단위 테스트 추가

`internal/dialect/` 패키지에 다음 테스트 파일을 추가한다:

| 파일 | 대상 |
|------|------|
| `mysql_test.go` | `MySQLDialect` |
| `mariadb_test.go` | `MariaDBDialect` |
| `sqlite_test.go` | `SQLiteDialect` |
| `mssql_test.go` | `MSSQLDialect` |

각 테스트 파일은 아래 테스트 케이스를 포함한다:

1. `TestMapOracleType_*` — 주요 Oracle 타입 → 대상 타입 매핑 검증
2. `TestCreateTableDDL_*` — 스키마 포함/미포함, NOT NULL 처리
3. `TestCreateIndexDDL_*` — 일반/UNIQUE/PK 인덱스, MSSQL object_id 조건 포함
4. `TestInsertStatement_*` — 단일 배치, 다중 배치 분할 (MSSQL 1000행 제한 포함)

---

## 5. 아키텍처 변경 요약 (Scope of Changes)

| 파일 | 변경 내용 |
|------|----------|
| `internal/dialect/mysql.go` | `MapOracleType`: VARCHAR2/CHAR precision 반영, VARCHAR 길이 범위 분기 |
| `internal/dialect/mssql.go` | `MapOracleType`: VARCHAR2/CHAR precision 반영, NUMBER 기본 매핑 수정 |
| `internal/dialect/mssql.go` | `CreateTableDDL`: TABLE_SCHEMA 조건 추가 |
| `internal/dialect/mssql.go` | `CreateIndexDDL`: sys.indexes + object_id 조건 추가 |
| `internal/dialect/mysql_test.go` | 신규 — MySQLDialect 단위 테스트 |
| `internal/dialect/mariadb_test.go` | 신규 — MariaDBDialect 단위 테스트 |
| `internal/dialect/sqlite_test.go` | 신규 — SQLiteDialect 단위 테스트 |
| `internal/dialect/mssql_test.go` | 신규 — MSSQLDialect 단위 테스트 |
| `internal/web/ws/tracker.go` | `MsgWarning` 상수, `Warning()` 메서드 추가 |
| `internal/migration/migration.go` | `WarningTracker` 인터페이스 정의, Dry-Run 대상 DB 연결 검증, Sequence 미지원 warning 호출 |
| `internal/web/server.go` | title 값 수정 |
| `internal/web/templates/index.html` | DDL 옵션 위치 재구성, 레이블 수정, warning 메시지 표시 로직 추가 |

---

## 6. 비기능 요구사항 (Non-Functional Requirements)

- **하위 호환성**: 모든 변경은 기존 CLI·API 인터페이스와 완전 호환되어야 한다.
- **테스트 통과**: `go test ./...` 전체가 통과해야 한다.
- **데이터 무결성**: VARCHAR 매핑 변경으로 인해 기존 PostgreSQL 출력 결과는 변경되지 않아야 한다.

---

## 7. 마일스톤 (Milestones)

1. **PRD 확정**: `docs/v7/prd.md` 작성 ✅
2. **타입 매핑 수정**: MySQL/MSSQL precision 반영
3. **MSSQL DDL 조건 수정**: TABLE_SCHEMA, object_id 필터
4. **WebSocket warning 구현**: tracker + 인터페이스 + 호출 지점
5. **Dry-Run 대상 DB 검증**: 연결 시도 로직 추가
6. **Web UI 개선**: DDL 섹션 재구성 + 레이블 수정 + warning 배너
7. **방언별 단위 테스트 추가**
8. **전체 테스트 통과 및 리뷰**

---

## 8. 미포함 사항 (Out of Scope)

다음은 이번 v7 범위에 포함되지 않으며, 별도 PRD로 분리합니다:

- 새로운 대상 DB 추가 (CockroachDB, TiDB 등)
- 소스 DB 다양화 (MySQL → PostgreSQL 등 Oracle 이외 소스)
- Web 서버 멀티 세션 동시 마이그레이션 지원
- 연결 풀 세밀 조정(MaxOpenConns 등) 설정 노출
