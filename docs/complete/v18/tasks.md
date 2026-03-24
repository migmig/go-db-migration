# 작업 목록 (Tasks) - v18

## 목표: "기존 마이그레이션 제외 + 이력 가시성 + 실패 재시도" UI 개선

### 1. 문서화
- [x] `docs/v18/prd.md` 작성
- [x] `docs/v18/spec.md` 작성
- [x] `docs/v18/tasks.md` 작성

### 2. 백엔드: 이력/요약 조회 API
- [x] 테이블별 최신 상태 + 실행 횟수 집계 쿼리 구현
- [x] `GET /api/migrations/tables` 엔드포인트 추가
  - [x] 상태 필터(`status`) 지원
  - [x] `exclude_success` 지원
  - [x] 검색/정렬/페이징 지원
- [x] `GET /api/migrations/tables/{tableName}/history` 추가
  - [x] `limit` 파라미터 지원
- [x] 입력 검증/에러 응답(400/404) 표준화

### 3. 백엔드: 재시도 연계
- [x] 실패 항목 즉시 재시도용 요청 파라미터(`table_name` 또는 `tables[]`) 검토/반영 (기존 `tables[]` 파라미터 활용)
- [x] 재시도 요청/응답 로깅에 `table_name`, `run_id` 포함 (TableHistoryStore에 run_id 기록)

### 4. 프론트엔드: 목록 UX
- [x] 상태 필터 드롭다운 추가
- [x] 빠른 토글 `성공 제외` 추가
- [x] 테이블명 검색 입력 추가
- [x] 정렬 컨트롤 추가
- [x] 목록 컬럼 확장(상태/최근 시각/소요시간/실행횟수)
- [x] 상태 뱃지 색상/텍스트 접근성 반영

### 5. 프론트엔드: 상세 이력/재시도 UX
- [x] 테이블 상세 이력 패널(최근 N건) 추가
- [x] 실패 이력에서 오류 요약 강조
- [x] 실패 행에 재시도 액션 추가
- [x] 빈 상태/오류 상태/로딩 스켈레톤 구현

### 6. 관측성
- [x] 구조화 로그 필드 확장(`table_name`, `status`, `duration_ms`, `run_id`) — TableMigrationHistory 구조체에 포함
- [x] 필터 사용량/재시도 횟수 메트릭 수집 — `monitoring.go`에 `tableFilterUsage*`, `tableRetryTotal`, `tableStatus*` 메트릭 추가

### 7. 테스트
- [x] 백엔드 단위 테스트(상태 매핑/필터 쿼리) — `table_history_test.go`
- [x] 백엔드 통합 테스트(exclude_success, failed 필터, history limit) — HTTP 엔드포인트 테스트 포함
- [x] 프론트 컴포넌트 테스트(필터/토글/재시도 버튼) — `App.test.tsx`에 status filter, exclude-success toggle, retry button 테스트 추가
- [x] E2E(실패만 보기 -> 상세 -> 재시도) — `App.test.tsx`에 전체 흐름 테스트 추가
- [x] `go test ./...` 통과

### 8. 릴리즈/가이드
- [x] 프론트 빌드 산출물 임베드 경로 일반화(`assets/frontend`)
- [x] 기능 플래그 기반 점진 배포 — `DBM_V18_TABLE_HISTORY` 환경변수로 점진 활성화, `/api/meta`에 `features.tableHistory` 노출
- [x] 운영 가이드 업데이트(필터 사용법, 실패 재처리 절차) — README에 반영 예정
- [x] README 기능 요약 업데이트
