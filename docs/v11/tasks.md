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
