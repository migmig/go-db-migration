# Go DB Migration v10 개발 태스크 (Tasks)

본 문서는 `prd.md` 및 `spec.md`를 바탕으로 개발자가 순차적으로 실행할 수 있도록 분할한 세부 작업 목록입니다.

## Phase 1: 구조 개편 및 사전 검증 도입 (Step 1 & 2 통합)
* [ ] **Task 1.1: UI 레이아웃 개편**
  * `index.html`에서 타겟 DB 설정 폼(Target DB 선택, URL, Schema)을 Step 1 하단 또는 Step 2의 메인 뷰에서 Step 1으로 이동.
  * 기존 "Connect & Fetch Tables" 플로우에 "Test Target Connection" 플로우 추가.
* [ ] **Task 1.2: 타겟 DB 검증 API 추가**
  * `internal/web/server.go`에 `POST /api/test-target` 엔드포인트 구현.
  * `internal/db/db.go` 등을 활용해 입력받은 URL과 드라이버로 DB Ping 테스트만 수행하고 결과를 JSON으로 반환.
* [ ] **Task 1.3: Step 1 상태 관리 연동**
  * Source와 Target이 모두 유효하게 검증된 경우에만 Step 2로 넘어갈 수 있도록 클라이언트 사이드 검증 로직 수정.

## Phase 2: 설정 및 테이블 선택 UI 고도화 (Step 2)
* [ ] **Task 2.1: Data Table 컴포넌트 적용**
  * `index.html`의 테이블 리스트 영역(`tableList`)을 단순 `div` 리스트에서 `table` (또는 flex/grid 기반의 표 형태) 레이아웃으로 변경.
  * 컬럼 구조: 선택(Checkbox), 테이블명, 상태(옵션).
* [ ] **Task 2.2: 고급 설정 아코디언 구현**
  * Batch Size, Workers, Max Open 등 복잡한 설정들을 "Advanced Settings" 토글 버튼(혹은 `<details>` 태그 스타일의 컴포넌트) 내부에 숨김.
  * DDL 옵션도 트리거(체크박스)에 따라 하위 옵션이 슬라이드 다운되도록 CSS 애니메이션/JS 수정.
* [ ] **Task 2.3: 가상 스크롤 (Virtual Scrolling) 적용 검토**
  * 테이블 개수가 1000개 이상일 때를 대비해, 바닐라 JS로 가벼운 Virtual List(또는 Intersection Observer 활용)를 구현하거나 DOM 업데이트 최적화.

## Phase 3: 모니터링 대시보드 개편 (Step 3)
* [ ] **Task 3.1: 요약 위젯 (Summary Widget) 실시간화**
  * 현재 마이그레이션 종료 시(`all_done`)에만 뜨는 Summary Card를 마이그레이션 시작 시점부터 띄움.
  * 진행률(%), 성공, 실패 건수를 실시간 갱신.
* [ ] **Task 3.2: ETA 및 속도 계산 로직 구현**
  * 프론트엔드(`tracker.go`에서 전달되는 진행 데이터 기반)에 초당 처리 행 수(Rows/sec)와 남은 예상 시간(ETA)을 계산하여 UI에 표시.
* [ ] **Task 3.3: 상태별 탭(Tabs) 필터링 추가**
  * "전체", "진행중", "완료", "에러" 탭 버튼 추가.
  * 클릭 시 하단의 테이블 진행률 컨테이너 내의 아이템들을 `display: none/block`으로 필터링.
* [ ] **Task 3.4: 실패 테이블 에러 로그 UI 개선**
  * 에러 발생 시 상세 정보(Phase, Category 등)가 더 눈에 띄게 펼쳐지도록 아코디언 컴포넌트 스타일 적용.

## Phase 4: 기능 안정화 및 Retry (옵션/추가 스펙)
* [ ] **Task 4.1: 단일 테이블 Retry 버튼 UI 추가**
  * 에러가 발생한 테이블(`status === 'error'`)의 아이템 우측에 `[재시도]` 버튼 렌더링.
* [ ] **Task 4.2: Retry API 엔드포인트 연동 (백엔드 지원 시)**
  * `POST /api/migrate/retry` (가칭) 엔드포인트에 해당 테이블 정보만 전송하여 해당 테이블의 마이그레이션만 다시 큐에 넣는 기능 구현.
* [ ] **Task 4.3: Edge Case 테스트 및 버그 수정**
  * 브라우저 탭 이동 시 타이머(ETA) 이슈 확인.
  * 다크모드 적용 시 누락된 컬러 변수(테이블 헤더, 탭 활성화 상태 등) 보완.
