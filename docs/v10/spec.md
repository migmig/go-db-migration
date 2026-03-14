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
