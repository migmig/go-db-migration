# 작업 목록 (Tasks) - v22

## 목표: 타겟 DB PostgreSQL 단일화 + 소스-타겟 테이블 비교 UI

### 1. 문서화
- [x] `docs/v22/prd.md` 작성
- [x] `docs/v22/spec.md` 작성
- [x] `docs/v22/tasks.md` 작성

---

### 2. FR-0: 타겟 DB 단일화 — 백엔드

#### 2.1 공통 헬퍼 (`internal/web/server.go`)
- [x] `requirePostgres(c *gin.Context, targetDB string) bool` 헬퍼 함수 구현

#### 2.2 `testTargetConnection` 정리 (`internal/web/server.go`)
- [x] `ShouldBindJSON` 직후 `requirePostgres` 호출 추가
- [x] `testTargetRequest.TargetDB` 필드를 선택적으로 변경
- [x] postgres 이외 타겟 연결 `else` 블록 제거 (pgPool 경로만 유지)

#### 2.3 `startMigrationHandler` 정리 (`internal/web/server.go`)
- [x] `validateMigrationRequest` 직후 `requirePostgres` 호출 추가
- [x] non-postgres 타겟 연결 `else` 블록 제거 (pgPool 경로만 유지)

#### 2.4 `precheckHandler` 정리 (`internal/web/precheck_handler.go`)
- [x] `ShouldBindJSON` 직후 `requirePostgres` 호출 추가
- [x] non-postgres 타겟 연결 분기 제거 (pgPool/postgres 경로만 유지)

---

### 3. FR-1: POST /api/target-tables 엔드포인트

#### 3.1 DB 레이어 (`internal/db/db.go`)
- [x] `FetchTargetTables(ctx context.Context, pool PGPool, schema string) ([]string, error)` 구현
  - [x] `information_schema.tables` 쿼리 (`table_schema=$1`, `table_type='BASE TABLE'`, `ORDER BY table_name`)
  - [x] `rows.Close()` defer 처리
  - [x] `rows.Err()` 반환 처리

#### 3.2 핸들러 (`internal/web/server.go`)
- [x] `targetTablesRequest` / `targetTablesResponse` 구조체 정의
- [x] `targetTablesHandler` 함수 구현
  - [x] 10초 `context.WithTimeout` 적용
  - [x] `db.ConnectPostgres` 연결 및 `defer pool.Close()`
  - [x] `isPermissionError(err)` 헬퍼로 403 / 500 구분
  - [x] 응답에 `fetchedAt` (RFC3339 UTC) 포함
- [x] `isPermissionError(err error) bool` 헬퍼 구현 (SQLSTATE `42501` / `"permission denied"` 감지)
- [x] `protected.POST("/target-tables", targetTablesHandler)` 라우트 등록

---

### 4. FR-0: 타겟 DB 단일화 — 프론트엔드 (`frontend/src/app/App.tsx`)

#### 4.1 TargetState 타입 변경
- [x] `TargetState`에서 `targetDb: string` 필드 제거
- [x] 신규 타입 추가: `TargetTableEntry`, `CompareState`, `CompareFilter`

#### 4.2 초기값 및 상태 업데이트 정리
- [x] `useState<TargetState>` 초기값에서 `targetDb` 제거
- [x] `useEffect` deps에서 `target.targetDb` 제거
- [x] `applyCredential()` 내 `targetDb: item.dbType || "postgres"` 라인 제거
- [x] `replayHistory()` / setTarget 처리에서 `targetDb` 제거

#### 4.3 API 호출부 정리
- [x] `testTarget()` — 요청 바디에서 `targetDb` 필드 제거
- [x] precheck API 요청 바디에서 `targetDb` 필드 제거
- [x] `startMigration()` — 요청 바디에서 `targetDb` 필드 제거

#### 4.4 드롭다운 → 고정 레이블 교체
- [x] Target DB `<select>` 컴포넌트를 PostgreSQL 고정 `<span>` 레이블로 교체

---

### 5. FR-2: 타겟 테이블 조회 UI (`frontend/src/app/App.tsx`)

- [x] `CompareState` 타입 정의 및 `compareState` 상태 추가
- [x] `compareFilter`, `compareSearch` 상태 추가
- [x] `fetchTargetTables()` async 함수 구현
  - [x] `POST /api/target-tables` 호출
  - [x] `busy` 플래그 on/off 처리
  - [x] 오류 시 `error` 필드 설정
- [x] 타겟 섹션 하단에 "타겟 테이블 조회" 버튼 추가
  - [x] 비활성화 조건: `compareState.busy || migrationBusy || !target.targetUrl || !target.schema`
  - [x] 조회 중 / 새로고침 텍스트 분기
- [x] 조회 성공 시 테이블 수 배지 표시
- [x] `fetchedAt` 마지막 조회 시각 표시
- [x] 인라인 오류 메시지 표시

---

### 6. FR-3: 소스-타겟 비교 패널 (`frontend/src/app/App.tsx`)

- [x] `compareEntries` useMemo 구현
  - [x] 소문자 정규화 후 `sourceSet` / `targetSet` 교차 비교
  - [x] 세 카테고리(`source_only` / `both` / `target_only`) 분류
  - [x] `precheckItems`에서 `sourceRowCount` / `targetRowCount` 조회 합산
  - [x] 정렬: 카테고리 순 → 테이블명 알파벳 순
- [x] 비교 패널 JSX 구현 (`compareEntries.length > 0` 조건부 렌더링)
  - [x] `<details>` / `<summary>` 래퍼로 접힘 가능하게 구현
  - [x] 요약 카드 × 3 (카테고리별 테이블 수)
  - [x] 카테고리 탭 필터 (`전체` / `소스만` / `양쪽` / `타겟만`)
  - [x] 테이블명 검색 입력
  - [x] 비교 테이블 컬럼: 테이블명 / 소스 / 타겟 / 소스 행 수 / 타겟 행 수 / 상태
  - [x] 카테고리 배지 색상 클래스 적용 (blue / emerald / amber)
  - [x] `row_diff` 배지 조건 (`both` + 행 수 불일치, orange)
  - [x] pre-check 미실행 시 안내 문구 표시
  - [x] 모바일 뷰 가로 스크롤 처리 (`overflow-x-auto`)

---

### 7. FR-4: 비교 기반 빠른 선택 (`frontend/src/app/App.tsx`)

- [x] `selectByCategory(category: TargetTableEntry["category"])` 함수 구현
  - [x] 대문자 복원(Oracle 원본 테이블명) 후 `selectedTables`에 추가
- [x] 테이블 선택 섹션에 버튼 추가 (비교 결과 있을 때만 표시)
  - [x] "소스에만 있는 테이블 선택" → `selectByCategory("source_only")`
  - [x] "양쪽에 있는 테이블 선택" → `selectByCategory("both")`

---

### 8. FR-5: pre-check 행 수 연동

- [x] `compareEntries` memo에서 `precheckItems` 기반 행 수 합산 (FR-3에서 구현)
- [x] `isRowDiff` 조건 처리: `both` + `sourceRowCount !== targetRowCount` → `row_diff` 배지

---

### 9. i18n

- [x] 비교 패널 신규 `tr()` 문자열 전체 추가 (spec.md 6절 목록 기준)
  - [x] `"Fetch Target Tables"` / `"타겟 테이블 조회"`
  - [x] `"Fetching..."` / `"조회 중..."`
  - [x] `"tables in target"` / `"개 타겟 테이블"`
  - [x] `"as of"` / `"기준 시각"`
  - [x] `"Refresh"` / `"새로고침"`
  - [x] `"Source vs Target Comparison"` / `"소스-타겟 비교"`
  - [x] `"Source only"` / `"소스에만"`, `"Both"` / `"양쪽"`, `"Target only"` / `"타겟에만"`
  - [x] `"Select source-only"` / `"소스에만 있는 테이블 선택"`
  - [x] `"Select both"` / `"양쪽에 있는 테이블 선택"`
  - [x] `"Row diff"` / `"행 수 불일치"`
  - [x] `"Run pre-check to see row counts"` / `"Pre-check 실행 후 행 수가 표시됩니다"`

---

### 10. 테스트

#### 10.1 백엔드 핸들러 (`internal/web/server_test.go`)
- [x] `POST /api/target-tables` — `targetUrl` 누락: 400 반환
- [x] `POST /api/target-tables` — `schema` 누락: 400 반환
- [x] `POST /api/target-tables` — JSON 파싱 실패: 400 반환
- [x] `POST /api/test-target` — `targetDb: "mysql"` 등 4종 전달 시 400 + v22 에러 메시지 검증
- [x] `POST /api/migrate` — `targetDb: "mysql"` 전달 시 400 반환
- [x] `requirePostgres` — 빈 targetDb는 통과하는지 검증

#### 10.2 프론트엔드
- [ ] `compareEntries` 메모: source-only / both / target-only 분류 검증
- [ ] 대소문자 정규화: 소스 `"USERS"` + 타겟 `"users"` → `both` 확인
- [ ] `selectByCategory("source_only")`: 해당 테이블만 추가 확인
- [ ] `isRowDiff` 조건 검증

#### 10.3 최종 확인
- [ ] `go test ./...` 전량 통과
- [ ] `npm run build` 빌드 오류 없음

---

### 11. 릴리즈 노트
- [x] README에 breaking change 명시 (타겟 DB 단일화, MariaDB/MySQL/MSSQL/SQLite 지원 종료)
- [x] CLI(`main.go`)에서 non-postgres targetDb 진입 시 오류 종료 처리 (`requirePostgres` 동등 로직)
- [x] `internal/db/db.go`에서 불필요한 드라이버 imports 제거 (`go-sql-driver/mysql`, `go-sqlite3`, `go-mssqldb`)
