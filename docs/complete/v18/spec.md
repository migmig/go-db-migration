# 기술 사양서 (Technical Specifications) - v18

## 1. 아키텍처 개요
v18은 웹 UI의 "테이블 단위 마이그레이션 운영 경험"을 개선하기 위해,
테이블별 실행 상태/이력 조회 기능을 도입하고 목록 필터링을 강화한다.

핵심 목표는 아래 3가지다.

1. 이미 완료된 대상(`성공`)을 즉시 제외할 수 있는 필터 제공
2. 테이블별 최근 상태/시간/실행 횟수 노출
3. 실패 항목 중심의 재시도 흐름 단축

---

## 2. 도메인 모델

### 2.1 상태(enum) 정의
- `not_started`: 이력 없음(미실행)
- `running`: 현재 실행 중
- `success`: 최근 실행 성공
- `failed`: 최근 실행 실패

> UI 노출 라벨 매핑
> - `not_started` -> `미실행`
> - `running` -> `진행중`
> - `success` -> `성공`
> - `failed` -> `실패`

### 2.2 테이블 요약 모델
```go
type TableMigrationSummary struct {
    TableName      string    `json:"table_name"`
    Status         string    `json:"status"` // not_started|running|success|failed
    LastStartedAt  time.Time `json:"last_started_at,omitempty"`
    LastFinishedAt time.Time `json:"last_finished_at,omitempty"`
    DurationMs     int64     `json:"duration_ms,omitempty"`
    RunCount       int64     `json:"run_count"`
    LastError      string    `json:"last_error,omitempty"`
}
```

### 2.3 이력 상세 모델
```go
type TableMigrationHistory struct {
    RunID         string    `json:"run_id"`
    TableName     string    `json:"table_name"`
    Status        string    `json:"status"`
    StartedAt     time.Time `json:"started_at"`
    FinishedAt    time.Time `json:"finished_at,omitempty"`
    DurationMs    int64     `json:"duration_ms,omitempty"`
    RowsProcessed int64     `json:"rows_processed,omitempty"`
    ErrorMessage  string    `json:"error_message,omitempty"`
}
```

---

## 3. 저장소/쿼리 설계

### 3.1 이력 저장소
기존 마이그레이션 실행 단위 로그에 테이블 단위 레코드를 추가/활용한다.
필수 필드:
- `run_id`
- `table_name`
- `status`
- `started_at`
- `finished_at`
- `duration_ms`
- `rows_processed`(optional)
- `error_message`(optional)

### 3.2 조회 쿼리
- 목록 조회: 테이블별 최신 1건 + 누적 run_count 집계
- 상세 조회: 특정 `table_name` 기준 최신 N건 내림차순

### 3.3 인덱스(권장)
- `(table_name, started_at DESC)`
- `(status, finished_at DESC)`

---

## 4. API 계약

### 4.1 목록 조회 API
`GET /api/migrations/tables`

쿼리 파라미터:
- `status` (optional): `not_started|running|success|failed`
- `exclude_success` (optional, bool)
- `search` (optional): 테이블명 prefix/contains
- `sort` (optional): `table_name|status|last_finished_at`
- `order` (optional): `asc|desc`
- `page`, `page_size` (optional)

응답(예시):
```json
{
  "items": [
    {
      "table_name": "users",
      "status": "failed",
      "last_started_at": "2026-03-17T09:00:00Z",
      "last_finished_at": "2026-03-17T09:00:10Z",
      "duration_ms": 10000,
      "run_count": 3,
      "last_error": "duplicate key"
    }
  ],
  "total": 1
}
```

### 4.2 상세 이력 API
`GET /api/migrations/tables/{tableName}/history?limit=20`

응답:
```json
{
  "table_name": "users",
  "items": [
    {
      "run_id": "run-20260317-01",
      "status": "failed",
      "started_at": "2026-03-17T09:00:00Z",
      "finished_at": "2026-03-17T09:00:10Z",
      "duration_ms": 10000,
      "rows_processed": 120,
      "error_message": "duplicate key"
    }
  ]
}
```

### 4.3 실패 재시도 API(연계)
기존 재시도 API를 사용하되, UI에서 현재 선택 테이블 context를 전달한다.
- `POST /api/migrate/retry`
- body에 `table_name`(또는 `tables[]`) 지원 확장 검토

---

## 5. 웹 UI 설계

### 5.1 목록 상단 컨트롤
- 상태 필터 드롭다운 (`전체/미실행/진행중/성공/실패`)
- 빠른 토글: `성공 제외`
- 검색창: 테이블명 검색
- 정렬 선택: 최근 실행 시각/상태/테이블명

### 5.2 목록 테이블 컬럼
- 테이블명
- 상태 뱃지
- 최근 실행 시각
- 소요 시간
- 실행 횟수
- 액션(상세, 실패 시 재시도)

### 5.3 상세 이력 패널
- 테이블 클릭 시 오른쪽 패널 또는 모달 표시
- 최근 N건 타임라인/리스트
- 실패 건은 오류 요약 강조

### 5.4 빈 상태/오류 상태
- 필터 결과 없음: 안내 문구 + 필터 초기화 액션
- API 에러: 에러 배너 + 재시도 버튼
- 로딩: 스켈레톤 행 표시

---

## 6. 로깅/관측성
- 주요 이벤트 구조화 로그(`slog`) 필드
  - `table_name`
  - `status`
  - `run_id`
  - `duration_ms`
- 신규 메트릭
  - `ui_table_filter_usage_total{filter=...}`
  - `migration_table_retry_total`
  - `migration_table_status_total{status=...}`

---

## 7. 성능/확장성
- 서버 사이드 필터/정렬/페이징 기본 적용
- 대량 테이블(1k+)에서도 첫 화면 응답 1초 내 목표
- 상세 이력은 lazy-load로 최초 요청 시에만 조회

---

## 8. 오류 처리 및 검증
- 허용되지 않은 `status/sort/order` 값은 400
- 존재하지 않는 테이블 상세 조회는 404
- `exclude_success=true` + `status=success` 동시 입력 시
  - 정책: `status` 우선 또는 400 중 하나로 명확화(구현 시 고정)

---

## 9. 테스트 전략

### 9.1 백엔드
- 단위 테스트
  - 상태 매핑(`not_started/running/success/failed`) 검증
  - 필터 조합 쿼리 빌더 검증
- 통합 테스트
  - `exclude_success=true` 동작 검증
  - `status=failed` + 정렬/검색 조합 검증
  - 상세 이력 limit 동작 검증

### 9.2 프론트엔드
- 컴포넌트 테스트
  - 상태 필터/토글 상호작용
  - 실패 항목 재시도 버튼 노출 조건
- E2E
  - 실패만 보기 -> 상세 이력 확인 -> 재시도 요청 흐름

---

## 10. 롤아웃
1. API read 경로(목록/상세) 우선 배포
2. UI 필터/이력 패널 점진 활성화(feature flag)
3. 실패 재시도 UX 활성화
4. KPI(재실행 감소/재처리 시간 단축) 관찰 후 기본 on
