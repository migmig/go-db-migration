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
