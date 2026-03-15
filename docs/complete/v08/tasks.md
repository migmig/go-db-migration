# 구현 작업 - v8 (확장성 및 엔터프라이즈 기능)

## 1. 웹 UI 다중 세션 지원
- [x] **세션 관리자(Session Manager) 구현**
  - [x] 여러 `WebSocketTracker` 인스턴스를 처리하기 위해 `ws/tracker.go`에 `SessionManager` 구조체를 생성.
  - [x] 글로벌 `tracker` 싱글톤 제거.
- [x] **세션 ID 생성 및 처리**
  - [x] `web/server.go`에서 연결하는 클라이언트를 위한 UUID 기반 `SessionID` 생성.
  - [x] 요청 페이로드에서 `SessionID`를 수락하고 파싱하도록 `/api/migrate` 엔드포인트 업데이트.
  - [x] 올바른 세션별 트래커를 `migration.Run()`에 주입.
- [x] **세션 정리 (가비지 컬렉션)**
  - [x] 연결이 끊기거나 오래된 세션을 안전하게 닫고 메모리에서 제거하는 로직 구현.
- [x] **UI 업데이트**
  - [x] 디버깅 및 추적을 위해 웹 UI(`index.html`)에 연결된 `SessionID` 표시.

## 2. 제약 조건(Constraints) 마이그레이션 지원
- [x] **기본값 (Default Values)**
  - [x] `DATA_DEFAULT`를 추출하기 위해 `db/oracle.go`의 `GetTableMetadata` 업데이트.
  - [x] 추출된 기본값을 포함하도록 `CREATE TABLE` DDL 생성 업데이트.
- [x] **외래 키(Foreign Key) 및 CHECK 제약 조건**
  - [x] FK 및 CHECK 제약 조건을 추출하기 위해 `db/oracle.go`에 `GetConstraintMetadata` 함수 구현.
  - [x] CLI 구성에 `--with-constraints` 플래그 추가.
  - [x] 웹 UI(`index.html`)에 "Include Constraints" 체크박스 추가.
  - [x] dialect 패키지에 `CreateConstraintDDL` 인터페이스 정의.
  - [x] 지원되는 모든 방언(PostgreSQL, MySQL, MariaDB, SQLite, MSSQL)에 대해 `CreateConstraintDDL` 구현.
- [x] **후처리(Post-processing) 실행**
  - [x] FK 제약 조건 실행을 연기(defer)하도록 `migration/migration.go` 수정.
  - [x] 데이터 삽입이 완전히 완료된 후 추출된 모든 FK에 대해 `ALTER TABLE ADD CONSTRAINT`를 실행하는 마지막 단계 구현.

## 3. 연결 풀 미세 조정 (Connection Pool Fine-Tuning)
- [x] **구성 업데이트**
  - [x] CLI 플래그 및 내부 Config 구조체에 `--db-max-open`, `--db-max-idle`, `--db-max-life` 매개변수 추가.
- [x] **연결 적용**
  - [x] Oracle 소스 `sql.DB` 인스턴스에 풀 설정 적용.
  - [x] Postgres 대상 `pgxpool.Config`에 풀 설정 적용.
  - [x] `db/connect.go`의 다른 모든 대상 `sql.DB` 인스턴스에 풀 설정 적용.
- [x] **UI 업데이트**
  - [x] `index.html`의 "고급 설정(Advanced Settings)" 섹션 내에 "DB Connection Pool Tuning" 입력 추가.

## 4. 재개 가능한 마이그레이션 (Checkpoints)
- [x] **상태 관리**
  - [x] 상태 처리를 위한 `migration/state.go` 모듈 생성.
  - [x] `.migration_state/{job_id}.json`에 진행 상황(오프셋 또는 최대 PK 값)을 저장/유지하는 기능 구현.
- [x] **재개(Resume) 기능**
  - [x] `--resume {job_id}` CLI 플래그를 추가하고 초기 시작 로직에 통합.
- [x] **데이터 추출 리팩토링**
  - [x] Oracle 데이터 추출을 위한 PK 기반 청크/페이지 매김 쿼리 구현.
  - [x] 기존 데이터를 다시 가져오지 않고 마지막으로 기록된 체크포인트에서 동적으로 추출을 시작할 수 있도록 보장.
