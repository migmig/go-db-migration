# Technical Specification: Web UI CLI 파라미터 통합 (v4)

## 1. 개요 (Introduction)

v4에서는 CLI 전용으로 남아있던 5개 파라미터(`-out`, `-per-table`, `-schema`, `-dry-run`, `-log-json`)를 Web UI에서 제어할 수 있도록 확장합니다. 백엔드 API 구조체 확장, 프론트엔드 UI 컨트롤 추가, Dry-Run 전용 WebSocket 메시지 타입 추가가 핵심 변경사항입니다.

## 2. 변경 파일 및 영향 범위

| 파일 | 변경 유형 | 설명 |
|------|----------|------|
| `internal/web/server.go` | 수정 | 요청 구조체 확장, Config 매핑 로직 수정, 입력값 검증 추가 |
| `internal/web/templates/index.html` | 수정 | 고급 설정 UI 확장, Dry-Run 토글, 조건부 표시, JS 로직 |
| `internal/web/ws/tracker.go` | 수정 | `MsgDryRunResult` 메시지 타입 및 `DryRunResult()` 메서드 추가 |
| `internal/migration/migration.go` | 수정 | Dry-Run 시 WebSocket tracker 연동 |
| `internal/logger/logger.go` | 수정 | 런타임 로그 모드 전환 함수 추가 |

## 3. 상세 설계 (Detailed Design)

### 3.1. 백엔드 API 변경

#### 3.1.1. `startMigrationRequest` 구조체 확장

**파일**: `internal/web/server.go:83`

기존 구조체에 5개 필드를 추가합니다:

```go
type startMigrationRequest struct {
    OracleURL string   `json:"oracleUrl" binding:"required"`
    Username  string   `json:"username" binding:"required"`
    Password  string   `json:"password" binding:"required"`
    Tables    []string `json:"tables" binding:"required"`
    Direct    bool     `json:"direct"`
    PGURL     string   `json:"pgUrl"`
    WithDDL   bool     `json:"withDdl"`
    BatchSize int      `json:"batchSize"`
    Workers   int      `json:"workers"`
    // v4 추가 필드
    OutFile  string `json:"outFile"`
    PerTable bool   `json:"perTable"`
    Schema   string `json:"schema"`
    DryRun   bool   `json:"dryRun"`
    LogJSON  bool   `json:"logJson"`
}
```

**하위 호환성**: 새 필드는 모두 JSON 바인딩 시 zero value가 기본값이므로, 기존 클라이언트가 이 필드를 생략해도 동작에 영향 없음.

#### 3.1.2. 입력값 검증 함수

**파일**: `internal/web/server.go` (새 함수)

```go
func validateMigrationRequest(req *startMigrationRequest) error {
    // outFile: 경로 구분자 차단
    if strings.ContainsAny(req.OutFile, "/\\") {
        return fmt.Errorf("outFile must not contain path separators")
    }
    // schema: SQL 인젝션 방지 (영문, 숫자, 언더스코어만 허용)
    if req.Schema != "" && !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(req.Schema) {
        return fmt.Errorf("schema name contains invalid characters")
    }
    // batchSize, workers: 음수 방지 (0은 기본값 적용으로 허용)
    if req.BatchSize < 0 {
        return fmt.Errorf("batchSize must be non-negative")
    }
    if req.Workers < 0 {
        return fmt.Errorf("workers must be non-negative")
    }
    return nil
}
```

#### 3.1.3. `startMigration` 핸들러 수정

**파일**: `internal/web/server.go:95` (`startMigration` 함수)

변경 포인트:

1. **검증 호출**: `ShouldBindJSON` 이후 `validateMigrationRequest` 호출
2. **기본값 처리**: `outFile`이 빈 문자열이면 `"migration.sql"` 설정
3. **Config 매핑 확장**: 기존 하드코딩된 `PerTable: true`를 `req.PerTable`로 변경
4. **LogJSON 처리**: `req.LogJSON`이 `true`이면 `logger.SetJSONMode(true)` 호출
5. **DryRun 분기**: `req.DryRun`이 `true`이면 Dry-Run 전용 로직 실행

```go
func startMigration(c *gin.Context) {
    var req startMigrationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
        return
    }

    if err := validateMigrationRequest(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    go func() {
        oracleDB, err := db.ConnectOracle(req.OracleURL, req.Username, req.Password)
        if err != nil {
            log.Printf("Failed to connect to Oracle: %v", err)
            tracker.AllDone("")
            return
        }
        defer oracleDB.Close()

        // LogJSON 모드 전환
        if req.LogJSON {
            logger.SetJSONMode(true)
            defer logger.SetJSONMode(false)
        }

        var pgPool db.PGPool
        if req.Direct && req.PGURL != "" {
            pgPool, err = db.ConnectPostgres(req.PGURL)
            if err != nil {
                log.Printf("Failed to connect to Postgres: %v", err)
                tracker.AllDone("")
                return
            }
            defer pgPool.Close()
        }

        workers := req.Workers
        if workers <= 0 {
            workers = 4
        }
        batchSize := req.BatchSize
        if batchSize <= 0 {
            batchSize = 1000
        }
        outFile := req.OutFile
        if outFile == "" {
            outFile = "migration.sql"
        }

        jobID := time.Now().Format("20060102150405")
        outDir := filepath.Join(os.TempDir(), "dbmigrator_"+jobID)
        if !req.Direct && !req.DryRun {
            if err := os.MkdirAll(outDir, 0755); err != nil {
                log.Printf("Failed to create temp directory: %v", err)
                return
            }
        }

        cfg := &config.Config{
            Tables:    req.Tables,
            Parallel:  true,
            Workers:   workers,
            BatchSize: batchSize,
            PerTable:  req.PerTable,
            OutFile:   outFile,
            Schema:    req.Schema,
            DryRun:    req.DryRun,
            OutputDir: outDir,
            PGURL:     req.PGURL,
            WithDDL:   req.WithDDL,
        }

        err = migration.Run(oracleDB, pgPool, cfg, tracker)
        if err != nil {
            log.Printf("Migration failed: %v", err)
            tracker.AllDone("")
        } else if req.DryRun {
            // Dry-Run은 다운로드 없이 완료
            tracker.AllDone("")
        } else if !req.Direct {
            zipFilePath := filepath.Join(os.TempDir(), "migration_"+jobID+".zip")
            if err := ziputil.ZipDirectory(outDir, zipFilePath); err != nil {
                log.Printf("Failed to create zip: %v", err)
                tracker.AllDone("")
            } else {
                tracker.AllDone("migration_" + jobID + ".zip")
            }
        } else {
            tracker.AllDone("")
        }

        if !req.Direct && !req.DryRun {
            os.RemoveAll(outDir)
        }
    }()

    c.JSON(http.StatusOK, gin.H{"message": "Migration started"})
}
```

### 3.2. WebSocket 프로토콜 확장

#### 3.2.1. Dry-Run 결과 메시지 타입

**파일**: `internal/web/ws/tracker.go`

```go
const (
    MsgInit         MsgType = "init"
    MsgUpdate       MsgType = "update"
    MsgDone         MsgType = "done"
    MsgError        MsgType = "error"
    MsgAllDone      MsgType = "all_done"
    MsgDryRunResult MsgType = "dry_run_result"  // NEW
)

type ProgressMsg struct {
    Type         MsgType `json:"type"`
    Table        string  `json:"table,omitempty"`
    Count        int     `json:"count,omitempty"`
    Total        int     `json:"total,omitempty"`
    ErrorMsg     string  `json:"error,omitempty"`
    ZipFileID    string  `json:"zip_file_id,omitempty"`
    ConnectionOk bool    `json:"connection_ok,omitempty"` // NEW: Dry-Run 연결 확인 결과
}
```

#### 3.2.2. `DryRunResult` 메서드 추가

```go
func (t *WebSocketTracker) DryRunResult(table string, totalRows int, connectionOk bool) {
    t.broadcast(ProgressMsg{
        Type:         MsgDryRunResult,
        Table:        table,
        Total:        totalRows,
        ConnectionOk: connectionOk,
    })
}
```

### 3.3. 마이그레이션 로직 수정 (Dry-Run + WebSocket)

**파일**: `internal/migration/migration.go:40-53`

현재 Dry-Run 로직은 `slog`로만 출력합니다. WebSocket tracker가 있을 경우 tracker를 통해 UI에 결과를 전달하도록 수정합니다.

```go
if cfg.DryRun {
    slog.Info("Dry run mode enabled. Verifying connectivity and estimating row counts.")
    for _, table := range cfg.Tables {
        var count int
        err := dbConn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
        if err != nil {
            slog.Error("failed to get row count", "table", table, "error", err)
            if tracker != nil {
                tracker.Error(table, err)
            }
            continue
        }
        slog.Info("table estimation", "table", table, "estimated_rows", count)
        if tracker != nil {
            // DryRunResult는 ProgressTracker 인터페이스에 optional이므로
            // 타입 단언으로 호출
            if drt, ok := tracker.(DryRunTracker); ok {
                drt.DryRunResult(table, count, true)
            }
        }
    }
    slog.Info("Dry run completed successfully.")
    return nil
}
```

#### 3.3.1. DryRunTracker 인터페이스

**파일**: `internal/migration/migration.go` (새 인터페이스)

기존 `ProgressTracker` 인터페이스를 깨뜨리지 않기 위해 별도의 인터페이스를 정의하고 타입 단언으로 호출합니다:

```go
// DryRunTracker는 Dry-Run 모드에서 추가 결과를 전송할 수 있는 tracker입니다.
type DryRunTracker interface {
    DryRunResult(table string, totalRows int, connectionOk bool)
}
```

`WebSocketTracker`는 이 인터페이스를 자동으로 만족합니다 (`DryRunResult` 메서드 추가 후).

### 3.4. 로거 런타임 전환

**파일**: `internal/logger/logger.go`

```go
// SetJSONMode는 런타임에 로그 출력 형식을 전환합니다.
func SetJSONMode(enabled bool) {
    var handler slog.Handler
    if enabled {
        handler = slog.NewJSONHandler(os.Stdout, nil)
    } else {
        handler = slog.NewTextHandler(os.Stdout, nil)
    }
    slog.SetDefault(slog.New(handler))
}
```

### 3.5. 프론트엔드 UI 변경

**파일**: `internal/web/templates/index.html`

#### 3.5.1. 고급 설정 섹션 확장 (Advanced Settings)

기존 `<details>` 요소 내부(line 308-320)를 확장합니다:

```html
<details style="margin-bottom: 1rem; cursor: pointer;">
    <summary style="font-weight: 600; font-size: 0.9rem; color: var(--text-muted); margin-bottom: 0.5rem;">Advanced Settings</summary>
    <div style="margin-top: 0.75rem; padding: 1rem; background: var(--bg-color); border-radius: var(--radius-md);">
        <!-- 기존 필드 -->
        <div style="display: flex; gap: 1rem;">
            <div class="form-group" style="flex: 1; margin-bottom: 0;">
                <label for="batchSize">Batch Size</label>
                <input type="text" id="batchSize" value="1000">
            </div>
            <div class="form-group" style="flex: 1; margin-bottom: 0;">
                <label for="workers">Parallel Workers</label>
                <input type="text" id="workers" value="4">
            </div>
        </div>

        <!-- 구분선 -->
        <hr style="border: none; border-top: 1px solid var(--border-color); margin: 1rem 0;">

        <!-- v4 새 필드: 출력 파일명, PG 스키마 -->
        <div id="file-settings" style="display: flex; gap: 1rem;">
            <div class="form-group" style="flex: 1; margin-bottom: 0;">
                <label for="outFile">Output Filename</label>
                <input type="text" id="outFile" value="migration.sql">
            </div>
            <div class="form-group" style="flex: 1; margin-bottom: 0;">
                <label for="schema">PG Schema</label>
                <input type="text" id="schema" placeholder="public">
            </div>
        </div>

        <!-- v4 새 필드: 체크박스 -->
        <div style="margin-top: 0.75rem;">
            <div id="perTableContainer" class="checkbox-container" style="margin-bottom: 0.5rem;">
                <input type="checkbox" id="perTable" checked>
                <label for="perTable">Per-Table File Output</label>
            </div>
            <div class="checkbox-container" style="margin-bottom: 0;">
                <input type="checkbox" id="logJson">
                <label for="logJson">JSON Logging</label>
            </div>
        </div>
    </div>
</details>
```

#### 3.5.2. Dry-Run 토글 추가

마이그레이션 버튼 직전(line 339 부근)에 Dry-Run 체크박스를 추가합니다:

```html
<div class="checkbox-container" style="margin-bottom: 1rem; margin-top: 1rem;">
    <input type="checkbox" id="dryRun">
    <label for="dryRun">Dry-Run (Verify connectivity & estimate row counts only)</label>
</div>

<button id="btn-migrate" class="btn-success" style="margin-top: 0.5rem;">Start Migration</button>
```

#### 3.5.3. JavaScript 변경사항

##### a) 조건부 표시 로직

`directMigration` 체크박스 이벤트 리스너를 확장합니다:

```javascript
directMigration.addEventListener('change', (e) => {
    const isDirect = e.target.checked;
    pgConfig.style.display = isDirect ? 'block' : 'none';

    // SQL File 전용 컨트롤 토글
    document.getElementById('file-settings').style.display = isDirect ? 'none' : 'flex';
    document.getElementById('perTableContainer').style.display = isDirect ? 'none' : 'flex';
});
```

##### b) Dry-Run 토글 시 버튼 텍스트 변경

```javascript
const dryRunCheckbox = document.getElementById('dryRun');
dryRunCheckbox.addEventListener('change', (e) => {
    btnMigrate.innerText = e.target.checked ? 'Run Verification' : 'Start Migration';
});
```

##### c) 마이그레이션 요청 payload 확장

`btnMigrate` 클릭 이벤트 핸들러 내 `fetch` 호출 부분을 확장합니다:

```javascript
const outFile = document.getElementById('outFile').value || 'migration.sql';
const schema = document.getElementById('schema').value;
const perTable = document.getElementById('perTable').checked;
const dryRun = document.getElementById('dryRun').checked;
const logJson = document.getElementById('logJson').checked;

const res = await fetch('/api/migrate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        oracleUrl,
        username,
        password,
        tables: selectedTables,
        direct: isDirect,
        pgUrl: pgUrl,
        withDdl: withDdl,
        batchSize: batchSize,
        workers: workers,
        outFile: outFile,
        perTable: perTable,
        schema: schema,
        dryRun: dryRun,
        logJson: logJson
    })
});
```

##### d) Dry-Run 결과 WebSocket 메시지 처리

`handleProgressMessage` 함수에 `dry_run_result` 타입 처리를 추가합니다:

```javascript
function handleProgressMessage(msg) {
    const container = document.getElementById('progress-container');

    if (msg.type === 'all_done') {
        currentZipId = msg.zip_file_id;
        if (currentZipId) {
            downloadBtn.style.display = 'flex';
        } else {
            downloadBtn.style.display = 'none';
        }
        btnMigrate.disabled = false;
        btnMigrate.innerText = dryRunCheckbox.checked ? 'Run Verification' : 'Start Migration';
        return;
    }

    // Dry-Run 결과 처리
    if (msg.type === 'dry_run_result') {
        let wrapper = document.getElementById(`prog-${msg.table}`);
        if (!wrapper) {
            wrapper = document.createElement('div');
            wrapper.id = `prog-${msg.table}`;
            wrapper.className = 'progress-container';
            container.appendChild(wrapper);
        }
        const statusIcon = msg.connection_ok
            ? '<svg style="width:14px;height:14px;vertical-align:middle;margin-right:4px;color:var(--success-color)" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>'
            : '<svg style="width:14px;height:14px;vertical-align:middle;margin-right:4px;color:var(--danger-color)" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>';
        wrapper.innerHTML = `
            <div class="progress-header">
                <span class="progress-title">${msg.table}</span>
                <span class="progress-text">
                    ${statusIcon}
                    Estimated rows: <strong>${(msg.total || 0).toLocaleString()}</strong>
                </span>
            </div>
        `;
        return;
    }

    // ... 기존 init/update/done/error 처리 유지
}
```

##### e) Dry-Run 모드 시 다운로드 버튼 숨김 및 진행 라벨 표시

Dry-Run 활성화 시 진행 상황 섹션 제목에 "(Dry-Run)" 접미사를 추가합니다:

```javascript
// btnMigrate 클릭 이벤트 핸들러 내
const progressTitle = document.querySelector('#progress-section h2');
if (dryRun) {
    progressTitle.innerText = '3. Verification Results (Dry-Run)';
} else {
    progressTitle.innerText = '3. Migration Progress';
}
```

## 4. 보안 고려사항

### 4.1. 입력값 검증
- `outFile`: 경로 구분자(`/`, `\`) 포함 시 거부하여 **Path Traversal** 방지
- `schema`: 정규식 `^[a-zA-Z_][a-zA-Z0-9_]*$`로 SQL 식별자만 허용하여 **SQL Injection** 방지
- 기존 `filepath.Base()` 기반 다운로드 경로 검증은 유지

### 4.2. 런타임 로그 모드 전환
- `SetJSONMode`는 글로벌 `slog.Default`를 변경하므로, 동시 요청 시 경합(race) 가능
- 현재 단일 사용자 로컬 환경이므로 허용 가능. 향후 멀티 유저 지원 시 per-request logger로 전환 필요

### 4.3. Dry-Run SQL 인젝션 방지
- Dry-Run의 `SELECT COUNT(*) FROM %s`에서 `tableName`은 `/api/tables` API가 Oracle `ALL_TABLES`에서 조회한 값만 사용되므로 안전
- 추가 방어가 필요하면 테이블명에 대해 `^[a-zA-Z_][a-zA-Z0-9_]*$` 검증 적용 가능

## 5. 테스트 계획

### 5.1. 단위 테스트

| 테스트 케이스 | 파일 | 설명 |
|--------------|------|------|
| `TestValidateMigrationRequest_ValidInput` | `server_test.go` | 정상 입력 통과 |
| `TestValidateMigrationRequest_PathTraversal` | `server_test.go` | `outFile`에 `/` 포함 시 에러 |
| `TestValidateMigrationRequest_InvalidSchema` | `server_test.go` | `schema`에 특수문자 포함 시 에러 |
| `TestDryRunResult_Broadcast` | `tracker_test.go` | `DryRunResult` 메서드가 올바른 JSON 전송 |
| `TestSetJSONMode` | `logger_test.go` | 로그 모드 전환 검증 |

### 5.2. 통합 테스트

| 시나리오 | 설명 |
|---------|------|
| SQL File + PerTable=false | 단일 파일 출력, `outFile` 이름 반영 확인 |
| SQL File + PerTable=true | 테이블별 파일 생성 확인 |
| SQL File + Schema 지정 | INSERT문에 스키마 접두사 포함 확인 |
| Dry-Run 모드 | 파일 생성 없이 row count만 반환 확인 |
| Dry-Run + WebSocket | `dry_run_result` 메시지 수신 확인 |
| 하위 호환성 | 새 필드 없는 요청이 기본값으로 정상 동작 확인 |

### 5.3. 프론트엔드 수동 테스트

| 시나리오 | 검증 항목 |
|---------|----------|
| Direct 모드 전환 | `outFile`, `perTable` 컨트롤 숨김 확인 |
| SQL File 모드 전환 | `outFile`, `perTable` 컨트롤 표시 확인 |
| Dry-Run 토글 | 버튼 텍스트 변경, 결과 표시 형식 확인 |
| Schema 입력 | placeholder `public` 표시, 빈값 허용 확인 |
| JSON Logging 토글 | 서버 로그 형식 변경 확인 (터미널에서) |
