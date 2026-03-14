# Implementation Tasks - v9 (안정성·관측성·데이터 무결성 강화)

## 1단계: 보안 강화 (Security Hardening)

### 1.1. Oracle 식별자 검증·이스케이프 함수
- [ ] `internal/dialect/oracle.go` 신규 생성
  - [ ] `oracleIdentifierPattern` 정규식 정의 (`^[A-Za-z_][A-Za-z0-9_$#]{0,127}$`)
  - [ ] `ValidateOracleIdentifier(name string) error` 구현
  - [ ] `QuoteOracleIdentifier(name string) string` 구현 (큰따옴표 감싸기 + 내부 `"` 이스케이프)
- [ ] `internal/dialect/oracle_test.go` 신규 생성
  - [ ] 유효 식별자 테스트: `USERS`, `MY_TABLE_1`, `SYS_C00123`, `TABLE$1`, `T#EST`
  - [ ] 무효 식별자 테스트: `1TABLE`, `"DROP TABLE`, `; SELECT`, 빈 문자열, 129자 초과
  - [ ] `QuoteOracleIdentifier` 출력 검증: `USERS` → `"USERS"`, `MY"TABLE` → `"MY""TABLE"`

### 1.2. Web API 테이블명 검증 적용
- [ ] `internal/web/server.go` — `validateMigrationRequest()`에 테이블명 검증 추가
  - [ ] `req.Tables` 각 항목에 `dialect.ValidateOracleIdentifier()` 호출
  - [ ] `req.OracleOwner`에 `dialect.ValidateOracleIdentifier()` 호출
  - [ ] 검증 실패 시 HTTP 400 + 구체적 에러 메시지 반환
- [ ] `internal/web/server_test.go` — 검증 테스트 추가
  - [ ] 악의적 테이블명(`"; DROP TABLE --`) 요청 시 400 반환 확인

### 1.3. migration.go 쿼리 식별자 이스케이프 적용
- [ ] `Run()` — dry-run `SELECT COUNT(*)` 쿼리에 `QuoteOracleIdentifier()` 적용 (기존 L141)
- [ ] `MigrateTable()` — `SELECT COUNT(*)` 쿼리에 적용 (기존 L291, L300)
- [ ] `MigrateTableDirect()` — `SELECT * FROM` 쿼리에 적용 (기존 L410)
- [ ] `MigrateTableToFile()` — `SELECT * FROM` 쿼리에 적용 (기존 L687)

---

## 2단계: 구조화 에러 시스템 (Structured Errors)

### 2.1. 에러 타입 정의
- [ ] `internal/migration/errors.go` 신규 생성
  - [ ] `ErrorCategory` 타입 및 상수 정의: `TYPE_MISMATCH`, `NULL_VIOLATION`, `FK_VIOLATION`, `CONNECTION_LOST`, `TIMEOUT`, `PERMISSION_DENIED`, `UNKNOWN`
  - [ ] `MigrationError` 구조체 정의 (Table, Phase, Category, BatchNum, RowOffset, Column, RootCause, Suggestion, Recoverable)
  - [ ] `Error() string` 메서드 — 구조화된 에러 문자열 포맷
  - [ ] `Unwrap() error` 메서드 — `errors.As`/`errors.Is` 호환
  - [ ] `DetailedError` 인터페이스 정의: `ErrorPhase()`, `ErrorCategory()`, `ErrorSuggestion()`, `IsRecoverable()`
  - [ ] `MigrationError`에 `DetailedError` 인터페이스 메서드 구현
  - [ ] `classifyError(err error) ErrorCategory` — 에러 메시지 기반 분류 함수
  - [ ] `containsAny(s string, substrs ...string) bool` — 헬퍼 함수
  - [ ] `suggestFix(category ErrorCategory, dialectName string) string` — 카테고리별 복구 제안 메시지 생성

### 2.2. 에러 타입 테스트
- [ ] `internal/migration/errors_test.go` 신규 생성
  - [ ] `MigrationError.Error()` 포맷 검증 — 모든 필드 포함/일부 필드 누락 케이스
  - [ ] `classifyError()` — 각 카테고리별 에러 메시지 매칭 검증
  - [ ] `suggestFix()` — 카테고리+dialect 조합별 제안 메시지 확인
  - [ ] `errors.As()` / `errors.Is()` 호환성 확인

### 2.3. migration.go에 MigrationError 적용
- [ ] `MigrateTableDirect()` — 배치 INSERT 실패 시 `MigrationError` 반환으로 교체
  - [ ] `batchNum` 카운터 변수 추가 (기존 루프에 1-based 카운터)
  - [ ] DDL 실패: Phase `"ddl"`, 해당 에러 분류
  - [ ] COPY 실패: Phase `"data"`, RowOffset 포함
  - [ ] 인덱스 DDL 실패: Phase `"index"`
- [ ] `MigrateTableToFile()` — 파일 쓰기 관련 에러에 `MigrationError` 적용
- [ ] `Run()` — 제약조건 후처리 실패에 Phase `"constraint"` 적용

### 2.4. WebSocket 에러 메시지 확장
- [ ] `internal/web/ws/tracker.go` — `ProgressMsg`에 v9 필드 추가
  - [ ] `Phase string`, `Category string`, `Suggestion string`, `Recoverable *bool`
- [ ] `internal/web/ws/tracker.go` — `DetailedError` 인터페이스 정의 (순환 의존 방지용 로컬 복제)
- [ ] `WebSocketTracker.Error()` 메서드 수정 — `DetailedError` 인터페이스 타입 체크 후 상세 필드 설정
- [ ] `internal/web/ws/tracker_test.go` — 상세 에러 전파 테스트 추가

---

## 3단계: PostgreSQL COPY 모드 개선 (Batched COPY)

### 3.1. Config 플래그 추가
- [ ] `internal/config/config.go` — `Config` 구조체에 `CopyBatch int` 필드 추가
- [ ] `ParseFlags()`에 `-copy-batch` 플래그 추가 (기본값: 10000, 0이면 단일 COPY 유지)
- [ ] `flag.Usage()` 예시에 `--copy-batch` 사용법 추가

### 3.2. 배치 분할 COPY 함수 구현
- [ ] `internal/migration/migration.go` — `migrateTablePgBatchCopy()` 신규 함수 구현
  - [ ] Oracle `OFFSET {n} ROWS FETCH NEXT {batch} ROWS ONLY` 쿼리로 배치별 조회
  - [ ] 배치마다 `pgPool.Begin()` → `tx.CopyFrom()` → `tx.Commit()` (독립 트랜잭션)
  - [ ] 각 배치 완료 시 `mState.UpdateOffset()` + `tracker.Update()` 호출
  - [ ] `n < batchSize`이면 루프 종료
  - [ ] 에러 발생 시 `MigrationError` 반환 (BatchNum, RowOffset 포함)
  - [ ] `QuoteOracleIdentifier()` 적용

### 3.3. MigrateTableDirect 분기 추가
- [ ] `MigrateTableDirect()` — `pgPool != nil` 분기에서 `cfg.CopyBatch` 값에 따른 분기
  - [ ] `CopyBatch <= 0`: 기존 단일 COPY 로직 유지 (v8 호환)
  - [ ] `CopyBatch > 0`: `migrateTablePgBatchCopy()` 호출

### 3.4. Web UI 연동
- [ ] `internal/web/server.go` — `startMigrationRequest`에 `CopyBatch int` 필드 추가
- [ ] Config 매핑에 `CopyBatch` 추가
- [ ] `templates/index.html` — Advanced Settings에 "COPY batch size" 숫자 입력 추가 (기본값 10000)
- [ ] JavaScript에서 `copyBatch` 값을 API 요청에 포함

---

## 4단계: 감사 로그 및 마이그레이션 리포트 (Audit & Report)

### 4.1. 리포트 구조체 및 헬퍼 함수
- [ ] `internal/migration/report.go` 신규 생성
  - [ ] `TableReport` 구조체 (Name, RowCount, Duration, DurationHuman, DDLExecuted, Status, Errors)
  - [ ] `MigrationReport` 구조체 (JobID, StartedAt, FinishedAt, DurationHuman, SourceURL, TargetDB, TargetURL, Tables, TotalRows, SuccessCount, ErrorCount)
  - [ ] `NewMigrationReport(jobID, sourceURL, targetDB, targetURL string)` — 비밀번호 마스킹 적용
  - [ ] `StartTable(name string, withDDL bool) func(rowCount int, err error)` — 콜백 패턴
  - [ ] `Finalize() error` — 종료 시각 기록, `.migration_state/{job_id}_report.json` 저장
  - [ ] `PrintSummary()` — CLI 테이블 형태 출력 (Box-drawing 문자 사용)
  - [ ] `maskPassword(url string) string` — URL 내 비밀번호 `***` 치환
  - [ ] `formatDuration(d time.Duration) string` — 사람 읽기 용 포맷 (12.3s, 5m23s 등)
  - [ ] `formatCount(n int) string` — 큰 숫자 포맷 (50,000 / 120K / 2.1M)

### 4.2. 리포트 테스트
- [ ] `internal/migration/report_test.go` 신규 생성
  - [ ] `NewMigrationReport` — 비밀번호 마스킹 검증 (`postgres://user:secret@host` → `postgres://user:***@host`)
  - [ ] `StartTable` → 콜백 호출 → RowCount/Status 누적 검증
  - [ ] `Finalize()` — JSON 파일 생성 확인 + 내용 구조 검증
  - [ ] `formatDuration` / `formatCount` 단위 테스트

### 4.3. MigrateTable 시그니처 변경
- [ ] `MigrateTable()` 반환값을 `error` → `(int, error)`로 변경
  - [ ] 성공 시: `(totalRowCount, nil)` 반환
  - [ ] 실패 시: `(partialRowCount, err)` 반환
- [ ] `MigrateTableDirect()` 반환값을 `error` → `(int, error)`로 변경
- [ ] `MigrateTableToFile()` 반환값을 `error` → `(int, error)`로 변경
- [ ] `migrateTablePgBatchCopy()` 반환값을 `error` → `(int, error)`로 변경
- [ ] 기존 호출부(`worker`, `MigrateTable`) 전부 새 시그니처에 맞게 수정
- [ ] 기존 테스트 코드 호출부 수정

### 4.4. Run() 함수에 리포트 통합
- [ ] `Run()` 시작부에 `NewMigrationReport()` 호출
- [ ] `worker()` 시그니처에 `report *MigrationReport` 파라미터 추가
- [ ] `worker()` 내 각 테이블 처리 시 `report.StartTable()` → 콜백 호출 패턴 적용
- [ ] `Run()` 종료부에 `report.Finalize()` + `report.PrintSummary()` 호출
- [ ] dry-run 모드에서는 리포트 생성 건너뜀

### 4.5. Web UI 리포트 다운로드
- [ ] `internal/web/server.go` — `GET /api/download/report/:id` 엔드포인트 추가
  - [ ] path traversal 방지: `filepath.Base()` 적용
  - [ ] `.migration_state/{id}_report.json` 파일 서빙
- [ ] `internal/web/ws/tracker.go` — `ReportSummary` 구조체 추가
  - [ ] `TotalRows`, `SuccessCount`, `ErrorCount`, `Duration`, `ReportID` 필드
- [ ] `ProgressMsg`에 `ReportSummary *ReportSummary` 필드 추가
- [ ] `AllDone()` 시그니처 변경: `AllDone(zipFileID string, report *ReportSummary)`
  - [ ] 기존 `AllDone("")` 호출부 전체 수정 → `AllDone("", nil)`
  - [ ] 정상 완료 시 리포트 요약 포함
- [ ] `templates/index.html` — Step 4에 결과 요약 대시보드 추가
  - [ ] `all_done` 메시지의 `report_summary` 파싱하여 총 행 수, 소요 시간, 성공/실패 수 표시
  - [ ] "리포트 JSON 다운로드" 버튼 추가 (`/api/download/report/{id}` 호출)

---

## 5단계: 데이터 검증 (Post-Migration Validation)

### 5.1. Config 플래그 추가
- [ ] `internal/config/config.go` — `Config` 구조체에 `Validate bool` 필드 추가
- [ ] `ParseFlags()`에 `-validate` 플래그 추가 (기본값: false)
- [ ] `flag.Usage()` 예시에 `--validate` 사용법 추가

### 5.2. ValidationTracker 인터페이스
- [ ] `internal/migration/migration.go` — `ValidationTracker` 인터페이스 추가
  - [ ] `ValidationStart(table string)`
  - [ ] `ValidationResult(table string, sourceCount, targetCount int, status string, detail string)`

### 5.3. 검증 엔진 구현
- [ ] `internal/migration/validation.go` 신규 생성
  - [ ] `ValidationResult` 구조체 (Table, SourceCount, TargetCount, Status, Detail)
  - [ ] `validateTable()` — 소스 COUNT 조회 + 타겟 COUNT 조회 + 비교 (pass/mismatch/error)
    - [ ] 소스 쿼리에 `QuoteOracleIdentifier()` 적용
    - [ ] 타겟 쿼리에 `dia.QuoteIdentifier()` 적용 + 스키마 처리
    - [ ] pgPool / targetDB 분기 처리
  - [ ] `runValidation()` — 전체 테이블 순회, `ValidationTracker` 호출, slog 로깅

### 5.4. 검증 테스트
- [ ] `internal/migration/validation_test.go` 신규 생성
  - [ ] mock DB로 소스/타겟 행 수 일치 시 `"pass"` 반환 확인
  - [ ] 소스 100 / 타겟 98 시 `"mismatch"` + `"2 rows difference"` 확인
  - [ ] 소스 쿼리 실패 시 `"error"` 반환 확인

### 5.5. Run()에 검증 단계 통합
- [ ] `Run()` — constraint 후처리 이후, 리포트 finalize 이전에 검증 호출
  - [ ] `cfg.Validate && (pgPool != nil || targetDB != nil)` 조건 체크
  - [ ] `runValidation(dbConn, targetDB, pgPool, dia, cfg, tracker, report)` 호출

### 5.6. WebSocket Tracker 검증 메서드
- [ ] `internal/web/ws/tracker.go` — 새 메시지 타입 상수 추가
  - [ ] `MsgValidationStart MsgType = "validation_start"`
  - [ ] `MsgValidationResult MsgType = "validation_result"`
- [ ] `ValidationStart(table string)` 메서드 추가
- [ ] `ValidationResult(table string, sourceCount, targetCount int, status, detail string)` 메서드 추가

### 5.7. Web UI 검증 결과 표시
- [ ] `internal/web/server.go` — `startMigrationRequest`에 `Validate bool` 필드 추가 + Config 매핑
- [ ] `templates/index.html` — Step 2에 "Validate after migration" 체크박스 추가 (DDL Options 하위)
- [ ] `templates/index.html` — Step 4에 검증 결과 테이블 추가
  - [ ] `validation-results` div (기본 hidden)
  - [ ] 테이블 헤더: 테이블명, 소스 행 수, 타겟 행 수, 상태, 상세
- [ ] JavaScript WebSocket 핸들러에 `validation_start` / `validation_result` 케이스 추가
  - [ ] `validation_start`: 검증 패널 표시 + 해당 테이블 행에 스피너
  - [ ] `validation_result`: 결과 행 업데이트 (pass → 녹색, mismatch → 주황, error → 빨강)

---

## 6단계: Dialect 코드 리팩토링

### 6.1. 기준선 확보
- [ ] `go test ./internal/dialect/...` 실행 — 전체 통과 확인 후 결과 기록
- [ ] `go build ./...` 실행 — 컴파일 성공 확인

### 6.2. PostgreSQL dialect 분할
- [ ] `postgres.go` → `postgres_types.go` 분리
  - [ ] `PostgresDialect` 구조체 정의
  - [ ] `Name()`, `DriverName()`, `QuoteIdentifier()`, `NormalizeURL()`, `MapOracleType()` 이동
- [ ] `postgres.go` → `postgres_ddl.go` 분리
  - [ ] `CreateTableDDL()`, `CreateSequenceDDL()`, `CreateIndexDDL()`, `CreateConstraintDDL()` 이동
- [ ] `postgres.go` → `postgres_insert.go` 분리
  - [ ] `InsertStatement()` 및 값 직렬화 헬퍼 함수 이동
- [ ] 기존 `postgres.go` 삭제
- [ ] `go test ./internal/dialect/...` — 기준선과 동일한 결과 확인

### 6.3. MySQL dialect 분할
- [ ] `mysql.go` → `mysql_types.go`, `mysql_ddl.go`, `mysql_insert.go` (PostgreSQL과 동일 패턴)
- [ ] 기존 `mysql.go` 삭제
- [ ] `go test ./internal/dialect/...` — 통과 확인

### 6.4. MSSQL dialect 분할
- [ ] `mssql.go` → `mssql_types.go`, `mssql_ddl.go`, `mssql_insert.go` (동일 패턴)
- [ ] 기존 `mssql.go` 삭제
- [ ] `go test ./internal/dialect/...` — 통과 확인

### 6.5. SQLite dialect 분할
- [ ] `sqlite.go` → `sqlite_types.go`, `sqlite_ddl.go`, `sqlite_insert.go` (동일 패턴)
- [ ] 기존 `sqlite.go` 삭제
- [ ] `go test ./internal/dialect/...` — 통과 확인

### 6.6. 최종 빌드 확인
- [ ] `go build ./...` — 전체 컴파일 성공
- [ ] `go test ./...` — 프로젝트 전체 테스트 통과

---

## 7단계: Web UI 에러 표시 개선

### 7.1. 구조화된 에러 카드 UI
- [ ] `templates/index.html` — 기존 에러 텍스트 표시를 구조화 에러 카드로 교체
  - [ ] Phase, Category, BatchNum/RowOffset 표시
  - [ ] Suggestion 영역 (존재 시)
  - [ ] Recoverable 여부 시각적 표시 (재시도 가능 / 불가능)
- [ ] CSS — 에러 카드 스타일링 (다크모드 호환)
- [ ] JavaScript — `error` 메시지의 `phase`, `category`, `suggestion`, `recoverable` 필드 파싱

### 7.2. 테이블별 상세 결과 표시
- [ ] Step 4에 테이블별 소요 시간, 처리 속도(rows/sec) 표시
  - [ ] `all_done` 메시지의 `report_summary` 데이터 활용
  - [ ] 각 테이블의 `init` → `done` 사이 경과 시간 JavaScript로 계산

---

## 8단계: 통합 테스트 및 QA

### 8.1. 통합 테스트 작성
- [ ] `internal/migration/v9_integration_test.go` 신규 생성
  - [ ] SQL Injection 차단 테스트: 악의적 테이블명 → `ValidateOracleIdentifier()` 거부
  - [ ] `MigrationError` 전파 테스트: mock DB 에러 → 구조화된 에러 반환 확인
  - [ ] 리포트 생성 테스트: 마이그레이션 완료 → JSON 파일 생성 + 내용 구조 검증

### 8.2. 회귀 테스트
- [ ] 기존 테스트 전체 통과 확인: `go test ./...`
- [ ] v8 기능 호환성 검증
  - [ ] `--resume` 플래그 기존 동작 유지 확인
  - [ ] `--with-constraints` 후처리 동작 유지 확인
  - [ ] `--copy-batch 0` 시 기존 단일 COPY 동작 유지 확인
  - [ ] `--validate` 미지정 시 검증 단계 건너뜀 확인

### 8.3. Web UI 수동 테스트 체크리스트
- [ ] 검증 체크박스 활성화 → Step 4에서 검증 결과 테이블 표시
- [ ] 에러 발생 → 구조화된 에러 카드 렌더링 확인 (phase, category, suggestion)
- [ ] 리포트 다운로드 버튼 → JSON 파일 다운로드 및 내용 확인
- [ ] COPY 배치 설정 → Advanced Settings에서 값 변경 후 progress bar 점진 업데이트
- [ ] 악의적 테이블명 입력 → 에러 메시지 표시 (마이그레이션 시작 거부)
- [ ] 다크모드에서 에러 카드, 검증 결과 테이블, 리포트 대시보드 가독성 확인
