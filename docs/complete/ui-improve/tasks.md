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
