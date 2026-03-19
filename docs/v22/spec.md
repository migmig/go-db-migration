# 기술 사양서 (Technical Specifications) - v22

## 1. 아키텍처 개요

v22는 **두 가지 축**으로 구성된다.

1. **타겟 DB 단일화**: 기존에 `postgres`, `mysql`, `mariadb`, `mssql`, `sqlite` 다섯 가지였던 타겟 DB 옵션을 `postgres`로 고정하고, 관련 분기 코드와 UI 선택 요소를 제거한다.
2. **소스-타겟 테이블 비교 UI**: `POST /api/target-tables` 엔드포인트를 신설하고, 프론트엔드에 비교 패널·빠른 선택·pre-check 연동을 추가한다.

핵심 설계 원칙:
- 기존 마이그레이션 엔진(DDL/DML 생성) 변경 없음
- `postgres` 이외의 `targetDb` 값을 받는 모든 API 진입점에서 즉시 `400 Bad Request` 반환
- 프론트엔드 `TargetState.targetDb` 필드 제거 → 타입 단순화

---

## 2. 도메인 모델 변경

### 2.1 프론트엔드 TargetState 타입 축소

**파일**: `frontend/src/app/App.tsx`

기존:
```typescript
type TargetState = {
  mode: "file" | "direct";
  targetDb: string;       // ← 제거
  targetUrl: string;
  schema: string;
};
```

변경 후:
```typescript
type TargetState = {
  mode: "file" | "direct";
  targetUrl: string;
  schema: string;
};
```

`targetDb` 필드가 제거되므로 `targetDb`를 참조하는 모든 코드 경로를 `"postgres"` 리터럴로 대체한다.

### 2.2 신규 프론트엔드 타입

```typescript
type TargetTableEntry = {
  name: string;
  inSource: boolean;
  inTarget: boolean;
  category: "source_only" | "both" | "target_only";
  sourceRowCount: number | null;
  targetRowCount: number | null;
};

type CompareState = {
  targetTables: string[];
  fetchedAt: string | null;
  busy: boolean;
  error: string | null;
};
```

### 2.3 신규 백엔드 요청/응답 구조체

**파일**: `internal/web/server.go`

```go
type targetTablesRequest struct {
    TargetURL string `json:"targetUrl" binding:"required"`
    Schema    string `json:"schema"    binding:"required"`
}

type targetTablesResponse struct {
    Tables    []string `json:"tables"`
    FetchedAt string   `json:"fetchedAt"` // time.RFC3339
}
```

---

## 3. 백엔드 설계

### 3.1 FR-0: 타겟 DB 단일화 — 입력 검증 강화

#### 3.1.1 공통 헬퍼

**파일**: `internal/web/server.go`

```go
// requirePostgres는 targetDb 값이 "postgres"가 아니면 400을 반환하고 false를 돌려준다.
// 핸들러에서 targetDb 검증이 필요한 모든 경로에서 호출한다.
func requirePostgres(c *gin.Context, targetDB string) bool {
    if targetDB != "" && targetDB != "postgres" {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "v22 이후 타겟 DB는 PostgreSQL만 지원합니다 (입력값: " + targetDB + ")",
        })
        return false
    }
    return true
}
```

#### 3.1.2 적용 대상 핸들러

| 핸들러 함수 | 파일 | 적용 위치 |
|---|---|---|
| `testTargetConnection` | `server.go` | `ShouldBindJSON` 직후 |
| `startMigrationHandler` | `server.go` | `validateMigrationRequest` 내부 또는 직후 |
| `precheckHandler` | `precheck_handler.go` | `ShouldBindJSON` 직후 |

#### 3.1.3 startMigrationHandler 분기 단순화

기존의 `targetDBName == "postgres"` 분기에서 `else` 블록(non-postgres 연결 처리)을 제거한다.

```go
// 변경 전
if req.Direct && targetURL != "" {
    if targetDBName == "postgres" {
        pgPool, err = db.ConnectPostgres(...)
        ...
    } else {                              // ← 이 블록 전체 삭제
        targetDB, err = db.ConnectTargetDB(...)
        ...
    }
}

// 변경 후
if req.Direct && targetURL != "" {
    pgPool, err = db.ConnectPostgres(targetURL, req.DBMaxOpen, req.DBMaxIdle, req.DBMaxLife)
    if err != nil { ... }
    defer pgPool.Close()
}
```

`testTargetConnection` 도 동일하게 `else` 블록을 제거하고 pgPool 경로만 유지한다.

#### 3.1.4 precheckHandler 단순화

**파일**: `internal/web/precheck_handler.go`

`targetConn` 분기에서 non-postgres 경로를 제거. `requirePostgres` 호출 후 항상 `db.ConnectPostgres`로 연결한다.

---

### 3.2 FR-1: POST /api/target-tables 핸들러 신설

**파일**: `internal/web/server.go`

#### 3.2.1 라우트 등록

```go
// RunServerWithAuth 내 protected 그룹에 추가
protected.POST("/target-tables", targetTablesHandler)
```

#### 3.2.2 핸들러 구현

```go
func targetTablesHandler(c *gin.Context) {
    var req targetTablesRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
        return
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
    defer cancel()

    pool, err := db.ConnectPostgres(req.TargetURL, 1, 1, 30)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "타겟 DB 연결 실패: " + err.Error()})
        return
    }
    defer pool.Close()

    tables, err := db.FetchTargetTables(ctx, pool, req.Schema)
    if err != nil {
        // information_schema 권한 부족 여부를 구분
        if isPermissionError(err) {
            c.JSON(http.StatusForbidden, gin.H{"error": "테이블 목록 조회 권한이 없습니다: " + err.Error()})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "테이블 목록 조회 실패: " + err.Error()})
        return
    }

    c.JSON(http.StatusOK, targetTablesResponse{
        Tables:    tables,
        FetchedAt: time.Now().UTC().Format(time.RFC3339),
    })
}

// isPermissionError는 PostgreSQL permission denied 오류 여부를 판별한다.
func isPermissionError(err error) bool {
    return strings.Contains(err.Error(), "permission denied") ||
        strings.Contains(err.Error(), "42501") // PostgreSQL SQLSTATE
}
```

#### 3.2.3 db.FetchTargetTables 신설

**파일**: `internal/db/db.go`

```go
// FetchTargetTables는 PostgreSQL information_schema에서 BASE TABLE 목록을 조회한다.
func FetchTargetTables(ctx context.Context, pool PGPool, schema string) ([]string, error) {
    const q = `
        SELECT table_name
        FROM information_schema.tables
        WHERE table_schema = $1
          AND table_type = 'BASE TABLE'
        ORDER BY table_name
    `
    rows, err := pool.Query(ctx, q, schema)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tables []string
    for rows.Next() {
        var name string
        if err := rows.Scan(&name); err != nil {
            return nil, err
        }
        tables = append(tables, name)
    }
    return tables, rows.Err()
}
```

---

## 4. 프론트엔드 설계

### 4.1 FR-0: 타겟 DB 드롭다운 제거

**파일**: `frontend/src/app/App.tsx`

#### 4.1.1 TargetState 초기값 변경

```typescript
// 변경 전
const [target, setTarget] = useState<TargetState>({
  mode: initialTarget.mode ?? "file",
  targetDb: initialTarget.targetDb ?? "postgres",
  targetUrl: initialTarget.targetUrl ?? "",
  schema: "",
});

// 변경 후
const [target, setTarget] = useState<TargetState>({
  mode: initialTarget.mode ?? "file",
  targetUrl: initialTarget.targetUrl ?? "",
  schema: "",
});
```

#### 4.1.2 드롭다운 JSX 교체

```tsx
// 변경 전 — Target DB select
<label className="block text-sm">
  <span>{tr("Target DB", "타깃 DB")}</span>
  <select value={target.targetDb} onChange={...}>
    <option value="postgres">PostgreSQL</option>
    <option value="mysql">MySQL</option>
    <option value="mariadb">MariaDB</option>
    <option value="sqlite">SQLite</option>
    <option value="mssql">MSSQL</option>
  </select>
</label>

// 변경 후 — 고정 레이블
<div className="block text-sm">
  <span className="mb-1 block text-slate-700">{tr("Target DB", "타깃 DB")}</span>
  <span className="inline-block rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700">
    PostgreSQL
  </span>
</div>
```

#### 4.1.3 API 호출부 targetDb 참조 제거

`target.targetDb`를 참조하던 모든 위치(`testTarget`, `startMigration`, `applyCredential`, `replayHistory` 등)에서 해당 필드 참조를 제거하거나 `"postgres"` 리터럴로 대체한다.

| 위치 | 변경 내용 |
|---|---|
| `testTarget()` | `targetDb: target.targetDb` → 제거 |
| `startMigration()` | `targetDb: target.targetDb` → 제거 (서버 기본값 postgres 사용) |
| `applyCredential()` | `targetDb: item.dbType \|\| "postgres"` → `targetDb` 필드 자체 제거 |
| `replayHistory()` / `buildReplayPayload` 응답 처리 | `targetDb` 필드 무시 |
| `loadTargetRecent()` | 반환 타입에서 `targetDb` 제거 |
| `TARGET_RECENT_KEY` localStorage 저장 | `targetDb` 필드 제외 |

---

### 4.2 FR-2: 타겟 테이블 조회 UI

#### 4.2.1 신규 상태

```typescript
const [compareState, setCompareState] = useState<CompareState>({
  targetTables: [],
  fetchedAt: null,
  busy: false,
  error: null,
});
```

#### 4.2.2 fetchTargetTables 함수

```typescript
async function fetchTargetTables() {
  if (!target.targetUrl || !target.schema) return;
  setCompareState((prev) => ({ ...prev, busy: true, error: null }));
  try {
    const res = await fetch("/api/target-tables", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ targetUrl: target.targetUrl, schema: target.schema }),
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.error ?? "조회 실패");
    setCompareState({
      targetTables: data.tables ?? [],
      fetchedAt: data.fetchedAt,
      busy: false,
      error: null,
    });
  } catch (e) {
    setCompareState((prev) => ({
      ...prev,
      busy: false,
      error: e instanceof Error ? e.message : "알 수 없는 오류",
    }));
  }
}
```

#### 4.2.3 버튼 및 배지 JSX (타겟 섹션 하단)

```tsx
<div className="mt-3 flex flex-wrap items-center gap-2">
  <button
    className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100 disabled:opacity-60"
    disabled={compareState.busy || migrationBusy || !target.targetUrl || !target.schema}
    onClick={() => void fetchTargetTables()}
    type="button"
  >
    {compareState.busy
      ? tr("Fetching...", "조회 중...")
      : tr("Fetch Target Tables", "타겟 테이블 조회")}
  </button>
  {compareState.targetTables.length > 0 && (
    <span className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-semibold text-slate-700">
      {compareState.targetTables.length} {tr("tables in target", "개 타겟 테이블")}
    </span>
  )}
  {compareState.fetchedAt && (
    <span className="text-xs text-slate-400">
      {tr("as of", "기준 시각")} {new Date(compareState.fetchedAt).toLocaleTimeString()}
    </span>
  )}
  {compareState.error && (
    <p className="w-full text-sm font-medium text-red-600">{compareState.error}</p>
  )}
</div>
```

---

### 4.3 FR-3: 소스-타겟 비교 패널

#### 4.3.1 비교 목록 도출 (메모이제이션)

```typescript
const compareEntries = useMemo((): TargetTableEntry[] => {
  if (allTables.length === 0 || compareState.targetTables.length === 0) return [];

  // Oracle 테이블명은 대문자 기본 → 소문자 정규화 후 비교
  const sourceSet = new Set(allTables.map((t) => t.toLowerCase()));
  const targetSet = new Set(compareState.targetTables.map((t) => t.toLowerCase()));

  const allNames = new Set([...sourceSet, ...targetSet]);
  return Array.from(allNames)
    .map((name): TargetTableEntry => {
      const inSource = sourceSet.has(name);
      const inTarget = targetSet.has(name);
      const category: TargetTableEntry["category"] =
        inSource && inTarget ? "both" : inSource ? "source_only" : "target_only";
      const precheckRow = precheckItems.find(
        (r) => r.table_name.toLowerCase() === name
      );
      return {
        name,
        inSource,
        inTarget,
        category,
        sourceRowCount: precheckRow?.source_row_count ?? null,
        targetRowCount: precheckRow?.target_row_count ?? null,
      };
    })
    .sort((a, b) => {
      const catOrder = { source_only: 0, both: 1, target_only: 2 };
      const diff = catOrder[a.category] - catOrder[b.category];
      return diff !== 0 ? diff : a.name.localeCompare(b.name);
    });
}, [allTables, compareState.targetTables, precheckItems]);
```

#### 4.3.2 비교 패널 상태

```typescript
type CompareFilter = "all" | "source_only" | "both" | "target_only";
const [compareFilter, setCompareFilter] = useState<CompareFilter>("all");
const [compareSearch, setCompareSearch] = useState("");
```

#### 4.3.3 비교 패널 JSX 구조 (테이블 선택 섹션 상단)

```
[비교 패널 — details/summary 래퍼로 접힘 가능]
  요약 카드 × 3 (source_only | both | target_only 개수)
  카테고리 탭: 전체 / 소스만 / 양쪽 / 타겟만
  검색 입력
  테이블: 테이블명 | 소스 | 타겟 | 소스 행 수 | 타겟 행 수 | 상태
```

배지 색상 클래스:

| 카테고리 | 배지 클래스 |
|---|---|
| `source_only` | `border-blue-300 bg-blue-100 text-blue-800` |
| `both` | `border-emerald-300 bg-emerald-100 text-emerald-800` |
| `target_only` | `border-amber-300 bg-amber-100 text-amber-800` |
| `row_diff` (both + 행 불일치) | `border-orange-300 bg-orange-100 text-orange-800` |

표시 조건: `compareEntries.length > 0` 일 때 렌더링.

---

### 4.4 FR-4: 비교 기반 빠른 선택

테이블 선택 섹션의 기존 버튼 영역에 추가 (비교 결과 있을 때만 렌더링):

```typescript
function selectByCategory(category: TargetTableEntry["category"]) {
  const names = new Set(
    compareEntries
      .filter((e) => e.category === category)
      .map((e) => e.name.toUpperCase()) // Oracle 원본 테이블명으로 복원
  );
  setSelectedTables((prev) => {
    const next = new Set(prev);
    allTables.forEach((t) => {
      if (names.has(t.toUpperCase())) next.add(t);
    });
    return Array.from(next);
  });
}
```

버튼:
- `{tr("Select source-only", "소스에만 있는 테이블 선택")}` → `selectByCategory("source_only")`
- `{tr("Select both", "양쪽에 있는 테이블 선택")}` → `selectByCategory("both")`

---

### 4.5 FR-5: pre-check 행 수 연동

`compareEntries` 계산 시 `precheckItems` 배열에서 `source_row_count` / `target_row_count`를 조회해 삽입한다(4.3.1 참조).

`row_diff` 배지 조건:
```typescript
const isRowDiff =
  entry.category === "both" &&
  entry.sourceRowCount !== null &&
  entry.targetRowCount !== null &&
  entry.sourceRowCount !== entry.targetRowCount;
```

---

## 5. API 계약

### 5.1 POST /api/target-tables

**요청**:
```json
{
  "targetUrl": "postgres://user:pass@host:5432/dbname",
  "schema": "public"
}
```

**응답 200**:
```json
{
  "tables": ["emp", "dept", "salary"],
  "fetchedAt": "2026-03-19T10:00:00Z"
}
```

**오류 응답**:

| 상황 | HTTP | `error` 필드 |
|---|---|---|
| 필수 필드 누락 | 400 | `"Invalid request parameters"` |
| postgres 이외 targetDb (다른 엔드포인트) | 400 | `"v22 이후 타겟 DB는 PostgreSQL만 지원합니다 (입력값: mysql)"` |
| 연결 실패 | 500 | `"타겟 DB 연결 실패: ..."` |
| 권한 부족 | 403 | `"테이블 목록 조회 권한이 없습니다: ..."` |
| 조회 실패 | 500 | `"테이블 목록 조회 실패: ..."` |

### 5.2 기존 API 변경 사항

| 엔드포인트 | 변경 내용 |
|---|---|
| `POST /api/test-target` | `targetDb != "postgres"` 시 400 반환; `else` 브랜치 제거 |
| `POST /api/migrate` | `targetDb != "" && targetDb != "postgres"` 시 400 반환; non-postgres 연결 분기 제거 |
| `POST /api/migrations/precheck` | `targetDb != "" && targetDb != "postgres"` 시 400 반환; non-postgres 연결 분기 제거 |

---

## 6. UI 텍스트 (i18n)

비교 패널에 추가되는 신규 `tr()` 문자열:

| 영어 | 한국어 |
|---|---|
| `"Fetch Target Tables"` | `"타겟 테이블 조회"` |
| `"Fetching..."` | `"조회 중..."` |
| `"tables in target"` | `"개 타겟 테이블"` |
| `"as of"` | `"기준 시각"` |
| `"Refresh"` | `"새로고침"` |
| `"Source vs Target Comparison"` | `"소스-타겟 비교"` |
| `"Source only"` | `"소스에만"` |
| `"Both"` | `"양쪽"` |
| `"Target only"` | `"타겟에만"` |
| `"Select source-only"` | `"소스에만 있는 테이블 선택"` |
| `"Select both"` | `"양쪽에 있는 테이블 선택"` |
| `"Row diff"` | `"행 수 불일치"` |
| `"Run pre-check to see row counts"` | `"Pre-check 실행 후 행 수가 표시됩니다"` |

---

## 7. 오류 처리

| 조건 | 처리 |
|---|---|
| schema가 빈 문자열인 상태에서 조회 버튼 클릭 | 버튼 비활성화로 원천 차단 |
| targetUrl이 빈 문자열인 상태에서 조회 버튼 클릭 | 버튼 비활성화로 원천 차단 |
| 타임아웃(10초) 초과 | `context deadline exceeded` 오류 → 500 반환 |
| 소스 목록 없이 타겟만 조회된 경우 | 비교 패널 미표시; 배지만 노출 |
| 마이그레이션 실행 중 조회 버튼 클릭 | `disabled={migrationBusy}` 로 차단 |
| non-postgres targetDb API 요청 | 400 즉시 반환 (`requirePostgres` 헬퍼) |

---

## 8. 테스트 전략

### 8.1 백엔드 단위 테스트

**파일**: `internal/db/db_test.go`
- `FetchTargetTables`: 스키마 `public`에 테이블 3개 삽입 후 조회 결과 검증 (통합 테스트, SQLite mock 불가 → `pgx` testcontainer 또는 기존 `v9_integration_test.go` 패턴 활용)
- `FetchTargetTables` 빈 스키마: 빈 슬라이스 반환 검증

**파일**: `internal/web/server_test.go`
- `POST /api/target-tables` — 정상 응답: `tables`, `fetchedAt` 포함 확인
- `POST /api/target-tables` — schema 누락: 400 반환
- `POST /api/test-target` — `targetDb: "mysql"` 전달 시 400 반환
- `POST /api/migrate` — `targetDb: "sqlite"` 전달 시 400 반환
- `POST /api/migrations/precheck` — `targetDb: "mssql"` 전달 시 400 반환

### 8.2 프론트엔드 단위 테스트

**파일**: `frontend/src/app/App.test.tsx`
- `compareEntries` 메모 함수: source-only 3개, both 2개, target-only 1개 케이스 검증
- 대소문자 정규화: 소스 `"USERS"`, 타겟 `"users"` → `both` 분류 확인
- `selectByCategory("source_only")`: 해당 테이블만 selectedTables에 추가되는지 확인
- `isRowDiff` 조건: both + 행 수 불일치 → `true`, both + 행 수 동일 → `false`

---

## 9. 롤아웃 계획

### 1차 배포 (FR-0 + FR-1 + FR-2)
- `requirePostgres` 헬퍼 적용 및 non-postgres 분기 제거
- `POST /api/target-tables` 엔드포인트 신설
- 프론트엔드 `TargetState.targetDb` 제거 및 드롭다운 → 고정 레이블 교체
- 조회 버튼·배지 UI 추가

### 2차 배포 (FR-3 + FR-4 + FR-5)
- 비교 패널 렌더링 (요약 카드, 탭 필터, 테이블)
- 빠른 선택 버튼 추가
- pre-check 행 수 연동 및 row_diff 배지

---

## 10. 오픈 이슈

- `FetchTargetTables`를 `information_schema` 대신 `pg_catalog.pg_tables`로 구현할 경우 권한 요구사항이 달라질 수 있음 — 최종 쿼리 확정 필요
- 비교 패널에서 500개 초과 테이블 처리: MVP는 경고 배너만 표시하고 전체 렌더링; 이후 가상 스크롤 도입 여부 논의
- `compareState`를 `localStorage`에 캐시할지 여부 — 페이지 새로고침 후 재조회 강제 vs 캐시 제공 간 UX 트레이드오프
