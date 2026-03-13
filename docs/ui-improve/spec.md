# Technical Specification - Web UI 전면 개선 (ui-improve)

> 변경 대상 파일: `internal/web/templates/index.html` (단일 파일)
> 백엔드 코드(Go) 변경 없음. 모든 API·WebSocket 인터페이스는 v8 그대로 유지.

---

## 1. Stepper 기반 단계 흐름

### 1.1. 마크업 구조

```
<div class="stepper-header">
  <div class="step" data-step="1">Step 1: Connect</div>
  <div class="step" data-step="2">Step 2: Configure</div>
  <div class="step" data-step="3">Step 3: Run</div>
</div>
<div class="step-content" id="step-1"> ... </div>
<div class="step-content" id="step-2"> ... </div>
<div class="step-content" id="step-3"> ... </div>
```

### 1.2. 상태 클래스

| 클래스 | 의미 |
|--------|------|
| `.step--active` | 현재 활성 단계 |
| `.step--done` | 완료된 단계 (체크 아이콘 표시) |
| `.step--disabled` | 아직 미도달 단계 (클릭 불가) |

### 1.3. 단계 전환 로직

- `goToStep(n)` 함수: 현재 단계 이하로만 이동 가능 (앞 단계 건너뛰기 불가)
- Step 1 완료 조건: `fetch /api/tables` 성공 응답 수신
- Step 2 완료 조건: `fetch /api/migrate` 성공 응답 수신 (마이그레이션 시작됨)
- Step 1 헤더 클릭 시: 현재 단계가 2 이상이면 Step 1로 복귀. 입력값 유지(패스워드 제외), 테이블 목록·진행 상황 초기화.
- `all_done` WebSocket 메시지 수신 후: "New Migration" 버튼 노출 → 클릭 시 `goToStep(1)` + 전체 상태 초기화.

---

## 2. Step 1 – Oracle 연결 개선

### 2.1. 연결 성공 배너

- `/api/tables` 200 응답 수신 시, 연결 폼 하단에 인라인 배너 표시:
  ```
  ✓ Connected: {oracleUrl}  —  {N} tables found
  ```
- 배너 클래스: `.conn-success-banner` (초록 배경, 아이콘 포함)
- Step 2로 전환되기 전 500ms 유지 후 전환 (사용자가 성공을 인지할 시간 확보)

### 2.2. 에러 처리

- 에러 메시지는 기존 `.error-msg` 유지
- HTTP 상태 코드 또는 응답 `error` 필드로 에러 유형 분기:
  - 연결 거부: "Oracle 서버에 연결할 수 없습니다. URL을 확인하세요."
  - 인증 실패: "사용자명 또는 패스워드가 올바르지 않습니다."
  - 기타: 서버 응답 메시지 그대로 표시

---

## 3. Step 2 – 설정 구조 개편

### 3.1. 레이아웃

- **1024px 이상**: 좌우 2컬럼 (`display: grid; grid-template-columns: 1fr 1fr; gap: 2rem`)
  - 좌: 테이블 선택 패널
  - 우: 마이그레이션 설정 패널
- **1024px 미만**: 단일 컬럼

### 3.2. 테이블 선택 패널

**인라인 검색**
- `<input type="text" id="tableSearch" placeholder="테이블 검색...">` 추가
- `input` 이벤트마다 `.table-item`의 표시 여부를 `tableName.toLowerCase().includes(query)` 기준으로 토글
- 검색 결과 없음 시: "검색 결과 없음" 문구 표시

**선택 카운터**
- `<span id="selectedCount">0 / N 선택됨</span>` 실시간 업데이트
- `.table-cb` 체크박스 change 이벤트 마다 카운트 갱신

**버튼 그룹**
- "전체 선택" / "전체 해제" 버튼 (기존 `checkAll` 체크박스를 버튼 2개로 교체)

### 3.3. 마이그레이션 설정 패널

기존 `<details>` + 평탄한 항목들을 아래 4개 섹션으로 나눈다. 각 섹션은 구분선(`<hr>`)과 소제목(`<h3>`)으로 시각 분리.

**섹션 A: 출력 방식**
- 라디오 버튼 2개: "SQL 파일 생성" / "Direct Migration (직접 이관)"
  - 기존 `directMigration` 체크박스 대체
- "SQL 파일 생성" 선택 시: Output Filename, Per-Table 옵션 표시
- "Direct Migration" 선택 시: Target URL 입력 표시

**섹션 B: 대상 데이터베이스**
- Target DB 드롭다운 (기존 `targetDb` 유지)
- Target URL 입력 (기존 `pgUrl` 유지)
- Schema 입력 (기존 `schema` 유지)
- 각 DB 선택에 따른 URL placeholder 동적 변경 로직 유지

**섹션 C: DDL 옵션**
- "CREATE TABLE 생성" 체크박스 (기존 `withDdl`)
- 하위 옵션 (DDL 체크 시 활성화):
  - "시퀀스 포함" (기존 `withSequences`)
  - "인덱스 포함" (기존 `withIndexes`)
  - "제약조건 포함 (FK, Check, Default)" (기존 `withConstraints`)
  - "Oracle 소유자(Owner)" 입력 (기존 `oracleOwner`)
- CLI 플래그 텍스트 (`--with-sequences` 등) 레이블에서 완전히 제거

**섹션 D: 고급 설정**
- Batch Size, Parallel Workers (기존 유지)
- DB Connection Pool: Max Open, Max Idle, Max Life (기존 유지)
- JSON Logging 체크박스 (기존 유지)
- Dry-Run 체크박스 (기존 유지, Step 2 하단으로 이동)

---

## 4. Step 3 – 진행 상황 대시보드

### 4.1. 전체 요약 진행바

```html
<div class="overall-progress">
  <div class="overall-progress-header">
    <span id="overall-label">0 / N 완료</span>
    <span id="overall-pct">0%</span>
  </div>
  <div class="progress-bar-bg">
    <div id="overall-fill" class="progress-bar-fill"></div>
  </div>
</div>
```

- `totalTables`: `btnMigrate` 클릭 시 `selectedTables.length` 로 설정
- `doneTables`: `msg.type === 'done'` 또는 `msg.type === 'error'` 수신 시 +1
- 완료율 = `doneTables / totalTables * 100`

### 4.2. 완료 요약 카드

`all_done` 메시지 수신 시 진행 목록 하단에 카드 추가:

```
┌──────────────────────────────────────┐
│  Migration Complete                  │
│  ✓ 성공  12   ✗ 실패  1   ⚠ 경고  2  │
│  소요 시간: 00:03:24                  │
│  [ZIP 다운로드]  [New Migration]      │
└──────────────────────────────────────┘
```

- 성공/실패 카운트: `done`/`error` 메시지 수신 수 집계
- 경고 카운트: `warning` 메시지 수신 수 집계 (중복 제거된 건수)
- 소요 시간: `btnMigrate` 클릭 시점부터 `all_done` 수신 시점까지 `Date.now()` 차이
- "New Migration" 버튼: `goToStep(1)` + 전체 상태 초기화

---

## 5. 다크모드

### 5.1. CSS 변수 정의

`:root`의 기존 변수 외에 다크 전용 변수 세트를 `[data-theme="dark"]` 셀렉터로 정의:

```css
[data-theme="dark"] {
  --bg-color: #0f172a;
  --card-bg: #1e293b;
  --text-main: #f1f5f9;
  --text-muted: #94a3b8;
  --border-color: #334155;
  --primary-color: #60a5fa;
  --primary-hover: #3b82f6;
  --success-color: #34d399;
  --danger-color: #f87171;
}
```

OS 자동 연동:
```css
@media (prefers-color-scheme: dark) {
  :root { /* 위 다크 변수 동일 적용 */ }
}
```

단, `data-theme` 속성이 명시적으로 설정된 경우 미디어 쿼리보다 우선 적용.

### 5.2. 토글 버튼

- 헤더 우상단 (Session ID 옆)에 `🌙 / ☀` 아이콘 버튼 배치
- 클릭 시: `document.documentElement.dataset.theme` 토글 (`light` ↔ `dark`)
- `localStorage.setItem('theme', value)` 로 저장
- 페이지 로드 시: `localStorage.getItem('theme')` → 없으면 `prefers-color-scheme` 감지

---

## 6. 레이아웃 & 반응형

### 6.1. 컨테이너

```css
.container {
  max-width: 1200px; /* 기존 800px → 1200px */
  width: 100%;
  padding: 0 1.5rem;
}
```

### 6.2. 헤더 바

- `<header>` 태그로 제목, Session ID, 다크모드 토글 통합
- `position: sticky; top: 0;` 로 스크롤 시에도 상단 고정
- `z-index: 100`

```html
<header class="app-header">
  <h1 class="app-title">{{ .title }}</h1>
  <div class="header-actions">
    <span class="session-info">Session: <code id="currentSessionId">{{ .sessionId }}</code></span>
    <button id="theme-toggle" aria-label="다크모드 토글">🌙</button>
  </div>
</header>
```

### 6.3. 반응형 브레이크포인트

| 브레이크포인트 | 레이아웃 변경 |
|---------------|-------------|
| `> 1024px` | Step 2: 2컬럼 그리드 |
| `≤ 1024px` | Step 2: 단일 컬럼 |
| `≤ 768px` | 폼 요소 전체 너비, 버튼 padding 증가, 폰트 소폭 축소 |

---

## 7. 접근성

- 모든 `<input>`, `<select>`에 연결된 `<label>` 또는 `aria-label` 보장
- 비활성 서브옵션(`withSequences` 등)에 `aria-disabled="true"` 추가
- 버튼 로딩 중 `aria-busy="true"`, `aria-label` 업데이트 (`"연결 중..."` 등)
- `.table-list`에 `role="listbox"`, 각 `.table-item`에 `role="option"` 추가 (스크린리더 호환)
- 포커스 링: `outline: none` 제거 후 `:focus-visible` 기반 커스텀 스타일 적용

```css
:focus-visible {
  outline: 2px solid var(--primary-color);
  outline-offset: 2px;
}
```

---

## 8. 언어 정비 기준

| 기존 표현 | 변경 표현 |
|-----------|-----------|
| `Sequence DDL 포함 (--with-sequences)` | `Sequence DDL 포함` |
| `Index DDL 포함 (--with-indexes)` | `Index DDL 포함` |
| `제약조건(Default, FK, Check) 포함 (--with-constraints)` | `제약조건 포함 (FK, Check, Default)` |
| `Oracle 소유자` | `Oracle Owner (소유자)` |
| `출력 대상 DB (Target DB)` | `Target Database` |
| 경고 배너 한국어 하드코딩 텍스트 | 서버 응답 메시지 그대로 표시 |
