# 운영자 가이드 - v19 Pre-check Row Count

## 1. 개요

v19 pre-check 기능은 Direct Insert 마이그레이션을 시작하기 전, Oracle 원본과 대상 DB의 테이블별 행 수를 자동으로 비교하여 실제 전송이 필요한 테이블만 선별합니다.

**주요 이점:**
- 행 수가 동일한 테이블의 불필요한 재전송 제거
- 마이그레이션 전 사전 확인으로 운영자 의사결정 지원
- 정책 기반 자동화로 수동 판단 작업 감소

---

## 2. Decision 종류

| Decision | 설명 | 기본 행동 |
|---|---|---|
| `transfer_required` | 원본/대상 행 수 불일치, 대상 테이블 미존재 | 전송 대상에 포함 |
| `skip_candidate` | 원본과 대상 행 수 동일 | 전송 생략 후보 |
| `count_check_failed` | 카운트 조회 실패(권한 부족, 타임아웃 등) | 정책에 따라 처리 |

---

## 3. Policy 선택 기준

### `strict` (기본값)
- `count_check_failed`가 1건이라도 있으면 해당 항목의 전송을 차단하고 경고를 표시합니다.
- **권장 상황:** 데이터 정합성이 중요한 운영 환경, 실패 원인 확인 후 재실행이 가능한 경우
- **주의:** `count_check_failed` 테이블이 있어도 전체 마이그레이션을 중단하지 않습니다. CLI `--precheck-row-count` 사용 시 `plan.Blocked=true`이면 실행이 중단됩니다.

### `best_effort`
- `count_check_failed` 항목은 경고 로그를 남기고 전송을 진행합니다.
- **권장 상황:** 일부 테이블의 카운트 조회 실패가 예상되거나, 실패해도 전송을 시도해야 하는 경우(예: 대상 DB 권한이 불완전한 경우)

### `skip_equal_rows`
- 행 수가 동일한 `skip_candidate` 테이블은 자동으로 전송 목록에서 제외됩니다.
- **권장 상황:** 증분 재실행 시나리오, 이미 동기화된 테이블을 건너뛰고 싶은 경우

---

## 4. 실패 대응 방법

### `count_check_failed` 발생 시

1. **원인 확인:** 응답의 `reason` 필드 확인
   - `"source count failed: timeout"` → Oracle 연결 타임아웃, 쿼리 성능 문제
   - `"permission denied"` → 대상 DB 계정 권한 부족
   - `"table not found"` → 대상 테이블 미생성

2. **대응 방법:**
   - 타임아웃: `timeoutMs` 파라미터를 늘리거나 DB 서버 상태 점검
   - 권한 부족: 대상 DB 계정에 `SELECT` 권한 부여 확인
   - 테이블 미존재: DDL 실행 후 재시도, 또는 `best_effort` 정책으로 전송 허용

3. **임시 우회:** policy를 `best_effort`로 변경하면 실패 항목도 전송 대상에 포함됩니다.

---

## 5. Web UI 사용법

1. 소스/타깃 연결 설정 후 테이블 선택
2. "Migration Options" 섹션에서 policy 선택 (`strict` / `best_effort` / `skip_equal_rows`)
3. **"Run Pre-check"** 버튼 클릭
4. 결과 확인:
   - 상단 요약 카드: 전체 / Transfer Required / Skip / Failed 건수
   - 필터 탭으로 decision별 테이블 목록 전환
   - `count_check_failed` 항목은 ⚠ 경고 아이콘과 함께 reason 표시
5. 결과 확인 후 **"Start Migration"** 클릭

> **참고:** pre-check 결과와 마이그레이션 실행은 별개입니다. `usePrecheckResults: true`를 API에 전달하면 pre-check 결과에 따라 전송 대상 테이블을 자동 필터링합니다.

---

## 6. CLI 사용법

### 기본 사전 점검 (dry-run)
```bash
./dbmigrator \
  -url oracle-host:1521/orcl \
  -user scott \
  -password tiger \
  -target-url postgres://user:pass@pg-host/mydb \
  -tables EMP,DEPT,PROJ \
  -precheck-row-count \
  -dry-run
```
dry-run 모드에서 pre-check 결과만 로그로 출력하고 실제 마이그레이션은 하지 않습니다.

### skip_equal_rows 정책으로 실행
```bash
./dbmigrator \
  -url oracle-host:1521/orcl \
  -user scott -password tiger \
  -target-url postgres://user:pass@pg-host/mydb \
  -tables EMP,DEPT,PROJ \
  -precheck-row-count \
  -precheck-policy skip_equal_rows
```
행 수가 동일한 테이블을 자동 제외하고 불일치 테이블만 마이그레이션합니다.

### 특정 decision만 출력 보기
```bash
./dbmigrator ... \
  -precheck-row-count \
  -precheck-filter transfer_required \
  -dry-run
```

---

## 7. API 사용 예시

### pre-check 실행
```http
POST /api/migrations/precheck
Content-Type: application/json

{
  "oracleUrl": "oracle://host:1521/orcl",
  "username": "scott",
  "password": "tiger",
  "tables": ["EMP", "DEPT", "PROJ"],
  "targetDb": "postgres",
  "targetUrl": "postgres://user:pass@pg-host/mydb",
  "policy": "skip_equal_rows",
  "concurrency": 4,
  "timeoutMs": 5000
}
```

### 결과 조회 (transfer_required만)
```http
GET /api/migrations/precheck/results?decision=transfer_required&page=1&page_size=50
```

### pre-check 결과 기반 마이그레이션 실행
```http
POST /api/migrate
Content-Type: application/json

{
  "oracleUrl": "...",
  "tables": ["EMP", "DEPT", "PROJ"],
  "usePrecheckResults": true,
  "precheckPolicy": "skip_equal_rows",
  ...
}
```

---

## 8. Feature Flag

`DBM_V19_PRECHECK=false` 환경변수로 pre-check 기능 전체를 비활성화할 수 있습니다.

```bash
export DBM_V19_PRECHECK=false
./dbmigrator -web
```

비활성화 시:
- `/api/migrations/precheck` 및 `/api/migrations/precheck/results` 엔드포인트가 404 반환
- Web UI에서 "Run Pre-check" 섹션이 숨겨집니다
- `/api/meta`의 `features.precheckRowCount`가 `false`로 반환됩니다

---

## 9. 성능 고려 사항

| 항목 | 기본값 | 권장 조정 |
|---|---|---|
| `concurrency` | 4 | 대량 테이블(500+): 8~16 |
| `timeoutMs` | 5000ms | 느린 DB: 10000~30000ms |

- `COUNT(*)` 쿼리는 테이블 크기에 따라 시간이 걸릴 수 있습니다.
- 수백만 행 이상의 테이블이 많을 경우 타임아웃을 조정하세요.
- `concurrency`가 너무 높으면 DB 연결 부하가 증가할 수 있습니다.

---

## 10. 로그 필드 참조

pre-check 실행 시 다음 구조화 로그가 기록됩니다:

```json
{
  "level": "INFO",
  "msg": "precheck result",
  "table_name": "EMP",
  "source_row_count": 1000,
  "target_row_count": 950,
  "decision": "transfer_required",
  "policy": "strict",
  "reason": ""
}
```

모니터링 메트릭 (`/api/monitoring/metrics`):
- `precheck.runTotal` - 총 pre-check 실행 횟수
- `precheck.tablesTotal` - 총 점검 테이블 수(누계)
- `precheck.transferRequiredTotal` - transfer_required 누계
- `precheck.skipCandidateTotal` - skip_candidate 누계
- `precheck.countCheckFailedTotal` - count_check_failed 누계
