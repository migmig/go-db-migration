# 기술 사양서 (Technical Specifications) - v20

## 1. 아키텍처 개요

v20은 **세 가지 운영 신뢰성 문제**를 해결한다.

1. **세션 메모리 누수**: `authSessionManager`가 만료 세션을 요청 경로에서만 정리하여 장기 운영 시 메모리가 축적된다.
2. **미인용 SQL 식별자**: `internal/db/db.go`의 `COUNT(*)` 쿼리가 테이블명을 문자열 연결로 구성해 인젝션 위험이 존재한다.
3. **재시도 전략 부재**: `CONNECTION_LOST`·`TIMEOUT` 오류 발생 시 자동 재시도 없이 즉시 실패 처리된다.

핵심 설계 원칙:
- 기존 마이그레이션 엔진(DDL/DML 생성) 변경 없음
- 하위 호환성 유지 — 기본값이 기존 동작과 동일
- 환경 변수로 세션 정책 튜닝 가능

---

## 2. 도메인 모델

### 2.1 세션 구조체 확장

**파일**: `internal/web/server.go`

기존:
```go
type authSession struct {
    UserID     int64
    Username   string
    CreatedAt  time.Time
    LastSeenAt time.Time
}
```

변경 후:
```go
type authSession struct {
    UserID     int64
    Username   string
    CreatedAt  time.Time
    LastSeenAt time.Time
    ExpiresAt  time.Time // 절대 만료 시각 (CreatedAt + absoluteTTL)
}
```

### 2.2 authSessionManager 확장

기존 필드에 추가:
```go
type authSessionManager struct {
    mu          sync.RWMutex
    sessions    map[string]authSession
    idleTTL     time.Duration
    absoluteTTL time.Duration
    maxSessions int       // 신규: 최대 동시 세션 수 (0이면 무제한)
    stopCleanup chan struct{} // 신규: 정리 고루틴 종료 신호
    metrics     *monitoringMetrics
}
```

### 2.3 재시도 이벤트 모델

**파일**: `internal/migration/errors.go` (신규 타입 추가)

```go
// RetryEvent는 재시도 발생 시 이벤트 버스로 전송되는 구조체이다.
type RetryEvent struct {
    TableName   string `json:"table_name"`
    Attempt     int    `json:"attempt"`
    MaxAttempts int    `json:"max_attempts"`
    ErrorMsg    string `json:"error_msg"`
    WaitSeconds int    `json:"wait_seconds"`
}
```

### 2.4 마이그레이션 결과 상태 확장

`partial_success` 상태를 기존 상태 목록에 추가한다.

**파일**: `internal/migration/state.go`

```go
const (
    StatusSuccess        = "success"
    StatusFailed         = "failed"
    StatusPartialSuccess = "partial_success" // 신규: skip_batch 정책으로 일부 배치 누락
)
```

---

## 3. 백엔드 설계

### 3.1 FR-1: 세션 자동 정리

**파일**: `internal/web/server.go`

#### 3.1.1 생성자 변경

`newAuthSessionManager` 파라미터에 `maxSessions int` 추가.
`stopCleanup` 채널 초기화 후 `startCleanupLoop` 고루틴 시작.

```go
func newAuthSessionManager(
    idleTTL, absoluteTTL time.Duration,
    maxSessions int,
    metrics ...*monitoringMetrics,
) *authSessionManager {
    m := &authSessionManager{
        sessions:    make(map[string]authSession),
        idleTTL:     idleTTL,
        absoluteTTL: absoluteTTL,
        maxSessions: maxSessions,
        stopCleanup: make(chan struct{}),
        metrics:     ...,
    }
    go m.startCleanupLoop(5 * time.Minute)
    return m
}
```

#### 3.1.2 정리 고루틴

```go
func (m *authSessionManager) startCleanupLoop(interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            m.purgeExpired()
        case <-m.stopCleanup:
            return
        }
    }
}

func (m *authSessionManager) purgeExpired() {
    now := time.Now()
    m.mu.Lock()
    defer m.mu.Unlock()
    for token, s := range m.sessions {
        if now.After(s.ExpiresAt) || now.Sub(s.LastSeenAt) > m.idleTTL {
            delete(m.sessions, token)
            m.metrics.recordSessionExpired()
        }
    }
}
```

#### 3.1.3 최대 세션 수 제한

`createSession` 호출 시 세션 수 초과 여부 확인.
초과 시 `ExpiresAt` 기준 가장 오래된 세션 1건 삭제 후 신규 세션 생성.

```go
func (m *authSessionManager) createSession(...) (string, authSession, error) {
    // ... 토큰 생성 ...
    m.mu.Lock()
    defer m.mu.Unlock()
    if m.maxSessions > 0 && len(m.sessions) >= m.maxSessions {
        m.evictOldest() // ExpiresAt 기준 최솟값 세션 삭제
    }
    // ExpiresAt = now + absoluteTTL
    s := authSession{..., ExpiresAt: now.Add(m.absoluteTTL)}
    m.sessions[token] = s
    return token, s, nil
}
```

#### 3.1.4 Graceful Shutdown 연동

`RunServerWithAuth` 내 서버 종료 시 `close(authSessions.stopCleanup)` 호출.

#### 3.1.5 환경 변수

| 변수 | 기본값 | 설명 |
|---|---|---|
| `DBM_MAX_SESSIONS` | `100` | 최대 동시 세션 수 (0 = 무제한) |
| `DBM_SESSION_CLEANUP_INTERVAL` | `5m` | 정리 주기 (Go `time.Duration` 형식) |

---

### 3.2 FR-2: SQL 식별자 인용

**파일**: `internal/db/db.go`

#### 3.2.1 문제 위치

```go
// 변경 전 (line 162)
err := d.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+tableName).Scan(&count)

// 변경 전 (line 174)
err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+tableName).Scan(&count)
```

#### 3.2.2 함수 시그니처 변경

`SQLDBCountFn`과 `PGPoolCountFn`에 `quoteIdentifier func(string) string` 파라미터 추가.

```go
func SQLDBCountFn(
    d *sql.DB,
    quoteIdentifier func(string) string,
) func(ctx context.Context, tableName string) (int, error) {
    return func(ctx context.Context, tableName string) (int, error) {
        quoted := quoteIdentifier(tableName)
        err := d.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+quoted).Scan(&count)
        ...
    }
}

func PGPoolCountFn(
    pool PGPool,
    quoteIdentifier func(string) string,
) func(ctx context.Context, tableName string) (int, error) {
    return func(ctx context.Context, tableName string) (int, error) {
        quoted := quoteIdentifier(tableName)
        err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+quoted).Scan(&count)
        ...
    }
}
```

#### 3.2.3 호출부 변경

`SQLDBCountFn`·`PGPoolCountFn` 호출 위치에서 해당 Dialect의 `QuoteIdentifier` 메서드를 전달한다.
`QuoteIdentifier`는 이미 `internal/dialect/dialect.go`의 인터페이스에 선언되어 있으며 모든 구현체에 존재한다.

---

### 3.3 FR-3: 수치형 입력 검증

**파일**: `internal/config/config.go`

`ParseFlags` 완료 후 `validateConfig(cfg *Config) error` 함수 호출.

```go
type configBound struct {
    min, max int
    name     string
}

var numericBounds = []struct {
    field *int
    configBound
}{
    {&cfg.BatchSize,  configBound{1, 100_000, "--batch"}},
    {&cfg.Workers,    configBound{1, 64,      "--workers"}},
    {&cfg.DBMaxOpen,  configBound{1, 1_000,   "--db-max-open"}},
    {&cfg.DBMaxIdle,  configBound{0, 1_000,   "--db-max-idle"}},
}

func validateConfig(cfg *Config) error {
    for _, b := range numericBounds {
        if *b.field < b.min || *b.field > b.max {
            return fmt.Errorf("%s: 값 범위는 %d~%d입니다 (입력값: %d)",
                b.name, b.min, b.max, *b.field)
        }
    }
    return nil
}
```

오류 발생 시 `fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)` 후 `os.Exit(1)`.

---

### 3.4 FR-4: 지수 백오프 재시도

**파일**: `internal/migration/retry.go` (신규)

```go
// RetryConfig는 재시도 정책을 정의한다.
type RetryConfig struct {
    MaxAttempts int           // 기본 3
    InitialWait time.Duration // 기본 1s
    Multiplier  float64       // 기본 2.0
    MaxWait     time.Duration // 기본 30s
}

// DefaultRetryConfig는 기본 재시도 설정을 반환한다.
func DefaultRetryConfig() RetryConfig {
    return RetryConfig{
        MaxAttempts: 3,
        InitialWait: time.Second,
        Multiplier:  2.0,
        MaxWait:     30 * time.Second,
    }
}

// WithRetry는 fn을 RecoverableError에 한해 지수 백오프로 재시도한다.
// eventFn이 nil이 아니면 재시도 시마다 RetryEvent를 전달한다.
func WithRetry(
    ctx context.Context,
    cfg RetryConfig,
    tableName string,
    eventFn func(RetryEvent),
    fn func() error,
) error {
    wait := cfg.InitialWait
    for attempt := 1; ; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }
        var migErr *MigrationError
        if !errors.As(err, &migErr) || !migErr.Recoverable {
            return err
        }
        if attempt >= cfg.MaxAttempts {
            return err
        }
        if eventFn != nil {
            eventFn(RetryEvent{
                TableName:   tableName,
                Attempt:     attempt,
                MaxAttempts: cfg.MaxAttempts,
                ErrorMsg:    err.Error(),
                WaitSeconds: int(wait.Seconds()),
            })
        }
        slog.Warn("migration retry",
            "table", tableName, "attempt", attempt,
            "wait_s", wait.Seconds(), "error", err)
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(wait):
        }
        wait = min(time.Duration(float64(wait)*cfg.Multiplier), cfg.MaxWait)
    }
}
```

**환경 변수**:

| 변수 | 기본값 | 설명 |
|---|---|---|
| `DBM_MAX_RETRIES` | `3` | 최대 재시도 횟수 |
| `DBM_RETRY_INITIAL_WAIT` | `1s` | 초기 대기 시간 |

---

### 3.5 FR-5: 부분 실패 허용 정책 (skip_batch)

**파일**: `internal/config/config.go` — `OnError string` 필드 추가
**파일**: `internal/migration/migration.go` — 배치 오류 처리 경로 분기

```go
// config.go
flag.StringVar(&cfg.OnError, "on-error", "fail_fast",
    "배치 오류 처리 정책: fail_fast | skip_batch")
```

배치 실행 루프 내 오류 처리:

```go
// migration.go (개략적 구조)
if err != nil {
    if cfg.OnError == "skip_batch" {
        skippedBatches++
        slog.Warn("batch skipped", "table", table, "batch", batchNum, "error", err)
        continue
    }
    return err
}
```

완료 후 `skippedBatches > 0`이면 상태를 `partial_success`로 기록.
최종 리포트에 누락 배치 수(`skipped_batches`)와 예상 누락 행 수(`estimated_skipped_rows`) 포함.

---

## 4. API 계약

### 4.1 재시도 이벤트 WebSocket 메시지

기존 WebSocket 이벤트 타입(`bus` 패키지) 에 `retry` 타입 추가.

```json
{
  "type": "retry",
  "payload": {
    "table_name": "EMP",
    "attempt": 1,
    "max_attempts": 3,
    "error_msg": "connection lost: EOF",
    "wait_seconds": 1
  }
}
```

### 4.2 마이그레이션 실행 API 파라미터 추가

`POST /api/migrate` 요청 바디에 `on_error` 필드 추가.

```json
{
  "tables": ["EMP", "DEPT"],
  "on_error": "skip_batch"
}
```

응답의 테이블별 결과에 `skipped_batches` 필드 추가:

```json
{
  "table_name": "EMP",
  "status": "partial_success",
  "skipped_batches": 2,
  "estimated_skipped_rows": 400
}
```

---

## 5. CLI 설계

### 5.1 신규/변경 플래그

| 플래그 | 타입 | 기본값 | 설명 |
|---|---|---|---|
| `--on-error` | string | `fail_fast` | 배치 오류 처리 정책 (`fail_fast` \| `skip_batch`) |

### 5.2 검증 오류 출력 형식

```
[ERROR] --batch: 값 범위는 1~100000입니다 (입력값: 0)
```

### 5.3 재시도 로그 형식

```
[WARN] 재시도 중 (1/3) table=EMP wait=1s error="connection lost: EOF"
```

### 5.4 부분 성공 요약

```
[WARN] partial_success: EMP — 2개 배치 건너뜀 (예상 누락 행: ~400건)
```

---

## 6. UI 설계

### 6.1 재시도 상태 표시

진행 패널(`MigrationProgress` 컴포넌트)에 재시도 상태 행 추가.
- 표시 조건: WebSocket `retry` 이벤트 수신 시
- 형식: `재시도 중 (1/3) — 1초 후 재시작`
- 재시도 성공 시 해당 행 자동 제거

### 6.2 partial_success 뱃지

기존 성공(`success`) 뱃지와 구분:
- `partial_success`: 노란색 뱃지 (`⚠ 부분 완료`)
- 호버 시 툴팁: `N개 배치 건너뜀 (예상 ~M건 누락)`

### 6.3 on-error 정책 선택 UI

마이그레이션 설정 패널에 오류 처리 정책 라디오 버튼 추가:
- `오류 시 중단` (fail_fast, 기본)
- `오류 배치 건너뛰기` (skip_batch)

---

## 7. 로깅/관측성

### 7.1 구조화 로그 필드

| 이벤트 | 필드 |
|---|---|
| 세션 정리 | `cleaned_count`, `remaining_count` |
| 세션 한도 초과 | `evicted_token_prefix`, `current_count`, `max_sessions` |
| 재시도 | `table`, `attempt`, `max_attempts`, `wait_s`, `error` |
| 배치 건너뜀 | `table`, `batch_num`, `estimated_rows`, `error` |

### 7.2 메트릭

기존 `monitoringMetrics`에 카운터 추가:

| 메트릭 | 설명 |
|---|---|
| `session_cleanup_total` | 정리 실행 횟수 |
| `session_evicted_total` | 한도 초과로 삭제된 세션 수 |
| `migration_retry_total{table}` | 테이블별 재시도 횟수 |
| `migration_batch_skipped_total{table}` | 건너뛴 배치 수 |
| `migration_partial_success_total` | partial_success 완료 테이블 수 |

---

## 8. 오류 처리

| 조건 | 처리 |
|---|---|
| `DBM_MAX_SESSIONS` 비정수 값 | 서버 시작 실패 + 오류 로그 |
| `--on-error` 허용되지 않은 값 | `[ERROR]` 출력 후 `os.Exit(1)` |
| 재시도 중 `ctx.Done()` | 즉시 중단, `context canceled` 오류 반환 |
| `skip_batch` + upsert 모드 | 허용; 건너뛴 배치는 upsert 재시도로 처리됨을 로그에 명시 |
| `partial_success` 상태에서 resume | resume 시 건너뛴 배치도 재실행 대상에 포함 |

---

## 9. 테스트 전략

### 9.1 세션 관리 단위 테스트 (`internal/web/server_test.go`)

- `purgeExpired`: 만료 세션 N개 삽입 → 정리 후 잔여 세션 수 검증
- `evictOldest`: maxSessions=2, 3번째 세션 생성 시 가장 오래된 세션 삭제 검증
- `createSession`: ExpiresAt이 `CreatedAt + absoluteTTL`과 일치하는지 검증
- 정리 고루틴: `startCleanupLoop` 호출 후 `stopCleanup` 전송 시 고루틴 정상 종료 검증

### 9.2 입력 검증 단위 테스트 (`internal/config/config_test.go`)

- `batch=0` → 오류 반환
- `workers=65` → 오류 반환
- `db-max-idle=-1` → 오류 반환
- `batch=1`, `batch=100000` → 정상 통과 (경계값)

### 9.3 재시도 단위 테스트 (`internal/migration/retry_test.go` 신규)

- 1회 실패 후 성공 → `attempt=2`에서 완료, 오류 nil 반환
- `MaxAttempts` 회 모두 실패 → 마지막 오류 반환
- `Recoverable=false` 오류 → 재시도 없이 즉시 반환
- `ctx.Cancel()` → `context canceled` 반환

### 9.4 skip_batch 통합 테스트 (`internal/migration/direct_test.go` 확장)

- 3배치 중 2번째에서 오류 발생 + `OnError="skip_batch"` → 1, 3 배치 완료, 상태 `partial_success`
- 3배치 중 2번째에서 오류 발생 + `OnError="fail_fast"` → 오류 반환, 상태 `failed`

### 9.5 SQL 인젝션 방어 테스트 (`internal/db/db_test.go`)

- `SQLDBCountFn`에 `tableName="users; DROP TABLE users--"` 전달 시 quoted 쿼리 문자열 검증
- 각 Dialect `QuoteIdentifier` 경계값 테스트 (특수문자, 예약어, 대소문자 혼합)

---

## 10. 롤아웃 계획

1. **1차 배포 (FR-1, FR-2)**
   - 세션 자동 정리 기본 활성화 (`DBM_MAX_SESSIONS=100`)
   - `SQLDBCountFn`/`PGPoolCountFn` 호출부 QuoteIdentifier 적용
   - 기존 테스트 전량 통과 확인

2. **2차 배포 (FR-3, FR-4)**
   - 입력 검증 활성화 — 기존 기본값이 모두 유효 범위 내이므로 기존 실행에 영향 없음
   - 재시도 기본 활성화 (`DBM_MAX_RETRIES=3`), 기존 fail-fast 동작은 재시도 소진 후 동일

3. **3차 배포 (FR-5)**
   - `--on-error skip_batch` CLI/UI 공개
   - `partial_success` 상태 Web UI 뱃지 활성화

---

## 11. 오픈 이슈

- `evictOldest` 구현 시 O(n) 순회 대신 세션 삽입 순서를 보조 슬라이스로 관리하는 방안 검토 (세션 수 > 500 환경 대비)
- `skip_batch` + PostgreSQL COPY 모드 조합: COPY는 배치 단위 롤백이 지원되지 않으므로 COPY 모드에서 `skip_batch` 허용 여부 정책 확정 필요
- 재시도 `WaitSeconds`를 WebSocket으로 전송 시 클라이언트 카운트다운 타이머 구현 여부
