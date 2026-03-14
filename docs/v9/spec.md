# Technical Specification - v9 (안정성·관측성·데이터 무결성 강화)

## 1. 입력 검증 및 보안 강화 (Security Hardening)

### 1.1. Oracle 식별자 검증 함수

`internal/dialect/` 패키지에 Oracle 소스 쿼리에서 사용하는 식별자 안전 함수를 추가한다.

```go
// internal/dialect/oracle.go (신규)

package dialect

import (
    "fmt"
    "regexp"
)

// oracleIdentifierPattern은 Oracle 식별자 규칙을 따르는 패턴이다.
// 알파벳/밑줄로 시작, 영숫자/밑줄/$/#만 허용, 최대 128자.
var oracleIdentifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_$#]{0,127}$`)

// ValidateOracleIdentifier는 문자열이 유효한 Oracle 식별자인지 검증한다.
func ValidateOracleIdentifier(name string) error {
    if !oracleIdentifierPattern.MatchString(name) {
        return fmt.Errorf("invalid Oracle identifier: %q", name)
    }
    return nil
}

// QuoteOracleIdentifier는 Oracle 식별자를 큰따옴표로 감싸 이스케이프한다.
// 내부 큰따옴표는 두 번 반복하여 이스케이프한다.
func QuoteOracleIdentifier(name string) string {
    escaped := strings.ReplaceAll(name, `"`, `""`)
    return fmt.Sprintf(`"%s"`, escaped)
}
```

### 1.2. 테이블명 검증 적용

**`internal/web/server.go`** — `validateMigrationRequest()`에 테이블명 검증을 추가한다:

```go
func validateMigrationRequest(req *startMigrationRequest) error {
    // 기존 검증 로직 유지...

    // 테이블명 검증 추가
    for _, table := range req.Tables {
        if err := dialect.ValidateOracleIdentifier(table); err != nil {
            return fmt.Errorf("invalid table name %q: %w", table, err)
        }
    }
    // OracleOwner 검증
    if req.OracleOwner != "" {
        if err := dialect.ValidateOracleIdentifier(req.OracleOwner); err != nil {
            return fmt.Errorf("invalid oracle owner %q: %w", req.OracleOwner, err)
        }
    }
    return nil
}
```

**`internal/migration/migration.go`** — 소스 쿼리에서 식별자 사용 시 `QuoteOracleIdentifier()` 적용:

```go
// 기존
query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
query := fmt.Sprintf("SELECT * FROM %s", tableName)

// 변경
quotedTable := dialect.QuoteOracleIdentifier(tableName)
query := fmt.Sprintf("SELECT COUNT(*) FROM %s", quotedTable)
query := fmt.Sprintf("SELECT * FROM %s", quotedTable)
```

적용 대상 함수:
- `Run()` — dry-run COUNT 쿼리 (migration.go:141)
- `MigrateTable()` — COUNT 쿼리 (migration.go:291, 300)
- `MigrateTableDirect()` — SELECT 쿼리 (migration.go:410)
- `MigrateTableToFile()` — SELECT 쿼리 (migration.go:687)

---

## 2. 구조화 에러 시스템 (Structured Error System)

### 2.1. 에러 타입 정의

`internal/migration/errors.go` (신규):

```go
package migration

import "fmt"

// ErrorCategory는 마이그레이션 에러의 분류 코드이다.
type ErrorCategory string

const (
    ErrTypeMismatch    ErrorCategory = "TYPE_MISMATCH"
    ErrNullViolation   ErrorCategory = "NULL_VIOLATION"
    ErrFKViolation     ErrorCategory = "FK_VIOLATION"
    ErrConnectionLost  ErrorCategory = "CONNECTION_LOST"
    ErrTimeout         ErrorCategory = "TIMEOUT"
    ErrPermissionDenied ErrorCategory = "PERMISSION_DENIED"
    ErrUnknown         ErrorCategory = "UNKNOWN"
)

// MigrationError는 마이그레이션 과정에서 발생하는 구조화된 에러이다.
type MigrationError struct {
    Table      string
    Phase      string        // "ddl", "data", "index", "constraint", "validation"
    Category   ErrorCategory
    BatchNum   int           // 1-based 배치 번호 (data phase에서만 유효)
    RowOffset  int           // 전체 행 기준 오프셋
    Column     string        // 문제 컬럼 (파악 가능한 경우)
    RootCause  error
    Suggestion string
    Recoverable bool
}

func (e *MigrationError) Error() string {
    msg := fmt.Sprintf("[%s] %s (table=%s, phase=%s", e.Category, e.RootCause, e.Table, e.Phase)
    if e.BatchNum > 0 {
        msg += fmt.Sprintf(", batch=%d", e.BatchNum)
    }
    if e.RowOffset > 0 {
        msg += fmt.Sprintf(", row=%d", e.RowOffset)
    }
    if e.Column != "" {
        msg += fmt.Sprintf(", column=%s", e.Column)
    }
    msg += ")"
    return msg
}

func (e *MigrationError) Unwrap() error {
    return e.RootCause
}

// classifyError는 DB 드라이버 에러 메시지를 분석하여 ErrorCategory를 결정한다.
func classifyError(err error) ErrorCategory {
    msg := err.Error()
    switch {
    case containsAny(msg, "data type", "type mismatch", "incompatible", "too long", "overflow"):
        return ErrTypeMismatch
    case containsAny(msg, "null", "NOT NULL", "cannot insert NULL"):
        return ErrNullViolation
    case containsAny(msg, "foreign key", "referential", "REFERENCES"):
        return ErrFKViolation
    case containsAny(msg, "connection", "reset", "broken pipe", "EOF", "refused"):
        return ErrConnectionLost
    case containsAny(msg, "timeout", "deadline"):
        return ErrTimeout
    case containsAny(msg, "permission", "denied", "privilege", "ORA-01031"):
        return ErrPermissionDenied
    default:
        return ErrUnknown
    }
}
```

### 2.2. MigrateTableDirect에 에러 컨텍스트 적용

`internal/migration/migration.go` — 배치 INSERT 실패 시:

```go
// 기존
return fmt.Errorf("failed to execute batch insert: %v\nstmt: %s", err, stmt)

// 변경
return &MigrationError{
    Table:       tableName,
    Phase:       "data",
    Category:    classifyError(err),
    BatchNum:    batchNum,
    RowOffset:   rowCount,
    RootCause:   err,
    Suggestion:  suggestFix(classifyError(err), dia.Name()),
    Recoverable: classifyError(err) != ErrConnectionLost,
}
```

DDL 실패, 인덱스 실패, 제약조건 실패에도 동일한 패턴을 적용한다. Phase 값은 각각 `"ddl"`, `"index"`, `"constraint"`로 설정한다.

### 2.3. WebSocket 에러 메시지 확장

**`internal/web/ws/tracker.go`** — `ProgressMsg` 구조체 확장:

```go
type ProgressMsg struct {
    // 기존 필드 유지...
    Type         MsgType `json:"type"`
    Table        string  `json:"table,omitempty"`
    Count        int     `json:"count,omitempty"`
    Total        int     `json:"total,omitempty"`
    ErrorMsg     string  `json:"error,omitempty"`
    Message      string  `json:"message,omitempty"`
    ZipFileID    string  `json:"zip_file_id,omitempty"`
    ConnectionOk bool    `json:"connection_ok,omitempty"`
    Object       string  `json:"object,omitempty"`
    ObjectName   string  `json:"object_name,omitempty"`
    Status       string  `json:"status,omitempty"`

    // v9 추가 필드
    Phase       string `json:"phase,omitempty"`
    Category    string `json:"category,omitempty"`
    Suggestion  string `json:"suggestion,omitempty"`
    Recoverable *bool  `json:"recoverable,omitempty"`
}
```

**`WebSocketTracker.Error()` 메서드 확장** — `MigrationError` 타입 체크 후 상세 필드 전송:

```go
func (t *WebSocketTracker) Error(table string, err error) {
    t.mu.Lock()
    delete(t.states, table)
    t.mu.Unlock()

    msg := ProgressMsg{
        Type:     MsgError,
        Table:    table,
        ErrorMsg: err.Error(),
    }

    // MigrationError인 경우 상세 필드 추가
    var migErr *MigrationError
    if errors.As(err, &migErr) {
        msg.Phase = migErr.Phase
        msg.Category = string(migErr.Category)
        msg.Suggestion = migErr.Suggestion
        recoverable := migErr.Recoverable
        msg.Recoverable = &recoverable
    }

    t.broadcast(msg)
}
```

> **주의**: `ws` 패키지에서 `migration.MigrationError`를 직접 임포트하면 순환 의존이 발생한다. 이를 방지하기 위해 `Error()` 메서드는 인터페이스 기반으로 상세 필드를 추출한다:

```go
// internal/migration/errors.go에 인터페이스 추가
type DetailedError interface {
    error
    ErrorPhase() string
    ErrorCategory() string
    ErrorSuggestion() string
    IsRecoverable() bool
}

func (e *MigrationError) ErrorPhase() string      { return e.Phase }
func (e *MigrationError) ErrorCategory() string    { return string(e.Category) }
func (e *MigrationError) ErrorSuggestion() string  { return e.Suggestion }
func (e *MigrationError) IsRecoverable() bool      { return e.Recoverable }
```

```go
// ws/tracker.go에서는 인터페이스로 접근
type DetailedError interface {
    ErrorPhase() string
    ErrorCategory() string
    ErrorSuggestion() string
    IsRecoverable() bool
}

func (t *WebSocketTracker) Error(table string, err error) {
    // ...
    if de, ok := err.(DetailedError); ok {
        msg.Phase = de.ErrorPhase()
        msg.Category = de.ErrorCategory()
        msg.Suggestion = de.ErrorSuggestion()
        recoverable := de.IsRecoverable()
        msg.Recoverable = &recoverable
    }
    t.broadcast(msg)
}
```

---

## 3. PostgreSQL COPY 모드 개선 (Batched COPY)

### 3.1. Config 플래그 추가

**`internal/config/config.go`**:

```go
type Config struct {
    // 기존 필드 유지...

    // v9 flags
    Validate  bool
    CopyBatch int  // 0이면 기존 단일 COPY 모드 유지
}
```

```go
// ParseFlags()에 추가
flag.BoolVar(&cfg.Validate, "validate", false, "마이그레이션 후 소스-타겟 데이터 검증 수행")
flag.IntVar(&cfg.CopyBatch, "copy-batch", 10000, "PostgreSQL COPY 모드 배치 크기 (0: 단일 COPY)")
```

### 3.2. MigrateTableDirect PostgreSQL 경로 변경

`internal/migration/migration.go` — `MigrateTableDirect()` 내 `pgPool != nil` 분기:

```go
if pgPool != nil {
    if cfg.CopyBatch <= 0 {
        // 기존 단일 COPY 모드 (v8 동작 유지)
        // ... 기존 코드 그대로 ...
    } else {
        // v9: 배치 분할 COPY 모드
        err = migrateTablePgBatchCopy(dbConn, pgPool, tableName, cfg, tracker, mState)
    }
}
```

**신규 함수 `migrateTablePgBatchCopy()`**:

```go
func migrateTablePgBatchCopy(
    dbConn *sql.DB,
    pgPool db.PGPool,
    tableName string,
    cfg *config.Config,
    tracker ProgressTracker,
    mState *MigrationState,
) error {
    tState := mState.GetState(tableName)
    offset := tState.Offset
    batchSize := cfg.CopyBatch
    quotedTable := dialect.QuoteOracleIdentifier(tableName)

    for {
        // Oracle에서 배치 단위 조회
        query := fmt.Sprintf(
            "SELECT * FROM %s OFFSET %d ROWS FETCH NEXT %d ROWS ONLY",
            quotedTable, offset, batchSize,
        )
        rows, err := dbConn.Query(query)
        if err != nil {
            return &MigrationError{
                Table: tableName, Phase: "data",
                Category: classifyError(err), RootCause: err,
            }
        }

        cols, _ := rows.Columns()
        source := &oracleCopySource{rows: rows, cols: cols}

        // 배치별 독립 트랜잭션
        ctx := context.Background()
        tx, err := pgPool.Begin(ctx)
        if err != nil {
            rows.Close()
            return &MigrationError{
                Table: tableName, Phase: "data",
                Category: ErrConnectionLost, RootCause: err,
            }
        }

        n, err := tx.CopyFrom(ctx, pgx.Identifier{cfg.Schema, tableName}, cols, source)
        rows.Close()

        if err != nil {
            tx.Rollback(ctx)
            return &MigrationError{
                Table: tableName, Phase: "data",
                Category: classifyError(err), RootCause: err,
                RowOffset: offset, BatchNum: (offset / batchSize) + 1,
            }
        }

        if err := tx.Commit(ctx); err != nil {
            return &MigrationError{
                Table: tableName, Phase: "data",
                Category: classifyError(err), RootCause: err,
            }
        }

        offset += int(n)
        mState.UpdateOffset(tableName, offset)

        if tracker != nil {
            tracker.Update(tableName, offset)
        }

        // n < batchSize이면 마지막 배치 — 루프 종료
        if int(n) < batchSize {
            break
        }
    }

    slog.Info("batched COPY migration finished", "table", tableName, "rows", offset)
    return nil
}
```

### 3.3. Web UI 연동

**`internal/web/server.go`** — `startMigrationRequest`에 필드 추가:

```go
type startMigrationRequest struct {
    // 기존 필드 유지...

    // v9 추가 필드
    Validate  bool `json:"validate"`
    CopyBatch int  `json:"copyBatch"`
}
```

Config 생성 시 매핑:
```go
cfg := &config.Config{
    // 기존 필드 유지...
    Validate:  req.Validate,
    CopyBatch: req.CopyBatch,
}
```

---

## 4. 감사 로그 및 마이그레이션 리포트 (Audit & Report)

### 4.1. 리포트 구조체

`internal/migration/report.go` (신규):

```go
package migration

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"
)

type TableReport struct {
    Name          string        `json:"name"`
    RowCount      int           `json:"row_count"`
    Duration      time.Duration `json:"duration_ns"`
    DurationHuman string        `json:"duration"`
    DDLExecuted   bool          `json:"ddl_executed"`
    Status        string        `json:"status"` // "ok", "error", "skipped"
    Errors        []string      `json:"errors,omitempty"`
}

type MigrationReport struct {
    JobID        string            `json:"job_id"`
    StartedAt    time.Time         `json:"started_at"`
    FinishedAt   time.Time         `json:"finished_at"`
    DurationHuman string           `json:"duration"`
    SourceURL    string            `json:"source_url"`  // 비밀번호 마스킹
    TargetDB     string            `json:"target_db"`
    TargetURL    string            `json:"target_url"`  // 비밀번호 마스킹
    Tables       []TableReport     `json:"tables"`
    TotalRows    int               `json:"total_rows"`
    SuccessCount int               `json:"success_count"`
    ErrorCount   int               `json:"error_count"`
    mu           sync.Mutex
}

func NewMigrationReport(jobID, sourceURL, targetDB, targetURL string) *MigrationReport {
    return &MigrationReport{
        JobID:     jobID,
        StartedAt: time.Now(),
        SourceURL: maskPassword(sourceURL),
        TargetDB:  targetDB,
        TargetURL: maskPassword(targetURL),
    }
}

// StartTable은 테이블 마이그레이션 시작을 기록하고 종료 시 호출할 콜백을 반환한다.
func (r *MigrationReport) StartTable(name string, withDDL bool) func(rowCount int, err error) {
    start := time.Now()
    return func(rowCount int, err error) {
        elapsed := time.Since(start)
        tr := TableReport{
            Name:          name,
            RowCount:      rowCount,
            Duration:      elapsed,
            DurationHuman: formatDuration(elapsed),
            DDLExecuted:   withDDL,
        }
        if err != nil {
            tr.Status = "error"
            tr.Errors = append(tr.Errors, err.Error())
        } else {
            tr.Status = "ok"
        }

        r.mu.Lock()
        r.Tables = append(r.Tables, tr)
        r.TotalRows += rowCount
        if err != nil {
            r.ErrorCount++
        } else {
            r.SuccessCount++
        }
        r.mu.Unlock()
    }
}

// Finalize는 리포트를 마무리하고 JSON 파일로 저장한다.
func (r *MigrationReport) Finalize() error {
    r.FinishedAt = time.Now()
    r.DurationHuman = formatDuration(r.FinishedAt.Sub(r.StartedAt))

    dir := ".migration_state"
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    path := filepath.Join(dir, fmt.Sprintf("%s_report.json", r.JobID))
    data, err := json.MarshalIndent(r, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}

// PrintSummary는 CLI에서 요약 테이블을 표준 출력에 출력한다.
func (r *MigrationReport) PrintSummary() {
    // 최대 컬럼 너비 계산
    maxName := 5 // "Table" 헤더 최소 길이
    for _, t := range r.Tables {
        if len(t.Name) > maxName {
            maxName = len(t.Name)
        }
    }

    border := fmt.Sprintf("┌─%s─┬──────────┬──────────┬─────────┐", strings.Repeat("─", maxName))
    header := fmt.Sprintf("│ %-*s │ %-8s │ %-8s │ %-7s │", maxName, "Table", "Rows", "Duration", "Status")
    sep    := fmt.Sprintf("├─%s─┼──────────┼──────────┼─────────┤", strings.Repeat("─", maxName))
    footer := fmt.Sprintf("└─%s─┴──────────┴──────────┴─────────┘", strings.Repeat("─", maxName))

    fmt.Println(border)
    fmt.Println(header)
    fmt.Println(sep)
    for _, t := range r.Tables {
        status := t.Status
        if status == "ok" {
            status = "OK"
        } else {
            status = "ERROR"
        }
        fmt.Printf("│ %-*s │ %8s │ %8s │ %-7s │\n",
            maxName, t.Name, formatCount(t.RowCount), t.DurationHuman, status)
    }
    fmt.Println(footer)
    fmt.Printf("Total: %s rows, %d ok, %d errors, %s elapsed\n",
        formatCount(r.TotalRows), r.SuccessCount, r.ErrorCount, r.DurationHuman)
}

// maskPassword는 URL에서 비밀번호를 "***"로 치환한다.
func maskPassword(url string) string {
    // "user:password@" 패턴에서 password를 마스킹
    // oracle://user:pass@host → oracle://user:***@host
    // postgres://user:pass@host → postgres://user:***@host
    // 간단한 패턴 매칭으로 처리
    // ... 구현 ...
}

func formatDuration(d time.Duration) string { /* ... */ }
func formatCount(n int) string              { /* ... */ }
```

### 4.2. Run() 함수에 리포트 통합

`internal/migration/migration.go` — `Run()` 함수 시작부에 리포트 생성, 종료부에 finalize:

```go
func Run(dbConn *sql.DB, targetDB *sql.DB, pgPool db.PGPool, dia dialect.Dialect, cfg *config.Config, tracker ProgressTracker) error {
    // ... 기존 초기화 ...

    report := NewMigrationReport(jobID, cfg.OracleURL, cfg.TargetDB, cfg.TargetURL)

    // ... dry-run 분기 (리포트 없이 기존 동작) ...

    // worker 호출 시 report 전달
    go worker(w, dbConn, targetDB, pgPool, dia, jobs, &wg, mainBuf, cfg, &outMutex, tracker, mState, report)

    // ... wg.Wait() ...
    // ... constraint post-processing ...

    // 검증 단계 (cfg.Validate일 때만)
    if cfg.Validate && (pgPool != nil || targetDB != nil) {
        runValidation(dbConn, targetDB, pgPool, dia, cfg, tracker, report)
    }

    // 리포트 저장 및 출력
    report.Finalize()
    report.PrintSummary()

    return nil
}
```

**worker 함수에서 리포트 기록**:

```go
func worker(id int, ..., report *MigrationReport) {
    defer wg.Done()
    for j := range jobs {
        finishTable := report.StartTable(j.tableName, cfg.WithDDL)
        err := MigrateTable(...)
        var rowCount int
        // rowCount는 MigrateTable의 반환값으로 변경 필요 (아래 참고)
        finishTable(rowCount, err)
        // ... 기존 에러 핸들링 ...
    }
}
```

> **MigrateTable 시그니처 변경**: 리포트 기록을 위해 처리된 행 수를 반환해야 한다.
> ```go
> // 기존
> func MigrateTable(...) error
> // 변경
> func MigrateTable(...) (int, error)
> ```
> 반환값: `(rowCount, nil)` 또는 `(partialRowCount, err)`

### 4.3. Web UI 리포트 다운로드

**`internal/web/server.go`** — 새 엔드포인트:

```go
api.GET("/download/report/:id", downloadReport)
```

```go
func downloadReport(c *gin.Context) {
    id := filepath.Base(c.Param("id"))
    reportPath := filepath.Join(".migration_state", id+"_report.json")
    if _, err := os.Stat(reportPath); os.IsNotExist(err) {
        c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
        return
    }
    c.Header("Content-Disposition", "attachment; filename="+id+"_report.json")
    c.Header("Content-Type", "application/json")
    c.File(reportPath)
}
```

**WebSocket 프로토콜** — `all_done` 메시지에 리포트 요약 포함:

```go
// tracker.AllDone 확장
func (t *WebSocketTracker) AllDone(zipFileID string, report *ReportSummary) {
    msg := ProgressMsg{
        Type:      MsgAllDone,
        ZipFileID: zipFileID,
    }
    if report != nil {
        msg.ReportSummary = report
    }
    t.broadcast(msg)
}
```

`ProgressMsg`에 추가:
```go
type ReportSummary struct {
    TotalRows    int    `json:"total_rows"`
    SuccessCount int    `json:"success_count"`
    ErrorCount   int    `json:"error_count"`
    Duration     string `json:"duration"`
    ReportID     string `json:"report_id"`
}

type ProgressMsg struct {
    // 기존 필드...
    ReportSummary *ReportSummary `json:"report_summary,omitempty"`
}
```

> **순환 의존 방지**: `ReportSummary`는 `ws` 패키지에 정의하고, `migration` 패키지의 `MigrationReport`에서 `ToSummary()` 메서드로 변환한다. 또는 `ws` 패키지에 독립적인 DTO를 두고, `server.go`에서 변환한다.

---

## 5. 마이그레이션 후 데이터 검증 (Post-Migration Validation)

### 5.1. ValidationTracker 인터페이스

`internal/migration/migration.go`에 추가:

```go
type ValidationTracker interface {
    ValidationStart(table string)
    ValidationResult(table string, sourceCount, targetCount int, status string, detail string)
}
```

### 5.2. 검증 엔진

`internal/migration/validation.go` (신규):

```go
package migration

import (
    "context"
    "crypto/sha256"
    "database/sql"
    "fmt"
    "log/slog"
    "strings"

    "dbmigrator/internal/config"
    "dbmigrator/internal/db"
    "dbmigrator/internal/dialect"
)

// runValidation은 직접 마이그레이션 후 소스-타겟 데이터를 비교 검증한다.
func runValidation(
    dbConn *sql.DB,
    targetDB *sql.DB,
    pgPool db.PGPool,
    dia dialect.Dialect,
    cfg *config.Config,
    tracker ProgressTracker,
    report *MigrationReport,
) {
    valTracker, hasValTracker := tracker.(ValidationTracker)

    for _, tableName := range cfg.Tables {
        if hasValTracker {
            valTracker.ValidationStart(tableName)
        }

        result := validateTable(dbConn, targetDB, pgPool, dia, tableName, cfg)

        if hasValTracker {
            valTracker.ValidationResult(
                tableName, result.SourceCount, result.TargetCount,
                result.Status, result.Detail,
            )
        }

        slog.Info("validation result",
            "table", tableName,
            "source_count", result.SourceCount,
            "target_count", result.TargetCount,
            "status", result.Status,
        )
    }
}

type ValidationResult struct {
    Table       string `json:"table"`
    SourceCount int    `json:"source_count"`
    TargetCount int    `json:"target_count"`
    Status      string `json:"status"` // "pass", "mismatch", "error"
    Detail      string `json:"detail,omitempty"`
}

func validateTable(
    dbConn *sql.DB,
    targetDB *sql.DB,
    pgPool db.PGPool,
    dia dialect.Dialect,
    tableName string,
    cfg *config.Config,
) ValidationResult {
    result := ValidationResult{Table: tableName}
    quotedSrc := dialect.QuoteOracleIdentifier(tableName)

    // 1. 소스 행 수 조회
    err := dbConn.QueryRow(
        fmt.Sprintf("SELECT COUNT(*) FROM %s", quotedSrc),
    ).Scan(&result.SourceCount)
    if err != nil {
        result.Status = "error"
        result.Detail = "source count query failed: " + err.Error()
        return result
    }

    // 2. 타겟 행 수 조회
    targetTable := dia.QuoteIdentifier(strings.ToLower(tableName))
    if cfg.Schema != "" {
        targetTable = dia.QuoteIdentifier(strings.ToLower(cfg.Schema)) + "." + targetTable
    }

    if pgPool != nil {
        err = pgPool.QueryRow(
            context.Background(),
            fmt.Sprintf("SELECT COUNT(*) FROM %s", targetTable),
        ).Scan(&result.TargetCount)
    } else if targetDB != nil {
        err = targetDB.QueryRow(
            fmt.Sprintf("SELECT COUNT(*) FROM %s", targetTable),
        ).Scan(&result.TargetCount)
    }
    if err != nil {
        result.Status = "error"
        result.Detail = "target count query failed: " + err.Error()
        return result
    }

    // 3. 비교
    if result.SourceCount != result.TargetCount {
        result.Status = "mismatch"
        diff := result.SourceCount - result.TargetCount
        result.Detail = fmt.Sprintf("%d rows difference", diff)
    } else {
        result.Status = "pass"
    }

    return result
}
```

### 5.3. WebSocket Tracker 확장

**`internal/web/ws/tracker.go`** — 새 메시지 타입 및 메서드:

```go
const (
    // 기존 상수 유지...
    MsgValidationStart  MsgType = "validation_start"
    MsgValidationResult MsgType = "validation_result"
)
```

```go
func (t *WebSocketTracker) ValidationStart(table string) {
    t.broadcast(ProgressMsg{
        Type:  MsgValidationStart,
        Table: table,
    })
}

func (t *WebSocketTracker) ValidationResult(table string, sourceCount, targetCount int, status, detail string) {
    t.broadcast(ProgressMsg{
        Type:    MsgValidationResult,
        Table:   table,
        Total:   sourceCount,          // source_count
        Count:   targetCount,          // target_count
        Status:  status,
        Message: detail,
    })
}
```

### 5.4. Web UI 결과 표시

**`templates/index.html`** — Step 4(결과) 영역에 검증 결과 테이블 추가:

```html
<div id="validation-results" class="validation-panel" style="display:none;">
    <h3>데이터 검증 결과</h3>
    <table class="validation-table">
        <thead>
            <tr>
                <th>테이블</th>
                <th>소스 행 수</th>
                <th>타겟 행 수</th>
                <th>상태</th>
                <th>상세</th>
            </tr>
        </thead>
        <tbody id="validation-tbody"></tbody>
    </table>
</div>
```

JavaScript WebSocket 핸들러에 추가:

```javascript
case 'validation_start':
    // 검증 패널 표시, 해당 테이블 행에 스피너 추가
    break;
case 'validation_result':
    // 결과 행 업데이트: pass → 녹색 체크, mismatch → 주황색 경고, error → 빨간색
    break;
```

---

## 6. Dialect 코드 구조 개선 (Refactoring)

### 6.1. 분할 전략

각 dialect 파일을 역할별 3개 파일로 분할한다. **패키지는 변경하지 않는다** (`package dialect` 유지).

| 현재 파일 | 분할 후 |
|-----------|---------|
| `postgres.go` (7,522줄) | `postgres_types.go` — 구조체, `Name()`, `DriverName()`, `QuoteIdentifier()`, `NormalizeURL()`, `MapOracleType()` |
| | `postgres_ddl.go` — `CreateTableDDL()`, `CreateSequenceDDL()`, `CreateIndexDDL()`, `CreateConstraintDDL()` |
| | `postgres_insert.go` — `InsertStatement()`, 값 직렬화 헬퍼 |
| `mysql.go` (7,161줄) | `mysql_types.go`, `mysql_ddl.go`, `mysql_insert.go` (동일 패턴) |
| `mssql.go` (8,650줄) | `mssql_types.go`, `mssql_ddl.go`, `mssql_insert.go` (동일 패턴) |
| `sqlite.go` (5,079줄) | `sqlite_types.go`, `sqlite_ddl.go`, `sqlite_insert.go` (동일 패턴) |
| `mariadb.go` (199줄) | 분할 없음 (얇은 래퍼) |

### 6.2. 분할 절차

1. **기존 테스트 전체 실행** — `go test ./internal/dialect/...` 결과 저장 (기준선)
2. **파일 분할 실행** — 구조체/메서드를 새 파일로 이동, `package dialect` 유지
3. **테스트 재실행** — 기준선과 동일한 결과 확인
4. **컴파일 확인** — `go build ./...` 성공 확인

### 6.3. 분할 원칙

- **공개 API 변경 없음**: `Dialect` 인터페이스, 구조체 이름, 메서드 시그니처 변경 없음
- **내부 헬퍼 함수**: 파일 간 공유가 필요한 경우 동일 패키지이므로 그대로 호출 가능
- **import 변경 없음**: 외부 패키지에서 `dialect.PostgresDialect`로 접근하는 코드 영향 없음

---

## 7. Web UI 변경 요약

### 7.1. Step 2(설정) 추가 항목

| 항목 | 위치 | UI 요소 | 기본값 |
|------|------|---------|--------|
| 마이그레이션 후 검증 | DDL Options 하위 | 체크박스 "Validate after migration" | 미체크 |
| COPY 배치 크기 | Advanced Settings 하위 | 숫자 입력 "COPY batch size" | 10000 |

### 7.2. Step 4(결과) 강화

- **결과 요약 대시보드**: 총 행 수, 소요 시간, 성공/실패 테이블 수, 처리 속도(rows/sec)
- **테이블별 상세**: 각 테이블의 행 수, 소요 시간, 상태를 표 형태로 표시
- **검증 결과 패널**: `--validate` 활성화 시 소스-타겟 행 수 비교 결과 표시
- **리포트 다운로드**: "리포트 JSON 다운로드" 버튼 추가
- **상세 에러 패널**: 에러 발생 시 phase, category, suggestion을 펼침(expandable) 형태로 표시

### 7.3. 에러 메시지 표시 개선

기존 단순 에러 텍스트 → 구조화된 에러 카드:

```
┌ ERROR: ORDERS ────────────────────────────┐
│ Phase: data                                │
│ Category: TYPE_MISMATCH                    │
│ Batch #42, Row offset 41023                │
│                                            │
│ column DESCRIPTION (CLOB → VARCHAR(255))   │
│ exceeds max length                         │
│                                            │
│ 💡 Suggestion: Target column DESCRIPTION   │
│    should be LONGTEXT or TEXT type          │
└────────────────────────────────────────────┘
```

---

## 8. 테스트 전략

### 8.1. 단위 테스트 (신규)

| 파일 | 테스트 내용 |
|------|------------|
| `dialect/oracle_test.go` | `ValidateOracleIdentifier()` — 유효/무효 식별자 패턴 테스트 |
| `migration/errors_test.go` | `MigrationError.Error()` 포맷, `classifyError()` 분류 정확성 |
| `migration/report_test.go` | `MigrationReport` 생성, `StartTable` → 콜백 → `Finalize()`, JSON 직렬화 검증 |
| `migration/validation_test.go` | `validateTable()` — mock DB로 행 수 일치/불일치 시나리오 |

### 8.2. 통합 테스트

| 시나리오 | 검증 항목 |
|---------|----------|
| 배치 분할 COPY | 10,000건 테이블을 batch=2000으로 분할 시 5개 배치로 완료되는지 확인, 중간 체크포인트 저장 확인 |
| 에러 컨텍스트 전파 | 타입 불일치를 유발하는 데이터 삽입 시 `MigrationError`의 필드가 올바르게 채워지는지 확인 |
| 검증 행 수 불일치 | 소스 100건, 타겟에서 일부 삭제 후 검증 시 `mismatch` 상태 반환 확인 |
| SQL Injection 차단 | 테이블명에 `"; DROP TABLE --` 전달 시 `ValidateOracleIdentifier()` 거부 확인 |
| 리포트 생성 | 마이그레이션 완료 후 `.migration_state/{job_id}_report.json` 파일이 올바른 구조로 생성되는지 확인 |
| Dialect 분할 회귀 | 파일 분할 후 기존 전체 테스트 스위트 통과 확인 |

### 8.3. Web UI 수동 테스트

| 시나리오 | 확인 항목 |
|---------|----------|
| 검증 체크박스 활성화 | Step 4에서 검증 결과 테이블 표시 여부 |
| 에러 발생 | 구조화된 에러 카드 렌더링 (phase, category, suggestion) |
| 리포트 다운로드 | JSON 파일 다운로드 및 내용 확인 |
| COPY 배치 설정 | Advanced Settings에서 값 변경 후 progress bar 점진 업데이트 확인 |
