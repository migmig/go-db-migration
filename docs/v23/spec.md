# 기술 사양서 (Technical Specifications) - v23

## 1. 개요

v23의 목적은 `frontend/src/app/App.tsx`의 대형 단일 파일 구조를 **기능 변경 없이 리팩토링**하여 유지보수성과 테스트 용이성을 높이는 것이다.

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

- **NFR-1. 동작 동일성**: 기능/요청/응답/렌더링 결과의 의미가 v22와 동일해야 한다.
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
