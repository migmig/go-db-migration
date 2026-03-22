# DBMigrator Project History

This file contains the consolidated history of all completed versions and improvements.

## Table of Contents
- [ui-improve](#ui-improve)
- [v01](#v01)
- [v02](#v02)
- [v03](#v03)
- [v04](#v04)
- [v05](#v05)
- [v06](#v06)
- [v07](#v07)
- [v08](#v08)
- [v09](#v09)
- [v10](#v10)
- [v11](#v11)
- [v12](#v12)
- [v13](#v13)
- [v14](#v14)
- [v15](#v15)
- [v16](#v16)
- [v17](#v17)
- [v18](#v18)
- [v19](#v19)
- [v20](#v20)
- [v22](#v22)
- [v23](#v23)
- [v24](#v24)
- [v25](#v25)
- [v26](#v26)

---
## <a name="ui-improve"></a> ui-improve

### prd.md

# PRD (Product Requirements Document) - Web UI 전면 개선 (ui-improve)

## 1. 개요 (Overview)

v8을 통해 멀티 세션, 커넥션 풀 제어, 제약조건 마이그레이션, 재개 기능 등 백엔드 엔터프라이즈 기능이 완성되었습니다.
이번 **ui-improve** 릴리스는 기능 추가 없이 **사용자 경험(UX)과 인터페이스(UI) 품질을 전면 개편**하는 데 집중합니다.
현재 UI는 단일 HTML 파일에 CSS·JS가 인라인으로 구성된 단순한 형태로, 기능 증가에 따른 복잡도를 감당하기 어려운 구조가 되었습니다.

---

## 2. 배경 및 문제 분석 (Background & Issues Found)

### 2.1. 레이아웃 – 좁은 단일 컬럼, 와이드 스크린 낭비

- **현상**: `max-width: 800px` 단일 컬럼 구조로, 1440px 이상 모니터에서 양쪽 여백이 과도하게 낭비됩니다.
- **문제**: 테이블 선택(Section 2)과 진행 상황(Section 3)을 동시에 확인하려면 스크롤이 필요하며, 정보 밀도가 낮습니다.

### 2.2. UX 흐름 – 단계 구분 없는 순차 스크롤

- **현상**: 3개 섹션이 조건부 `display: none/block`으로 순차 노출됩니다.
- **문제**: 사용자가 현재 어느 단계에 있는지, 이전 단계로 돌아갈 수 있는지 명확하지 않습니다. 완료 후 재시작 흐름도 불분명합니다.

### 2.3. Advanced Settings – 발견성 낮고 정보 구조 혼재

- **현상**: `<details>` 태그로 숨겨진 고급 설정 안에 Target DB 선택, 배치/워커 설정, DB 커넥션 풀, 출력 파일, DDL 옵션이 뒤섞여 있습니다.
- **문제**: Direct Migration 체크박스는 Advanced Settings 바깥에 위치하는 등 설정 항목들의 논리적 그룹핑이 일관성이 없습니다.

### 2.4. 테이블 선택 – 검색·요약 기능 부재

- **현상**: 테이블 목록이 고정 높이 300px 스크롤 영역으로만 표시됩니다.
- **문제**: 수십~수백 개 테이블 환경에서 원하는 테이블을 찾을 인라인 검색 수단이 없습니다. 선택된 테이블 수도 표시되지 않아 상태 파악이 어렵습니다.

### 2.5. 진행 상황 – 전체 요약 없음

- **현상**: 테이블별 개별 진행바만 표시됩니다.
- **문제**: 전체 N개 테이블 중 M개 완료 같은 요약 지표가 없어, 대규모 마이그레이션 시 전체 진행률을 파악하기 어렵습니다. 완료 시 최종 결과 요약(성공/실패/경고 건수)도 표시되지 않습니다.

### 2.6. 다크모드 미지원

- **현상**: 라이트 모드 고정 배경(`--bg-color: #f8fafc`).
- **문제**: CSS 변수 구조로 다크모드 추가가 기술적으로 용이하지만 구현이 없습니다.

### 2.7. 언어 혼재 & 레이블 품질

- **현상**: `Oracle 소유자`, `--with-sequences`, `--with-constraints` 같은 CLI 플래그 노출, 한국어·영어가 혼재된 레이블.
- **문제**: 사용자 대상 도구임에도 개발자 관점의 표현이 그대로 노출되어 있습니다.

### 2.8. 연결 성공 피드백 부재

- **현상**: 연결 성공 시 Section 2가 나타나는 것만으로 확인합니다.
- **문제**: 명시적인 성공 표시(연결된 Oracle URL, 테이블 수) 없이 섹션이 전환되어 사용자가 성공 여부를 직관적으로 인지하기 어렵습니다.

### 2.9. 접근성 & 반응형 미흡

- **현상**: `aria-*` 속성이 없고, 모바일 레이아웃이 기본 수준에 머뭅니다.
- **문제**: 시맨틱 접근성이 부족하며 소형 화면에서 UX가 저하됩니다.

---

## 3. 목표 (Goals)

1. **명확한 단계 흐름(Stepper)**: 사용자가 "연결 → 설정 → 실행 → 결과" 단계를 직관적으로 인지하고 이전 단계로 복귀할 수 있게 한다.
2. **설정 그룹핑 개선**: 논리적으로 연관된 설정 항목을 탭 또는 섹션으로 명확히 분리한다.
3. **테이블 선택 UX 강화**: 인라인 검색·필터, 선택 수 카운터를 추가한다.
4. **진행 상황 대시보드화**: 전체 진행률 및 완료 후 결과 요약을 제공한다.
5. **다크모드 지원**: OS 설정 자동 연동 및 수동 토글을 지원한다.
6. **레이블·언어 정비**: CLI 플래그 표현 제거, 한국어/영어 혼재 통일.
7. **반응형 & 접근성 기반 구축**: 와이드 2컬럼 레이아웃, 모바일 대응, 기본 aria 속성 추가.

---

## 4. 기능 요구사항 (Functional Requirements)

### 4.1. Stepper 기반 단계 흐름

- **단계 구성**: `Step 1: Connect` → `Step 2: Configure` → `Step 3: Run`
- **상태 표시**: 각 스텝 헤더에 완료(`✓`)/진행 중/비활성 상태를 시각적으로 표시한다.
- **뒤로 가기**: Step 2 이후 단계에서 이전 스텝 헤더를 클릭하면 해당 단계로 복귀할 수 있다.
  - Step 1 복귀 시: 연결 폼 재노출, 테이블 목록·진행 상황 초기화.
- **완료 후 재시작**: Step 3 진행 완료 후 "New Migration" 버튼으로 Step 1로 초기화.

### 4.2. Step 1 – Oracle 연결 개선

- **연결 성공 배너**: 연결 성공 시 "Connected: `{url}` — {N} tables found" 인라인 배너 표시.
- **연결 실패 처리**: 에러 메시지에 구체적인 원인(타임아웃, 인증 오류 등) 분류 표시.
- **입력값 유지**: Step 1으로 복귀해도 이전 입력값(URL, 사용자명)은 유지된다. (패스워드는 보안상 초기화)

### 4.3. Step 2 – 설정 구조 개편

설정을 두 개의 논리 그룹으로 분리한다.

**그룹 A: 테이블 선택**
- 인라인 검색 입력창: 테이블 목록 실시간 필터링 (대소문자 무시)
- 선택 카운터: `N / Total 선택됨` 실시간 표시
- "Select All" / "Clear All" 토글 버튼

**그룹 B: 마이그레이션 설정** (기존 Advanced Settings 대체)
- 탭 또는 명시적 섹션으로 구분:
  - **출력 방식**: SQL 파일 생성 vs. Direct Migration 토글 (라디오 버튼 또는 세그먼트 컨트롤로 교체)
  - **타겟 DB 설정**: Target DB 선택, Target URL, Schema
  - **성능 설정**: Batch Size, Workers, DB Connection Pool (Max Open/Idle/Life)
  - **DDL 옵션**: DDL 생성 여부, Sequences, Indexes, Constraints, Oracle Owner
  - **기타**: Per-Table 출력, Output Filename, JSON Logging, Dry-Run
- CLI 플래그 텍스트(`--with-sequences` 등) 제거 → 사람이 읽을 수 있는 레이블로 교체
- 한국어/영어 혼재 정비 (기본 영어 UI, 필요시 설명 툴팁)

### 4.4. Step 3 – 진행 상황 대시보드

- **전체 요약 바**: 상단에 "전체 N개 중 M개 완료 (P%)" 진행바 추가
- **완료 요약 카드**: `all_done` 이벤트 수신 시 표시
  - 성공 건수 / 실패 건수 / 경고 건수
  - 총 소요 시간
  - ZIP 다운로드 버튼 (기존 기능 유지)
- **개별 진행 항목**: 기존 테이블별 진행바 유지, 에러 발생 시 오류 메시지 인라인 표시 개선
- **경고 배너**: 기존 `warning-banner` 스타일 유지, 동일 메시지 중복 방지 로직 유지

### 4.5. 다크모드

- `prefers-color-scheme: dark` 미디어 쿼리로 OS 설정 자동 연동
- 헤더 영역에 수동 토글 버튼 추가 (라이트/다크 전환, `localStorage`에 기본값 저장)
- 다크모드 CSS 변수 세트 정의:
  - `--bg-color: #0f172a`, `--card-bg: #1e293b`, `--text-main: #f1f5f9`, `--border-color: #334155`

### 4.6. 레이아웃 & 반응형

- **와이드 레이아웃**: `max-width: 1200px`, Step 2에서 테이블 선택(좌)/설정(우) 2컬럼 레이아웃 (breakpoint: 1024px 이하에서 단일 컬럼 전환)
- **모바일 대응**: 768px 이하에서 폼 요소 전체 너비, 버튼 크기 증가
- **Session ID 표시**: `position: absolute` 대신 헤더 내 고정 위치로 이동 (스크롤 무관하게 안정적 표시)

### 4.7. 접근성 기본 지원

- 모든 인터랙티브 요소에 `aria-label` 또는 연결된 `<label>` 보장
- 버튼 로딩 상태에 `aria-busy="true"` 추가
- 포커스 링(focus ring) 스타일 명시 (현재 `outline: none` 일부 제거)
- 색상 대비 WCAG AA 기준 충족 확인

---

## 5. 아키텍처 및 변경 범위 (Scope of Changes)

| 변경 영역 | 파일 | 상세 내용 |
|-----------|------|-----------|
| **HTML 구조** | `internal/web/templates/index.html` | Stepper 마크업 추가, 섹션 재구성, aria 속성 추가 |
| **CSS** | (동일 파일 `<style>`) | 다크모드 변수, 와이드 레이아웃, 반응형 미디어 쿼리, Stepper 스타일 |
| **JavaScript** | (동일 파일 `<script>`) | 단계 전환 로직, 테이블 인라인 검색, 전체 진행률 계산, 다크모드 토글, 결과 요약 렌더링 |
| **백엔드 변경** | 없음 | API·서버 코드 변경 없음. 순수 프론트엔드 개선 |

> **`web/templates/index.html`** (구버전 사본)은 이번 개선 대상에서 제외하거나 삭제를 검토한다.

---

## 6. 비기능 요구사항 (Non-Functional Requirements)

- **외부 의존성 없음**: CDN·외부 라이브러리 추가 없이 순수 HTML/CSS/JS로 구현한다. 바이너리 임베딩 방식을 유지한다.
- **기존 기능 100% 호환**: v8의 모든 API 요청/응답 구조, WebSocket 메시지 포맷을 변경하지 않는다.
- **로딩 성능 유지**: 파일 크기가 현재(37KB) 대비 50% 이상 증가하지 않도록 코드 최적화에 주의한다.
- **브라우저 호환**: Chrome 최신, Firefox 최신, Safari 최신 기준 동작을 보장한다.

---

## 7. 마일스톤 (Milestones)

1. **1단계: 구조 재편** — Stepper 마크업, 섹션 재조합, Step 2 그룹 A·B 분리
2. **2단계: 테이블 UX** — 인라인 검색, 선택 카운터, 레이블 정비
3. **3단계: 진행 대시보드** — 전체 요약 바, 완료 요약 카드, 경고 개선
4. **4단계: 다크모드 & 레이아웃** — 다크모드 변수·토글, 와이드 2컬럼, 반응형
5. **5단계: 접근성 & QA** — aria 속성, 포커스 링, WCAG 대비 확인, 브라우저 교차 검증

---

## 8. 기대 효과 (Expected Outcomes)

- DBA·개발자가 수십~수백 개 테이블 마이그레이션 설정 시 오조작 없이 빠르게 설정 완료
- 단계별 진행 상황이 명확히 보여 모니터링 부담 감소
- 다크모드로 장시간 작업 시 눈의 피로 감소
- CLI 플래그 표현 제거로 비개발자 사용자(DBA)도 직관적으로 사용 가능


### spec.md

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


### tasks.md

# Implementation Tasks - Web UI 전면 개선 (ui-improve)

> 모든 작업은 `internal/web/templates/index.html` 단일 파일 내에서 수행.
> 백엔드(Go) 코드 변경 없음.

---

## 1단계: 구조 재편 – Stepper & 헤더 ✅

- [x] **앱 헤더 구성**
  - [x] `<header class="app-header">` 마크업 추가 (제목 + Session ID + 다크모드 토글 버튼 포함)
  - [x] `position: sticky; top: 0` 헤더 CSS 적용
  - [x] 기존 `.session-info` (`position: absolute`) 제거 → 헤더 내로 이동
  - [x] 다크모드 토글 버튼 (`#theme-toggle`) 추가 (5단계 기능도 함께 구현)

- [x] **Stepper 마크업 추가**
  - [x] `.stepper-header` + 3개 `.step[data-step]` 마크업 추가
  - [x] 각 스텝 상태 클래스 CSS 정의: `.step--active`, `.step--done`, `.step--disabled`
  - [x] 기존 섹션 3개를 `.step-content#step-1/2/3` 으로 교체

- [x] **`goToStep(n)` 함수 구현**
  - [x] 단계 전환 시 `.step-content` 표시/숨김 처리
  - [x] 스텝 헤더 상태 클래스 업데이트
  - [x] Step 1 클릭 복귀 시 테이블 목록·진행 상황 초기화 로직

- [x] **단계 전환 트리거 연결**
  - [x] `/api/tables` 성공 → 연결 성공 배너 600ms 후 `showStep(2)`
  - [x] `/api/migrate` 클릭 → `showStep(3)`
  - [x] `all_done` 수신 → "New Migration" 버튼 노출
  - [x] "New Migration" 클릭 → `goToStep(1)` + 전체 상태 초기화

---

## 2단계: Step 1 개선 – 연결 피드백 ✅

- [x] **연결 성공 배너**
  - [x] `.conn-success-banner` CSS 스타일 정의 (초록 배경, 체크 아이콘)
  - [x] `/api/tables` 200 응답 시 배너 렌더링: `"✓ Connected: {url} — {N} tables found"`
  - [x] 500ms 후 `goToStep(2)` 전환

- [x] **에러 메시지 개선**
  - [x] HTTP 상태코드 / `error` 필드 기반 에러 유형 분기 처리
  - [x] 연결 거부 / 인증 실패 / 기타 메시지 분기

---

## 3단계: Step 2 개선 – 테이블 UX & 설정 재구성 ✅

- [x] **테이블 선택 패널 – 인라인 검색**
  - [x] `#tableSearch` 입력창 추가 (테이블 목록 위)
  - [x] `input` 이벤트로 `.table-item` 실시간 필터링
  - [x] 검색 결과 없음 안내 문구 처리

- [x] **테이블 선택 패널 – 선택 카운터**
  - [x] `#selectedCount` 스팬 추가 (`0 / N 선택됨`)
  - [x] `.table-cb` change 이벤트 및 renderTableList 완료 후 카운터 갱신 함수 구현
  - [x] `checkAll` 체크박스를 "전체 선택" / "전체 해제" 버튼 2개로 교체

- [x] **마이그레이션 설정 패널 – 레이아웃**
  - [x] 기존 `<details>` 제거
  - [x] 섹션 A·B·C·D 소제목(`<h3>`) + 구분선(`<hr>`) 구조로 교체

- [x] **섹션 A: 출력 방식**
  - [x] 기존 `directMigration` 체크박스 → 라디오 버튼 2개로 교체 ("SQL 파일 생성" / "Direct Migration")
  - [x] 각 라디오 선택에 따른 조건부 필드 표시/숨김 유지

- [x] **섹션 B: 대상 데이터베이스**
  - [x] Target DB 드롭다운, Target URL, Schema 입력 배치
  - [x] `handleTargetDbChange()` 로직 유지

- [x] **섹션 C: DDL 옵션**
  - [x] `withDdl`, `withSequences`, `withIndexes`, `withConstraints`, `oracleOwner` 재배치
  - [x] CLI 플래그 텍스트 (`--with-*`) 레이블에서 제거
  - [x] 언어 정비 테이블 기준으로 레이블 교체

- [x] **섹션 D: 고급 설정**
  - [x] Batch Size, Workers, DB Pool (Max Open/Idle/Life), JSON Logging, Dry-Run 재배치
  - [x] `dryRun` 체크박스를 Step 2 하단 (Start 버튼 위)으로 이동

- [x] **2컬럼 그리드 레이아웃 CSS**
  - [x] Step 2 컨텐츠에 `display: grid; grid-template-columns: 1fr 1fr` 적용
  - [x] `@media (max-width: 1024px)` 에서 단일 컬럼 전환

---

## 4단계: Step 3 개선 – 진행 대시보드 ✅

- [x] **전체 요약 진행바**
  - [x] `.overall-progress` 마크업 및 CSS 추가 (Step 3 최상단)
  - [x] `totalTables` 변수: migrate 클릭 시 `selectedTables.length` 설정
  - [x] `doneTables` 카운터: `done`/`error` 메시지 수신마다 +1
  - [x] 전체 진행율 `#overall-fill`, `#overall-label`, `#overall-pct` 업데이트 함수

- [x] **완료 요약 카드**
  - [x] `.summary-card` CSS 정의
  - [x] `all_done` 수신 시 성공/실패/경고 카운트, 소요 시간 집계 후 카드 렌더링
  - [x] ZIP 다운로드 버튼 카드 내로 이동 (기존 `#download-btn` 연동 유지)
  - [x] "New Migration" 버튼 카드 내 배치

---

## 5단계: 다크모드 ✅

- [x] **CSS 변수 다크 세트 정의**
  - [x] `[data-theme="dark"]` 셀렉터에 다크 변수 추가
  - [x] `@media (prefers-color-scheme: dark)` 에 동일 변수 적용
  - [x] `warning-banner` 다크모드 색상 대응

- [x] **토글 버튼 기능 구현**
  - [x] 클릭 시 `document.documentElement.dataset.theme` 토글
  - [x] `localStorage` 저장/불러오기
  - [x] 페이지 로드 시 저장값 또는 OS 설정 감지하여 초기 테마 적용
  - [x] 버튼 아이콘 `🌙` / `☀️` 전환

---

## 6단계: 레이아웃 & 반응형 & 접근성 ✅

- [x] **컨테이너 max-width 확장**
  - [x] `.container` `max-width: 800px` → `1200px`

- [x] **반응형 미디어 쿼리**
  - [x] `@media (max-width: 768px)`: 폼 전체 너비, 버튼 padding 증가

- [x] **접근성 마크업**
  - [x] 모든 `<input>`, `<select>` `aria-label` 또는 `<label>` 연결 확인 및 보완
  - [x] 비활성 하위 옵션에 `aria-disabled="true"` 추가
  - [x] 버튼 로딩 상태 `aria-busy="true"` 처리
  - [x] `.table-list` `role="listbox"`, `.table-item` `role="option"` 추가
  - [x] `outline: none` 제거 → `:focus-visible` 커스텀 스타일 적용

- [x] **구버전 사본 처리**
  - [x] `web/templates/index.html` 삭제 여부 확인 후 제거 또는 동기화

---

## 7단계: QA & 검증

- [ ] **기능 회귀 테스트**
  - [ ] Step 1: Oracle 연결, 테이블 조회 정상 동작
  - [ ] Step 2: SQL 파일 생성 모드 전체 옵션 정상 전달
  - [ ] Step 2: Direct Migration 모드 전체 옵션 정상 전달
  - [ ] Step 3: WebSocket 진행 메시지 타입별 렌더링 (`init`, `update`, `done`, `error`, `warning`, `ddl_progress`, `dry_run_result`, `all_done`)
  - [ ] ZIP 다운로드 정상 동작
  - [ ] Session ID 표시 정상

- [ ] **크로스 브라우저 확인**
  - [ ] Chrome 최신
  - [ ] Firefox 최신
  - [ ] Safari 최신

- [ ] **반응형 확인**
  - [ ] 1440px 와이드 레이아웃
  - [ ] 1024px 브레이크포인트 전환
  - [ ] 375px 모바일 레이아웃

- [ ] **다크모드 확인**
  - [ ] OS 다크 설정 자동 감지
  - [ ] 토글 버튼 수동 전환
  - [ ] `localStorage` 유지 확인


---
## <a name="v01"></a> v01

### prd.md

# 제품 요구사항 문서 (PRD): Oracle에서 PostgreSQL 데이터 마이그레이션 CLI

## 1. 개요
이 프로젝트는 Go 언어로 구축된 명령줄 인터페이스(CLI) 애플리케이션으로, Oracle 데이터베이스에서 데이터를 추출하여 PostgreSQL 호환 `INSERT` SQL 스크립트를 생성하도록 설계되었습니다. 이 도구는 성능과 유연성에 중점을 두며, 대량 삽입(Bulk Insert) 기능, 올바른 데이터 타입 매핑(예: CLOB, BLOB, Timestamp) 및 병렬 처리, 테이블별 출력 파일 생성과 같은 고급 실행 옵션을 제공합니다.

## 2. 목표
- **데이터 내보내기:** Oracle 데이터베이스에서 지정된 테이블을 조회합니다.
- **데이터 변환:** Oracle 전용 데이터 타입(BLOB, RAW, CLOB, DATE, TIMESTAMP)을 적절한 PostgreSQL 리터럴 구문으로 변환합니다 (예: bytea에 대한 `\x...` 16진수 인코딩, 적절하게 이스케이프된 문자열, 정확한 타임스탬프 문자열).
- **효율성:** 개별 행 삽입 대신 `BULK INSERT` 문을 생성하여 PostgreSQL 가져오기 오버헤드를 줄입니다.
- **병렬 처리 및 구성:** 테이블을 동시에 처리하고 결과를 테이블별로 개별 `.sql` 파일로 출력하는 옵션을 제공합니다.

## 3. 범위 및 기능
### 핵심 기능
- 순수 Go 드라이버(CGO/Oracle Instant Client 불필요)를 사용하여 Oracle 데이터베이스에 연결합니다.
- CLI 플래그를 통해 연결 자격 증명, URL 및 쉼표로 구분된 테이블 목록을 허용합니다.
- `NULL` 값을 원활하게 처리합니다.
- 과도한 메모리 소비를 방지하기 위해 `SELECT * FROM <table>`을 반복적으로 추출합니다.
- `INSERT INTO` 문을 작성하기 전에 행을 구성 가능한 크기(기본값 1000)까지 일괄 처리(Batch)합니다.

### 고급 출력 옵션 (신규)
- **`--per-table` 플래그:**
  - 활성화되면 모든 SQL 명령을 단일 출력 파일(예: `migration.sql`)에 작성하는 대신 테이블별로 `<테이블명>_migration.sql` 형식의 개별 파일을 생성합니다.
- **`--parallel` 플래그:**
  - Go 루틴(`sync.WaitGroup`)을 사용하여 여러 테이블에 대한 동시 추출 및 파일 쓰기를 활성화합니다.
  - 여러 테이블이 지정된 경우 대규모 스키마의 추출 속도를 크게 높입니다.
  - `--parallel`과 단일 출력 파일(`--per-table` 없음)이 모두 사용되는 경우 데이터 인터리빙/손상을 방지하기 위해 Mutex를 사용하여 파일 쓰기를 안전하게 동기화해야 합니다.

## 4. CLI 인수
| 플래그 | 설명 | 기본값 | 필수 여부 |
| --- | --- | --- | --- |
| `-url` | Oracle DB 연결 URL / DSN | 없음 | 예 |
| `-user` | 데이터베이스 사용자명 | 없음 | 예 |
| `-password`| 데이터베이스 비밀번호 | 없음 | 예 |
| `-tables` | 쉼표로 구분된 테이블 목록 | 없음 | 예 |
| `-out` | 출력 SQL 파일명 (`-per-table`이 false인 경우) | `migration.sql` | 아니오 |
| `-batch` | 대량 삽입(Bulk Insert) 문당 행 수 | `1000` | 아니오 |
| `-per-table`| `<tablename>.sql`이라는 별도의 파일로 출력 | `false` | 아니오 |
| `-parallel`| 테이블을 동시에 처리 | `false` | 아니오 |

## 5. 비기능적 요구사항
- **언어:** Go 1.21 이상
- **드라이버:** `github.com/sijms/go-ora/v2`
- **성능:** 결과를 효율적으로 스트리밍하고 큰 테이블에 대해서도 메모리 사용량을 낮게 유지합니다. 병렬 작업을 위해 고루틴(Goroutine)을 사용합니다.
- **안전성:** 문자열 열에서 작은따옴표를 엄격하게 이스케이프하여 생성된 파일의 SQL 인젝션 위험을 방지합니다. 필요한 경우 Mutex를 사용하여 공유 파일에 동시에 씁니다.

## 6. 향후 개선 사항
- DDL 생성 (CREATE TABLE 스키마).
- 중간 파일 없이 직접 PostgreSQL 삽입.
- 공간(Spatial) 데이터와 같은 더 복잡한 데이터 타입 지원.


### spec.md

# 기술 명세서: Oracle에서 PostgreSQL 데이터 마이그레이션 CLI

## 1. 소개
이 문서는 Oracle 데이터베이스에서 PostgreSQL 호환 SQL 스크립트로 데이터를 마이그레이션하는 Go 기반 CLI 도구의 기술적 설계 및 구현 세부 사항을 정의합니다.

## 2. 기술 스택
- **언어:** Go 1.21 이상
- **Oracle 드라이버:** `github.com/sijms/go-ora/v2` (순수 Go 드라이버, Oracle Instant Client 불필요)
- **동시성:** Go 표준 라이브러리 `sync` (WaitGroups, Mutex) 및 `channels`

## 3. 아키텍처 및 설계
이 도구는 단일 바이너리 CLI로 작동합니다. 테이블 크기에 관계없이 적은 메모리 사용량을 보장하기 위해 스트리밍 아키텍처를 따릅니다.

### 3.1 데이터 흐름
1. **연결:** 제공된 DSN을 사용하여 소스 Oracle DB에 대한 연결을 설정합니다.
2. **메타데이터 발견:** 각 테이블에 대해 열 이름과 타입을 쿼리합니다.
3. **스트리밍 추출:** `SELECT * FROM <table>`을 실행하고 `sql.Rows`를 사용하여 행을 반복합니다.
4. **변환:** Oracle 타입을 PostgreSQL 호환 리터럴로 변환합니다.
5. **일괄 처리(Batching):** `batch` 크기에 도달할 때까지 행을 메모리에 누적합니다.
6. **쓰기:** `INSERT INTO` 문을 대상 파일에 형식화하고 씁니다.

## 4. 데이터 타입 매핑
이 도구는 다음 변환을 처리해야 합니다:

| Oracle 타입 | PostgreSQL 타입 | 변환 로직 |
| --- | --- | --- |
| `VARCHAR2`, `CHAR`, `NVARCHAR2` | `text` / `varchar` | 작은따옴표(`'`)를 두 개(`''`)로 만들어 이스케이프합니다. |
| `NUMBER` | `numeric` / `int` / `float` | 직접적인 문자열 표현. |
| `DATE`, `TIMESTAMP` | `timestamp` | `YYYY-MM-DD HH24:MI:SS.FF` 형식으로 포맷합니다. |
| `CLOB` | `text` | 큰 문자열로 취급하고 이스케이프 처리를 합니다. |
| `BLOB`, `RAW` | `bytea` | `\x...` 16진수 형식으로 변환합니다. |
| `NULL` | `NULL` | 명시적으로 `NULL`을 씁니다. |

## 5. CLI 인터페이스
애플리케이션은 표준 `flag` 패키지 또는 `cobra`와 같은 라이브러리를 사용합니다.

| 플래그 | 타입 | 설명 |
| --- | --- | --- |
| `-url` | string | Oracle DSN (예: `oracle://user:pass@host:port/service`) |
| `-user` | string | DB 사용자명 |
| `-password` | string | DB 비밀번호 |
| `-tables` | string | 쉼표로 구분된 테이블 목록 (예: `USERS,ORDERS`) |
| `-out` | string | 출력 파일명 (기본값: `migration.sql`) |
| `-batch` | int | `INSERT` 문당 행 수 (기본값: `1000`) |
| `-per-table` | bool | 별도의 파일 생성: `<TABLE>_migration.sql` |
| `-parallel` | bool | 여러 테이블을 동시에 처리 |

## 6. 구현 세부 사항

### 6.1 병렬 처리 (`--parallel`)
- 테이블 처리 루틴의 완료를 추적하기 위해 `sync.WaitGroup`을 사용합니다.
- 필요한 경우 동시성을 제한합니다 (PRD에서는 제한을 명시하지 않았지만, 작업자 풀(Worker Pool)은 향후 개선 사항이 될 수 있습니다).

### 6.2 파일 처리 및 동기화
- **테이블별 모드(Per-Table Mode):** 각 고루틴이 자체 파일을 열고 씁니다. 테이블 작업자 간의 동기화가 필요하지 않습니다.
- **단일 파일 모드 + 병렬 처리:** 서로 다른 테이블의 일괄 삽입(Bulk Insert) 블록이 섞이지 않도록 공유 `io.Writer` 또는 `*os.File`을 `sync.Mutex`로 보호해야 합니다.

### 6.3 성능 최적화
- **버퍼링된 I/O:** 시스템 호출을 최소화하기 위해 모든 파일 작업에 `bufio.Writer`를 사용합니다.
- **메모리 관리:** 행은 하나씩 처리되며, 디스크에 플러시하기 전에 현재 행 배치의 내용만 메모리에 유지됩니다.

## 7. 보안 및 안전성
- **SQL 인젝션:** 이 도구는 수동 실행을 위한 스크립트를 생성하므로 문자열 값을 이스케이프해야 합니다. Oracle 드라이버는 값을 `interface{}`로 반환해야 하며, 그런 다음 안전하게 캐스팅하고 형식화합니다.
- **자격 증명 처리:** 비밀번호는 플래그를 통해 허용되어야 하지만, 향후 버전에서는 더 나은 보안을 위해 환경 변수를 지원해야 합니다.


### task.md

# 구현 작업: Oracle에서 PostgreSQL 데이터 마이그레이션 CLI

## 1단계: 프로젝트 설정 및 CLI 뼈대 구성
- [x] Go 모듈 초기화 (`go mod init`).
- [x] Oracle 드라이버 종속성 설치: `github.com/sijms/go-ora/v2`.
- [x] CLI 플래그 파싱 구현 (URL, User, Password, Tables, Out, Batch, PerTable, Parallel).
- [x] 필수 플래그 유효성 검사 및 기본 도움말/사용법 출력 제공.

## 2단계: 데이터베이스 연결 및 메타데이터
- [x] `sql.Open`을 사용하여 Oracle 연결 로직 구현.
- [x] 지정된 테이블 이름에 대한 열 이름과 타입을 가져오는 함수 구현.
- [x] 자격 증명과 도달 가능성을 확인하기 위한 견고한 연결 테스트 생성.

## 3단계: 핵심 데이터 추출 및 변환
- [x] `sql.Rows.Next()`를 사용하여 스트리밍 행 추출 구현.
- [x] 데이터 타입 매핑 로직 구현:
    - [x] 문자열 이스케이프 (VARCHAR2, CLOB).
    - [x] 숫자 포맷팅 (NUMBER).
    - [x] 타임스탬프 포맷팅 (DATE, TIMESTAMP).
    - [x] 이진 데이터에 대한 16진수 인코딩 (BLOB, RAW).
    - [x] NULL 처리.
- [x] 행을 `INSERT INTO` 문으로 그룹화하는 기본 배치 로직 구현.

## 4단계: 출력 관리
- [x] 단일 파일 출력 작성기 구현 (기본 모드).
- [x] `--per-table` 로직 구현: 테이블별로 개별 파일 생성.
- [x] 성능 향상을 위해 `bufio`를 사용한 버퍼링된 쓰기 구현.

## 5단계: 동시성 및 병렬 처리
- [x] `sync.WaitGroup`을 사용하여 `--parallel` 처리 구현.
- [x] `sync.Mutex`를 사용하여 단일 파일 모드에 대한 스레드 안전 쓰기 구현.
- [x] 고루틴에서 기본 프로세스로 적절한 오류 처리 및 전파 보장.

## 6단계: 검증 및 개선
- [x] 타입 변환 로직에 대한 단위 테스트 추가.
- [x] 통합 테스트 추가 (모의 객체 사용 또는 사용 가능한 경우 테스트 Oracle 인스턴스 사용).
- [x] 생성된 SQL을 PostgreSQL 대상에 대해 수동으로 검증.
- [x] 문서 업데이트 및 최종 코드 정리.


---
## <a name="v02"></a> v02

### prd.md

# 제품 요구사항 문서 (PRD): Oracle에서 PostgreSQL 마이그레이션 도구 v2

## 1. 개요
초기 CLI 도구의 성공을 바탕으로, v2는 프로덕션 규모의 마이그레이션을 지원하기 위해 견고성, 성능 및 기능을 향상시키는 것을 목표로 합니다. 이 버전은 직접 데이터베이스 삽입, 스키마 생성 및 더 나은 리소스 관리에 중점을 둡니다.

## 2. 목표
- **직접 마이그레이션(Direct Migration):** 중간 SQL 파일 없이 Oracle에서 PostgreSQL로의 직접 데이터 전송을 지원합니다.
- **스키마 자동 검색:** Oracle 테이블 구조를 기반으로 `CREATE TABLE` DDL을 생성합니다.
- **성능 향상:** 작업자 풀(Worker Pool)을 사용하여 동시 테이블 처리를 보다 효율적으로 관리합니다.
- **검증 강화:** 연결성을 검증하고 데이터 볼륨을 추정하기 위한 "예행 연습(dry run)" 모드를 제공합니다.

## 3. 새로운 기능

### 3.1 직접 PostgreSQL 삽입
- **새 플래그:** `--pg-url`
- 지정된 경우 도구는 대상 PostgreSQL 데이터베이스에 연결하고 `COPY` 또는 `INSERT` 명령을 직접 실행합니다.
- `github.com/lib/pq` 또는 `github.com/jackc/pgx/v5`를 사용합니다.

### 3.2 DDL 생성
- **새 플래그:** `--with-ddl`
- `INSERT` 문 전에 `CREATE TABLE` 문을 생성합니다.
- Oracle 타입을 가장 적절한 PostgreSQL 타입으로 자동 매핑합니다 (예: `VARCHAR2` -> `text`, `NUMBER(10,0)` -> `integer`).

### 3.3 병렬 처리를 위한 작업자 풀
- **새 플래그:** `--workers` (기본값: 4)
- 테이블당 하나의 고루틴을 생성하는 대신, 도구는 고정된 수의 작업자를 사용하여 테이블 대기열을 처리하므로 데이터베이스의 리소스 고갈을 방지합니다.

### 3.4 예행 연습(Dry Run) 모드
- **새 플래그:** `--dry-run`
- Oracle에 연결하고, 권한을 확인하고, 지정된 테이블의 행 수를 세고, 파일을 쓰거나 데이터를 삽입하지 않고 예상되는 마이그레이션 계획을 보고합니다.

### 3.5 구조화된 로깅
- 구조화되고 레벨이 지정된 로깅(JSON 또는 Text)을 위해 표준 `log`를 `log/slog`로 교체합니다.

## 4. 기술 요구사항
- **언어:** Go 1.22 이상
- **드라이버:**
  - Oracle: `github.com/sijms/go-ora/v2`
  - PostgreSQL: `github.com/jackc/pgx/v5`
- **동시성:** 채널(channels) 및 `sync.WaitGroup`을 사용하는 작업자 풀 패턴.

## 5. 범위 및 제약 사항
- 마이그레이션은 데이터 및 기본 스키마에 중점을 둡니다. 프로시저, 트리거 또는 뷰와 같은 복잡한 객체는 v2의 범위를 벗어납니다.
- 메모리 급증을 방지하기 위해 대용량 BLOB/CLOB 처리를 최적화해야 합니다 (스트리밍).


### spec.md

# 기술 명세서: Oracle에서 PostgreSQL 마이그레이션 도구 v2

## 1. 소개
이 명세서는 직접 데이터베이스 마이그레이션, 스키마 검색 및 최적화된 리소스 관리에 중점을 둔 마이그레이션 도구 v2의 기술 아키텍처 및 설계를 정의합니다.

## 2. 업데이트된 기술 스택
- **Go 1.22 이상**
- **Oracle 드라이버:** `github.com/sijms/go-ora/v2`
- **PostgreSQL 드라이버:** `github.com/jackc/pgx/v5`
- **로깅:** `log/slog` (구조화된 로깅)

## 3. 아키텍처 개요

### 3.1 마이그레이션 모드
1. **파일 기반 (v1 레거시):** Oracle -> 메모리 -> SQL 파일.
2. **직접 마이그레이션 (v2):** Oracle -> 메모리 -> PostgreSQL (`COPY` 또는 배치 `INSERT`).

### 3.2 컴포넌트 설계
- **디스패처(Dispatcher):** 테이블 목록을 읽고, 작업을 생성하며, 작업자 풀을 관리합니다.
- **작업자(Worker):** 채널에서 작업(테이블 이름)을 소비하고, Oracle에서 추출을 수행하며, 파일 또는 PostgreSQL에 쓰는 작업을 처리합니다.
- **DDL 생성기(DDL Generator):** Oracle 메타데이터를 쿼리하여 호환되는 PostgreSQL `CREATE TABLE` 문을 구성합니다.

## 4. 구현 세부 사항

### 4.1 작업자 풀 (Worker Pool)
- 작업 채널과 `sync.WaitGroup`을 사용하여 작업자 풀을 구현합니다.
- `--workers` 플래그는 동시 테이블 처리기의 수를 결정합니다.
- 경합을 피하기 위해 각 작업자는 고유한 Oracle 및 (선택적으로) PostgreSQL 연결을 유지하거나, 스레드 안전 풀을 사용합니다.

### 4.2 직접 PostgreSQL 마이그레이션
- `pgx.Conn` 또는 `pgxpool.Pool`을 사용합니다.
- **기본 방법:** 고성능 대량 로드를 위해 `pgx.Conn.CopyFrom`을 통한 `COPY` 명령.
- **대체 방법:** 매개변수화된 쿼리를 사용한 배치 `INSERT` 문.

### 4.3 DDL 매핑 (Oracle에서 PostgreSQL로)
| Oracle 타입 | PostgreSQL 타입 | 참고 |
| --- | --- | --- |
| `NUMBER(*, 0)` | `bigint` / `integer` | 정밀도 기반. |
| `NUMBER(*, >0)` | `numeric` | |
| `VARCHAR2(n)`, `NVARCHAR2(n)` | `text` 또는 `varchar(n)` | |
| `DATE`, `TIMESTAMP` | `timestamp` | |
| `CLOB` | `text` | |
| `BLOB`, `RAW` | `bytea` | |

### 4.4 예행 연습 (Dry Run) 로직
- `--dry-run`이 활성화된 경우:
  - Oracle에 대한 연결을 설정합니다.
  - 각 테이블에 대해 `SELECT COUNT(*) FROM table`을 쿼리합니다.
  - 보고: "테이블 X: 약 Y개의 행이 마이그레이션됩니다."
  - 출력 파일을 열거나 대상에 삽입을 실행하지 않습니다.

### 4.5 구조화된 로깅 (`slog`)
- JSON 플래그가 설정된 경우 `slog.New(slog.NewJSONHandler(os.Stdout, nil))`로 전역 로거를 초기화하고, 그렇지 않은 경우 TextHandler를 사용합니다.
- 로그 컨텍스트: `slog.Info("processing table", "table", tableName, "status", "started")`.

## 5. 보안 및 안전성
- **PostgreSQL DSN:** `--pg-url` 플래그 또는 `PG_URL` 환경 변수를 통해 처리합니다.
- **트랜잭션 안전성:** 직접 삽입의 경우, 테이블당 원자적 결과를 보장하기 위해 각 테이블 마이그레이션을 트랜잭션으로 래핑하는 것을 고려하십시오.


### task.md

# 구현 작업: Oracle에서 PostgreSQL 마이그레이션 도구 v2

## 1단계: 종속성 관리 및 인프라
- [x] `go.mod`에서 Go를 1.22 이상으로 업그레이드.
- [x] PostgreSQL 드라이버 설치: `github.com/jackc/pgx/v5`.
- [x] 애플리케이션 전반에 걸쳐 구조화된 로깅을 위해 `log/slog` 구현.
- [x] 새로운 플래그(`--pg-url`, `--workers`, `--with-ddl`, `--dry-run`)를 처리하기 위한 구성(configuration) 구조체 구현.
- [x] 내부 패키지로 코드베이스 모듈화.

## 2단계: 작업자 풀 및 동시성
- [x] `n`개의 작업자 풀을 관리하기 위한 `Dispatcher` ( `Run` 함수 내) 구현.
- [x] `Job` 구조체 및 스레드 안전 작업자 메커니즘 구현.
- [x] 테이블 처리를 위한 단순한 `sync.WaitGroup` 루프를 작업자 풀로 교체.
- [x] 적절하고 우아한 종료 보장 (Run이 반환되기 전에 모든 작업자가 완료됨).

## 3단계: 직접 마이그레이션 구현
- [x] `pgxpool`을 사용한 PostgreSQL 연결 풀 관리 구현.
- [x] 고속 데이터 전송을 위해 `pgx.Conn.CopyFrom`을 사용하는 `DirectWriter` ( `MigrateTableDirect` 내) 구현.
- [ ] 호환성을 위해 매개변수화된 쿼리를 사용하는 대체(fallback) 배치 `INSERT` 메커니즘 구현.
- [x] 테이블 마이그레이션별 트랜잭션 지원 추가.

## 4단계: 스키마 및 DDL 생성
- [x] 정밀도, 스케일 및 제약 조건을 위한 Oracle 메타데이터 검색 ( `GetTableMetadata` 내) 구현.
- [x] Oracle 타입에서 PostgreSQL 타입으로의 매핑 함수 (`MapOracleToPostgres`) 구현.
- [x] `CREATE TABLE` 스크립트 생성 로직 (`GenerateCreateTableDDL`) 구현.
- [x] 데이터 삽입 전에 DDL을 실행하기 위한 `--with-ddl` 실행 흐름 추가.

## 5단계: 예행 연습 및 검증
- [x] 연결성을 검증하고 예상 행 수를 보고하기 위한 `--dry-run` 로직 ( `Run` 내) 구현.
- [x] 마이그레이션을 시작하기 전에 대상 테이블이 존재하는지 확인하는 검증 ( `MigrateTableDirect` 내) 구현.
- [x] 사전 점검 (풀 생성 및 dry-run 중 연결성 확인) 추가.

## 6단계: 테스트 및 품질 보증
- [x] `slog`를 사용하도록 단위 테스트 업데이트 (패키지 리팩토링을 통해 암시적으로 수행됨).
- [x] 작업자 풀 및 작업 디스패칭을 위한 새로운 단위 테스트 추가 (`worker_test.go`).
- [x] Oracle 및 PostgreSQL을 모두 시뮬레이션하기 위해 `pgx` 및 `sqlmock`을 사용하는 통합 테스트 추가 (`direct_test.go`).
- [ ] 파일 기반 마이그레이션과 직접 마이그레이션을 비교하는 성능 벤치마킹 수행.
- [x] `README.md`의 문서 및 예제 업데이트.


---
## <a name="v03"></a> v03

### prd.md

# PRD (Product Requirements Document) - Web UI Addition (v3)

## 1. 개요 (Overview)
기존 CLI 전용으로 동작하던 Oracle to PostgreSQL 데이터 마이그레이션 도구(dbmigrator)에 **Web UI 모드**를 추가합니다. 사용자는 로컬 환경에서 웹 브라우저를 통해 마이그레이션 작업을 보다 직관적으로 설정, 실행, 모니터링할 수 있습니다. 생성된 SQL 스크립트 파일들은 최종적으로 압축 파일(.zip) 형태로 제공됩니다.

## 2. 목표 (Goals)
- **사용자 편의성 향상**: 터미널 명령어를 외우지 않고도 웹 UI에서 DB 접속 정보를 입력하고 작업을 실행할 수 있게 합니다.
- **가시성 확보**: 다수의 테이블을 병렬로 마이그레이션하는 과정을 실시간 프로그레스 바(Progress Bar)와 처리 건수(Count)로 시각화합니다.
- **테이블 선택 기능**: 접속한 Oracle DB의 전체 테이블 목록을 조회하고, `LIKE` 검색을 통해 원하는 테이블만 필터링하여 선택할 수 있는 기능을 제공합니다.
- **결과물 다운로드**: 마이그레이션 결과물인 다수의 `.sql` 파일들을 하나의 `.zip` 파일로 묶어서 웹에서 즉시 다운로드할 수 있게 합니다.
- **호환성 유지**: 기존 CLI 방식의 실행 옵션은 그대로 유지합니다.

## 3. 기능 요구사항 (Functional Requirements)

### 3.1. 실행 모드
- `dbmigrator --web` (또는 유사한 플래그/명령) 실행 시 내장 웹 서버(Gin 기반)가 구동됩니다.
- CLI 실행 방식은 기존과 동일하게 유지됩니다.

### 3.2. 데이터베이스 연결 및 테이블 검색
- 사용자는 Web UI 폼에 Oracle DB 접속 정보(`URL`, `Username`, `Password`)를 입력합니다.
- 사용자가 "연결" 버튼을 클릭하면, 서버는 DB에 접속하여 테이블 목록을 가져옵니다.
- **테이블 검색 (LIKE 검색)**: 입력 폼에 `테이블명 검색 (LIKE)` 필드를 제공하여, 특정 패턴(예: `USER_%`)에 일치하는 테이블만 필터링하여 화면에 리스트업할 수 있습니다.
- 조회된 테이블 목록은 체크박스 형태로 표시되며, 사용자는 마이그레이션할 테이블을 다중 선택할 수 있습니다.

### 3.3. 마이그레이션 실행 및 실시간 모니터링 (WebSocket)
- 사용자가 선택한 테이블들에 대해 마이그레이션(SQL 스크립트 생성) 작업을 시작합니다.
- 작업은 기존 로직을 활용하여 **병렬(Parallel)**로 처리됩니다.
- **WebSocket 연동**: 서버는 작업 진행 상황을 WebSocket을 통해 브라우저로 실시간 푸시(Push)합니다.
- **UI 표시**: 브라우저는 수신된 데이터를 바탕으로 각 테이블별 진행률(%)을 **프로그레스 바**로 표시하고, 현재까지 **처리된 레코드 수(Count) / 총 레코드 수**를 텍스트로 함께 보여줍니다.

### 3.4. 결과물 다운로드 (ZIP)
- 마이그레이션이 완료된 SQL 스크립트 파일들은 서버의 임시 디렉토리(Temp)에 생성되거나 메모리 버퍼에서 직접 압축됩니다.
- 모든 작업이 성공적으로 완료되면 `.zip` 파일로 압축됩니다.
- Web UI에 다운로드 버튼이 활성화되며, 사용자는 해당 버튼을 클릭하여 압축 파일을 로컬 PC로 다운로드합니다.
- 다운로드 후 임시 파일들은 정리(Cleanup)되어야 합니다.

## 4. 비기능 요구사항 (Non-Functional Requirements)
- **환경**: 단일 사용자(Single User) 로컬 구동 환경으로 가정합니다. 복잡한 사용자 인증(Authentication)이나 세션 관리는 제외합니다.
- **기술 스택**:
  - Backend: Go (Gin 웹 프레임워크)
  - Frontend: 기본 HTML, Vanilla JavaScript, CSS (Go `html/template` 활용)
  - 통신: REST API (테이블 조회 등), WebSocket (실시간 진행 상황)
- **성능**: 대용량 테이블 마이그레이션 시 서버 메모리 고갈이나 브라우저 렌더링 지연이 발생하지 않도록 주의합니다. (WebSocket 이벤트 발생 주기 조절 등)

## 5. UI/UX 구성 (초안)
1. **DB 연결 섹션**:
   - `DB URL` (예: localhost:1521/XE)
   - `Username`
   - `Password`
   - `테이블명 검색 (LIKE, 선택사항)`
   - [연결 및 테이블 조회] 버튼
2. **테이블 선택 섹션**:
   - (조회된 테이블 목록이 체크박스와 함께 표시됨)
   - [전체 선택/해제] 체크박스
   - [마이그레이션 시작] 버튼
3. **진행 상황 모니터링 섹션**:
   - 선택된 테이블들의 목록과 각 테이블 우측에 실시간 프로그레스 바 표시.
   - 텍스트 예시: `USERS: [████████░░] 80% (8,000 / 10,000)`
4. **결과 다운로드 섹션**:
   - 모든 프로그레스 바가 100% 도달 시 표시.
   - [결과물 다운로드 (.zip)] 버튼

## 6. 마일스톤 (Milestones)
1. `docs/v3/prd.md` 작성 및 요구사항 확정 (완료)
2. Gin 웹 서버 초기화 및 정적/템플릿 파일 라우팅 설정
3. DB 연결 및 테이블 목록 조회(LIKE 검색 포함) API 구현
4. Web UI 뼈대 작성 (HTML/JS) 및 조회 API 연동
5. 마이그레이션 백그라운드 작업 및 WebSocket 실시간 진행률 연동
6. ZIP 압축 생성 및 다운로드 API 구현
7. Web UI 최종 연동 (프로그레스 바 시각화, 다운로드 버튼 활성화)
8. 기존 CLI 명령어 호환성 테스트 및 `--web` 옵션 분기 처리
9. 최종 리뷰 및 코드 정리


### spec.md

# Technical Specification: Web UI Addition (v3)

## 1. 개요 (Introduction)
v3에서는 기존 CLI 기반의 dbmigrator에 사용자 친화적인 웹 인터페이스를 추가합니다. Go의 Gin 프레임워크를 사용하여 경량 웹 서버를 구축하고, WebSocket을 통해 실시간 마이그레이션 진행 상태를 시각화합니다.

## 2. 기술 스택 (Technical Stack)
- **Backend**: Go 1.21+, Gin Web Framework
- **Frontend**: HTML5, Vanilla JavaScript, CSS3 (Google Fonts 'Inter')
- **Real-time Communication**: WebSockets (`github.com/gorilla/websocket`)
- **Zip Generation**: Standard `archive/zip` library

## 3. 상세 설계 (Detailed Design)

### 3.1. Web Server Architecture
- **Framework**: `gin-gonic/gin`을 사용하여 라우팅 및 미들웨어를 관리합니다.
- **Static Assets**: HTML 템플릿은 `web/templates`에, 정적 자산은 `web/static`에 위치합니다. (현재는 `index.html`에 스타일과 스크립트가 내장된 형태)
- **Concurrency**: 각 마이그레이션 요청은 별도의 고루틴에서 실행되어 서버의 응답성을 유지합니다.

### 3.2. API 엔드포인트
- `POST /api/tables`: Oracle DB 연결 정보를 받아 테이블 목록을 반환합니다. (`LIKE` 검색 지원)
- `POST /api/migrate`: 선택된 테이블들에 대해 마이그레이션을 시작합니다. (비동기 처리)
- `GET /api/progress`: WebSocket 연결을 통해 실시간 이벤트를 전송합니다.
- `GET /api/download/:id`: 생성된 ZIP 파일을 다운로드합니다.

### 3.3. WebSocket 프로토콜 명세 (JSON)
- **Init**: `{"type": "init", "table": "NAME", "total": 1000}`
- **Update**: `{"type": "update", "table": "NAME", "count": 500}`
- **Done**: `{"type": "done", "table": "NAME"}`
- **Error**: `{"type": "error", "table": "NAME", "error": "MSG"}`
- **All Done**: `{"type": "all_done", "zip_file_id": "FILENAME.zip"}`

### 3.4. 마이그레이션 및 ZIP 처리
- 웹 모드에서의 마이그레이션은 항상 `PerTable: true` 옵션을 사용합니다.
- 결과물은 `os.TempDir()` 하위의 임시 디렉토리에 생성됩니다.
- 모든 테이블 작업 완료 후 `ziputil`을 통해 디렉토리를 압축하고, 원본 SQL 파일들은 즉시 삭제합니다.
- ZIP 파일은 다운로드 후 약 5분 뒤에 자동 삭제되도록 스케줄링됩니다.

## 4. 보안 고려사항
- **경로 트래버스 방지**: 다운로드 API에서 `filepath.Base()`를 사용하여 권한 없는 파일 접근을 차단합니다.
- **제한된 로컬 환경**: 본 도구는 로컬 구동용이며, 외부 노출 시 추가적인 인증 레이어가 필요합니다.


### task.md

# Implementation Tasks: Web UI Addition (v3)

## Phase 1: Web Server Infrastructure
- [x] Gin 프레임워크 의존성 추가 및 초기화
- [x] `--web` 플래그 및 실행 모드 분기 로직 구현
- [x] HTML 템플릿 및 정적 파일 라우팅 설정
- [x] 프로젝트 레이아웃 구성 (`web/templates`, `internal/web`)

## Phase 2: Core Table Management API
- [x] Oracle DB 연결 및 테이블 목록 조회 API (`POST /api/tables`)
- [x] `LIKE` 필터를 통한 테이블 검색 기능 구현
- [x] 프론트엔드 테이블 리스트 렌더링 및 체크박스 선택 로직

## Phase 3: Real-time Progress Tracking (WebSocket)
- [x] WebSocket Tracker 구현 (`internal/web/ws`)
- [x] `migration.Run`에 `ProgressTracker` 인터페이스 도입 및 연동
- [x] 프론트엔드 WebSocket 클라이언트 구현 및 프로그레스 바 시각화
- [x] 실시간 처리 건수 표시 기능

## Phase 4: Output Management & ZIP
- [x] ZIP 압축 유틸리티 구현 (`internal/web/ziputil`)
- [x] 작업 완료 후 자동 ZIP 압축 및 임시 SQL 파일 정리 로직
- [x] ZIP 파일 다운로드 API 구현 (`GET /api/download/:id`)
- [x] 다운로드 후 일정 시간 뒤 ZIP 파일 자동 삭제 (Cleanup)

## Phase 5: UI/UX Refinement & Next Steps
- [x] Vanilla JS 기반의 현대적 UI 디자인 (Glassmorphism, Inter font)
- [x] 에러 핸들링 및 사용자 알림 UI
- [x] Web UI에서 PostgreSQL 직접 마이그레이션(Direct Copy) 옵션 추가
- [x] UI에서 Batch Size 및 Worker 수 설정 기능 추가
- [x] 대용량 테이블 처리 시 WebSocket 이벤트 Throttling 최적화


---
## <a name="v04"></a> v04

### prd.md

# PRD (Product Requirements Document) - Web UI CLI 파라미터 통합 (v4)

## 1. 개요 (Overview)

현재 Web UI에서는 CLI 모드에서 제공하는 일부 파라미터만 제어 가능합니다. CLI 전용으로 남아있는 5개 파라미터(`-out`, `-per-table`, `-schema`, `-dry-run`, `-log-json`)를 Web UI에서도 표시 및 제어할 수 있도록 확장합니다.

## 2. 배경 (Background)

### 2.1. 현재 상태

| 파라미터 | CLI 플래그 | Web UI 지원 | 비고 |
|----------|-----------|-------------|------|
| Oracle URL | `-url` | O | 연결 폼 |
| Username | `-user` | O | 연결 폼 |
| Password | `-password` | O | 연결 폼 |
| Tables | `-tables` | O | 체크박스 선택 |
| Batch Size | `-batch` | O | 고급 설정 |
| Workers | `-workers` | O | 고급 설정 |
| PostgreSQL URL | `-pg-url` | O | Direct 모드 |
| DDL 생성 | `-with-ddl` | O | Direct 모드 체크박스 |
| **Output 파일명** | `-out` | **X** | 기본값 `migration.sql` |
| **테이블별 파일 분리** | `-per-table` | **X** | 기본값 `false` |
| **PostgreSQL 스키마** | `-schema` | **X** | 기본값 빈 문자열 |
| **Dry-Run 모드** | `-dry-run` | **X** | 기본값 `false` |
| **JSON 로깅** | `-log-json` | **X** | 기본값 `false` |

### 2.2. 문제점
- Web UI 사용자는 출력 파일명을 지정하거나 테이블별 파일 분리를 선택할 수 없음
- PostgreSQL 스키마를 지정하려면 CLI로 전환해야 함
- 실제 마이그레이션 전 사전 검증(Dry-Run)을 Web UI에서 수행할 수 없음
- 디버깅용 JSON 로그 출력을 Web UI에서 활성화할 수 없음

## 3. 목표 (Goals)

- **CLI-Web 기능 동등성(Feature Parity)**: CLI에서 가능한 모든 설정을 Web UI에서도 동일하게 제어 가능하게 합니다.
- **직관적인 UI 배치**: 기존 고급 설정(Advanced Settings) 섹션을 확장하여 새 파라미터를 자연스럽게 배치합니다.
- **하위 호환성 유지**: 기존 API 요청에 새 필드가 없는 경우 기본값으로 동작합니다.

## 4. 기능 요구사항 (Functional Requirements)

### 4.1. 출력 파일명 설정 (`-out`)

- **UI 컨트롤**: 텍스트 입력 필드
- **위치**: 고급 설정(Advanced Settings) 섹션 내
- **기본값**: `migration.sql`
- **적용 범위**: SQL File 모드에서만 활성화 (Direct 모드에서는 비활성화/숨김)
- **동작**: ZIP 파일 내 생성되는 SQL 파일의 기본 이름으로 사용됨. `-per-table` 활성화 시에는 접두사(prefix)로 활용되거나 무시될 수 있음

### 4.2. 테이블별 파일 분리 (`-per-table`)

- **UI 컨트롤**: 토글 스위치 또는 체크박스
- **위치**: 고급 설정 섹션 내
- **기본값**: `false` (비활성화)
- **적용 범위**: SQL File 모드에서만 활성화
- **동작**:
  - 비활성화 시: 모든 테이블의 INSERT 문이 단일 SQL 파일에 기록됨
  - 활성화 시: 각 테이블별로 별도의 `.sql` 파일이 생성됨 (예: `USERS.sql`, `ORDERS.sql`)
- **참고**: 현재 Web UI는 내부적으로 항상 `PerTable=true`로 동작 중. 이 옵션을 노출하여 사용자가 단일 파일 출력도 선택 가능하게 함

### 4.3. PostgreSQL 스키마 지정 (`-schema`)

- **UI 컨트롤**: 텍스트 입력 필드
- **위치**: 고급 설정 섹션 내
- **기본값**: 빈 문자열 (미지정)
- **Placeholder**: `public`
- **적용 범위**: SQL File 모드, Direct 모드 모두 적용
- **동작**:
  - 값이 지정되면 생성되는 DDL 및 INSERT 문에서 테이블명 앞에 스키마가 붙음 (예: `myschema.USERS`)
  - `-with-ddl` 활성화 시 `CREATE TABLE` 문에도 스키마가 반영됨
  - 빈 값인 경우 기존 동작과 동일 (스키마 미지정)

### 4.4. Dry-Run 모드 (`-dry-run`)

- **UI 컨트롤**: 토글 스위치 또는 체크박스
- **위치**: 마이그레이션 시작 버튼 근처 (잘 보이는 위치)
- **기본값**: `false` (비활성화)
- **동작**:
  - 활성화 시: 실제 데이터 마이그레이션 없이 다음을 수행
    - Oracle DB 연결 확인
    - PostgreSQL DB 연결 확인 (Direct 모드 시)
    - 각 테이블의 총 레코드 수(Row Count) 추정
  - 결과를 WebSocket을 통해 실시간으로 UI에 표시
  - **UI 변화**:
    - Dry-Run 활성화 시 "마이그레이션 시작" 버튼 텍스트가 "사전 검증 실행"으로 변경
    - 진행 상황 섹션에 "Dry-Run 모드" 라벨 표시
    - 완료 후 다운로드 버튼 대신 검증 결과 요약 표시

### 4.5. JSON 로깅 (`-log-json`)

- **UI 컨트롤**: 토글 스위치 또는 체크박스
- **위치**: 고급 설정 섹션 내
- **기본값**: `false` (비활성화)
- **동작**:
  - 활성화 시: 서버 측 로그 출력이 구조화된 JSON 형식으로 전환됨
  - 디버깅 및 로그 수집 시스템 연동에 유용
- **선택적 확장**: 향후 Web UI에 로그 뷰어 패널을 추가하여 JSON 로그를 실시간으로 표시할 수 있음

## 5. API 변경사항 (API Changes)

### 5.1. `POST /api/migrate` 요청 구조 확장

```json
{
  "oracleUrl": "localhost:1521/XE",
  "username": "user",
  "password": "pass",
  "tables": ["USERS", "ORDERS"],
  "direct": false,
  "pgUrl": "",
  "withDdl": true,
  "batchSize": 1000,
  "workers": 4,
  "outFile": "migration.sql",
  "perTable": true,
  "schema": "public",
  "dryRun": false,
  "logJson": false
}
```

### 5.2. 서버 측 `startMigrationRequest` 구조체 변경

```go
type startMigrationRequest struct {
    OracleURL string   `json:"oracleUrl" binding:"required"`
    Username  string   `json:"username" binding:"required"`
    Password  string   `json:"password" binding:"required"`
    Tables    []string `json:"tables" binding:"required"`
    Direct    bool     `json:"direct"`
    PGURL     string   `json:"pgUrl"`
    WithDDL   bool     `json:"withDdl"`
    BatchSize int      `json:"batchSize"`
    Workers   int      `json:"workers"`
    // v4 추가 필드
    OutFile   string   `json:"outFile"`
    PerTable  bool     `json:"perTable"`
    Schema    string   `json:"schema"`
    DryRun    bool     `json:"dryRun"`
    LogJSON   bool     `json:"logJson"`
}
```

### 5.3. WebSocket 메시지 확장 (Dry-Run 모드)

Dry-Run 모드 전용 메시지 타입 추가:

```json
{
  "type": "dry_run_result",
  "table": "USERS",
  "total": 150000,
  "estimatedTime": "~2m30s",
  "connectionOk": true
}
```

## 6. UI/UX 설계 (UI/UX Design)

### 6.1. 고급 설정 섹션 확장

기존 고급 설정(Advanced Settings) `<details>` 요소 내부에 다음 항목을 추가합니다:

```
[고급 설정] ▾
┌─────────────────────────────────────────────────┐
│  Batch Size:     [1000        ]                 │
│  Workers:        [4           ]                 │
│  ─────────────────────────────────               │
│  Output 파일명:  [migration.sql]  ← NEW         │
│  PG 스키마:      [public      ]  ← NEW         │
│  ─────────────────────────────────               │
│  ☐ 테이블별 파일 분리              ← NEW         │
│  ☐ JSON 로깅                      ← NEW         │
└─────────────────────────────────────────────────┘
```

### 6.2. Dry-Run 컨트롤

마이그레이션 시작 버튼 영역에 Dry-Run 토글을 배치합니다:

```
┌─────────────────────────────────────────────────┐
│  ☐ Dry-Run (사전 검증만 수행)      ← NEW        │
│                                                 │
│  [ 🚀 마이그레이션 시작 ]                         │
│  ──── Dry-Run 활성화 시 ────                     │
│  [ 🔍 사전 검증 실행 ]                            │
└─────────────────────────────────────────────────┘
```

### 6.3. 조건부 표시 규칙

| 컨트롤 | SQL File 모드 | Direct 모드 |
|--------|--------------|-------------|
| Output 파일명 | 표시 | 숨김 |
| 테이블별 파일 분리 | 표시 | 숨김 |
| PG 스키마 | 표시 | 표시 |
| Dry-Run | 표시 | 표시 |
| JSON 로깅 | 표시 | 표시 |

## 7. 비기능 요구사항 (Non-Functional Requirements)

- **하위 호환성**: 새 필드가 요청에 포함되지 않을 경우 기존 기본값으로 동작해야 합니다.
- **유효성 검증**:
  - `outFile`에 경로 구분자(`/`, `\`)가 포함된 경우 거부
  - `schema`에 SQL 인젝션 가능한 문자가 포함된 경우 거부
  - `batchSize`, `workers`는 양의 정수만 허용
- **반응형 UI**: 새 컨트롤들이 모바일/작은 화면에서도 정상 표시되어야 합니다.
- **접근성**: 모든 입력 필드에 적절한 `<label>` 및 `aria-` 속성을 부여합니다.

## 8. 영향 범위 (Scope of Changes)

### 8.1. 변경 파일 목록

| 파일 | 변경 내용 |
|------|----------|
| `internal/web/server.go` | `startMigrationRequest` 구조체 확장, 핸들러 로직 수정 |
| `internal/web/templates/index.html` | 고급 설정 UI 확장, Dry-Run 토글, 조건부 표시 로직 |
| `internal/web/ws/tracker.go` | Dry-Run 결과 메시지 타입 추가 |
| `internal/migration/migration.go` | Dry-Run 로직 분기 처리 |
| `internal/logger/logger.go` | 런타임 로그 모드 전환 지원 (선택적) |

### 8.2. 변경하지 않는 파일

| 파일 | 사유 |
|------|------|
| `internal/config/config.go` | CLI 플래그 파싱은 기존 유지, Web API에서 직접 매핑 |
| `main.go` | 진입점 변경 없음 |

## 9. 마일스톤 (Milestones)

1. **PRD 확정**: `docs/v4/prd.md` 작성 및 리뷰
2. **API 확장**: `startMigrationRequest` 구조체 및 핸들러 수정
3. **UI 확장**: 고급 설정 섹션에 새 컨트롤 추가
4. **Dry-Run 구현**: 사전 검증 로직 및 WebSocket 메시지 확장
5. **조건부 표시**: SQL File / Direct 모드에 따른 UI 토글 로직
6. **유효성 검증**: 입력값 검증 로직 추가 (프론트엔드 + 백엔드)
7. **테스트**: 단위 테스트 및 통합 테스트 작성
8. **최종 리뷰**: 코드 리뷰 및 문서 정리

## 10. 향후 확장 고려사항 (Future Considerations)

- **로그 뷰어**: Web UI에 실시간 로그 패널을 추가하여 JSON 로그를 시각적으로 확인
- **설정 프리셋**: 자주 사용하는 설정 조합을 저장/로드하는 기능
- **웹 서버 포트 설정**: 현재 하드코딩된 `8080` 포트를 Web UI에서 설정 가능하게 확장


### spec.md

# Technical Specification: Web UI CLI 파라미터 통합 (v4)

## 1. 개요 (Introduction)

v4에서는 CLI 전용으로 남아있던 5개 파라미터(`-out`, `-per-table`, `-schema`, `-dry-run`, `-log-json`)를 Web UI에서 제어할 수 있도록 확장합니다. 백엔드 API 구조체 확장, 프론트엔드 UI 컨트롤 추가, Dry-Run 전용 WebSocket 메시지 타입 추가가 핵심 변경사항입니다.

## 2. 변경 파일 및 영향 범위

| 파일 | 변경 유형 | 설명 |
|------|----------|------|
| `internal/web/server.go` | 수정 | 요청 구조체 확장, Config 매핑 로직 수정, 입력값 검증 추가 |
| `internal/web/templates/index.html` | 수정 | 고급 설정 UI 확장, Dry-Run 토글, 조건부 표시, JS 로직 |
| `internal/web/ws/tracker.go` | 수정 | `MsgDryRunResult` 메시지 타입 및 `DryRunResult()` 메서드 추가 |
| `internal/migration/migration.go` | 수정 | Dry-Run 시 WebSocket tracker 연동 |
| `internal/logger/logger.go` | 수정 | 런타임 로그 모드 전환 함수 추가 |

## 3. 상세 설계 (Detailed Design)

### 3.1. 백엔드 API 변경

#### 3.1.1. `startMigrationRequest` 구조체 확장

**파일**: `internal/web/server.go:83`

기존 구조체에 5개 필드를 추가합니다:

```go
type startMigrationRequest struct {
    OracleURL string   `json:"oracleUrl" binding:"required"`
    Username  string   `json:"username" binding:"required"`
    Password  string   `json:"password" binding:"required"`
    Tables    []string `json:"tables" binding:"required"`
    Direct    bool     `json:"direct"`
    PGURL     string   `json:"pgUrl"`
    WithDDL   bool     `json:"withDdl"`
    BatchSize int      `json:"batchSize"`
    Workers   int      `json:"workers"`
    // v4 추가 필드
    OutFile  string `json:"outFile"`
    PerTable bool   `json:"perTable"`
    Schema   string `json:"schema"`
    DryRun   bool   `json:"dryRun"`
    LogJSON  bool   `json:"logJson"`
}
```

**하위 호환성**: 새 필드는 모두 JSON 바인딩 시 zero value가 기본값이므로, 기존 클라이언트가 이 필드를 생략해도 동작에 영향 없음.

#### 3.1.2. 입력값 검증 함수

**파일**: `internal/web/server.go` (새 함수)

```go
func validateMigrationRequest(req *startMigrationRequest) error {
    // outFile: 경로 구분자 차단
    if strings.ContainsAny(req.OutFile, "/\\") {
        return fmt.Errorf("outFile must not contain path separators")
    }
    // schema: SQL 인젝션 방지 (영문, 숫자, 언더스코어만 허용)
    if req.Schema != "" && !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(req.Schema) {
        return fmt.Errorf("schema name contains invalid characters")
    }
    // batchSize, workers: 음수 방지 (0은 기본값 적용으로 허용)
    if req.BatchSize < 0 {
        return fmt.Errorf("batchSize must be non-negative")
    }
    if req.Workers < 0 {
        return fmt.Errorf("workers must be non-negative")
    }
    return nil
}
```

#### 3.1.3. `startMigration` 핸들러 수정

**파일**: `internal/web/server.go:95` (`startMigration` 함수)

변경 포인트:

1. **검증 호출**: `ShouldBindJSON` 이후 `validateMigrationRequest` 호출
2. **기본값 처리**: `outFile`이 빈 문자열이면 `"migration.sql"` 설정
3. **Config 매핑 확장**: 기존 하드코딩된 `PerTable: true`를 `req.PerTable`로 변경
4. **LogJSON 처리**: `req.LogJSON`이 `true`이면 `logger.SetJSONMode(true)` 호출
5. **DryRun 분기**: `req.DryRun`이 `true`이면 Dry-Run 전용 로직 실행

```go
func startMigration(c *gin.Context) {
    var req startMigrationRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
        return
    }

    if err := validateMigrationRequest(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    go func() {
        oracleDB, err := db.ConnectOracle(req.OracleURL, req.Username, req.Password)
        if err != nil {
            log.Printf("Failed to connect to Oracle: %v", err)
            tracker.AllDone("")
            return
        }
        defer oracleDB.Close()

        // LogJSON 모드 전환
        if req.LogJSON {
            logger.SetJSONMode(true)
            defer logger.SetJSONMode(false)
        }

        var pgPool db.PGPool
        if req.Direct && req.PGURL != "" {
            pgPool, err = db.ConnectPostgres(req.PGURL)
            if err != nil {
                log.Printf("Failed to connect to Postgres: %v", err)
                tracker.AllDone("")
                return
            }
            defer pgPool.Close()
        }

        workers := req.Workers
        if workers <= 0 {
            workers = 4
        }
        batchSize := req.BatchSize
        if batchSize <= 0 {
            batchSize = 1000
        }
        outFile := req.OutFile
        if outFile == "" {
            outFile = "migration.sql"
        }

        jobID := time.Now().Format("20060102150405")
        outDir := filepath.Join(os.TempDir(), "dbmigrator_"+jobID)
        if !req.Direct && !req.DryRun {
            if err := os.MkdirAll(outDir, 0755); err != nil {
                log.Printf("Failed to create temp directory: %v", err)
                return
            }
        }

        cfg := &config.Config{
            Tables:    req.Tables,
            Parallel:  true,
            Workers:   workers,
            BatchSize: batchSize,
            PerTable:  req.PerTable,
            OutFile:   outFile,
            Schema:    req.Schema,
            DryRun:    req.DryRun,
            OutputDir: outDir,
            PGURL:     req.PGURL,
            WithDDL:   req.WithDDL,
        }

        err = migration.Run(oracleDB, pgPool, cfg, tracker)
        if err != nil {
            log.Printf("Migration failed: %v", err)
            tracker.AllDone("")
        } else if req.DryRun {
            // Dry-Run은 다운로드 없이 완료
            tracker.AllDone("")
        } else if !req.Direct {
            zipFilePath := filepath.Join(os.TempDir(), "migration_"+jobID+".zip")
            if err := ziputil.ZipDirectory(outDir, zipFilePath); err != nil {
                log.Printf("Failed to create zip: %v", err)
                tracker.AllDone("")
            } else {
                tracker.AllDone("migration_" + jobID + ".zip")
            }
        } else {
            tracker.AllDone("")
        }

        if !req.Direct && !req.DryRun {
            os.RemoveAll(outDir)
        }
    }()

    c.JSON(http.StatusOK, gin.H{"message": "Migration started"})
}
```

### 3.2. WebSocket 프로토콜 확장

#### 3.2.1. Dry-Run 결과 메시지 타입

**파일**: `internal/web/ws/tracker.go`

```go
const (
    MsgInit         MsgType = "init"
    MsgUpdate       MsgType = "update"
    MsgDone         MsgType = "done"
    MsgError        MsgType = "error"
    MsgAllDone      MsgType = "all_done"
    MsgDryRunResult MsgType = "dry_run_result"  // NEW
)

type ProgressMsg struct {
    Type         MsgType `json:"type"`
    Table        string  `json:"table,omitempty"`
    Count        int     `json:"count,omitempty"`
    Total        int     `json:"total,omitempty"`
    ErrorMsg     string  `json:"error,omitempty"`
    ZipFileID    string  `json:"zip_file_id,omitempty"`
    ConnectionOk bool    `json:"connection_ok,omitempty"` // NEW: Dry-Run 연결 확인 결과
}
```

#### 3.2.2. `DryRunResult` 메서드 추가

```go
func (t *WebSocketTracker) DryRunResult(table string, totalRows int, connectionOk bool) {
    t.broadcast(ProgressMsg{
        Type:         MsgDryRunResult,
        Table:        table,
        Total:        totalRows,
        ConnectionOk: connectionOk,
    })
}
```

### 3.3. 마이그레이션 로직 수정 (Dry-Run + WebSocket)

**파일**: `internal/migration/migration.go:40-53`

현재 Dry-Run 로직은 `slog`로만 출력합니다. WebSocket tracker가 있을 경우 tracker를 통해 UI에 결과를 전달하도록 수정합니다.

```go
if cfg.DryRun {
    slog.Info("Dry run mode enabled. Verifying connectivity and estimating row counts.")
    for _, table := range cfg.Tables {
        var count int
        err := dbConn.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count)
        if err != nil {
            slog.Error("failed to get row count", "table", table, "error", err)
            if tracker != nil {
                tracker.Error(table, err)
            }
            continue
        }
        slog.Info("table estimation", "table", table, "estimated_rows", count)
        if tracker != nil {
            // DryRunResult는 ProgressTracker 인터페이스에 optional이므로
            // 타입 단언으로 호출
            if drt, ok := tracker.(DryRunTracker); ok {
                drt.DryRunResult(table, count, true)
            }
        }
    }
    slog.Info("Dry run completed successfully.")
    return nil
}
```

#### 3.3.1. DryRunTracker 인터페이스

**파일**: `internal/migration/migration.go` (새 인터페이스)

기존 `ProgressTracker` 인터페이스를 깨뜨리지 않기 위해 별도의 인터페이스를 정의하고 타입 단언으로 호출합니다:

```go
// DryRunTracker는 Dry-Run 모드에서 추가 결과를 전송할 수 있는 tracker입니다.
type DryRunTracker interface {
    DryRunResult(table string, totalRows int, connectionOk bool)
}
```

`WebSocketTracker`는 이 인터페이스를 자동으로 만족합니다 (`DryRunResult` 메서드 추가 후).

### 3.4. 로거 런타임 전환

**파일**: `internal/logger/logger.go`

```go
// SetJSONMode는 런타임에 로그 출력 형식을 전환합니다.
func SetJSONMode(enabled bool) {
    var handler slog.Handler
    if enabled {
        handler = slog.NewJSONHandler(os.Stdout, nil)
    } else {
        handler = slog.NewTextHandler(os.Stdout, nil)
    }
    slog.SetDefault(slog.New(handler))
}
```

### 3.5. 프론트엔드 UI 변경

**파일**: `internal/web/templates/index.html`

#### 3.5.1. 고급 설정 섹션 확장 (Advanced Settings)

기존 `<details>` 요소 내부(line 308-320)를 확장합니다:

```html
<details style="margin-bottom: 1rem; cursor: pointer;">
    <summary style="font-weight: 600; font-size: 0.9rem; color: var(--text-muted); margin-bottom: 0.5rem;">Advanced Settings</summary>
    <div style="margin-top: 0.75rem; padding: 1rem; background: var(--bg-color); border-radius: var(--radius-md);">
        <!-- 기존 필드 -->
        <div style="display: flex; gap: 1rem;">
            <div class="form-group" style="flex: 1; margin-bottom: 0;">
                <label for="batchSize">Batch Size</label>
                <input type="text" id="batchSize" value="1000">
            </div>
            <div class="form-group" style="flex: 1; margin-bottom: 0;">
                <label for="workers">Parallel Workers</label>
                <input type="text" id="workers" value="4">
            </div>
        </div>

        <!-- 구분선 -->
        <hr style="border: none; border-top: 1px solid var(--border-color); margin: 1rem 0;">

        <!-- v4 새 필드: 출력 파일명, PG 스키마 -->
        <div id="file-settings" style="display: flex; gap: 1rem;">
            <div class="form-group" style="flex: 1; margin-bottom: 0;">
                <label for="outFile">Output Filename</label>
                <input type="text" id="outFile" value="migration.sql">
            </div>
            <div class="form-group" style="flex: 1; margin-bottom: 0;">
                <label for="schema">PG Schema</label>
                <input type="text" id="schema" placeholder="public">
            </div>
        </div>

        <!-- v4 새 필드: 체크박스 -->
        <div style="margin-top: 0.75rem;">
            <div id="perTableContainer" class="checkbox-container" style="margin-bottom: 0.5rem;">
                <input type="checkbox" id="perTable" checked>
                <label for="perTable">Per-Table File Output</label>
            </div>
            <div class="checkbox-container" style="margin-bottom: 0;">
                <input type="checkbox" id="logJson">
                <label for="logJson">JSON Logging</label>
            </div>
        </div>
    </div>
</details>
```

#### 3.5.2. Dry-Run 토글 추가

마이그레이션 버튼 직전(line 339 부근)에 Dry-Run 체크박스를 추가합니다:

```html
<div class="checkbox-container" style="margin-bottom: 1rem; margin-top: 1rem;">
    <input type="checkbox" id="dryRun">
    <label for="dryRun">Dry-Run (Verify connectivity & estimate row counts only)</label>
</div>

<button id="btn-migrate" class="btn-success" style="margin-top: 0.5rem;">Start Migration</button>
```

#### 3.5.3. JavaScript 변경사항

##### a) 조건부 표시 로직

`directMigration` 체크박스 이벤트 리스너를 확장합니다:

```javascript
directMigration.addEventListener('change', (e) => {
    const isDirect = e.target.checked;
    pgConfig.style.display = isDirect ? 'block' : 'none';

    // SQL File 전용 컨트롤 토글
    document.getElementById('file-settings').style.display = isDirect ? 'none' : 'flex';
    document.getElementById('perTableContainer').style.display = isDirect ? 'none' : 'flex';
});
```

##### b) Dry-Run 토글 시 버튼 텍스트 변경

```javascript
const dryRunCheckbox = document.getElementById('dryRun');
dryRunCheckbox.addEventListener('change', (e) => {
    btnMigrate.innerText = e.target.checked ? 'Run Verification' : 'Start Migration';
});
```

##### c) 마이그레이션 요청 payload 확장

`btnMigrate` 클릭 이벤트 핸들러 내 `fetch` 호출 부분을 확장합니다:

```javascript
const outFile = document.getElementById('outFile').value || 'migration.sql';
const schema = document.getElementById('schema').value;
const perTable = document.getElementById('perTable').checked;
const dryRun = document.getElementById('dryRun').checked;
const logJson = document.getElementById('logJson').checked;

const res = await fetch('/api/migrate', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        oracleUrl,
        username,
        password,
        tables: selectedTables,
        direct: isDirect,
        pgUrl: pgUrl,
        withDdl: withDdl,
        batchSize: batchSize,
        workers: workers,
        outFile: outFile,
        perTable: perTable,
        schema: schema,
        dryRun: dryRun,
        logJson: logJson
    })
});
```

##### d) Dry-Run 결과 WebSocket 메시지 처리

`handleProgressMessage` 함수에 `dry_run_result` 타입 처리를 추가합니다:

```javascript
function handleProgressMessage(msg) {
    const container = document.getElementById('progress-container');

    if (msg.type === 'all_done') {
        currentZipId = msg.zip_file_id;
        if (currentZipId) {
            downloadBtn.style.display = 'flex';
        } else {
            downloadBtn.style.display = 'none';
        }
        btnMigrate.disabled = false;
        btnMigrate.innerText = dryRunCheckbox.checked ? 'Run Verification' : 'Start Migration';
        return;
    }

    // Dry-Run 결과 처리
    if (msg.type === 'dry_run_result') {
        let wrapper = document.getElementById(`prog-${msg.table}`);
        if (!wrapper) {
            wrapper = document.createElement('div');
            wrapper.id = `prog-${msg.table}`;
            wrapper.className = 'progress-container';
            container.appendChild(wrapper);
        }
        const statusIcon = msg.connection_ok
            ? '<svg style="width:14px;height:14px;vertical-align:middle;margin-right:4px;color:var(--success-color)" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>'
            : '<svg style="width:14px;height:14px;vertical-align:middle;margin-right:4px;color:var(--danger-color)" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>';
        wrapper.innerHTML = `
            <div class="progress-header">
                <span class="progress-title">${msg.table}</span>
                <span class="progress-text">
                    ${statusIcon}
                    Estimated rows: <strong>${(msg.total || 0).toLocaleString()}</strong>
                </span>
            </div>
        `;
        return;
    }

    // ... 기존 init/update/done/error 처리 유지
}
```

##### e) Dry-Run 모드 시 다운로드 버튼 숨김 및 진행 라벨 표시

Dry-Run 활성화 시 진행 상황 섹션 제목에 "(Dry-Run)" 접미사를 추가합니다:

```javascript
// btnMigrate 클릭 이벤트 핸들러 내
const progressTitle = document.querySelector('#progress-section h2');
if (dryRun) {
    progressTitle.innerText = '3. Verification Results (Dry-Run)';
} else {
    progressTitle.innerText = '3. Migration Progress';
}
```

## 4. 보안 고려사항

### 4.1. 입력값 검증
- `outFile`: 경로 구분자(`/`, `\`) 포함 시 거부하여 **Path Traversal** 방지
- `schema`: 정규식 `^[a-zA-Z_][a-zA-Z0-9_]*$`로 SQL 식별자만 허용하여 **SQL Injection** 방지
- 기존 `filepath.Base()` 기반 다운로드 경로 검증은 유지

### 4.2. 런타임 로그 모드 전환
- `SetJSONMode`는 글로벌 `slog.Default`를 변경하므로, 동시 요청 시 경합(race) 가능
- 현재 단일 사용자 로컬 환경이므로 허용 가능. 향후 멀티 유저 지원 시 per-request logger로 전환 필요

### 4.3. Dry-Run SQL 인젝션 방지
- Dry-Run의 `SELECT COUNT(*) FROM %s`에서 `tableName`은 `/api/tables` API가 Oracle `ALL_TABLES`에서 조회한 값만 사용되므로 안전
- 추가 방어가 필요하면 테이블명에 대해 `^[a-zA-Z_][a-zA-Z0-9_]*$` 검증 적용 가능

## 5. 테스트 계획

### 5.1. 단위 테스트

| 테스트 케이스 | 파일 | 설명 |
|--------------|------|------|
| `TestValidateMigrationRequest_ValidInput` | `server_test.go` | 정상 입력 통과 |
| `TestValidateMigrationRequest_PathTraversal` | `server_test.go` | `outFile`에 `/` 포함 시 에러 |
| `TestValidateMigrationRequest_InvalidSchema` | `server_test.go` | `schema`에 특수문자 포함 시 에러 |
| `TestDryRunResult_Broadcast` | `tracker_test.go` | `DryRunResult` 메서드가 올바른 JSON 전송 |
| `TestSetJSONMode` | `logger_test.go` | 로그 모드 전환 검증 |

### 5.2. 통합 테스트

| 시나리오 | 설명 |
|---------|------|
| SQL File + PerTable=false | 단일 파일 출력, `outFile` 이름 반영 확인 |
| SQL File + PerTable=true | 테이블별 파일 생성 확인 |
| SQL File + Schema 지정 | INSERT문에 스키마 접두사 포함 확인 |
| Dry-Run 모드 | 파일 생성 없이 row count만 반환 확인 |
| Dry-Run + WebSocket | `dry_run_result` 메시지 수신 확인 |
| 하위 호환성 | 새 필드 없는 요청이 기본값으로 정상 동작 확인 |

### 5.3. 프론트엔드 수동 테스트

| 시나리오 | 검증 항목 |
|---------|----------|
| Direct 모드 전환 | `outFile`, `perTable` 컨트롤 숨김 확인 |
| SQL File 모드 전환 | `outFile`, `perTable` 컨트롤 표시 확인 |
| Dry-Run 토글 | 버튼 텍스트 변경, 결과 표시 형식 확인 |
| Schema 입력 | placeholder `public` 표시, 빈값 허용 확인 |
| JSON Logging 토글 | 서버 로그 형식 변경 확인 (터미널에서) |


### task.md

# Implementation Tasks: Web UI CLI 파라미터 통합 (v4)

## Phase 1: 백엔드 API 구조체 확장 및 검증

### 1-1. `startMigrationRequest` 구조체 확장
- [x] `internal/web/server.go`의 `startMigrationRequest`에 5개 필드 추가
  - `OutFile string json:"outFile"`
  - `PerTable bool json:"perTable"`
  - `Schema string json:"schema"`
  - `DryRun bool json:"dryRun"`
  - `LogJSON bool json:"logJson"`
- [x] `import`에 `regexp`, `strings`, `fmt` 추가 (미포함 시)

### 1-2. 입력값 검증 함수 구현
- [x] `validateMigrationRequest()` 함수 신규 작성
  - `outFile`에 경로 구분자(`/`, `\`) 포함 시 에러 반환
  - `schema`에 SQL 식별자 패턴(`^[a-zA-Z_][a-zA-Z0-9_]*$`) 외 문자 포함 시 에러 반환
  - `batchSize`, `workers` 음수 방지
- [x] `startMigration` 핸들러에서 `ShouldBindJSON` 직후 `validateMigrationRequest` 호출 추가

### 1-3. `startMigration` 핸들러 Config 매핑 수정
- [x] `outFile` 기본값 처리: 빈 문자열이면 `"migration.sql"` 설정
- [x] `PerTable: true` 하드코딩 → `req.PerTable`로 변경
- [x] `cfg`에 `OutFile`, `Schema`, `DryRun` 필드 매핑 추가
- [x] DryRun 시 임시 디렉토리 생성 스킵 로직 추가 (`!req.Direct && !req.DryRun`)
- [x] DryRun 완료 시 ZIP 생성 없이 `tracker.AllDone("")` 호출
- [x] DryRun 시 임시 디렉토리 삭제 스킵 (`!req.Direct && !req.DryRun`)

## Phase 2: WebSocket 프로토콜 확장 (Dry-Run)

### 2-1. Tracker 메시지 타입 추가
- [x] `internal/web/ws/tracker.go`에 `MsgDryRunResult MsgType = "dry_run_result"` 상수 추가
- [x] `ProgressMsg` 구조체에 `ConnectionOk bool json:"connection_ok,omitempty"` 필드 추가

### 2-2. `DryRunResult` 메서드 구현
- [x] `WebSocketTracker`에 `DryRunResult(table string, totalRows int, connectionOk bool)` 메서드 추가
- [x] 해당 메서드에서 `MsgDryRunResult` 타입으로 broadcast

## Phase 3: 마이그레이션 로직 Dry-Run WebSocket 연동

### 3-1. `DryRunTracker` 인터페이스 정의
- [x] `internal/migration/migration.go`에 `DryRunTracker` 인터페이스 추가
  - `DryRunResult(table string, totalRows int, connectionOk bool)`

### 3-2. Dry-Run 로직에 tracker 연동
- [x] 기존 Dry-Run 블록(`migration.go:40-53`)에서 tracker nil 체크 후 타입 단언으로 `DryRunResult` 호출
- [x] 에러 발생 시 `tracker.Error(table, err)` 호출 추가

## Phase 4: 로거 런타임 전환

### 4-1. `SetJSONMode` 함수 구현
- [x] `internal/logger/logger.go`에 `SetJSONMode(enabled bool)` 함수 추가
  - `enabled=true`: `slog.NewJSONHandler` → `slog.SetDefault`
  - `enabled=false`: `slog.NewTextHandler` → `slog.SetDefault`

### 4-2. 핸들러에서 LogJSON 연동
- [x] `startMigration` 고루틴 내에서 `req.LogJSON` true 시 `logger.SetJSONMode(true)` 호출
- [x] `defer logger.SetJSONMode(false)`로 원복 처리

## Phase 5: 프론트엔드 UI 확장

### 5-1. 고급 설정(Advanced Settings) 섹션 확장
- [x] 기존 Batch Size / Workers 아래에 구분선(`<hr>`) 추가
- [x] Output Filename 텍스트 입력 필드 추가 (기본값: `migration.sql`)
- [x] PG Schema 텍스트 입력 필드 추가 (placeholder: `public`)
- [x] Per-Table File Output 체크박스 추가 (기본값: checked)
- [x] JSON Logging 체크박스 추가 (기본값: unchecked)

### 5-2. Dry-Run 토글 UI 추가
- [x] 마이그레이션 시작 버튼 직전에 Dry-Run 체크박스 배치
- [x] 라벨: "Dry-Run (Verify connectivity & estimate row counts only)"

### 5-3. 조건부 표시 JavaScript 로직
- [x] `directMigration` 이벤트 리스너 확장: Direct 모드 시 `outFile`, `perTable` 컨트롤 숨김
- [x] SQL File 모드 전환 시 해당 컨트롤 다시 표시

### 5-4. Dry-Run 토글 JavaScript 로직
- [x] `dryRun` 체크박스 change 이벤트: 버튼 텍스트 "Start Migration" ↔ "Run Verification" 전환

### 5-5. 마이그레이션 요청 payload 확장
- [x] `btnMigrate` 클릭 핸들러에서 5개 새 필드 값 수집
- [x] `fetch('/api/migrate')` body에 `outFile`, `perTable`, `schema`, `dryRun`, `logJson` 추가

### 5-6. Dry-Run WebSocket 메시지 처리
- [x] `handleProgressMessage`에 `dry_run_result` 타입 분기 추가
- [x] Dry-Run 결과 UI 렌더링: 테이블명 + 연결 상태 아이콘 + 예상 row count
- [x] `all_done` 시 Dry-Run 모드면 다운로드 버튼 숨김 유지
- [x] Dry-Run 모드 시 진행 섹션 제목을 "3. Verification Results (Dry-Run)"으로 변경

## Phase 6: 테스트

### 6-1. 단위 테스트 작성
- [x] `TestValidateMigrationRequest_ValidInput` - 정상 입력 통과
- [x] `TestValidateMigrationRequest_PathTraversal` - outFile 경로 구분자 거부
- [x] `TestValidateMigrationRequest_InvalidSchema` - schema 특수문자 거부
- [x] `TestDryRunResult_Broadcast` - DryRunResult WebSocket 메시지 검증
- [x] `TestSetJSONMode` - 로그 모드 전환 검증

### 6-2. 통합 테스트 작성
- [x] SQL File + PerTable=false: 단일 파일 출력 및 outFile 이름 반영
- [x] SQL File + PerTable=true: 테이블별 파일 생성
- [x] SQL File + Schema 지정: INSERT문 스키마 접두사 포함
- [x] Dry-Run 모드: 파일 미생성, row count만 반환
- [x] 하위 호환성: 새 필드 미포함 요청 정상 동작

### 6-3. 프론트엔드 수동 테스트
- [ ] Direct ↔ SQL File 모드 전환 시 컨트롤 표시/숨김 확인
- [ ] Dry-Run 토글 시 버튼 텍스트 및 결과 UI 확인
- [ ] Schema 입력 placeholder 및 빈값 허용 확인
- [ ] JSON Logging 토글 시 서버 로그 형식 변경 확인

## Phase 7: 동기화 및 마무리

### 7-1. 템플릿 동기화
- [x] `internal/web/templates/index.html` 변경사항을 `web/templates/index.html`에 동기화

### 7-2. 빌드 및 검증
- [x] `go build` 성공 확인
- [x] `go test ./...` 전체 테스트 통과 확인
- [x] `go vet ./...` 정적 분석 통과 확인

### 7-3. 문서 정리
- [ ] `docs/v4/prd.md` 최종 확인
- [ ] `docs/v4/spec.md` 최종 확인
- [x] `docs/v4/task.md` 완료 항목 체크


---
## <a name="v05"></a> v05

### prd.md

# PRD (Product Requirements Document) - Sequence / Index 마이그레이션 지원 (v5)

## 1. 개요 (Overview)

현재 `--with-ddl` 옵션은 Oracle 테이블의 컬럼 정의(`CREATE TABLE`)만 PostgreSQL로 변환합니다.
실제 운영 DB 이전에는 **Sequence**(자동 증가 값 생성기)와 **Index**(검색 성능 구조체)도 함께 옮겨야 합니다.
이 버전에서는 `--with-sequences`, `--with-indexes` 두 옵션을 추가하여 테이블 DDL과 함께
관련 Sequence·Index를 Oracle에서 읽어 PostgreSQL 호환 DDL로 변환·출력합니다.

---

## 2. 배경 (Background)

### 2.1. 현재 한계

| 객체 | `--with-ddl` 지원 | 비고 |
|------|:-----------------:|------|
| 테이블 컬럼 정의 | O | `CREATE TABLE IF NOT EXISTS` 생성 |
| Sequence | X | Oracle `CREATE SEQUENCE` → PG 미변환 |
| 일반 Index | X | Oracle `CREATE INDEX` → PG 미변환 |
| Unique Index | X | Oracle `UNIQUE` → PG 미변환 |
| Primary Key | X | Oracle `PRIMARY KEY` 제약 → PG 미변환 |
| Foreign Key | △ | 범위 외 (v5 미포함) |

### 2.2. Oracle ↔ PostgreSQL 객체 대응

| Oracle | PostgreSQL |
|--------|-----------|
| `CREATE SEQUENCE s START WITH n INCREMENT BY m MINVALUE a MAXVALUE b CYCLE/NOCYCLE` | `CREATE SEQUENCE IF NOT EXISTS s START n INCREMENT m MINVALUE a MAXVALUE b [CYCLE]` |
| `CREATE INDEX i ON t (col1, col2)` | `CREATE INDEX IF NOT EXISTS i ON t (col1, col2)` |
| `CREATE UNIQUE INDEX i ON t (col)` | `CREATE UNIQUE INDEX IF NOT EXISTS i ON t (col)` |
| `PRIMARY KEY (col)` 제약 (테이블 단위) | `ALTER TABLE t ADD PRIMARY KEY (col)` |

---

## 3. 목표 (Goals)

- Oracle DB의 지정된 테이블과 **연관된 Sequence·Index**를 자동으로 추출해 PostgreSQL DDL로 변환합니다.
- 기존 `--with-ddl` 플래그와 **독립적으로** 선택 가능하게 합니다 (조합 자유).
- SQL File 출력 모드와 Direct(PG 직접 실행) 모드 모두 지원합니다.
- Web UI에서도 동일하게 제어할 수 있도록 확장합니다.
- 하위 호환성을 유지합니다 (기존 옵션 미설정 시 동작 불변).

---

## 4. 기능 요구사항 (Functional Requirements)

### 4.1. Sequence 마이그레이션 (`--with-sequences`)

#### 4.1.1. Oracle 메타데이터 조회

`ALL_SEQUENCES` 뷰에서 지정 테이블과 **이름이 연관된** Sequence를 조회합니다.

연관 판별 기준 (우선순위 순):
1. 테이블의 컬럼 중 `DEFAULT` 값에 해당 Sequence의 `.NEXTVAL`이 포함된 경우 (`ALL_TAB_COLUMNS.DATA_DEFAULT`)
2. Sequence 이름이 `<TABLE_NAME>_SEQ`, `<TABLE_NAME>_ID_SEQ`, `SEQ_<TABLE_NAME>` 패턴인 경우
3. `--sequences` 플래그로 명시적으로 지정한 Sequence 이름

조회 쿼리 (안):
```sql
SELECT sequence_name, min_value, max_value, increment_by,
       cycle_flag, last_number
FROM   all_sequences
WHERE  sequence_owner = :owner
  AND  sequence_name IN (
         SELECT REGEXP_SUBSTR(data_default, '[A-Z0-9_$#]+', 1, 1)
         FROM   all_tab_columns
         WHERE  table_name = :table
           AND  data_default LIKE '%.NEXTVAL%'
         UNION ALL
         SELECT sequence_name
         FROM   all_sequences
         WHERE  sequence_name IN (
                  :table || '_SEQ',
                  :table || '_ID_SEQ',
                  'SEQ_' || :table
                )
       )
```

#### 4.1.2. PostgreSQL DDL 생성

```sql
CREATE SEQUENCE IF NOT EXISTS {schema.}seq_name
    START WITH {last_number}
    INCREMENT BY {increment_by}
    MINVALUE {min_value}
    MAXVALUE {max_value}
    [CYCLE | NO CYCLE];
```

- `last_number`는 Oracle의 현재 Sequence 값을 반영해 충돌 없이 이어받음
- `MAXVALUE`가 Oracle 기본값(`9999999999999999999999999999`) 이상이면 PostgreSQL 기본값으로 생략
- DDL은 `CREATE TABLE` **이전**에 출력되어 `DEFAULT nextval(...)` 컬럼 정의보다 앞에 위치

#### 4.1.3. 컬럼 기본값 연동

`--with-ddl`과 함께 사용할 경우, `DEFAULT` 값에 `.NEXTVAL`이 있는 컬럼은 PostgreSQL DDL에서:

```sql
column_name bigint DEFAULT nextval('schema.seq_name')
```

로 자동 변환합니다.

---

### 4.2. Index 마이그레이션 (`--with-indexes`)

#### 4.2.1. Oracle 메타데이터 조회

`ALL_INDEXES`, `ALL_IND_COLUMNS` 뷰에서 테이블의 Index 목록과 컬럼 정보를 조회합니다.

```sql
SELECT i.index_name,
       i.uniqueness,
       i.index_type,
       c.column_name,
       c.column_position,
       c.descend
FROM   all_indexes i
JOIN   all_ind_columns c
  ON   c.index_name  = i.index_name
 AND   c.table_owner = i.owner
WHERE  i.table_name  = :table
  AND  i.owner       = :owner
  AND  i.index_type IN ('NORMAL', 'FUNCTION-BASED NORMAL')
ORDER  BY i.index_name, c.column_position
```

제외 대상:
- `index_type = 'LOB'` (BLOB/CLOB 관리용 내부 인덱스)
- Oracle 내부 PK 인덱스(`index_name` 패턴 `SYS_C%`)는 `ALTER TABLE ADD PRIMARY KEY`로 대체

#### 4.2.2. PostgreSQL DDL 생성

```sql
-- 일반 인덱스
CREATE INDEX IF NOT EXISTS {index_name} ON {schema.}{table_name} ({col1} [DESC], {col2});

-- Unique 인덱스
CREATE UNIQUE INDEX IF NOT EXISTS {index_name} ON {schema.}{table_name} ({col});

-- Primary Key (SYS_C% 계열)
ALTER TABLE {schema.}{table_name} ADD PRIMARY KEY ({pk_col});
```

- `DESCEND = 'DESC'`인 컬럼은 `col DESC` 으로 표현
- Function-based index (`FUNCTION-BASED NORMAL`)의 경우 표현식 그대로 사용
- DDL은 `INSERT` 문 **이후** (또는 `CREATE TABLE` 바로 다음) 에 출력

#### 4.2.3. 출력 순서

```
-- Sequence DDL (--with-sequences)
CREATE SEQUENCE IF NOT EXISTS ...;

-- Table DDL (--with-ddl)
CREATE TABLE IF NOT EXISTS ...;

-- Index DDL (--with-indexes)
CREATE INDEX IF NOT EXISTS ...;
ALTER TABLE ... ADD PRIMARY KEY (...);

-- Data INSERT
INSERT INTO ... VALUES ...;
```

---

### 4.3. CLI 플래그

| 플래그 | 타입 | 기본값 | 설명 |
|--------|------|--------|------|
| `--with-sequences` | bool | `false` | 연관 Sequence DDL 포함 |
| `--with-indexes` | bool | `false` | 연관 Index DDL 포함 |
| `--sequences` | string | `""` | 추가로 포함할 Sequence 이름 목록 (쉼표 구분) |
| `--oracle-owner` | string | `""` | Oracle 스키마(소유자) 이름. 미지정 시 `-user` 값 사용 |

#### 조합 예시

```bash
# 테이블 DDL + Sequence + Index 모두 포함하여 SQL 파일로 출력
dbmigrator -url localhost:1521/ORCL -user scott -password tiger \
  -tables USERS,ORDERS \
  -with-ddl -with-sequences -with-indexes

# 직접 마이그레이션 + Index만
dbmigrator -url localhost:1521/ORCL -user scott -password tiger \
  -tables USERS \
  -pg-url postgres://pguser:pgpass@localhost:5432/mydb \
  -with-ddl -with-indexes -schema myschema

# Sequence 이름 명시 지정
dbmigrator ... -with-sequences -sequences SEQ_USERS,SEQ_ORDERS
```

---

### 4.4. Web UI 확장

#### 4.4.1. 고급 설정 섹션에 추가

기존 `--with-ddl` 체크박스 아래에 다음을 추가합니다:

```
[고급 설정] ▾
┌──────────────────────────────────────────────────────┐
│  ...                                                 │
│  ☑ DDL 생성 (CREATE TABLE)         ← 기존            │
│  ☐ Sequence 포함 (--with-sequences)  ← NEW          │
│  ☐ Index 포함   (--with-indexes)     ← NEW          │
│  Oracle 소유자: [           ]         ← NEW          │
└──────────────────────────────────────────────────────┘
```

- `Sequence 포함`, `Index 포함` 체크박스는 `DDL 생성`이 체크된 경우에만 활성화(enable) 처리
- `Oracle 소유자` 입력은 비어있으면 연결 시 사용한 Username으로 서버 측 대체

#### 4.4.2. API 요청 필드 추가

```json
{
  "withSequences": true,
  "withIndexes": true,
  "oracleOwner": "SCOTT"
}
```

#### 4.4.3. WebSocket 진행 메시지 확장

Sequence/Index DDL 실행 중 진행 상황을 알리는 새 메시지 타입:

```json
{ "type": "ddl_progress", "object": "sequence", "name": "SEQ_USERS", "status": "ok" }
{ "type": "ddl_progress", "object": "index",    "name": "IDX_USERS_EMAIL", "status": "ok" }
{ "type": "ddl_progress", "object": "index",    "name": "IDX_ORDERS_DATE", "status": "error", "error": "..." }
```

---

## 5. 비기능 요구사항 (Non-Functional Requirements)

- **멱등성**: `IF NOT EXISTS` 사용으로 재실행 시 오류 없이 스킵
- **오류 격리**: 특정 Sequence/Index 변환 실패 시 해당 객체만 경고 로그 후 계속 진행 (전체 중단 없음)
- **권한 최소화**: `ALL_SEQUENCES`, `ALL_INDEXES`, `ALL_IND_COLUMNS`, `ALL_TAB_COLUMNS` 조회 권한만 필요
- **성능**: 메타데이터 조회는 마이그레이션 시작 전 1회만 수행, 데이터 이전 성능에 영향 없음
- **하위 호환성**: `--with-sequences`, `--with-indexes` 미지정 시 기존 동작 완전 유지

---

## 6. 영향 범위 (Scope of Changes)

| 파일 | 변경 내용 |
|------|----------|
| `internal/config/config.go` | `WithSequences`, `WithIndexes`, `Sequences`, `OracleOwner` 필드 및 플래그 추가 |
| `internal/migration/ddl.go` | `GetSequenceMetadata`, `GenerateSequenceDDL`, `GetIndexMetadata`, `GenerateIndexDDL` 함수 추가 |
| `internal/migration/migration.go` | `MigrateTable` 내에서 Sequence/Index DDL 출력 로직 추가 |
| `internal/web/server.go` | `startMigrationRequest`에 `WithSequences`, `WithIndexes`, `OracleOwner` 필드 추가 |
| `internal/web/ws/tracker.go` | `MsgDDLProgress` 메시지 타입 및 `DDLProgress()` 메서드 추가 |
| `internal/web/templates/index.html` | 고급 설정 체크박스 2개 + Oracle 소유자 입력 추가 |

---

## 7. 마일스톤 (Milestones)

1. **PRD 확정**: `docs/v5/prd.md` 작성 ✅
2. **Config 확장**: 4개 필드 및 플래그 추가
3. **DDL 로직 구현**: `ddl.go`에 Sequence/Index 메타조회·변환 함수 구현
4. **마이그레이션 연동**: `migration.go` 출력 순서 통합
5. **WebSocket 확장**: `ddl_progress` 메시지 타입 추가
6. **Web UI 확장**: 체크박스·입력 필드 추가, 조건부 활성화 로직
7. **테스트**: 단위 테스트(메타조회 mock, DDL 생성 검증) + 통합 테스트
8. **최종 리뷰 및 문서 정리**

---

## 8. 향후 확장 고려사항 (Future Considerations)

- **Foreign Key** 마이그레이션 (`--with-fk`): 참조 무결성 제약 변환
- **Trigger** 마이그레이션 (`--with-triggers`): Oracle PL/SQL → PostgreSQL PL/pgSQL 변환 (복잡도 높음)
- **View** 마이그레이션 (`--with-views`): 연관 뷰 DDL 변환
- **파티션** 지원: Oracle Range/List 파티션 → PostgreSQL 파티션 테이블 변환


### task.md

# Implementation Tasks: Sequence / Index 마이그레이션 지원 (v5)

## Phase 1: Config 확장

### 1-1. `Config` 구조체 필드 추가
- [x] `internal/config/config.go`의 `Config`에 4개 필드 추가
  - `WithSequences bool`
  - `WithIndexes bool`
  - `Sequences string` (쉼표 구분 명시 지정용)
  - `OracleOwner string`

### 1-2. CLI 플래그 등록
- [x] `--with-sequences` bool 플래그 추가 (`"연관 Sequence DDL 포함"`)
- [x] `--with-indexes` bool 플래그 추가 (`"연관 Index DDL 포함"`)
- [x] `--sequences` string 플래그 추가 (`"추가 포함할 Sequence 이름 목록 (쉼표 구분)"`)
- [x] `--oracle-owner` string 플래그 추가 (`"Oracle 스키마 소유자 (미지정 시 -user 값 사용)"`)

### 1-3. 사용 예시 업데이트
- [x] `flag.Usage` 예시 블록에 `--with-sequences`, `--with-indexes` 조합 예시 추가

---

## Phase 2: DDL 로직 구현 (`ddl.go`)

### 2-1. Sequence 메타데이터 조회
- [x] `SequenceMetadata` 구조체 정의
  - `Name string`
  - `MinValue int64`
  - `MaxValue string` (Oracle 최대값은 28자리 → string)
  - `IncrementBy int64`
  - `CycleFlag string` (`"Y"` / `"N"`)
  - `LastNumber int64`
- [x] `GetSequenceMetadata(db *sql.DB, tableName, owner string, extraNames []string) ([]SequenceMetadata, error)` 구현
  - `ALL_TAB_COLUMNS.DATA_DEFAULT`에서 `.NEXTVAL` 포함 컬럼으로 연관 Sequence 이름 추출
  - 이름 패턴(`<TABLE>_SEQ`, `<TABLE>_ID_SEQ`, `SEQ_<TABLE>`) 조회
  - `extraNames`(명시 지정 목록) 병합, 중복 제거
  - `ALL_SEQUENCES` 에서 메타데이터 조회

### 2-2. Sequence DDL 생성
- [x] `GenerateSequenceDDL(seq SequenceMetadata, schema string) string` 구현
  - `CREATE SEQUENCE IF NOT EXISTS {schema.}name` 출력
  - `START WITH last_number`
  - `INCREMENT BY increment_by`
  - `MINVALUE min_value`
  - Oracle 기본 MAXVALUE(28자리 9) 이상이면 `MAXVALUE` 절 생략
  - `CYCLE` / `NO CYCLE` 조건 처리

### 2-3. Index 메타데이터 조회
- [x] `IndexMetadata` 구조체 정의
  - `Name string`
  - `Uniqueness string` (`"UNIQUE"` / `"NONUNIQUE"`)
  - `IndexType string` (`"NORMAL"` / `"FUNCTION-BASED NORMAL"`)
  - `IsPK bool` (index_name이 `SYS_C%` 패턴)
  - `Columns []IndexColumn`
- [x] `IndexColumn` 구조체 정의
  - `Name string`
  - `Position int`
  - `Descend string` (`"ASC"` / `"DESC"`)
- [x] `GetIndexMetadata(db *sql.DB, tableName, owner string) ([]IndexMetadata, error)` 구현
  - `ALL_INDEXES` + `ALL_IND_COLUMNS` JOIN 조회
  - `index_type IN ('NORMAL', 'FUNCTION-BASED NORMAL')` 필터
  - `LOB` 인덱스 제외
  - `SYS_C%` 패턴 → `IsPK = true` 표시

### 2-4. Index DDL 생성
- [x] `GenerateIndexDDL(idx IndexMetadata, tableName, schema string) string` 구현
  - `IsPK=true`: `ALTER TABLE {schema.}table ADD PRIMARY KEY (col)` 출력
  - `Uniqueness="UNIQUE"`: `CREATE UNIQUE INDEX IF NOT EXISTS` 출력
  - 일반: `CREATE INDEX IF NOT EXISTS` 출력
  - `Descend="DESC"` 컬럼은 `col DESC` 로 표현

---

## Phase 3: 마이그레이션 연동 (`migration.go`)

### 3-1. `MigrateTableToFile` 확장
- [x] `cfg.WithSequences` true 시 `GetSequenceMetadata` 호출 후 `GenerateSequenceDDL` 결과를 `CREATE TABLE` **이전**에 출력
- [x] `cfg.WithIndexes` true 시 `GetIndexMetadata` 호출 후 `GenerateIndexDDL` 결과를 `CREATE TABLE` **이후**, INSERT **이전**에 출력
- [x] Sequence/Index 조회 실패 시 경고 로그 후 해당 객체 스킵 (전체 중단 없음)

### 3-2. `MigrateTableDirect` 확장
- [x] `cfg.WithSequences` true 시 Sequence DDL을 `pgPool.Exec`으로 실행
- [x] `cfg.WithIndexes` true 시 Index DDL을 `pgPool.Exec`으로 실행 (COPY 완료 후)
- [x] 각 DDL 실행 후 tracker가 있으면 `DDLProgress` 메시지 전송

### 3-3. OracleOwner 기본값 처리
- [x] `cfg.OracleOwner`가 빈 문자열이면 `cfg.User`를 대문자 변환하여 사용

---

## Phase 4: WebSocket 프로토콜 확장

### 4-1. 메시지 타입 및 구조체 추가
- [x] `internal/web/ws/tracker.go`에 `MsgDDLProgress MsgType = "ddl_progress"` 상수 추가
- [x] `ProgressMsg`에 다음 필드 추가
  - `Object string json:"object,omitempty"` (`"sequence"` / `"index"`)
  - `ObjectName string json:"object_name,omitempty"`
  - `Status string json:"status,omitempty"` (`"ok"` / `"error"`)

### 4-2. `DDLProgress` 메서드 구현
- [x] `WebSocketTracker`에 `DDLProgress(object, name, status string, err error)` 메서드 추가
- [x] `status="error"` 시 `ErrorMsg` 필드에 에러 내용 포함

---

## Phase 5: Web API 확장 (`server.go`)

### 5-1. `startMigrationRequest` 구조체 확장
- [x] `WithSequences bool json:"withSequences"` 필드 추가
- [x] `WithIndexes bool json:"withIndexes"` 필드 추가
- [x] `OracleOwner string json:"oracleOwner"` 필드 추가

### 5-2. Config 매핑
- [x] `startMigration` 핸들러에서 `cfg.WithSequences`, `cfg.WithIndexes`, `cfg.OracleOwner` 매핑 추가

---

## Phase 6: Web UI 확장 (`index.html`)

### 6-1. 고급 설정 섹션 체크박스 추가
- [x] `--with-ddl` 체크박스 아래에 구분 없이 연속 배치
  - `Sequence 포함 (--with-sequences)` 체크박스 추가 (id: `withSequences`, 기본값: unchecked)
  - `Index 포함 (--with-indexes)` 체크박스 추가 (id: `withIndexes`, 기본값: unchecked)
- [x] `Oracle 소유자` 텍스트 입력 추가 (id: `oracleOwner`, placeholder: `접속 계정과 동일`)

### 6-2. 조건부 활성화 JavaScript 로직
- [x] `withDdl` 체크박스 change 이벤트: unchecked 시 `withSequences`, `withIndexes` 체크 해제 및 비활성화
- [x] `withDdl` checked 시 두 체크박스 활성화

### 6-3. 요청 payload 확장
- [x] `btnMigrate` 클릭 핸들러에서 `withSequences`, `withIndexes`, `oracleOwner` 값 수집 및 전송

### 6-4. DDL Progress WebSocket 메시지 처리
- [x] `handleProgressMessage`에 `ddl_progress` 타입 분기 추가
- [x] 진행 목록에 `[sequence] SEQ_USERS ✓` / `[index] IDX_USERS_EMAIL ✓` 형태로 표시
- [x] `status="error"` 시 경고 아이콘 + 에러 메시지 표시

---

## Phase 7: 테스트

### 7-1. 단위 테스트 (`ddl_test.go` 확장)
- [x] `TestGenerateSequenceDDL_Basic` - 기본 Sequence DDL 생성 검증
- [x] `TestGenerateSequenceDDL_MaxValueOmit` - Oracle 기본 MAXVALUE 생략 검증
- [x] `TestGenerateSequenceDDL_Cycle` - CYCLE 옵션 반영 검증
- [x] `TestGenerateSequenceDDL_WithSchema` - 스키마 접두사 포함 검증
- [x] `TestGenerateIndexDDL_Normal` - 일반 Index DDL 검증
- [x] `TestGenerateIndexDDL_Unique` - Unique Index DDL 검증
- [x] `TestGenerateIndexDDL_PrimaryKey` - PK → ALTER TABLE 변환 검증
- [x] `TestGenerateIndexDDL_Descend` - DESC 컬럼 표현 검증
- [x] `TestDDLProgress_Broadcast` - DDLProgress WebSocket 메시지 검증

### 7-2. 통합 테스트 (`v5_integration_test.go` 신규)
- [x] `WithSequences=true`: Sequence DDL이 CREATE TABLE 이전에 출력되는지 검증
- [x] `WithIndexes=true`: Index DDL이 CREATE TABLE 이후에 출력되는지 검증
- [x] `WithSequences=false, WithIndexes=false`: 기존 동작 완전 유지 검증
- [x] OracleOwner 기본값: 빈 문자열 시 User 값 대문자로 대체 검증

### 7-3. 프론트엔드 수동 테스트
- [x] `withDdl` 미체크 시 `withSequences`, `withIndexes` 비활성화 확인
- [x] `ddl_progress` 메시지 UI 렌더링 확인

---

## Phase 8: 동기화 및 마무리

### 8-1. 템플릿 동기화
- [x] `internal/web/templates/index.html` 변경사항을 `web/templates/index.html`에 동기화

### 8-2. 빌드 및 검증
- [x] `go build` 성공 확인
- [x] `go test ./...` 전체 테스트 통과 확인
- [x] `go vet ./...` 정적 분석 통과 확인

### 8-3. 문서 정리
- [x] `docs/v5/prd.md` 최종 확인
- [x] `docs/v5/task.md` 완료 항목 체크


---
## <a name="v06"></a> v06

### prd.md

# PRD (Product Requirements Document) - 출력 대상 DB 선택 지원 (v6)

## 1. 개요 (Overview)

현재 `dbmigrator`는 Oracle → **PostgreSQL** 고정 경로만 지원합니다.
이 버전에서는 `--target-db` 옵션을 추가하여 출력 대상 DB를 **PostgreSQL·MySQL·MariaDB·SQLite·MSSQL** 중 선택할 수 있도록 합니다.
SQL 파일 출력 모드와 Direct 마이그레이션 모드 모두 선택된 DB의 방언(dialect)에 맞게 DDL·DML을 생성합니다.

---

## 2. 배경 (Background)

### 2.1. 현재 한계

| 출력 대상 | 현재 지원 | 비고 |
|-----------|:---------:|------|
| PostgreSQL | O | 기본 출력 대상 |
| MySQL / MariaDB | X | 문법·타입 차이로 그대로 사용 불가 |
| SQLite | X | 로컬 테스트 및 임베디드 용도 |
| MSSQL (SQL Server) | X | 기업 환경 이전 수요 |

### 2.2. 방언별 주요 차이

| 항목 | PostgreSQL | MySQL/MariaDB | SQLite | MSSQL |
|------|-----------|---------------|--------|-------|
| 자동 증가 | `SERIAL` / `BIGSERIAL` / `nextval()` | `AUTO_INCREMENT` | `AUTOINCREMENT` | `IDENTITY(1,1)` |
| 문자열 타입 | `TEXT`, `VARCHAR(n)` | `TEXT`, `VARCHAR(n)` | `TEXT` | `NVARCHAR(n)`, `NTEXT` |
| 날짜/시간 | `TIMESTAMP` | `DATETIME` | `TEXT` (ISO 8601) | `DATETIME2` |
| CLOB/BLOB | `TEXT`, `BYTEA` | `LONGTEXT`, `LONGBLOB` | `TEXT`, `BLOB` | `NVARCHAR(MAX)`, `VARBINARY(MAX)` |
| FLOAT | `NUMERIC(p,s)` | `DECIMAL(p,s)` | `REAL` | `DECIMAL(p,s)` |
| DDL 가드 | `IF NOT EXISTS` | `IF NOT EXISTS` | `IF NOT EXISTS` | 별도 존재 체크 필요 |
| 식별자 인용 | `"name"` | `` `name` `` | `"name"` | `[name]` |
| INSERT 다중 행 | `VALUES (),()`  | `VALUES (),()`  | `VALUES (),()` | `VALUES (),()` (2008+) |
| Sequence | 네이티브 지원 | 미지원 (`AUTO_INCREMENT`로 대체) | 미지원 | `SEQUENCE` (2012+) |
| Index 문법 | `CREATE INDEX IF NOT EXISTS` | `CREATE INDEX` (IF NOT EXISTS MySQL 8+) | `CREATE INDEX IF NOT EXISTS` | `CREATE INDEX` |

---

## 3. 목표 (Goals)

- `--target-db` 하나의 플래그로 출력 방언 전환이 가능하도록 합니다.
- Oracle 타입 → 대상 DB 타입 매핑 테이블을 방언별로 분리·관리합니다.
- DDL(CREATE TABLE, Sequence, Index), DML(INSERT) 모두 대상 DB에 맞는 문법으로 생성합니다.
- Direct 마이그레이션 모드에서 대상 DB에 맞는 드라이버로 연결합니다.
- Web UI의 연결 설정이 선택된 대상 DB에 맞게 동적으로 변경됩니다.
- PostgreSQL 기본값 유지 — 기존 사용자에게 아무 변화 없음.

---

## 4. 기능 요구사항 (Functional Requirements)

### 4.1. `--target-db` 플래그

| 값 | 설명 | Direct 드라이버 | 기본 포트 |
|----|------|----------------|---------|
| `postgres` | PostgreSQL (기본값) | `pgx/v5` | 5432 |
| `mysql` | MySQL 8.x | `go-sql-driver/mysql` | 3306 |
| `mariadb` | MariaDB 10.x+ | `go-sql-driver/mysql` | 3306 |
| `sqlite` | SQLite3 (파일 경로) | `mattn/go-sqlite3` | N/A |
| `mssql` | SQL Server 2019+ | `microsoft/go-mssqldb` | 1433 |

- 미지정 시 `postgres`로 동작 (완전 하위 호환)
- SQL 파일 출력 모드에서도 방언에 맞는 DDL·DML을 생성

### 4.2. Oracle 타입 매핑

방언별 타입 변환 테이블. `TypeMapper` 인터페이스를 구현하는 방언별 struct로 분리합니다.

| Oracle 타입 | PostgreSQL | MySQL/MariaDB | SQLite | MSSQL |
|------------|-----------|---------------|--------|-------|
| `NUMBER(p,0)` p≤4 | `SMALLINT` | `SMALLINT` | `INTEGER` | `SMALLINT` |
| `NUMBER(p,0)` p≤9 | `INTEGER` | `INT` | `INTEGER` | `INT` |
| `NUMBER(p,0)` p≤18 | `BIGINT` | `BIGINT` | `INTEGER` | `BIGINT` |
| `NUMBER(p,s)` s>0 | `NUMERIC(p,s)` | `DECIMAL(p,s)` | `REAL` | `DECIMAL(p,s)` |
| `VARCHAR2(n)` | `VARCHAR(n)` | `VARCHAR(n)` | `TEXT` | `NVARCHAR(n)` |
| `CHAR(n)` | `CHAR(n)` | `CHAR(n)` | `TEXT` | `NCHAR(n)` |
| `CLOB` | `TEXT` | `LONGTEXT` | `TEXT` | `NVARCHAR(MAX)` |
| `BLOB` | `BYTEA` | `LONGBLOB` | `BLOB` | `VARBINARY(MAX)` |
| `DATE` | `TIMESTAMP` | `DATETIME` | `TEXT` | `DATETIME2` |
| `TIMESTAMP(n)` | `TIMESTAMP(n)` | `DATETIME(n)` | `TEXT` | `DATETIME2(n)` |
| `FLOAT` | `DOUBLE PRECISION` | `DOUBLE` | `REAL` | `FLOAT` |

### 4.3. DDL 방언 차이 처리

#### 4.3.1. CREATE TABLE

```sql
-- PostgreSQL (기존)
CREATE TABLE IF NOT EXISTS "schema"."USERS" ( ... );

-- MySQL/MariaDB
CREATE TABLE IF NOT EXISTS `schema`.`USERS` ( ... );

-- SQLite
CREATE TABLE IF NOT EXISTS "USERS" ( ... );  -- 스키마 미지원

-- MSSQL
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'USERS')
CREATE TABLE [schema].[USERS] ( ... );
```

#### 4.3.2. Sequence DDL (`--with-sequences`)

| 대상 DB | 처리 방식 |
|---------|---------|
| PostgreSQL | `CREATE SEQUENCE IF NOT EXISTS ...` (기존) |
| MSSQL (2012+) | `CREATE SEQUENCE ... START WITH ... INCREMENT BY ...` |
| MySQL/MariaDB | Sequence 미지원 → 해당 컬럼을 `AUTO_INCREMENT`로 변환, 경고 로그 출력 |
| SQLite | Sequence 미지원 → `AUTOINCREMENT` 키워드로 대체, 경고 로그 출력 |

#### 4.3.3. Index DDL (`--with-indexes`)

```sql
-- PostgreSQL (기존)
CREATE INDEX IF NOT EXISTS idx_name ON "table" (col);

-- MySQL 5.7 이하 / MariaDB 10.4 이하
CREATE INDEX idx_name ON `table` (col);

-- MySQL 8.0+ / MariaDB 10.5+
CREATE INDEX IF NOT EXISTS idx_name ON `table` (col);

-- MSSQL
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_name')
    CREATE INDEX idx_name ON [schema].[table] (col);
```

#### 4.3.4. INSERT 문

- 기존 배치 INSERT(`VALUES (), ()`) 방식은 PostgreSQL·MySQL·SQLite·MSSQL(2008+) 모두 지원
- MSSQL: 최대 1000행 제한 → BatchSize 초과 시 자동으로 분할 출력

### 4.4. Direct 마이그레이션 연결

`--target-db` 값에 따라 대상 연결 URL 형식이 달라집니다.

| 대상 | URL 형식 | 플래그 |
|------|---------|--------|
| PostgreSQL | `postgres://user:pass@host:port/db` | `--pg-url` (기존 유지) |
| MySQL/MariaDB | `user:pass@tcp(host:port)/db` | `--target-url` (신규) |
| SQLite | `/path/to/file.db` | `--target-url` (신규) |
| MSSQL | `sqlserver://user:pass@host:port?database=db` | `--target-url` (신규) |

- 기존 `--pg-url`은 `--target-db postgres` 시 그대로 사용 가능 (하위 호환)
- `--target-db`가 postgres 이외이면서 `--pg-url`이 지정된 경우 경고 로그 출력

### 4.5. CLI 플래그 요약

| 플래그 | 타입 | 기본값 | 설명 |
|--------|------|--------|------|
| `--target-db` | string | `postgres` | 출력 대상 DB 종류 (`postgres`/`mysql`/`mariadb`/`sqlite`/`mssql`) |
| `--target-url` | string | `""` | 대상 DB 연결 URL (PostgreSQL 외 Direct 마이그레이션 시) |

#### CLI 조합 예시

```bash
# MySQL로 SQL 파일 출력 (DDL 포함)
dbmigrator -url localhost:1521/ORCL -user scott -password tiger \
  -tables USERS,ORDERS -with-ddl \
  --target-db mysql

# MariaDB 직접 마이그레이션
dbmigrator -url localhost:1521/ORCL -user scott -password tiger \
  -tables USERS \
  --target-db mariadb \
  --target-url "scott:tiger@tcp(localhost:3306)/mydb"

# SQLite 파일 생성 (테스트 환경)
dbmigrator -url localhost:1521/ORCL -user scott -password tiger \
  -tables USERS -with-ddl \
  --target-db sqlite

# MSSQL 직접 마이그레이션
dbmigrator -url localhost:1521/ORCL -user scott -password tiger \
  -tables USERS -with-ddl -with-indexes \
  --target-db mssql \
  --target-url "sqlserver://sa:pass@localhost:1433?database=mydb"
```

### 4.6. Web UI 확장

#### 4.6.1. 대상 DB 선택 드롭다운

고급 설정 섹션 상단에 대상 DB 선택 드롭다운을 추가합니다.

```
[고급 설정] ▾
┌──────────────────────────────────────────────────────┐
│  출력 대상 DB: [PostgreSQL ▾]          ← NEW          │
│                                                      │
│  Direct 마이그레이션                                   │
│  ☐ Direct Migration                                  │
│  대상 URL: [postgres://... or mysql://...]  ← 동적 변경│
│                                                      │
│  ...                                                 │
└──────────────────────────────────────────────────────┘
```

#### 4.6.2. 동적 UI 변화

| 선택 | `target-url` placeholder | `schema` 입력 | Sequence 체크박스 |
|------|--------------------------|-------------|-----------------|
| PostgreSQL | `postgres://user:pass@host:5432/db` | 표시 | 활성 |
| MySQL/MariaDB | `user:pass@tcp(host:3306)/db` | 표시 | 비활성 + 경고 툴팁 |
| SQLite | `/path/to/file.db` | 숨김 | 비활성 + 경고 툴팁 |
| MSSQL | `sqlserver://user:pass@host:1433?database=db` | 표시 | 활성 (2012+) |

- MySQL/MariaDB/SQLite 선택 시 `--with-sequences`가 체크되어 있으면 체크 해제 및 경고 표시
- SQLite 선택 시 `schema` 입력 필드 숨김 (SQLite는 스키마 미지원)

#### 4.6.3. API 요청 필드 추가

```json
{
  "targetDb": "mysql",
  "targetUrl": "user:pass@tcp(localhost:3306)/mydb"
}
```

#### 4.6.4. WebSocket 경고 메시지

```json
{ "type": "warning", "message": "MySQL은 Sequence를 지원하지 않습니다. AUTO_INCREMENT로 대체됩니다." }
```

---

## 5. 아키텍처 설계

### 5.1. Dialect 인터페이스

```go
// internal/dialect/dialect.go
type Dialect interface {
    Name() string
    QuoteIdentifier(name string) string
    MapOracleType(oracleType string, precision, scale int) string
    CreateTableDDL(tableName, schema string, cols []ColumnDef) string
    CreateSequenceDDL(seq SequenceMetadata, schema string) (string, bool) // (ddl, supported)
    CreateIndexDDL(idx IndexMetadata, tableName, schema string) string
    InsertStatement(tableName, schema string, cols []string, rows [][]any, batchSize int) []string
    DriverName() string
    NormalizeURL(url string) string
}
```

### 5.2. 방언별 구현체

```
internal/dialect/
├── dialect.go          # Dialect 인터페이스 + 공통 유틸
├── postgres.go         # PostgresDialect (기존 로직 이관)
├── mysql.go            # MySQLDialect
├── mariadb.go          # MariaDBDialect (MySQLDialect 임베딩 + 미세 차이)
├── sqlite.go           # SQLiteDialect
└── mssql.go            # MSSQLDialect
```

### 5.3. 기존 코드 영향

| 파일 | 변경 내용 |
|------|----------|
| `internal/config/config.go` | `TargetDB`, `TargetURL` 필드 및 플래그 추가 |
| `internal/dialect/` | 신규 패키지 (Dialect 인터페이스 + 5개 구현체) |
| `internal/migration/ddl.go` | `GenerateSequenceDDL`, `GenerateIndexDDL` → Dialect 위임 |
| `internal/migration/migration.go` | Dialect 선택 후 주입, Direct 모드 연결 URL 분기 |
| `internal/web/server.go` | `targetDb`, `targetUrl` 필드 추가, Config 매핑 |
| `internal/web/templates/index.html` | 대상 DB 드롭다운, URL placeholder 동적 변경 로직 |

---

## 6. 비기능 요구사항 (Non-Functional Requirements)

- **하위 호환성**: `--target-db` 미지정 시 완전히 기존 동작과 동일
- **Dialect 확장성**: 새 DB 추가 시 `Dialect` 인터페이스 구현체만 추가하면 됨
- **경고 투명성**: 방언 제약(Sequence 미지원 등) 발생 시 사용자에게 명확한 경고 출력
- **테스트 용이성**: Dialect 인터페이스로 단위 테스트에서 mock 교체 가능

---

## 7. 마일스톤 (Milestones)

1. **PRD 확정**: `docs/v6/prd.md` 작성 ✅
2. **Dialect 인터페이스 설계**: `internal/dialect/dialect.go` 정의
3. **PostgresDialect 구현**: 기존 로직 이관 (동작 동일 보장)
4. **MySQLDialect / MariaDBDialect 구현**: 타입 매핑 + DDL 생성
5. **SQLiteDialect 구현**: 타입 매핑 + DDL 생성
6. **MSSQLDialect 구현**: 타입 매핑 + DDL 생성 + MSSQL 배치 제한 처리
7. **Config 확장**: `--target-db`, `--target-url` 플래그 추가
8. **migration.go 연동**: Dialect 주입 및 Direct 모드 드라이버 분기
9. **Web API/UI 확장**: 드롭다운 + placeholder 동적 변경
10. **테스트**: 방언별 DDL·DML 출력 단위 테스트 + 통합 테스트
11. **최종 리뷰 및 문서 정리**

---

## 8. 향후 확장 고려사항 (Future Considerations)

- **CockroachDB** 지원: PostgreSQL 호환 방언으로 비교적 적은 변경
- **TiDB** 지원: MySQL 호환 방언
- **Oracle → Oracle** 지원: 동일 Oracle 버전 간 테이블 복제
- **소스 DB 다양화**: MySQL → PostgreSQL 등 Oracle 이외 소스 지원 (별도 PRD)


### spec.md

# Technical Specification (Spec) - Multi-Target DB Migration (v6)

## 1. 개요 (Overview)

본 문서는 `dbmigrator` v6의 핵심 기능인 **출력 대상 DB(Target DB) 다변화**에 대한 기술적 명세를 정의합니다.
기존 PostgreSQL 단일 지원 구조에서 벗어나, MySQL, MariaDB, SQLite, MSSQL 등 다양한 데이터베이스 방언(Dialect)을 지원하기 위한 아키텍처, 인터페이스 설계, 타입 매핑, DDL/DML 생성 전략, CLI 플래그 및 Web UI 확장에 대한 세부 사항을 다룹니다.

## 2. 아키텍처 설계 (Architecture Design)

다양한 대상 데이터베이스를 지원하기 위해 방언을 추상화하는 `Dialect` 인터페이스를 도입합니다.
`dbmigrator`의 핵심 로직(`migration`, `ddl` 등)은 특정 DB에 종속되지 않고 `Dialect` 인터페이스에 의존하여 유연성을 확보합니다.

### 2.1. Dialect 인터페이스 (`internal/dialect/dialect.go`)

각 대상 DB별로 고유한 동작(식별자 인용, 타입 매핑, DDL 생성, INSERT 문법 차이 등)을 캡슐화합니다.

```go
package dialect

// Dialect defines the interface for different target database dialects.
type Dialect interface {
	// Name returns the dialect name (e.g., "postgres", "mysql").
	Name() string

	// QuoteIdentifier quotes an identifier (e.g., table or column name) according to the dialect.
	QuoteIdentifier(name string) string

	// MapOracleType maps an Oracle data type to the target database type.
	MapOracleType(oracleType string, precision, scale int) string

	// CreateTableDDL generates the CREATE TABLE DDL.
	CreateTableDDL(tableName, schema string, cols []ColumnDef) string

	// CreateSequenceDDL generates the CREATE SEQUENCE DDL.
	// Returns a boolean indicating whether the target DB supports sequences.
	CreateSequenceDDL(seq SequenceMetadata, schema string) (string, bool)

	// CreateIndexDDL generates the CREATE INDEX DDL.
	CreateIndexDDL(idx IndexMetadata, tableName, schema string) string

	// InsertStatement generates batch INSERT statements.
	InsertStatement(tableName, schema string, cols []string, rows [][]any, batchSize int) []string

	// DriverName returns the Go SQL driver name (e.g., "pgx", "mysql", "sqlite3", "sqlserver").
	DriverName() string

	// NormalizeURL standardizes the connection URL for the target driver.
	NormalizeURL(url string) string
}
```

### 2.2. 방언별 구현체 (Dialect Implementations)

- **PostgreSQL (`internal/dialect/postgres.go`)**: 기존 동작(v1~v5)을 유지 (`pgx/v5` 드라이버 사용)
- **MySQL (`internal/dialect/mysql.go`)**: MySQL 8.x 대상 (`go-sql-driver/mysql` 드라이버 사용)
- **MariaDB (`internal/dialect/mariadb.go`)**: MySQL 구현체를 확장하거나 임베딩하여 미세 조정 (`go-sql-driver/mysql` 드라이버 사용)
- **SQLite (`internal/dialect/sqlite.go`)**: SQLite3 대상. 로컬 파일 시스템을 사용하며 스키마 개념 무시 (`mattn/go-sqlite3` 드라이버 사용)
- **MSSQL (`internal/dialect/mssql.go`)**: SQL Server 2019+ 대상. 배치 제한 등 고려 (`microsoft/go-mssqldb` 드라이버 사용)

## 3. 기능 상세 명세 (Functional Specifications)

### 3.1. CLI 플래그 및 Config 구조체 확장

`internal/config/config.go`에 다중 DB 지원을 위한 플래그를 추가합니다.

- `--target-db` (string, default: `"postgres"`): 대상 DB 선택. 유효값: `"postgres"`, `"mysql"`, `"mariadb"`, `"sqlite"`, `"mssql"`
- `--target-url` (string, default: `""`): PostgreSQL 외 DB로 Direct 마이그레이션 시 사용하는 URL 형식.

#### 하위 호환성 (Backward Compatibility)
- 기존 `--pg-url`은 `--target-db postgres` 일 때 `--target-url`처럼 처리됩니다.
- `--target-db`를 명시하지 않으면 기본적으로 `"postgres"`로 동작하여 기존 사용자 환경에 영향을 주지 않습니다.

### 3.2. Oracle 타입 매핑 (Type Mapping)

각 방언 구현체(`MapOracleType`)는 아래 테이블을 기준으로 매핑을 수행합니다.

| Oracle 타입 | PostgreSQL | MySQL/MariaDB | SQLite | MSSQL |
|------------|-----------|---------------|--------|-------|
| `NUMBER(p,0)` p≤4 | `SMALLINT` | `SMALLINT` | `INTEGER` | `SMALLINT` |
| `NUMBER(p,0)` p≤9 | `INTEGER` | `INT` | `INTEGER` | `INT` |
| `NUMBER(p,0)` p≤18| `BIGINT` | `BIGINT` | `INTEGER` | `BIGINT` |
| `NUMBER(p,s)` s>0 | `NUMERIC(p,s)` | `DECIMAL(p,s)` | `REAL` | `DECIMAL(p,s)` |
| `VARCHAR2(n)` | `VARCHAR(n)` | `VARCHAR(n)` | `TEXT` | `NVARCHAR(n)` |
| `CHAR(n)` | `CHAR(n)` | `CHAR(n)` | `TEXT` | `NCHAR(n)` |
| `CLOB` | `TEXT` | `LONGTEXT` | `TEXT` | `NVARCHAR(MAX)` |
| `BLOB` | `BYTEA` | `LONGBLOB` | `BLOB` | `VARBINARY(MAX)` |
| `DATE` | `TIMESTAMP` | `DATETIME` | `TEXT` | `DATETIME2` |
| `TIMESTAMP(n)`| `TIMESTAMP(n)` | `DATETIME(n)` | `TEXT` | `DATETIME2(n)` |
| `FLOAT` | `DOUBLE PRECISION`| `DOUBLE` | `REAL` | `FLOAT` |

### 3.3. DDL (CREATE TABLE, Sequence, Index) 생성 차이 처리

#### 3.3.1. 테이블 생성 (`CreateTableDDL`)
- **PostgreSQL**: `CREATE TABLE IF NOT EXISTS "schema"."TABLE" (...)`
- **MySQL/MariaDB**: `CREATE TABLE IF NOT EXISTS \`schema\`.\`TABLE\` (...)` (백틱 사용)
- **SQLite**: `CREATE TABLE IF NOT EXISTS "TABLE" (...)` (스키마 무시)
- **MSSQL**:
  ```sql
  IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'TABLE')
  CREATE TABLE [schema].[TABLE] (...)
  ```

#### 3.3.2. Sequence 생성 (`CreateSequenceDDL`)
- **PostgreSQL**: 지원. (기존 v5 유지)
- **MSSQL (2012+)**: 지원. `CREATE SEQUENCE [schema].[SEQ_NAME] START WITH ... INCREMENT BY ...`
- **MySQL / MariaDB / SQLite**:
  - `CreateSequenceDDL`은 빈 문자열과 `false`(미지원)를 반환.
  - 마이그레이션 엔진에서 해당 컬럼을 `AUTO_INCREMENT` (MySQL) 또는 `AUTOINCREMENT` (SQLite)로 자동 대체 처리.
  - 사용자에게 Sequence 미지원에 대한 경고(Warning)를 출력.

#### 3.3.3. 인덱스 생성 (`CreateIndexDDL`)
- **PostgreSQL**: `CREATE INDEX IF NOT EXISTS idx_name ON "schema"."table" (col)`
- **MySQL/MariaDB**: MySQL 8.0+에서는 `IF NOT EXISTS`를 지원하므로 활용. (이전 버전 호환성 필요 시 분기 고려)
- **SQLite**: `CREATE INDEX IF NOT EXISTS "idx_name" ON "table" ("col")`
- **MSSQL**:
  ```sql
  IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_name')
  CREATE INDEX idx_name ON [schema].[table] (col)
  ```

### 3.4. DML (INSERT) 생성 및 배치 처리 차이

- 다중 행 INSERT(`VALUES (), ()`)는 모든 타겟 DB에서 기본적으로 지원됨.
- **MSSQL 배치 제한 (`InsertStatement` 특징)**:
  - MSSQL은 단일 `INSERT` 문에서 `VALUES` 절 뒤에 올 수 있는 최대 행의 수가 1000행으로 제한됨.
  - `MSSQLDialect` 구현체의 `InsertStatement` 메서드는 전달된 `rows` 배열의 길이가 1000을 초과할 경우(예: `-batch 2000`), 자동으로 1000행 단위로 청크(chunk)를 분할하여 여러 개의 `INSERT` 문 배열을 반환해야 함.

### 3.5. Web UI 및 API 변경사항

#### 3.5.1. UI/UX 요구사항
- "고급 설정" 섹션 상단에 "출력 대상 DB" (Target DB) 드롭다운 (PostgreSQL, MySQL, MariaDB, SQLite, MSSQL) 추가.
- 선택된 대상에 따라 대상 URL 입력창의 PlaceHolder가 동적으로 변경됨.
  - PG: `postgres://user:pass@host:5432/db`
  - MySQL: `user:pass@tcp(host:3306)/db`
  - SQLite: `/path/to/file.db`
  - MSSQL: `sqlserver://user:pass@host:1433?database=db`
- SQLite 선택 시, `schema` 입력창 비활성화 또는 숨김 (스키마 미지원).
- MySQL/SQLite 선택 시, `--with-sequences`가 체크되어 있다면 체크 해제 후 비활성화, 툴팁으로 미지원 사유 안내.

#### 3.5.2. API 확장
마이그레이션 시작 요청 페이로드(`internal/web/server.go`)에 필드 추가:
```json
{
  ...
  "targetDb": "mysql",
  "targetUrl": "user:pass@tcp(localhost:3306)/mydb",
  ...
}
```

#### 3.5.3. WebSocket 메시지 확장
방언 제약사항에 따른 경고를 프론트엔드로 전달하기 위해 `warning` 타입의 메시지를 추가 활용.

## 4. 모듈 간 의존성 변경 요약 (Scope of Changes)

1. **`internal/dialect/*`**: 신규 패키지 및 5개의 방언 구현체 추가.
2. **`internal/config/config.go`**: `--target-db`, `--target-url` 처리.
3. **`internal/migration/ddl.go`**: DDL 템플릿 로직이 `Dialect` 인터페이스의 메서드 호출로 변경.
4. **`internal/migration/migration.go`**: Direct 모드에서 `Dialect` 드라이버 정보를 바탕으로 `sql.Open` 인자 동적 결정. `Insert` 구문 생성 시 `Dialect.InsertStatement` 위임.
5. **`internal/web/*`**: `targetDb`, `targetUrl` 파라미터 핸들링 추가 및 Web UI 템플릿(JS 포함) 업데이트.

## 5. 테스트 전략 (Testing Strategy)

- **Unit Test**: `internal/dialect/` 아래 각 구현체들의 `MapOracleType`, `CreateTableDDL`, `InsertStatement`(특히 MSSQL 1000 row 분할 검증)에 대한 테스트 작성.
- **Integration Test**: Direct 마이그레이션 시 `Dialect` 선택에 따라 올바른 드라이버명과 URL 포맷으로 `database/sql` 초기화가 이뤄지는지 검증.
- **E2E Test / File Output**: SQL 파일 출력 시 각 방언별 특성이 쿼리에 올바르게 묻어나는지 텍스트 파일 내용 검증.


### tasks.md

# v6 구현 작업

## 1. 방언(Dialect) 인터페이스
- [ ] `internal/dialect/dialect.go`를 생성하고 `Dialect` 인터페이스를 정의.

## 2. 방언 구현
- [ ] PostgreSQL용 `internal/dialect/postgres.go` 생성 (기존 로직 마이그레이션).
- [ ] MySQL용 `internal/dialect/mysql.go` 생성.
- [ ] MariaDB용 `internal/dialect/mariadb.go` 생성.
- [ ] SQLite용 `internal/dialect/sqlite.go` 생성.
- [ ] MSSQL용 `internal/dialect/mssql.go` 생성.

## 3. 구성(Config) 확장
- [ ] `--target-db` 및 `--target-url` 플래그를 추가하기 위해 `internal/config/config.go` 업데이트.
- [ ] `--pg-url`에 대한 하위 호환성 보장.

## 4. DDL 리팩토링
- [ ] DDL 생성(CreateTable, Sequence, Index)에 `Dialect` 인터페이스를 사용하도록 `internal/migration/ddl.go` 업데이트.

## 5. 마이그레이션 로직 리팩토링
- [ ] `Dialect`를 주입하기 위해 `internal/migration/migration.go` 업데이트.
- [ ] `Dialect`를 기반으로 동적 드라이버 초기화를 사용하도록 `MigrateTableDirect` 업데이트.
- [ ] `Dialect.InsertStatement`를 사용하도록 `ProcessRow` 및 `WriteBatch` 업데이트.

## 6. 웹 UI 및 API
- [ ] `targetDb` 및 `targetUrl`을 처리하기 위해 `internal/web/server.go` 업데이트.
- [ ] 대상 DB 드롭다운 및 동적 UI 변경 사항을 추가하기 위해 `internal/web/templates/index.html` 업데이트.

## 7. 테스트
- [ ] `Dialect` 구현체(`internal/dialect/*`)에 대한 단위 테스트 작성.
- [ ] 새로운 대상에 대한 통합 테스트 업데이트.


---
## <a name="v07"></a> v07

### prd.md

# PRD (Product Requirements Document) - v6 품질 개선 및 버그 수정 (v7)

## 1. 개요 (Overview)

v6에서 멀티 타겟 DB 지원(`--target-db`)이 도입되었습니다.
이 버전(v7)은 v6 코드베이스를 면밀히 검토하여 발견된 **버그 수정**, **기능 누락**, **Web UI 불일치**, **테스트 공백**을 체계적으로 해소하는 **품질 개선 릴리스**입니다.
신규 기능 추가보다 기존 기능의 완성도와 안정성을 높이는 데 집중합니다.

---

## 2. 배경 및 문제 분석 (Background & Issues Found)

코드베이스 분석을 통해 다음 7가지 카테고리의 개선점을 발견했습니다.

### 2.1. 타입 매핑 정밀도 부족

| 방언 | 대상 타입 | 현재 동작 | 문제 |
|------|-----------|-----------|------|
| MySQL | `VARCHAR2(n)` | `VARCHAR(255)` 하드코딩 | precision 무시 → 데이터 손실 가능 (`VARCHAR2(4000)` 등) |
| MySQL | `CHAR(n)` | `CHAR(255)` 하드코딩 | precision 무시 |
| MSSQL | `VARCHAR2(n)` | `NVARCHAR(MAX)` 전체 적용 | precision 있을 경우 `NVARCHAR(n)` 사용이 성능상 유리 |
| MSSQL | `CHAR(n)` | `NCHAR(255)` 하드코딩 | precision 무시 |
| MSSQL | `NUMBER` 精度 없음 | `FLOAT` | PRD 표에서는 `DECIMAL` 계열로 명시됨 |

### 2.2. MSSQL DDL 조건 검사 미흡

#### 2.2.1. `CreateTableDDL` — `TABLE_SCHEMA` 조건 누락

현재 코드(`mssql.go`):
```sql
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'tablename')
```
동일한 이름의 테이블이 다른 스키마에 있을 경우 `IF NOT EXISTS`가 잘못된 결과를 반환하여 테이블 생성을 건너뜁니다.

#### 2.2.2. `CreateIndexDDL` — `object_id` 필터 누락

현재 코드(`mssql.go`):
```sql
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_name')
```
인덱스 이름이 다른 테이블에 같은 이름으로 존재할 경우 중복으로 오인하여 인덱스를 생성하지 않습니다.

### 2.3. Web UI — DDL 옵션 파일 출력 모드 미노출

`index.html`에서 `withDdl`, `withSequences`, `withIndexes`, `oracleOwner` 옵션이 **Direct Migration 토글(`directMigration`) 영역 내부**에만 위치합니다.
CLI에서는 `--with-ddl --with-sequences --with-indexes` 플래그가 파일 출력 모드에서도 완전히 동작하지만, Web UI에서는 파일 출력 모드일 때 이 옵션들에 접근할 방법이 없습니다.

```
현재 UI 구조:
[Direct Migration 체크박스] ← 해제 상태
  └─ (숨겨짐) pgUrl
  └─ (숨겨짐) withDdl ← 파일 모드에서 접근 불가
  └─ (숨겨짐) withSequences
  └─ (숨겨짐) withIndexes
  └─ (숨겨짐) oracleOwner
```

### 2.4. Web UI — 레이블/제목 PostgreSQL 종속 잔재

| 위치 | 현재 값 | 문제 |
|------|---------|------|
| `server.go` title | `"Oracle to PostgreSQL Migrator"` | 멀티 타겟 DB 지원 후 부적절 |
| `index.html` Schema 레이블 | `"PG Schema"` | PostgreSQL 전용 레이블 |
| `index.html` `<html lang>` | `"en"` | 한국어 혼용 UI와 불일치 |

### 2.5. WebSocket `warning` 메시지 타입 미구현

v6 PRD §4.6.4에서 다음을 명시했습니다:
```json
{ "type": "warning", "message": "MySQL은 Sequence를 지원하지 않습니다. AUTO_INCREMENT로 대체됩니다." }
```
그러나 `ws/tracker.go`에 `MsgType = "warning"`이 없으며, 마이그레이션 엔진에서도 `tracker.Warning(...)` 호출이 없습니다.
결과적으로 Sequence 미지원 방언 사용 시 사용자는 WebSocket 메시지가 아닌 서버 로그로만 경고를 확인할 수 있습니다.

### 2.6. Dry-Run — 대상 DB 연결 검증 미지원

`migration.go`의 Dry-Run 모드는 Oracle DB에 `SELECT COUNT(*)` 쿼리만 수행합니다.
`DryRunResult`의 `connectionOk`가 항상 `true`로 전달되므로, Direct 마이그레이션 대상 URL이 잘못된 경우에도 Dry-Run이 성공으로 표시됩니다.

### 2.7. 테스트 커버리지 부족

`internal/dialect/` 패키지의 MySQL, MariaDB, SQLite, MSSQL 구현체에 대한 단위 테스트가 없습니다.
(v6 `tasks.md` §7 항목이 TODO 상태로 남아 있음)

---

## 3. 목표 (Goals)

1. MySQL / MSSQL 타입 매핑 시 precision 값을 올바르게 반영한다.
2. MSSQL `CreateTableDDL` · `CreateIndexDDL`에 스키마/테이블 범위 조건을 추가한다.
3. Web UI 파일 출력 모드에서 DDL 관련 옵션(with-ddl, sequences, indexes)을 노출한다.
4. UI의 PostgreSQL 종속 레이블을 방언 중립적으로 수정한다.
5. WebSocket `warning` 메시지 타입을 구현하고 Sequence 미지원 경고를 프론트엔드로 전달한다.
6. Dry-Run 시 대상 DB 연결도 함께 검증하고 결과를 UI에 표시한다.
7. 4개 방언(MySQL, MariaDB, SQLite, MSSQL) 단위 테스트를 추가한다.

---

## 4. 기능 요구사항 (Functional Requirements)

### 4.1. 타입 매핑 정확도 개선

#### 4.1.1. MySQL `MapOracleType` 수정

| Oracle 타입 | 현재 | 수정 후 |
|------------|------|---------|
| `VARCHAR2(n)` | `VARCHAR(255)` | `VARCHAR(n)` (n ≤ 16383이면 그대로, 초과 시 `LONGTEXT`) |
| `CHAR(n)` | `CHAR(255)` 하드코딩 | `CHAR(n)` (precision 반영) |

- MySQL utf8mb4 환경에서 최대 Row 크기(65535 bytes)를 고려하여, 단일 컬럼 `VARCHAR(n)`의 n이 16383(≈ 65535÷4)을 초과하면 자동으로 `LONGTEXT`로 매핑한다.

#### 4.1.2. MSSQL `MapOracleType` 수정

| Oracle 타입 | 현재 | 수정 후 |
|------------|------|---------|
| `VARCHAR2(n)` | `NVARCHAR(MAX)` | precision ≤ 4000이면 `NVARCHAR(n)`, 초과이면 `NVARCHAR(MAX)` |
| `CHAR(n)` | `NCHAR(255)` | `NCHAR(n)` (precision 반영, 최대 4000) |
| `NUMBER` (precision 없음) | `FLOAT` | `NUMERIC` (Oracle NUMBER 기본 매핑과 일치) |

### 4.2. MSSQL DDL 조건 검사 수정

#### 4.2.1. `CreateTableDDL` — `TABLE_SCHEMA` 조건 추가

```sql
-- 수정 전
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES
               WHERE TABLE_NAME = 'tablename')

-- 수정 후 (schema 지정 시)
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES
               WHERE TABLE_SCHEMA = 'schemaname' AND TABLE_NAME = 'tablename')

-- 수정 후 (schema 미지정 시, dbo 기본값 사용)
IF NOT EXISTS (SELECT 1 FROM INFORMATION_SCHEMA.TABLES
               WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = 'tablename')
```

#### 4.2.2. `CreateIndexDDL` — `object_id` 필터 추가

```sql
-- 수정 전
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_name')

-- 수정 후
IF NOT EXISTS (
    SELECT 1 FROM sys.indexes i
    JOIN sys.objects o ON i.object_id = o.object_id
    WHERE i.name = 'idx_name'
      AND o.name = 'tablename'
      AND SCHEMA_NAME(o.schema_id) = 'schemaname'
)
```

### 4.3. Web UI — DDL 옵션 파일 출력 모드 노출

현재 Direct Migration 토글 하위에만 있는 DDL 관련 옵션을 **공통 DDL 설정 섹션**으로 분리한다.

```
[목표 UI 구조]
Advanced Settings ▾
  ├─ 출력 대상 DB: [드롭다운]
  ├─ Batch Size / Workers
  └─ DDL 설정 (항상 표시, 파일 출력·Direct 공통)
       ├─ ☑ CREATE TABLE DDL 포함 (--with-ddl)
       ├─   ├─ ☐ Sequence DDL 포함 (--with-sequences)  ← withDdl 체크 시 활성
       ├─   ├─ ☐ Index DDL 포함 (--with-indexes)
       └─   └─ Oracle 소유자 입력 (oracle-owner)

[Direct Migration 체크박스]
  └─ (표시 시) 대상 URL 입력
```

- DDL 체크박스 비활성 상태에서 하위 옵션(sequences, indexes, oracleOwner)은 비활성화된다.
- Direct Migration 영역과 파일 출력 영역 모두 동일한 DDL 설정 섹션을 참조한다.

### 4.4. Web UI — 레이블/제목 수정

| 위치 | 현재 | 수정 후 |
|------|------|---------|
| `server.go` title 값 | `"Oracle to PostgreSQL Migrator"` | `"Oracle DB Migrator"` |
| Schema 입력 레이블 | `"PG Schema"` | `"Schema"` |

### 4.5. WebSocket `warning` 메시지 타입 구현

#### 4.5.1. `ws/tracker.go` 확장

```go
const (
    // 기존
    MsgInit         MsgType = "init"
    MsgUpdate       MsgType = "update"
    MsgDone         MsgType = "done"
    MsgError        MsgType = "error"
    MsgAllDone      MsgType = "all_done"
    MsgDryRunResult MsgType = "dry_run_result"
    MsgDDLProgress  MsgType = "ddl_progress"
    // 신규
    MsgWarning      MsgType = "warning"  // ← 추가
)
```

`WebSocketTracker`에 `Warning(message string)` 메서드를 추가한다.

#### 4.5.2. `ProgressTracker` 인터페이스 확장

```go
// WarningTracker extends ProgressTracker with warning broadcasting.
type WarningTracker interface {
    Warning(message string)
}
```

#### 4.5.3. 호출 위치

`MigrateTableToFile` / `MigrateTableDirect` 내부에서 Sequence DDL이 미지원(`supported == false`)인 경우:

```go
if wt, ok := tracker.(WarningTracker); ok {
    wt.Warning(fmt.Sprintf(
        "%s은(는) Sequence를 지원하지 않습니다. --with-sequences 옵션은 무시됩니다.",
        dia.Name(),
    ))
}
```

#### 4.5.4. Web UI `handleProgressMessage` 확장

```js
if (msg.type === 'warning') {
    // 경고 배너 또는 토스트 표시
    showWarningBanner(msg.message);
    return;
}
```

- 경고 메시지는 진행 컨테이너 상단에 노란색 배너로 표시된다.
- 동일한 경고가 여러 번 오면 중복 표시하지 않는다.

### 4.6. Dry-Run 대상 DB 연결 검증

Dry-Run 모드(`cfg.DryRun == true`)에서 `--target-url`이 지정된 경우, 대상 DB 연결을 시도하고 결과를 `DryRunResult.ConnectionOk`에 반영한다.

```go
// migration.go Run() 내 DryRun 블록
if cfg.TargetURL != "" {
    // 연결 시도 후 즉시 Close
    connOk = tryConnectTarget(dia, cfg.TargetURL)
}
for _, table := range cfg.Tables {
    // ...
    dryTracker.DryRunResult(table, count, connOk)
}
```

- 연결 실패 시 `connectionOk = false`로 전송하고, UI에 대상 DB 연결 실패 표시를 추가한다.

### 4.7. 방언별 단위 테스트 추가

`internal/dialect/` 패키지에 다음 테스트 파일을 추가한다:

| 파일 | 대상 |
|------|------|
| `mysql_test.go` | `MySQLDialect` |
| `mariadb_test.go` | `MariaDBDialect` |
| `sqlite_test.go` | `SQLiteDialect` |
| `mssql_test.go` | `MSSQLDialect` |

각 테스트 파일은 아래 테스트 케이스를 포함한다:

1. `TestMapOracleType_*` — 주요 Oracle 타입 → 대상 타입 매핑 검증
2. `TestCreateTableDDL_*` — 스키마 포함/미포함, NOT NULL 처리
3. `TestCreateIndexDDL_*` — 일반/UNIQUE/PK 인덱스, MSSQL object_id 조건 포함
4. `TestInsertStatement_*` — 단일 배치, 다중 배치 분할 (MSSQL 1000행 제한 포함)

---

## 5. 아키텍처 변경 요약 (Scope of Changes)

| 파일 | 변경 내용 |
|------|----------|
| `internal/dialect/mysql.go` | `MapOracleType`: VARCHAR2/CHAR precision 반영, VARCHAR 길이 범위 분기 |
| `internal/dialect/mssql.go` | `MapOracleType`: VARCHAR2/CHAR precision 반영, NUMBER 기본 매핑 수정 |
| `internal/dialect/mssql.go` | `CreateTableDDL`: TABLE_SCHEMA 조건 추가 |
| `internal/dialect/mssql.go` | `CreateIndexDDL`: sys.indexes + object_id 조건 추가 |
| `internal/dialect/mysql_test.go` | 신규 — MySQLDialect 단위 테스트 |
| `internal/dialect/mariadb_test.go` | 신규 — MariaDBDialect 단위 테스트 |
| `internal/dialect/sqlite_test.go` | 신규 — SQLiteDialect 단위 테스트 |
| `internal/dialect/mssql_test.go` | 신규 — MSSQLDialect 단위 테스트 |
| `internal/web/ws/tracker.go` | `MsgWarning` 상수, `Warning()` 메서드 추가 |
| `internal/migration/migration.go` | `WarningTracker` 인터페이스 정의, Dry-Run 대상 DB 연결 검증, Sequence 미지원 warning 호출 |
| `internal/web/server.go` | title 값 수정 |
| `internal/web/templates/index.html` | DDL 옵션 위치 재구성, 레이블 수정, warning 메시지 표시 로직 추가 |

---

## 6. 비기능 요구사항 (Non-Functional Requirements)

- **하위 호환성**: 모든 변경은 기존 CLI·API 인터페이스와 완전 호환되어야 한다.
- **테스트 통과**: `go test ./...` 전체가 통과해야 한다.
- **데이터 무결성**: VARCHAR 매핑 변경으로 인해 기존 PostgreSQL 출력 결과는 변경되지 않아야 한다.

---

## 7. 마일스톤 (Milestones)

1. **PRD 확정**: `docs/v7/prd.md` 작성 ✅
2. **타입 매핑 수정**: MySQL/MSSQL precision 반영
3. **MSSQL DDL 조건 수정**: TABLE_SCHEMA, object_id 필터
4. **WebSocket warning 구현**: tracker + 인터페이스 + 호출 지점
5. **Dry-Run 대상 DB 검증**: 연결 시도 로직 추가
6. **Web UI 개선**: DDL 섹션 재구성 + 레이블 수정 + warning 배너
7. **방언별 단위 테스트 추가**
8. **전체 테스트 통과 및 리뷰**

---

## 8. 미포함 사항 (Out of Scope)

다음은 이번 v7 범위에 포함되지 않으며, 별도 PRD로 분리합니다:

- 새로운 대상 DB 추가 (CockroachDB, TiDB 등)
- 소스 DB 다양화 (MySQL → PostgreSQL 등 Oracle 이외 소스)
- Web 서버 멀티 세션 동시 마이그레이션 지원
- 연결 풀 세밀 조정(MaxOpenConns 등) 설정 노출


### spec.md

# 기술 명세서 - v7 (품질 개선 및 버그 수정)

## 1. 타입 매핑 정밀도 개선
### 1.1. MySQL `MapOracleType`
- `VARCHAR2(n)`: `VARCHAR(n)`으로 매핑. 만약 `n > 16383`인 경우, `LONGTEXT`로 매핑.
- `CHAR(n)`: `CHAR(n)`으로 매핑 (정밀도 포함).

### 1.2. MSSQL `MapOracleType`
- `VARCHAR2(n)`: 만약 `n <= 4000`인 경우, `NVARCHAR(n)`으로 매핑. 만약 `n > 4000`인 경우, `NVARCHAR(MAX)`로 매핑.
- `CHAR(n)`: `NCHAR(n)`으로 매핑 (정밀도 포함, 최대 4000).
- `NUMBER` (정밀도 없음): `NUMERIC`으로 매핑.

## 2. MSSQL DDL 조건 확인
### 2.1. `CreateTableDDL`
- 다른 스키마에 동일한 이름의 테이블이 존재할 때 발생하는 오탐(false positive)을 방지하기 위해 `IF NOT EXISTS` 검사에 `TABLE_SCHEMA` 조건을 추가합니다.
- 스키마가 지정되지 않은 경우 기본값을 `dbo`로 설정합니다.

### 2.2. `CreateIndexDDL`
- `sys.objects`와 `sys.indexes`를 `object_id` 기준으로 조인하여 테이블 이름과 스키마로 엄격하게 필터링함으로써, 다른 테이블 간의 중복된 인덱스 이름 충돌을 방지합니다.

## 3. 웹 UI 개선
### 3.1. DDL 옵션 가시성
- DDL 관련 옵션(`--with-ddl`, `--with-sequences`, `--with-indexes`, `oracleOwner`)을 "Direct Migration" 토글 섹션에서 분리합니다.
- 이를 "고급 설정(Advanced Settings)" 아래의 공통 "DDL 설정" 섹션에 표시하여, 파일 출력(File Output)과 직접 마이그레이션(Direct Migration) 모드 모두에 적용 가능하고 눈에 띄게 만듭니다.

### 3.2. 레이블 및 제목 수정 (PostgreSQL 종속성 제거)
- `server.go` HTML 제목을 `"Oracle to PostgreSQL Migrator"`에서 `"Oracle DB Migrator"`로 업데이트합니다.
- `index.html`에서 Schema 입력 레이블을 `"PG Schema"`에서 `"Schema"`로 업데이트합니다.

## 4. WebSocket 경고 메시지 구현
- `ws/tracker.go`에 `MsgWarning MsgType = "warning"`을 추가합니다.
- `WebSocketTracker`에 `Warning(message string)` 메서드를 구현합니다.
- `ProgressTracker` 인터페이스를 `WarningTracker` 인터페이스로 확장합니다.
- 특정 방언(dialect)이 시퀀스 DDL을 지원하지 않을 때(예: MySQL) 마이그레이션 중에 경고 메시지를 발생시킵니다.
- 웹 UI(`index.html`)의 `handleProgressMessage`를 업데이트하여 `warning` 이벤트를 수신할 때 진행 컨테이너 상단에 경고 배너(노란색)를 표시합니다.

## 5. 예행 연습(Dry-Run) 대상 DB 연결 검증
- 예행 연습 모드(`cfg.DryRun == true`)에서 `--target-url`이 지정된 경우 대상 데이터베이스 연결을 시도합니다.
- `DryRunResult.ConnectionOk`를 통해 연결 성공/실패 결과를 전송합니다.
- 대상 DB 연결 상태를 웹 UI에 반영합니다.

## 6. 단위 테스트 추가
- `internal/dialect/`에 테스트 파일을 생성합니다:
  - `mysql_test.go`
  - `mariadb_test.go`
  - `sqlite_test.go`
  - `mssql_test.go`
- 테스트 케이스는 다음을 포괄해야 합니다:
  1. `TestMapOracleType_*`
  2. `TestCreateTableDDL_*`
  3. `TestCreateIndexDDL_*`
  4. `TestInsertStatement_*`


### tasks.md

# v7 Implementation Tasks

## 1. MySQL 타입 매핑 precision 반영

> **파일**: `internal/dialect/mysql.go`
> **Spec 참조**: §2

- [x] `MapOracleType`에서 `VARCHAR2`와 `CHAR`의 합쳐진 `case` 분기를 분리한다.
- [x] `VARCHAR2` 분기: precision > 0이면 `VARCHAR(n)` 반환, n > 16383이면 `LONGTEXT` 반환.
- [x] `CHAR` 분기: precision > 0이면 `CHAR(n)` 반환, 기본값 `CHAR(255)` 유지.
- [x] 기존 `NUMBER`, `DATE`, `CLOB`, `BLOB`, `FLOAT` 분기는 변경하지 않는다.

## 2. MSSQL 타입 매핑 precision 반영

> **파일**: `internal/dialect/mssql.go`
> **Spec 참조**: §3

- [x] `MapOracleType`에서 `VARCHAR2` 분기: precision > 0 && ≤ 4000이면 `NVARCHAR(n)`, 초과이면 `NVARCHAR(MAX)`.
- [x] `CHAR` 분기: precision > 0이면 `NCHAR(n)` (최대 4000 클램핑), 기본값 `NCHAR(255)` 유지.
- [x] `NUMBER` (precision 없음) 기본 반환값을 `FLOAT` → `NUMERIC`으로 변경한다.

## 3. MSSQL `CreateTableDDL` — TABLE_SCHEMA 조건 추가

> **파일**: `internal/dialect/mssql.go`
> **Spec 참조**: §4.1

- [x] `IF NOT EXISTS` 쿼리에 `TABLE_SCHEMA` 조건을 추가한다.
- [x] schema가 빈 문자열이면 `dbo`를 기본값으로 사용한다.
- [x] 기존 `TABLE_NAME` 조건은 유지한다.

## 4. MSSQL `CreateIndexDDL` — object_id 필터 추가

> **파일**: `internal/dialect/mssql.go`
> **Spec 참조**: §4.2

- [x] `IF NOT EXISTS` 쿼리를 `sys.indexes i JOIN sys.objects o ON i.object_id = o.object_id` 형태로 변경한다.
- [x] `WHERE` 절에 `i.name`, `o.name` (테이블명), `SCHEMA_NAME(o.schema_id)` (스키마명) 조건을 추가한다.
- [x] schema가 빈 문자열이면 `dbo`를 기본값으로 사용한다.
- [x] `IsPK == true`인 경우의 `ALTER TABLE ADD PRIMARY KEY` 로직은 변경하지 않는다.

## 5. WebSocket `warning` 메시지 타입 — tracker 확장

> **파일**: `internal/web/ws/tracker.go`
> **Spec 참조**: §5.1

- [x] `MsgWarning MsgType = "warning"` 상수를 추가한다.
- [x] `ProgressMsg` 구조체에 `Message string \`json:"message,omitempty"\`` 필드를 추가한다.
- [x] `WebSocketTracker`에 `Warning(message string)` 메서드를 추가한다. (`broadcast`로 `MsgWarning` + `Message` 전송)

## 6. `WarningTracker` 인터페이스 및 호출 지점

> **파일**: `internal/migration/migration.go`
> **Spec 참조**: §5.2, §5.3

- [x] `WarningTracker` 인터페이스를 정의한다: `Warning(message string)`.
- [x] `MigrateTableDirect` 내 `cfg.WithSequences` 블록에서 `supported == false`일 때 `WarningTracker.Warning()` 호출을 추가한다.
- [x] `MigrateTableToFile` 내 동일 위치에 같은 패턴을 적용한다.

## 7. Dry-Run 대상 DB 연결 검증

> **파일**: `internal/migration/migration.go`
> **Spec 참조**: §6

- [x] `tryConnectTarget(dia dialect.Dialect, targetURL string) bool` 헬퍼 함수를 추가한다.
  - PostgreSQL: `db.ConnectPostgres` → 즉시 Close.
  - 기타: `sql.Open` + `Ping()` → 즉시 Close.
- [x] `Run()` 함수의 DryRun 블록에서 `cfg.TargetURL != ""`이면 `tryConnectTarget`를 호출한다.
- [x] 결과를 `connOk` 변수에 저장하고, `DryRunResult(table, count, connOk)`에 전달한다.

## 8. Web UI — 타이틀 및 레이블 수정

> **파일**: `internal/web/server.go`, `internal/web/templates/index.html`
> **Spec 참조**: §7.1, §7.2

- [x] `server.go`에서 title 값을 `"Oracle to PostgreSQL Migrator"` → `"Oracle DB Migrator"`로 변경한다.
- [x] `index.html`에서 Schema 입력 필드 레이블을 `"PG Schema"` → `"Schema"`로 변경한다.

## 9. Web UI — DDL 옵션 파일 출력 모드 노출

> **파일**: `internal/web/templates/index.html`
> **Spec 참조**: §7.3

- [x] DDL 관련 옵션(`withDdl`, `withSequences`, `withIndexes`, `oracleOwner`)을 Direct Migration 토글 내부에서 **공통 DDL 설정 섹션**으로 분리한다.
- [x] DDL 설정 섹션은 Direct Migration 체크 여부와 무관하게 항상 표시되도록 한다.
- [x] `withDdl` 체크박스 변경 이벤트: 해제 시 하위 옵션(sequences, indexes, oracleOwner)을 숨기고 값을 초기화한다.
- [x] Direct Migration 토글 내부에는 대상 URL 입력만 남긴다.

## 10. Web UI — warning 배너 및 Dry-Run 연결 실패 표시

> **파일**: `internal/web/templates/index.html`
> **Spec 참조**: §7.4, §7.5

- [x] `handleProgressMessage`에 `msg.type === 'warning'` 분기를 추가하여 `showWarningBanner(msg.message)`를 호출한다.
- [x] `showWarningBanner` 함수를 구현한다: `Set`으로 중복 방지, 진행 컨테이너 상단에 노란색 배너 삽입.
- [x] `.warning-banner` CSS 스타일을 추가한다 (`#fff3cd` 배경, `#ffc107` 테두리, `#856404` 텍스트).
- [x] `dry_run_result` 처리에서 `connection_ok === false`이면 대상 DB 연결 실패 경고 배너를 표시한다.

## 11. 단위 테스트 — MySQL

> **파일**: `internal/dialect/mysql_test.go` (신규)
> **Spec 참조**: §8

- [x] `TestMapOracleType_MySQL` — VARCHAR2 precision 반영 (정상, 초과→LONGTEXT, 미지정→255), CHAR precision 반영, NUMBER, DATE, CLOB, BLOB, FLOAT 매핑.
- [x] `TestCreateTableDDL_MySQL` — 스키마 포함/미포함, NOT NULL 처리, `IF NOT EXISTS` 포함 검증.
- [x] `TestCreateIndexDDL_MySQL` — 일반 인덱스, UNIQUE 인덱스, PRIMARY KEY (ALTER TABLE).
- [x] `TestInsertStatement_MySQL` — 단일 배치, 다중 배치 분할, 값 포맷팅 (문자열 이스케이프, NULL, 시간).

## 12. 단위 테스트 — MariaDB

> **파일**: `internal/dialect/mariadb_test.go` (신규)
> **Spec 참조**: §8.3

- [x] `TestMariaDB_Name` — `Name()` 반환값이 `"mariadb"`인지 검증.
- [x] `TestMariaDB_InheritsMySQL` — MySQL과 동일한 타입 매핑 (VARCHAR2 precision 포함) 확인.
- [x] `TestCreateTableDDL_MariaDB` — MySQL과 동일한 DDL 생성 확인.
- [x] `TestInsertStatement_MariaDB` — MySQL과 동일한 INSERT 문 생성 확인.

## 13. 단위 테스트 — SQLite

> **파일**: `internal/dialect/sqlite_test.go` (신규)
> **Spec 참조**: §8

- [x] `TestMapOracleType_SQLite` — VARCHAR2→TEXT, NUMBER→INTEGER/REAL, DATE→TEXT, CLOB→TEXT, BLOB→BLOB, FLOAT→REAL.
- [x] `TestCreateTableDDL_SQLite` — 스키마 무시 (schema 전달해도 DDL에 미포함), NOT NULL, `IF NOT EXISTS`.
- [x] `TestCreateIndexDDL_SQLite` — 일반 인덱스 `IF NOT EXISTS`, UNIQUE, PK 건너뛰기 (빈 문자열 반환).
- [x] `TestInsertStatement_SQLite` — 단일 배치, 다중 배치 분할, 값 포맷팅.

## 14. 단위 테스트 — MSSQL

> **파일**: `internal/dialect/mssql_test.go` (신규)
> **Spec 참조**: §8

- [x] `TestMapOracleType_MSSQL` — VARCHAR2 precision ≤4000 → NVARCHAR(n), >4000 → NVARCHAR(MAX), CHAR precision → NCHAR(n), NUMBER 무precision → NUMERIC (v7 수정).
- [x] `TestCreateTableDDL_MSSQL` — TABLE_SCHEMA 조건 포함 (schema 지정/미지정→dbo), NOT NULL, 컬럼 타입 (v7 수정).
- [x] `TestCreateIndexDDL_MSSQL` — sys.objects JOIN + object_id 필터 (v7 수정), 스키마 포함/미포함→dbo, UNIQUE, PK (ALTER TABLE).
- [x] `TestInsertStatement_MSSQL` — 1000행 배치 제한 (batchSize > 1000 → 자동 분할), 값 포맷팅 (N'' 유니코드, 0x 바이너리).

## 15. WebSocket tracker 테스트 확장

> **파일**: `internal/web/ws/tracker_test.go`
> **Spec 참조**: §10.2

- [x] `TestWarning` — `Warning()` 호출 시 `MsgWarning` 타입 + `Message` 필드가 올바르게 브로드캐스트되는지 검증.

## 16. 전체 테스트 통과 확인

- [x] `go test ./...` 실행하여 전체 테스트 통과를 확인한다.
- [x] `--target-db postgres` 출력이 v7 수정 전과 동일한지 확인한다 (PostgreSQL 하위 호환성).


---
## <a name="v08"></a> v08

### prd.md

# PRD (Product Requirements Document) - v8 확장성 및 엔터프라이즈 기능 강화 (v8)

## 1. 개요 (Overview)

v7을 통해 타입 매핑의 정밀도, MSSQL DDL의 정확성, Web UI의 사용성 및 코드베이스의 안정성(테스트 커버리지)이 크게 개선되었습니다.
이번 **v8 릴리스**는 이를 바탕으로 **운영 환경에서의 다중 사용자 지원, 대규모 데이터 마이그레이션 안정성, 그리고 스키마 마이그레이션의 완전성**을 확보하는 데 집중하는 **엔터프라이즈 스케일업 릴리스**입니다.

---

## 2. 배경 및 문제 분석 (Background & Issues Found)

현재 v7 시스템에서는 엔터프라이즈 규모의 운영을 위해 다음 4가지 핵심적인 개선 영역이 식별되었습니다.

### 2.1. Web UI 멀티 세션 간섭 (Global State Issue)
- **현상**: `web/server.go`에 `tracker` 객체가 전역 변수(`var tracker = ws.NewWebSocketTracker()`)로 선언되어 있습니다.
- **문제**: 두 명 이상의 사용자가 동시에 웹 브라우저를 열어 다른 마이그레이션 작업을 수행할 경우, 진행 상황(Progress)과 경고(Warning) 메시지가 모든 사용자에게 브로드캐스트되어 데이터와 UI가 혼재되는 치명적인 문제가 발생합니다.

### 2.2. 제약조건(Constraints) 마이그레이션 누락
- **현상**: 현재 스키마 마이그레이션은 테이블 구조(컬럼 타입, NOT NULL), 인덱스, 시퀀스, 그리고 Primary Key(인덱스 생성 시 포함)만 지원합니다.
- **문제**: Foreign Key, Check 제약조건, 그리고 컬럼의 Default 값이 누락되어 완전한 데이터베이스 스키마 이관이 불가능합니다. 이는 이관 후 무결성 검증을 수동으로 해야 하는 부담을 낳습니다.

### 2.3. DB 연결 자원(Connection Pool) 통제권 부재
- **현상**: `--workers`를 늘릴 경우, 시스템은 기본값에 의존하여 DB 연결을 무한정 생성하려 시도합니다.
- **문제**: 타겟 DB나 소스 DB의 `max_connections`를 초과하여 연결 거부(Connection Refused) 오류가 발생할 수 있습니다. 사용자가 `MaxOpenConns`, `MaxIdleConns`, `ConnMaxLifetime` 등을 설정할 수 있는 인터페이스가 없습니다.

### 2.4. 대용량 테이블 마이그레이션 실패 시 재개(Resume) 불가
- **현상**: 수천만 건의 데이터를 가진 테이블을 마이그레이션하던 중 네트워크 단절이나 OOM으로 실패할 경우, 처음부터 다시 복사해야 합니다.
- **문제**: 진행 상태를 저장하는 체크포인트(Checkpoint) 기능이 없어 대용량 마이그레이션 시 시간과 자원의 낭비가 큽니다.

---

## 3. 목표 (Goals)

1. **완벽한 격리(Isolation)**: Web UI에서 사용자별 세션(Session ID)을 발급하여 멀티 테넌트 동시 마이그레이션을 지원한다.
2. **스키마 완전성(Completeness)**: Foreign Key, Check 제약조건 및 컬럼 Default Value를 추출하고 DDL을 생성한다.
3. **자원 최적화(Resource Management)**: DB 커넥션 풀을 세밀하게 제어할 수 있는 설정(CLI/UI)을 제공하여 DB 과부하를 방지한다.
4. **회복 탄력성(Resilience)**: 마이그레이션 실패 시 중단된 지점부터 다시 시작할 수 있는 체크포인트(Resumable) 기반 복구 기능을 도입한다.

---

## 4. 기능 요구사항 (Functional Requirements)

### 4.1. Web UI 다중 세션 지원 (Multi-Session WebSocket)

- **세션 발급**: 사용자가 웹페이지에 접속할 때 UUID 기반의 `SessionID`를 발급한다.
- **Tracker 분리**: `WebSocketTracker`를 전역 싱글톤에서 세션별 인스턴스로 분리 관리하는 `SessionManager` 구조로 변경한다.
- **API 연동**: `/api/migrate` 요청 시 `SessionID`를 페이로드로 받아, 해당 세션에 종속된 Tracker를 `migration.Run()`에 주입한다.
- **UI 표시**: 현재 연결된 Session ID를 화면 하단에 표시하여 문제 발생 시 추적이 용이하게 한다.

### 4.2. 추가 제약조건(Constraints) 마이그레이션 지원

- **Default Value 지원**:
  - `GetTableMetadata` 실행 시 Oracle의 `DATA_DEFAULT` 값을 조회하여 `CREATE TABLE` 구문에 포함시킨다.
- **Foreign Key 및 Check 제약조건**:
  - `GetConstraintMetadata` 함수를 신설하여 `ALL_CONSTRAINTS`에서 FK, CHECK 조건을 추출한다.
  - `--with-constraints` (UI: "제약조건 포함") 플래그를 추가한다.
  - **순서 보장 전략**: Foreign Key는 참조되는 테이블이 먼저 생성되어야 하므로, 모든 테이블의 `CREATE TABLE` 및 데이터 이관(INSERT/COPY)이 완전히 종료된 후, **마지막 단계에 `ALTER TABLE ADD CONSTRAINT` 형태로 일괄 실행**한다.

### 4.3. Connection Pool 세밀 조정 설정

- **설정 항목 추가 (Config / CLI / UI)**:
  - `--db-max-open` (기본값: 0, 무제한)
  - `--db-max-idle` (기본값: 2)
  - `--db-max-life` (기본값: 0, 무제한)
- **적용 대상**:
  - Oracle 소스 DB 연결 (`sql.DB.SetMaxOpenConns` 등 적용)
  - Postgres 대상 풀 (`pgxpool.Config.MaxConns`)
  - 기타 대상 DB (`sql.DB.SetMaxOpenConns`)
- **Web UI**: "Advanced Settings" 하위에 DB Connection Pool 튜닝 섹션을 추가한다.

### 4.4. Resumable 마이그레이션 (체크포인트)

- **진행 상태 저장**:
  - 로컬 파일시스템에 `.migration_state/{job_id}.json` 형태로 각 테이블의 완료된 Offset(또는 Primary Key의 Max Value)을 기록한다.
- **재개 명령**:
  - `--resume {job_id}` 플래그를 통해 실패한 마이그레이션을 다시 시작한다.
- **구현 제약**:
  - 단순 Batch 기반(LIMIT/OFFSET) 혹은 PK 기반 분할(Chunking) 조회를 Oracle 소스 조회 쿼리에 적용해야 한다. v8에서는 가장 범용적인 PK 기반 Chunking 조회를 우선 도입한다.

---

## 5. 아키텍처 및 인터페이스 변경 (Scope of Changes)

| 변경 영역 | 파일 / 모듈 | 상세 내용 |
|-----------|-------------|----------|
| **세션 관리** | `web/server.go`, `ws/tracker.go` | SessionManager 도입, 전역 변수 제거, Gin Route 파라미터로 SessionID 연동. |
| **메타데이터** | `db/oracle.go` | `DATA_DEFAULT` 컬럼 추가, `FetchConstraints` 쿼리 및 DTO 추가. |
| **DDL 생성** | `dialect/*.go` | `CreateConstraintDDL` 인터페이스 추가. 각 방언별 FK, CHECK 구문 구현. |
| **마이그레이션** | `migration/migration.go` | 전체 데이터 이관 루프 종료 후, Constraint DDL을 별도로 실행하는 후처리(Post-processing) 단계 추가. |
| **DB 연결** | `db/connect.go` | Config 구조체에 커넥션 풀 속성 수용 및 `sql.DB` 인스턴스에 적용. |
| **상태 관리** | `migration/state.go` (신규) | 체크포인트 저장/로드, 재개 시 Oracle 조회 쿼리의 `WHERE PK > ?` 동적 생성. |
| **Web UI** | `templates/index.html` | Session ID 핸들링, 풀링 설정 입력란, Constraints 체크박스 추가. |

---

## 6. 비기능 요구사항 (Non-Functional Requirements)

- **메모리 안정성**: 멀티 세션 지원 시, 오래된/종료된 세션의 Tracker와 WebSocket 연결은 즉시 가비지 컬렉션(GC) 되도록 확실한 정리(Cleanup) 로직을 구현해야 한다.
- **하위 호환성**: `--resume`나 추가 커넥션 풀 파라미터를 입력하지 않더라도 v7과 동일하게 기본값으로 안전하게 동작해야 한다.

---

## 7. 마일스톤 (Milestones)

1. **1단계: 동시성 아키텍처 개편** (Web UI 다중 세션 적용 및 GC 로직 구현)
2. **2단계: 자원 통제 도입** (Connection Pool 파라미터 적용 및 CLI/UI 노출)
3. **3단계: 스키마 완전성 확보** (Default, FK, Check 메타데이터 추출 및 후처리 로직 구현)
4. **4단계: Resumable 엔진 개발** (상태 저장 파일 I/O 및 PK 기반 청크 쿼리 지원)
5. **5단계: 통합 테스트 및 QA** (대용량 단절 테스트, 멀티 브라우저 동시 이관 테스트)

---

## 8. 기대 효과 (Expected Outcomes)

v8 업데이트를 통해 단일 사용자용 툴에서 **팀 단위 운영 및 대규모 데이터 센터 이관**에 적합한 엔터프라이즈급 안정성을 확보하게 됩니다. 특히 중단 없는 재개 기능과 정확한 제약조건 생성은 실무 데이터베이스 엔지니어(DBA)들의 수동 후속 작업 시간을 극적으로 단축시킬 것입니다.


### spec.md

# 기술 명세서 - v8 (확장성 및 엔터프라이즈 기능)

## 1. 웹 UI 다중 세션 지원
- **세션 관리**: 글로벌 싱글톤 패턴을 대체하여 여러 `WebSocketTracker` 인스턴스를 처리하기 위해 `web/server.go` 및 `ws/tracker.go`에 `SessionManager`를 도입합니다.
- **세션 ID**: 웹페이지에 접속할 때 연결되는 각 클라이언트에 대해 UUID 기반의 `SessionID`를 생성합니다.
- **API 통합**: 페이로드에서 `SessionID`를 수락하도록 `/api/migrate` 엔드포인트를 수정하여 마이그레이션(`migration.Run()`)에 올바른 세션별 트래커가 주입되도록 보장합니다.
- **가비지 컬렉션(Garbage Collection)**: 오래되거나 연결이 끊긴 세션을 메모리에서 안전하게 제거하는 정리 로직을 구현합니다.
- **UI 추적**: 디버깅 목적으로 UI에 현재 연결된 세션 ID를 표시합니다.

## 2. 제약 조건(Constraints) 마이그레이션 지원
### 2.1. 기본값 (Default Values)
- `GetTableMetadata` (`db/oracle.go`)의 Oracle 메타데이터에서 `DATA_DEFAULT`를 추출합니다.
- 추출된 기본값을 생성된 `CREATE TABLE` DDL 문에 통합합니다.

### 2.2. 외래 키(Foreign Key) 및 CHECK 제약 조건
- **메타데이터 추출**: `ALL_CONSTRAINTS`에서 외래 키(FK) 및 CHECK 제약 조건을 추출하는 `GetConstraintMetadata` 함수를 생성합니다.
- **CLI/UI 통합**: `--with-constraints` 플래그 (웹 UI: "Include Constraints" 체크박스)를 추가합니다.
- **방언(Dialect) DDL**: `CreateConstraintDDL` 인터페이스를 정의하고 지원되는 각 방언에 대한 제약 조건 구문을 구현합니다.
- **후처리(Post-processing) 실행**: 참조 무결성 순서를 보장하기 위해 모든 테이블 구조가 생성되고 모든 데이터가 완전히 삽입/복사된 후 마지막 단계로 FK에 대해서만 `ALTER TABLE ADD CONSTRAINT`를 실행합니다.

## 3. 연결 풀 미세 조정 (Connection Pool Fine-Tuning)
- **구성 속성**: 고급 연결 풀 매개변수를 추가합니다:
  - `--db-max-open` (기본값: 0, 무제한)
  - `--db-max-idle` (기본값: 2)
  - `--db-max-life` (기본값: 0, 무제한)
- **애플리케이션 적용**: 이러한 설정을 Oracle 소스 `sql.DB` 인스턴스, Postgres 대상 `pgxpool.Config` 및 `db/connect.go`의 다른 모든 대상 `sql.DB` 인스턴스에 직접 적용합니다.
- **웹 UI**: "고급 설정(Advanced Settings)" 내의 새로운 "DB Connection Pool Tuning" 섹션 아래에 이러한 설정을 노출합니다.

## 4. 재개 가능한 마이그레이션 (Checkpoints)
- **상태 관리**: 마이그레이션 상태를 관리하기 위해 새로운 모듈 `migration/state.go`를 생성합니다.
- **체크포인트(Checkpointing)**: 각 처리된 테이블에 대해 완료된 오프셋 또는 최대 기본 키(Primary Key) 값을 기록하여 진행 상황을 로컬 파일 시스템(`.migration_state/{job_id}.json`)에 유지합니다.
- **재개 명령(Resume Command)**: 중단된 마이그레이션을 마지막으로 성공한 체크포인트에서 재개하기 위해 `--resume {job_id}` CLI 플래그를 도입합니다.
- **청크 쿼리(Chunking Queries)**: 데이터 전체를 다시 가져오지 않고 중단된 위치에서 올바르게 데이터 가져오기를 계속하기 위해 Oracle 데이터 추출용 PK 기반 청크/페이지 매김 쿼리를 구현합니다.


### tasks.md

# 구현 작업 - v8 (확장성 및 엔터프라이즈 기능)

## 1. 웹 UI 다중 세션 지원
- [x] **세션 관리자(Session Manager) 구현**
  - [x] 여러 `WebSocketTracker` 인스턴스를 처리하기 위해 `ws/tracker.go`에 `SessionManager` 구조체를 생성.
  - [x] 글로벌 `tracker` 싱글톤 제거.
- [x] **세션 ID 생성 및 처리**
  - [x] `web/server.go`에서 연결하는 클라이언트를 위한 UUID 기반 `SessionID` 생성.
  - [x] 요청 페이로드에서 `SessionID`를 수락하고 파싱하도록 `/api/migrate` 엔드포인트 업데이트.
  - [x] 올바른 세션별 트래커를 `migration.Run()`에 주입.
- [x] **세션 정리 (가비지 컬렉션)**
  - [x] 연결이 끊기거나 오래된 세션을 안전하게 닫고 메모리에서 제거하는 로직 구현.
- [x] **UI 업데이트**
  - [x] 디버깅 및 추적을 위해 웹 UI(`index.html`)에 연결된 `SessionID` 표시.

## 2. 제약 조건(Constraints) 마이그레이션 지원
- [x] **기본값 (Default Values)**
  - [x] `DATA_DEFAULT`를 추출하기 위해 `db/oracle.go`의 `GetTableMetadata` 업데이트.
  - [x] 추출된 기본값을 포함하도록 `CREATE TABLE` DDL 생성 업데이트.
- [x] **외래 키(Foreign Key) 및 CHECK 제약 조건**
  - [x] FK 및 CHECK 제약 조건을 추출하기 위해 `db/oracle.go`에 `GetConstraintMetadata` 함수 구현.
  - [x] CLI 구성에 `--with-constraints` 플래그 추가.
  - [x] 웹 UI(`index.html`)에 "Include Constraints" 체크박스 추가.
  - [x] dialect 패키지에 `CreateConstraintDDL` 인터페이스 정의.
  - [x] 지원되는 모든 방언(PostgreSQL, MySQL, MariaDB, SQLite, MSSQL)에 대해 `CreateConstraintDDL` 구현.
- [x] **후처리(Post-processing) 실행**
  - [x] FK 제약 조건 실행을 연기(defer)하도록 `migration/migration.go` 수정.
  - [x] 데이터 삽입이 완전히 완료된 후 추출된 모든 FK에 대해 `ALTER TABLE ADD CONSTRAINT`를 실행하는 마지막 단계 구현.

## 3. 연결 풀 미세 조정 (Connection Pool Fine-Tuning)
- [x] **구성 업데이트**
  - [x] CLI 플래그 및 내부 Config 구조체에 `--db-max-open`, `--db-max-idle`, `--db-max-life` 매개변수 추가.
- [x] **연결 적용**
  - [x] Oracle 소스 `sql.DB` 인스턴스에 풀 설정 적용.
  - [x] Postgres 대상 `pgxpool.Config`에 풀 설정 적용.
  - [x] `db/connect.go`의 다른 모든 대상 `sql.DB` 인스턴스에 풀 설정 적용.
- [x] **UI 업데이트**
  - [x] `index.html`의 "고급 설정(Advanced Settings)" 섹션 내에 "DB Connection Pool Tuning" 입력 추가.

## 4. 재개 가능한 마이그레이션 (Checkpoints)
- [x] **상태 관리**
  - [x] 상태 처리를 위한 `migration/state.go` 모듈 생성.
  - [x] `.migration_state/{job_id}.json`에 진행 상황(오프셋 또는 최대 PK 값)을 저장/유지하는 기능 구현.
- [x] **재개(Resume) 기능**
  - [x] `--resume {job_id}` CLI 플래그를 추가하고 초기 시작 로직에 통합.
- [x] **데이터 추출 리팩토링**
  - [x] Oracle 데이터 추출을 위한 PK 기반 청크/페이지 매김 쿼리 구현.
  - [x] 기존 데이터를 다시 가져오지 않고 마지막으로 기록된 체크포인트에서 동적으로 추출을 시작할 수 있도록 보장.


---
## <a name="v09"></a> v09

### prd.md

# PRD (Product Requirements Document) - v9 안정성·관측성·데이터 무결성 강화

## 1. 개요 (Overview)

v8까지 엔터프라이즈 스케일업(멀티 세션, 제약조건, 커넥션 풀, 체크포인트)이 완료되고, ui-improve 브랜치에서 2컬럼 레이아웃·다크모드·접근성이 도입되었습니다.

이번 **v9 릴리스**는 실 운영 환경에서 드러나는 **데이터 무결성 검증, 마이그레이션 관측성(Observability), 에러 복구 개선, 그리고 코드 구조의 유지보수성**에 집중하는 **운영 안정화 릴리스**입니다.

---

## 2. 배경 및 문제 분석 (Background & Issues Found)

### 2.1. 데이터 무결성 검증 수단 부재

- **현상**: 마이그레이션 완료 후 소스(Oracle)와 타겟 DB 간의 데이터 정합성을 확인할 방법이 없습니다.
- **문제**: `Done` 상태가 되어도 실제 행 수 불일치, 컬럼 값 누락, 인코딩 불일치 등을 감지할 수 없어, DBA가 수동으로 `SELECT COUNT(*)` 비교나 샘플링 검증을 해야 합니다. 대규모 환경일수록 이 수동 검증 비용이 기하급수적으로 증가합니다.

### 2.2. 마이그레이션 감사 로그(Audit Trail) 부재

- **현상**: 마이그레이션 수행 이력(시작 시각, 종료 시각, 테이블별 행 수, 오류 내역, 수행자)이 체계적으로 기록되지 않습니다.
- **문제**: 운영 환경에서 "언제, 누가, 어떤 테이블을, 얼마나 이관했는가"를 사후 추적할 수 없습니다. 장애 발생 시 원인 분석이 어렵고, 컴플라이언스 요건(감사 추적)을 충족하지 못합니다.

### 2.3. 에러 발생 시 컨텍스트 손실

- **현상**: 에러 메시지가 `fmt.Errorf("failed to execute batch insert: %v", err)` 형태로 전파되어, 어느 배치의 몇 번째 행에서 어떤 컬럼 값으로 인해 실패했는지 알 수 없습니다.
- **문제**: 타입 불일치(예: Oracle `NUMBER`가 `NULL`인데 타겟에서 `NOT NULL`)나 인코딩 문제 발생 시 디버깅에 수십 분~수 시간이 소요됩니다. Web UI에서도 `"error"` 타입 메시지만 표시되어 사용자가 원인을 파악하기 어렵습니다.

### 2.4. PostgreSQL COPY 모드에서 진행률 미표시 및 부분 재개 불가

- **현상**: `pgx.CopyFrom`은 전체 데이터를 한 번에 스트리밍하므로, 진행 중간에 `tracker.Update()`가 호출되지 않습니다. 또한 COPY 중간에 실패하면 전체 테이블을 처음부터 다시 복사해야 합니다.
- **문제**: 대용량 테이블(수천만 건) 마이그레이션 시 Web UI가 0% → 100%로 갑자기 점프하여 사용자 경험이 나쁘고, 실패 시 체크포인트 기반 재개가 불가능합니다. v8에서 도입한 `MigrationState`가 COPY 모드에서는 무용지물입니다.

### 2.5. SQL Injection 취약점

- **현상**: `migration.go`에서 `fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)` 및 `fmt.Sprintf("SELECT * FROM %s", tableName)` 형태로 테이블명을 직접 쿼리에 삽입합니다.
- **문제**: Web UI를 통해 악의적인 테이블명이 전달될 경우 SQL Injection이 가능합니다. `validateMigrationRequest()`에서 테이블명 검증이 누락되어 있습니다.

### 2.6. 대형 소스 파일의 유지보수성 저하

- **현상**: Dialect 구현 파일들이 5,000~8,650줄에 달하며(`mssql.go` 8,650줄, `postgres.go` 7,522줄, `mysql.go` 7,161줄), `index.html`이 2,000줄 이상의 CSS/JS를 인라인으로 포함합니다.
- **문제**: 단일 파일이 커질수록 변경 시 merge conflict 확률 증가, IDE 성능 저하, 코드 리뷰 효율 감소가 발생합니다. 특히 dialect 파일은 DDL 생성과 INSERT 생성이 단일 파일에 혼재되어 있어 관심사 분리가 미흡합니다.

---

## 3. 목표 (Goals)

1. **데이터 무결성 보장**: 마이그레이션 후 소스-타겟 간 자동 검증(행 수 비교 + 체크섬 샘플링)을 수행하여 정합성을 보증한다.
2. **완전한 관측성**: 마이그레이션 전 과정을 구조화된 감사 로그로 기록하고, 결과 리포트를 생성한다.
3. **정밀한 에러 진단**: 실패 시 배치 번호, 행 오프셋, 문제 컬럼, 원인 분류를 포함한 상세 에러 컨텍스트를 제공한다.
4. **COPY 모드 개선**: PostgreSQL 대상 대용량 마이그레이션에서도 실시간 진행률 표시와 배치 단위 재개를 지원한다.
5. **보안 강화**: 모든 사용자 입력(테이블명, 스키마명)에 대한 검증을 강화하여 SQL Injection을 방지한다.
6. **코드 구조 개선**: Dialect 파일 분할 및 Web UI 에셋 분리로 유지보수성을 확보한다.

---

## 4. 기능 요구사항 (Functional Requirements)

### 4.1. 마이그레이션 후 데이터 검증 (Post-Migration Validation)

- **행 수(Row Count) 비교**:
  - 각 테이블에 대해 소스(`SELECT COUNT(*) FROM {table}`)와 타겟의 행 수를 비교한다.
  - 불일치 시 `warning` 레벨로 로그를 남기고, WebSocket으로 `validation_result` 메시지를 전송한다.

- **체크섬 샘플링 (선택적)**:
  - `--validate` 플래그(UI: "마이그레이션 후 검증") 활성화 시, 각 테이블에서 무작위 N개(기본 1000개) 행을 추출하여 소스/타겟의 해시 값을 비교한다.
  - 체크섬 알고리즘: 각 행의 모든 컬럼 값을 문자열로 직렬화 후 SHA-256 해시 비교.

- **검증 결과 리포트**:
  - JSON 형식의 검증 결과 파일(`validation_{job_id}.json`)을 생성한다.
  - Web UI에서 결과 단계(Step 4)에 검증 상태(Pass/Fail/Warning)를 테이블별로 표시한다.

- **WebSocket 프로토콜 확장**:
  ```json
  { "type": "validation_start", "table": "USERS" }
  { "type": "validation_result", "table": "USERS", "source_count": 50000, "target_count": 50000, "status": "pass" }
  { "type": "validation_result", "table": "ORDERS", "source_count": 120000, "target_count": 119998, "status": "mismatch", "detail": "2 rows missing" }
  ```

### 4.2. 구조화된 감사 로그 (Audit Trail)

- **마이그레이션 리포트 파일**:
  - 각 마이그레이션 수행 시 `.migration_state/{job_id}_report.json` 파일에 다음 정보를 기록한다:
    - `job_id`, 시작/종료 시각, 소스 DB 접속 정보(URL, 비밀번호 제외), 타겟 DB 종류 및 접속 정보
    - 테이블별: 행 수, 소요 시간(초), DDL 수행 여부, 오류 목록
    - 전체 요약: 총 행 수, 총 소요 시간, 성공/실패 테이블 수

- **Web UI 결과 대시보드 강화**:
  - Step 4(결과)에 테이블별 소요 시간, 처리 속도(rows/sec), 오류 수를 표 형태로 표시한다.
  - 리포트 JSON 다운로드 버튼을 추가한다.

- **CLI 출력**:
  - 마이그레이션 종료 시 요약 테이블을 터미널에 출력한다:
    ```
    ┌──────────┬────────┬──────────┬─────────┐
    │ Table    │ Rows   │ Duration │ Status  │
    ├──────────┼────────┼──────────┼─────────┤
    │ USERS    │ 50,000 │ 12.3s    │ OK      │
    │ ORDERS   │ 120K   │ 45.1s    │ OK      │
    │ LOGS     │ 2.1M   │ 5m23s    │ ERROR   │
    └──────────┴────────┴──────────┴─────────┘
    ```

### 4.3. 상세 에러 컨텍스트 (Rich Error Context)

- **구조화 에러 타입 도입**:
  ```go
  type MigrationError struct {
      Table      string
      Phase      string // "ddl", "data", "index", "constraint", "validation"
      BatchNum   int
      RowOffset  int
      Column     string // 문제 컬럼 (파악 가능한 경우)
      RootCause  error
      Suggestion string // 사용자에게 제안할 복구 방법
  }
  ```

- **에러 분류 체계**:
  - `TYPE_MISMATCH`: 데이터 타입 호환 불가 (예: Oracle CLOB → MySQL VARCHAR(255) 초과)
  - `NULL_VIOLATION`: NOT NULL 컬럼에 NULL 삽입 시도
  - `FK_VIOLATION`: 참조 무결성 위반
  - `CONNECTION_LOST`: 네트워크 단절
  - `TIMEOUT`: 쿼리 타임아웃
  - `PERMISSION_DENIED`: 권한 부족

- **WebSocket 에러 메시지 강화**:
  ```json
  {
    "type": "error",
    "table": "ORDERS",
    "error": "TYPE_MISMATCH: column DESCRIPTION (CLOB → VARCHAR(255)) exceeds max length at batch 42, row 41023",
    "suggestion": "Target column DESCRIPTION should be LONGTEXT or TEXT type",
    "phase": "data",
    "recoverable": true
  }
  ```

### 4.4. PostgreSQL COPY 모드 개선 — 배치 분할 COPY

- **배치 단위 COPY 전환**:
  - 기존: `CopyFrom`으로 전체 테이블을 단일 스트림 전송
  - 개선: `SELECT * FROM {table} ORDER BY ROWID OFFSET {n} ROWS FETCH NEXT {batch} ROWS ONLY` → 배치별 `CopyFrom` 실행
  - 각 배치 완료 시 `tracker.Update()` 호출 및 `MigrationState.UpdateOffset()` 저장

- **구현 전략**:
  - `--copy-batch` 플래그(기본값: 10000)로 COPY 배치 크기를 별도 제어한다.
  - 각 배치는 독립 트랜잭션으로 실행하되, 실패 시 해당 배치만 재시도한다.
  - 성능 저하를 최소화하기 위해 Oracle 측 `FETCH NEXT` 기반 페이징을 사용한다.

- **성능 벤치마크 기준**:
  - 배치 분할 COPY는 단일 COPY 대비 최대 20% 성능 저하를 허용한다.
  - 20% 이상 저하 시 기존 단일 COPY 모드를 `--copy-batch 0`으로 선택 가능하게 한다.

### 4.5. 입력 검증 및 보안 강화

- **테이블명 검증**:
  - Oracle 식별자 규칙에 따라 `^[A-Za-z_][A-Za-z0-9_$#]{0,127}$` 패턴으로 테이블명을 검증한다.
  - SQL 쿼리에서 테이블명 사용 시 반드시 `QuoteIdentifier()`를 통해 이스케이프한다.
  - `validateMigrationRequest()`에 테이블명 검증 로직을 추가한다.

- **쿼리 파라미터화**:
  - `fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)` → 식별자 검증 후 `QuoteIdentifier()` 적용
  - Oracle은 prepared statement에서 식별자를 파라미터로 받지 못하므로, 허용목록(allowlist) 기반 검증을 적용한다.

- **Oracle 식별자 Quoting 함수 추가**:
  - `dialect.QuoteOracleIdentifier(name string) string` 함수를 신설하여 Oracle 쿼리에서도 안전한 식별자 사용을 보장한다.

### 4.6. Dialect 코드 구조 개선

- **파일 분할 전략**:
  - 각 dialect 디렉토리를 하위 패키지로 분리하지 않고, 단일 패키지 내에서 파일만 분할한다:
    ```
    internal/dialect/
    ├── dialect.go              # 인터페이스 정의 (변경 없음)
    ├── postgres_ddl.go         # CreateTableDDL, CreateSequenceDDL, CreateIndexDDL, CreateConstraintDDL
    ├── postgres_insert.go      # InsertStatement, 값 직렬화
    ├── postgres_types.go       # MapOracleType, QuoteIdentifier, NormalizeURL
    ├── mysql_ddl.go
    ├── mysql_insert.go
    ├── mysql_types.go
    ├── mssql_ddl.go
    ├── mssql_insert.go
    ├── mssql_types.go
    ├── sqlite_ddl.go
    ├── sqlite_insert.go
    ├── sqlite_types.go
    ├── mariadb.go              # 얇은 래퍼이므로 단일 파일 유지
    └── *_test.go               # 기존 테스트 파일 유지 (분할 불필요)
    ```

- **분할 원칙**:
  - 구조체 정의와 `Name()`, `DriverName()` 등 메타 함수는 `*_types.go`에 배치
  - DDL 생성 로직은 `*_ddl.go`에 배치
  - INSERT/값 직렬화 로직은 `*_insert.go`에 배치
  - 기존 테스트는 변경하지 않음 (파일 분할은 리팩토링이지 기능 변경이 아님)

---

## 5. 아키텍처 및 인터페이스 변경 (Scope of Changes)

| 변경 영역 | 파일 / 모듈 | 상세 내용 |
|-----------|-------------|----------|
| **검증 엔진** | `migration/validation.go` (신규) | `ValidateTable()`: 소스-타겟 행 수 비교, 체크섬 샘플링 로직 |
| **감사 로그** | `migration/report.go` (신규) | `MigrationReport` 구조체, JSON 직렬화, CLI 테이블 출력 |
| **에러 타입** | `migration/errors.go` (신규) | `MigrationError` 구조화 에러, 에러 분류 상수 |
| **COPY 개선** | `migration/migration.go` | `MigrateTableDirect()` 내 PostgreSQL COPY를 배치 분할 루프로 교체 |
| **입력 검증** | `web/server.go`, `migration/migration.go` | 테이블명 검증 함수 추가, `QuoteOracleIdentifier()` 적용 |
| **Dialect 분할** | `dialect/*.go` | 기존 대형 파일을 `*_ddl.go`, `*_insert.go`, `*_types.go`로 분할 |
| **Config** | `config/config.go` | `--validate`, `--copy-batch` 플래그 추가 |
| **WebSocket** | `web/ws/tracker.go` | `ValidationStart()`, `ValidationResult()` 메서드 추가 |
| **Web UI** | `templates/index.html` | 검증 결과 표시, 리포트 다운로드, 상세 에러 패널 |
| **ProgressTracker** | `migration/migration.go` | `ValidationTracker` 인터페이스 추가 |

---

## 6. 비기능 요구사항 (Non-Functional Requirements)

- **성능**: 검증 단계는 마이그레이션 시간의 10% 이내로 완료되어야 한다. 체크섬 샘플링은 전수 검사가 아닌 통계적 샘플링이므로 대용량 테이블에서도 일정한 시간(최대 30초/테이블)을 유지한다.
- **하위 호환성**: `--validate`, `--copy-batch` 미지정 시 v8과 동일하게 동작한다. 기존 CLI/Web UI 워크플로우에 영향 없음.
- **보안**: 모든 사용자 입력 경로에 대한 식별자 검증을 적용하여, 알려진 SQL Injection 패턴을 차단한다.
- **코드 품질**: Dialect 파일 분할 후 기존 테스트가 100% 통과해야 한다. 새로운 기능에 대한 유닛 테스트 커버리지 80% 이상.

---

## 7. 마일스톤 (Milestones)

1. **1단계: 보안 강화** — 테이블명/스키마명 입력 검증, `QuoteOracleIdentifier()` 도입, SQL Injection 방어
2. **2단계: 에러 구조화** — `MigrationError` 타입 도입, 에러 분류 체계, WebSocket 에러 메시지 강화
3. **3단계: COPY 모드 개선** — 배치 분할 COPY, 진행률 표시, 체크포인트 연동
4. **4단계: 감사 로그 및 리포트** — `MigrationReport` 생성, CLI 요약 테이블, Web UI 결과 대시보드 강화
5. **5단계: 데이터 검증** — 행 수 비교, 체크섬 샘플링, 검증 결과 WebSocket 전송
6. **6단계: Dialect 리팩토링** — 대형 파일 분할, 기존 테스트 통과 확인
7. **7단계: 통합 테스트 및 QA** — 전체 시나리오 테스트, 성능 벤치마크, 보안 검증

---

## 8. 기대 효과 (Expected Outcomes)

v9 업데이트를 통해 **"마이그레이션을 수행했으나 결과를 신뢰할 수 없다"는 운영 현장의 핵심 불안 요소**를 제거합니다.

- **데이터 검증**: 자동화된 소스-타겟 정합성 검증으로 수동 검증 시간을 90% 이상 절감
- **감사 추적**: 구조화된 마이그레이션 리포트로 컴플라이언스 요건 충족 및 장애 시 원인 분석 시간 단축
- **에러 진단**: 상세 에러 컨텍스트로 디버깅 시간을 수십 분에서 수 분으로 단축
- **COPY 모드**: 대용량 PostgreSQL 마이그레이션에서도 실시간 진행률과 안정적 재개를 보장
- **보안**: SQL Injection 취약점 제거로 Web UI의 프로덕션 환경 배포 안전성 확보
- **유지보수**: 코드 구조 개선으로 향후 기능 추가 및 버그 수정 속도 향상


### spec.md

# Technical Specification - v9 (안정성·관측성·데이터 무결성 강화)

## 1. 입력 검증 및 보안 강화 (Security Hardening)

### 1.1. Oracle 식별자 검증 함수

`internal/dialect/` 패키지에 Oracle 소스 쿼리에서 사용하는 식별자 안전 함수를 추가한다.

```go
// internal/dialect/oracle.go (신규)

package dialect

import (
    "fmt"
    "regexp"
)

// oracleIdentifierPattern은 Oracle 식별자 규칙을 따르는 패턴이다.
// 알파벳/밑줄로 시작, 영숫자/밑줄/$/#만 허용, 최대 128자.
var oracleIdentifierPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_$#]{0,127}$`)

// ValidateOracleIdentifier는 문자열이 유효한 Oracle 식별자인지 검증한다.
func ValidateOracleIdentifier(name string) error {
    if !oracleIdentifierPattern.MatchString(name) {
        return fmt.Errorf("invalid Oracle identifier: %q", name)
    }
    return nil
}

// QuoteOracleIdentifier는 Oracle 식별자를 큰따옴표로 감싸 이스케이프한다.
// 내부 큰따옴표는 두 번 반복하여 이스케이프한다.
func QuoteOracleIdentifier(name string) string {
    escaped := strings.ReplaceAll(name, `"`, `""`)
    return fmt.Sprintf(`"%s"`, escaped)
}
```

### 1.2. 테이블명 검증 적용

**`internal/web/server.go`** — `validateMigrationRequest()`에 테이블명 검증을 추가한다:

```go
func validateMigrationRequest(req *startMigrationRequest) error {
    // 기존 검증 로직 유지...

    // 테이블명 검증 추가
    for _, table := range req.Tables {
        if err := dialect.ValidateOracleIdentifier(table); err != nil {
            return fmt.Errorf("invalid table name %q: %w", table, err)
        }
    }
    // OracleOwner 검증
    if req.OracleOwner != "" {
        if err := dialect.ValidateOracleIdentifier(req.OracleOwner); err != nil {
            return fmt.Errorf("invalid oracle owner %q: %w", req.OracleOwner, err)
        }
    }
    return nil
}
```

**`internal/migration/migration.go`** — 소스 쿼리에서 식별자 사용 시 `QuoteOracleIdentifier()` 적용:

```go
// 기존
query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
query := fmt.Sprintf("SELECT * FROM %s", tableName)

// 변경
quotedTable := dialect.QuoteOracleIdentifier(tableName)
query := fmt.Sprintf("SELECT COUNT(*) FROM %s", quotedTable)
query := fmt.Sprintf("SELECT * FROM %s", quotedTable)
```

적용 대상 함수:
- `Run()` — dry-run COUNT 쿼리 (migration.go:141)
- `MigrateTable()` — COUNT 쿼리 (migration.go:291, 300)
- `MigrateTableDirect()` — SELECT 쿼리 (migration.go:410)
- `MigrateTableToFile()` — SELECT 쿼리 (migration.go:687)

---

## 2. 구조화 에러 시스템 (Structured Error System)

### 2.1. 에러 타입 정의

`internal/migration/errors.go` (신규):

```go
package migration

import "fmt"

// ErrorCategory는 마이그레이션 에러의 분류 코드이다.
type ErrorCategory string

const (
    ErrTypeMismatch    ErrorCategory = "TYPE_MISMATCH"
    ErrNullViolation   ErrorCategory = "NULL_VIOLATION"
    ErrFKViolation     ErrorCategory = "FK_VIOLATION"
    ErrConnectionLost  ErrorCategory = "CONNECTION_LOST"
    ErrTimeout         ErrorCategory = "TIMEOUT"
    ErrPermissionDenied ErrorCategory = "PERMISSION_DENIED"
    ErrUnknown         ErrorCategory = "UNKNOWN"
)

// MigrationError는 마이그레이션 과정에서 발생하는 구조화된 에러이다.
type MigrationError struct {
    Table      string
    Phase      string        // "ddl", "data", "index", "constraint", "validation"
    Category   ErrorCategory
    BatchNum   int           // 1-based 배치 번호 (data phase에서만 유효)
    RowOffset  int           // 전체 행 기준 오프셋
    Column     string        // 문제 컬럼 (파악 가능한 경우)
    RootCause  error
    Suggestion string
    Recoverable bool
}

func (e *MigrationError) Error() string {
    msg := fmt.Sprintf("[%s] %s (table=%s, phase=%s", e.Category, e.RootCause, e.Table, e.Phase)
    if e.BatchNum > 0 {
        msg += fmt.Sprintf(", batch=%d", e.BatchNum)
    }
    if e.RowOffset > 0 {
        msg += fmt.Sprintf(", row=%d", e.RowOffset)
    }
    if e.Column != "" {
        msg += fmt.Sprintf(", column=%s", e.Column)
    }
    msg += ")"
    return msg
}

func (e *MigrationError) Unwrap() error {
    return e.RootCause
}

// classifyError는 DB 드라이버 에러 메시지를 분석하여 ErrorCategory를 결정한다.
func classifyError(err error) ErrorCategory {
    msg := err.Error()
    switch {
    case containsAny(msg, "data type", "type mismatch", "incompatible", "too long", "overflow"):
        return ErrTypeMismatch
    case containsAny(msg, "null", "NOT NULL", "cannot insert NULL"):
        return ErrNullViolation
    case containsAny(msg, "foreign key", "referential", "REFERENCES"):
        return ErrFKViolation
    case containsAny(msg, "connection", "reset", "broken pipe", "EOF", "refused"):
        return ErrConnectionLost
    case containsAny(msg, "timeout", "deadline"):
        return ErrTimeout
    case containsAny(msg, "permission", "denied", "privilege", "ORA-01031"):
        return ErrPermissionDenied
    default:
        return ErrUnknown
    }
}
```

### 2.2. MigrateTableDirect에 에러 컨텍스트 적용

`internal/migration/migration.go` — 배치 INSERT 실패 시:

```go
// 기존
return fmt.Errorf("failed to execute batch insert: %v\nstmt: %s", err, stmt)

// 변경
return &MigrationError{
    Table:       tableName,
    Phase:       "data",
    Category:    classifyError(err),
    BatchNum:    batchNum,
    RowOffset:   rowCount,
    RootCause:   err,
    Suggestion:  suggestFix(classifyError(err), dia.Name()),
    Recoverable: classifyError(err) != ErrConnectionLost,
}
```

DDL 실패, 인덱스 실패, 제약조건 실패에도 동일한 패턴을 적용한다. Phase 값은 각각 `"ddl"`, `"index"`, `"constraint"`로 설정한다.

### 2.3. WebSocket 에러 메시지 확장

**`internal/web/ws/tracker.go`** — `ProgressMsg` 구조체 확장:

```go
type ProgressMsg struct {
    // 기존 필드 유지...
    Type         MsgType `json:"type"`
    Table        string  `json:"table,omitempty"`
    Count        int     `json:"count,omitempty"`
    Total        int     `json:"total,omitempty"`
    ErrorMsg     string  `json:"error,omitempty"`
    Message      string  `json:"message,omitempty"`
    ZipFileID    string  `json:"zip_file_id,omitempty"`
    ConnectionOk bool    `json:"connection_ok,omitempty"`
    Object       string  `json:"object,omitempty"`
    ObjectName   string  `json:"object_name,omitempty"`
    Status       string  `json:"status,omitempty"`

    // v9 추가 필드
    Phase       string `json:"phase,omitempty"`
    Category    string `json:"category,omitempty"`
    Suggestion  string `json:"suggestion,omitempty"`
    Recoverable *bool  `json:"recoverable,omitempty"`
}
```

**`WebSocketTracker.Error()` 메서드 확장** — `MigrationError` 타입 체크 후 상세 필드 전송:

```go
func (t *WebSocketTracker) Error(table string, err error) {
    t.mu.Lock()
    delete(t.states, table)
    t.mu.Unlock()

    msg := ProgressMsg{
        Type:     MsgError,
        Table:    table,
        ErrorMsg: err.Error(),
    }

    // MigrationError인 경우 상세 필드 추가
    var migErr *MigrationError
    if errors.As(err, &migErr) {
        msg.Phase = migErr.Phase
        msg.Category = string(migErr.Category)
        msg.Suggestion = migErr.Suggestion
        recoverable := migErr.Recoverable
        msg.Recoverable = &recoverable
    }

    t.broadcast(msg)
}
```

> **주의**: `ws` 패키지에서 `migration.MigrationError`를 직접 임포트하면 순환 의존이 발생한다. 이를 방지하기 위해 `Error()` 메서드는 인터페이스 기반으로 상세 필드를 추출한다:

```go
// internal/migration/errors.go에 인터페이스 추가
type DetailedError interface {
    error
    ErrorPhase() string
    ErrorCategory() string
    ErrorSuggestion() string
    IsRecoverable() bool
}

func (e *MigrationError) ErrorPhase() string      { return e.Phase }
func (e *MigrationError) ErrorCategory() string    { return string(e.Category) }
func (e *MigrationError) ErrorSuggestion() string  { return e.Suggestion }
func (e *MigrationError) IsRecoverable() bool      { return e.Recoverable }
```

```go
// ws/tracker.go에서는 인터페이스로 접근
type DetailedError interface {
    ErrorPhase() string
    ErrorCategory() string
    ErrorSuggestion() string
    IsRecoverable() bool
}

func (t *WebSocketTracker) Error(table string, err error) {
    // ...
    if de, ok := err.(DetailedError); ok {
        msg.Phase = de.ErrorPhase()
        msg.Category = de.ErrorCategory()
        msg.Suggestion = de.ErrorSuggestion()
        recoverable := de.IsRecoverable()
        msg.Recoverable = &recoverable
    }
    t.broadcast(msg)
}
```

---

## 3. PostgreSQL COPY 모드 개선 (Batched COPY)

### 3.1. Config 플래그 추가

**`internal/config/config.go`**:

```go
type Config struct {
    // 기존 필드 유지...

    // v9 flags
    Validate  bool
    CopyBatch int  // 0이면 기존 단일 COPY 모드 유지
}
```

```go
// ParseFlags()에 추가
flag.BoolVar(&cfg.Validate, "validate", false, "마이그레이션 후 소스-타겟 데이터 검증 수행")
flag.IntVar(&cfg.CopyBatch, "copy-batch", 10000, "PostgreSQL COPY 모드 배치 크기 (0: 단일 COPY)")
```

### 3.2. MigrateTableDirect PostgreSQL 경로 변경

`internal/migration/migration.go` — `MigrateTableDirect()` 내 `pgPool != nil` 분기:

```go
if pgPool != nil {
    if cfg.CopyBatch <= 0 {
        // 기존 단일 COPY 모드 (v8 동작 유지)
        // ... 기존 코드 그대로 ...
    } else {
        // v9: 배치 분할 COPY 모드
        err = migrateTablePgBatchCopy(dbConn, pgPool, tableName, cfg, tracker, mState)
    }
}
```

**신규 함수 `migrateTablePgBatchCopy()`**:

```go
func migrateTablePgBatchCopy(
    dbConn *sql.DB,
    pgPool db.PGPool,
    tableName string,
    cfg *config.Config,
    tracker ProgressTracker,
    mState *MigrationState,
) error {
    tState := mState.GetState(tableName)
    offset := tState.Offset
    batchSize := cfg.CopyBatch
    quotedTable := dialect.QuoteOracleIdentifier(tableName)

    for {
        // Oracle에서 배치 단위 조회
        query := fmt.Sprintf(
            "SELECT * FROM %s OFFSET %d ROWS FETCH NEXT %d ROWS ONLY",
            quotedTable, offset, batchSize,
        )
        rows, err := dbConn.Query(query)
        if err != nil {
            return &MigrationError{
                Table: tableName, Phase: "data",
                Category: classifyError(err), RootCause: err,
            }
        }

        cols, _ := rows.Columns()
        source := &oracleCopySource{rows: rows, cols: cols}

        // 배치별 독립 트랜잭션
        ctx := context.Background()
        tx, err := pgPool.Begin(ctx)
        if err != nil {
            rows.Close()
            return &MigrationError{
                Table: tableName, Phase: "data",
                Category: ErrConnectionLost, RootCause: err,
            }
        }

        n, err := tx.CopyFrom(ctx, pgx.Identifier{cfg.Schema, tableName}, cols, source)
        rows.Close()

        if err != nil {
            tx.Rollback(ctx)
            return &MigrationError{
                Table: tableName, Phase: "data",
                Category: classifyError(err), RootCause: err,
                RowOffset: offset, BatchNum: (offset / batchSize) + 1,
            }
        }

        if err := tx.Commit(ctx); err != nil {
            return &MigrationError{
                Table: tableName, Phase: "data",
                Category: classifyError(err), RootCause: err,
            }
        }

        offset += int(n)
        mState.UpdateOffset(tableName, offset)

        if tracker != nil {
            tracker.Update(tableName, offset)
        }

        // n < batchSize이면 마지막 배치 — 루프 종료
        if int(n) < batchSize {
            break
        }
    }

    slog.Info("batched COPY migration finished", "table", tableName, "rows", offset)
    return nil
}
```

### 3.3. Web UI 연동

**`internal/web/server.go`** — `startMigrationRequest`에 필드 추가:

```go
type startMigrationRequest struct {
    // 기존 필드 유지...

    // v9 추가 필드
    Validate  bool `json:"validate"`
    CopyBatch int  `json:"copyBatch"`
}
```

Config 생성 시 매핑:
```go
cfg := &config.Config{
    // 기존 필드 유지...
    Validate:  req.Validate,
    CopyBatch: req.CopyBatch,
}
```

---

## 4. 감사 로그 및 마이그레이션 리포트 (Audit & Report)

### 4.1. 리포트 구조체

`internal/migration/report.go` (신규):

```go
package migration

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "sync"
    "time"
)

type TableReport struct {
    Name          string        `json:"name"`
    RowCount      int           `json:"row_count"`
    Duration      time.Duration `json:"duration_ns"`
    DurationHuman string        `json:"duration"`
    DDLExecuted   bool          `json:"ddl_executed"`
    Status        string        `json:"status"` // "ok", "error", "skipped"
    Errors        []string      `json:"errors,omitempty"`
}

type MigrationReport struct {
    JobID        string            `json:"job_id"`
    StartedAt    time.Time         `json:"started_at"`
    FinishedAt   time.Time         `json:"finished_at"`
    DurationHuman string           `json:"duration"`
    SourceURL    string            `json:"source_url"`  // 비밀번호 마스킹
    TargetDB     string            `json:"target_db"`
    TargetURL    string            `json:"target_url"`  // 비밀번호 마스킹
    Tables       []TableReport     `json:"tables"`
    TotalRows    int               `json:"total_rows"`
    SuccessCount int               `json:"success_count"`
    ErrorCount   int               `json:"error_count"`
    mu           sync.Mutex
}

func NewMigrationReport(jobID, sourceURL, targetDB, targetURL string) *MigrationReport {
    return &MigrationReport{
        JobID:     jobID,
        StartedAt: time.Now(),
        SourceURL: maskPassword(sourceURL),
        TargetDB:  targetDB,
        TargetURL: maskPassword(targetURL),
    }
}

// StartTable은 테이블 마이그레이션 시작을 기록하고 종료 시 호출할 콜백을 반환한다.
func (r *MigrationReport) StartTable(name string, withDDL bool) func(rowCount int, err error) {
    start := time.Now()
    return func(rowCount int, err error) {
        elapsed := time.Since(start)
        tr := TableReport{
            Name:          name,
            RowCount:      rowCount,
            Duration:      elapsed,
            DurationHuman: formatDuration(elapsed),
            DDLExecuted:   withDDL,
        }
        if err != nil {
            tr.Status = "error"
            tr.Errors = append(tr.Errors, err.Error())
        } else {
            tr.Status = "ok"
        }

        r.mu.Lock()
        r.Tables = append(r.Tables, tr)
        r.TotalRows += rowCount
        if err != nil {
            r.ErrorCount++
        } else {
            r.SuccessCount++
        }
        r.mu.Unlock()
    }
}

// Finalize는 리포트를 마무리하고 JSON 파일로 저장한다.
func (r *MigrationReport) Finalize() error {
    r.FinishedAt = time.Now()
    r.DurationHuman = formatDuration(r.FinishedAt.Sub(r.StartedAt))

    dir := ".migration_state"
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }

    path := filepath.Join(dir, fmt.Sprintf("%s_report.json", r.JobID))
    data, err := json.MarshalIndent(r, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}

// PrintSummary는 CLI에서 요약 테이블을 표준 출력에 출력한다.
func (r *MigrationReport) PrintSummary() {
    // 최대 컬럼 너비 계산
    maxName := 5 // "Table" 헤더 최소 길이
    for _, t := range r.Tables {
        if len(t.Name) > maxName {
            maxName = len(t.Name)
        }
    }

    border := fmt.Sprintf("┌─%s─┬──────────┬──────────┬─────────┐", strings.Repeat("─", maxName))
    header := fmt.Sprintf("│ %-*s │ %-8s │ %-8s │ %-7s │", maxName, "Table", "Rows", "Duration", "Status")
    sep    := fmt.Sprintf("├─%s─┼──────────┼──────────┼─────────┤", strings.Repeat("─", maxName))
    footer := fmt.Sprintf("└─%s─┴──────────┴──────────┴─────────┘", strings.Repeat("─", maxName))

    fmt.Println(border)
    fmt.Println(header)
    fmt.Println(sep)
    for _, t := range r.Tables {
        status := t.Status
        if status == "ok" {
            status = "OK"
        } else {
            status = "ERROR"
        }
        fmt.Printf("│ %-*s │ %8s │ %8s │ %-7s │\n",
            maxName, t.Name, formatCount(t.RowCount), t.DurationHuman, status)
    }
    fmt.Println(footer)
    fmt.Printf("Total: %s rows, %d ok, %d errors, %s elapsed\n",
        formatCount(r.TotalRows), r.SuccessCount, r.ErrorCount, r.DurationHuman)
}

// maskPassword는 URL에서 비밀번호를 "***"로 치환한다.
func maskPassword(url string) string {
    // "user:password@" 패턴에서 password를 마스킹
    // oracle://user:pass@host → oracle://user:***@host
    // postgres://user:pass@host → postgres://user:***@host
    // 간단한 패턴 매칭으로 처리
    // ... 구현 ...
}

func formatDuration(d time.Duration) string { /* ... */ }
func formatCount(n int) string              { /* ... */ }
```

### 4.2. Run() 함수에 리포트 통합

`internal/migration/migration.go` — `Run()` 함수 시작부에 리포트 생성, 종료부에 finalize:

```go
func Run(dbConn *sql.DB, targetDB *sql.DB, pgPool db.PGPool, dia dialect.Dialect, cfg *config.Config, tracker ProgressTracker) error {
    // ... 기존 초기화 ...

    report := NewMigrationReport(jobID, cfg.OracleURL, cfg.TargetDB, cfg.TargetURL)

    // ... dry-run 분기 (리포트 없이 기존 동작) ...

    // worker 호출 시 report 전달
    go worker(w, dbConn, targetDB, pgPool, dia, jobs, &wg, mainBuf, cfg, &outMutex, tracker, mState, report)

    // ... wg.Wait() ...
    // ... constraint post-processing ...

    // 검증 단계 (cfg.Validate일 때만)
    if cfg.Validate && (pgPool != nil || targetDB != nil) {
        runValidation(dbConn, targetDB, pgPool, dia, cfg, tracker, report)
    }

    // 리포트 저장 및 출력
    report.Finalize()
    report.PrintSummary()

    return nil
}
```

**worker 함수에서 리포트 기록**:

```go
func worker(id int, ..., report *MigrationReport) {
    defer wg.Done()
    for j := range jobs {
        finishTable := report.StartTable(j.tableName, cfg.WithDDL)
        err := MigrateTable(...)
        var rowCount int
        // rowCount는 MigrateTable의 반환값으로 변경 필요 (아래 참고)
        finishTable(rowCount, err)
        // ... 기존 에러 핸들링 ...
    }
}
```

> **MigrateTable 시그니처 변경**: 리포트 기록을 위해 처리된 행 수를 반환해야 한다.
> ```go
> // 기존
> func MigrateTable(...) error
> // 변경
> func MigrateTable(...) (int, error)
> ```
> 반환값: `(rowCount, nil)` 또는 `(partialRowCount, err)`

### 4.3. Web UI 리포트 다운로드

**`internal/web/server.go`** — 새 엔드포인트:

```go
api.GET("/download/report/:id", downloadReport)
```

```go
func downloadReport(c *gin.Context) {
    id := filepath.Base(c.Param("id"))
    reportPath := filepath.Join(".migration_state", id+"_report.json")
    if _, err := os.Stat(reportPath); os.IsNotExist(err) {
        c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
        return
    }
    c.Header("Content-Disposition", "attachment; filename="+id+"_report.json")
    c.Header("Content-Type", "application/json")
    c.File(reportPath)
}
```

**WebSocket 프로토콜** — `all_done` 메시지에 리포트 요약 포함:

```go
// tracker.AllDone 확장
func (t *WebSocketTracker) AllDone(zipFileID string, report *ReportSummary) {
    msg := ProgressMsg{
        Type:      MsgAllDone,
        ZipFileID: zipFileID,
    }
    if report != nil {
        msg.ReportSummary = report
    }
    t.broadcast(msg)
}
```

`ProgressMsg`에 추가:
```go
type ReportSummary struct {
    TotalRows    int    `json:"total_rows"`
    SuccessCount int    `json:"success_count"`
    ErrorCount   int    `json:"error_count"`
    Duration     string `json:"duration"`
    ReportID     string `json:"report_id"`
}

type ProgressMsg struct {
    // 기존 필드...
    ReportSummary *ReportSummary `json:"report_summary,omitempty"`
}
```

> **순환 의존 방지**: `ReportSummary`는 `ws` 패키지에 정의하고, `migration` 패키지의 `MigrationReport`에서 `ToSummary()` 메서드로 변환한다. 또는 `ws` 패키지에 독립적인 DTO를 두고, `server.go`에서 변환한다.

---

## 5. 마이그레이션 후 데이터 검증 (Post-Migration Validation)

### 5.1. ValidationTracker 인터페이스

`internal/migration/migration.go`에 추가:

```go
type ValidationTracker interface {
    ValidationStart(table string)
    ValidationResult(table string, sourceCount, targetCount int, status string, detail string)
}
```

### 5.2. 검증 엔진

`internal/migration/validation.go` (신규):

```go
package migration

import (
    "context"
    "crypto/sha256"
    "database/sql"
    "fmt"
    "log/slog"
    "strings"

    "dbmigrator/internal/config"
    "dbmigrator/internal/db"
    "dbmigrator/internal/dialect"
)

// runValidation은 직접 마이그레이션 후 소스-타겟 데이터를 비교 검증한다.
func runValidation(
    dbConn *sql.DB,
    targetDB *sql.DB,
    pgPool db.PGPool,
    dia dialect.Dialect,
    cfg *config.Config,
    tracker ProgressTracker,
    report *MigrationReport,
) {
    valTracker, hasValTracker := tracker.(ValidationTracker)

    for _, tableName := range cfg.Tables {
        if hasValTracker {
            valTracker.ValidationStart(tableName)
        }

        result := validateTable(dbConn, targetDB, pgPool, dia, tableName, cfg)

        if hasValTracker {
            valTracker.ValidationResult(
                tableName, result.SourceCount, result.TargetCount,
                result.Status, result.Detail,
            )
        }

        slog.Info("validation result",
            "table", tableName,
            "source_count", result.SourceCount,
            "target_count", result.TargetCount,
            "status", result.Status,
        )
    }
}

type ValidationResult struct {
    Table       string `json:"table"`
    SourceCount int    `json:"source_count"`
    TargetCount int    `json:"target_count"`
    Status      string `json:"status"` // "pass", "mismatch", "error"
    Detail      string `json:"detail,omitempty"`
}

func validateTable(
    dbConn *sql.DB,
    targetDB *sql.DB,
    pgPool db.PGPool,
    dia dialect.Dialect,
    tableName string,
    cfg *config.Config,
) ValidationResult {
    result := ValidationResult{Table: tableName}
    quotedSrc := dialect.QuoteOracleIdentifier(tableName)

    // 1. 소스 행 수 조회
    err := dbConn.QueryRow(
        fmt.Sprintf("SELECT COUNT(*) FROM %s", quotedSrc),
    ).Scan(&result.SourceCount)
    if err != nil {
        result.Status = "error"
        result.Detail = "source count query failed: " + err.Error()
        return result
    }

    // 2. 타겟 행 수 조회
    targetTable := dia.QuoteIdentifier(strings.ToLower(tableName))
    if cfg.Schema != "" {
        targetTable = dia.QuoteIdentifier(strings.ToLower(cfg.Schema)) + "." + targetTable
    }

    if pgPool != nil {
        err = pgPool.QueryRow(
            context.Background(),
            fmt.Sprintf("SELECT COUNT(*) FROM %s", targetTable),
        ).Scan(&result.TargetCount)
    } else if targetDB != nil {
        err = targetDB.QueryRow(
            fmt.Sprintf("SELECT COUNT(*) FROM %s", targetTable),
        ).Scan(&result.TargetCount)
    }
    if err != nil {
        result.Status = "error"
        result.Detail = "target count query failed: " + err.Error()
        return result
    }

    // 3. 비교
    if result.SourceCount != result.TargetCount {
        result.Status = "mismatch"
        diff := result.SourceCount - result.TargetCount
        result.Detail = fmt.Sprintf("%d rows difference", diff)
    } else {
        result.Status = "pass"
    }

    return result
}
```

### 5.3. WebSocket Tracker 확장

**`internal/web/ws/tracker.go`** — 새 메시지 타입 및 메서드:

```go
const (
    // 기존 상수 유지...
    MsgValidationStart  MsgType = "validation_start"
    MsgValidationResult MsgType = "validation_result"
)
```

```go
func (t *WebSocketTracker) ValidationStart(table string) {
    t.broadcast(ProgressMsg{
        Type:  MsgValidationStart,
        Table: table,
    })
}

func (t *WebSocketTracker) ValidationResult(table string, sourceCount, targetCount int, status, detail string) {
    t.broadcast(ProgressMsg{
        Type:    MsgValidationResult,
        Table:   table,
        Total:   sourceCount,          // source_count
        Count:   targetCount,          // target_count
        Status:  status,
        Message: detail,
    })
}
```

### 5.4. Web UI 결과 표시

**`templates/index.html`** — Step 4(결과) 영역에 검증 결과 테이블 추가:

```html
<div id="validation-results" class="validation-panel" style="display:none;">
    <h3>데이터 검증 결과</h3>
    <table class="validation-table">
        <thead>
            <tr>
                <th>테이블</th>
                <th>소스 행 수</th>
                <th>타겟 행 수</th>
                <th>상태</th>
                <th>상세</th>
            </tr>
        </thead>
        <tbody id="validation-tbody"></tbody>
    </table>
</div>
```

JavaScript WebSocket 핸들러에 추가:

```javascript
case 'validation_start':
    // 검증 패널 표시, 해당 테이블 행에 스피너 추가
    break;
case 'validation_result':
    // 결과 행 업데이트: pass → 녹색 체크, mismatch → 주황색 경고, error → 빨간색
    break;
```

---

## 6. Dialect 코드 구조 개선 (Refactoring)

### 6.1. 분할 전략

각 dialect 파일을 역할별 3개 파일로 분할한다. **패키지는 변경하지 않는다** (`package dialect` 유지).

| 현재 파일 | 분할 후 |
|-----------|---------|
| `postgres.go` (7,522줄) | `postgres_types.go` — 구조체, `Name()`, `DriverName()`, `QuoteIdentifier()`, `NormalizeURL()`, `MapOracleType()` |
| | `postgres_ddl.go` — `CreateTableDDL()`, `CreateSequenceDDL()`, `CreateIndexDDL()`, `CreateConstraintDDL()` |
| | `postgres_insert.go` — `InsertStatement()`, 값 직렬화 헬퍼 |
| `mysql.go` (7,161줄) | `mysql_types.go`, `mysql_ddl.go`, `mysql_insert.go` (동일 패턴) |
| `mssql.go` (8,650줄) | `mssql_types.go`, `mssql_ddl.go`, `mssql_insert.go` (동일 패턴) |
| `sqlite.go` (5,079줄) | `sqlite_types.go`, `sqlite_ddl.go`, `sqlite_insert.go` (동일 패턴) |
| `mariadb.go` (199줄) | 분할 없음 (얇은 래퍼) |

### 6.2. 분할 절차

1. **기존 테스트 전체 실행** — `go test ./internal/dialect/...` 결과 저장 (기준선)
2. **파일 분할 실행** — 구조체/메서드를 새 파일로 이동, `package dialect` 유지
3. **테스트 재실행** — 기준선과 동일한 결과 확인
4. **컴파일 확인** — `go build ./...` 성공 확인

### 6.3. 분할 원칙

- **공개 API 변경 없음**: `Dialect` 인터페이스, 구조체 이름, 메서드 시그니처 변경 없음
- **내부 헬퍼 함수**: 파일 간 공유가 필요한 경우 동일 패키지이므로 그대로 호출 가능
- **import 변경 없음**: 외부 패키지에서 `dialect.PostgresDialect`로 접근하는 코드 영향 없음

---

## 7. Web UI 변경 요약

### 7.1. Step 2(설정) 추가 항목

| 항목 | 위치 | UI 요소 | 기본값 |
|------|------|---------|--------|
| 마이그레이션 후 검증 | DDL Options 하위 | 체크박스 "Validate after migration" | 미체크 |
| COPY 배치 크기 | Advanced Settings 하위 | 숫자 입력 "COPY batch size" | 10000 |

### 7.2. Step 4(결과) 강화

- **결과 요약 대시보드**: 총 행 수, 소요 시간, 성공/실패 테이블 수, 처리 속도(rows/sec)
- **테이블별 상세**: 각 테이블의 행 수, 소요 시간, 상태를 표 형태로 표시
- **검증 결과 패널**: `--validate` 활성화 시 소스-타겟 행 수 비교 결과 표시
- **리포트 다운로드**: "리포트 JSON 다운로드" 버튼 추가
- **상세 에러 패널**: 에러 발생 시 phase, category, suggestion을 펼침(expandable) 형태로 표시

### 7.3. 에러 메시지 표시 개선

기존 단순 에러 텍스트 → 구조화된 에러 카드:

```
┌ ERROR: ORDERS ────────────────────────────┐
│ Phase: data                                │
│ Category: TYPE_MISMATCH                    │
│ Batch #42, Row offset 41023                │
│                                            │
│ column DESCRIPTION (CLOB → VARCHAR(255))   │
│ exceeds max length                         │
│                                            │
│ 💡 Suggestion: Target column DESCRIPTION   │
│    should be LONGTEXT or TEXT type          │
└────────────────────────────────────────────┘
```

---

## 8. 테스트 전략

### 8.1. 단위 테스트 (신규)

| 파일 | 테스트 내용 |
|------|------------|
| `dialect/oracle_test.go` | `ValidateOracleIdentifier()` — 유효/무효 식별자 패턴 테스트 |
| `migration/errors_test.go` | `MigrationError.Error()` 포맷, `classifyError()` 분류 정확성 |
| `migration/report_test.go` | `MigrationReport` 생성, `StartTable` → 콜백 → `Finalize()`, JSON 직렬화 검증 |
| `migration/validation_test.go` | `validateTable()` — mock DB로 행 수 일치/불일치 시나리오 |

### 8.2. 통합 테스트

| 시나리오 | 검증 항목 |
|---------|----------|
| 배치 분할 COPY | 10,000건 테이블을 batch=2000으로 분할 시 5개 배치로 완료되는지 확인, 중간 체크포인트 저장 확인 |
| 에러 컨텍스트 전파 | 타입 불일치를 유발하는 데이터 삽입 시 `MigrationError`의 필드가 올바르게 채워지는지 확인 |
| 검증 행 수 불일치 | 소스 100건, 타겟에서 일부 삭제 후 검증 시 `mismatch` 상태 반환 확인 |
| SQL Injection 차단 | 테이블명에 `"; DROP TABLE --` 전달 시 `ValidateOracleIdentifier()` 거부 확인 |
| 리포트 생성 | 마이그레이션 완료 후 `.migration_state/{job_id}_report.json` 파일이 올바른 구조로 생성되는지 확인 |
| Dialect 분할 회귀 | 파일 분할 후 기존 전체 테스트 스위트 통과 확인 |

### 8.3. Web UI 수동 테스트

| 시나리오 | 확인 항목 |
|---------|----------|
| 검증 체크박스 활성화 | Step 4에서 검증 결과 테이블 표시 여부 |
| 에러 발생 | 구조화된 에러 카드 렌더링 (phase, category, suggestion) |
| 리포트 다운로드 | JSON 파일 다운로드 및 내용 확인 |
| COPY 배치 설정 | Advanced Settings에서 값 변경 후 progress bar 점진 업데이트 확인 |


### tasks.md

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


---
## <a name="v10"></a> v10

### prd.md

# Go DB Migration v10 UI 개선 PRD (Product Requirements Document)

## 1. 개요 (Overview)
현재 Go DB Migration 도구의 웹 UI는 단계별(Stepper) 마법사 형태를 갖추고 있으나, 기능이 추가됨에 따라 설정 화면(Step 2)이 비대해지고 복잡해졌습니다. 또한 대량의 테이블 마이그레이션 시 진행 상황(Step 3)을 모니터링하기 어렵다는 한계가 있습니다.
v10의 목표는 사용자 경험(UX)을 심층적으로 개선하여 **직관적이고 안정적인 대규모 데이터 마이그레이션 모니터링 및 제어 환경**을 제공하는 것입니다.

## 2. 현재 UI 문제점 분석
* **정보 밀집도 과다 (Step 2)**: 좌측 테이블 선택과 우측 설정이 한 화면에 혼재되어 복잡. 특히 우측의 설정(출력, 타겟 DB, DDL, 고급 설정 등)이 길게 나열되어 초보자가 접근하기 어려움.
* **타겟 DB 검증 부재**: Oracle 소스 DB는 연결 시 바로 검증(Step 1)되지만, 타겟 DB(PostgreSQL, MySQL 등)는 설정을 마치고 마이그레이션을 시작해야만 연결 오류를 알 수 있음.
* **대규모 테이블 모니터링 한계 (Step 3)**: 100개 이상의 테이블 마이그레이션 시, 화면에 수많은 진행률 바가 나열되어 가독성이 떨어짐. 실패/지연 중인 테이블을 한눈에 파악하기 힘듦.
* **피드백 부족**: 개별 테이블의 스키마 구조나 예상 데이터 크기를 설정 단계(Step 2)에서 미리 파악할 수 없어 배치 크기(Batch Size)나 워커 수를 최적화하기 어려움.
* **상태 전환의 불편함**: 마이그레이션 중 에러 발생 시, 설정을 일부 수정하고 다시 시도하려면 1단계부터 초기화되거나 새로고침을 해야 함.

## 3. 개선 목표 및 핵심 UX
1. **Progressive Disclosure (점진적 정보 노출)**: 기본 설정과 고급 설정을 분리하여 UI 복잡도 감소. (고급 설정은 Accordion이나 Modal로 숨겨 화면을 깔끔하게 유지)
2. **사전 검증(Pre-flight Check) 강화**: 마이그레이션 실행 전 Source와 Target DB 간의 연결 상태 및 스키마 호환성, 권한을 미리 테스트하는 기능 도입.
3. **대시보드형 모니터링**: 대량의 테이블 처리 상태를 요약 대시보드(전체 진행률, 성공/실패 수량, 속도 ETA 등)로 먼저 보여주고, 필터링/페이징 가능한 테이블 리스트로 세부 진행 상태 제공.
4. **Resilience (복원력) 및 편의성**: 오류 발생 시 전체 중단이 아닌, 실패한 테이블만 재시도(Retry)할 수 있는 인터페이스 제공.

## 4. 상세 화면 및 기능 명세

### 4.1. Step 1: Source & Target Connection (연결 단계 통합)
* **변경점**: 기존 Step 1(Oracle만 연결)과 Step 2의 타겟 DB 설정 일부를 통합하여 **Connection Phase**로 개편.
* **기능**:
  * Source DB (Oracle) 정보 입력 및 `[연결 테스트 및 테이블 로드]` 버튼 제공.
  * Target DB (PostgreSQL, MySQL 등) 정보 입력 및 `[연결 테스트]` 버튼 제공.
  * Target DB 검증이 성공적으로 완료되어야 다음 단계 이동 시 안정성 확보.

### 4.2. Step 2: Configuration & Selection (설정 및 테이블 선택)
* **테이블 그리드 고도화**:
  * 단순 체크박스 리스트가 아닌 **Data Table 형태**로 제공.
  * 테이블명 외에 컬럼 수, 예상 로우(Row) 수 등의 메타데이터 추가 표시(가능한 경우).
  * 검색, 정렬, 전체 선택/해제 기능 고도화.
* **설정 UI 최적화**:
  * **기본 설정**: 마이그레이션 모드(Direct/SQL 파일), DDL 생성 여부 등 필수 옵션만 메인 뷰에 노출.
  * **고급 설정 (Advanced)**: 배치 크기, 병렬 워커 수, COPY 배치 크기, DB Connection Pool 등의 옵션은 "고급 설정(Advanced Settings)" 토글 패널 내부에 배치.

### 4.3. Step 3: Execution & Monitoring (실행 및 모니터링)
* **Pre-flight Check (사전 점검 확인 창)**:
  * 마이그레이션 시작 전 모달 팝업으로 선택된 테이블 수, 대상 타겟 DB, 주의 사항을 요약해서 보여주고 최종 `[Start Migration]` 트리거.
* **모니터링 대시보드 개편**:
  * **상단 요약 (Summary Widget)**: 전체 진행률(Overall Progress), 남은 시간(ETA), 처리 속도(Rows/sec) 실시간 갱신.
  * **상태별 탭 (Tabs)**: `전체(100)` | `대기(50)` | `진행중(4)` | `완료(40)` | `에러(6)` 형태로 필터 탭 제공. 대량의 테이블이 나열될 때 원하는 상태만 쉽게 필터링.
  * **에러 로그 및 복구**: 에러가 난 테이블은 별도 강조 표시되며, 클릭 시 상세 로그(에러 원인, Phase 등)를 보여주는 사이드 패널 또는 아코디언 노출.
  * 기능상 가능하다면 개별 실패 테이블에 대한 `[재시도(Retry)]` 버튼 제공.

## 5. 비기능적 요구사항 (Non-functional Requirements)
* **렌더링 성능 최적화 (Virtualization)**: 1,000개 이상의 테이블 목록 렌더링 및 프로그레스 바 실시간 업데이트 시 DOM 과부하를 방지하기 위해 Virtual Scrolling 기법 검토 및 적용.
* **다크 모드 및 접근성 보완**: 테마 대비를 높이고 텍스트 가독성을 최적화하여 시인성 개선.
* **세션 상태 유지**: 우발적인 새로고침이나 탭 닫힘을 대비해 LocalStorage에 입력 폼 데이터(비밀번호 제외) 및 진행 상태 토큰 임시 저장.

## 6. 마일스톤 및 작업 분류
* **Phase 1: 기획 및 구조 개편** - 타겟 DB 설정 1단계 이동 및 UI 레이아웃 초안(Wireframe) 작성.
* **Phase 2: UI/UX 컴포넌트 재구성** - 고급 설정 분리, 테이블 선택 화면 Data Table화 적용.
* **Phase 3: 모니터링 대시보드 구현** - 상태별 탭, ETA 계산 로직 추가, 가상 스크롤(Virtual List) 적용 여부 결정.
* **Phase 4: 사용자 피드백 반영 및 안정화** - 재시도(Retry) 로직 연결, 예외 상황 처리 고도화 및 최종 버그 수정.


### spec.md

# Go DB Migration v10 UI 개선 기능 명세서 (Technical Specification)

## 1. 개요
본 문서는 `docs/v10/prd.md`에 정의된 요구사항을 바탕으로, Go DB Migration 도구의 프론트엔드(UI/UX) 및 백엔드(API) 변경 사항을 정의하는 기술 명세서입니다. 핵심 목표는 설정 화면의 복잡도를 낮추고, 대규모 데이터 이관 시 모니터링 편의성과 복원력(Resilience)을 강화하는 것입니다.

## 2. 시스템 아키텍처 변경 사항
기존의 모놀리식 Stepper 구조를 유지하되, 상태 관리와 UI 렌더링 방식을 고도화합니다. 특히 수천 개의 테이블 마이그레이션 이벤트를 처리하기 위해 프론트엔드의 DOM 렌더링 방식을 최적화해야 합니다.

* **프론트엔드**: Vanilla JS + HTML/CSS 유지. 외부 무거운 프레임워크(React/Vue 등) 도입 없이, 경량화된 Virtual Scrolling 기법 및 상태 관리 패턴을 바닐라 환경에서 구현.
* **백엔드 (Go)**: Target DB 사전 검증 API 엔드포인트 추가, Retry(재시도)를 위한 잡(Job) 관리 기능 확장.

## 3. 화면별 상세 명세

### 3.1. Step 1: Source & Target Connection
기존 Step 1(Oracle만 연결)과 Step 2의 타겟 DB 설정 일부를 통합합니다.

#### UI 컴포넌트
* **Source DB (Oracle) Card**:
  * 입력 폼: URL, Username, Password, Table Filter (LIKE).
  * 액션 버튼: `[Connect & Fetch Tables]` -> 클릭 시 대상 테이블 목록을 백그라운드로 가져옴과 동시에 Source 연결 유효성 검증.
* **Target DB Card** (신규 추가):
  * 입력 폼: Target DB 선택 (Select Box: PostgreSQL, MySQL, MariaDB, SQLite, MSSQL), Target URL, Schema.
  * 액션 버튼: `[Test Target Connection]` -> 클릭 시 Target DB 연결 유효성 및 권한 검증.
* **상태 표시**: 연결 성공/실패 배너 (초록색/빨간색).

#### API 엔드포인트
* `POST /api/tables` (기존 유지, Target 검증은 별도 분리 고려)
* `POST /api/test_target_connection` (신규): Target DB 접속 정보를 받아 핑(Ping) 테스트 후 결과 반환.

### 3.2. Step 2: Configuration & Selection
복잡했던 설정 화면을 점진적 정보 노출(Progressive Disclosure) 방식으로 개선하고 테이블 선택 UI를 고도화합니다.

#### UI 컴포넌트
* **Table Selection Data Table**:
  * 기존의 단순 리스트를 Table 형태로 변경.
  * 컬럼: Checkbox(선택), 테이블명, (가능한 경우) 예상 로우 수.
  * 헤더: 검색창, 전체 선택/해제 버튼, `n / m 선택됨` 카운터.
  * 성능: 1,000개 이상의 테이블 렌더링 시 브라우저 버벅임을 방지하기 위해 가상 스크롤(Virtual Scrolling) 구현.
* **Migration Mode (기본 설정)**:
  * 모드 선택: "SQL 파일 생성", "Direct Migration (직접 이관)".
  * "SQL 파일 생성" 선택 시: 출력 파일명, 테이블별 분할 저장 옵션만 노출.
* **DDL Options (아코디언 형태 적용 가능)**:
  * "CREATE TABLE DDL 생성/실행" 체크 시 세부 옵션(인덱스, 제약조건 등) 노출.
* **Advanced Settings (고급 설정 패널)**:
  * 기본적으로 접혀있는 아코디언(Accordion) 또는 토글 버튼으로 제공.
  * 항목: Batch Size, Parallel Workers, COPY Batch, DB Pool (Max Open/Idle/Life), JSON Logging, Dry-Run 등.

### 3.3. Step 3: Execution & Monitoring
진행 상황 모니터링을 대시보드 형태로 전면 개편합니다.

#### UI 컴포넌트
* **Pre-flight Check Modal (신규)**:
  * Step 2에서 `[Start Migration]` 클릭 시 팝업.
  * 내용: "Target DB: PostgreSQL, 선택된 테이블: 150개, 모드: Direct Migration. 진행하시겠습니까?"
  * 액션: `[Confirm]`, `[Cancel]`.
* **Summary Dashboard Widget**:
  * Metrics: 전체 진행률 바(%), 성공 건수, 실패 건수, 남은 예상 시간(ETA), 실시간 처리 속도(Rows/sec).
* **상태별 탭 (Status Tabs)**:
  * `All (전체)`, `Pending (대기)`, `Running (진행중)`, `Completed (완료)`, `Error (에러)`.
  * 탭 클릭 시 하단의 테이블 진행률 목록 필터링.
* **테이블별 Progress Item**:
  * 성공/진행중인 테이블: 기존 형태의 컴팩트한 Progress Bar 유지.
  * 에러 발생 테이블: 붉은색 강조, 에러 메시지 축약 표시. 클릭 시 세부 로그 및 **`[Retry]` 버튼 노출**.

#### 실시간 데이터 처리 (WebSocket)
* 진행률(ETA/속도) 계산을 위해 프론트엔드에 간단한 윈도우(Window) 기반 속도 계산 로직 추가.
* 에러 발생 시 `error` 타입 메시지에 Retry에 필요한 식별자(Job ID, Table Name 등) 포함되도록 백엔드 구조 검토 필요 (이미 제공되는 경우 활용).

### 4. 백엔드(API) 주요 변경 사항 요약
1. **Target DB 연결 테스트 API**: 설정 전 미리 타겟 DB의 상태를 검증할 수 있는 `GET/POST /api/test-target` 엔드포인트 구현.
2. **테이블 메타데이터 제공 확대**: `POST /api/tables` 호출 시 단순 테이블 이름 배열(`[]string`) 뿐만 아니라, 향후 Data Table에서 활용할 수 있도록 객체 배열(`[{name: "T1", row_count: 1000}, ...]`) 형태의 데이터 반환 검토 (성능 문제로 지연될 경우 옵션화).
3. **단일 테이블 재시도 (Retry) API (선택/Phase 4)**: 마이그레이션 실패 시, 전체를 다시 돌리는 것이 아니라 특정 테이블만 다시 큐에 넣고 실행할 수 있는 `POST /api/migrate/retry` 엔드포인트 추가 검토.

## 5. 단계별 개발 계획 (Implementation Phases)
본 스펙은 다음과 같은 순서로 개발됩니다.

* **Task 1: 타겟 DB 설정 이동 및 검증 API 추가** (Step 1 통합, 사전 검증 기능 구현)
* **Task 2: UI 레이아웃 및 컴포넌트 재구성** (Step 2 고급 설정 아코디언화, 점진적 노출 적용)
* **Task 3: 테이블 리스트 Data Table 및 가상 스크롤 적용** (대규모 테이블 렌더링 성능 최적화)
* **Task 4: 모니터링 대시보드 개편** (Step 3 요약 위젯, 상태별 탭 필터, ETA 계산 로직 추가)
* **Task 5: 단일 테이블 재시도(Retry) 기능 연결** (에러 처리 및 복원력 강화)


### tasks.md

# Go DB Migration v10 개발 태스크 (Tasks)

본 문서는 `prd.md` 및 `spec.md`를 바탕으로 개발자가 순차적으로 실행할 수 있도록 분할한 세부 작업 목록입니다.

## Phase 1: 구조 개편 및 사전 검증 도입 (Step 1 & 2 통합)
* [x] **Task 1.1: UI 레이아웃 개편**
  * `index.html`에서 타겟 DB 설정 폼(Target DB 선택, URL, Schema)을 Step 1 하단 또는 Step 2의 메인 뷰에서 Step 1으로 이동.
  * 기존 "Connect & Fetch Tables" 플로우에 "Test Target Connection" 플로우 추가.
* [x] **Task 1.2: 타겟 DB 검증 API 추가**
  * `internal/web/server.go`에 `POST /api/test-target` 엔드포인트 구현.
  * `internal/db/db.go` 등을 활용해 입력받은 URL과 드라이버로 DB Ping 테스트만 수행하고 결과를 JSON으로 반환.
* [x] **Task 1.3: Step 1 상태 관리 연동**
  * Source와 Target이 모두 유효하게 검증된 경우에만 Step 2로 넘어갈 수 있도록 클라이언트 사이드 검증 로직 수정.

## Phase 2: 설정 및 테이블 선택 UI 고도화 (Step 2)
* [x] **Task 2.1: Data Table 컴포넌트 적용**
  * `index.html`의 테이블 리스트 영역(`tableList`)을 단순 `div` 리스트에서 `table` (또는 flex/grid 기반의 표 형태) 레이아웃으로 변경.
  * 컬럼 구조: 선택(Checkbox), 테이블명, 상태(옵션).
* [x] **Task 2.2: 고급 설정 아코디언 구현**
  * Batch Size, Workers, Max Open 등 복잡한 설정들을 "Advanced Settings" 토글 버튼(혹은 `<details>` 태그 스타일의 컴포넌트) 내부에 숨김.
  * DDL 옵션도 트리거(체크박스)에 따라 하위 옵션이 슬라이드 다운되도록 CSS 애니메이션/JS 수정.
* [ ] **Task 2.3: 가상 스크롤 (Virtual Scrolling) 적용 검토**
  * 테이블 개수가 1000개 이상일 때를 대비해, 바닐라 JS로 가벼운 Virtual List(또는 Intersection Observer 활용)를 구현하거나 DOM 업데이트 최적화.

## Phase 3: 모니터링 대시보드 개편 (Step 3)
* [x] **Task 3.1: 요약 위젯 (Summary Widget) 실시간화**
  * 현재 마이그레이션 종료 시(`all_done`)에만 뜨는 Summary Card를 마이그레이션 시작 시점부터 띄움.
  * 진행률(%), 성공, 실패 건수를 실시간 갱신.
* [x] **Task 3.2: ETA 및 속도 계산 로직 구현**
  * 프론트엔드(`tracker.go`에서 전달되는 진행 데이터 기반)에 초당 처리 행 수(Rows/sec)와 남은 예상 시간(ETA)을 계산하여 UI에 표시.
* [x] **Task 3.3: 상태별 탭(Tabs) 필터링 추가**
  * "전체", "진행중", "완료", "에러" 탭 버튼 추가.
  * 클릭 시 하단의 테이블 진행률 컨테이너 내의 아이템들을 `display: none/block`으로 필터링.
* [x] **Task 3.4: 실패 테이블 에러 로그 UI 개선**
  * 에러 발생 시 상세 정보(Phase, Category 등)가 더 눈에 띄게 펼쳐지도록 아코디언 컴포넌트 스타일 적용.

## Phase 4: 기능 안정화 및 Retry (옵션/추가 스펙)
* [x] **Task 4.1: 단일 테이블 Retry 버튼 UI 추가**
  * 에러가 발생한 테이블(`status === 'error'`)의 아이템 우측에 `[재시도]` 버튼 렌더링.
* [x] **Task 4.2: Retry API 엔드포인트 연동 (백엔드 지원 시)**
  * `POST /api/migrate/retry` (가칭) 엔드포인트에 해당 테이블 정보만 전송하여 해당 테이블의 마이그레이션만 다시 큐에 넣는 기능 구현.
* [x] **Task 4.3: Edge Case 테스트 및 버그 수정**
  * 브라우저 탭 이동 시 타이머(ETA) 이슈 확인.
  * 다크모드 적용 시 누락된 컬러 변수(테이블 헤더, 탭 활성화 상태 등) 보완.

## Phase 5: 테스트 케이스 작성 여부 (Test Coverage Check)
* [x] **Task 5.1: 타겟 DB 검증 API 테스트**
  * `POST /api/test-target` 엔드포인트에 대한 성공 및 실패 케이스 검증.
* [x] **Task 5.2: 재시도(Retry) API 테스트**
  * `POST /api/migrate/retry` 엔드포인트 요청에 대한 처리 여부 확인.
* [x] **Task 5.3: 입력값 검증 (Validation) 테스트 확장**
  * DB 커넥션 풀(DB Max Open/Idle/Life) 등 새로운 설정값에 대한 유효성 테스트 작성.


---
## <a name="v11"></a> v11

### prd.md

# Go DB Migration v11 심층 분석 및 개선 PRD (Product Requirements Document)

## 1. 개요 (Overview)
v10 업데이트를 통해 Go DB Migration 도구의 기본 웹 UI 구조가 점진적 정보 노출(Progressive Disclosure) 및 사전 검증(Pre-flight Check) 중심으로 개편되었고, 대량 데이터 마이그레이션을 위한 기본 모니터링 대시보드가 도입되었습니다.
하지만 실시간 데이터 처리의 **응답성 한계**, **대용량 단일 테이블 처리의 병목**, 그리고 에러 발생 시의 **수동 복구(Manual Recovery)의 번거로움** 등 엔터프라이즈급 안정성과 사용성을 확보하기 위한 개선의 여지가 남아있습니다.

v11의 핵심 목표는 **실시간 양방향 통신(WebSocket) 기반의 즉각적인 모니터링**, **대형 테이블 자동 파티셔닝(Chunking)을 통한 마이그레이션 성능 극대화**, 그리고 **능동적 에러 복구(Auto-Healing) 메커니즘**을 도입하여 마이그레이션 도구의 완성도를 한 단계 끌어올리는 것입니다.

## 2. v10(현재)의 한계점 심층 분석

### 2.1. 웹 UI 및 모니터링 측면 (UI/UX)
* **단방향 폴링(Polling)의 한계**: 클라이언트가 서버에 주기적으로 상태를 요청(Polling)하는 구조로 인해, 네트워크 오버헤드가 발생하고 실시간 진행률(Progress) 반영에 지연(Latency)이 발생합니다.
* **시각적 인사이트 부족**: 테이블 간의 연관 관계(Foreign Key)나 마이그레이션 순서를 시각적으로 보여주는 기능이 부재하여 복잡한 스키마 구조 파악이 어렵습니다.
* **과거 마이그레이션 히스토리 추적 불가**: 브라우저를 새로고침하거나 종료하면 이전 마이그레이션 실행 이력(History) 및 로그 기록이 휘발되어 감사(Audit)가 불가능합니다.

### 2.2. 성능 및 마이그레이션 코어 로직 측면 (Performance & Core)
* **초대형 단일 테이블의 병목 현상**: 수천만 건 이상의 데이터를 가진 단일 테이블 마이그레이션 시, 단일 워커가 순차 처리하므로 전체 마이그레이션 완료 시간이 크게 지연됩니다.
* **메모리 최적화 부족**: 고해상도 LOB(BLOB, CLOB) 데이터를 대량으로 마이그레이션 할 때, 배치(Batch) 메모리 점유율이 높아져 OOM(Out of Memory) 현상이 발생할 위험이 있습니다.

### 2.3. 안정성 및 에러 처리 (Stability & Error Handling)
* **수동 재시도의 한계**: 에러 발생 시 개별 테이블 단위로 재시도(Retry) 버튼을 눌러야 하는 번거로움이 존재하며, 일시적인 네트워크 순단 등에 대한 자동 재시도(Auto-Retry) 로직이 부족합니다.
* **무중단 마이그레이션 미지원**: 마이그레이션 도중 타겟 DB 접속 장애 시, 큐(Queue)에 있는 작업을 멈추지 않고 그대로 실패 처리하여 복원력이 떨어집니다.

---

## 3. v11 개선 목표 및 핵심 요구사항

### 3.1. 실시간 모니터링 고도화 (WebSocket 도입)
* **WebSocket 기반 양방향 통신**: 서버-클라이언트 간 통신을 기존 HTTP Polling에서 WebSocket 구조로 전면 개편.
  * 초당 수백 건의 마이그레이션 상태 변경을 오버헤드 없이 즉각적으로 UI에 반영.
  * 서버 리소스 절약 및 UI 응답성 극대화.
* **고급 대시보드 리포팅**: IOPS(Input/Output Operations Per Second), 네트워크 대역폭 사용량, CPU/메모리 실시간 사용량 등 인프라 레벨의 메트릭 추가 제공.

### 3.2. 대용량 테이블 분할 병렬 처리 (Table Chunking)
* **지능형 테이블 파티셔닝**: 수백만 로우(Row) 이상의 테이블을 식별하고, Primary Key나 숫자/날짜형 인덱스 컬럼을 기준으로 N개의 청크(Chunk)로 자동 분할.
* **청크 단위 워커 할당**: 분할된 청크를 여러 워커(Worker)에 분산 할당하여 **단일 테이블 내비 병렬 처리(Intra-table Parallelism)** 지원. (기존은 테이블 단위 병렬 처리만 지원)
* **UI 반영**: 대형 테이블 클릭 시 하위 청크들의 개별 진행 상황을 Tree 구조나 아코디언 형태로 확인 가능.

### 3.3. 자동 복구 및 안정성 강화 (Auto-Healing & Resilience)
* **스마트 자동 재시도 (Smart Auto-Retry)**:
  * 네트워크 타임아웃, 일시적 Lock 등으로 인한 실패를 감지하여 설정된 횟수만큼 자동 재시도(Exponential Backoff 적용).
  * 영구적 에러(예: 스키마 불일치)와 일시적 에러를 구분하여 지능적으로 대처.
* **Pause & Resume (일시정지 및 재개)**: 전체 마이그레이션을 일시정지시키고 타겟 DB의 상태(예: 디스크 용량 확보)를 정비한 후 끊긴 시점부터 재개(Resume)할 수 있는 기능.

### 3.4. 사용자 편의성 및 기타 백엔드 개선 (UX & Backend)
* **마이그레이션 히스토리 로컬 스토리지/SQLite 영속화**: 마이그레이션 세션 기록, 성공/실패 로그, 통계 데이터를 내부 SQLite 등에 저장하여 언제든 이전 내역(History 탭)을 열람하고 보고서(PDF/CSV)로 다운로드.
* **스키마 토폴로지 뷰어 (Topology Viewer)**: Oracle 데이터 딕셔너리를 분석하여 테이블 간의 FK 종속성을 시각적 노드 그래프로 렌더링. 종속성에 맞춘 안전한 마이그레이션 순서 자동 제안.
* **클라우드 스토리지 백업 (옵션)**: SQL Dump 모드 사용 시 결과물을 로컬뿐만 아니라 AWS S3, Google Cloud Storage 등으로 직접 스트리밍 업로드 기능 추가.

---

## 4. 비기능적 요구사항 (Non-functional Requirements)
* **무중단 하위 호환성 (Backward Compatibility)**: 기존 v10 CLI 명령어 및 설정 파일(`config.json`)을 수정 없이 그대로 v11에서도 사용할 수 있어야 함.
* **Zero Dependency 원칙 유지**: 순수 Go 언어만으로 빌드되며, 별도의 메시지 큐(Redis, RabbitMQ)나 무거운 외부 런타임 설치 없이 단일 실행 파일로 제공되어야 함.
* **LOB 데이터 스트리밍 메모리 제한**: LOB 데이터를 메모리에 전체 버퍼링하지 않고 `io.Reader`와 `io.Writer`를 활용한 파이프라인 스트리밍을 통해 최대 메모리 사용량을 예측 가능하게 고정(Cap).

## 5. 마일스톤 및 릴리스 계획
* **Phase 1: 기반 아키텍처 개선** - WebSocket 통신망 구축 및 내부 Event Bus(Pub/Sub) 구조 도입.
* **Phase 2: 고성능 코어 개발** - 단일 테이블 Chunking 로직 및 LOB 스트리밍 최적화 적용.
* **Phase 3: 자동 복구 및 상태 영속화** - 스마트 Auto-Retry 로직, Pause/Resume 기능 추가, 로컬 SQLite 기반 히스토리 저장소 구현.
* **Phase 4: UI/UX 고도화** - Topology Viewer, 고급 실시간 대시보드 뷰어, 상태별 하위 청크 트래킹 UI 연동.
* **Phase 5: 알파 테스트 및 안정화** - OOM 및 네트워크 순단 테스트 진행, 최종 문서화(v11).


### spec.md

# Go DB Migration v11 기능 명세서 (Technical Specification)

## 1. 개요 (Overview)
본 문서는 `docs/v11/prd.md`에 정의된 요구사항을 바탕으로, Go DB Migration 도구의 핵심 아키텍처 및 백엔드/프론트엔드 변경 사항을 구체화한 기술 명세서입니다. 핵심 목표는 실시간 양방향 통신(WebSocket) 기반의 모니터링, 대형 테이블 자동 파티셔닝(Chunking)을 통한 성능 극대화, 그리고 능동적 에러 복구(Auto-Healing) 메커니즘의 도입입니다.

## 2. 시스템 아키텍처 변경 사항

### 2.1. 실시간 양방향 통신 (WebSocket)
기존 HTTP Polling 방식의 한계를 극복하기 위해 WebSocket을 도입합니다.
* **프론트엔드**: 브라우저 기본 `WebSocket` API를 사용하여 서버와 지속적인 연결 유지. 상태 갱신은 푸시(Push) 방식으로 수신.
* **백엔드 (Go)**: `nhooyr.io/websocket` 또는 `gorilla/websocket`을 활용하여 웹소켓 서버 구축. 다중 클라이언트 브로드캐스팅 지원.
* **Event Bus**: 마이그레이션 워커(Worker)와 웹소켓 브로드캐스터 간의 결합도를 낮추기 위해, 내부적인 채널(Channel) 기반의 Pub/Sub Event Bus 구조 도입.

### 2.2. 테이블 파티셔닝 및 병렬 처리 (Table Chunking)
단일 테이블 처리의 병목을 해소하기 위해 청크(Chunk) 기반 분할 처리를 구현합니다.
* **청킹 전략**: 로우(Row) 수가 설정된 임계값(예: 100만 건) 이상인 테이블을 식별. Primary Key나 인덱스된 숫자/날짜 컬럼을 기준으로 `SELECT` 쿼리에 `WHERE` 범위를 주어 N개의 청크로 분할.
* **워커 풀 구조 개선**: 워커 풀의 작업 단위를 '테이블'에서 '청크(Chunk)'로 세분화. 여러 워커가 단일 테이블의 서로 다른 청크를 동시에 처리 가능 (Intra-table Parallelism).
* **상태 추적**: `Tracker`는 각 청크의 완료 상태를 집계하여 테이블 전체의 진행률로 변환.

### 2.3. 에러 복구 및 영속화 (Auto-Healing & Persistence)
안정성을 위해 스마트 재시도와 히스토리 영속성을 추가합니다.
* **스마트 자동 재시도 (Smart Auto-Retry)**: 일시적 에러(네트워크 순단, DB 락 등)와 영구적 에러(스키마 불일치 등)를 분류. 일시적 에러에 대해 Exponential Backoff 적용하여 N회 자동 재시도.
* **Pause & Resume**: 전체 마이그레이션 프로세스를 일시 중지하고, 이후 끊긴 지점(완료되지 않은 청크)부터 재개할 수 있는 제어 기능 추가.
* **SQLite 기반 히스토리 영속화**: 로컬 SQLite DB(`history.db`)를 도입하여, 마이그레이션 세션 기록(시작/종료 시간, 통계, 에러 로그) 저장. UI의 History 탭에서 조회.

## 3. 세부 컴포넌트 명세

### 3.1. WebSocket 프로토콜 및 메시지
`ws://<host>/api/ws` 엔드포인트를 통해 JSON 포맷 메시지 송수신.

* **Progress 메시지 (진행률 갱신)**:
```json
{
  "type": "progress",
  "table": "ORDERS",
  "chunk_id": "ORDERS_chunk_1",
  "total": 5000000,
  "count": 10000,
  "status": "running"
}
```
* **Metrics 메시지 (대시보드 메트릭)**:
```json
{
  "type": "metrics",
  "iops": 4500,
  "network_rx_mbps": 12.5,
  "network_tx_mbps": 50.2,
  "cpu_usage_pct": 45.2,
  "mem_usage_mb": 1024
}
```
* **Error 메시지 (에러 발생)**:
```json
{
  "type": "error",
  "table": "USERS",
  "chunk_id": "USERS_chunk_3",
  "error_msg": "Network timeout during INSERT",
  "retry_count": 1,
  "will_retry": true
}
```

### 3.2. 청킹(Chunking) 로직
* **분할 알고리즘**:
  * 숫자형 PK: `MIN(PK)`, `MAX(PK)` 조회 후 균등 간격으로 분할.
  * 날짜형 인덱스: 월/일 단위 분할.
  * PK 부재 시: PostgreSQL `ctid`, Oracle `ROWID` 등 물리적 식별자 활용 또는 청킹 생략(단일 처리).
* **동시성 제어**: 청크 단위 결과가 합산될 때 Race Condition을 방지하기 위해 `sync.Mutex` 또는 원자적 연산(Atomic) 사용.

### 3.3. SQLite 테이블 스키마 설계
`internal/db/history.go`에서 관리될 스키마:
* `migration_jobs`: `job_id(PK)`, `start_time`, `end_time`, `status`, `total_tables`, `success_tables`, `error_tables`.
* `table_progress`: `job_id(FK)`, `table_name(PK)`, `total_rows`, `migrated_rows`, `status`.
* `error_logs`: `log_id(PK)`, `job_id(FK)`, `table_name`, `chunk_id`, `error_msg`, `created_at`.

### 3.4. 메모리 관리 (LOB 스트리밍)
LOB(BLOB, CLOB) 데이터 처리 시 메모리 누수 방지.
* `sql.Rows`의 `Scan` 결과를 대형 버퍼에 담지 않고, `io.Reader`로 래핑하여 타겟 DB 드라이버의 파이프라인(스트리밍 입력) 기능 직접 활용. 최대 배치 메모리 고정.

## 4. 백엔드 API 변경 사항
* `GET /api/ws`: WebSocket 업그레이드 엔드포인트.
* `POST /api/migrate/pause`: 현재 마이그레이션 세션 일시 중지.
* `POST /api/migrate/resume`: 일시 중지된 마이그레이션 세션 재개.
* `GET /api/history`: SQLite에 저장된 마이그레이션 세션 목록 조회.
* `GET /api/history/:job_id`: 특정 마이그레이션 세션의 상세 내역 조회.

## 5. 단계별 개발 계획 (Implementation Phases)
본 스펙은 다음과 같은 순서로 개발됩니다.

* **Phase 1: 기반 아키텍처 개편** (WebSocket, Event Bus 구축)
* **Phase 2: 고성능 코어 엔진** (Chunking 로직, Chunk 기반 워커 풀, LOB 스트리밍)
* **Phase 3: 복원력 및 상태 영속화** (스마트 재시도, Pause/Resume, SQLite 연동)
* **Phase 4: 웹 UI 고도화** (대시보드 메트릭 연동, 청크 상태 트리 UI, History 탭 구현)
* **Phase 5: 안정화 및 최적화** (동시성 테스트, OOM 벤치마크, 문서화)


### tasks.md

# Go DB Migration v11 개발 태스크 (Tasks)

본 문서는 `prd.md` 및 `spec.md`를 바탕으로 개발자가 순차적으로 실행할 수 있도록 분할한 세부 작업 목록입니다.

## Phase 1: 실시간 모니터링 고도화 (WebSocket 도입)
* [x] **Task 1.1: WebSocket 서버 아키텍처 구축**
  * `internal/web/ws/` 디렉토리에 WebSocket 핸들러 및 연결 관리(Connection Manager) 구현.
  * 기존 HTTP 단방향 Polling 엔드포인트를 대체할 `ws://<host>/api/ws` 엔드포인트 라우팅 추가.
* [x] **Task 1.2: 내부 Event Bus (Pub/Sub) 구현**
  * 마이그레이션 코어 로직(Worker)과 웹소켓 브로드캐스터 간의 의존성 분리를 위한 경량 Event Bus 구조 도입.
  * 상태 변경, 에러 발생, 진행률 업데이트 이벤트를 정의하고 발행(Publish) 및 구독(Subscribe) 로직 작성.
* [x] **Task 1.3: 프론트엔드 WebSocket 클라이언트 연동**
  * `ui.js` (또는 관련 프론트엔드 스크립트)에서 기존 `setInterval` 기반의 Polling 로직 제거.
  * WebSocket API(`new WebSocket()`)를 사용하여 서버 연결, 재연결(Reconnection) 로직 및 수신된 이벤트 처리 함수 구현.
* [x] **Task 1.4: 실시간 대시보드 UI 반영**
  * WebSocket을 통해 수신된 초당 처리량(IOPS), 네트워크 대역폭, 예상 남은 시간(ETA) 등의 메트릭을 UI 대시보드 위젯에 즉각적으로(지연 없이) 렌더링.

## Phase 2: 대용량 테이블 분할 병렬 처리 (Table Chunking)
* [ ] **Task 2.1: 지능형 청킹(Chunking) 로직 구현**
  * 테이블 메타데이터를 분석하여 전체 Row 수를 파악하고, 설정된 임계치(예: 100만 건) 초과 시 테이블을 N개의 청크로 분할하는 로직 (`internal/migration/chunking.go` 등) 구현.
  * Primary Key 또는 인덱스된 숫자/날짜 컬럼을 기준으로 안전하게 범위를 나누는 쿼리 생성기(Dialect 별) 작성.
* [ ] **Task 2.2: 워커 풀(Worker Pool) 구조 개선**
  * 기존 '테이블 단위'로 할당되던 작업을 '청크 단위'로 할당할 수 있도록 Task Queue 구조 변경.
  * 여러 워커가 동일한 테이블의 서로 다른 청크를 동시에 처리(Intra-table Parallelism)할 수 있도록 동시성 제어 강화.
* [ ] **Task 2.3: LOB 데이터 스트리밍 파이프라인 최적화**
  * OOM 방지를 위해 대용량 BLOB/CLOB 컬럼 처리 시 메모리 버퍼 대신 `io.Reader`/`io.Writer`를 활용한 파이프라인 스트리밍 로직 전면 적용.
* [ ] **Task 2.4: 청크 진행 상황 UI 트리 구조 반영**
  * 프론트엔드 테이블 리스트에서 대형 테이블 클릭 시, 하위 청크들의 개별 진행률(Progress Bar)이 아코디언 또는 트리 형태로 노출되도록 UI 컴포넌트 확장.

## Phase 3: 능동적 에러 복구 및 안정성 (Auto-Healing)
* [ ] **Task 3.1: 에러 분류 및 스마트 재시도(Auto-Retry) 로직**
  * 데이터베이스 에러 코드를 분석하여 일시적 네트워크 에러, Lock 타임아웃 등(Retry 가능)과 스키마 불일치(Retry 불가)로 분류.
  * Retry 가능한 에러 발생 시 Exponential Backoff 알고리즘을 적용하여 지정된 횟수만큼 자동 재시도하는 래퍼(Wrapper) 함수 구현.
* [ ] **Task 3.2: Pause & Resume (일시정지/재개) 제어 기능**
  * 전체 마이그레이션 세션에 대한 'Pause' 및 'Resume' 상태를 관리하는 컨트롤러 구현.
  * 웹 UI에 [일시정지], [재개] 버튼 추가 및 관련 API 연동.
  * 일시정지 시 현재 진행 중인 청크까지만 완료하고 워커를 대기 상태로 전환하는 로직 작성.

## Phase 4: 마이그레이션 히스토리 영속화 및 부가 기능
* [ ] **Task 4.1: SQLite 기반 로컬 히스토리 저장소 구축**
  * `internal/db/history.go` (가칭)를 생성하여 마이그레이션 세션 정보, 통계, 에러 로그를 저장할 SQLite 스키마(테이블 생성 쿼리 포함) 작성.
  * 마이그레이션 시작/종료, 에러 발생 시점에 비동기로 DB에 기록하는 로직 구현.
* [ ] **Task 4.2: 히스토리 뷰어 탭 구현 (UI)**
  * 프론트엔드에 'History' 탭 추가.
  * SQLite에서 과거 실행 이력을 페이징하여 불러오는 REST API(`GET /api/history`) 연동 및 테이블 형태로 출력.
* [ ] **Task 4.3: 클라우드 스토리지 스트리밍 (선택 사항/SQL 덤프 모드)**
  * 옵션 활성화 시 생성된 SQL Dump 파일을 로컬 디스크뿐만 아니라 AWS S3 호환 API를 통해 스트리밍 업로드하는 기능 추가. (Go `io.Pipe` 등 활용)

## Phase 5: 안정화 및 최적화
* [ ] **Task 5.1: 통합 테스트 및 성능 검증**
  * 1,000만 건 이상의 Dummy Data가 있는 단일 테이블 환경에서 청크 분할 및 메모리 사용량 추이 모니터링 테스트.
  * 임의로 네트워크를 단절시켜 Auto-Retry 및 Pause/Resume 정상 동작 확인.
* [ ] **Task 5.2: OOM 벤치마크 및 문서화**
  * 대용량 LOB 컬럼을 다수 포함한 테이블을 마이그레이션하여 메모리 사용량이 고정(Cap)되는지 확인.
  * `v11` 기능에 대한 README 및 사용자 가이드 업데이트.


---
## <a name="v12"></a> v12

### prd.md

# Go DB Migration v12 PRD

## 배경
CLI 사용자가 매번 긴 플래그를 기억해야 해서 사용성이 떨어집니다.

## 목표
- Bash, Zsh, Fish, PowerShell 환경에서 dbmigrator 플래그 자동완성을 제공한다.
- 기존 마이그레이션 실행 경로와 하위 호환성을 유지한다.

## 요구사항
1. `-completion <shell>` 플래그를 제공한다.
2. 지원 쉘: `bash`, `zsh`, `fish`, `powershell`.
3. 자동완성 모드에서는 DB 접속 필수 플래그 검증을 건너뛴다.
4. 지원하지 않는 쉘 입력 시 명확한 에러를 반환한다.


### spec.md

# Go DB Migration v12 Technical Spec

## 설계
- `internal/config/config.go`에 `CompletionShell` 필드를 추가한다.
- `generateCompletionScript(shell string)` 헬퍼로 쉘별 스크립트를 생성한다.
- `ParseFlags()`에서 `-completion`이 설정되면 스크립트를 stdout으로 출력하고 조기 반환한다.

## 호환성
- 기존 플래그 동작은 유지한다.
- completion 모드 외에는 기존 필수 플래그 검증 로직을 그대로 적용한다.

## 테스트
- `-completion=bash` 입력 시 필수 플래그 없이도 성공하는지 검증.
- completion 출력에 핵심 스니펫이 포함되는지 검증.
- 미지원 쉘 입력 시 에러를 반환하는지 검증.


### tasks.md

# Go DB Migration v12 개발 태스크

- [x] `Config`에 `CompletionShell` 필드 추가
- [x] `-completion` CLI 플래그 파싱 추가
- [x] 쉘별 자동완성 스크립트 생성기 구현 (bash/zsh/fish/powershell)
- [x] completion 모드 조기 반환 로직 추가 (필수 플래그 검증 건너뜀)
- [x] 미지원 쉘 에러 처리 추가
- [x] 단위 테스트 추가
- [x] README 버전/플래그/사용 예시 문서 업데이트


---
## <a name="v13"></a> v13

### prd.md

# Product Requirements Document (PRD) - v13

## 1. 개요 (Overview)
현재 DB 마이그레이션 도구는 `-completion` 플래그를 통해 다양한 쉘(bash, zsh, fish, powershell)에 대한 자동완성 스크립트를 제공하고 있습니다. 하지만 사용자가 쉘 종류를 명시하지 않고 `-completion` 플래그만 단독으로 입력할 경우, Go 표준 `flag` 패키지의 기본 에러 메시지("flag needs an argument")가 출력되어 사용자 경험이 다소 불친절한 문제가 있습니다. 
본 기능 개선은 `-completion` 플래그만 입력되었을 때, **현재 사용 중인 쉘 환경 변수(`$SHELL`)를 자동으로 감지하여 해당 쉘에 맞는 스크립트를 출력**하고, 만약 감지할 수 없거나 지원하지 않는 쉘인 경우에는 올바른 사용법을 친절하게 안내하도록 개선하는 것을 목표로 합니다.

## 2. 문제점 (Problem Statement)
- 사용자가 `./dbmigrator -completion`을 실행하면 `flag needs an argument: -completion` 이라는 단순 에러만 발생합니다.
- 사용자는 자신의 쉘 이름(bash, zsh 등)을 매번 명시적으로 입력해야 하는 번거로움이 있습니다.

## 3. 목표 (Goals)
- `-completion` 플래그만 입력 시, `$SHELL` 환경 변수를 통해 현재 쉘을 자동으로 감지하여 해당 자동완성 스크립트를 출력합니다.
- 자동 감지가 불가능하거나 지원되지 않는 쉘일 경우, 사용법(예시 및 지원 쉘 목록)을 화면에 출력합니다.
- 기존의 정상적인 사용법(`-completion=bash`, `-completion zsh` 등)은 아무런 영향 없이 기존대로 동작해야 합니다.

## 4. 상세 요구사항 (Detailed Requirements)
1. **현재 쉘 자동 감지 및 스크립트 출력**
   - `-completion` 플래그에 인자가 주어지지 않은 상태로 프로그램이 실행될 경우, `os.Getenv("SHELL")` 등을 통해 쉘을 감지합니다.
   - 감지된 쉘 이름(예: `bash`, `zsh`, `fish`, `powershell`)이 지원하는 쉘 목록에 포함되어 있다면 해당 스크립트를 출력하고 프로그램이 정상(0) 종료됩니다.
2. **사용자 친화적 안내 메시지 출력**
   - 쉘을 감지하지 못했거나 지원하지 않는 쉘일 경우, 아래와 같은 형태의 안내 메시지를 표준 출력(혹은 표준 에러)으로 제공합니다.
     ```text
     자동 감지된 쉘이 지원되지 않거나 알 수 없습니다.

     사용법:
       -completion <shell>

     지원하는 쉘(shell):
       bash, zsh, fish, powershell

     사용 예시:
       ./dbmigrator -completion bash > /etc/bash_completion.d/dbmigrator
       ./dbmigrator -completion zsh > ~/.zsh/completions/_dbmigrator
     ```
3. **기존 동작 유지**
   - 지원하는 쉘 이름이 정상적으로 주어졌을 때는 안내 메시지 없이 기존처럼 해당 쉘의 자동완성 스크립트만 출력하고 종료되어야 합니다.
4. **오류 처리 및 종료 코드**
   - 인자 없이 `-completion`만 입력되고 자동 감지마저 실패하여 사용법을 출력한 경우, 프로그램은 적절한 에러 종료 코드(예: `1`)와 함께 종료되어야 합니다.

## 5. 비기능 요구사항 (Non-functional Requirements)
- Go 표준 라이브러리의 `flag` 패키지 한계를 극복하기 위해 `os.Args`를 사전 검사하는 등 부작용이 없는 깔끔한 방식으로 구현되어야 합니다.
- 기존의 테스트 코드들을 깨뜨리지 않아야 하며, 새로운 동작에 대한 단위 테스트가 추가되어야 합니다.

### spec.md

# 기술 사양서 (Technical Specifications) - v13

## 1. 아키텍처 및 구현 방향
Go 표준 라이브러리의 `flag` 패키지는 `StringVar` 플래그에 인자가 주어지지 않았을 때 에러 메시지를 출력하고 즉시 `os.Exit(2)`를 호출합니다. 이를 방지하고 사용자 친화적인 메시지를 제공 및 쉘을 자동 감지하기 위해서는 `flag.Parse()`가 실행되기 전에 `os.Args`를 전처리(pre-process)하는 방법이 가장 직관적이고 안정적입니다.

### 1.1. `os.Args` 전처리 및 `$SHELL` 자동 감지 방식 도입
`internal/config/config.go`의 `ParseFlags()` 함수 시작 부분에 인자 검사 로직을 추가합니다.

- 사용자가 입력한 명령어 인자(`os.Args[1:]`)를 순회합니다.
- `-completion` 인자가 발견되었을 때, 다음 조건 중 하나에 해당하면(단독으로 쓰인 경우) 쉘 감지를 시도합니다:
  1. `-completion`이 마지막 인자인 경우.
  2. `-completion` 바로 다음 인자가 `-` 문자로 시작하는 경우 (즉, 다른 플래그인 경우).

**쉘 자동 감지 로직:**
- `os.Getenv("SHELL")`을 통해 현재 쉘 경로를 가져옵니다.
- 경로에 `bash`, `zsh`, `fish`, `pwsh` (또는 `powershell`) 문자열이 포함되어 있는지 확인합니다.
- 매칭되는 쉘이 지원 목록(`bash, zsh, fish, powershell`)에 있다면, `generateCompletionScript()`를 호출하여 스크립트를 출력하고 프로그램 실행을 정상 종료(`os.Exit(0)`)합니다.
- 매칭되지 않거나 빈 값이라면 사용법을 출력하고 프로그램 실행을 에러 상태로 중단(`os.Exit(1)`)합니다.

### 1.2. 사용법 출력 함수 추가
`printCompletionUsage()` 헬퍼 함수를 추가하여 표준 출력(혹은 표준 에러)에 아래 정보를 제공합니다.
- 올바른 사용법
- 지원 가능한 쉘 목록 (bash, zsh, fish, powershell)
- 간단한 실행 예시

### 1.3. 코드 구현 (예시)
```go
func detectShell() string {
    shellEnv := strings.ToLower(os.Getenv("SHELL"))
    if strings.Contains(shellEnv, "bash") { return "bash" }
    if strings.Contains(shellEnv, "zsh") { return "zsh" }
    if strings.Contains(shellEnv, "fish") { return "fish" }
    if strings.Contains(shellEnv, "pwsh") || strings.Contains(shellEnv, "powershell") { return "powershell" }
    return ""
}

func ParseFlags() (*Config, error) {
    // flag.Parse() 호출 전, -completion 단독 사용 예외 처리
    args := os.Args[1:]
    for i, arg := range args {
        if arg == "-completion" || arg == "--completion" {
            // 마지막 인자이거나 다음 인자가 또 다른 플래그일 때 (인자 없음)
            if i+1 == len(args) || strings.HasPrefix(args[i+1], "-") {
                detected := detectShell()
                if detected != "" {
                    script, _ := generateCompletionScript(detected)
                    fmt.Println(script)
                    os.Exit(0)
                } else {
                    printCompletionUsage()
                    os.Exit(1)
                }
            }
        }
    }

    // 기존 flag 초기화 및 파싱 로직
    // ...
}
```

## 2. 테스트 방안
- **통합/단위 테스트 (Unit Test):** 
  - `config_test.go` 파일 내에서, `os.Args`를 임의로 조작(`[]string{"cmd", "-completion"}`)하고 임시로 `os.Setenv("SHELL", "/bin/zsh")`를 설정한 뒤 `ParseFlags()`(또는 분리된 검증 로직)를 호출했을 때, 기존처럼 `flag needs an argument` 패닉/에러가 아니라 해당 쉘의 스크립트가 잘 반환(또는 출력)되는지 검증합니다.
  - 지원하지 않는 쉘의 경우 적절한 사용법 안내 텍스트가 출력되는지 확인합니다.
  - (주의: `os.Exit`을 피하기 위해, 테스트가 용이하도록 `ParseFlags()` 내부 로직을 리팩토링하여 에러나 값을 반환하게 하는 것이 필요할 수 있습니다.)
- **기존 테스트 호환성:** 기존에 `-completion=bash` 등의 정상 케이스에 대한 테스트가 깨지지 않는지 확인합니다.

### tasks.md

# 작업 목록 (Tasks) - v13

## 목표: `-completion` 플래그 단독 사용성 개선 및 쉘 자동 감지

### 1. 설계 (Design & Documentation)
- [x] `docs/v13/prd.md` 작성 (요구사항 문서)
- [x] `docs/v13/spec.md` 작성 (기술 사양서)
- [x] `docs/v13/tasks.md` 작성 (작업 목록)

### 2. 구현 (Implementation)
- [x] `internal/config/config.go`에 `detectShell()` 헬퍼 함수 추가 (`$SHELL` 환경 변수 기반 감지)
- [x] `internal/config/config.go`에 `printCompletionUsage()` 헬퍼 함수 추가
  - [x] `bash, zsh, fish, powershell`에 대한 사용법 및 자동 감지 실패 안내 텍스트 작성
- [x] `internal/config/config.go`의 `ParseFlags()` 함수 시작 부분에 `os.Args` 사전 검사 로직 추가
  - [x] `-completion`이 인자 없이 단독으로 쓰인 경우 `detectShell()` 호출
  - [x] 쉘이 정상적으로 감지되면 `generateCompletionScript()`를 통해 출력 후 정상 종료(`os.Exit(0)`)
  - [x] 쉘을 감지하지 못하면 `printCompletionUsage()` 호출 후 에러 종료(`os.Exit(1)`)

### 3. 테스트 (Testing)
- [x] `internal/config/config_test.go`에 새로운 시나리오 테스트 추가
  - [x] `-completion` 단독 실행 및 `$SHELL` 환경 변수에 따른 정상 스크립트 반환 검증 (`os.Exit` 문제 우회를 위한 테스트 로직 구성)
  - [x] `-completion` 단독 실행 시 `$SHELL` 감지 실패할 경우 사용법 텍스트 출력 검증
- [x] 기존 통합/단위 테스트(`.go` 테스트들) 실행 및 통과 여부 확인 (`go test ./...`)

### 4. 문서 업데이트 (Documentation)
- [x] `README.md` 내용 수정
  - [x] 기능 설명 부분에 `-completion` 단독 실행 시 현재 쉘 자동 감지 기능이 지원된다는 문구 추가

---
## <a name="v14"></a> v14

### prd.md

# Product Requirements Document (PRD) - v14

## 1. 개요 (Overview)
현재 Web UI의 입력 필드(예: 소스/타깃 DB 연결 정보, 스키마/테이블 입력, 옵션 값)는 사용자가 반복적으로 유사한 값을 다시 입력해야 하는 경우가 많습니다. 이로 인해 작업 시간이 늘어나고, 오타 및 설정 불일치가 발생할 수 있습니다.

본 기능은 **사용자가 최근 입력한 값을 브라우저에 저장하고, 동일/유사 입력 필드에서 자동완성 제안으로 재사용**할 수 있도록 하여 입력 효율과 정확성을 높이는 것을 목표로 합니다.

## 2. 문제점 (Problem Statement)
- 반복 마이그레이션 작업 시 동일한 값(호스트, 포트, DB명, 사용자명, 스키마 등)을 매번 재입력해야 합니다.
- 재입력 과정에서 오타/누락으로 인해 실행 실패 가능성이 증가합니다.
- 입력 이력 재사용 기능이 없어 사용자의 작업 맥락이 단절됩니다.

## 3. 목표 (Goals)
- 최근 입력값 기반 자동완성 제안을 통해 사용자의 반복 입력 시간을 줄입니다.
- 중요한 입력값 재사용 시 오타를 줄여 설정 정확도를 향상합니다.
- 사용자가 자동완성 이력을 통제(삭제/초기화)할 수 있도록 하여 UX 및 프라이버시를 함께 보장합니다.
- DB URL/ID/PASS 핵심 접속 정보는 상단 공통 영역에서 관리하고, 새로고침·재시작·재접속 후에도 복원되어 빠르게 작업을 재개할 수 있게 합니다.

## 4. 범위 (Scope)
### 4.1 In Scope
1. Web UI 주요 텍스트 입력 필드에 최근 입력 자동완성 제안 제공
2. 브라우저 로컬 스토리지 기반 입력 이력 저장/조회
3. 필드별 최근 N개(예: 5~10개) 이력 유지 및 중복 제거
4. 자동완성 제안 선택 시 즉시 입력값 반영
5. 사용자가 필드 이력 또는 전체 이력을 삭제할 수 있는 UI 제공
6. DB URL/ID/PASS를 화면 상단 공통 연결 정보 영역에 배치하고 세션 재진입 시 자동 복원

### 4.2 Out of Scope
- 서버(DB/API)에 입력 이력 저장 및 계정 간 동기화
- AI 기반 의미 추론형 추천
- 비정형 대용량 데이터 자동완성(로그/본문 편집기 수준)

## 5. 사용자 시나리오 (User Scenarios)
1. 사용자는 이전에 입력했던 `source host`를 다시 입력하려고 필드를 클릭합니다.
2. 입력창 하단(또는 브라우저 기본 자동완성 UI)에 최근 사용한 host 목록이 제안됩니다.
3. 사용자가 제안을 선택하면 값이 즉시 채워지고 폼 제출이 가능합니다.
4. 사용자는 “이력 삭제” 액션으로 해당 필드의 저장값을 제거할 수 있습니다.
5. 사용자는 DB URL/ID/PASS를 상단 공통 입력 영역에 한 번 입력한 뒤, 브라우저를 새로고침하거나 다시 접속해도 값이 유지되어 즉시 작업을 이어갈 수 있습니다.

## 6. 상세 요구사항 (Detailed Requirements)
1. **저장 시점**
   - 사용자가 폼 제출 또는 입력 확정(blur/enter) 시 유효한 값을 이력에 저장합니다.
2. **저장 규칙**
   - 빈 문자열, 공백-only 값은 저장하지 않습니다.
   - 동일 값은 중복 저장하지 않고 최신 순으로 재정렬합니다.
   - 필드별 최대 저장 개수(N)를 초과하면 오래된 값부터 제거합니다.
3. **자동완성 노출 규칙**
   - 해당 필드 포커스 시 최근값을 우선 제안합니다.
   - 사용자가 타이핑하면 prefix 기준으로 필터링된 제안을 노출합니다.
4. **이력 관리**
   - 필드별 삭제(예: “최근 source host 지우기”)와 전체 초기화(예: “입력 이력 전체 삭제”)를 지원합니다.
5. **호환성**
   - 브라우저가 localStorage를 지원하지 않거나 비활성화된 경우에도 폼 기본 동작은 정상 유지되어야 합니다.
6. **상단 공통 접속 정보 유지 (DB URL/ID/PASS)**
   - Web UI 상단에 DB 접속 공통 정보 입력 섹션(DB URL, ID, PASS)을 제공합니다.
   - 사용자가 값을 입력/수정하면 브라우저 저장소에 즉시(또는 submit 시) 반영합니다.
   - 페이지 새로고침, 브라우저 재시작, 동일 브라우저 재접속 시 마지막 입력값을 자동 복원합니다.
   - 사용자는 상단 영역에서 개별 항목 삭제 또는 전체 초기화를 수행할 수 있습니다.

## 7. 비기능 요구사항 (Non-functional Requirements)
- **성능:** 자동완성 제안 조회/필터링은 입력 지연을 체감하지 않도록 경량으로 동작해야 합니다.
- **안정성:** localStorage 접근 오류(권한/용량/비활성화) 발생 시 예외를 안전하게 처리하고 기능을 우아하게 비활성화해야 합니다.
- **보안/프라이버시:** 일반 자동완성 이력에서는 비밀번호/토큰 등 민감정보 필드를 저장 대상에서 제외해야 합니다. 단, 상단 공통 접속 정보의 PASS는 사용자 명시 동의(예: "비밀번호 기억")가 활성화된 경우에 한해 저장할 수 있어야 하며 기본값은 비활성화입니다.
- **사용성:** 키보드 내비게이션(↑/↓/Enter/Esc) 또는 브라우저 기본 UX와 충돌 없이 자연스럽게 동작해야 합니다.

## 8. 성공 지표 (Success Metrics)
- 동일 세션/반복 작업에서 평균 입력 시간 감소
- 동일 입력 필드의 재입력 횟수 대비 자동완성 선택 비율 증가
- 오입력으로 인한 실행 실패 케이스 감소(정성/정량)

## 9. 수용 기준 (Acceptance Criteria)
1. 사용자가 동일 필드에 값을 2회 이상 입력하면, 이후 포커스 시 최근값 제안이 노출됩니다.
2. 중복값은 1개만 유지되며 가장 최근 사용 순으로 정렬됩니다.
3. 민감정보 필드(예: password)는 기본적으로 저장/제안되지 않으며, 상단 공통 PASS는 "비밀번호 기억" 옵션을 사용자가 명시적으로 켠 경우에만 예외적으로 저장됩니다.
4. 사용자가 필드별/전체 이력 삭제를 수행하면 즉시 반영됩니다.
5. localStorage 미지원 환경에서도 입력/실행의 핵심 기능은 정상 동작합니다.
6. 사용자가 상단 DB URL/ID/PASS를 저장한 경우, 새로고침/재접속 후 자동 복원됩니다.
7. PASS 저장은 "비밀번호 기억"을 사용자가 명시적으로 활성화한 경우에만 동작합니다.

## 10. 리스크 및 대응 (Risks & Mitigations)
- **리스크:** 공유 PC 환경에서 입력 이력이 노출될 수 있음  
  **대응:** 명확한 이력 삭제 UI 제공, 민감정보 저장 금지, 안내 문구 추가
- **리스크:** PASS 영속 저장으로 인한 보안 우려  
  **대응:** 기본 비활성화(옵트인), 저장 상태 시각화, 즉시 삭제 버튼 제공, 가능한 경우 브라우저 암호화 저장소 사용 검토
- **리스크:** 브라우저별 자동완성 동작 차이  
  **대응:** 커스텀 제안 UI 또는 표준 datalist 기반의 호환성 검증 테스트 수행

## 11. 오픈 이슈 (Open Questions)
- 필드별 최대 저장 개수 N의 기본값은 5, 10 중 무엇이 적절한가?
- 자동완성 제안 UI를 브라우저 기본(`datalist`)로 시작할지, 커스텀 드롭다운으로 구현할지?
- 이력 저장 트리거를 `submit` 중심으로 할지, `blur`까지 확장할지?


### spec.md

# 기술 사양서 (Technical Specifications) - v14

## 1. 아키텍처 및 구현 방향
v14 기능은 서버 저장 없이 브라우저 측 상태(localStorage)를 활용해 구현합니다. 핵심은 (1) 필드별 최근 입력 자동완성, (2) 상단 공통 DB URL/ID/PASS 복원, (3) 민감정보 보호(기본 비저장 + PASS 옵트인)입니다.

구현은 `internal/web/templates/index.html`(폼 구조/UI) + `internal/web/templates/chart.js`(클라이언트 로직) 조합으로 진행하며, 기존 API 스키마 및 백엔드 핸들러(`internal/web/server.go`)는 변경 없이 호환됩니다.

---

## 2. 데이터 모델 (Client-side)

### 2.1 localStorage 키 설계
- 접두사: `dbmigrator:webui:v14:`
- 필드별 최근 이력: `dbmigrator:webui:v14:history:<fieldKey>`
  - 예: `history:sourceHost`, `history:schema`, `history:oracleOwner`
- 상단 공통 접속정보: `dbmigrator:webui:v14:sharedConnection`
  - JSON 구조:
    ```json
    {
      "dbUrl": "...",
      "dbId": "...",
      "dbPass": "...",
      "rememberPass": false,
      "updatedAt": "2026-01-01T00:00:00Z"
    }
    ```

### 2.2 저장 규칙
- trim 후 빈 문자열은 저장하지 않음
- 필드별 중복값 제거 후 최신값을 배열 선두로 이동
- 필드별 최대 N개(기본 10개) 유지
- `password`, `token`, `secret` 계열 필드는 이력 저장 대상에서 제외
- 상단 `dbPass`는 `rememberPass=true`일 때만 저장

---

## 3. UI/UX 사양

### 3.1 상단 공통 연결 영역
- 배치: 페이지 상단(현재 헤더/폼 시작 영역 인접)에 `DB URL`, `DB ID`, `DB PASS`, `비밀번호 기억` 체크박스 제공
- 동작:
  1. 진입 시 `sharedConnection` 읽어서 입력값 자동 복원
  2. 사용자가 수정 시 디바운스(예: 300ms) 또는 submit 시 저장
  3. "초기화" 버튼으로 URL/ID/PASS + rememberPass 일괄 삭제

### 3.2 필드 자동완성
- 대상: 텍스트/검색형 입력 필드 중 민감정보 제외
- 노출:
  - 포커스 시 최근값 목록 노출
  - 입력 중 prefix 매칭 필터
- 선택:
  - 클릭 또는 Enter로 값 반영
- 접근성:
  - ↑/↓로 항목 이동, Enter 선택, Esc 닫기
  - blur 시 드롭다운 닫기

### 3.3 이력 삭제 UX
- 필드 단위 삭제: 입력 우측 또는 설정 메뉴에서 "최근값 지우기"
- 전체 삭제: "입력 이력 전체 삭제" 액션 제공
- 삭제 직후 UI 및 localStorage 즉시 동기화

---

## 4. 클라이언트 로직 세부

### 4.1 모듈화 함수(예시)
- `loadHistory(fieldKey): string[]`
- `saveHistory(fieldKey, value): void`
- `clearHistory(fieldKey?): void` (없으면 전체 삭제)
- `loadSharedConnection(): SharedConnection`
- `saveSharedConnection(model): void`
- `bindAutocomplete(inputEl, fieldKey, options): void`
- `isSensitiveField(fieldKey|inputName): boolean`

### 4.2 예외 처리
- `localStorage` 접근 시 `try/catch`로 감싸고 실패 시 no-op 처리
- JSON 파싱 실패 시 해당 키 삭제 후 기본값 복구
- quota 초과 시 오래된 이력부터 정리 후 재시도

### 4.3 기존 동작 호환성
- localStorage 사용 불가 시 자동완성/복원만 비활성화되고,
  - 입력
  - 테이블 조회
  - 마이그레이션 실행
  - 결과 다운로드
  의 기존 핵심 플로우는 그대로 동작해야 함

---

## 5. 보안/프라이버시 가이드
- 기본 정책: 민감정보 자동 저장 금지
- PASS 저장은 옵트인(`rememberPass=true`)일 때만 허용
- 공유 기기 사용 경고 문구를 상단 영역 근처에 배치
- 가능하면 PASS는 마스킹 상태 유지, 토글 보기 기능은 별도 검토

---

## 6. 테스트 방안

### 6.1 단위 테스트 (JS)
- 중복 제거/최신순 정렬 검증
- 최대 N개 제한 검증
- 민감필드 저장 제외 검증
- rememberPass false 시 PASS 미저장 검증

### 6.2 통합/UI 테스트
- 페이지 재로드 시 상단 URL/ID 복원 검증
- rememberPass true/false에 따른 PASS 복원 분기 검증
- 자동완성 포커스/타이핑/선택/삭제 동작 검증
- localStorage 비활성화 환경에서 graceful fallback 검증

### 6.3 회귀 테스트
- 기존 `/api/tables`, `/api/start`, `/api/retry` 요청 payload가 변경되지 않는지 확인
- 다크모드/반응형 레이아웃에서 상단 공통 영역 UI 깨짐 여부 확인

---

## 7. 단계별 롤아웃 제안
1. Phase 1: 필드별 최근 이력 자동완성(민감정보 제외)
2. Phase 2: 상단 공통 DB URL/ID/PASS + rememberPass
3. Phase 3: 삭제 UX 고도화 및 접근성 개선(키보드 내비게이션)


### tasks.md

# 작업 목록 (Tasks) - v14

## 목표: Web UI 최근 입력 자동완성 + 상단 DB URL/ID/PASS 영속 복원

### 1. 설계/문서화 (Design & Documentation)
- [x] `docs/v14/prd.md` 작성 및 보완
- [x] `docs/v14/spec.md` 작성
- [x] `docs/v14/tasks.md` 작성

### 2. 프론트엔드 구현 (Implementation - Web UI)
- [x] `internal/web/templates/index.html`
  - [x] 상단 공통 연결 정보 섹션(DB URL/ID/PASS, 비밀번호 기억, 초기화 버튼) 추가
  - [x] 자동완성 제안 UI 컨테이너(또는 datalist) 배치
  - [x] 필드별 이력 삭제/전체 삭제 액션 UI 추가
- [x] `internal/web/templates/chart.js`
  - [x] localStorage 키/모델 정의 (`history:*`, `sharedConnection`)
  - [x] 필드별 최근 이력 save/load/clear 유틸 구현 (중복 제거, 최근순, N개 제한)
  - [x] 민감 필드 제외 로직 구현 (`password/token/secret`)
  - [x] 상단 공통 DB URL/ID/PASS 복원 및 저장 로직 구현
  - [x] PASS 저장 옵트인(`rememberPass`) 분기 처리
  - [x] 자동완성 노출/선택/키보드 내비게이션 구현
  - [x] localStorage 예외 처리(graceful fallback) 추가

### 3. 백엔드/호환성 점검 (Compatibility)
- [x] `internal/web/server.go` 및 기존 API 계약 점검
  - [x] 프론트 변경 후도 기존 요청 JSON 필드와 바인딩이 깨지지 않는지 확인
  - [x] 세션/웹소켓/다운로드 흐름 영향 없음 확인

### 4. 테스트 (Testing)
- [x] 단위 테스트
  - [x] 이력 정렬/중복 제거/개수 제한 검증
  - [x] 민감정보 미저장 및 rememberPass 분기 검증
- [x] UI/통합 테스트
  - [x] 새로고침/재접속 시 상단 DB URL/ID 복원 검증
  - [x] rememberPass true 시 PASS 복원, false 시 미복원 검증
  - [x] 자동완성 표시/선택/삭제 시나리오 검증
  - [x] localStorage 미지원 환경 fallback 검증
- [x] 회귀 테스트
  - [x] `go test ./...` 실행

### 5. 문서 업데이트 (Documentation)
- [x] `README.md` 업데이트
  - [x] Web UI 입력 이력 자동완성 기능 설명 추가
  - [x] 상단 DB URL/ID/PASS 복원 및 비밀번호 기억(옵트인) 동작 안내
  - [x] 개인정보/보안 주의사항(공용 PC) 명시


---
## <a name="v15"></a> v15

### prd.md

# Product Requirements Document (PRD) - v15

## 1. 개요 (Overview)
현재 Web UI는 단일 사용자 또는 익명 환경을 가정하고 있어, 다수의 사용자가 동시에 또는 독립적으로 접근하여 사용할 때 마이그레이션 연결 정보(DB URL, ID, Password)와 작업 내역이 혼재되는 문제가 있습니다.
본 기능은 **사용자 로그인 기능을 도입하여 사용자별로 DB 접속 정보를 안전하게 저장/불러오기 하고, 본인의 과거 마이그레이션 작업 내역을 독립적으로 조회**할 수 있는 환경을 제공하는 것을 목표로 합니다. 추가로 관리자를 위한 CLI 기반의 계정 관리 기능도 포함합니다.

## 2. 문제점 (Problem Statement)
- 다중 사용자가 이용할 경우, 이전에 입력된 DB 연결 정보(URL, ID, PASS)가 다른 사용자에게 노출될 위험이 있습니다.
- 사용자별로 자주 사용하는 DB 연결 정보가 다름에도 불구하고, 이를 개별적으로 저장하고 불러오는 기능이 없어 매번 재입력해야 합니다.
- 과거 마이그레이션 실행 내역이 전체 공유되거나 추적이 어려워, 개별 사용자가 본인의 작업 히스토리를 확인하고 재사용하기 어렵습니다.
- 사용자 계정의 비밀번호 분실 등의 예외 상황을 처리할 수 있는 중앙 관리(Admin) 수단이 부재합니다.

## 3. 목표 (Goals)
- 사용자 인증(로그인) 시스템을 구축하여 개별 워크스페이스(세션)를 분리합니다.
- 사용자별로 자주 사용하는 DB 접속 정보(URL, ID, PASS)를 프로필에 저장하고, 작업 시 쉽게 불러올 수 있도록 합니다.
- 사용자 본인이 실행했던 과거 마이그레이션 작업 내역을 목록 형태로 제공하여 언제든 다시 열람할 수 있도록 합니다.
- 시스템 관리자가 터미널 환경(CLI)에서 사용자 계정을 제어하고 비밀번호를 초기화할 수 있는 별도의 관리 도구를 제공합니다.

## 4. 범위 (Scope)
### 4.1 In Scope
1. **로그인/로그아웃 기능**:
   - ID/Password 기반의 기본 인증
   - 세션(Session) 또는 JWT 토큰을 이용한 로그인 상태 유지
2. **사용자별 DB 접속 정보 관리**:
   - 마이그레이션 연결 정보(DB URL, ID, Password)를 사용자 프로필 하위에 다수 저장(Credential 저장소 역할)
   - 저장된 접속 정보를 목록에서 선택하여 메인 화면에 자동 완성/불러오기 기능
3. **사용자별 작업 내역 조회 (History)**:
   - 과거 실행한 마이그레이션 내역(성공/실패 여부, 날짜, 소스/타겟 정보 등)을 사용자 계정에 귀속시켜 저장
   - Web UI 내 '내 작업 내역(My History)' 페이지/섹션 추가
   - 기존 작업 내역 클릭 시 해당 설정으로 폼 데이터 채우기 (재작업 용이)
4. **관리자용 CLI (Admin Shell) 기능**:
   - 서버 환경(터미널)에서 구동 가능한 관리자용 커맨드 제공
   - 사용자 목록 조회, 계정 수동 생성 및 삭제
   - 사용자 비밀번호 강제 초기화(재설정) 기능

### 4.2 Out of Scope
- OAuth2(구글, 깃허브 등) 소셜 로그인 연동 (추후 페이즈 고려)
- Web UI 기반의 복잡한 사용자 그룹(Role-based Access Control) 관리 및 관리자 대시보드
- 여러 사용자 간의 프로젝트 공유 기능

## 5. 사용자 시나리오 (User Scenarios)
1. **접속 및 로그인**: 사용자가 Web UI에 접속하면 로그인 화면이 표시됩니다. 계정을 입력하고 로그인하여 메인 화면으로 진입합니다.
2. **DB 정보 저장 및 불러오기**:
   - 사용자는 "DB 연결 정보 관리" 메뉴에서 본인이 자주 쓰는 소스/타겟 DB 정보(URL, 계정, 비밀번호)를 이름표(Alias)와 함께 저장해 둡니다.
   - 마이그레이션 화면에서 '저장된 정보 불러오기' 드롭다운을 통해 설정값을 원클릭으로 입력 필드에 채웁니다.
3. **마이그레이션 실행 및 내역 조회**:
   - 설정을 마치고 마이그레이션을 실행합니다.
   - 작업 완료 후 "내 작업 내역" 탭으로 이동하면 방금 실행한 작업이 목록 최상단에 추가된 것을 확인합니다.
   - 며칠 뒤, 동일한 내역을 다시 실행하기 위해 내 작업 내역에서 해당 항목의 "이 설정으로 다시 실행" 버튼을 클릭합니다.
4. **관리자의 계정 관리 (CLI)**:
   - 한 사용자가 비밀번호를 잊어버려 시스템 관리자에게 초기화를 요청합니다.
   - 관리자는 서버의 터미널 환경에서 CLI 명령어(`users reset-password ...`)를 실행하여 사용자의 비밀번호를 새로운 값으로 초기화하고 이를 사용자에게 안내합니다.

## 6. 상세 요구사항 (Detailed Requirements)
1. **사용자 인증 (Authentication)**
   - 비밀번호는 서버에 평문으로 저장하지 않고 단방향 암호화(Bcrypt 등) 해시로 저장해야 합니다.
   - 인증 방식은 세션 쿠키 혹은 JWT를 사용하며, 일정 시간 미사용 시 자동 로그아웃 처리를 합니다.
2. **DB 접속 정보 저장소 (Credential Store)**
   - 스키마 설계 시 `users` 테이블과 1:N 관계를 가지는 `db_credentials` (가칭) 테이블을 구성합니다.
   - 저장 항목: 프로필명(Alias), DB 타입(MySQL, PG 등), URL(Host/Port), Username, Password(암호화 저장 권장).
3. **마이그레이션 이력 관리 (History)**
   - `migration_history` 테이블에 `user_id` 컬럼을 추가/연결하여 소유자를 구분합니다.
   - 작업 일시, 소스/타겟 정보, 옵션, 결과(성공/실패), 로그 요약 정보가 포함되어야 합니다.
4. **관리자용 CLI (Admin Shell)**
   - 앱 구동 시 서브 커맨드 형태 또는 별도의 관리자용 플래그를 통해 Admin 기능을 지원합니다.
   - 지원해야 할 필수 명령어:
     - `users list`: 등록된 전체 사용자 목록 조회 (ID, 생성일시 등)
     - `users add <username> <password>`: 신규 사용자 수동 생성
     - `users reset-password <username> <new_password>`: 특정 사용자 비밀번호 강제 변경
     - `users delete <username>`: 사용자 계정 삭제 (연결된 접속 정보 및 마이그레이션 이력의 Cascade 삭제 처리 필요)
5. **UI/UX 추가 사항**
   - 상단 네비게이션(GNB)에 "로그인/로그아웃", "내 정보(접속 정보 관리)", "작업 내역" 메뉴 추가.
   - 폼 입력란 근처에 "저장된 접속 정보 불러오기" 위젯 배치.

## 7. 비기능 요구사항 (Non-functional Requirements)
- **보안**: 다른 사용자의 접속 정보나 작업 내역에 접근할 수 없도록 철저한 권한 체크(Authorization)가 필요합니다. 사용자의 DB Password는 DB에 저장 시 대칭키 암호화(AES 등)를 거쳐 저장하여 탈취 위험을 최소화하는 것을 권장합니다.
- **성능**: 마이그레이션 내역이 많아질 것을 대비해 내역 조회 시 페이지네이션(Pagination) 기능을 적용해야 합니다.
- **호환성**: 기존 단일 사용자 모드에서 저장되었던 상태 데이터의 마이그레이션 혹은 하위 호환을 고려해야 합니다.

## 8. 성공 지표 (Success Metrics)
- 로그인 사용자 수 및 1인당 저장된 접속 정보(Credential) 평균 개수.
- 접속 정보 불러오기 위젯 사용률 증가(직접 입력 대비 비율).
- 과거 작업 내역을 통한 "재실행" 기능의 전환율.

## 9. 수용 기준 (Acceptance Criteria)
1. 사용자는 계정 생성(또는 발급) 후 로그인/로그아웃이 가능해야 합니다.
2. 로그인하지 않은 사용자는 메인 마이그레이션 기능에 접근할 수 없어야 합니다.
3. 사용자는 여러 개의 DB 접속 정보를 저장하고, 목록에서 선택하여 폼에 자동 반영할 수 있습니다.
4. 사용자는 본인이 실행한 과거 작업 내역만 볼 수 있으며, 타인의 내역은 볼 수 없습니다.
5. 저장된 DB Password는 평문으로 DB(내부 저장소, 예: SQLite)에 저장되지 않아야 합니다(암호화).
6. 시스템 관리자는 제공되는 CLI 명령어(`users add`, `users reset-password` 등)를 통해 터미널에서 신규 사용자를 추가하거나 기존 사용자의 비밀번호를 성공적으로 초기화할 수 있어야 합니다.

## 10. 리스크 및 대응 (Risks & Mitigations)
- **리스크**: DB 접속 정보(비밀번호)의 유출 위험.
  - **대응**: 애플리케이션 레벨에서 설정된 마스터 키를 이용해 비밀번호를 양방향 암호화하여 저장하고, 메모리상에서만 복호화하여 사용합니다.
- **리스크**: 기존 싱글 유저 환경에 익숙한 사용자의 불편함.
  - **대응**: 최초 실행 시 단일 계정 자동 생성(admin/admin) 혹은 "로컬 전용 무인증 모드" 설정 플래그 제공을 고려합니다.

## 11. 오픈 이슈 (Open Questions)
- 싱글 데스크탑용 CLI 환경과 서버 환경을 동시 지원할 경우, 로컬 실행 시 강제 로그인 적용 여부? (flag 옵션으로 auth-enable 여부를 제어하는 방향 검토)
- DB 접속 정보의 암복호화를 위한 마스터 키(Secret Key)는 어떻게 주입/관리할 것인가? (환경 변수 `.env` 처리 등)

### spec.md

# 기술 사양서 (Technical Specifications) - v15

## 1. 아키텍처 개요
v15는 기존 단일 사용자 기반 Web UI를 **사용자 인증 기반 멀티 사용자 구조**로 확장한다. 핵심 목표는 다음 4가지다.

1. 로그인/로그아웃 및 세션 검증
2. 사용자별 DB 접속정보(Credential) 안전 저장/조회
3. 사용자별 마이그레이션 이력 조회/재실행
4. 관리자 CLI를 통한 계정 라이프사이클 관리

구현은 기존 구조를 최대한 유지하며 `internal/web`, `internal/db`, `internal/migration`, `internal/config`에 인증/권한 계층을 추가하는 방향으로 진행한다.

---

## 2. 컴포넌트 설계

### 2.1 서버 컴포넌트
- `internal/web/server.go`
  - 로그인/로그아웃/세션 체크 API 추가
  - 인증 미들웨어(`requireAuth`) 도입
  - 사용자별 Credential/History API 엔드포인트 추가
- `internal/db/db.go`
  - 사용자, 자격증명, 이력 테이블 생성/마이그레이션 로직 추가
  - 사용자 스코프 CRUD 메서드 추가
- `internal/migration`
  - 작업 실행 시 `user_id` 컨텍스트를 수집하여 history 저장
- `main.go`
  - Admin CLI 서브커맨드 진입점 추가 (`users list/add/reset-password/delete`)

### 2.2 클라이언트 컴포넌트
- `internal/web/templates/index.html`
  - 비인증 상태에서 로그인 UI 렌더링
  - 인증 후 GNB: 내 정보, 내 작업 내역, 로그아웃
  - 저장된 접속 정보 불러오기 드롭다운/모달
- `internal/web/templates/chart.js`
  - 인증 상태 확인 및 세션 만료 처리
  - Credential 목록 조회/적용/저장/삭제
  - My History 조회/페이지네이션/재실행 입력값 반영

---

## 3. 데이터 모델 및 스키마

### 3.1 `users`
- `id` INTEGER PK AUTOINCREMENT
- `username` TEXT UNIQUE NOT NULL
- `password_hash` TEXT NOT NULL (bcrypt)
- `is_admin` BOOLEAN NOT NULL DEFAULT 0
- `created_at` DATETIME NOT NULL
- `updated_at` DATETIME NOT NULL

인덱스:
- `ux_users_username (username)`

### 3.2 `db_credentials`
- `id` INTEGER PK AUTOINCREMENT
- `user_id` INTEGER NOT NULL FK -> `users(id)` ON DELETE CASCADE
- `alias` TEXT NOT NULL
- `db_type` TEXT NOT NULL
- `host` TEXT NOT NULL
- `port` INTEGER NULL
- `database_name` TEXT NULL
- `username` TEXT NOT NULL
- `password_enc` TEXT NOT NULL (AES-GCM 암호문 + nonce 포함 포맷)
- `created_at` DATETIME NOT NULL
- `updated_at` DATETIME NOT NULL

인덱스:
- `ix_db_credentials_user_id (user_id)`
- `ux_db_credentials_user_alias (user_id, alias)`

### 3.3 `migration_history`
- 기존 컬럼 + `user_id` INTEGER NOT NULL FK -> `users(id)` ON DELETE CASCADE
- `status` TEXT (`success`/`failed`)
- `source_summary`, `target_summary`, `options_json`, `log_summary`
- `created_at` DATETIME NOT NULL

인덱스:
- `ix_migration_history_user_created_at (user_id, created_at DESC)`

### 3.4 마이그레이션 전략
1. 신규 설치: 위 3개 테이블을 최신 스키마로 생성
2. 기존 설치 업그레이드:
   - `users` 생성 + 기본 관리자 계정 정책 적용(초기 비밀번호 강제 변경 권장)
   - `migration_history`에 `user_id` nullable로 추가 후 데이터 백필
   - 백필 이후 `NOT NULL` 제약 적용
3. 롤백: 스키마 롤백보다는 forward-only 정책 권장

---

## 4. 인증/인가 설계

### 4.1 인증 방식
- 기본: 서버 세션 쿠키 기반
- 쿠키 속성:
  - `HttpOnly=true`
  - `Secure=true` (TLS 환경)
  - `SameSite=Lax`
  - 만료: idle timeout (예: 30분) + absolute timeout (예: 24시간)

### 4.2 패스워드 정책
- 저장: bcrypt(hash cost 기본 10~12)
- 최소 길이(예: 8자 이상) 검증
- 관리자 CLI reset 시 임시 비밀번호 발급 가능

### 4.3 인가 규칙
- 모든 business API는 `requireAuth` 필수
- Credential/History 조회/수정/삭제는 `WHERE user_id = session.user_id` 강제
- Admin CLI는 웹 인증과 별개이며 로컬 실행 권한을 전제로 함

---

## 5. 암호화 및 비밀정보 처리

### 5.1 DB 비밀번호 저장
- `db_credentials.password_enc`는 AES-GCM으로 암호화 저장
- 마스터 키는 환경 변수(`DBM_MASTER_KEY`) 또는 설정 파일에서 주입
- 키가 없으면 서버 시작 실패(fail-fast)

### 5.2 메모리 처리
- 복호화된 비밀번호는 연결 직전 최소 범위에서만 사용
- 로깅/에러 메시지에 비밀번호 출력 금지

### 5.3 감사 로그
- 로그인 성공/실패, 비밀번호 초기화, 사용자 삭제는 구조화 로그로 기록
- 로그에 민감정보(원문 비밀번호, 암호문 전문) 포함 금지

---

## 6. API 사양

### 6.1 인증 API
- `POST /api/auth/login`
  - req: `{ "username": "...", "password": "..." }`
  - res: `200 { "ok": true, "user": {"id":1,"username":"..."} }`
- `POST /api/auth/logout`
  - 세션 무효화
- `GET /api/auth/me`
  - 로그인 사용자 정보 반환

### 6.2 Credential API
- `GET /api/credentials`
- `POST /api/credentials`
- `PUT /api/credentials/:id`
- `DELETE /api/credentials/:id`
- 공통: 본인 소유 데이터만 접근 가능

### 6.3 My History API
- `GET /api/history?page=1&pageSize=20`
  - 사용자 본인 이력만 반환
- `GET /api/history/:id`
  - 상세 조회 (권한 체크)
- `POST /api/history/:id/replay`
  - 이력의 설정값을 현재 입력 폼에 복원할 수 있는 payload 반환

---

## 7. UI/UX 사양

### 7.1 로그인 게이트
- 미인증 시 메인 마이그레이션 화면 대신 로그인 폼 노출
- 로그인 성공 시 기존 메인 화면 렌더
- 세션 만료 시 토스트 + 로그인 화면 리다이렉트

### 7.2 상단 네비게이션(GNB)
- 메뉴: `내 정보(접속정보)`, `내 작업 내역`, `로그아웃`
- 현재 사용자명 표시

### 7.3 접속정보 관리
- 별칭(alias) 기반 목록
- 생성/수정/삭제 모달 제공
- "불러오기" 클릭 시 소스/타겟 폼 자동 채움

### 7.4 내 작업 내역
- 열: 실행일시, 소스/타겟 요약, 결과 상태
- 페이지네이션
- "이 설정으로 다시 실행" 액션 제공

---

## 8. 관리자 CLI 사양

### 8.1 커맨드 구조
- `go-db-migration users list`
- `go-db-migration users add <username> <password>`
- `go-db-migration users reset-password <username> <newPassword>`
- `go-db-migration users delete <username>`

### 8.2 동작 규칙
- 출력은 표준출력(성공), 오류는 표준에러 + non-zero exit code
- `delete`는 연관 Credential/History cascade 삭제
- `reset-password`는 대상 유저 존재 여부 검증 후 해시 갱신

---

## 9. 테스트 전략

### 9.1 단위 테스트
- 비밀번호 해시/검증 유틸
- AES-GCM 암복호화 유틸
- 권한 필터(`user_id` 스코프) 검증

### 9.2 통합 테스트
- 로그인/로그아웃/세션 만료 시나리오
- 사용자 A/B 분리(credential, history 상호 비가시성)
- history pagination 및 replay payload 검증
- admin CLI 명령 정상/예외 케이스

### 9.3 보안 회귀
- 인증 없이 보호 API 접근 시 401
- 타 사용자 리소스 접근 시 403/404
- 민감정보 로그 노출 여부 점검

---

## 10. 롤아웃/운영
1. DB 스키마 마이그레이션 적용
2. 마스터 키 설정 및 배포
3. 초기 관리자 계정 생성
4. 기능 플래그(`auth-enabled`)로 단계적 오픈 고려
5. 모니터링:
   - 로그인 실패율
   - 세션 만료율
   - credential/historical API 오류율

---

## 11. 오픈 이슈 및 결정 필요사항
- 로컬 단일 사용자 모드 유지 여부(`--auth-enabled=false`) 확정 필요
- 초기 관리자 계정 생성 정책(환경변수 vs 첫 실행 인터랙션) 확정 필요
- 마스터 키 로테이션 전략(다중 키 버전 관리) 설계 필요


### tasks.md

# 작업 목록 (Tasks) - v15

## 목표: 인증 기반 멀티유저 Web UI + 사용자별 접속정보/이력 분리 + 관리자 CLI

### 1. 설계/문서화 (Design & Documentation)
- [x] `docs/v15/prd.md` 작성 및 보완
- [x] `docs/v15/spec.md` 작성
- [x] `docs/v15/tasks.md` 작성
- [x] `README.md` 기능 섹션/플래그/운영 가이드 업데이트
  - [x] 인증 모드 개요 및 로그인 흐름 추가
  - [x] 관리자 CLI(`users list/add/reset-password/delete`) 사용 예시 추가
  - [x] 비밀키(`DBM_MASTER_KEY`) 설정 가이드 추가

### 2. DB 스키마/저장소 구현 (Database & Repository)
- [x] `internal/db/db.go`
  - [x] `users` 테이블 생성/마이그레이션 추가
  - [x] `db_credentials` 테이블 생성/마이그레이션 추가
  - [x] `migration_history.user_id` 스키마 확장 및 인덱스 추가
  - [x] 업그레이드 백필 로직(user_id 매핑) 구현
- [x] 사용자 저장소 메서드 구현
  - [x] `CreateUser`, `GetUserByUsername`, `ListUsers`, `DeleteUser`, `ResetPassword`
- [x] 접속정보 저장소 메서드 구현
  - [x] `CreateCredential`, `ListCredentialsByUser`, `UpdateCredential`, `DeleteCredential`
- [x] 이력 저장소 메서드 구현
  - [x] `InsertHistory(userID, ...)`, `ListHistoryByUser(page,pageSize)`, `GetHistoryByID(userID,id)`

### 3. 보안 유틸 구현 (Security)
- [x] 비밀번호 해시 유틸 추가
  - [x] bcrypt 해시/검증 함수
  - [x] 최소 길이/정책 검증
- [x] Credential 비밀번호 암복호화 유틸 추가
  - [x] AES-GCM 암호화/복호화 구현
  - [x] nonce/포맷 직렬화 규칙 정의
  - [x] 키 누락 시 fail-fast 처리
- [x] 민감정보 로깅 차단
  - [x] 구조화 로그 필드 점검(원문 비밀번호/암호문 미노출)

### 4. 인증/인가 서버 구현 (Web Backend)
- [x] `internal/web/server.go`
  - [x] `POST /api/auth/login`
  - [x] `POST /api/auth/logout`
  - [x] `GET /api/auth/me`
  - [x] 인증 미들웨어(`requireAuth`) 적용
- [x] 보호 API 사용자 스코프 강제
  - [x] credentials CRUD에서 `user_id` 소유권 체크
  - [x] history 조회/상세/재실행에서 `user_id` 소유권 체크
- [x] 세션 정책 적용
  - [x] 쿠키(HttpOnly, SameSite, Secure) 설정
  - [x] idle/absolute timeout 처리

### 5. 마이그레이션 실행 경로 연동 (Migration Flow)
- [x] `internal/migration` 연동
  - [x] 실행 컨텍스트에 `user_id` 주입
  - [x] 실행 완료 시 사용자 귀속 이력 저장
  - [x] retry/replay 경로의 사용자 권한 검증

### 6. 프론트엔드 구현 (Web UI)
- [x] `internal/web/templates/index.html`
  - [x] 로그인 화면/폼 추가
  - [x] 인증 후 GNB(내 정보/내 작업 내역/로그아웃) 추가
  - [x] 저장된 접속정보 불러오기 UI(드롭다운/모달) 추가
  - [x] 내 작업 내역 섹션(목록/상태/재실행 버튼) 추가
- [x] `internal/web/templates/chart.js`
  - [x] 인증 상태 확인/세션 만료 처리
  - [x] credentials API 연동(조회/생성/수정/삭제/불러오기)
  - [x] history API 연동(페이지네이션/상세/재실행)
  - [x] 로그인/로그아웃 이벤트 핸들링

### 7. 관리자 CLI 구현 (Admin Shell)
- [x] `main.go` 커맨드 엔트리 추가
  - [x] `users list`
  - [x] `users add <username> <password>`
  - [x] `users reset-password <username> <newPassword>`
  - [x] `users delete <username>`
- [x] 에러 처리/종료코드 표준화
  - [x] 성공 시 0, 실패 시 non-zero
  - [x] 오류 메시지 표준에러 출력

### 8. 테스트 (Testing)
- [x] 단위 테스트
  - [x] bcrypt 해시/검증
  - [x] AES-GCM 암복호화/키 오류
  - [x] user_id 스코프 필터(권한) 검증
- [x] 통합 테스트
  - [x] 로그인/로그아웃/세션 만료 시나리오
  - [x] 사용자 A/B 데이터 격리 검증(credentials/history)
  - [x] history pagination/replay 검증
  - [x] 관리자 CLI 명령 정상/예외 케이스
- [x] 회귀 테스트
  - [x] `go test ./...` 실행

### 9. 운영/릴리즈 준비 (Rollout)
- [x] 배포 전 체크리스트
  - [x] `DBM_MASTER_KEY` 주입 확인
  - [x] 초기 관리자 계정 생성/비밀번호 변경 절차 점검
  - [x] 인증 기능 플래그(`auth-enabled`) 기본값 확정 (`false`)
- [x] 모니터링 지표 추가
  - [x] 로그인 실패율, 세션 만료율
  - [x] credentials/history API 오류율


---
## <a name="v16"></a> v16

### prd.md

# Product Requirements Document (PRD) - v16

## 1. 개요 (Overview)
v15에서 인증 기반 멀티유저 기능은 동작하지만, DB 연결 입력 UX가 `Quick Shared Connection`, `Source/Target 직접 입력`, `저장된 접속정보(내 정보)`로 분산되어 사용자 입장에서 어떤 경로를 우선 사용해야 하는지 직관성이 떨어집니다.
v16의 목표는 **연결 설정 UX를 단일 흐름으로 통합**해, 사용자가 "불러오기 → 확인 → 적용" 순서로 빠르게 작업을 시작하도록 만드는 것입니다.

## 2. 문제점 (Problem Statement)
- 동일한 연결 데이터를 여러 섹션에서 관리해 중복 입력/중복 학습 비용이 발생합니다.
- `Quick`과 `저장된 접속정보`의 역할 경계가 불명확해 사용자 혼란이 발생합니다.
- Source/Target 각각에 맞는 저장 연결을 찾고 적용하는 과정이 길고 우회적입니다.
- 로그인 후 첫 화면에서 핵심 작업(연결 후 마이그레이션)보다 관리 UI가 먼저 보이는 인상을 줍니다.

## 3. 목표 (Goals)
- 연결 입력 경로를 단일화해 초기 진입 시 의사결정 부담을 줄입니다.
- Source(Oracle)/Target(DB) 역할별로 저장 연결을 즉시 불러올 수 있도록 합니다.
- 저장된 연결 목록은 "관리 화면"이 아니라 "폼 입력 보조 도구"로 동작하도록 재배치합니다.
- 기존 인증/권한/이력 기능은 유지하면서 UI 정보구조(IA)만 개선합니다.

## 4. 범위 (Scope)
### 4.1 In Scope
1. **용어/정보구조 정리**
   - `내 정보` 메뉴 명칭을 `저장된 연결`로 통일
   - `Quick Shared Connection`을 기본 접힘 상태의 `최근 입력(선택)` 영역으로 재정의
2. **역할별 연결 불러오기 강화**
   - Source URL 필드 옆 `저장된 연결 불러오기` 버튼
   - Target URL 필드 옆 `저장된 연결 불러오기` 버튼
   - 클릭 시 저장 목록이 역할별(`source`/`target`)로 필터링되어 표시
3. **저장된 연결 패널 동작 개선**
   - 패널 진입 시 기본은 전체 목록
   - 역할별 진입 시 해당 역할 목록만 노출하고 안내 문구 표시
4. **기존 흐름과의 호환**
   - 기존 credential 저장/수정/삭제 API와 데이터 스키마는 변경하지 않음
   - 기존 history/replay 흐름은 변경하지 않음

### 4.2 Out of Scope
- 신규 백엔드 API 추가 (예: 추천/랭킹/태그)
- 대규모 디자인 리뉴얼(테마/컴포넌트 시스템 교체)
- 모바일 전용 별도 화면 설계
- OAuth, RBAC 등 인증 체계 확장

## 5. 사용자 시나리오 (User Scenarios)
1. **소스 연결 불러오기**
   - 사용자가 Source URL 옆 `저장된 연결 불러오기`를 클릭합니다.
   - 저장된 Source(Oracle) 목록만 보이고, 항목 선택 시 Oracle URL/ID/PASS가 즉시 반영됩니다.
2. **타깃 연결 불러오기**
   - 사용자가 Target URL 옆 `저장된 연결 불러오기`를 클릭합니다.
   - Direct 모드가 자동 활성화되고 Target 목록만 보이며, 선택 시 Target DB/URL이 반영됩니다.
3. **수동 입력 + 최근 입력 보조**
   - 사용자는 기본적으로 메인 폼에서 직접 입력합니다.
   - 필요 시 접힌 `최근 입력(선택)`을 펼쳐 직전 입력값을 빠르게 재사용합니다.
4. **저장된 연결 관리**
   - 상단 `저장된 연결` 메뉴에서 전체 목록을 보고 수정/삭제합니다.
   - 관리 후 같은 화면에서 즉시 현재 폼으로 불러옵니다.

## 6. 상세 요구사항 (Detailed Requirements)
1. **IA/카피 정합성**
   - UI 전역에서 `저장된 연결` 용어를 일관되게 사용합니다.
   - `Quick`이라는 용어는 제거하고 `최근 입력(선택)`으로 대체합니다.
2. **역할 기반 필터링**
   - 필터 상태: `all | source | target`
   - Source 필터: `dbType === 'oracle'`
   - Target 필터: `dbType !== 'oracle'`
3. **사용성 요구사항**
   - 역할별 진입 시 빈 목록 메시지를 역할 맞춤형으로 표시합니다.
   - 역할별 진입 시 목록 상단 보조 문구로 현재 필터 상태를 명확히 표시합니다.
4. **상태 동기화**
   - Source credential 적용 시 Oracle 입력 필드 및 최근 입력 동기화 유지
   - Target credential 적용 시 Direct 모드/Target DB 선택 상태와 동기화
5. **접근성/반응형**
   - 필드 옆 액션 버튼은 모바일에서 줄바꿈 레이아웃 지원
   - 버튼 라벨은 동작 목적이 분명해야 함(`저장된 연결 불러오기`)

## 7. 비기능 요구사항 (Non-functional Requirements)
- **안정성**: 기존 API 계약과 데이터 포맷을 깨지 않아야 합니다.
- **성능**: 역할 필터링은 클라이언트 단에서 즉시 처리(추가 API 호출 없음).
- **보안**: 권한/세션 검증 로직은 기존과 동일하게 유지됩니다.
- **회귀 방지**: 로그인/로그아웃, credential CRUD, history replay 경로가 기존처럼 동작해야 합니다.

## 8. 성공 지표 (Success Metrics)
- 연결 설정 완료까지 평균 클릭 수 감소
- 첫 마이그레이션 시작까지 시간(TTFM: Time To First Migration) 단축
- `저장된 연결 불러오기` 사용률 증가
- 연결 관련 사용자 문의(“어디서 불러오나”) 감소

## 9. 수용 기준 (Acceptance Criteria)
1. 로그인 후 상단 메뉴에서 `내 정보`가 아닌 `저장된 연결`이 보여야 합니다.
2. `최근 입력(선택)` 영역은 기본 접힘 상태여야 합니다.
3. Source URL 옆 버튼 클릭 시 Source 연결만 목록에 노출되어야 합니다.
4. Target URL 옆 버튼 클릭 시 Target 연결만 목록에 노출되어야 합니다.
5. Source 연결 적용 시 Oracle URL/Username/Password가 즉시 채워져야 합니다.
6. Target 연결 적용 시 Direct 모드와 Target DB/URL이 즉시 반영되어야 합니다.
7. 전체 목록 진입(`저장된 연결` 메뉴) 시 Source/Target 모두 보여야 합니다.
8. 기존 로그인/이력/마이그레이션 실행 회귀가 없어야 합니다.

## 10. 리스크 및 대응 (Risks & Mitigations)
- **리스크**: 기존 사용자가 Quick 영역 위치 변경에 혼란을 느낄 수 있음
  - **대응**: `최근 입력(선택)` 라벨과 안내 문구 제공, 기본 기능은 유지
- **리스크**: 역할 자동 필터가 의도와 다르게 느껴질 수 있음
  - **대응**: 필터 상태 문구(`Source만 표시`, `Target만 표시`, `전체 목록`)를 명시
- **리스크**: 프런트 상태 동기화 누락으로 값 덮어쓰기 이슈 발생 가능
  - **대응**: credential apply 경로에 대한 UI 회귀 테스트 보강

## 11. 오픈 이슈 (Open Questions)
- `저장된 연결` 패널을 현재 인라인 카드로 유지할지, 모달/사이드 패널로 전환할지
- `최근 입력`에 Target URL/Target DB까지 포함할지(현재는 Oracle 중심)
- 장기적으로 `저장된 연결`에 태그/검색/즐겨찾기 기능이 필요한지

## 12. 기술 전략 보강 (Frontend Modernization Addendum)
v16은 UX 개선을 안정적으로 확장하기 위해 프런트 구현 전략을 다음과 같이 보강합니다.

1. **프런트 스택 전환**
   - 현재 단일 `index.html` + 대형 인라인 스크립트 구조를 `Vite + React + Tailwind CSS` 기반 컴포넌트 구조로 전환합니다.
   - 목적은 스타일·상태·로직의 분리, 재사용성 향상, 변경 리스크 축소입니다.

2. **오프라인 런타임 보장**
   - 실행 시(Node/Vite dev server 없이) `./dbmigrator -web`만으로 동작해야 합니다.
   - 빌드 산출물(`dist`)은 Go 바이너리(또는 실행 디렉토리 정적 파일)로 서빙합니다.
   - CDN 의존 자산(원격 JS/CSS/폰트)은 사용하지 않습니다.

3. **단계적 마이그레이션**
   - 1차는 Step 1(연결 설정) 화면부터 React 컴포넌트로 전환하고, 기존 API 계약은 유지합니다.
   - 이후 Step 2/3, History/Monitoring 영역을 순차적으로 이전합니다.


### spec.md

# 기술 사양서 (Technical Specifications) - v16

## 1. 아키텍처 개요
v16은 v15의 인증/권한/이력 백엔드 구조를 유지하면서, 프런트엔드를 `Vite + React + Tailwind CSS`로 전환한다.
핵심 목표는 다음 3가지다.

1. 연결 UX 단일화(`저장된 연결` 중심)
2. 프런트 코드베이스 모듈화(컴포넌트/상태 분리)
3. 런타임 오프라인 동작 보장(Node 미의존)

---

## 2. 컴포넌트 설계

### 2.1 서버 컴포넌트
- `internal/web/server.go`
  - `/api/*` 기존 계약 유지
  - SPA 정적 파일 서빙 및 라우팅 fallback 처리
- `internal/web/assets` (신규, go:embed 대상)
  - `frontend/dist` 산출물 포함

### 2.2 클라이언트 컴포넌트
- `frontend/` (신규)
  - `React 19 + TypeScript + Vite`
  - `Tailwind CSS` 기반 UI 레이어
  - 인증 상태, credential, history, migration 설정 상태를 컴포넌트 단위로 분리

---

## 3. 디렉터리 구조(안)

```text
frontend/
  index.html
  package.json
  vite.config.ts
  tailwind.config.ts
  postcss.config.js
  src/
    main.tsx
    app/
      App.tsx
      routes.tsx
    features/
      auth/
      credentials/
      history/
      migration/
    shared/
      api/
      ui/
      hooks/
      styles/
```

---

## 4. 빌드/배포/오프라인 전략

### 4.1 빌드 단계
1. `frontend` 의존성 설치 (`npm ci` 혹은 `pnpm i --frozen-lockfile`)
2. `vite build` 수행
3. 생성된 `frontend/dist`를 Go에서 정적 서빙

### 4.2 런타임 단계(오프라인)
- `./dbmigrator -web` 실행 시 내장 정적 파일만으로 UI 제공
- Node, npm, Vite 프로세스가 런타임에 필요하지 않음
- 외부 CDN 의존 없음(폰트/아이콘/스크립트 로컬 번들)

### 4.3 fallback 라우팅
- 브라우저 직접 진입(` / `, `/history` 등) 시 `index.html`로 fallback
- `/api/*`, `/static/*`는 기존 API/정적 라우팅 우선

---

## 5. UI/상태 모델

### 5.1 도메인 상태
- `auth`: 로그인 사용자, 세션 만료 상태
- `connections`: source/target 현재 입력, 최근 입력, 저장된 연결 목록/필터
- `history`: 내 작업 이력 목록/페이지네이션/replay payload
- `migration`: stepper 단계, 선택 테이블, 실행 옵션, 실시간 진행 상태

### 5.2 핵심 UX 규칙
- `저장된 연결`이 연결 데이터의 단일 관리 지점
- Source/Target 필드 옆 `저장된 연결 불러오기`는 역할 기반 필터(`source/target`)를 강제
- `최근 입력(선택)`은 보조 기능이며 기본 접힘 상태

---

## 6. API 계약

v16에서 API 스펙은 v15와 동일하게 유지한다.

- 인증: `/api/auth/login`, `/api/auth/logout`, `/api/auth/me`
- 접속정보: `/api/credentials` CRUD
- 이력: `/api/history`, `/api/history/:id`, `/api/history/:id/replay`
- 마이그레이션: `/api/tables`, `/api/migrate`, `/api/migrate/retry`, `/api/test-target`
- 모니터링: `/api/monitoring/metrics`

---

## 7. 보안/운영 고려사항

1. **민감정보 처리**
   - 브라우저 저장소(localStorage)는 비밀번호 저장을 opt-in으로 제한
   - 저장된 연결 비밀번호는 서버측 암호화 저장 정책(v15) 유지
2. **세션 정책**
   - 기존 쿠키 정책(HttpOnly/SameSite/idle+absolute timeout) 유지
3. **오프라인 운영**
   - 사내망/폐쇄망 환경에서도 바이너리 단독 실행 가능해야 함

---

## 8. 테스트 전략

### 8.1 프런트
- 단위 테스트: 역할 필터, 폼 상태 동기화, replay payload 적용
- 컴포넌트 테스트: Source/Target 불러오기 플로우, 세션 만료 UI

### 8.2 서버/통합
- 기존 `go test ./...` 회귀 유지
- 정적 파일 서빙 및 SPA fallback 라우팅 테스트 추가

### 8.3 E2E(선택)
- 로그인 → 저장된 Source 불러오기 → 테이블 조회
- Target 불러오기 → 연결 테스트 → 마이그레이션 시작

---

## 9. 마이그레이션 전략

1. **Phase 1**
   - `frontend` 스캐폴딩, 빌드 파이프라인, Go 정적 서빙 연결
2. **Phase 2**
   - Step 1(연결 화면) React 전환
3. **Phase 3**
   - Step 2/3, history/monitoring 전환
4. **Phase 4**
   - 기존 템플릿 잔존 코드 정리

---

## 10. 결정 필요사항

- 패키지 매니저 표준(`npm` vs `pnpm`)
- 프런트 테스트 도구(Vitest + Testing Library) 도입 범위
- 기존 `index.html`을 병행 유지할 기간(rollback window)


### tasks.md

# 작업 목록 (Tasks) - v16

## 목표: Vite/React/Tailwind 전환 + 오프라인 런타임 보장 + 연결 UX 단일화

### 1. 설계/문서화 (Design & Documentation)
- [x] `docs/v16/prd.md` 작성
- [x] `docs/v16/spec.md` 작성
- [x] `docs/v16/tasks.md` 작성
- [x] `README.md` 개발/빌드/오프라인 실행 가이드 보강
  - [x] 프런트 빌드 명령 및 산출물 위치
  - [x] 런타임 오프라인 동작 방식(노드 미의존) 설명

### 2. 프런트 프로젝트 스캐폴딩 (Frontend Scaffolding)
- [x] `frontend/` 초기화
  - [x] Vite + React + TypeScript 설정
  - [x] Tailwind CSS + PostCSS 설정
  - [ ] ESLint/기본 스크립트(`dev/build/test`) 정리
- [x] 공통 앱 골격 구성
  - [x] `src/app/App.tsx`
  - [x] `src/shared/api/client.ts` (fetch 래퍼)
  - [x] 전역 스타일/테마 변수 정리

### 3. 오프라인 서빙 경로 구현 (Offline Runtime Path)
- [x] Go 정적 파일 서빙 경로 추가
  - [x] `frontend/dist` 자산 embed 또는 배포 경로 연결
  - [x] `/api/*` 우선 라우팅 + SPA fallback 적용
- [x] 외부 CDN 의존 제거
  - [x] 폰트/아이콘/스크립트 로컬 번들화 확인

### 4. Step 1 UI 전환 (Connection UX)
- [x] Source/Target 연결 화면 React 컴포넌트 전환
  - [x] Source URL 옆 `저장된 연결 불러오기`
  - [x] Target URL 옆 `저장된 연결 불러오기`
  - [x] `최근 입력(선택)` 기본 접힘
- [x] 저장된 연결 패널 UX 정리
  - [x] 필터 상태(`all/source/target`) 명시
  - [x] 역할별 빈 상태 메시지 반영

### 5. 인증/세션/이력 연동 (Auth & History)
- [x] 인증 상태 게이트 연동(`/api/auth/me`)
- [x] 로그인/로그아웃 플로우 React 상태로 이전
- [x] 내 작업 이력 조회/재실행 플로우 이전

### 6. 마이그레이션 실행 영역 이전 (Step 2/3)
- [x] 테이블 선택/옵션 폼 React 전환
- [x] 실행/진행률/요약 카드 React 전환
- [x] WebSocket 수신 및 상태 갱신 로직 이전

### 7. 테스트 (Testing)
- [x] 프런트 단위/컴포넌트 테스트
  - [x] 역할별 불러오기 필터
  - [x] replay payload 폼 반영
  - [x] 세션 만료 처리 UI
- [x] 서버 회귀 테스트
  - [x] `go test ./...`
  - [x] 정적 파일 서빙 + SPA fallback 경로

### 8. 롤아웃/전환 (Rollout)
- [ ] 병행 운영 전략 확정
  - [ ] 구 UI fallback 유지 기간 정의
  - [ ] 롤백 절차 문서화
- [ ] 최종 전환 체크리스트
  - [ ] 오프라인 환경 실행 검증
  - [ ] 인증 모드(`-auth-enabled`) 동작 검증
  - [ ] 주요 사용자 시나리오 E2E 확인


---
## <a name="v17"></a> v17

### prd.md

# Product Requirements Document (PRD) - v17

## 1. 개요 (Overview)
현재 마이그레이션 과정에서 테이블 관련 스크립트(테이블/제약조건/인덱스 등)와 시퀀스 스크립트가 단일 흐름으로 취급되어,
조회 단계와 실행 단계에서 원하는 객체만 선택적으로 다루기 어렵습니다.
v17의 목표는 **테이블 계열 스크립트와 시퀀스 스크립트를 분리하여 조회·검토·마이그레이션을 독립적으로 수행**할 수 있게 하는 것입니다.

## 2. 문제점 (Problem Statement)
- 스키마 조회 결과에서 테이블/시퀀스가 혼합되어 확인 가독성이 떨어집니다.
- “테이블만 먼저 이관 후 시퀀스 별도 반영” 같은 운영 시나리오를 수행하기 어렵습니다.
- 실패 시 재시도 범위를 세밀하게 통제하기 어렵고, 영향도 파악이 늦어집니다.
- Dry-run 검토 시에도 객체 유형별 SQL 검증이 한 번에 섞여 리스크 판단이 어렵습니다.

## 3. 목표 (Goals)
- 테이블 계열과 시퀀스 계열의 **조회(Discovery) 결과를 분리**해 표시합니다.
- 마이그레이션 실행 시 객체 유형별로 **독립 실행**(테이블만/시퀀스만/둘 다)을 지원합니다.
- Dry-run 및 리포트에서 유형별 결과를 구분해 운영 의사결정을 빠르게 만듭니다.
- 기존 전체 마이그레이션 흐름과 하위 호환성을 유지합니다.

## 4. 범위 (Scope)

### 4.1 In Scope
1. **조회 분리**
   - 소스 메타데이터 조회 시 결과를 `table-related`와 `sequence`로 분류합니다.
   - UI/CLI에서 유형별 카운트 및 목록을 분리 표시합니다.
2. **실행 분리**
   - 실행 옵션에 객체 유형 선택을 추가합니다.
   - 실행 모드: `all`(기본), `tables-only`, `sequences-only`.
3. **스크립트 생성 분리**
   - 내부 DDL 생성 파이프라인에서 테이블 계열 SQL과 시퀀스 SQL을 분리 보관합니다.
   - 필요 시 유형별 파일 출력(예: `tables.sql`, `sequences.sql`)을 지원합니다.
4. **결과/리포트 분리**
   - 진행률, 성공/실패, 경고를 유형별로 집계합니다.
   - 실패 재시도 시 유형 단위 재실행이 가능해야 합니다.

### 4.2 Out of Scope
- 뷰/트리거/프로시저 등 추가 객체 유형 세분화
- DB 벤더별 DDL 최적화 로직 대규모 개편
- 권한/계정/시스템 오브젝트 마이그레이션

## 5. 사용자 시나리오 (User Scenarios)
1. **사전 점검 운영자**
   - 운영자는 먼저 `tables-only` dry-run을 수행해 테이블 생성 및 제약조건 충돌 가능성을 확인합니다.
2. **단계적 배포 운영자**
   - 운영자는 1차로 테이블 계열 이관 후 검증을 마치고, 2차 윈도우에 `sequences-only`를 실행합니다.
3. **장애 복구 담당자**
   - 시퀀스 반영 단계에서 실패가 발생하면, 테이블 작업 영향 없이 `sequences-only` 재실행으로 복구합니다.

## 6. 상세 요구사항 (Detailed Requirements)
1. **객체 분류 규칙**
   - `table-related`: table, pk/fk/uk/check, index 등 테이블 종속 DDL.
   - `sequence`: sequence 생성/증분/시작값 관련 DDL.
2. **옵션/파라미터 요구사항**
   - CLI 플래그 또는 UI 선택값으로 실행 대상을 지정할 수 있어야 합니다.
   - 지정이 없으면 기존과 동일하게 `all` 동작이어야 합니다.
3. **Dry-run 분리 출력**
   - Dry-run 시 유형별 SQL 블록과 요약 통계를 각각 제공합니다.
4. **실행 순서 정책**
   - `all` 모드에서는 기본 순서를 `tables -> sequences`로 고정합니다.
   - `sequences-only`는 테이블 생성 단계 없이 독립 실행됩니다.
5. **리포트/로그**
   - 로그 필드에 `object_group`(tables/sequences)를 포함합니다.
   - 최종 리포트에 그룹별 성공/실패/스킵 건수를 표시합니다.
6. **호환성**
   - 기존 플래그/기본 실행 경로를 깨지 않아야 하며, 미지정 시 기존 사용자 경험과 동일해야 합니다.

## 7. 비기능 요구사항 (Non-functional Requirements)
- **안정성:** 그룹 분리 도입이 기존 전체 마이그레이션 성공률을 저해하지 않아야 합니다.
- **관찰성:** 그룹별 로그/리포트로 실패 지점을 즉시 식별 가능해야 합니다.
- **성능:** 분리 집계 추가로 인해 총 실행시간 증가를 최소화합니다.
- **보안:** 기존 연결정보/자격증명 처리 방식은 변경하지 않습니다.

## 8. 성공 지표 (Success Metrics)
- `tables-only`, `sequences-only` 실행 사용률
- 장애 시 평균 복구 시간(MTTR) 단축
- Dry-run 결과 검토 시간 감소
- 객체 유형 혼합으로 인한 운영 문의 건수 감소

## 9. 수용 기준 (Acceptance Criteria)
1. 조회 결과 화면/출력에서 테이블 계열과 시퀀스 계열이 구분되어 보여야 합니다.
2. 사용자는 `all`, `tables-only`, `sequences-only` 중 하나를 선택해 실행할 수 있어야 합니다.
3. `tables-only` 실행 시 시퀀스 DDL이 실행되지 않아야 합니다.
4. `sequences-only` 실행 시 테이블 계열 DDL이 실행되지 않아야 합니다.
5. `all` 실행 시 테이블 계열 후 시퀀스 계열 순으로 처리되어야 합니다.
6. Dry-run/최종 리포트에서 그룹별 통계가 분리되어 표시되어야 합니다.
7. 기존 기본 실행(옵션 미지정)은 기존 동작과 호환되어야 합니다.

## 10. 리스크 및 대응 (Risks & Mitigations)
- **리스크:** 그룹 경계 정의가 DB 벤더별로 달라 분류 오류 가능
  - **대응:** 벤더별 분류 매핑 테스트 케이스 추가 및 회귀 테스트 강화
- **리스크:** 분리 실행 중 의존성 누락(예: 시퀀스 참조)으로 런타임 오류 발생 가능
  - **대응:** `sequences-only` 실행 전 사전 검증/경고 메시지 제공
- **리스크:** 운영자가 모드를 잘못 선택할 가능성
  - **대응:** UI/CLI에 모드 설명 및 기본값(`all`) 안내를 명확히 표기

## 11. 오픈 이슈 (Open Questions)
- 시퀀스 현재값(Last Number/Cache/Order) 동기화 정책을 벤더별로 어디까지 보장할지
- 테이블 계열에 포함될 객체 범위(인덱스/제약조건)를 옵션으로 더 쪼갤지
- 유형별 SQL 아티팩트 파일 보관 정책(실행 후 삭제 vs 영구 보관)


### spec.md

# 기술 사양서 (Technical Specifications) - v17

## 1. 아키텍처 개요
v17은 v16의 마이그레이션 파이프라인을 확장하여 DDL 처리 대상을 `tables` 그룹과 `sequences` 그룹으로 분리한다.
핵심 목표는 다음 3가지다.

1. 조회(Discovery) 결과의 객체 유형 분리
2. 실행(Execute) 단계의 그룹 단위 선택/재시도
3. Dry-run/리포트/로그의 그룹별 관찰성 강화

---

## 2. 도메인 모델 및 분류 규칙

### 2.1 ObjectGroup 정의
- `all`: 전체(기본값)
- `tables`: 테이블 계열
- `sequences`: 시퀀스 계열

### 2.2 분류 규칙
- `tables`
  - table DDL
  - pk/fk/uk/check constraint DDL
  - index DDL
  - 테이블 종속 부가 DDL
- `sequences`
  - sequence 생성 DDL
  - increment/start/cache/order 관련 DDL

### 2.3 내부 데이터 구조(안)
```go
type ObjectGroup string

const (
    ObjectGroupAll       ObjectGroup = "all"
    ObjectGroupTables    ObjectGroup = "tables"
    ObjectGroupSequences ObjectGroup = "sequences"
)

type GroupedScripts struct {
    TablesSQL    []string
    SequencesSQL []string
}

type GroupedStats struct {
    Tables    GroupStats
    Sequences GroupStats
}
```

---

## 3. 컴포넌트 설계

### 3.1 Migration 파이프라인 변경
- 스키마 조회 결과를 `GroupedMetadata` 형태로 저장
- SQL 생성 단계에서 `GroupedScripts` 생성
- 실행기(Executor)는 전달받은 `ObjectGroup`에 따라 수행 대상을 제한

### 3.2 실행 순서 정책
- `all`: `tables -> sequences` 고정
- `tables`: tables 그룹만 실행
- `sequences`: sequences 그룹만 실행

### 3.3 재시도 정책
- 실패 이력에 `object_group` 필드를 저장
- 재시도 시 `retry_group` 미지정이면 원래 그룹 유지
- `all` 실패 후 `sequences`만 선택 재시도 가능

---

## 4. 인터페이스 계약

### 4.1 CLI 플래그
- 신규 플래그(안): `--object-group`
  - 허용값: `all|tables|sequences`
  - 기본값: `all`
- 기존 명령/플래그와 호환되도록 미지정 시 기존 동작 유지

### 4.2 Web UI
- 실행 옵션에 `마이그레이션 대상` 선택 UI 추가
  - 전체
  - 테이블 계열만
  - 시퀀스만
- 조회 결과 패널에서 그룹별 카운트/목록을 분리 표시

### 4.3 API 스키마 확장(안)
- `POST /api/migrate`
  - 요청 필드: `object_group`(optional, default=`all`)
- `POST /api/migrate/retry`
  - 요청 필드: `object_group`(optional)
- 응답/이력에 그룹별 통계 포함
  - `stats.tables.success|failed|skipped`
  - `stats.sequences.success|failed|skipped`

---

## 5. Dry-run / 리포트 / 로깅

### 5.1 Dry-run 출력 규칙
- 출력 섹션을 `TABLES SQL` / `SEQUENCES SQL`로 분리
- 그룹별 SQL 건수 및 예상 영향도 표시

### 5.2 실행 리포트
- 그룹별 처리 건수(success/failed/skipped) 집계
- 최종 요약은 전체 + 그룹별 상세를 함께 표기

### 5.3 구조화 로그
- 모든 주요 이벤트에 `object_group` 필드 포함
- 예시 이벤트
  - `discovery.completed`
  - `script.generated`
  - `migration.started`
  - `migration.statement.failed`
  - `migration.completed`

---

## 6. 호환성 및 마이그레이션 전략

### 6.1 하위 호환성
- `object_group` 미지정 시 기존과 동일한 전체 실행
- 기존 이력 데이터(그룹 필드 없음) 조회 시 `all`로 간주

### 6.2 단계적 적용
1. 내부 파이프라인 그룹 분리(기능 플래그 off)
2. CLI/API에 `object_group` 노출
3. UI 분리 노출 및 리포트 강화
4. 회귀 테스트 완료 후 기본 활성화

---

## 7. 오류 처리/검증

### 7.1 입력 검증
- 허용되지 않은 `object_group` 값은 400 에러
- 에러 메시지에 허용값 명시

### 7.2 의존성 경고
- `sequences-only` 실행 시 대상 시퀀스의 참조 테이블 미존재 가능성 경고 로그 출력
- dry-run 단계에서 경고 목록에 포함

### 7.3 실패 격리
- 한 그룹 실패가 다른 그룹 실행 여부에 영향을 주는 정책을 명확화
  - 기본: `all`에서 tables 실패 시 sequences 단계 진입하지 않음
  - 옵션 정책은 후속 버전에서 확장

---

## 8. 테스트 전략

### 8.1 단위 테스트
- 분류기(Classifier)
  - table/constraint/index/sequence 분류 정확성
- 실행 선택기(Selector)
  - `all|tables|sequences`별 실행 목록 검증

### 8.2 통합 테스트
- `tables-only` 실행 시 sequence SQL 미실행 보장
- `sequences-only` 실행 시 table SQL 미실행 보장
- `all` 실행 시 순서 `tables -> sequences` 보장

### 8.3 회귀 테스트
- `object_group` 미지정 경로에서 기존 결과 동일성 검증
- 기존 이력 replay 시 동작 호환성 검증

---

## 9. 운영 관측 지표
- 실행 모드 사용률: `all/tables/sequences`
- 그룹별 실패율 및 재시도 성공률
- `sequences-only` 복구 시나리오 MTTR
- Dry-run 검토 소요 시간 변화

---

## 10. 결정 필요사항
- `--object-group` 네이밍 확정 여부(`--target-group` 대안)
- `all` 모드에서 tables 실패 시 sequences 계속 진행 옵션 제공 여부
- 유형별 SQL 아티팩트(`tables.sql`, `sequences.sql`) 기본 보관 정책


### tasks.md

# 작업 목록 (Tasks) - v17

## 목표: 테이블 계열/시퀀스 계열 분리 조회 + 그룹 단위 실행/재시도 + 리포트 분리

### 1. 설계/문서화 (Design & Documentation)
- [x] `docs/v17/prd.md` 작성
- [x] `docs/v17/spec.md` 작성
- [x] `docs/v17/tasks.md` 작성
- [x] `README.md` 업데이트
  - [x] 객체 그룹 실행 모드(`all/tables/sequences`) 설명 추가
  - [x] 신규 옵션/요청 필드(`--object-group`, `object_group`) 문서화
  - [x] Dry-run/리포트의 그룹별 출력 예시 추가

### 2. 도메인 모델/분류기 구현 (Domain & Classifier)
- [x] `ObjectGroup` 타입/상수 도입
  - [x] `all`, `tables`, `sequences` 정의
  - [x] 기본값 `all` 처리 경로 반영
- [x] DDL 분류기 구현
  - [x] table/constraint/index 등 `tables` 분류
  - [x] sequence 관련 DDL `sequences` 분류
  - [x] 분류 불가/애매 케이스 경고 로깅
- [x] 그룹별 컨테이너 구조 도입
  - [x] `GroupedMetadata`(조회 결과)
  - [x] `GroupedScripts`(생성 SQL)
  - [x] `GroupedStats`(집계 결과)

### 3. 조회/스크립트 생성 파이프라인 분리 (Discovery & Script Generation)
- [x] 메타데이터 조회 결과를 그룹별로 분리 저장
- [x] SQL 생성 시 그룹별 산출물 분리
  - [x] `tables.sql` 성격의 SQL 묶음
  - [x] `sequences.sql` 성격의 SQL 묶음
- [x] Dry-run 출력 섹션 분리
  - [x] `TABLES SQL`
  - [x] `SEQUENCES SQL`

### 4. 실행기(Executor) 그룹 선택 로직 (Execution)
- [x] 실행 입력에 `object_group` 반영
- [x] 실행 모드별 대상 선택 구현
  - [x] `all`: `tables -> sequences` 순서 고정 
  - [x] `tables`: tables만 실행
  - [x] `sequences`: sequences만 실행
- [x] 실패 격리 정책 반영
  - [x] 기본 정책: `all`에서 tables 실패 시 sequences 미진입
  - [x] 정책 이벤트 구조화 로그로 기록

### 5. API/웹 서버 계약 확장 (API Contract)
- [x] `POST /api/migrate` 요청 바디에 `object_group`(optional) 지원
- [x] `POST /api/migrate/retry` 요청 바디에 `object_group`(optional) 지원
- [x] 입력 검증
  - [x] 허용값 외 입력 시 400 반환
  - [x] 허용값 안내 메시지 포함
- [x] 응답/이력 모델 확장
  - [x] 그룹별 통계(`stats.tables`, `stats.sequences`) 노출
  - [x] 이력에 `object_group` 저장(미존재 legacy는 `all`로 해석)

### 6. Web UI 반영 (Frontend)
- [x] 실행 옵션에 `마이그레이션 대상` 선택 UI 추가
  - [x] 전체
  - [x] 테이블 계열만
  - [x] 시퀀스만
- [x] 조회 결과 패널 그룹 분리
  - [x] 그룹별 카운트 표시
  - [x] 그룹별 목록/접기-펼치기 UX 정리
- [x] Dry-run/결과 요약 카드 분리
  - [x] 그룹별 성공/실패/스킵 통계 표시

### 7. 로그/리포트/관측성 강화 (Observability)
- [x] 주요 이벤트 로그에 `object_group` 필드 추가
  - [x] `discovery.completed`
  - [x] `script.generated`
  - [x] `migration.started`
  - [x] `migration.statement.failed`
  - [x] `migration.completed`
- [x] 최종 리포트 포맷 확장
  - [x] 전체 요약 + 그룹별 상세 병기
- [x] 운영 지표 수집 항목 추가
  - [x] 모드별 사용률(`all/tables/sequences`)
  - [x] 그룹별 실패율/재시도 성공률

### 8. 테스트 (Testing)
- [x] 단위 테스트
  - [x] 분류기: table/constraint/index/sequence 분류 정확성
  - [x] 선택기: `all|tables|sequences`별 실행 목록 검증
  - [x] 검증기: 잘못된 `object_group` 입력 검증
- [x] 통합 테스트
  - [x] `tables-only`에서 sequence SQL 미실행 검증
  - [x] `sequences-only`에서 table SQL 미실행 검증
  - [x] `all` 순서(`tables -> sequences`) 검증
- [x] 회귀 테스트
  - [x] 옵션 미지정(`all` 기본) 기존 동작 동일성
  - [x] 기존 이력 replay 호환성
  - [x] `go test ./...` 통과

### 9. 롤아웃/운영 가이드 (Rollout)
- [x] 단계적 릴리즈 계획 반영
  - [x] 내부 기능 플래그 기반 점진 활성화
  - [x] 운영팀 대상 모드 선택 가이드 배포
- [x] 장애 대응 플레이북 보강
  - [x] `sequences-only` 복구 절차 문서화
  - [x] 모드 오선택 방지 체크리스트 추가
- [x] 최종 배포 체크
  - [x] dry-run 검토 절차 준수 확인
  - [x] 로그/리포트 대시보드 필드 누락 점검


### rollout.md

# v17 롤아웃/운영 가이드

## 1. 단계적 릴리즈 계획

### 1-1. 제어 포인트

- `DBM_OBJECT_GROUP_UI_ENABLED=true`
  - v16 UI에 `Migration target` 선택기와 그룹별 조회/결과 패널을 노출한다.
- `DBM_OBJECT_GROUP_UI_ENABLED=false`
  - v16 UI를 legacy `all` 모드로 고정한다.
  - 백엔드의 `object_group` 호환성은 유지하지만 일반 운영자는 v17 분리 실행 UI를 보지 않는다.

### 1-2. 권장 활성화 순서

1. 스테이징
   - `DBM_OBJECT_GROUP_UI_ENABLED=true`
   - `all`, `tables`, `sequences` 3개 모드로 dry-run과 실제 실행을 각각 1회 이상 검증한다.
2. 내부 운영자 제한 오픈
   - 운영 담당 서버 또는 내부 접근 가능한 배포 슬롯에서만 `DBM_OBJECT_GROUP_UI_ENABLED=true`
   - 일반 사용자 서버는 `false`로 유지한다.
3. 전체 오픈
   - 운영 지표에서 모드별 실패율과 재시도 성공률이 허용 범위인지 확인한 뒤 전체 서버에 `true`를 반영한다.

## 2. 운영팀 모드 선택 가이드

- `all`
  - 기본값이다.
  - 테이블/데이터 이관과 시퀀스 반영을 한 번에 수행해야 할 때 사용한다.
  - 테이블 단계 실패 시 시퀀스 단계는 자동 스킵된다.
- `tables`
  - 테이블/데이터 경로만 먼저 검증하거나 시퀀스 변경을 의도적으로 제외해야 할 때 사용한다.
  - 대규모 데이터 적재 후 시퀀스는 별도 승인 절차로 분리하려는 운영 절차에 적합하다.
- `sequences`
  - 테이블/데이터 적재가 이미 완료되었고 시퀀스 보정만 따로 반영해야 할 때 사용한다.
  - 복구 작업이나 재시도 작업에 우선 적용한다.

## 3. 장애 대응 플레이북

### 3-1. `sequences-only` 복구 절차

1. 실패 이력에서 대상 작업의 `report_id`, `object_group`, `stats.tables`, `stats.sequences`를 확인한다.
2. 테이블 그룹이 정상 완료되었고 시퀀스 그룹만 실패했는지 판단한다.
3. 실패 원인이 권한, 이름 충돌, 대상 DB 접속 일시 장애인지 로그의 `migration.statement.failed` 필드로 확인한다.
4. 원인 제거 후 `object_group=sequences`로 재실행하거나 History replay로 `sequences` 모드를 재적용한다.
5. 완료 후 리포트에서 `stats.sequences.error_count == 0`인지 다시 확인한다.

### 3-2. 모드 오선택 방지 체크리스트

- 데이터와 DDL을 함께 반영해야 하면 `all`인지 확인한다.
- 시퀀스 반영을 제외하려는 명확한 운영 사유가 없으면 `tables`를 선택하지 않는다.
- 이미 데이터 적재가 끝난 복구 작업이 아니면 `sequences`를 선택하지 않는다.
- Dry-run 결과의 `TABLES SQL` / `SEQUENCES SQL` 섹션이 기대한 범위와 일치하는지 검토한다.

## 4. 최종 배포 체크

- dry-run 결과에서 대상 테이블 수와 시퀀스 수가 운영 변경 요청과 일치하는지 확인한다.
- 완료 리포트와 WebSocket 요약에 `object_group`, `stats.tables`, `stats.sequences`가 모두 노출되는지 확인한다.
- 운영 대시보드 또는 `/api/monitoring/metrics`에서 `migrations.all|tables|sequences` 지표가 수집되는지 확인한다.
- History replay 시 legacy 이력이 `all`로 복원되는지 샘플 1건 이상 확인한다.


---
## <a name="v18"></a> v18

### prd.md

# PRD: UI 테이블 단위 마이그레이션 사용성 개선

## 1. 배경
현재 UI에서 테이블을 하나씩 조회하고 개별적으로 마이그레이션하는 흐름은 다음과 같은 불편이 있다.

- 이미 마이그레이션이 완료된 테이블을 매번 수동으로 구분해야 한다.
- 대상 테이블이 많을수록 중복 작업(재조회/재선택)이 발생한다.
- "무엇을 이미 했는지"에 대한 이력 가시성이 부족해 운영자가 상태를 추적하기 어렵다.

이로 인해 운영 시간이 증가하고, 동일 테이블 재실행 같은 실수 가능성이 높아진다.

## 2. 문제 정의
UI 기반 마이그레이션에서 사용자가 가장 자주 겪는 문제는 다음 두 가지다.

1. **필터링 부재**: 이미 마이그레이션한 테이블을 제외하고 보고 싶어도 바로 걸러지지 않는다.
2. **이력/상태 가시성 부족**: 테이블별 성공/실패/진행 상태를 한 화면에서 파악하기 어렵다.

## 3. 목표
- 사용자가 "미마이그레이션 대상"에 집중할 수 있도록 필터/옵션을 제공한다.
- 테이블별 마이그레이션 이력을 UI에서 즉시 확인할 수 있게 한다.
- 반복 작업과 재실행 실수를 줄여 작업 효율을 향상한다.

## 4. 비목표(Non-goals)
- 마이그레이션 엔진의 핵심 실행 로직(DDL/DML 생성 방식) 변경.
- 신규 DB 타입 지원 추가.
- 대규모 권한/인증 체계 변경.

## 5. 사용자 스토리
- 운영자로서, 이미 완료된 테이블을 제외하고 목록을 보고 싶다.
- 운영자로서, 각 테이블의 최근 마이그레이션 결과(성공/실패/시간/소요시간)를 보고 싶다.
- 운영자로서, 실패한 테이블만 빠르게 필터링해서 재시도하고 싶다.

## 6. 기능 요구사항

### FR-1. 상태 기반 필터
- 테이블 목록 상단에 상태 필터를 제공한다.
- 기본값은 `전체`이며, 최소 아래 옵션을 제공한다.
  - `미실행`
  - `성공`
  - `실패`
  - `진행중`
- "이미 마이그레이션 완료 제외"를 위한 빠른 토글(예: `성공 제외`)을 제공한다.

### FR-2. 테이블별 이력 표시
- 목록에서 테이블별로 아래 정보를 표시한다.
  - 최근 실행 상태
  - 최근 실행 시각
  - 실행 소요 시간
  - 누적 실행 횟수
- 상태는 색상/뱃지로 직관적으로 구분한다.

### FR-3. 상세 이력 패널
- 특정 테이블 선택 시 상세 이력(최근 N건)을 확인할 수 있다.
- 이력 항목에는 최소 아래 필드를 포함한다.
  - 실행 시작/종료 시각
  - 결과 상태
  - 처리 건수(가능한 경우)
  - 오류 메시지 요약(실패 시)

### FR-4. 재시도 중심 UX
- `실패만 보기` 필터 제공.
- 실패 항목에서 즉시 재시도 액션을 제공한다.
- 재시도 시 이전 실패 원인(요약)을 확인할 수 있다.

### FR-5. 정렬/검색
- 테이블명 검색 지원.
- 최근 실행 시각, 상태 기준 정렬 지원.

## 7. 데이터 및 이력 모델 요구
- 테이블 단위 마이그레이션 실행 이력을 저장/조회할 수 있어야 한다.
- 최소 저장 필드:
  - table_name
  - status
  - started_at
  - finished_at
  - duration_ms
  - rows_processed(옵션)
  - error_message(옵션)
  - run_id
- UI 조회 성능을 위해 상태/최근시각 기준 인덱싱을 고려한다.

## 8. UX 요구사항
- 상태 필터와 검색창은 목록 상단에 고정 배치.
- 상태 정보는 텍스트+컬러를 함께 사용해 접근성 확보.
- 빈 상태(예: 실패 없음, 미실행 없음) 화면 메시지 제공.
- 로딩/실패 시 스켈레톤, 에러 배너 등 피드백 제공.

## 9. 성공 지표 (KPI)
- 전체 대상 중 완료 테이블 식별 시간 50% 단축.
- 운영자의 수동 재조회 횟수 30% 이상 감소.
- 동일 테이블 불필요 재실행 건수 30% 이상 감소.
- 실패 테이블 재처리 리드타임 20% 단축.

## 10. 수용 기준 (Acceptance Criteria)
1. 사용자는 `성공 제외` 옵션으로 미완료 대상만 즉시 조회할 수 있다.
2. 모든 테이블 행에서 최근 실행 상태/시각/소요시간을 확인할 수 있다.
3. 사용자는 `실패만 보기`로 실패 대상만 필터링하고 재시도할 수 있다.
4. 특정 테이블 클릭 시 최근 이력 N건이 표시된다.
5. 이력이 없는 테이블은 `미실행` 상태로 명확히 표시된다.

## 11. 릴리즈 범위 제안
- **1차(MVP)**: 상태 필터, 성공 제외 토글, 최근 1회 상태 표시.
- **2차**: 상세 이력 패널(최근 N건), 실패 즉시 재시도 UX.
- **3차**: 고급 통계(실패 원인 집계, 테이블별 추세).

## 12. 리스크 및 대응
- 이력 데이터 누락 시 상태 오표시 가능성
  - 대응: 실행 완료 시점 write 보장, 실패 시 재시도 큐와 동기화
- 테이블 수가 많을 때 목록 성능 저하
  - 대응: 서버 사이드 페이징/필터링, 인덱스 최적화
- 상태 정의 불일치(백엔드/프론트)
  - 대응: 공통 enum 계약 및 계약 테스트 추가


### spec.md

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


### tasks.md

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


---
## <a name="v19"></a> v19

### prd.md

# PRD: Direct Insert 대상 사전 행 수 비교 기반 전송 필요성 판단

## 1. 배경
현재 Direct Insert 방식 마이그레이션에서는 Oracle 원본 테이블을 조회한 뒤 대상 DB로 데이터를 전송한다. 이때 대상 DB에 이미 동일한 데이터가 적재되어 있어도 전송이 수행될 수 있어 불필요한 I/O, 처리 시간 증가, 운영 리스크(중복 검증 부담)가 발생한다.

특히 운영자는 "어떤 테이블을 다시 전송해야 하는지"를 사전에 판단하기 어렵고, 이를 위해 수동으로 원본/대상 건수를 비교해야 한다.

## 2. 문제 정의
- 마이그레이션 전 단계에서 **원본(Oracle) 행 수 vs 대상 DB 행 수** 비교가 자동화되어 있지 않다.
- 전송 필요 여부를 판단할 **기준(rule)** 과 **필터링 옵션**이 없어, 이미 동기화된 테이블도 동일하게 처리된다.
- 결과적으로 전체 수행 시간이 늘어나고, 운영자 의사결정이 수동 절차에 의존한다.

## 3. 목표
- 마이그레이션 시작 전 테이블별 행 수 비교를 수행해 **전송 필요 여부를 자동 판정**한다.
- 운영자가 조건 기반으로 대상 테이블을 빠르게 필터링할 수 있도록 기준/필터를 제공한다.
- 불필요한 Direct Insert 작업을 줄여 실행 시간을 단축한다.

## 4. 비목표 (Non-goals)
- 행 수 비교를 넘어선 레코드 단위 정합성 검증(checksum, diff) 제공.
- DDL 스키마 변경 감지/동기화 자동화.
- CDC(변경분 추적) 방식 도입.

## 5. 사용자 스토리
- 운영자로서, 마이그레이션 전에 원본/대상 행 수를 비교해 전송이 꼭 필요한 테이블만 보고 싶다.
- 운영자로서, 행 수가 동일한 테이블은 기본적으로 제외하고 불일치 테이블만 전송하고 싶다.
- 운영자로서, 비교 결과를 기준으로 예외 정책(무조건 전송/건너뛰기)을 선택하고 싶다.

## 6. 기능 요구사항

### FR-1. 사전 행 수 수집
- Direct Insert 모드 실행 전, 선택된 각 테이블에 대해 아래 값을 수집한다.
  - `source_row_count` (Oracle)
  - `target_row_count` (대상 DB)
- 수집 실패 시 상태를 `count_check_failed`로 표기하고 원인 메시지를 기록한다.

### FR-2. 전송 필요 여부 판정 규칙
- 기본 판정 규칙:
  - `source_row_count == target_row_count` → `skip_candidate` (전송 불필요 후보)
  - `source_row_count != target_row_count` → `transfer_required`
- 추가 판정 상태:
  - 대상 테이블 미존재 또는 조회 불가 → `transfer_required` (사유 포함)
  - 카운트 조회 실패 → 정책 기반 처리 (`fail_closed` 또는 `force_transfer`)

### FR-3. 기준(Policy) 옵션
- 최소 아래 정책을 제공한다.
  1. `strict`(기본): 비교 실패가 1건이라도 있으면 해당 테이블 전송 차단 또는 작업 중단.
  2. `best_effort`: 비교 실패 테이블은 경고 후 전송 수행.
  3. `skip_equal_rows`: 행 수 동일 테이블은 자동 제외하고 불일치/미확인 테이블만 전송.

### FR-4. 필터링 옵션
- 비교 결과 기반 필터를 제공한다.
  - `all`
  - `transfer_required` (불일치/대상 없음/정책상 전송 필요)
  - `skip_candidate` (행 수 동일)
  - `count_check_failed`
- 필터는 UI/CLI 양쪽에서 동일한 의미로 동작해야 한다.

### FR-5. 실행 전 요약(Pre-check Summary)
- 마이그레이션 시작 전에 요약 정보를 제공한다.
  - 전체 대상 테이블 수
  - 전송 필요 테이블 수
  - 전송 생략 후보 수
  - 비교 실패 수
- 사용자가 요약을 확인하고 실행 여부를 결정할 수 있어야 한다.

### FR-6. 감사 및 추적 로그
- 테이블별 비교 결과와 최종 판정 결과를 구조화 로그로 남긴다.
- 최소 로그 필드:
  - `table_name`
  - `source_row_count`
  - `target_row_count`
  - `decision` (`transfer_required`/`skip_candidate`/`count_check_failed`)
  - `policy`
  - `reason`

## 7. 데이터 모델/계약 요구사항
- 사전 비교 결과를 표현하는 공통 모델을 정의한다.
- 최소 필드:
  - `table_name`
  - `source_row_count` (nullable)
  - `target_row_count` (nullable)
  - `count_delta` (nullable, `source - target`)
  - `decision`
  - `reason`
  - `checked_at`
- API/내부 이벤트에서 decision enum은 고정값 계약으로 관리한다.

## 8. UX 요구사항
- 테이블 목록에 `원본 건수`, `대상 건수`, `차이`, `판정` 컬럼을 표시한다.
- 기본 필터는 `transfer_required`로 제공해 즉시 실행 대상 중심으로 보이게 한다.
- `skip_candidate`는 회색 톤 등 낮은 강조로 표시해 오퍼레이터 집중도를 높인다.
- `count_check_failed` 항목은 경고 아이콘과 사유 툴팁을 제공한다.

## 9. 성공 지표 (KPI)
- 행 수 동일 테이블의 불필요 전송 건수 80% 이상 감소.
- 전체 Direct Insert 실행 시간 30% 이상 단축(대상/데이터 규모별 평균).
- 사전 점검 후 수동 제외 작업(수기 판단) 50% 이상 감소.

## 10. 수용 기준 (Acceptance Criteria)
1. Direct Insert 사전 점검 시 모든 선택 테이블에 대해 원본/대상 행 수 비교 결과가 생성된다.
2. 기본 정책에서 행 수 동일 테이블은 `skip_candidate`로 분류되고 전송 대상에서 제외 가능하다.
3. 운영자는 `transfer_required`, `skip_candidate`, `count_check_failed` 필터로 목록을 즉시 전환할 수 있다.
4. 비교 실패 항목은 이유가 표시되고, 정책에 따라 차단 또는 진행된다.
5. 실행 전 요약에 전송 필요/생략/실패 건수가 정확히 집계되어 표시된다.

## 11. 릴리즈 범위 제안
- **1차(MVP)**: 사전 행 수 비교, 기본 판정(decision), `transfer_required` 중심 필터, 요약 표시.
- **2차**: 정책 확장(`strict`, `best_effort`, `skip_equal_rows`), 실패 처리 UX 강화.
- **3차**: 대규모 테이블 성능 최적화(병렬 카운트, 타임아웃/재시도 전략), 리포트 고도화.

## 12. 리스크 및 대응
- `COUNT(*)` 비용 증가로 사전 점검 시간이 길어질 수 있음
  - 대응: 병렬 실행 제한, 타임아웃, 샘플링/통계값 활용 옵션(추후)
- 대상 DB 권한 부족으로 카운트 조회 실패 가능
  - 대응: 연결 사전 검증, 실패 정책 선택지 제공
- 행 수 동일하지만 데이터 내용이 다른 경우 오판 가능
  - 대응: 본 기능은 "전송 필요성 1차 판단"임을 명시하고 추후 checksum 옵션 검토


### spec.md

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


### tasks.md

# 작업 목록 (Tasks) - v19

## 목표: Direct Insert 사전 행 수 비교 기반 전송 필요성 판단 체계 도입

### 1. 문서화
- [x] `docs/v19/prd.md` 작성
- [x] `docs/v19/spec.md` 작성
- [x] `docs/v19/tasks.md` 작성

### 2. 백엔드: pre-check 엔진
- [x] 테이블별 source/target COUNT 수집 모듈 추가 (`internal/migration/precheck_engine.go`)
- [x] decision 판정 로직 구현 (`transfer_required`, `skip_candidate`, `count_check_failed`)
- [x] policy 적용기 구현 (`strict`, `best_effort`, `skip_equal_rows`)
- [x] 병렬 처리/타임아웃/재시도(필요시) 설정 반영 (worker pool, context timeout)

### 3. 백엔드: API/연계
- [x] `POST /api/migrations/precheck` 구현
- [x] pre-check 결과 필터 조회 API 구현 (`GET /api/migrations/precheck/results`)
- [x] 기존 마이그레이션 실행 API와 `use_precheck` 연계 (`usePrecheckResults` 파라미터)
- [x] 입력 검증(정책/필터 enum) 및 에러 코드 표준화

### 4. 프론트엔드: pre-check UX
- [x] pre-check 실행 버튼 및 요약 카드 추가
- [x] 결과 테이블 컬럼(원본/대상/차이/판정) 추가
- [x] decision 필터 탭 추가 (`all`, `transfer_required`, `skip_candidate`, `count_check_failed`)
- [x] 실패 항목 reason 툴팁/경고 표시

### 5. CLI
- [x] `--precheck-row-count` 플래그 추가
- [x] `--precheck-policy` 플래그 추가
- [x] `--precheck-filter` 플래그 추가
- [x] dry-run 출력 포맷 확장

### 6. 관측성
- [x] 구조화 로그 필드 추가 (`table_name`, `decision`, `policy`, `reason` 등)
- [x] pre-check 관련 메트릭 추가 (`precheckRunTotal`, `precheckTablesTotal` 등)
- [x] 모니터링 메트릭 테스트 추가 (`TestMonitoringPrecheckMetrics`)

### 7. 테스트
- [x] 단위 테스트: decision/policy 판정
- [x] 통합 테스트: pre-check API 및 필터링 (`precheck_handler_test.go`)
- [x] 실행 연계 테스트: pre-check 기반 실제 전송 대상 축소 (`precheck_engine_test.go`)
- [x] 성능 테스트: 대량 테이블 pre-check 처리 시간 (`precheck_bench_test.go`)

### 8. 운영 가이드/릴리즈
- [x] README에 신규 플래그 및 pre-check 모드 설명 반영
- [x] feature flag(`DBM_V19_PRECHECK`)로 점진 배포
- [x] 운영자 가이드(정책 선택 기준, 실패 대응) 추가 (`docs/v19/operator-guide.md`)


### operator-guide.md

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


---
## <a name="v20"></a> v20

### prd.md

# PRD: 세션 보안 강화, 입력 검증, 마이그레이션 안정성 개선

## 1. 배경

현재 go-db-migration은 v19까지 기능 확장을 거듭하며 프로덕션 환경에서 사용 가능한 수준으로 성숙했다. 그러나 코드베이스 전반에 걸쳐 다음 세 가지 구조적 문제가 잠재되어 있다.

1. **세션 메모리 누수**: 웹 서버의 인증 세션 맵(authSessionManager)이 만료된 세션을 자동으로 정리하지 않아, 장기 운영 시 메모리가 무제한 증가하는 리스크가 있다.
2. **입력 검증 미흡**: 일부 코드 경로에서 테이블명을 SQL `COUNT(*)` 쿼리에 직접 연결해 SQL 인젝션 위험이 존재하며, 배치 크기·워커 수 등 수치형 입력도 음수/0 값 처리가 불명확하다.
3. **마이그레이션 재시도 불완전**: 일시적 네트워크/DB 오류 발생 시 전체 테이블 재시도만 가능하고, 부분 실패(중간 배치 오류) 복구나 지수 백오프(exponential backoff) 전략이 구현되어 있지 않다.

이 문제들은 기능이 아닌 **운영 신뢰성과 보안**에 직결되며, 규모가 커질수록 운영 리스크가 증가한다.

## 2. 문제 정의

### P1. 세션 메모리 누수 및 세션 한도 부재
- 만료된 세션(30분 유휴, 24시간 절대)이 서버 재시작 전까지 메모리에 잔존한다.
- 최대 동시 세션 수 제한이 없어 DoS 또는 의도치 않은 메모리 폭증에 취약하다.
- 세션 만료 확인과 사용 사이에 경쟁 조건(race condition)이 존재한다.

### P2. SQL 인젝션 방어 미흡
- `db/db.go` 내 사전 점검 COUNT 쿼리 등 일부 경로에서 `"SELECT COUNT(*) FROM " + tableName` 패턴이 사용된다.
- 테이블명 quoting이 일관되지 않아 비표준 문자나 예약어 포함 시 실패하거나 악용될 수 있다.
- 배치 크기(batch size), 워커 수(workers), 청크 크기(chunk size) 등 수치형 CLI 인자에 대한 범위 검증이 미흡하다.

### P3. 재시도·복구 전략 불완전
- 트랜잭션 일시 오류 시 단순 로그 기록만 하고 자동 재시도가 이루어지지 않는다.
- 실패한 배치를 건너뛰고 나머지를 계속 처리하는 "부분 허용" 정책이 없다.
- 연결 복구(connection recovery) 시 고정 대기 없이 즉시 재시도해 DB 부하를 가중시킬 수 있다.

## 3. 목표

- 장기 운영 시 서버 메모리 사용량이 예측 가능하고 안정적으로 유지된다.
- 모든 동적 SQL 식별자가 안전하게 이스케이프/인용되어 인젝션 위험이 제거된다.
- 일시적 오류 발생 시 지수 백오프로 자동 재시도하고, 정책에 따라 부분 실패를 허용하거나 차단한다.

## 4. 비목표 (Non-goals)

- 세션 저장소를 Redis 등 외부 스토어로 전환.
- OAuth, OIDC 등 외부 인증 체계 도입.
- 마이그레이션 엔진의 DDL/DML 생성 로직 변경.
- 신규 DB 타입 추가.
- 레코드 단위 checksum 검증(v19 Non-goal 연장).

## 5. 사용자 스토리

- 운영자로서, 서버를 수개월 동안 재시작 없이 운영해도 메모리 사용량이 안정적으로 유지되길 바란다.
- 보안 담당자로서, 테이블명이 포함된 모든 SQL이 안전하게 처리되어 인젝션 위험이 없음을 확인하고 싶다.
- 운영자로서, 네트워크 일시 장애로 인한 마이그레이션 중단 시 자동으로 재시도되어 수동 개입 없이 완료되길 바란다.
- 운영자로서, 일부 배치 실패 시 "실패 건너뛰기" 정책을 선택해 나머지 데이터를 계속 마이그레이션하고 싶다.

## 6. 기능 요구사항

### FR-1. 세션 자동 정리 (Session Cleanup)

- 서버 시작 시 백그라운드 고루틴을 통해 주기적으로 만료된 세션을 정리한다.
  - 기본 정리 주기: 5분 (환경 변수 `DBM_SESSION_CLEANUP_INTERVAL`로 조정 가능)
- 최대 동시 세션 수를 제한한다.
  - 기본값: 100 (환경 변수 `DBM_MAX_SESSIONS`로 조정 가능)
  - 한도 초과 시 가장 오래된 세션을 만료 처리 후 신규 세션 발급
- 세션 확인과 갱신을 단일 잠금(lock) 범위 내에서 원자적으로 처리해 경쟁 조건을 제거한다.
- 서버 종료(graceful shutdown) 시 정리 고루틴이 안전하게 종료된다.

### FR-2. SQL 식별자 안전 처리 (Identifier Quoting)

- 모든 동적 SQL에서 테이블명·컬럼명은 DB 방언(dialect)별 `QuoteIdentifier` 함수를 통해 이스케이프한다.
  - Oracle: `"name"` (대문자 변환 포함)
  - PostgreSQL: `"name"`
  - MySQL/MariaDB: `` `name` ``
  - MSSQL: `[name]`
  - SQLite: `"name"`
- 사전 점검 COUNT 쿼리, 마이그레이션 DML, DDL 생성 등 모든 SQL 생성 경로를 일관되게 적용한다.
- Dialect 인터페이스에 `QuoteIdentifier(name string) string` 메서드를 공식 추가한다.

### FR-3. 수치형 입력 검증 (Numeric Input Validation)

- CLI 파싱 단계(`config` 패키지)에서 아래 인자의 유효성을 검증한다.

| 인자 | 최솟값 | 최댓값 | 기본값 |
|---|---|---|---|
| `batch-size` | 1 | 100,000 | 500 |
| `workers` | 1 | 64 | 4 |
| `chunk-size` | 1 | 10,000,000 | — |
| `db-max-open` | 1 | 1,000 | 10 |
| `db-max-idle` | 0 | 1,000 | 5 |

- 범위 초과 시 명확한 오류 메시지를 출력하고 프로세스를 종료한다.
- Web UI를 통한 입력도 동일 규칙으로 서버 측 검증을 수행한다.

### FR-4. 지수 백오프 자동 재시도 (Exponential Backoff Retry)

- 마이그레이션 실행 중 다음 오류 범주에서 자동 재시도를 수행한다.
  - `CONNECTION_LOST`, `TIMEOUT` (MigrationError.Category 기준)
- 재시도 전략:
  - 초기 대기: 1초
  - 최대 재시도 횟수: 3회 (환경 변수 `DBM_MAX_RETRIES`로 조정 가능)
  - 백오프 계수: 2 (1s → 2s → 4s)
  - 최대 대기: 30초
- 재시도 횟수 소진 후에도 실패 시 기존 오류 처리 흐름(실패 기록, 이벤트 발행)을 따른다.
- 재시도 시도 횟수와 최종 결과를 구조화 로그로 기록한다.

### FR-5. 부분 실패 허용 정책 (Skip-on-Error Policy)

- 마이그레이션 실행 시 배치 단위 실패 처리 정책을 선택할 수 있다.

| 정책 | 설명 |
|---|---|
| `fail_fast` (기본) | 첫 번째 배치 오류 시 해당 테이블 중단 |
| `skip_batch` | 오류 배치를 건너뛰고 다음 배치 계속 진행, 최종 상태는 `partial_success` |

- CLI 옵션: `-on-error [fail_fast|skip_batch]`
- 건너뛴 배치 수와 예상 누락 행 수를 요약 리포트에 포함한다.
- `skip_batch` 모드에서도 재시도 후 실패한 경우에만 건너뛴다(FR-4와 결합).

## 7. 데이터 모델/계약 요구사항

### 세션 메타데이터 확장
```go
type AuthSession struct {
    Token     string
    UserID    string
    CreatedAt time.Time
    LastUsedAt time.Time
    ExpiresAt time.Time   // 절대 만료 시각 (신규)
}
```

### Dialect 인터페이스 확장
```go
type Dialect interface {
    // ... 기존 메서드 ...
    QuoteIdentifier(name string) string  // 신규
}
```

### 재시도 결과 이벤트 확장
```go
type RetryEvent struct {
    TableName   string
    Attempt     int
    MaxAttempts int
    ErrorMsg    string
    WaitSeconds int
}
```

## 8. UX 요구사항

- 세션 만료 시 Web UI에서 `401 Unauthorized` 응답과 함께 로그인 화면으로 자동 리다이렉트한다.
- 자동 재시도 발생 시 Web UI 진행 패널에 "재시도 중 (N/M)" 상태 표시를 노출한다.
- `skip_batch` 모드에서 건너뛴 배치가 있을 경우, 테이블 완료 뱃지를 `partial` (노란색)으로 표시한다.
- 입력 검증 오류는 CLI에서 `[ERROR] --batch-size: 최솟값은 1입니다 (입력값: 0)` 형식으로 즉시 출력한다.

## 9. 성공 지표 (KPI)

- 24시간 지속 운영 시 서버 메모리 증가량 < 50MB (기존 대비).
- 동적 SQL 생성 경로 내 미인용 식별자 사용 0건 (코드 리뷰 기준).
- 일시적 연결 오류로 인한 테이블 마이그레이션 실패율 50% 이상 감소.
- 수치형 인자 경계 위반에 대한 테스트 커버리지 100%.

## 10. 수용 기준 (Acceptance Criteria)

1. 서버 기동 후 5분 주기로 만료 세션이 삭제되고, `DBM_MAX_SESSIONS` 한도 초과 시 오래된 세션이 자동 만료된다.
2. 모든 Dialect 구현체에 `QuoteIdentifier`가 추가되고, 사전 점검 COUNT 쿼리를 포함한 모든 동적 SQL에 적용된다.
3. `--batch-size 0` 또는 `--workers 100` 입력 시 오류 메시지와 함께 즉시 종료된다.
4. CONNECTION_LOST/TIMEOUT 오류 시 최대 3회 지수 백오프 재시도 후 성공 또는 최종 실패 처리된다.
5. `--on-error skip_batch` 옵션 사용 시 오류 배치를 건너뛰고 나머지 배치를 완료하며, 최종 상태가 `partial_success`로 기록된다.

## 11. 릴리즈 범위 제안

- **1차(MVP)**: 세션 자동 정리(FR-1), SQL 식별자 인용 통일(FR-2).
- **2차**: 수치형 입력 검증(FR-3), 지수 백오프 재시도(FR-4).
- **3차**: 부분 실패 허용 정책(FR-5), Web UI 상태 표시 연동.

## 12. 리스크 및 대응

| 리스크 | 대응 |
|---|---|
| 기존 Dialect 구현에 `QuoteIdentifier` 추가 시 컴파일 오류 | 인터페이스 변경 전 모든 구현체에 메서드 추가 후 인터페이스 반영 |
| 세션 정리 고루틴이 잠금 경쟁을 유발할 수 있음 | 정리 주기를 충분히 길게(5분) 설정하고, tryLock 패턴 또는 별도 만료 인덱스 사용 검토 |
| 지수 백오프 중 중복 삽입(duplicate insert) 발생 가능 | Upsert 모드와의 병행 시 멱등성 보장 여부 명시; 비upsert 모드에서는 재시도 전 배치 롤백 필수 |
| `skip_batch`로 인한 데이터 누락 미인지 | 요약 리포트에 누락 배치 수 및 예상 행 수를 명시; 운영자 확인 후 재실행 권장 문구 출력 |


### spec.md

# 기술 사양서 (Technical Specifications) - v20

## 1. 아키텍처 개요

v20은 **세 가지 운영 신뢰성 문제**를 해결한다.

1. **세션 메모리 누수**: `authSessionManager`가 만료 세션을 요청 경로에서만 정리하여 장기 운영 시 메모리가 축적된다.
2. **미인용 SQL 식별자**: `internal/db/db.go`의 `COUNT(*)` 쿼리가 테이블명을 문자열 연결로 구성해 인젝션 위험이 존재한다.
3. **재시도 전략 부재**: `CONNECTION_LOST`·`TIMEOUT` 오류 발생 시 자동 재시도 없이 즉시 실패 처리된다.

핵심 설계 원칙:
- 기존 마이그레이션 엔진(DDL/DML 생성) 변경 없음
- 하위 호환성 유지 — 기본값이 기존 동작과 동일
- 환경 변수로 세션 정책 튜닝 가능

---

## 2. 도메인 모델

### 2.1 세션 구조체 확장

**파일**: `internal/web/server.go`

기존:
```go
type authSession struct {
    UserID     int64
    Username   string
    CreatedAt  time.Time
    LastSeenAt time.Time
}
```

변경 후:
```go
type authSession struct {
    UserID     int64
    Username   string
    CreatedAt  time.Time
    LastSeenAt time.Time
    ExpiresAt  time.Time // 절대 만료 시각 (CreatedAt + absoluteTTL)
}
```

### 2.2 authSessionManager 확장

기존 필드에 추가:
```go
type authSessionManager struct {
    mu          sync.RWMutex
    sessions    map[string]authSession
    idleTTL     time.Duration
    absoluteTTL time.Duration
    maxSessions int       // 신규: 최대 동시 세션 수 (0이면 무제한)
    stopCleanup chan struct{} // 신규: 정리 고루틴 종료 신호
    metrics     *monitoringMetrics
}
```

### 2.3 재시도 이벤트 모델

**파일**: `internal/migration/errors.go` (신규 타입 추가)

```go
// RetryEvent는 재시도 발생 시 이벤트 버스로 전송되는 구조체이다.
type RetryEvent struct {
    TableName   string `json:"table_name"`
    Attempt     int    `json:"attempt"`
    MaxAttempts int    `json:"max_attempts"`
    ErrorMsg    string `json:"error_msg"`
    WaitSeconds int    `json:"wait_seconds"`
}
```

### 2.4 마이그레이션 결과 상태 확장

`partial_success` 상태를 기존 상태 목록에 추가한다.

**파일**: `internal/migration/state.go`

```go
const (
    StatusSuccess        = "success"
    StatusFailed         = "failed"
    StatusPartialSuccess = "partial_success" // 신규: skip_batch 정책으로 일부 배치 누락
)
```

---

## 3. 백엔드 설계

### 3.1 FR-1: 세션 자동 정리

**파일**: `internal/web/server.go`

#### 3.1.1 생성자 변경

`newAuthSessionManager` 파라미터에 `maxSessions int` 추가.
`stopCleanup` 채널 초기화 후 `startCleanupLoop` 고루틴 시작.

```go
func newAuthSessionManager(
    idleTTL, absoluteTTL time.Duration,
    maxSessions int,
    metrics ...*monitoringMetrics,
) *authSessionManager {
    m := &authSessionManager{
        sessions:    make(map[string]authSession),
        idleTTL:     idleTTL,
        absoluteTTL: absoluteTTL,
        maxSessions: maxSessions,
        stopCleanup: make(chan struct{}),
        metrics:     ...,
    }
    go m.startCleanupLoop(5 * time.Minute)
    return m
}
```

#### 3.1.2 정리 고루틴

```go
func (m *authSessionManager) startCleanupLoop(interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            m.purgeExpired()
        case <-m.stopCleanup:
            return
        }
    }
}

func (m *authSessionManager) purgeExpired() {
    now := time.Now()
    m.mu.Lock()
    defer m.mu.Unlock()
    for token, s := range m.sessions {
        if now.After(s.ExpiresAt) || now.Sub(s.LastSeenAt) > m.idleTTL {
            delete(m.sessions, token)
            m.metrics.recordSessionExpired()
        }
    }
}
```

#### 3.1.3 최대 세션 수 제한

`createSession` 호출 시 세션 수 초과 여부 확인.
초과 시 `ExpiresAt` 기준 가장 오래된 세션 1건 삭제 후 신규 세션 생성.

```go
func (m *authSessionManager) createSession(...) (string, authSession, error) {
    // ... 토큰 생성 ...
    m.mu.Lock()
    defer m.mu.Unlock()
    if m.maxSessions > 0 && len(m.sessions) >= m.maxSessions {
        m.evictOldest() // ExpiresAt 기준 최솟값 세션 삭제
    }
    // ExpiresAt = now + absoluteTTL
    s := authSession{..., ExpiresAt: now.Add(m.absoluteTTL)}
    m.sessions[token] = s
    return token, s, nil
}
```

#### 3.1.4 Graceful Shutdown 연동

`RunServerWithAuth` 내 서버 종료 시 `close(authSessions.stopCleanup)` 호출.

#### 3.1.5 환경 변수

| 변수 | 기본값 | 설명 |
|---|---|---|
| `DBM_MAX_SESSIONS` | `100` | 최대 동시 세션 수 (0 = 무제한) |
| `DBM_SESSION_CLEANUP_INTERVAL` | `5m` | 정리 주기 (Go `time.Duration` 형식) |

---

### 3.2 FR-2: SQL 식별자 인용

**파일**: `internal/db/db.go`

#### 3.2.1 문제 위치

```go
// 변경 전 (line 162)
err := d.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+tableName).Scan(&count)

// 변경 전 (line 174)
err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+tableName).Scan(&count)
```

#### 3.2.2 함수 시그니처 변경

`SQLDBCountFn`과 `PGPoolCountFn`에 `quoteIdentifier func(string) string` 파라미터 추가.

```go
func SQLDBCountFn(
    d *sql.DB,
    quoteIdentifier func(string) string,
) func(ctx context.Context, tableName string) (int, error) {
    return func(ctx context.Context, tableName string) (int, error) {
        quoted := quoteIdentifier(tableName)
        err := d.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+quoted).Scan(&count)
        ...
    }
}

func PGPoolCountFn(
    pool PGPool,
    quoteIdentifier func(string) string,
) func(ctx context.Context, tableName string) (int, error) {
    return func(ctx context.Context, tableName string) (int, error) {
        quoted := quoteIdentifier(tableName)
        err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM "+quoted).Scan(&count)
        ...
    }
}
```

#### 3.2.3 호출부 변경

`SQLDBCountFn`·`PGPoolCountFn` 호출 위치에서 해당 Dialect의 `QuoteIdentifier` 메서드를 전달한다.
`QuoteIdentifier`는 이미 `internal/dialect/dialect.go`의 인터페이스에 선언되어 있으며 모든 구현체에 존재한다.

---

### 3.3 FR-3: 수치형 입력 검증

**파일**: `internal/config/config.go`

`ParseFlags` 완료 후 `validateConfig(cfg *Config) error` 함수 호출.

```go
type configBound struct {
    min, max int
    name     string
}

var numericBounds = []struct {
    field *int
    configBound
}{
    {&cfg.BatchSize,  configBound{1, 100_000, "--batch"}},
    {&cfg.Workers,    configBound{1, 64,      "--workers"}},
    {&cfg.DBMaxOpen,  configBound{1, 1_000,   "--db-max-open"}},
    {&cfg.DBMaxIdle,  configBound{0, 1_000,   "--db-max-idle"}},
}

func validateConfig(cfg *Config) error {
    for _, b := range numericBounds {
        if *b.field < b.min || *b.field > b.max {
            return fmt.Errorf("%s: 값 범위는 %d~%d입니다 (입력값: %d)",
                b.name, b.min, b.max, *b.field)
        }
    }
    return nil
}
```

오류 발생 시 `fmt.Fprintf(os.Stderr, "[ERROR] %v\n", err)` 후 `os.Exit(1)`.

---

### 3.4 FR-4: 지수 백오프 재시도

**파일**: `internal/migration/retry.go` (신규)

```go
// RetryConfig는 재시도 정책을 정의한다.
type RetryConfig struct {
    MaxAttempts int           // 기본 3
    InitialWait time.Duration // 기본 1s
    Multiplier  float64       // 기본 2.0
    MaxWait     time.Duration // 기본 30s
}

// DefaultRetryConfig는 기본 재시도 설정을 반환한다.
func DefaultRetryConfig() RetryConfig {
    return RetryConfig{
        MaxAttempts: 3,
        InitialWait: time.Second,
        Multiplier:  2.0,
        MaxWait:     30 * time.Second,
    }
}

// WithRetry는 fn을 RecoverableError에 한해 지수 백오프로 재시도한다.
// eventFn이 nil이 아니면 재시도 시마다 RetryEvent를 전달한다.
func WithRetry(
    ctx context.Context,
    cfg RetryConfig,
    tableName string,
    eventFn func(RetryEvent),
    fn func() error,
) error {
    wait := cfg.InitialWait
    for attempt := 1; ; attempt++ {
        err := fn()
        if err == nil {
            return nil
        }
        var migErr *MigrationError
        if !errors.As(err, &migErr) || !migErr.Recoverable {
            return err
        }
        if attempt >= cfg.MaxAttempts {
            return err
        }
        if eventFn != nil {
            eventFn(RetryEvent{
                TableName:   tableName,
                Attempt:     attempt,
                MaxAttempts: cfg.MaxAttempts,
                ErrorMsg:    err.Error(),
                WaitSeconds: int(wait.Seconds()),
            })
        }
        slog.Warn("migration retry",
            "table", tableName, "attempt", attempt,
            "wait_s", wait.Seconds(), "error", err)
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(wait):
        }
        wait = min(time.Duration(float64(wait)*cfg.Multiplier), cfg.MaxWait)
    }
}
```

**환경 변수**:

| 변수 | 기본값 | 설명 |
|---|---|---|
| `DBM_MAX_RETRIES` | `3` | 최대 재시도 횟수 |
| `DBM_RETRY_INITIAL_WAIT` | `1s` | 초기 대기 시간 |

---

### 3.5 FR-5: 부분 실패 허용 정책 (skip_batch)

**파일**: `internal/config/config.go` — `OnError string` 필드 추가
**파일**: `internal/migration/migration.go` — 배치 오류 처리 경로 분기

```go
// config.go
flag.StringVar(&cfg.OnError, "on-error", "fail_fast",
    "배치 오류 처리 정책: fail_fast | skip_batch")
```

배치 실행 루프 내 오류 처리:

```go
// migration.go (개략적 구조)
if err != nil {
    if cfg.OnError == "skip_batch" {
        skippedBatches++
        slog.Warn("batch skipped", "table", table, "batch", batchNum, "error", err)
        continue
    }
    return err
}
```

완료 후 `skippedBatches > 0`이면 상태를 `partial_success`로 기록.
최종 리포트에 누락 배치 수(`skipped_batches`)와 예상 누락 행 수(`estimated_skipped_rows`) 포함.

---

## 4. API 계약

### 4.1 재시도 이벤트 WebSocket 메시지

기존 WebSocket 이벤트 타입(`bus` 패키지) 에 `retry` 타입 추가.

```json
{
  "type": "retry",
  "payload": {
    "table_name": "EMP",
    "attempt": 1,
    "max_attempts": 3,
    "error_msg": "connection lost: EOF",
    "wait_seconds": 1
  }
}
```

### 4.2 마이그레이션 실행 API 파라미터 추가

`POST /api/migrate` 요청 바디에 `on_error` 필드 추가.

```json
{
  "tables": ["EMP", "DEPT"],
  "on_error": "skip_batch"
}
```

응답의 테이블별 결과에 `skipped_batches` 필드 추가:

```json
{
  "table_name": "EMP",
  "status": "partial_success",
  "skipped_batches": 2,
  "estimated_skipped_rows": 400
}
```

---

## 5. CLI 설계

### 5.1 신규/변경 플래그

| 플래그 | 타입 | 기본값 | 설명 |
|---|---|---|---|
| `--on-error` | string | `fail_fast` | 배치 오류 처리 정책 (`fail_fast` \| `skip_batch`) |

### 5.2 검증 오류 출력 형식

```
[ERROR] --batch: 값 범위는 1~100000입니다 (입력값: 0)
```

### 5.3 재시도 로그 형식

```
[WARN] 재시도 중 (1/3) table=EMP wait=1s error="connection lost: EOF"
```

### 5.4 부분 성공 요약

```
[WARN] partial_success: EMP — 2개 배치 건너뜀 (예상 누락 행: ~400건)
```

---

## 6. UI 설계

### 6.1 재시도 상태 표시

진행 패널(`MigrationProgress` 컴포넌트)에 재시도 상태 행 추가.
- 표시 조건: WebSocket `retry` 이벤트 수신 시
- 형식: `재시도 중 (1/3) — 1초 후 재시작`
- 재시도 성공 시 해당 행 자동 제거

### 6.2 partial_success 뱃지

기존 성공(`success`) 뱃지와 구분:
- `partial_success`: 노란색 뱃지 (`⚠ 부분 완료`)
- 호버 시 툴팁: `N개 배치 건너뜀 (예상 ~M건 누락)`

### 6.3 on-error 정책 선택 UI

마이그레이션 설정 패널에 오류 처리 정책 라디오 버튼 추가:
- `오류 시 중단` (fail_fast, 기본)
- `오류 배치 건너뛰기` (skip_batch)

---

## 7. 로깅/관측성

### 7.1 구조화 로그 필드

| 이벤트 | 필드 |
|---|---|
| 세션 정리 | `cleaned_count`, `remaining_count` |
| 세션 한도 초과 | `evicted_token_prefix`, `current_count`, `max_sessions` |
| 재시도 | `table`, `attempt`, `max_attempts`, `wait_s`, `error` |
| 배치 건너뜀 | `table`, `batch_num`, `estimated_rows`, `error` |

### 7.2 메트릭

기존 `monitoringMetrics`에 카운터 추가:

| 메트릭 | 설명 |
|---|---|
| `session_cleanup_total` | 정리 실행 횟수 |
| `session_evicted_total` | 한도 초과로 삭제된 세션 수 |
| `migration_retry_total{table}` | 테이블별 재시도 횟수 |
| `migration_batch_skipped_total{table}` | 건너뛴 배치 수 |
| `migration_partial_success_total` | partial_success 완료 테이블 수 |

---

## 8. 오류 처리

| 조건 | 처리 |
|---|---|
| `DBM_MAX_SESSIONS` 비정수 값 | 서버 시작 실패 + 오류 로그 |
| `--on-error` 허용되지 않은 값 | `[ERROR]` 출력 후 `os.Exit(1)` |
| 재시도 중 `ctx.Done()` | 즉시 중단, `context canceled` 오류 반환 |
| `skip_batch` + upsert 모드 | 허용; 건너뛴 배치는 upsert 재시도로 처리됨을 로그에 명시 |
| `partial_success` 상태에서 resume | resume 시 건너뛴 배치도 재실행 대상에 포함 |

---

## 9. 테스트 전략

### 9.1 세션 관리 단위 테스트 (`internal/web/server_test.go`)

- `purgeExpired`: 만료 세션 N개 삽입 → 정리 후 잔여 세션 수 검증
- `evictOldest`: maxSessions=2, 3번째 세션 생성 시 가장 오래된 세션 삭제 검증
- `createSession`: ExpiresAt이 `CreatedAt + absoluteTTL`과 일치하는지 검증
- 정리 고루틴: `startCleanupLoop` 호출 후 `stopCleanup` 전송 시 고루틴 정상 종료 검증

### 9.2 입력 검증 단위 테스트 (`internal/config/config_test.go`)

- `batch=0` → 오류 반환
- `workers=65` → 오류 반환
- `db-max-idle=-1` → 오류 반환
- `batch=1`, `batch=100000` → 정상 통과 (경계값)

### 9.3 재시도 단위 테스트 (`internal/migration/retry_test.go` 신규)

- 1회 실패 후 성공 → `attempt=2`에서 완료, 오류 nil 반환
- `MaxAttempts` 회 모두 실패 → 마지막 오류 반환
- `Recoverable=false` 오류 → 재시도 없이 즉시 반환
- `ctx.Cancel()` → `context canceled` 반환

### 9.4 skip_batch 통합 테스트 (`internal/migration/direct_test.go` 확장)

- 3배치 중 2번째에서 오류 발생 + `OnError="skip_batch"` → 1, 3 배치 완료, 상태 `partial_success`
- 3배치 중 2번째에서 오류 발생 + `OnError="fail_fast"` → 오류 반환, 상태 `failed`

### 9.5 SQL 인젝션 방어 테스트 (`internal/db/db_test.go`)

- `SQLDBCountFn`에 `tableName="users; DROP TABLE users--"` 전달 시 quoted 쿼리 문자열 검증
- 각 Dialect `QuoteIdentifier` 경계값 테스트 (특수문자, 예약어, 대소문자 혼합)

---

## 10. 롤아웃 계획

1. **1차 배포 (FR-1, FR-2)**
   - 세션 자동 정리 기본 활성화 (`DBM_MAX_SESSIONS=100`)
   - `SQLDBCountFn`/`PGPoolCountFn` 호출부 QuoteIdentifier 적용
   - 기존 테스트 전량 통과 확인

2. **2차 배포 (FR-3, FR-4)**
   - 입력 검증 활성화 — 기존 기본값이 모두 유효 범위 내이므로 기존 실행에 영향 없음
   - 재시도 기본 활성화 (`DBM_MAX_RETRIES=3`), 기존 fail-fast 동작은 재시도 소진 후 동일

3. **3차 배포 (FR-5)**
   - `--on-error skip_batch` CLI/UI 공개
   - `partial_success` 상태 Web UI 뱃지 활성화

---

## 11. 오픈 이슈

- `evictOldest` 구현 시 O(n) 순회 대신 세션 삽입 순서를 보조 슬라이스로 관리하는 방안 검토 (세션 수 > 500 환경 대비)
- `skip_batch` + PostgreSQL COPY 모드 조합: COPY는 배치 단위 롤백이 지원되지 않으므로 COPY 모드에서 `skip_batch` 허용 여부 정책 확정 필요
- 재시도 `WaitSeconds`를 WebSocket으로 전송 시 클라이언트 카운트다운 타이머 구현 여부


### tasks.md

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


---
## <a name="v22"></a> v22

### prd.md

# PRD: 타겟 DB 테이블 목록 조회 및 소스-타겟 비교 UI

## 1. 배경

현재 go-db-migration은 소스(Oracle)의 테이블 목록을 조회해 마이그레이션 대상을 선택하는 흐름을 제공하고 있다. 그러나 **타겟 DB에 어떤 테이블이 존재하는지를 UI에서 직접 확인하는 방법이 없다.** 운영자는 다음과 같은 판단을 별도 DB 클라이언트 없이 내려야 하는 상황이다.

- "타겟 DB에 이미 같은 이름의 테이블이 있는가?"
- "이전 마이그레이션 이후 소스에서 삭제된 테이블이 타겟에 잔존하는가?"
- "소스 테이블 중 아직 한 번도 마이그레이션되지 않은 테이블은 어느 것인가?"

이 정보를 UI 안에서 비교 뷰로 제공하면 마이그레이션 전략 수립, 불필요한 테이블 정리, 누락 탐지 등 운영 전반의 의사결정 속도와 정확성이 높아진다.

**v22의 타겟 DB는 PostgreSQL로 단일화한다.** 타겟 DB 선택 옵션(MariaDB, MySQL, MSSQL, SQLite)은 UI 및 모든 기능에서 제거한다.

## 2. 문제 정의

### P1. 타겟 현황 파악을 위해 외부 도구에 의존
- 타겟 DB(PostgreSQL)에 어떤 테이블이 존재하는지 확인하려면 별도의 DB 클라이언트(DBeaver, psql 등)를 실행해야 한다.
- UI 내에서 소스와 타겟을 동시에 비교할 수 없어 작업 맥락이 단절된다.

### P2. 소스에만 있는 테이블과 타겟에만 있는 테이블의 구분 불가
- 마이그레이션 후 소스에서 테이블이 삭제되더라도 타겟에 잔존하는 테이블을 식별할 방법이 없다.
- 소스에 새로 생긴 테이블이 타겟에 아직 없는 경우, 운영자가 수동으로 확인하지 않으면 누락된다.

### P3. 마이그레이션 대상 테이블 선택의 비효율
- 현재는 소스 테이블 전체 목록에서 수동으로 선택해야 하며, "타겟에 없는 것만 선택"하는 빠른 선택 수단이 없다.
- 이미 타겟에 동일한 테이블이 존재하는 경우에도 구별 없이 목록에 나열된다.

## 3. 목표

- 타겟(PostgreSQL) 연결 후 테이블 목록을 조회하여 소스(Oracle) 목록과 나란히 비교할 수 있는 UI를 제공한다.
- 소스만 존재 / 양쪽 존재 / 타겟만 존재 세 가지 카테고리로 테이블을 분류하여 상태를 직관적으로 파악할 수 있게 한다.
- "소스에만 있는 테이블 전체 선택" 등 비교 결과 기반의 빠른 선택 기능으로 마이그레이션 대상 지정 속도를 높인다.
- 타겟 DB 옵션을 PostgreSQL로 단일화하여 UI와 백엔드의 복잡도를 줄인다.

## 4. 비목표 (Non-goals)

- 타겟 테이블의 DDL(컬럼 구조) 상세 비교 — 컬럼 수준 스키마 diff는 별도 기능으로 고려.
- 타겟 DB에서 테이블 삭제/생성 등 DDL 실행.
- 소스가 Oracle이 아닌 경우의 타겟 비교 (소스는 Oracle 고정).
- 뷰(View), 트리거, 프로시저 등 테이블 이외 객체 비교.
- 타겟 테이블 목록의 실시간 폴링(자동 갱신).
- MariaDB, MySQL, MSSQL, SQLite 타겟 지원 — v22에서 타겟 DB는 PostgreSQL로 단일화하며, 이 DB 타입들은 UI 선택 옵션 및 모든 기능에서 제거한다.

## 5. 사용자 스토리

- 운영자로서, PostgreSQL 타겟에 이미 존재하는 테이블을 Oracle 소스 목록과 나란히 보고 싶다. 그래야 어떤 테이블을 새로 마이그레이션해야 할지 즉시 판단할 수 있다.
- 운영자로서, 소스에는 없지만 타겟에만 남아 있는 테이블 목록을 확인해서 정리 대상 후보를 파악하고 싶다.
- 운영자로서, "소스에만 있는 테이블 전체 선택" 버튼 한 번으로 아직 마이그레이션되지 않은 테이블들을 빠르게 선택하고 싶다.
- 운영자로서, 소스와 타겟 양쪽에 동일한 테이블이 있는 경우 행 수 차이를 바로 확인하여 재마이그레이션 필요 여부를 판단하고 싶다.
- 운영자로서, 타겟 테이블 목록을 다시 새로고침하는 버튼이 있어서 마이그레이션 진행 중에도 최신 현황을 확인하고 싶다.

## 6. 기능 요구사항

### FR-0. 타겟 DB 옵션 단일화 (PostgreSQL)

- 타겟 DB 선택 드롭다운에서 PostgreSQL 이외의 옵션(MariaDB, MySQL, MSSQL, SQLite)을 제거한다.
- 타겟 DB 타입은 항상 `postgres`로 고정되며, 관련 UI 선택 요소를 제거하거나 읽기 전용으로 표시한다.
- 백엔드의 타겟 연결·마이그레이션·사전 점검 등 모든 API 경로에서 PostgreSQL 이외의 DB 타입 입력을 거부한다(`400 Bad Request`).

### FR-1. 타겟 테이블 목록 조회 API

- 타겟 연결 정보(URL, 스키마)를 받아 PostgreSQL의 테이블 목록을 반환하는 새 API 엔드포인트를 추가한다.
  - `POST /api/target-tables`
  - 요청: `{ targetUrl, schema }`
  - 응답: `{ tables: string[], fetchedAt: string }`
- 조회 쿼리:

```sql
SELECT table_name
FROM information_schema.tables
WHERE table_schema = $1
  AND table_type = 'BASE TABLE'
ORDER BY table_name
```

- 연결 오류 또는 권한 부족 시 명확한 오류 메시지를 반환한다.
- 조회는 읽기 전용(SELECT만 수행)이며 타임아웃은 10초로 제한한다.

### FR-2. 타겟 테이블 목록 조회 UI

- 타겟 섹션("2. Target")의 연결 테스트 버튼 옆에 **"타겟 테이블 조회"** 버튼을 추가한다.
  - 조회 성공 시: 타겟 테이블 수를 배지로 표시하고 비교 패널을 활성화한다.
  - 조회 중: 버튼에 로딩 스피너 표시.
  - 조회 실패 시: 인라인 오류 메시지 표시.
- 이전에 조회한 결과가 있으면 마지막 조회 시각(`fetchedAt`)을 함께 표시한다.
- "새로고침" 버튼으로 언제든지 재조회할 수 있다.

### FR-3. 소스-타겟 비교 패널

- 소스 테이블 목록(Oracle 조회 완료)과 타겟 테이블 목록(FR-2) 모두 준비되면 **비교 패널**이 테이블 선택 섹션 상단에 나타난다.
- 비교 결과는 다음 세 카테고리로 분류한다:

| 카테고리 | 정의 | 배지 색상 |
|---|---|---|
| `source_only` | 소스에만 존재 (타겟에 없음) | 파란색 |
| `both` | 소스와 타겟 양쪽에 존재 | 초록색 |
| `target_only` | 타겟에만 존재 (소스에 없음) | 노란색 |

- 비교 패널 구성:
  - 상단 요약 카드: 각 카테고리별 테이블 수 표시.
  - 카테고리 필터 탭: `전체` / `소스만` / `양쪽` / `타겟만` 탭 선택 시 아래 목록 필터링.
  - 테이블 목록 테이블:

| 컬럼 | 설명 |
|---|---|
| 테이블명 | 테이블 이름 |
| 소스 | ✓ 또는 — |
| 타겟 | ✓ 또는 — |
| 소스 행 수 | pre-check 결과가 있으면 표시, 없으면 — |
| 타겟 행 수 | pre-check 결과가 있으면 표시, 없으면 — |
| 상태 | `source_only` / `both` / `target_only` 배지 |

- 비교 테이블은 기본적으로 카테고리 순(source_only → both → target_only) + 테이블명 알파벳 순으로 정렬한다.
- 검색 입력 필드로 테이블명 필터링을 지원한다.

### FR-4. 비교 기반 빠른 선택

- 테이블 선택 섹션("3. Table Selection")에 비교 결과 기반 빠른 선택 버튼을 추가한다:

| 버튼 | 동작 |
|---|---|
| 소스에만 있는 테이블 선택 | `source_only` 카테고리의 테이블 전체를 체크박스 선택 |
| 양쪽에 있는 테이블 선택 | `both` 카테고리의 테이블 전체를 체크박스 선택 |
| 선택 해제 (비교) | 비교 기반으로 선택된 항목만 해제 |

- 버튼은 비교 결과(타겟 테이블 목록)가 조회된 경우에만 활성화된다.
- 기존 "현재 목록 전체 선택" / "현재 목록 선택 해제" 버튼과 병행하여 동작한다.

### FR-5. pre-check 결과와의 연동

- 기존 Pre-check Row Count 결과가 있으면 비교 패널의 소스/타겟 행 수 컬럼에 자동으로 반영한다.
- `both` 카테고리이면서 소스 행 수 ≠ 타겟 행 수인 테이블은 행 수 차이를 강조 표시(`row_diff` 배지, 주황색)한다.
- pre-check 미실행 상태에서는 행 수 컬럼을 `—`로 표시하고, "Pre-check 실행 후 행 수가 표시됩니다" 안내 문구를 노출한다.

## 7. API 계약 요구사항

### POST /api/target-tables 요청/응답

```go
// 요청
type TargetTablesRequest struct {
    TargetURL string `json:"targetUrl"`
    Schema    string `json:"schema"`
}

// 응답
type TargetTablesResponse struct {
    Tables    []string `json:"tables"`
    FetchedAt string   `json:"fetchedAt"` // RFC3339
}
```

### 프론트엔드 상태 확장

```typescript
type TargetTableEntry = {
  name: string;
  inSource: boolean;
  inTarget: boolean;
  category: "source_only" | "both" | "target_only";
  sourceRowCount: number | null;  // pre-check 결과
  targetRowCount: number | null;  // pre-check 결과
};

type CompareState = {
  targetTables: string[];
  fetchedAt: string | null;
  busy: boolean;
  error: string | null;
};
```

## 8. UX 요구사항

- 타겟 DB 선택 드롭다운을 제거하고 "PostgreSQL" 고정 레이블로 대체한다.
- 타겟 테이블 조회는 타겟 연결 테스트와 독립적으로 실행 가능하다(테스트 없이도 조회 버튼 클릭 가능).
- 소스 테이블이 조회되지 않은 상태에서 타겟 테이블만 조회한 경우, 비교 패널은 표시되지 않고 타겟 테이블 수만 배지로 안내한다.
- 비교 패널은 기본적으로 접힘(collapsed) 상태로 표시하며, 사용자가 펼칠 수 있다.
- 마이그레이션 실행 중에는 "타겟 테이블 조회" 버튼을 비활성화한다.
- 한국어/영어 전환 시 비교 패널의 모든 레이블도 즉시 전환된다.
- 모바일 뷰(작은 화면)에서는 비교 테이블을 가로 스크롤로 처리한다.

## 9. 성공 지표 (KPI)

- 타겟 테이블 조회 API 응답 시간 p95 < 3초.
- 비교 패널 사용 후 "소스만 있는 테이블 선택" 버튼으로 대상 지정하는 사용자 비율 측정.
- 타겟에만 존재하는 테이블 발견으로 인한 운영자 수동 개입(정리 작업) 케이스 파악.

## 10. 수용 기준 (Acceptance Criteria)

1. 타겟 DB 선택 드롭다운이 제거되고 PostgreSQL 고정 레이블이 표시된다.
2. PostgreSQL 이외의 DB 타입을 API로 전달하면 `400 Bad Request`가 반환된다.
3. 타겟 연결 정보 입력 후 "타겟 테이블 조회" 버튼 클릭 시 PostgreSQL의 테이블 목록이 조회되어 배지에 수가 표시된다.
4. 소스와 타겟 목록이 모두 준비되면 비교 패널이 나타나고 `source_only` / `both` / `target_only` 카테고리 별 수가 요약 카드에 표시된다.
5. 카테고리 탭 필터 선택 시 해당 카테고리의 테이블만 목록에 표시된다.
6. "소스에만 있는 테이블 선택" 버튼 클릭 시 `source_only` 테이블이 마이그레이션 체크박스에 선택된다.
7. pre-check 결과가 있으면 비교 패널 행 수 컬럼에 반영되고, `both` + 행 수 불일치 테이블에는 `row_diff` 배지가 표시된다.
8. 타겟 연결 실패 또는 권한 부족 시 UI에 명확한 오류 메시지가 표시된다.
9. 한국어/영어 전환 시 비교 패널의 모든 텍스트가 즉시 전환된다.

## 11. 릴리즈 범위 제안

- **1차(MVP)**: 타겟 DB 단일화(FR-0) + 타겟 테이블 조회 API(FR-1) + 조회 버튼 및 배지 UI(FR-2) + 기본 비교 패널(FR-3).
- **2차**: 비교 기반 빠른 선택(FR-4) + pre-check 결과 연동(FR-5).
- **3차**: 비교 결과 CSV 내보내기, 마이그레이션 완료 후 자동 재조회 및 차이 강조 표시.

## 12. 리스크 및 대응

| 리스크 | 대응 |
|---|---|
| 기존 MariaDB/MySQL/MSSQL/SQLite 타겟을 사용하던 운영 환경에서 업그레이드 시 설정 손실 | 릴리즈 노트에 타겟 DB 단일화 breaking change 명시; 기존 설정 파일의 타겟 DB 타입이 postgres가 아닌 경우 서버 시작 시 경고 로그 출력 |
| 타겟 DB 계정에 `information_schema` 조회 권한이 없는 경우 | 권한 부족 오류를 명확히 구분하여 "테이블 목록 조회 권한이 없습니다" 메시지 출력 |
| 테이블 수가 수천 개인 경우 UI 성능 저하 | 1차 MVP는 500개 초과 시 경고 표시; 이후 가상 스크롤 도입 검토 |
| 소스 테이블명 대소문자와 타겟 테이블명 불일치 | Oracle은 대문자 기본이므로 비교 시 양측 테이블명을 소문자로 정규화하여 매칭 |
| 마이그레이션 진행 중 타겟 테이블 목록이 변경될 수 있음 | 조회 시각(`fetchedAt`) 표시로 최신성 인지 유도; 마이그레이션 실행 중 재조회 비활성화 |


### spec.md

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


### tasks.md

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
- [x] `compareEntries` 메모: source-only / both / target-only 분류 검증
- [x] 대소문자 정규화: 소스 `"USERS"` + 타겟 `"users"` → `both` 확인
- [x] `selectByCategory("source_only")`: 해당 테이블만 추가 확인
- [x] `isRowDiff` 조건 검증

#### 10.3 최종 확인
- [ ] `go test ./...` 전량 통과
- [ ] `npm run build` 빌드 오류 없음

---

### 11. 릴리즈 노트
- [x] README에 breaking change 명시 (타겟 DB 단일화, MariaDB/MySQL/MSSQL/SQLite 지원 종료)
- [x] CLI(`main.go`)에서 non-postgres targetDb 진입 시 오류 종료 처리 (`requirePostgres` 동등 로직)
- [x] `internal/db/db.go`에서 불필요한 드라이버 imports 제거 (`go-sql-driver/mysql`, `go-sqlite3`, `go-mssqldb`)


---
## <a name="v23"></a> v23

### prd.md

App.tsx 분리 계획
Context
frontend/src/app/App.tsx가 3,279줄로 비대해져 있음. 타입, 상수, 유틸리티, UI 컴포넌트가 모두 한 파일에 있어 유지보수가 어려운 상태. 이를 논리적 단위로 분리하여 가독성과 유지보수성을 높임.
분리 구조
1. app/types.ts — 타입 정의 (lines 14-186)
모든 로컬 타입 추출:
	∙	RoleFilter, NoticeTone, PrecheckDecisionFilter, WsStatus, TableRunStatus, TableHistoryStatusFilter, TableSortOption, ObjectGroup
	∙	SourceState, TargetState, TargetTableEntry, CompareState, CompareFilter, SourceRecent
	∙	MigrationOptions, TableRunState, TableHistoryState, TableHistoryDetail
	∙	ValidationState, DdlEvent, DiscoverySummary, ReportSummary, GroupStats, GroupedStats
	∙	WsProgressMsg, MetricsState, Locale
2. app/constants.ts — 상수 (lines 188-252)
	∙	SOURCE_RECENT_KEY, SOURCE_REMEMBER_KEY, TARGET_RECENT_KEY, UI_LOCALE_KEY
	∙	UI_TEXT (다국어 사전)
	∙	DEFAULT_OPTIONS
3. app/utils.ts — 유틸리티 함수 (lines 254-427)
	∙	normalizeTableKey, formatHistoryTime
	∙	loadRememberPassword, loadSourceRecent, loadTargetRecent, loadLocale
	∙	toBool, toNumber, toString, toObjectGroup, toStringArray
	∙	isObjectGroupModeEnabled, createSessionId
	∙	wsStatusLabel, tableStatusLabel, tableStatusBadgeClass
	∙	historyStatusLabel, historyStatusBadgeClass, parseReplayedTables
4. app/HeaderBar.tsx — 헤더 영역 (~lines 1730-1786)
	∙	앱 타이틀, 언어 토글, 인증 상태, 로그인/로그아웃 버튼
	∙	히스토리/연결 버튼
5. app/RecentSource.tsx — 최근 소스 저장 영역 (~lines 1788-1816)
	∙	비밀번호 기억 체크박스, 복원/초기화 버튼
6. app/ConnectionForms.tsx — 소스/타겟 연결 폼 (~lines 1818-2017)
	∙	Source (Oracle) 폼, Target (PostgreSQL) 폼
	∙	Compare 상태 표시
7. app/TableSelection.tsx — 테이블 선택 영역 (~lines 2020-2480)
	∙	테이블 목록, 필터, 정렬, 검색
	∙	Object group 선택
	∙	전체 선택/해제
8. app/MigrationOptionsPanel.tsx — 마이그레이션 옵션 (~lines 2480-2838)
	∙	옵션 체크박스, 고급 설정
	∙	Pre-check 섹션
	∙	마이그레이션 시작 버튼
9. app/RunStatus.tsx — 실행 상태 (~lines 2842-3111)
	∙	진행률 바, 통계 카드
	∙	테이블별 진행 상태 목록
	∙	메트릭스, 유효성 검증 결과
10. app/CredentialsPanel.tsx — 저장된 연결 패널 (~lines 3115-3190)
	∙	사이드 패널 (aside), 필터, 연결 목록
11. app/HistoryPanel.tsx — 히스토리 패널 (~lines 3190-3250)
	∙	사이드 패널, 히스토리 목록, 재실행 버튼
12. app/LoginModal.tsx — 로그인 모달 (~lines 3250-3278)
	∙	로그인 폼 오버레이
App.tsx (최종)
	∙	모든 state (useState), effects (useEffect), useMemo, handler 함수 유지
	∙	각 컴포넌트를 import하여 조합
	∙	props로 필요한 state와 handler 전달
구현 순서
	1.	types.ts 생성 → App.tsx에서 import로 교체
	2.	constants.ts 생성 → App.tsx에서 import로 교체
	3.	utils.ts 생성 → App.tsx에서 import로 교체
	4.	UI 컴포넌트 추출 (각각):
	∙	LoginModal → HeaderBar → RecentSource → ConnectionForms → TableSelection → MigrationOptionsPanel → RunStatus → CredentialsPanel → HistoryPanel
	∙	작은 컴포넌트부터 먼저 추출 (의존성이 적은 것부터)
	5.	App.tsx를 orchestrator로 정리
핵심 원칙
	∙	동작 변경 없음: 순수 리팩토링, 기능 변경 없음
	∙	Props 패턴: 각 컴포넌트에 필요한 state/handler를 props로 전달
	∙	t/tr 함수: locale 기반 t/tr 헬퍼는 utils.ts에서 팩토리로 export하거나, 각 컴포넌트에 locale prop 전달
	∙	flat 구조: frontend/src/app/ 아래 파일로 유지 (하위 디렉토리 없음)
검증
	∙	npm run build (또는 npx vite build) 성공 확인
	∙	npm run test 기존 테스트 통과 확인
	∙	UI 동작 동일 (시각적 변경 없음)

### spec.md

# 기술 사양서 (Technical Specifications)

## 1. 개요

이번 리팩토링의 목적은 `frontend/src/app/App.tsx`의 대형 단일 파일 구조를 **기능 변경 없이 리팩토링**하여 유지보수성과 테스트 용이성을 높이는 것이다.

- 현재 상태: `App.tsx` 단일 파일에 타입/상수/유틸/핸들러/UI가 혼재
- 목표 상태: 타입, 상수, 유틸, 화면 컴포넌트를 파일 단위로 분리하고 `App.tsx`는 orchestration 역할만 담당
- 비목표(Non-goal): UX 변경, API 스펙 변경, 비즈니스 로직 변경

---

## 2. 파일 구조

분리 후 구조(평면 구조 유지):

- `frontend/src/app/App.tsx`
- `frontend/src/app/types.ts`
- `frontend/src/app/constants.ts`
- `frontend/src/app/utils.ts`
- `frontend/src/app/HeaderBar.tsx`
- `frontend/src/app/RecentSource.tsx`
- `frontend/src/app/ConnectionForms.tsx`
- `frontend/src/app/TableSelection.tsx`
- `frontend/src/app/MigrationOptionsPanel.tsx`
- `frontend/src/app/RunStatus.tsx`
- `frontend/src/app/CredentialsPanel.tsx`
- `frontend/src/app/HistoryPanel.tsx`
- `frontend/src/app/LoginModal.tsx`

---

## 3. 기능 요구사항 (FR)

### FR-1. 타입 분리 (`types.ts`)

`App.tsx` 내 로컬 타입 정의를 `types.ts`로 이동하고 재-export/import 구조를 정리한다.

대상 예시:
- RoleFilter, NoticeTone, PrecheckDecisionFilter
- WsStatus, TableRunStatus, TableHistoryStatusFilter, TableSortOption
- SourceState, TargetState, TargetTableEntry, CompareState
- MigrationOptions, TableRunState, TableHistoryState, ValidationState
- DdlEvent, DiscoverySummary, ReportSummary, GroupStats, GroupedStats
- WsProgressMsg, MetricsState, Locale

### FR-2. 상수 분리 (`constants.ts`)

다음 상수를 `constants.ts`로 이동한다.
- 로컬 스토리지 키: `SOURCE_RECENT_KEY`, `SOURCE_REMEMBER_KEY`, `TARGET_RECENT_KEY`, `UI_LOCALE_KEY`
- 다국어 사전: `UI_TEXT`
- 초기 옵션: `DEFAULT_OPTIONS`

### FR-3. 유틸 분리 (`utils.ts`)

App 내부 유틸을 `utils.ts`로 분리한다.

대상 예시:
- key/time 유틸: `normalizeTableKey`, `formatHistoryTime`
- 저장소 로딩: `loadRememberPassword`, `loadSourceRecent`, `loadTargetRecent`, `loadLocale`
- 파서: `toBool`, `toNumber`, `toString`, `toObjectGroup`, `toStringArray`
- 도메인 유틸: `isObjectGroupModeEnabled`, `createSessionId`
- 상태 라벨/배지: `wsStatusLabel`, `tableStatusLabel`, `tableStatusBadgeClass`, `historyStatusLabel`, `historyStatusBadgeClass`
- 히스토리 파서: `parseReplayedTables`

### FR-4. UI 컴포넌트 분리

UI 블록을 의미 단위로 컴포넌트로 추출한다.

- `LoginModal`: 로그인 오버레이
- `HeaderBar`: 타이틀/언어 토글/세션 버튼
- `RecentSource`: 최근 접속정보 및 비밀번호 기억
- `ConnectionForms`: 소스/타겟 입력 폼
- `TableSelection`: 테이블 목록, 필터/정렬, object-group
- `MigrationOptionsPanel`: 실행 옵션 + pre-check
- `RunStatus`: 진행률, 테이블 상태, 메트릭
- `CredentialsPanel`: 저장된 연결 목록
- `HistoryPanel`: 실행 이력/재실행

### FR-5. App.tsx Orchestrator화

`App.tsx`는 아래 책임만 유지한다.
- 상태 관리(`useState`, `useMemo`, `useEffect`)
- API 호출/핸들러 함수
- 하위 컴포넌트 조합 및 props 전달

---

## 4. 비기능 요구사항 (NFR)

- **NFR-1. 동작 동일성**: 기능/요청/응답/렌더링 결과의 의미가 이전 릴리스와 동일해야 한다.
- **NFR-2. 타입 안정성**: `npm run typecheck` 통과
- **NFR-3. 빌드 안정성**: `npm run build` 통과
- **NFR-4. 테스트 안정성**: 기존 `App.test.tsx` 및 관련 테스트 통과
- **NFR-5. 구조 일관성**: `frontend/src/app` 하위 디렉터리 추가 없이 평면 구조 유지

---

## 5. 인터페이스/호환성

- 백엔드 API 변경 없음
- WebSocket 메시지 포맷 변경 없음
- LocalStorage key 변경 없음
- i18n 키/문구 의미 변경 없음

---

## 6. 구현 순서

1. `types.ts` 생성 및 타입 이동
2. `constants.ts` 생성 및 상수 이동
3. `utils.ts` 생성 및 유틸 이동
4. 의존성 낮은 UI부터 컴포넌트 추출
   - `LoginModal` → `HeaderBar` → `RecentSource`
5. 폼/선택/옵션/상태 영역 컴포넌트 추출
   - `ConnectionForms` → `TableSelection` → `MigrationOptionsPanel` → `RunStatus`
6. 패널 컴포넌트 추출
   - `CredentialsPanel` → `HistoryPanel`
7. `App.tsx` import/props 정리 및 dead code 제거

---

## 7. 검증 계획

- 프론트엔드
  - `cd frontend && npm run test`
  - `cd frontend && npm run typecheck`
  - `cd frontend && npm run build`
- 회귀 검증
  - 로그인/로그아웃
  - 연결 테스트/테이블 조회
  - pre-check 실행
  - 마이그레이션 시작/진행률 표시
  - 히스토리 조회/재실행

---

## 8. 리스크 및 대응

- 리스크: props drilling 증가로 컴포넌트 시그니처 복잡도 상승
  - 대응: props 타입 별도 선언, 그룹화된 props 인터페이스 사용
- 리스크: 유틸 분리 중 참조 누락
  - 대응: 타입체크 + 테스트 + 빌드 3중 검증
- 리스크: 렌더링 순서/조건 변경으로 미세 UI 회귀
  - 대응: 기존 조건식 유지, 단계별 추출 후 즉시 테스트


### tasks.md

# 작업 목록 (Tasks)

## 목표: App.tsx 리팩토링 분리 + 완료 문서 정리

### 1) 문서 정리
- [x] `docs/v23-prd.md` → `docs/v23/prd.md`로 이동
- [x] `docs/v22/*` → `docs/complete/v22/*`로 정리
- [x] `docs/v23/spec.md` 작성
- [x] `docs/v23/tasks.md` 작성

---

### 2) 타입/상수/유틸 분리
- [x] `frontend/src/app/types.ts` 생성 및 타입 이동
- [x] `frontend/src/app/constants.ts` 생성 및 상수 이동
- [x] `frontend/src/app/utils.ts` 생성 및 유틸 이동
- [x] `App.tsx` 내부 중복 타입/상수/유틸 제거

---

### 3) UI 컴포넌트 추출
- [x] `LoginModal.tsx` 추출
- [x] `HeaderBar.tsx` 추출
- [x] `RecentSource.tsx` 추출
- [x] `ConnectionForms.tsx` 추출
- [x] `TableSelection.tsx` 추출
- [x] `MigrationOptionsPanel.tsx` 추출
- [x] `RunStatus.tsx` 추출
- [x] `CredentialsPanel.tsx` 추출
- [x] `HistoryPanel.tsx` 추출

---

### 4) App.tsx Orchestrator 정리
- [x] 하위 컴포넌트 import/props 연결
- [x] 핸들러/useEffect/useMemo 의존성 점검
- [x] dead code 및 사용하지 않는 import 제거

---

### 5) 검증
- [x] `cd frontend && npm run test`
- [x] `cd frontend && npm run typecheck`
- [x] `cd frontend && npm run build`
- [x] 주요 수동 시나리오 회귀 확인 (로그인/연결/pre-check/실행/히스토리)


---
## <a name="v24"></a> v24

### prd.md

# Product Requirements Document (PRD)

## 1. 개요
**프로젝트명:** 마이그레이션 Web UI/UX 고도화 (v24)  
**목표:** 최근 도입된 스텝 위저드(Step Wizard) 기반의 레이아웃 위에서, 개별 컴포넌트들의 사용성과 시각적 완성도를 엔터프라이즈급 수준으로 끌어올린다.

## 2. 배경 및 필요성
기존 마이그레이션 도구는 모든 기능이 한 화면에 노출되어 인지 부하가 높았으나, 스텝 위저드 도입으로 전체적인 흐름은 개선되었다. 
그러나 단계별 화면 내부(특히 테이블 선택과 마이그레이션 세부 옵션, 실행 모니터링)는 여전히 정보 밀집도가 높아 UX 개선이 절실하다.
- 수백 개의 테이블 중 원하는 것을 선택했는지 한눈에 확인하기 어렵다.
- 고급 성능 튜닝 옵션이 기본 옵션과 섞여 있어 초보 사용자에게 혼란을 준다.
- 실행 상태 화면이 단순한 텍스트와 프로그레스 바 위주라 한눈에 들어오는 가시성이 부족하다.
- 개발자와 DBA가 주로 사용하는 도구임에도 눈의 피로를 덜어주는 다크 모드를 지원하지 않는다.

## 3. 핵심 목표
1. **인지 부하 감소:** 복잡한 설정은 숨기고 자주 쓰는 필수 기능만 직관적으로 노출.
2. **조작 편의성 극대화:** 대량의 데이터를 다루는 테이블 선택 UI를 명확하게 개편.
3. **가시성 향상:** 데이터를 그래픽적 요소(차트, 통계 카드)로 표현하여 직관적인 모니터링 제공.
4. **접근성 및 테마:** 시스템 설정 및 사용자 선호도에 따른 다크/라이트 모드 환경 제공.

## 4. 세부 요구사항 (Requirements)

### 4.1 마이그레이션 옵션 그룹화 (Basic vs Advanced)
- **위치:** Step 2 - `MigrationOptionsPanel`
- **요구사항:**
  - 사용 빈도가 높은 **기본 설정(Basic)**과 성능/시스템 튜닝에 가까운 **고급 설정(Advanced)**을 분리한다.
  - **기본 설정 노출 항목:** `Object Group`, `Dry Run`, `With DDL`, `Validate`, `Truncate Target` 등
  - **고급 설정 숨김 항목:** `Workers`, `Batch Size`, `Copy Batch`, `DB Max Open/Idle/Life`, `Log JSON`, `Out File` 등
  - 고급 설정은 **아코디언(Accordion, 접기/펴기)** UI로 감싸 기본적으로 접혀(Collapsed) 있도록 처리한다.

### 4.2 듀얼 리스트박스 (Dual Listbox / Transfer List) 테이블 선택
- **위치:** Step 2 - `TableSelection`
- **요구사항:**
  - 기존의 단일 리스트 + 체크박스 형태를 탈피하여 2개의 패널로 분리한다.
    - **왼쪽 패널:** '선택 가능한 테이블 (Available)'
    - **오른쪽 패널:** '선택된 테이블 (Selected)'
  - 패널 사이에 이동을 위한 컨트롤 버튼을 배치한다: `>` (단일 이동), `<` (선택 해제), `>>` (전체 이동), `<<` (전체 선택 해제).
  - 각 패널 상단에 독립적인 검색(Search) 필드를 제공하여 방대한 테이블 목록을 쉽게 필터링할 수 있어야 한다.

### 4.3 모니터링 대시보드 시각화 강화
- **위치:** Step 3 - `RunStatusPanel`
- **요구사항:**
  - **통계 카드 (Stat Cards):** 화면 최상단에 마이그레이션 핵심 지표를 큰 폰트의 카드로 배치한다. (전체 진행률, 처리된 행 수, 초당 처리 속도(Rows/sec), 예상 남은 시간(ETA)).
  - **시각적 차트 도입:** 테이블 진행 상태를 단순히 텍스트 목록이 아닌 **원형 차트(Donut Chart)** 또는 진행률 서클 형태로 표현하여 직관성을 높인다.
  - **서버 메트릭 강조:** 웹소켓으로 수신되는 CPU / Memory 사용량을 실시간으로 업데이트되는 뱃지 또는 미니 그래프로 시각화한다.

### 4.4 다크 모드 (Dark Mode) 지원
- **위치:** 글로벌 (앱 전체)
- **요구사항:**
  - TailwindCSS의 `dark:` variant를 사용하여 라이트 모드와 대비되는 심미적인 다크 테마 컬러 팔레트를 적용한다. (예: `bg-slate-900`, `text-slate-200` 등).
  - `HeaderBar`에 테마 전환 스위치 (☀️ / 🌙 아이콘)를 배치한다.
  - 사용자가 설정한 테마 상태는 브라우저 `localStorage` (키: `ui_theme`)에 영구 저장하여 새로고침 시에도 유지되도록 한다.
  - (선택) 시스템 테마(OS 기본 설정)를 감지하여 기본값으로 적용한다 (`prefers-color-scheme: dark`).

## 5. 제외 범위 (Out of Scope)
- 백엔드(Go) API의 비즈니스 로직 변경 및 추가. (순수 프론트엔드 React/TailwindCSS 개선으로 한정).
- 외부 차트 라이브러리(Chart.js, Recharts 등) 도입. 번들 사이즈 증가를 막기 위해 SVG 기반의 가벼운 커스텀 컴포넌트(Donut Chart 등)를 직접 구현하는 것을 권장.

## 6. 예상 결과물 (Deliverables)
- 변경된 React 컴포넌트 (`TableSelection.tsx`, `MigrationOptionsPanel.tsx`, `RunStatusPanel.tsx`, `HeaderBar.tsx`, `App.tsx` 등).
- SVG 기반의 심플 원형 차트 / 통계 카드 컴포넌트 추가.
- 다크 모드 제어를 위한 `useTheme` 커스텀 훅.


### spec.md

# 기술 명세서 (Technical Specification)

## 1. 개요
**버전:** v24  
**목표:** PRD에 정의된 4가지 주요 UI/UX 개선안(마이그레이션 옵션 그룹화, 듀얼 리스트박스, 대시보드 시각화, 다크 모드)을 구현하기 위한 프론트엔드 기술 상세 명세.

## 2. 기술 스택 및 제약 사항
- **프레임워크:** React 18 (TypeScript)
- **스타일링:** TailwindCSS v3+
- **상태 관리:** React Hooks (`useState`, `useEffect`, `useMemo` 등)
- **제약 사항:** 번들 사이즈 최소화를 위해 외부 무거운 UI/차트 라이브러리(Material-UI, Chart.js 등) 사용을 지양하고, **순수 TailwindCSS와 HTML/SVG 태그**를 활용하여 직접 구현한다.

---

## 3. 기능별 구현 명세

### 3.1 다크 모드 (Dark Mode) 지원
1. **Tailwind 설정 (`tailwind.config.ts`):**
   - `darkMode: 'class'` 옵션이 활성화되어 있는지 확인/추가한다.
2. **커스텀 훅 (`useTheme.ts`):**
   - `theme` 상태 관리 (`'light' | 'dark'`).
   - 초기화 로직: `localStorage.getItem('ui_theme')` 값을 우선 확인하고, 값이 없다면 OS 기본 설정(`window.matchMedia('(prefers-color-scheme: dark)').matches`)을 따른다.
   - 테마 변경 시 `document.documentElement.classList.toggle('dark', isDark)`를 실행하여 `<html>` 태그에 `dark` 클래스를 주입한다.
3. **UI 토글 (`HeaderBar.tsx`):**
   - 우측 상단 사용자 메뉴 옆에 ☀️/🌙 (Sun/Moon) 토글 버튼을 추가한다.
4. **스타일 적용:**
   - 프로젝트 전체 주요 컨테이너에 `dark:bg-slate-900`, `dark:text-slate-200`, `dark:border-slate-700` 등 다크 테마 유틸리티 클래스를 일괄 적용한다.

### 3.2 테이블 선택: 듀얼 리스트박스 (Transfer List)
- **컴포넌트:** `TableSelection.tsx`
- **상태 관리 변경:**
  - `selectedTables`: 기존 상태 유지.
  - `availableTables`: `allTables` 중 `selectedTables`에 없는 항목들 (Derived State).
  - 왼쪽/오른쪽 각각을 위한 검색어 상태 (`leftSearch`, `rightSearch`).
  - (선택) 개별 이동을 위한 내부 선택 상태 (`leftChecked`, `rightChecked`).
- **UI 레이아웃 (Grid 3-cols):**
  - **Left Panel (Available):** 검색 바 + 가상 스크롤 또는 오버플로우가 적용된 리스트.
  - **Center Controls:** 세로로 정렬된 4개의 버튼 (`>>` 전체 추가, `>` 선택 추가, `<` 선택 제거, `<<` 전체 제거).
  - **Right Panel (Selected):** 검색 바 + 오버플로우가 적용된 리스트.
- **최적화:** 방대한 테이블 렌더링 시 브라우저 버벅임을 막기 위해 `useMemo`를 통한 리스트 필터링 메모이제이션 필수.

### 3.3 마이그레이션 옵션 그룹화 (Accordion)
- **컴포넌트:** `MigrationOptionsPanel.tsx`
- **상태 관리:**
  - `isAdvancedOpen` (boolean): 고급 설정 패널의 열림/닫힘 상태.
- **UI 구조:**
  - **Basic Options Area:** `Object Group`, `With DDL`, `Truncate`, `Validate`, `Dry Run` 등 핵심 스위치 및 라디오 버튼 상단 배치.
  - **Advanced Toggle Button:** `고급 설정 보기 ▼` / `고급 설정 숨기기 ▲` 토글 버튼.
  - **Advanced Options Area:** `isAdvancedOpen`이 true일 때만 렌더링. `Batch Size`, `Workers`, `DB Max Open` 등 숫자 입력 폼 위주로 배치.

### 3.4 모니터링 대시보드 시각화 (Stat Cards & Donut Chart)
- **컴포넌트:** `RunStatusPanel.tsx`
- **통계 카드 (Stat Cards):**
  - `metrics` (CPU/Mem), `overallPercent`, `processedRows`, `rowsPerSecond`, `etaSeconds` 데이터를 활용하여 화면 상단에 4~5개의 Grid 카드로 렌더링.
  - 예: [ ⏱ 남은 시간: 05:20 ] [ 🚀 속도: 15,000 rows/s ] [ 📈 진행률: 85% ]
- **진행률 원형 차트 (Donut Chart):**
  - `<svg>` 태그 기반의 커스텀 컴포넌트 구현 (`CircularProgress.tsx` 분리 또는 내부 구현).
  - `<circle>`의 `stroke-dasharray`와 `stroke-dashoffset` 속성을 계산식 `(100 - percent) / 100 * circumference`에 대입하여 애니메이션 효과 적용.
- **상세 테이블 목록:**
  - 기존 테이블별 진행 상태(Progress)는 차트 하단에 컴팩트한 리스트/그리드 형태로 유지하되, 성공/실패 여부를 색상(Green/Red) 뱃지로 더 명확히 표현.

## 4. 데이터 플로우 및 인터페이스
본 개선안은 철저하게 **View 레이어의 표현 방식 변경**에 집중되어 있으므로, 기존 API 통신 인터페이스(`/api/migrate`, `/api/tables`, Websocket 등)와 Payload 스키마는 **전혀 변경하지 않는다.**

## 5. 단계별 적용 전략
안정적인 도입을 위해 다음 순서로 점진적 리팩토링 및 커밋을 진행한다.
1. `tailwind.config.ts` 다크 모드 활성화 및 `useTheme` 훅 구현
2. `MigrationOptionsPanel.tsx` 옵션 그룹화
3. `TableSelection.tsx` 듀얼 리스트박스 구현
4. `RunStatusPanel.tsx` 시각화 컴포넌트(차트, 통계 뱃지) 구현
5. 전체 화면 다크 모드 스타일(`dark:`) 튜닝 및 QA


### tasks.md

# 구현 작업 목록 (Tasks)

## [Phase 1] 기반 설정 및 다크 모드 (Dark Mode) 구현
- [ ] `frontend/tailwind.config.ts` 파일에 `darkMode: 'class'` 옵션 추가
- [ ] `frontend/src/app/hooks/useTheme.ts` 커스텀 훅 생성 (상태 관리 및 `localStorage` 연동, `<html class="dark">` 토글 로직)
- [ ] `frontend/src/app/components/HeaderBar.tsx`에 다크 모드 토글 버튼(해/달 아이콘) 추가
- [ ] `frontend/src/app/App.tsx` 최상단 컨테이너에 다크 모드용 배경/텍스트 색상(`dark:bg-slate-900 dark:text-slate-100`) 적용

## [Phase 2] 마이그레이션 옵션 그룹화 (Options Grouping)
- [ ] `frontend/src/app/components/MigrationOptionsPanel.tsx` 내부 상태에 `isAdvancedOpen` 추가
- [ ] 기존 옵션들을 분류:
  - **Basic Options:** `Object Group`, `With DDL`, `Validate`, `Truncate Target`, `Dry Run`
  - **Advanced Options:** `Batch Size`, `Copy Batch`, `Workers`, `Log JSON`, `DB Max Open/Idle/Life`, `Out File`
- [ ] Advanced 영역을 감싸는 컨테이너와 토글 버튼(`고급 설정 보기 ▼` 등) UI 구현 및 Tailwind 트랜지션(transition) 적용
- [ ] 다크 모드 지원을 위한 폼 필드 스타일 업데이트 (`dark:bg-slate-800`, `dark:border-slate-600` 등)

## [Phase 3] 듀얼 리스트박스 테이블 선택 (Dual Listbox)
- [ ] `frontend/src/app/components/TableSelection.tsx` 파일 백업 및 리팩토링 준비
- [ ] `availableTables` (선택되지 않은 테이블)와 `selectedTables`로 데이터 분리하는 `useMemo` 작성
- [ ] 3-Column Grid 레이아웃(왼쪽 리스트, 가운데 컨트롤, 오른쪽 리스트) 마크업 추가
- [ ] 양쪽 패널 각각에 대한 독립적인 검색 상태(`leftSearch`, `rightSearch`) 추가 및 필터링 로직 구현
- [ ] 중간 컨트롤 버튼(`>`, `<`, `>>`, `<<`) 핸들러 함수 구현
- [ ] 다크 모드 호환 리스트 아이템 UI 적용

## [Phase 4] 모니터링 대시보드 시각화 (Dashboard Enhancements)
- [ ] `frontend/src/app/components/RunStatusPanel.tsx` 수정 준비
- [ ] 상단 통계 카드(Stat Cards) Grid 마크업 추가:
  - 진행률 (`overallPercent`%)
  - 남은 시간 (`etaSeconds` 포맷팅)
  - 속도 (`rowsPerSecond` rows/s)
  - 처리된 행 수 (`processedRows` / `expectedRows`)
- [ ] 커스텀 SVG 기반의 진행률 서클(Donut Chart) 컴포넌트 내부 구현
- [ ] 서버 Metrics(CPU, Mem)를 보여주는 별도 배지 또는 우측 상단 UI 추가
- [ ] 다크 모드 시 차트 색상 및 카드 배경색 호환 확인

## [Phase 5] 최종 점검 및 최적화
- [ ] 각 컴포넌트(`App.tsx`, 모달 패널, Credentials 폼 등) 다크 모드 호환성 통합 점검
- [ ] `npm run verify:fast` (또는 `tsc --noEmit`) 실행하여 타입스크립트 에러 확인
- [ ] 로컬 테스트 서버 구동 및 전체 시나리오(로그인 -> 연결 -> 옵션 토글 -> 테이블 이동 -> 실행 모니터링) E2E 수동 테스트
- [ ] 변경 사항 커밋 및 푸시


---
## <a name="v25"></a> v25

### prd.md

# PRD: UI Readability & Wizard Flow Correction (v25)

## 1. Goal
Improve UI text contrast for better readability and ensure the 3-step wizard can be fully reset/restarted after a migration run is completed.

## 2. Problem Statement
- **Low Contrast:** Many labels and descriptions use `text-slate-500` or `text-slate-400`, which is too light against the background, especially in Dark Mode.
- **Restart Loop:** Once a migration reaches Phase 3 (Monitoring), there is no clear path to return to Step 1 (Source/Target config) to start a new session without refreshing the page.

## 3. Requirements
- **Color Audit:** Replace low-contrast grey text (`slate-500/400`) with higher contrast alternatives (`slate-700/600` for Light, `slate-300/200` for Dark) where important information is displayed.
- **Reset Logic:** 
    - The "Close Monitoring" button in `RunStatusPanel` must trigger a full state reset.
    - `App.tsx` must reset the `step` state to `1` when `resetRunState` is called.
- **Visual Feedback:** Ensure active interactive elements are clearly distinguishable from disabled/readonly text.

## 4. Success Criteria
- WCAG-compliant contrast for primary text.
- User can go from Monitoring back to Step 1 with a single click and start a new migration.
- `make offline` (all tests) continues to pass.


### tasks.md

# Tasks: UI Readability & Wizard Flow (v25)

## Phase 1: Contrast Fixes
- [ ] Audit `App.tsx` and change `slate-500` to `slate-700` (Light) and `slate-400` to `slate-300` (Dark).
- [ ] Audit `TableSelection.tsx` for table names and counts readability.
- [ ] Audit `MigrationOptionsPanel.tsx` for labels and help text.
- [ ] Audit `RunStatusPanel.tsx` for monitoring metrics and table list.

## Phase 2: Flow Fixes
- [ ] Update `App.tsx` `handleResetRunState` to include `setStep(1)`.
- [ ] Ensure `useMigrationRun.ts` `resetRunState` properly clears `runSessionId` and `migrationBusy`.
- [ ] Verify that clicking "Close Monitoring" correctly navigates the user back to Step 1.

## Phase 3: Validation
- [ ] Run `make offline` to ensure no test regressions.
- [ ] Manual check of Step 1 -> 2 -> 3 -> Reset -> Step 1 flow.


---
## <a name="v26"></a> v26

### prd.md

# PRD: UI Optimization for Large Scale & Large Monitors (v26)

## 1. Goal
Improve the user experience when migrating a large number of tables (100+) and optimize the layout for modern large/wide monitors to utilize available screen real estate effectively.

## 2. Problem Statement
- **Narrow Layout:** On large monitors, the UI is restricted to `max-w-7xl` (1280px), leaving excessive whitespace on the sides and compressing complex components.
- **Table List Scaling:** When selecting or monitoring many tables, the lists either feel too small (Step 2) or grow excessively long (Step 3), making it difficult to maintain a global view.
- **Visual Breakage:** Long table names or many concurrent status updates can cause layout shifting or wrapping that degrades readability.

## 3. Requirements

### 3.1 Fluid Layout for Large Screens
- **Expanded Container:** Increase the maximum container width from `max-w-7xl` to `max-w-[1600px]` (or `max-w-[90vw]`) to give components more room to breathe on wide displays.
- **Responsive Grids:** 
    - In `RunStatusPanel`, increase grid columns for wider screens (e.g., `2xl:grid-cols-4`, `3xl:grid-cols-6`).
    - Adjust `TableSelection` side-by-side layout to use more horizontal space.

### 3.2 Scalable Table Selection (Step 2)
- **Increased Height:** The available/selected table lists should use more vertical space (`h-[500px]` or `min-h-[40vh]`) instead of the current `h-64`.
- **Top-aligned Controls:** Selection buttons (Add/Remove) are moved to the top of each panel for easier access when many tables are present.
- **Multi-term Search:** Support comma-separated search terms in both "Available" and "Selected" table filters (e.g., `TB_USER, TB_ORDER`).
- **Long Name Handling:** Ensure table names truncate gracefully with tooltips or provide enough horizontal width to avoid line breaks.

### 3.3 Optimized Progress Monitoring (Step 3)
- **Scrollable Progress Section:** The "Detailed Table Progress" grid should be contained within a scrollable area (`max-h-[60vh]`) with a sticky header if necessary, preventing the entire page from becoming excessively long.
- **Filtering:** Add a quick filter (search box) specifically for the "Detailed Table Progress" list to quickly locate specific tables among hundreds.
- **Compact View Option:** (Optional/Stretch) Add a toggle to switch between "Card" view and "Row" view for table progress.

### 3.4 Visual Polish
- **Sticky Actions:** Ensure primary actions (Next, Start Migration, Stop) remain easily accessible even when lists are scrolled.
- **Loading States:** Ensure smooth transitions when loading large metadata sets.

## 4. Success Criteria
- UI fills wide screens appropriately without looking "stretched."
- 200+ tables can be selected and monitored without layout breakage or excessive page-level scrolling.
- "Search" and "Filter" functionality remains performant with high table counts.
- `make offline` (all tests) continues to pass.


### spec.md

# SPEC: UI Optimization for Large Scale & Large Monitors (v26)

## 1. Introduction
This spec details the technical changes required to support 4K/Ultrawide monitors and large metadata sets (100+ tables) in the migration UI.

## 2. Structural Layout Adjustments

### 2.1 Main Container (`App.tsx`)
- Change: `max-w-7xl` (1280px) to `max-w-[1600px]` in Step 2 & 3.
- Reasoning: 1280px is too narrow for large displays, making side-by-side table lists feel cramped. 1600px provides a better balance between line length and space utilization.

### 2.2 Table Selection Component (`TableSelection.tsx`)
- Change: Height of available/selected table containers from `h-64` to `h-[500px]`.
- Change: Relocate "Add All/Selected" and "Remove All/Selected" buttons to the top of their respective panels (Left/Right) for better visibility.
- Change: Update search filtering to split input by `,` and perform `OR` matching on multiple terms.
- Change: Added `truncate` and `hover:whitespace-normal` (or tooltips) for table names.
- Change: Ensure horizontal scroll in table body to prevent long names from pushing buttons off-screen.

### 2.3 Run Status Panel (`RunStatusPanel.tsx`)
- Change: Added `max-h-[60vh] overflow-auto` container around the table progress grid.
- Change: Responsive grid adjustment:
    ```tsx
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 3xl:grid-cols-5">
    ```
- Feature: Added a search input above the progress grid to filter tables by name.

## 3. UI/UX Refinement
- Step indicators should remain centered.
- Summary cards in Phase 3 should use more horizontal space to avoid excessive stacking.

## 4. Technical Configuration

### 4.1 Tailwind Configuration (`tailwind.config.ts`)
Add `3xl` breakpoint:
```ts
theme: {
  extend: {
    screens: {
      '3xl': '1920px',
    },
    // ...
  }
}
```

## 5. Non-Functional Requirements
- **Performance:** Filtering 500+ tables in-memory must be instantaneous (< 16ms).
- **Accessibility:** Ensure all new filters and scrollable areas are keyboard-navigable.


### tasks.md

# Tasks: UI Optimization for Large Scale & Large Monitors (v26)

## 1. Research & Analysis
- [x] Verify Tailwind configuration for custom breakpoints (if any).
- [x] Measure layout performance with 100+ simulated tables.

## 2. Layout Implementation
- [x] Update `App.tsx` container width: `max-w-7xl` -> `max-w-[1600px]`.
- [x] In `TableSelection.tsx`, update table list heights to `h-[500px]` and improve horizontal wrapping.
- [x] In `RunStatusPanel.tsx`, wrap table progress grid in a scrollable container with `max-h-[60vh]`.
- [x] Adjust grid columns for `RunStatusPanel` for `2xl` and `3xl` screens.
- [x] Move Table Selection buttons to the top for better accessibility.

## 3. Performance & Features
- [x] Implement search filter for detailed table progress in `RunStatusPanel.tsx`.
- [x] Support comma-separated multi-term search in Table Selection.
- [x] Add Google Login integration (Backend & Frontend).
- [x] Add Database Connection History saving logic.

## 4. Testing & Validation
- [x] Verify layout on high-resolution displays.
- [x] Test with "large" metadata (100+ tables).
- [x] Ensure mobile view (`sm` and `lg`) is still working as expected.
- [x] Run `make offline` to verify no regressions in unit/integration tests.


