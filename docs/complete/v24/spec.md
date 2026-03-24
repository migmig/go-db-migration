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
