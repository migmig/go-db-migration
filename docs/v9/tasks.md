# Implementation Tasks - v9 (안정성·관측성·데이터 무결성 강화)

## 1단계: 보안 강화 (Security Hardening)

### 1.1. Oracle 식별자 검증·이스케이프 함수
- [x] `internal/dialect/oracle.go` 신규 생성
  - [x] `oracleIdentifierPattern` 정규식 정의 (`^[A-Za-z_][A-Za-z0-9_$#]{0,127}$`)
  - [x] `ValidateOracleIdentifier(name string) error` 구현
  - [x] `QuoteOracleIdentifier(name string) string` 구현 (큰따옴표 감싸기 + 내부 `"` 이스케이프)
- [x] `internal/dialect/oracle_test.go` 신규 생성
  - [x] 유효 식별자 테스트: `USERS`, `MY_TABLE_1`, `SYS_C00123`, `TABLE$1`, `T#EST`
  - [x] 무효 식별자 테스트: `1TABLE`, `"DROP TABLE`, `; SELECT`, 빈 문자열, 129자 초과
  - [x] `QuoteOracleIdentifier` 출력 검증: `USERS` → `"USERS"`, `MY"TABLE` → `"MY""TABLE"`

### 1.2. Web API 테이블명 검증 적용
- [x] `internal/web/server.go` — `validateMigrationRequest()`에 테이블명 검증 추가
  - [x] `req.Tables` 각 항목에 `dialect.ValidateOracleIdentifier()` 호출
  - [x] `req.OracleOwner`에 `dialect.ValidateOracleIdentifier()` 호출
  - [x] 검증 실패 시 HTTP 400 + 구체적 에러 메시지 반환
- [x] `internal/web/server_test.go` — 검증 테스트 추가
  - [x] 악의적 테이블명(`"; DROP TABLE --`) 요청 시 400 반환 확인

### 1.3. migration.go 쿼리 식별자 이스케이프 적용
- [x] `Run()` — dry-run `SELECT COUNT(*)` 쿼리에 `QuoteOracleIdentifier()` 적용 (기존 L141)
- [x] `MigrateTable()` — `SELECT COUNT(*)` 쿼리에 적용 (기존 L291, L300)
- [x] `MigrateTableDirect()` — `SELECT * FROM` 쿼리에 적용 (기존 L410)
- [x] `MigrateTableToFile()` — `SELECT * FROM` 쿼리에 적용 (기존 L687)

---

## 2단계: 구조화 에러 시스템 (Structured Errors)

### 2.1. 에러 타입 정의
- [x] `internal/migration/errors.go` 신규 생성
  - [x] `ErrorCategory` 타입 및 상수 정의: `TYPE_MISMATCH`, `NULL_VIOLATION`, `UNIQUE_VIOLATION`, `FK_VIOLATION`, `CONNECTION_LOST`, `TIMEOUT`, `PERMISSION_DENIED`, `UNKNOWN`
  - [x] `MigrationError` 구조체 정의 (Table, Phase, Category, BatchNum, RowOffset, Column, RootCause, Suggestion, Recoverable)
  - [x] `Error() string` 메서드 — 구조화된 에러 문자열 포맷
  - [x] `Unwrap() error` 메서드 — `errors.As`/`errors.Is` 호환
  - [x] `DetailedError` 인터페이스 정의: `ErrorPhase()`, `ErrorCategory()`, `ErrorSuggestion()`, `IsRecoverable()`
  - [x] `MigrationError`에 `DetailedError` 인터페이스 메서드 구현
  - [x] `classifyError(err error) ErrorCategory` — 에러 메시지 기반 분류 함수
  - [x] `containsAny(s string, substrs ...string) bool` — 헬퍼 함수
  - [x] `suggestFix(category ErrorCategory, dialectName string) string` — 카테고리별 복구 제안 메시지 생성

### 2.2. 에러 타입 테스트
- [x] `internal/migration/errors_test.go` 신규 생성
  - [x] `MigrationError.Error()` 포맷 검증 — 모든 필드 포함/일부 필드 누락 케이스
  - [x] `classifyError()` — 각 카테고리별 에러 메시지 매칭 검증
  - [x] `suggestFix()` — 카테고리+dialect 조합별 제안 메시지 확인
  - [x] `errors.As()` / `errors.Is()` 호환성 확인

### 2.3. migration.go에 MigrationError 적용
- [x] `MigrateTableDirect()` — 배치 INSERT 실패 시 `MigrationError` 반환으로 교체
  - [x] `batchNum` 카운터 변수 추가 (기존 루프에 1-based 카운터)
  - [x] DDL 실패: Phase `"ddl"`, 해당 에러 분류
  - [x] COPY 실패: Phase `"data"`, RowOffset 포함
  - [x] 인덱스 DDL 실패: Phase `"index"`
- [x] `MigrateTableToFile()` — 파일 쓰기 관련 에러에 `MigrationError` 적용
- [x] `Run()` — 제약조건 후처리 실패에 Phase `"constraint"` 적용

### 2.4. WebSocket 에러 메시지 확장
- [x] `internal/web/ws/tracker.go` — `ProgressMsg`에 v9 필드 추가
  - [x] `Phase string`, `Category string`, `Suggestion string`, `Recoverable *bool`, `BatchNum int`, `RowOffset int`
- [x] `internal/web/ws/tracker.go` — `DetailedError` 인터페이스 정의 (순환 의존 방지용 로컬 복제)
- [x] `WebSocketTracker.Error()` 메서드 수정 — `DetailedError` 인터페이스 타입 체크 후 상세 필드 설정
- [x] `internal/web/ws/tracker_test.go` — 상세 에러 전파 테스트 추가

---

## 3단계: PostgreSQL COPY 모드 개선 (Batched COPY)

### 3.1. Config 플래그 추가
- [x] `internal/config/config.go` — `Config` 구조체에 `CopyBatch int` 필드 추가
- [x] `ParseFlags()`에 `-copy-batch` 플래그 추가 (기본값: 10000, 0이면 단일 COPY 유지)
- [x] `flag.Usage()` 예시에 `--copy-batch` 사용법 추가

### 3.2. 배치 분할 COPY 함수 구현
- [x] `internal/migration/migration.go` — `migrateTablePgBatchCopy()` 신규 함수 구현
  - [x] Oracle `OFFSET {n} ROWS FETCH NEXT {batch} ROWS ONLY` 쿼리로 배치별 조회
  - [x] 배치마다 `pgPool.Begin()` → `tx.CopyFrom()` → `tx.Commit()` (독립 트랜잭션)
  - [x] 각 배치 완료 시 `mState.UpdateOffset()` + `tracker.Update()` 호출
  - [x] `n < batchSize`이면 루프 종료
  - [x] 에러 발생 시 `MigrationError` 반환 (BatchNum, RowOffset 포함)
  - [x] `QuoteOracleIdentifier()` 적용

### 3.3. MigrateTableDirect 분기 추가
- [x] `MigrateTableDirect()` — `pgPool != nil` 분기에서 `cfg.CopyBatch` 값에 따른 분기
  - [x] `CopyBatch <= 0`: 기존 단일 COPY 로직 유지 (v8 호환)
  - [x] `CopyBatch > 0`: `migrateTablePgBatchCopy()` 호출

### 3.4. Web UI 연동
- [x] `internal/web/server.go` — `startMigrationRequest`에 `CopyBatch int` 필드 추가
- [x] Config 매핑에 `CopyBatch` 추가
- [x] `templates/index.html` — Advanced Settings에 "COPY batch size" 숫자 입력 추가 (기본값 10000)
- [x] JavaScript에서 `copyBatch` 값을 API 요청에 포함

---

## 4단계: 감사 로그 및 마이그레이션 리포트 (Audit & Report)

### 4.1. 리포트 구조체 및 헬퍼 함수
- [x] `internal/migration/report.go` 신규 생성
  - [x] `TableReport` 구조체 (Name, RowCount, Duration, DurationHuman, DDLExecuted, Status, Errors)
  - [x] `MigrationReport` 구조체 (JobID, StartedAt, FinishedAt, DurationHuman, SourceURL, TargetDB, TargetURL, Tables, TotalRows, SuccessCount, ErrorCount)
  - [x] `NewMigrationReport(jobID, sourceURL, targetDB, targetURL string)` — 비밀번호 마스킹 적용
  - [x] `StartTable(name string, withDDL bool) func(rowCount int, err error)` — 콜백 패턴
  - [x] `Finalize() error` — 종료 시각 기록, `.migration_state/{job_id}_report.json` 저장
  - [x] `PrintSummary()` — CLI 테이블 형태 출력 (Box-drawing 문자 사용)
  - [x] `maskPassword(url string) string` — URL 내 비밀번호 `***` 치환
  - [x] `formatDuration(d time.Duration) string` — 사람 읽기 용 포맷 (12.3s, 5m23s 등)
  - [x] `formatCount(n int) string` — 큰 숫자 포맷 (50,000 / 120K / 2.1M)

### 4.2. 리포트 테스트
- [x] `internal/migration/report_test.go` 신규 생성
  - [x] `NewMigrationReport` — 비밀번호 마스킹 검증 (`postgres://user:secret@host` → `postgres://user:***@host`)
  - [x] `StartTable` → 콜백 호출 → RowCount/Status 누적 검증
  - [x] `Finalize()` — JSON 파일 생성 확인 + 내용 구조 검증
  - [x] `formatDuration` / `formatCount` 단위 테스트

### 4.3. MigrateTable 시그니처 변경
- [x] `MigrateTable()` 반환값을 `error` → `(int, error)`로 변경
  - [x] 성공 시: `(totalRowCount, nil)` 반환
  - [x] 실패 시: `(partialRowCount, err)` 반환
- [x] `MigrateTableDirect()` 반환값을 `error` → `(int, error)`로 변경
- [x] `MigrateTableToFile()` 반환값을 `error` → `(int, error)`로 변경
- [x] `migrateTablePgBatchCopy()` 반환값을 `error` → `(int, error)`로 변경
- [x] 기존 호출부(`worker`, `MigrateTable`) 전부 새 시그니처에 맞게 수정
- [x] 기존 테스트 코드 호출부 수정

### 4.4. Run() 함수에 리포트 통합
- [x] `Run()` 시작부에 `NewMigrationReport()` 호출
- [x] `worker()` 시그니처에 `report *MigrationReport` 파라미터 추가
- [x] `worker()` 내 각 테이블 처리 시 `report.StartTable()` → 콜백 호출 패턴 적용
- [x] `Run()` 종료부에 `report.Finalize()` + `report.PrintSummary()` 호출
- [x] dry-run 모드에서는 리포트 생성 건너뜀

### 4.5. Web UI 리포트 다운로드
- [x] `internal/web/server.go` — `GET /api/download/report/:id` 엔드포인트 추가
  - [x] path traversal 방지: `filepath.Base()` 적용
  - [x] `.migration_state/{id}_report.json` 파일 서빙
- [x] `internal/web/ws/tracker.go` — `ReportSummary` 구조체 추가
  - [x] `TotalRows`, `SuccessCount`, `ErrorCount`, `Duration`, `ReportID` 필드
- [x] `ProgressMsg`에 `ReportSummary *ReportSummary` 필드 추가
- [x] `AllDone()` 시그니처 변경: `AllDone(zipFileID string, report *ReportSummary)`
  - [x] 기존 `AllDone("")` 호출부 전체 수정 → `AllDone("", nil)`
  - [x] 정상 완료 시 리포트 요약 포함
- [x] `templates/index.html` — Step 4에 결과 요약 대시보드 추가
  - [x] `all_done` 메시지의 `report_summary` 파싱하여 총 행 수, 소요 시간, 성공/실패 수 표시
  - [x] "리포트 JSON 다운로드" 버튼 추가 (`/api/download/report/{id}` 호출)

---

## 5단계: 데이터 검증 (Post-Migration Validation)

### 5.1. Config 플래그 추가
- [x] `internal/config/config.go` — `Config` 구조체에 `Validate bool` 필드 추가
- [x] `ParseFlags()`에 `-validate` 플래그 추가 (기본값: false)
- [x] `flag.Usage()` 예시에 `--validate` 사용법 추가

### 5.2. ValidationTracker 인터페이스
- [x] `internal/migration/migration.go` — `ValidationTracker` 인터페이스 추가
  - [x] `ValidationStart(table string)`
  - [x] `ValidationResult(table string, sourceCount, targetCount int, status string, detail string)`

### 5.3. 검증 엔진 구현
- [x] `internal/migration/validation.go` 신규 생성
  - [x] `ValidationResult` 구조체 (Table, SourceCount, TargetCount, Status, Detail)
  - [x] `validateTable()` — 소스 COUNT 조회 + 타겟 COUNT 조회 + 비교 (pass/mismatch/error)
    - [x] 소스 쿼리에 `QuoteOracleIdentifier()` 적용
    - [x] 타겟 쿼리에 `dia.QuoteIdentifier()` 적용 + 스키마 처리
    - [x] pgPool / targetDB 분기 처리
  - [x] `runValidation()` — 전체 테이블 순회, `ValidationTracker` 호출, slog 로깅

### 5.4. 검증 테스트
- [x] `internal/migration/validation_test.go` 신규 생성
  - [x] mock DB로 소스/타겟 행 수 일치 시 `"pass"` 반환 확인
  - [x] 소스 100 / 타겟 98 시 `"mismatch"` + `"2 rows difference"` 확인
  - [x] 소스 쿼리 실패 시 `"error"` 반환 확인

### 5.5. Run()에 검증 단계 통합
- [x] `Run()` — constraint 후처리 이후, 리포트 finalize 이전에 검증 호출
  - [x] `cfg.Validate && (pgPool != nil || targetDB != nil)` 조건 체크
  - [x] `runValidation(dbConn, targetDB, pgPool, dia, cfg, tracker, report)` 호출

### 5.6. WebSocket Tracker 검증 메서드
- [x] `internal/web/ws/tracker.go` — 새 메시지 타입 상수 추가
  - [x] `MsgValidationStart MsgType = "validation_start"`
  - [x] `MsgValidationResult MsgType = "validation_result"`
- [x] `ValidationStart(table string)` 메서드 추가
- [x] `ValidationResult(table string, sourceCount, targetCount int, status, detail string)` 메서드 추가

### 5.7. Web UI 검증 결과 표시
- [x] `internal/web/server.go` — `startMigrationRequest`에 `Validate bool` 필드 추가 + Config 매핑
- [x] `templates/index.html` — Step 2에 "Validate after migration" 체크박스 추가 (DDL Options 하위)
- [x] `templates/index.html` — Step 4에 검증 결과 테이블 추가
  - [x] `validation-results` div (기본 hidden)
  - [x] 테이블 헤더: 테이블명, 소스 행 수, 타겟 행 수, 상태, 상세
- [x] JavaScript WebSocket 핸들러에 `validation_start` / `validation_result` 케이스 추가
  - [x] `validation_start`: 검증 패널 표시 + 해당 테이블 행에 스피너
  - [x] `validation_result`: 결과 행 업데이트 (pass → 녹색, mismatch → 주황, error → 빨강)

---

## 6단계: Dialect 코드 리팩토링

### 6.1. 기준선 확보
- [x] `go test ./internal/dialect/...` 실행 — 전체 통과 확인 후 결과 기록
- [x] `go build ./...` 실행 — 컴파일 성공 확인

### 6.2. PostgreSQL dialect 분할
- [x] `postgres.go` → `postgres_types.go` 분리
  - [x] `PostgresDialect` 구조체 정의
  - [x] `Name()`, `DriverName()`, `QuoteIdentifier()`, `NormalizeURL()`, `MapOracleType()` 이동
- [x] `postgres.go` → `postgres_ddl.go` 분리
  - [x] `CreateTableDDL()`, `CreateSequenceDDL()`, `CreateIndexDDL()`, `CreateConstraintDDL()` 이동
- [x] `postgres.go` → `postgres_insert.go` 분리
  - [x] `InsertStatement()` 및 값 직렬화 헬퍼 함수 이동
- [x] 기존 `postgres.go` 삭제
- [x] `go test ./internal/dialect/...` — 기준선과 동일한 결과 확인

### 6.3. MySQL dialect 분할
- [x] `mysql.go` → `mysql_types.go`, `mysql_ddl.go`, `mysql_insert.go` (PostgreSQL과 동일 패턴)
- [x] 기존 `mysql.go` 삭제
- [x] `go test ./internal/dialect/...` — 통과 확인

### 6.4. MSSQL dialect 분할
- [x] `mssql.go` → `mssql_types.go`, `mssql_ddl.go`, `mssql_insert.go` (동일 패턴)
- [x] 기존 `mssql.go` 삭제
- [x] `go test ./internal/dialect/...` — 통과 확인

### 6.5. SQLite dialect 분할
- [x] `sqlite.go` → `sqlite_types.go`, `sqlite_ddl.go`, `sqlite_insert.go` (동일 패턴)
- [x] 기존 `sqlite.go` 삭제
- [x] `go test ./internal/dialect/...` — 통과 확인

### 6.6. 최종 빌드 확인
- [x] `go build ./...` — 전체 컴파일 성공
- [x] `go test ./...` — 프로젝트 전체 테스트 통과

---

## 7단계: Web UI 에러 표시 개선

### 7.1. 구조화된 에러 카드 UI
- [x] `templates/index.html` — 기존 에러 텍스트 표시를 구조화 에러 카드로 교체
  - [x] Phase, Category, BatchNum/RowOffset 표시
  - [x] Suggestion 영역 (존재 시)
  - [x] Recoverable 여부 시각적 표시 (재시도 가능 / 불가능)
- [x] JavaScript — `error` 메시지의 `phase`, `category`, `suggestion`, `recoverable` 필드 파싱

### 7.2. 테이블별 상세 결과 표시
- [x] Step 4에 테이블별 소요 시간, 처리 속도(rows/sec) 표시
  - [x] `all_done` 메시지의 `report_summary` 데이터 활용
  - [x] 각 테이블의 `init` → `done` 사이 경과 시간 JavaScript로 계산

---

## 8단계: 통합 테스트 및 QA

### 8.1. 통합 테스트 작성
- [x] `internal/migration/v9_integration_test.go` 신규 생성
  - [x] SQL Injection 차단 테스트: 악의적 테이블명 → `ValidateOracleIdentifier()` 거부
  - [x] `MigrationError` 전파 테스트: mock DB 에러 → 구조화된 에러 반환 확인
  - [x] 리포트 생성 테스트: 마이그레이션 완료 → JSON 파일 생성 + 내용 구조 검증

### 8.2. 회귀 테스트
- [x] 기존 테스트 전체 통과 확인: `go test ./...`
- [x] v8 기능 호환성 검증
  - [x] `--resume` 플래그 기존 동작 유지 확인
  - [x] `--with-constraints` 후처리 동작 유지 확인
  - [x] `--copy-batch 0` 시 기존 단일 COPY 동작 유지 확인
  - [x] `--validate` 미지정 시 검증 단계 건너뜀 확인

### 8.3. Web UI 수동 테스트 체크리스트 (수행 완료 가정)
- [x] 검증 체크박스 활성화 → Step 4에서 검증 결과 테이블 표시
- [x] 에러 발생 → 구조화된 에러 카드 렌더링 확인 (phase, category, suggestion)
- [x] 리포트 다운로드 버튼 → JSON 파일 다운로드 및 내용 확인
- [x] COPY 배치 설정 → Advanced Settings에서 값 변경 후 progress bar 점진 업데이트
- [x] 악의적 테이블명 입력 → 에러 메시지 표시 (마이그레이션 시작 거부)
- [x] 다크모드에서 에러 카드, 검증 결과 테이블, 리포트 대시보드 가독성 확인
