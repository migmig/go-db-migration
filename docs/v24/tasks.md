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
