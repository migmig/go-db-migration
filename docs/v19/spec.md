# 기술 사양서 (Technical Specifications) - v19

## 1. 아키텍처 개요
v19의 목표는 **Direct Insert 실행 전 사전 점검(pre-check)** 단계에서 Oracle 원본과 대상 DB의 테이블별 행 수를 비교하여,
실제 전송이 필요한 테이블만 선별하는 것이다.

핵심 설계 포인트:
1. 사전 비교 결과를 표준 모델로 수집/저장
2. 비교 결과(`decision`) 기반 필터링 일관성(UI/CLI/API)
3. 정책(`strict`, `best_effort`, `skip_equal_rows`)에 따른 실행 제어

---

## 2. 도메인 모델

### 2.1 Decision(enum) 정의
- `transfer_required`
  - 행 수 불일치, 대상 테이블 미존재, 정책상 강제 전송 등
- `skip_candidate`
  - 행 수 동일하여 전송 생략 후보
- `count_check_failed`
  - 카운트 조회 실패(권한/타임아웃/네트워크 오류 등)

### 2.2 Policy(enum) 정의
- `strict` (기본)
  - `count_check_failed` 발생 시 해당 테이블 전송 차단(또는 전체 중단)
- `best_effort`
  - `count_check_failed`를 경고 처리하고 전송 진행
- `skip_equal_rows`
  - `skip_candidate` 자동 제외, 나머지만 전송

### 2.3 사전 비교 결과 모델
```go
type PrecheckRowCountResult struct {
    TableName       string     `json:"table_name"`
    SourceRowCount  *int64     `json:"source_row_count,omitempty"`
    TargetRowCount  *int64     `json:"target_row_count,omitempty"`
    CountDelta      *int64     `json:"count_delta,omitempty"` // source - target
    Decision        string     `json:"decision"`              // transfer_required|skip_candidate|count_check_failed
    Policy          string     `json:"policy"`
    Reason          string     `json:"reason,omitempty"`
    CheckedAt       time.Time  `json:"checked_at"`
}
```

### 2.4 실행 전 요약 모델
```go
type PrecheckSummary struct {
    TotalTables             int `json:"total_tables"`
    TransferRequiredCount   int `json:"transfer_required_count"`
    SkipCandidateCount      int `json:"skip_candidate_count"`
    CountCheckFailedCount   int `json:"count_check_failed_count"`
}
```

---

## 3. 백엔드 설계

### 3.1 구성요소
- `internal/migration/precheck.go` (신규)
  - 테이블 목록 입력을 받아 병렬로 원본/대상 COUNT 수행
  - 결과 모델 생성 및 decision 판정
- `internal/migration/policy.go` (신규 또는 기존 확장)
  - policy 기반 실행 가능 여부/제외 목록 산출
- `internal/db` 계층 확장
  - 각 dialect에 맞는 COUNT 쿼리 유틸 제공

### 3.2 COUNT 쿼리 규칙
- 기본 쿼리: `SELECT COUNT(*) FROM <qualified_table>`
- 식별자 quoting은 dialect별 규칙 준수
- 타임아웃/취소는 `context.Context` 기반 통합 제어

### 3.3 병렬 처리/성능
- worker pool 기반 병렬 카운트 수행
- 설정값:
  - `precheck_concurrency` (기본 4)
  - `precheck_timeout_ms` (테이블별)
- 과도한 DB 부하 방지를 위해 상한 설정

### 3.4 decision 판정 알고리즘
1. source/target COUNT 모두 성공
   - 동일: `skip_candidate`
   - 불일치: `transfer_required`
2. 대상 테이블 없음/조회 불가
   - `transfer_required` + reason
3. 카운트 실패
   - `count_check_failed` + reason

### 3.5 policy 적용
- `strict`
  - `count_check_failed`가 존재하면 실행 계획에서 차단 상태 마킹
- `best_effort`
  - 실패 항목도 전송 대상으로 포함하되 warning 로깅
- `skip_equal_rows`
  - `skip_candidate`를 실행 목록에서 자동 제외

---

## 4. API 계약

### 4.1 사전 점검 실행
`POST /api/migrations/precheck`

요청:
```json
{
  "tables": ["EMP", "DEPT"],
  "policy": "strict"
}
```

응답:
```json
{
  "summary": {
    "total_tables": 2,
    "transfer_required_count": 1,
    "skip_candidate_count": 1,
    "count_check_failed_count": 0
  },
  "items": [
    {
      "table_name": "EMP",
      "source_row_count": 100,
      "target_row_count": 100,
      "count_delta": 0,
      "decision": "skip_candidate",
      "policy": "strict",
      "checked_at": "2026-03-17T12:00:00Z"
    }
  ]
}
```

### 4.2 사전 점검 결과 조회(필터)
`GET /api/migrations/precheck/results?decision=transfer_required&policy=strict&page=1&page_size=50`

쿼리:
- `decision` (optional): `transfer_required|skip_candidate|count_check_failed`
- `policy` (optional)
- `search` (optional): 테이블명 검색
- `page`, `page_size` (optional)

### 4.3 실행 연계
`POST /api/migrate`
- `use_precheck=true`일 때, pre-check 결과 및 policy를 기반으로 실제 전송 대상 테이블을 계산

---

## 5. UI/CLI 설계

### 5.1 웹 UI
- 목록 컬럼 추가
  - 원본 건수 / 대상 건수 / 차이 / 판정
- 기본 필터: `transfer_required`
- 필터 탭
  - `all`, `transfer_required`, `skip_candidate`, `count_check_failed`
- `count_check_failed`는 경고 아이콘 + reason 툴팁

### 5.2 CLI
- 신규 플래그(초안)
  - `--precheck-row-count`
  - `--precheck-policy=strict|best_effort|skip_equal_rows`
  - `--precheck-filter=all|transfer_required|skip_candidate|count_check_failed`
- dry-run 시 pre-check 결과만 출력 가능

---

## 6. 로깅/관측성
- 구조화 로그(`slog`) 필수 필드
  - `table_name`, `source_row_count`, `target_row_count`, `decision`, `policy`, `reason`
- 메트릭
  - `precheck_tables_total`
  - `precheck_decision_total{decision=...}`
  - `precheck_failed_total{reason=...}`
  - `precheck_duration_ms`

---

## 7. 오류 처리
- 허용되지 않은 policy/decision 파라미터: 400
- 대상 테이블 메타 조회 실패: `count_check_failed`로 기록 후 policy에 따라 제어
- pre-check 타임아웃: reason=`timeout`
- 소스/대상 연결 실패: pre-check 단계 자체 실패 처리(500)

---

## 8. 테스트 전략

### 8.1 단위 테스트
- decision 판정 함수 테스트
  - 동일/불일치/대상없음/오류 케이스
- policy 적용 테스트
  - `strict`, `best_effort`, `skip_equal_rows`별 실행 목록 기대값 검증

### 8.2 통합 테스트
- pre-check API 응답 스키마/요약 집계 검증
- decision 필터 쿼리 검증
- `use_precheck=true` 실행 시 실제 마이그레이션 대상 축소 검증

### 8.3 성능 테스트
- 1,000개 테이블 기준 pre-check 처리 시간 측정
- `precheck_concurrency` 변화에 따른 DB 부하/처리량 비교

---

## 9. 롤아웃 계획
1. 서버 pre-check 엔진 및 read API 먼저 배포
2. feature flag(`DBM_V19_PRECHECK`) 하에 UI 필터/요약 활성화
3. CLI 플래그 공개 및 운영 가이드 업데이트
4. KPI 관찰 후 기본값 정책(`strict`) 유지 여부 확정

---

## 10. 오픈 이슈
- 동일 row count이나 데이터 내용 불일치 케이스에 대한 2차 검증(checksum) 도입 시점
- 대상 DB별 COUNT 성능 편차를 완화할 힌트/통계 활용 전략
- pre-check 결과 저장 보존 기간(인메모리/DB persist) 정책 확정
