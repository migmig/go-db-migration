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
