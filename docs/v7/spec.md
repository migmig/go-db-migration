# Technical Specification (Spec) - v6 품질 개선 및 버그 수정 (v7)

## 1. 개요 (Overview)

본 문서는 `dbmigrator` v7의 기술적 명세를 정의합니다.
v7은 v6에서 도입된 멀티 타겟 DB 지원 코드베이스를 면밀히 검토하여 발견된 **버그 수정**, **기능 누락**, **Web UI 불일치**, **테스트 공백**을 체계적으로 해소하는 **품질 개선 릴리스**입니다.

변경 대상 파일:

| 파일 | 변경 유형 |
|------|-----------|
| `internal/dialect/mysql.go` | 수정 — 타입 매핑 precision 반영 |
| `internal/dialect/mssql.go` | 수정 — 타입 매핑 precision 반영, DDL 조건 강화 |
| `internal/dialect/mysql_test.go` | 신규 — 단위 테스트 |
| `internal/dialect/mariadb_test.go` | 신규 — 단위 테스트 |
| `internal/dialect/sqlite_test.go` | 신규 — 단위 테스트 |
| `internal/dialect/mssql_test.go` | 신규 — 단위 테스트 |
| `internal/web/ws/tracker.go` | 수정 — `MsgWarning` 상수 및 `Warning()` 메서드 추가 |
| `internal/migration/migration.go` | 수정 — `WarningTracker` 인터페이스, Dry-Run 대상 DB 검증, warning 호출 |
| `internal/web/server.go` | 수정 — title 값 변경 |
| `internal/web/templates/index.html` | 수정 — DDL 옵션 위치 재구성, 레이블 수정, warning 배너 |

---

## 2. MySQL 타입 매핑 정확도 개선 (`internal/dialect/mysql.go`)

### 2.1. 현재 코드 문제

`MapOracleType` 메서드에서 `VARCHAR2`와 `CHAR` 타입의 precision 값을 무시하고 하드코딩된 값을 반환합니다.

```go
// 현재 (버그)
case strings.Contains(oracleType, "VARCHAR2") || strings.Contains(oracleType, "CHAR"):
    return "VARCHAR(255)"
```

### 2.2. 수정 명세

`MapOracleType` 메서드의 `VARCHAR2`/`CHAR` 분기를 분리하고 precision을 반영합니다.

```go
func (d *MySQLDialect) MapOracleType(oracleType string, precision, scale int) string {
    oracleType = strings.ToUpper(oracleType)

    switch {
    case strings.Contains(oracleType, "VARCHAR2"):
        if precision > 0 {
            if precision > 16383 {
                return "LONGTEXT"
            }
            return fmt.Sprintf("VARCHAR(%d)", precision)
        }
        return "VARCHAR(255)"

    case strings.Contains(oracleType, "CHAR"):
        if precision > 0 {
            return fmt.Sprintf("CHAR(%d)", precision)
        }
        return "CHAR(255)"

    // ... 나머지 분기는 기존과 동일
    }
}
```

#### 2.2.1. VARCHAR2 매핑 규칙

| Oracle 입력 | precision | 결과 | 근거 |
|-------------|-----------|------|------|
| `VARCHAR2(100)` | 100 | `VARCHAR(100)` | precision 그대로 반영 |
| `VARCHAR2(4000)` | 4000 | `VARCHAR(4000)` | MySQL utf8mb4 기준 16383 이하 |
| `VARCHAR2(20000)` | 20000 | `LONGTEXT` | 16383 초과 → LONGTEXT |
| `VARCHAR2` (precision 없음) | 0 | `VARCHAR(255)` | 기존 기본값 유지 |

- **16383 임계값 근거**: MySQL utf8mb4(4바이트 문자셋) 환경에서 단일 `VARCHAR(n)`의 최대 바이트 크기는 65,535 bytes. `65535 ÷ 4 = 16383.75`이므로 n ≤ 16383이 유효 범위.

#### 2.2.2. CHAR 매핑 규칙

| Oracle 입력 | precision | 결과 |
|-------------|-----------|------|
| `CHAR(10)` | 10 | `CHAR(10)` |
| `CHAR` (precision 없음) | 0 | `CHAR(255)` |

- MySQL `CHAR` 최대 길이는 255이므로, Oracle `CHAR(n)`에서 n > 255인 경우는 실제로 발생하지 않음 (Oracle CHAR 최대 2000 bytes이나, MySQL 제한에 맞춰 그대로 전달).

### 2.3. 영향 범위

- `MariaDBDialect`는 `MySQLDialect`를 임베딩하므로 동일하게 적용됩니다.
- PostgreSQL의 `VARCHAR2` → `text` 매핑은 변경 없음 (precision 무시가 의도된 동작).

---

## 3. MSSQL 타입 매핑 정확도 개선 (`internal/dialect/mssql.go`)

### 3.1. 현재 코드 문제

```go
// 현재 (버그)
case strings.Contains(oracleType, "VARCHAR2"):
    return "NVARCHAR(MAX)"           // precision 무시
case strings.Contains(oracleType, "CHAR"):
    return "NCHAR(255)"              // precision 무시
// ...
    return "FLOAT"                   // NUMBER(precision 없음) → FLOAT (PRD에서는 NUMERIC 계열)
```

### 3.2. 수정 명세

```go
func (d *MSSQLDialect) MapOracleType(oracleType string, precision, scale int) string {
    oracleType = strings.ToUpper(oracleType)

    switch {
    case strings.Contains(oracleType, "VARCHAR2"):
        if precision > 0 {
            if precision <= 4000 {
                return fmt.Sprintf("NVARCHAR(%d)", precision)
            }
            return "NVARCHAR(MAX)"
        }
        return "NVARCHAR(MAX)"

    case strings.Contains(oracleType, "CHAR"):
        if precision > 0 {
            if precision > 4000 {
                return "NCHAR(4000)"
            }
            return fmt.Sprintf("NCHAR(%d)", precision)
        }
        return "NCHAR(255)"

    case oracleType == "NUMBER":
        if precision > 0 {
            if scale > 0 {
                return fmt.Sprintf("DECIMAL(%d, %d)", precision, scale)
            }
            if precision <= 4 {
                return "SMALLINT"
            }
            if precision <= 9 {
                return "INT"
            }
            return "BIGINT"
        }
        return "NUMERIC"    // 변경: FLOAT → NUMERIC

    // ... 나머지 분기는 기존과 동일
    }
}
```

#### 3.2.1. VARCHAR2 매핑 규칙

| Oracle 입력 | precision | 결과 | 근거 |
|-------------|-----------|------|------|
| `VARCHAR2(500)` | 500 | `NVARCHAR(500)` | precision 반영 |
| `VARCHAR2(4000)` | 4000 | `NVARCHAR(4000)` | MSSQL NVARCHAR 최대 |
| `VARCHAR2(8000)` | 8000 | `NVARCHAR(MAX)` | 4000 초과 → MAX |
| `VARCHAR2` (precision 없음) | 0 | `NVARCHAR(MAX)` | 기존 기본값 유지 |

- **4000 임계값 근거**: MSSQL `NVARCHAR(n)`의 최대 n은 4000(유니코드 문자). 초과 시 `NVARCHAR(MAX)` 사용 필수.

#### 3.2.2. CHAR 매핑 규칙

| Oracle 입력 | precision | 결과 |
|-------------|-----------|------|
| `CHAR(50)` | 50 | `NCHAR(50)` |
| `CHAR(5000)` | 5000 | `NCHAR(4000)` |
| `CHAR` (precision 없음) | 0 | `NCHAR(255)` |

- MSSQL `NCHAR` 최대 길이는 4000. 초과 시 4000으로 클램핑.

#### 3.2.3. NUMBER (precision 없음) 매핑 변경

| 변경 전 | 변경 후 | 근거 |
|---------|---------|------|
| `FLOAT` | `NUMERIC` | Oracle `NUMBER`는 정밀 소수 타입. `FLOAT`는 근사값 타입으로 데이터 정밀도 손실 가능. `NUMERIC`이 Oracle `NUMBER`의 동작과 일치 |

---

## 4. MSSQL DDL 조건 검사 강화 (`internal/dialect/mssql.go`)

### 4.1. `CreateTableDDL` — `TABLE_SCHEMA` 조건 추가

#### 4.1.1. 현재 코드 문제

```go
// 현재 (버그)
sb.WriteString(fmt.Sprintf(
    "IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = '%s')\n",
    bareTableName,
))
```

동일한 이름의 테이블이 다른 스키마(예: `hr.employees`, `sales.employees`)에 존재할 경우, `TABLE_SCHEMA` 조건이 없어 `IF NOT EXISTS`가 잘못된 결과를 반환합니다.

#### 4.1.2. 수정 명세

```go
func (d *MSSQLDialect) CreateTableDDL(tableName, schema string, cols []ColumnDef) string {
    fullTableName := d.QuoteIdentifier(strings.ToLower(tableName))
    if schema != "" {
        fullTableName = fmt.Sprintf("%s.%s",
            d.QuoteIdentifier(strings.ToLower(schema)), fullTableName)
    }

    var sb strings.Builder
    bareTableName := strings.ToLower(tableName)

    // schema 조건 포함
    effectiveSchema := strings.ToLower(schema)
    if effectiveSchema == "" {
        effectiveSchema = "dbo"   // MSSQL 기본 스키마
    }

    sb.WriteString(fmt.Sprintf(
        "IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = '%s' AND TABLE_NAME = '%s')\n",
        effectiveSchema, bareTableName,
    ))
    sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", fullTableName))

    // ... 컬럼 정의는 기존과 동일
}
```

#### 4.1.3. 생성 결과 예시

**schema 지정 시** (`schema = "sales"`):
```sql
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES
               WHERE TABLE_SCHEMA = 'sales' AND TABLE_NAME = 'employees')
CREATE TABLE [sales].[employees] (
    [id] INT NOT NULL,
    [name] NVARCHAR(100)
);
```

**schema 미지정 시** (`schema = ""`):
```sql
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES
               WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = 'employees')
CREATE TABLE [employees] (
    [id] INT NOT NULL,
    [name] NVARCHAR(100)
);
```

### 4.2. `CreateIndexDDL` — `object_id` 필터 추가

#### 4.2.1. 현재 코드 문제

```go
// 현재 (버그)
sb.WriteString(fmt.Sprintf(
    "IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = '%s')\n",
    indexName,
))
```

인덱스 이름이 다른 테이블에 동일하게 존재할 경우, 인덱스 생성이 의도치 않게 건너뛰어집니다.

#### 4.2.2. 수정 명세

```go
func (d *MSSQLDialect) CreateIndexDDL(idx IndexMetadata, tableName, schema string) string {
    // ... 기존 table, colExprs 구성 동일

    if idx.IsPK {
        return fmt.Sprintf("ALTER TABLE %s ADD PRIMARY KEY (%s);\n", table, colList)
    }

    indexName := strings.ToLower(idx.Name)
    quotedIndexName := d.QuoteIdentifier(indexName)

    uniqueStr := ""
    if idx.Uniqueness == "UNIQUE" {
        uniqueStr = "UNIQUE "
    }

    // schema/table 범위 조건 포함
    bareTableName := strings.ToLower(tableName)
    effectiveSchema := strings.ToLower(schema)
    if effectiveSchema == "" {
        effectiveSchema = "dbo"
    }

    var sb strings.Builder
    sb.WriteString(fmt.Sprintf(
        "IF NOT EXISTS (\n"+
        "    SELECT 1 FROM sys.indexes i\n"+
        "    JOIN sys.objects o ON i.object_id = o.object_id\n"+
        "    WHERE i.name = '%s'\n"+
        "      AND o.name = '%s'\n"+
        "      AND SCHEMA_NAME(o.schema_id) = '%s'\n"+
        ")\n",
        indexName, bareTableName, effectiveSchema,
    ))
    sb.WriteString(fmt.Sprintf(
        "    CREATE %sINDEX %s ON %s (%s);\n",
        uniqueStr, quotedIndexName, table, colList,
    ))

    return sb.String()
}
```

#### 4.2.3. 생성 결과 예시

```sql
IF NOT EXISTS (
    SELECT 1 FROM sys.indexes i
    JOIN sys.objects o ON i.object_id = o.object_id
    WHERE i.name = 'idx_emp_name'
      AND o.name = 'employees'
      AND SCHEMA_NAME(o.schema_id) = 'sales'
)
    CREATE INDEX [idx_emp_name] ON [sales].[employees] ([name]);
```

---

## 5. WebSocket `warning` 메시지 타입 구현

### 5.1. `ws/tracker.go` 확장

#### 5.1.1. 상수 추가

```go
const (
    MsgInit         MsgType = "init"
    MsgUpdate       MsgType = "update"
    MsgDone         MsgType = "done"
    MsgError        MsgType = "error"
    MsgAllDone      MsgType = "all_done"
    MsgDryRunResult MsgType = "dry_run_result"
    MsgDDLProgress  MsgType = "ddl_progress"
    MsgWarning      MsgType = "warning"     // 신규
)
```

#### 5.1.2. `ProgressMsg` 구조체 확장

기존 `ProgressMsg`의 `ErrorMsg` 필드(string)를 경고 메시지에도 재사용합니다. `warning` 타입에서는 `ErrorMsg` 필드가 실질적 경고 메시지를 담습니다.

추가로, 전용 `Message` 필드를 도입하여 경고 메시지의 의미를 명확히 합니다.

```go
type ProgressMsg struct {
    Type         MsgType `json:"type"`
    Table        string  `json:"table,omitempty"`
    Count        int     `json:"count,omitempty"`
    Total        int     `json:"total,omitempty"`
    ErrorMsg     string  `json:"error,omitempty"`
    Message      string  `json:"message,omitempty"`   // 신규 — warning 메시지용
    ZipFileID    string  `json:"zip_file_id,omitempty"`
    ConnectionOk bool    `json:"connection_ok,omitempty"`
    Object       string  `json:"object,omitempty"`
    ObjectName   string  `json:"object_name,omitempty"`
    Status       string  `json:"status,omitempty"`
}
```

#### 5.1.3. `Warning()` 메서드 추가

```go
func (t *WebSocketTracker) Warning(message string) {
    t.broadcast(ProgressMsg{
        Type:    MsgWarning,
        Message: message,
    })
}
```

### 5.2. `WarningTracker` 인터페이스 정의 (`internal/migration/migration.go`)

기존 `ProgressTracker`, `DryRunTracker`, `DDLProgressTracker`와 동일한 패턴으로 선택적 확장 인터페이스를 정의합니다.

```go
// WarningTracker extends ProgressTracker with warning broadcasting.
type WarningTracker interface {
    Warning(message string)
}
```

### 5.3. 호출 지점 — Sequence 미지원 경고

`MigrateTableToFile`과 `MigrateTableDirect` 내부에서 Sequence DDL이 미지원(`supported == false`)인 경우 경고를 전송합니다.

#### 5.3.1. `MigrateTableDirect` 내 호출 위치

```go
// cfg.WithSequences 블록 내부, CreateSequenceDDL 호출 후
ddl, supported := GenerateSequenceDDL(seq, cfg.Schema, dia)
if !supported || ddl == "" {
    slog.Warn("Sequence not supported by dialect", "dialect", dia.Name(), "sequence", seq.Name)
    // 신규: WebSocket warning 전송
    if wt, ok := tracker.(WarningTracker); ok {
        wt.Warning(fmt.Sprintf(
            "%s은(는) Sequence를 지원하지 않습니다. --with-sequences 옵션은 무시됩니다.",
            dia.Name(),
        ))
    }
    continue
}
```

#### 5.3.2. `MigrateTableToFile` 내 동일 위치에 같은 패턴 적용

```go
ddl, supported := GenerateSequenceDDL(seq, cfg.Schema, dia)
if !supported || ddl == "" {
    slog.Warn("Sequence not supported by dialect", "dialect", dia.Name(), "sequence", seq.Name)
    if wt, ok := tracker.(WarningTracker); ok {
        wt.Warning(fmt.Sprintf(
            "%s은(는) Sequence를 지원하지 않습니다. --with-sequences 옵션은 무시됩니다.",
            dia.Name(),
        ))
    }
    continue
}
```

### 5.4. 중복 경고 방지

동일한 방언에 대해 Sequence 미지원 경고가 테이블마다 반복 전송되지 않도록, `MigrateTableDirect`/`MigrateTableToFile` 호출 전 상위 레벨(`Run` 또는 `worker`)에서 경고를 1회만 전송하는 것이 이상적이나, 현재 아키텍처에서 worker 간 상태 공유가 복잡하므로 **프론트엔드에서 중복 방지**하는 전략을 채택합니다.

---

## 6. Dry-Run 대상 DB 연결 검증 (`internal/migration/migration.go`)

### 6.1. 현재 코드 문제

```go
// 현재 — DryRun 블록
if hasDryTracker {
    dryTracker.DryRunResult(table, count, true)  // connectionOk 항상 true
}
```

`cfg.TargetURL`이 지정되어 있더라도 대상 DB 연결을 검증하지 않아, 잘못된 URL도 성공으로 표시됩니다.

### 6.2. 수정 명세

`Run()` 함수의 DryRun 블록에 대상 DB 연결 검증 로직을 추가합니다.

```go
if cfg.DryRun {
    slog.Info("Dry run mode enabled. Verifying connectivity and estimating row counts.")
    dryTracker, hasDryTracker := tracker.(DryRunTracker)

    // 신규: 대상 DB 연결 검증
    connOk := true
    if cfg.TargetURL != "" {
        connOk = tryConnectTarget(dia, cfg.TargetURL)
        if !connOk {
            slog.Warn("Target DB connection failed during dry-run",
                "targetDB", dia.Name(), "url", cfg.TargetURL)
        } else {
            slog.Info("Target DB connection verified",
                "targetDB", dia.Name())
        }
    }

    for _, table := range cfg.Tables {
        var count int
        err := dbConn.QueryRow(
            fmt.Sprintf("SELECT COUNT(*) FROM %s", table),
        ).Scan(&count)
        if err != nil {
            slog.Error("failed to get row count for table", "table", table, "error", err)
            if tracker != nil {
                tracker.Error(table, err)
            }
            continue
        }
        slog.Info("table estimation", "table", table, "estimated_rows", count)
        if hasDryTracker {
            dryTracker.DryRunResult(table, count, connOk)  // 실제 검증 결과 반영
        }
    }
    slog.Info("Dry run completed successfully.")
    return nil
}
```

### 6.3. `tryConnectTarget` 헬퍼 함수

```go
// tryConnectTarget attempts to open and ping the target database.
// Returns true if the connection succeeds, false otherwise.
func tryConnectTarget(dia dialect.Dialect, targetURL string) bool {
    if dia.Name() == "postgres" {
        // PostgreSQL은 pgx 풀을 사용하므로 별도 처리
        pool, err := db.ConnectPostgres(targetURL)
        if err != nil {
            return false
        }
        pool.Close()
        return true
    }

    conn, err := sql.Open(dia.DriverName(), dia.NormalizeURL(targetURL))
    if err != nil {
        return false
    }
    defer conn.Close()

    if err := conn.Ping(); err != nil {
        return false
    }
    return true
}
```

- `sql.Open`만으로는 실제 연결이 성립되지 않으므로 `Ping()`을 반드시 호출합니다.
- 연결 성공 시 즉시 `Close()`하여 리소스를 해제합니다.

---

## 7. Web UI 개선 (`internal/web/server.go`, `internal/web/templates/index.html`)

### 7.1. 타이틀 수정 (`server.go`)

```go
// 변경 전
c.HTML(http.StatusOK, "index.html", gin.H{
    "title": "Oracle to PostgreSQL Migrator",
})

// 변경 후
c.HTML(http.StatusOK, "index.html", gin.H{
    "title": "Oracle DB Migrator",
})
```

### 7.2. Schema 레이블 수정 (`index.html`)

| 위치 | 변경 전 | 변경 후 |
|------|---------|---------|
| Schema 입력 필드 레이블 | `PG Schema` | `Schema` |

### 7.3. DDL 옵션 파일 출력 모드 노출 (`index.html`)

#### 7.3.1. 현재 구조 (문제)

```
[Direct Migration 체크박스] ← 해제 시 하위 전체 숨김
  └─ pgUrl / targetUrl
  └─ withDdl          ← 파일 모드에서 접근 불가
  └─ withSequences
  └─ withIndexes
  └─ oracleOwner
```

#### 7.3.2. 변경 후 구조

DDL 관련 옵션을 Direct Migration 토글과 독립적인 **공통 DDL 설정 섹션**으로 분리합니다.

```
Advanced Settings ▾
  ├─ 출력 대상 DB: [드롭다운]
  ├─ Batch Size / Workers / Schema
  └─ DDL 설정 (항상 표시)
       ├─ ☑ CREATE TABLE DDL 포함 (withDdl)
       │    ├─ ☐ Sequence DDL 포함 (withSequences)   ← withDdl 체크 시 활성
       │    ├─ ☐ Index DDL 포함 (withIndexes)         ← withDdl 체크 시 활성
       │    └─ Oracle 소유자 입력 (oracleOwner)        ← withDdl 체크 시 활성
       └─ (withDdl 미체크 시 하위 비활성)

[Direct Migration 체크박스]
  └─ (표시 시) 대상 URL 입력만 노출
```

#### 7.3.3. JavaScript 동작 명세

```javascript
// withDdl 체크박스 변경 이벤트
document.getElementById('withDdl').addEventListener('change', function() {
    const ddlSubOptions = document.getElementById('ddlSubOptions');
    if (this.checked) {
        ddlSubOptions.style.display = 'block';
    } else {
        ddlSubOptions.style.display = 'none';
        // 하위 옵션 초기화
        document.getElementById('withSequences').checked = false;
        document.getElementById('withIndexes').checked = false;
        document.getElementById('oracleOwner').value = '';
    }
});
```

- DDL 체크박스(`withDdl`)가 해제되면 하위 옵션(sequences, indexes, oracleOwner)을 숨기고 값을 초기화합니다.
- Direct Migration 체크 여부와 무관하게 DDL 설정 섹션은 항상 표시됩니다.

### 7.4. WebSocket `warning` 메시지 처리 (`index.html`)

#### 7.4.1. `handleProgressMessage` 함수 확장

```javascript
function handleProgressMessage(msg) {
    // ... 기존 분기

    if (msg.type === 'warning') {
        showWarningBanner(msg.message);
        return;
    }

    // ... 기존 분기 계속
}
```

#### 7.4.2. `showWarningBanner` 함수

```javascript
const shownWarnings = new Set();

function showWarningBanner(message) {
    // 중복 방지
    if (shownWarnings.has(message)) {
        return;
    }
    shownWarnings.add(message);

    const container = document.getElementById('progressContainer');
    const banner = document.createElement('div');
    banner.className = 'warning-banner';
    banner.textContent = '⚠ ' + message;
    container.insertBefore(banner, container.firstChild);
}
```

#### 7.4.3. CSS 스타일

```css
.warning-banner {
    background-color: #fff3cd;
    border: 1px solid #ffc107;
    color: #856404;
    padding: 10px 15px;
    border-radius: 4px;
    margin-bottom: 10px;
    font-size: 14px;
}
```

### 7.5. Dry-Run 대상 DB 연결 실패 표시

기존 `dry_run_result` 메시지의 `connection_ok` 필드가 `false`인 경우 UI에 연결 실패를 표시합니다.

```javascript
if (msg.type === 'dry_run_result') {
    // 기존 로직
    addDryRunRow(msg.table, msg.total, msg.connection_ok);

    // 신규: connection_ok가 false이면 경고 배너 추가
    if (!msg.connection_ok) {
        showWarningBanner('대상 DB 연결에 실패했습니다. URL을 확인해 주세요.');
    }
}
```

---

## 8. 방언별 단위 테스트 추가 (`internal/dialect/`)

### 8.1. 테스트 파일 구조

| 파일 | 대상 구현체 |
|------|------------|
| `mysql_test.go` | `MySQLDialect` |
| `mariadb_test.go` | `MariaDBDialect` |
| `sqlite_test.go` | `SQLiteDialect` |
| `mssql_test.go` | `MSSQLDialect` |

### 8.2. 공통 테스트 케이스 패턴

모든 방언 테스트 파일은 다음 4개 범주의 테스트를 포함합니다.

#### 8.2.1. `TestMapOracleType_<Dialect>`

table-driven 테스트로 주요 Oracle 타입 매핑을 검증합니다.

```go
func TestMapOracleType_MySQL(t *testing.T) {
    d := &MySQLDialect{}
    tests := []struct {
        name      string
        oraType   string
        precision int
        scale     int
        want      string
    }{
        {"VARCHAR2 with precision", "VARCHAR2", 100, 0, "VARCHAR(100)"},
        {"VARCHAR2 over limit",    "VARCHAR2", 20000, 0, "LONGTEXT"},
        {"VARCHAR2 no precision",  "VARCHAR2", 0, 0, "VARCHAR(255)"},
        {"CHAR with precision",    "CHAR", 10, 0, "CHAR(10)"},
        {"CHAR no precision",      "CHAR", 0, 0, "CHAR(255)"},
        {"NUMBER with scale",      "NUMBER", 10, 2, "DECIMAL(10, 2)"},
        {"NUMBER p<=4",            "NUMBER", 3, 0, "SMALLINT"},
        {"NUMBER p<=9",            "NUMBER", 8, 0, "INT"},
        {"NUMBER p>9",             "NUMBER", 15, 0, "BIGINT"},
        {"NUMBER no precision",    "NUMBER", 0, 0, "DOUBLE"},
        {"DATE",                   "DATE", 0, 0, "DATETIME"},
        {"CLOB",                   "CLOB", 0, 0, "LONGTEXT"},
        {"BLOB",                   "BLOB", 0, 0, "LONGBLOB"},
        {"FLOAT",                  "FLOAT", 0, 0, "DOUBLE"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := d.MapOracleType(tt.oraType, tt.precision, tt.scale)
            if got != tt.want {
                t.Errorf("MapOracleType(%q, %d, %d) = %q, want %q",
                    tt.oraType, tt.precision, tt.scale, got, tt.want)
            }
        })
    }
}
```

MSSQL 테스트에는 v7 수정 항목(NVARCHAR precision, NUMBER→NUMERIC)에 대한 케이스를 반드시 포함합니다.

```go
// MSSQL 특수 케이스
{"VARCHAR2 precision<=4000", "VARCHAR2", 500, 0, "NVARCHAR(500)"},
{"VARCHAR2 precision>4000",  "VARCHAR2", 8000, 0, "NVARCHAR(MAX)"},
{"NUMBER no precision",      "NUMBER", 0, 0, "NUMERIC"},     // FLOAT이 아닌 NUMERIC
```

#### 8.2.2. `TestCreateTableDDL_<Dialect>`

스키마 포함/미포함, NOT NULL 처리를 검증합니다.

```go
func TestCreateTableDDL_MSSQL(t *testing.T) {
    d := &MSSQLDialect{}
    cols := []ColumnDef{
        {Name: "ID", Type: "NUMBER", Precision: sql.NullInt64{Int64: 9, Valid: true}, Nullable: "N"},
        {Name: "NAME", Type: "VARCHAR2", Precision: sql.NullInt64{Int64: 100, Valid: true}, Nullable: "Y"},
    }

    t.Run("with schema", func(t *testing.T) {
        ddl := d.CreateTableDDL("employees", "hr", cols)

        // TABLE_SCHEMA 조건 포함 검증 (v7 수정)
        if !strings.Contains(ddl, "TABLE_SCHEMA = 'hr'") {
            t.Error("expected TABLE_SCHEMA condition in DDL")
        }
        if !strings.Contains(ddl, "TABLE_NAME = 'employees'") {
            t.Error("expected TABLE_NAME condition in DDL")
        }
    })

    t.Run("without schema uses dbo", func(t *testing.T) {
        ddl := d.CreateTableDDL("employees", "", cols)

        // dbo 기본 스키마 검증 (v7 수정)
        if !strings.Contains(ddl, "TABLE_SCHEMA = 'dbo'") {
            t.Error("expected TABLE_SCHEMA = 'dbo' as default")
        }
    })
}
```

#### 8.2.3. `TestCreateIndexDDL_<Dialect>`

일반/UNIQUE/PK 인덱스 생성을 검증하며, MSSQL은 `sys.objects` JOIN 조건을 포함합니다.

```go
func TestCreateIndexDDL_MSSQL(t *testing.T) {
    d := &MSSQLDialect{}

    t.Run("regular index with object_id filter", func(t *testing.T) {
        idx := IndexMetadata{
            Name:       "idx_emp_name",
            Uniqueness: "NONUNIQUE",
            Columns:    []IndexColumn{{Name: "name", Position: 1}},
        }
        ddl := d.CreateIndexDDL(idx, "employees", "hr")

        // v7 수정: sys.objects JOIN 포함 검증
        if !strings.Contains(ddl, "sys.objects o ON i.object_id = o.object_id") {
            t.Error("expected object_id join in DDL")
        }
        if !strings.Contains(ddl, "o.name = 'employees'") {
            t.Error("expected table name filter in DDL")
        }
        if !strings.Contains(ddl, "SCHEMA_NAME(o.schema_id) = 'hr'") {
            t.Error("expected schema filter in DDL")
        }
    })

    t.Run("primary key", func(t *testing.T) {
        idx := IndexMetadata{
            Name:    "pk_emp",
            IsPK:    true,
            Columns: []IndexColumn{{Name: "id", Position: 1}},
        }
        ddl := d.CreateIndexDDL(idx, "employees", "hr")

        if !strings.Contains(ddl, "ALTER TABLE") {
            t.Error("PK should use ALTER TABLE ADD PRIMARY KEY")
        }
    })
}
```

#### 8.2.4. `TestInsertStatement_<Dialect>`

단일 배치, 다중 배치 분할을 검증합니다. MSSQL은 1000행 제한을 포함합니다.

```go
func TestInsertStatement_MSSQL(t *testing.T) {
    d := &MSSQLDialect{}

    t.Run("batch limit 1000", func(t *testing.T) {
        // 1500행, batchSize 2000 요청 → MSSQL은 1000으로 클램핑
        rows := make([][]any, 1500)
        for i := range rows {
            rows[i] = []any{i, fmt.Sprintf("name_%d", i)}
        }

        stmts := d.InsertStatement("emp", "", []string{"id", "name"}, rows, 2000)

        // 1000 + 500 = 2개의 INSERT 문
        if len(stmts) != 2 {
            t.Errorf("expected 2 statements, got %d", len(stmts))
        }
    })
}
```

### 8.3. MariaDB 테스트 특이사항

`MariaDBDialect`는 `MySQLDialect`를 임베딩하므로, 테스트는 `Name()` 반환값 검증과 MySQL 테스트에서 커버하지 않는 MariaDB 고유 동작(있다면)에 집중합니다.

```go
func TestMariaDB_Name(t *testing.T) {
    d := &MariaDBDialect{}
    if d.Name() != "mariadb" {
        t.Errorf("expected 'mariadb', got %q", d.Name())
    }
}

func TestMariaDB_InheritsMySQL(t *testing.T) {
    d := &MariaDBDialect{}
    // MySQL과 동일한 매핑을 확인
    got := d.MapOracleType("VARCHAR2", 100, 0)
    if got != "VARCHAR(100)" {
        t.Errorf("expected VARCHAR(100), got %q", got)
    }
}
```

---

## 9. 하위 호환성 (Backward Compatibility)

### 9.1. CLI 인터페이스

- 모든 기존 CLI 플래그는 변경 없이 동작합니다.
- `--target-db`가 `postgres`이거나 미지정일 때 PostgreSQL 출력 결과는 **일체 변경되지 않습니다**.

### 9.2. API 인터페이스

- `POST /api/migrate` 요청 페이로드의 기존 필드는 변경되지 않습니다.
- WebSocket 메시지에 `warning` 타입이 추가되며, 기존 메시지 타입의 스키마는 변경되지 않습니다.
- `ProgressMsg`에 `message` 필드가 추가되나, JSON에서 `omitempty`로 처리되어 기존 클라이언트에 영향 없음.

### 9.3. 데이터 무결성

- PostgreSQL `MapOracleType`은 `VARCHAR2` → `text`로 매핑하므로 precision 변경의 영향을 받지 않습니다.
- MySQL/MSSQL의 타입 매핑 변경은 **v6 최초 도입 이후 첫 수정**이므로 실질적 하위 호환성 문제는 없습니다.

---

## 10. 테스트 전략 (Testing Strategy)

### 10.1. 단위 테스트

- `internal/dialect/mysql_test.go` — MySQL 타입 매핑 (precision 반영 포함), DDL, INSERT
- `internal/dialect/mariadb_test.go` — MariaDB 이름, MySQL 상속 동작
- `internal/dialect/sqlite_test.go` — SQLite 타입 매핑, DDL (스키마 무시), INSERT
- `internal/dialect/mssql_test.go` — MSSQL 타입 매핑 (precision, NUMBER→NUMERIC), DDL (TABLE_SCHEMA, object_id), INSERT (1000행 분할)

### 10.2. 기존 테스트 영향

- `internal/migration/` 패키지의 기존 integration 테스트는 `ProgressTracker` mock에 의존하므로, `WarningTracker` 추가 시 mock 수정이 필요할 수 있습니다.
- `internal/web/ws/tracker_test.go`에 `Warning()` 메서드 테스트 케이스를 추가합니다.

### 10.3. 검증 기준

- `go test ./...` 전체 통과.
- v7 수정 전후로 `--target-db postgres` 출력이 동일한지 diff로 검증.

---

## 11. 구현 순서 (Implementation Order)

아래 순서는 의존성을 고려한 권장 구현 순서입니다.

| 단계 | 작업 | 의존성 |
|------|------|--------|
| 1 | MySQL `MapOracleType` precision 반영 | 없음 |
| 2 | MSSQL `MapOracleType` precision 반영 + NUMBER→NUMERIC | 없음 |
| 3 | MSSQL `CreateTableDDL` TABLE_SCHEMA 조건 추가 | 없음 |
| 4 | MSSQL `CreateIndexDDL` object_id 조건 추가 | 없음 |
| 5 | 4개 방언 단위 테스트 작성 | 단계 1~4 완료 후 |
| 6 | `ws/tracker.go` MsgWarning + Warning() 추가 | 없음 |
| 7 | `migration.go` WarningTracker 인터페이스 + 호출 지점 | 단계 6 |
| 8 | `migration.go` Dry-Run 대상 DB 연결 검증 | 없음 |
| 9 | `server.go` title 수정 | 없음 |
| 10 | `index.html` DDL 옵션 재구성 + 레이블 수정 + warning 배너 | 단계 6, 8 |
| 11 | 전체 테스트 통과 확인 | 단계 1~10 |
