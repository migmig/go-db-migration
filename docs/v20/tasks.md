# 작업 목록 (Tasks) - v20

## 목표: 세션 보안 강화·SQL 식별자 인용·입력 검증·재시도 정책 도입

### 1. 문서화
- [x] `docs/v20/spec.md` 작성
- [x] `docs/v20/tasks.md` 작성

### 2. FR-1: 세션 자동 정리 (`internal/web/server.go`)
- [x] `authSession` 구조체에 `ExpiresAt time.Time` 필드 추가
- [x] `authSessionManager`에 `maxSessions int`, `stopCleanup chan struct{}` 필드 추가
- [x] `newAuthSessionManager` 시그니처에 `maxSessions int` 파라미터 추가
- [x] `purgeExpired()` 메서드 구현 (만료·유휴 세션 일괄 삭제)
- [x] `startCleanupLoop(interval time.Duration)` 고루틴 구현
- [x] `evictOldest()` 메서드 구현 (최대 세션 수 초과 시 가장 오래된 세션 삭제)
- [x] `createSession` 내 `ExpiresAt` 설정 및 세션 한도 초과 처리 추가
- [x] `RunServerWithAuth` 종료 시 `close(stopCleanup)` 연동
- [x] 환경변수 `DBM_MAX_SESSIONS`, `DBM_SESSION_CLEANUP_INTERVAL` 파싱 적용

### 3. FR-2: SQL 식별자 인용 (`internal/db/db.go`)
- [x] `SQLDBCountFn` 시그니처에 `quoteIdentifier func(string) string` 파라미터 추가
- [x] `PGPoolCountFn` 시그니처에 `quoteIdentifier func(string) string` 파라미터 추가
- [x] 두 함수 내부의 문자열 연결 `COUNT(*)` 쿼리를 `quoteIdentifier(tableName)` 적용으로 교체
- [x] 호출부에서 해당 Dialect의 `QuoteIdentifier` 메서드 전달

### 4. FR-3: 수치형 입력 검증 (`internal/config/config.go`)
- [x] `validateConfig(cfg *Config) error` 함수 구현
  - [x] `--batch` 범위: 1 ~ 100,000
  - [x] `--workers` 범위: 1 ~ 64
  - [x] `--db-max-open` 범위: 1 ~ 1,000
  - [x] `--db-max-idle` 범위: 0 ~ 1,000
- [x] `ParseFlags` 완료 후 `validateConfig` 호출 및 오류 시 `os.Exit(1)`

### 5. FR-4: 지수 백오프 재시도 (`internal/migration/retry.go` 신규)
- [x] `RetryConfig` 구조체 정의 (`MaxAttempts`, `InitialWait`, `Multiplier`, `MaxWait`)
- [x] `DefaultRetryConfig()` 함수 구현
- [x] `RetryEvent` 구조체 추가 (`internal/migration/errors.go`)
- [x] `WithRetry(ctx, cfg, tableName, eventFn, fn)` 함수 구현
  - [x] `MigrationError.Recoverable=true` 경우에만 재시도
  - [x] `ctx.Done()` 시 즉시 중단
  - [x] 재시도 발생 시 `slog.Warn` 로그 출력
  - [x] `eventFn`을 통해 `RetryEvent` 전달
- [x] 마이그레이션 엔진 내 `ErrConnectionLost`·`ErrTimeout` 발생 위치에 `WithRetry` 적용
- [x] 환경변수 `DBM_MAX_RETRIES`, `DBM_RETRY_INITIAL_WAIT` 파싱 적용

### 6. FR-5: 부분 실패 허용 정책 (`skip_batch`)
- [x] `internal/migration/state.go`에 `StatusPartialSuccess = "partial_success"` 추가
- [x] `internal/config/config.go`에 `OnError string` 필드 및 `--on-error` 플래그 추가
- [ ] `internal/migration/migration.go` 배치 루프 내 `OnError="skip_batch"` 분기 처리
  - [x] 건너뛴 배치 수(`skippedBatches`) 카운트
  - [x] 완료 후 상태를 `partial_success`로 기록
- [x] 최종 리포트에 `skipped_batches`, `estimated_skipped_rows` 필드 추가

### 7. 웹소켓/API 연계
- [x] `bus` 패키지에 `retry` 이벤트 타입 추가
- [x] `POST /api/migrate` 요청 바디에 `on_error` 필드 파싱 지원
- [ ] 테이블별 응답에 `skipped_batches`, `estimated_skipped_rows` 필드 추가

### 8. UI
- [ ] 진행 패널에 재시도 상태 행 추가 (WebSocket `retry` 이벤트 수신 시 표시)
- [ ] `partial_success` 뱃지 추가 (노란색, 호버 툴팁)
- [ ] 마이그레이션 설정 패널에 `on-error` 정책 라디오 버튼 추가

### 9. 관측성
- [ ] `session_cleanup_total`, `session_evicted_total` 메트릭 추가 (`monitoring.go`)
- [ ] `migration_retry_total{table}`, `migration_batch_skipped_total{table}` 메트릭 추가
- [ ] `migration_partial_success_total` 메트릭 추가
- [ ] 구조화 로그 필드 추가 (`cleaned_count`, `evicted_token_prefix`, `attempt`, `batch_num`)

### 10. 테스트
- [ ] 세션 단위 테스트 (`internal/web/server_test.go`)
  - [x] `purgeExpired`: 만료 세션 삭제 검증
  - [x] `evictOldest`: maxSessions 초과 시 가장 오래된 세션 삭제 검증
  - [x] `createSession`: `ExpiresAt` 값 검증
  - [x] `startCleanupLoop` 종료 검증
- [x] 입력 검증 단위 테스트 (`internal/config/config_test.go`)
  - [x] 경계값 이하/이상/경계값 정상 통과 검증
- [x] 재시도 단위 테스트 (`internal/migration/retry_test.go` 신규)
  - [x] 1회 실패 후 성공, MaxAttempts 소진, Recoverable=false, ctx.Cancel 시나리오
- [ ] skip_batch 통합 테스트 (`internal/migration/direct_test.go` 확장)
  - [ ] 3배치 중 2번째 오류 + skip_batch → partial_success
  - [ ] 3배치 중 2번째 오류 + fail_fast → failed
- [ ] SQL 인젝션 방어 테스트 (`internal/db/db_test.go`)
  - [x] 특수문자 테이블명 QuoteIdentifier 적용 검증
- [ ] `go test ./...` 전량 통과

### 11. CLI/릴리즈
- [x] `--on-error` 플래그 도움말 및 자동완성(zsh/fish/bash) 업데이트
- [ ] feature flag (`DBM_V20_*`) 기반 점진 배포 설정
- [x] README 업데이트 (신규 플래그, 환경변수, 오류 처리 정책 설명)
