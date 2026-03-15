# Implementation Tasks: Web UI CLI 파라미터 통합 (v4)

## Phase 1: 백엔드 API 구조체 확장 및 검증

### 1-1. `startMigrationRequest` 구조체 확장
- [x] `internal/web/server.go`의 `startMigrationRequest`에 5개 필드 추가
  - `OutFile string json:"outFile"`
  - `PerTable bool json:"perTable"`
  - `Schema string json:"schema"`
  - `DryRun bool json:"dryRun"`
  - `LogJSON bool json:"logJson"`
- [x] `import`에 `regexp`, `strings`, `fmt` 추가 (미포함 시)

### 1-2. 입력값 검증 함수 구현
- [x] `validateMigrationRequest()` 함수 신규 작성
  - `outFile`에 경로 구분자(`/`, `\`) 포함 시 에러 반환
  - `schema`에 SQL 식별자 패턴(`^[a-zA-Z_][a-zA-Z0-9_]*$`) 외 문자 포함 시 에러 반환
  - `batchSize`, `workers` 음수 방지
- [x] `startMigration` 핸들러에서 `ShouldBindJSON` 직후 `validateMigrationRequest` 호출 추가

### 1-3. `startMigration` 핸들러 Config 매핑 수정
- [x] `outFile` 기본값 처리: 빈 문자열이면 `"migration.sql"` 설정
- [x] `PerTable: true` 하드코딩 → `req.PerTable`로 변경
- [x] `cfg`에 `OutFile`, `Schema`, `DryRun` 필드 매핑 추가
- [x] DryRun 시 임시 디렉토리 생성 스킵 로직 추가 (`!req.Direct && !req.DryRun`)
- [x] DryRun 완료 시 ZIP 생성 없이 `tracker.AllDone("")` 호출
- [x] DryRun 시 임시 디렉토리 삭제 스킵 (`!req.Direct && !req.DryRun`)

## Phase 2: WebSocket 프로토콜 확장 (Dry-Run)

### 2-1. Tracker 메시지 타입 추가
- [x] `internal/web/ws/tracker.go`에 `MsgDryRunResult MsgType = "dry_run_result"` 상수 추가
- [x] `ProgressMsg` 구조체에 `ConnectionOk bool json:"connection_ok,omitempty"` 필드 추가

### 2-2. `DryRunResult` 메서드 구현
- [x] `WebSocketTracker`에 `DryRunResult(table string, totalRows int, connectionOk bool)` 메서드 추가
- [x] 해당 메서드에서 `MsgDryRunResult` 타입으로 broadcast

## Phase 3: 마이그레이션 로직 Dry-Run WebSocket 연동

### 3-1. `DryRunTracker` 인터페이스 정의
- [x] `internal/migration/migration.go`에 `DryRunTracker` 인터페이스 추가
  - `DryRunResult(table string, totalRows int, connectionOk bool)`

### 3-2. Dry-Run 로직에 tracker 연동
- [x] 기존 Dry-Run 블록(`migration.go:40-53`)에서 tracker nil 체크 후 타입 단언으로 `DryRunResult` 호출
- [x] 에러 발생 시 `tracker.Error(table, err)` 호출 추가

## Phase 4: 로거 런타임 전환

### 4-1. `SetJSONMode` 함수 구현
- [x] `internal/logger/logger.go`에 `SetJSONMode(enabled bool)` 함수 추가
  - `enabled=true`: `slog.NewJSONHandler` → `slog.SetDefault`
  - `enabled=false`: `slog.NewTextHandler` → `slog.SetDefault`

### 4-2. 핸들러에서 LogJSON 연동
- [x] `startMigration` 고루틴 내에서 `req.LogJSON` true 시 `logger.SetJSONMode(true)` 호출
- [x] `defer logger.SetJSONMode(false)`로 원복 처리

## Phase 5: 프론트엔드 UI 확장

### 5-1. 고급 설정(Advanced Settings) 섹션 확장
- [x] 기존 Batch Size / Workers 아래에 구분선(`<hr>`) 추가
- [x] Output Filename 텍스트 입력 필드 추가 (기본값: `migration.sql`)
- [x] PG Schema 텍스트 입력 필드 추가 (placeholder: `public`)
- [x] Per-Table File Output 체크박스 추가 (기본값: checked)
- [x] JSON Logging 체크박스 추가 (기본값: unchecked)

### 5-2. Dry-Run 토글 UI 추가
- [x] 마이그레이션 시작 버튼 직전에 Dry-Run 체크박스 배치
- [x] 라벨: "Dry-Run (Verify connectivity & estimate row counts only)"

### 5-3. 조건부 표시 JavaScript 로직
- [x] `directMigration` 이벤트 리스너 확장: Direct 모드 시 `outFile`, `perTable` 컨트롤 숨김
- [x] SQL File 모드 전환 시 해당 컨트롤 다시 표시

### 5-4. Dry-Run 토글 JavaScript 로직
- [x] `dryRun` 체크박스 change 이벤트: 버튼 텍스트 "Start Migration" ↔ "Run Verification" 전환

### 5-5. 마이그레이션 요청 payload 확장
- [x] `btnMigrate` 클릭 핸들러에서 5개 새 필드 값 수집
- [x] `fetch('/api/migrate')` body에 `outFile`, `perTable`, `schema`, `dryRun`, `logJson` 추가

### 5-6. Dry-Run WebSocket 메시지 처리
- [x] `handleProgressMessage`에 `dry_run_result` 타입 분기 추가
- [x] Dry-Run 결과 UI 렌더링: 테이블명 + 연결 상태 아이콘 + 예상 row count
- [x] `all_done` 시 Dry-Run 모드면 다운로드 버튼 숨김 유지
- [x] Dry-Run 모드 시 진행 섹션 제목을 "3. Verification Results (Dry-Run)"으로 변경

## Phase 6: 테스트

### 6-1. 단위 테스트 작성
- [x] `TestValidateMigrationRequest_ValidInput` - 정상 입력 통과
- [x] `TestValidateMigrationRequest_PathTraversal` - outFile 경로 구분자 거부
- [x] `TestValidateMigrationRequest_InvalidSchema` - schema 특수문자 거부
- [x] `TestDryRunResult_Broadcast` - DryRunResult WebSocket 메시지 검증
- [x] `TestSetJSONMode` - 로그 모드 전환 검증

### 6-2. 통합 테스트 작성
- [x] SQL File + PerTable=false: 단일 파일 출력 및 outFile 이름 반영
- [x] SQL File + PerTable=true: 테이블별 파일 생성
- [x] SQL File + Schema 지정: INSERT문 스키마 접두사 포함
- [x] Dry-Run 모드: 파일 미생성, row count만 반환
- [x] 하위 호환성: 새 필드 미포함 요청 정상 동작

### 6-3. 프론트엔드 수동 테스트
- [ ] Direct ↔ SQL File 모드 전환 시 컨트롤 표시/숨김 확인
- [ ] Dry-Run 토글 시 버튼 텍스트 및 결과 UI 확인
- [ ] Schema 입력 placeholder 및 빈값 허용 확인
- [ ] JSON Logging 토글 시 서버 로그 형식 변경 확인

## Phase 7: 동기화 및 마무리

### 7-1. 템플릿 동기화
- [x] `internal/web/templates/index.html` 변경사항을 `web/templates/index.html`에 동기화

### 7-2. 빌드 및 검증
- [x] `go build` 성공 확인
- [x] `go test ./...` 전체 테스트 통과 확인
- [x] `go vet ./...` 정적 분석 통과 확인

### 7-3. 문서 정리
- [ ] `docs/v4/prd.md` 최종 확인
- [ ] `docs/v4/spec.md` 최종 확인
- [x] `docs/v4/task.md` 완료 항목 체크
