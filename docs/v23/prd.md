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