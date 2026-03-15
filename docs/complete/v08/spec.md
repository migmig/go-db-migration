# 기술 명세서 - v8 (확장성 및 엔터프라이즈 기능)

## 1. 웹 UI 다중 세션 지원
- **세션 관리**: 글로벌 싱글톤 패턴을 대체하여 여러 `WebSocketTracker` 인스턴스를 처리하기 위해 `web/server.go` 및 `ws/tracker.go`에 `SessionManager`를 도입합니다.
- **세션 ID**: 웹페이지에 접속할 때 연결되는 각 클라이언트에 대해 UUID 기반의 `SessionID`를 생성합니다.
- **API 통합**: 페이로드에서 `SessionID`를 수락하도록 `/api/migrate` 엔드포인트를 수정하여 마이그레이션(`migration.Run()`)에 올바른 세션별 트래커가 주입되도록 보장합니다.
- **가비지 컬렉션(Garbage Collection)**: 오래되거나 연결이 끊긴 세션을 메모리에서 안전하게 제거하는 정리 로직을 구현합니다.
- **UI 추적**: 디버깅 목적으로 UI에 현재 연결된 세션 ID를 표시합니다.

## 2. 제약 조건(Constraints) 마이그레이션 지원
### 2.1. 기본값 (Default Values)
- `GetTableMetadata` (`db/oracle.go`)의 Oracle 메타데이터에서 `DATA_DEFAULT`를 추출합니다.
- 추출된 기본값을 생성된 `CREATE TABLE` DDL 문에 통합합니다.

### 2.2. 외래 키(Foreign Key) 및 CHECK 제약 조건
- **메타데이터 추출**: `ALL_CONSTRAINTS`에서 외래 키(FK) 및 CHECK 제약 조건을 추출하는 `GetConstraintMetadata` 함수를 생성합니다.
- **CLI/UI 통합**: `--with-constraints` 플래그 (웹 UI: "Include Constraints" 체크박스)를 추가합니다.
- **방언(Dialect) DDL**: `CreateConstraintDDL` 인터페이스를 정의하고 지원되는 각 방언에 대한 제약 조건 구문을 구현합니다.
- **후처리(Post-processing) 실행**: 참조 무결성 순서를 보장하기 위해 모든 테이블 구조가 생성되고 모든 데이터가 완전히 삽입/복사된 후 마지막 단계로 FK에 대해서만 `ALTER TABLE ADD CONSTRAINT`를 실행합니다.

## 3. 연결 풀 미세 조정 (Connection Pool Fine-Tuning)
- **구성 속성**: 고급 연결 풀 매개변수를 추가합니다:
  - `--db-max-open` (기본값: 0, 무제한)
  - `--db-max-idle` (기본값: 2)
  - `--db-max-life` (기본값: 0, 무제한)
- **애플리케이션 적용**: 이러한 설정을 Oracle 소스 `sql.DB` 인스턴스, Postgres 대상 `pgxpool.Config` 및 `db/connect.go`의 다른 모든 대상 `sql.DB` 인스턴스에 직접 적용합니다.
- **웹 UI**: "고급 설정(Advanced Settings)" 내의 새로운 "DB Connection Pool Tuning" 섹션 아래에 이러한 설정을 노출합니다.

## 4. 재개 가능한 마이그레이션 (Checkpoints)
- **상태 관리**: 마이그레이션 상태를 관리하기 위해 새로운 모듈 `migration/state.go`를 생성합니다.
- **체크포인트(Checkpointing)**: 각 처리된 테이블에 대해 완료된 오프셋 또는 최대 기본 키(Primary Key) 값을 기록하여 진행 상황을 로컬 파일 시스템(`.migration_state/{job_id}.json`)에 유지합니다.
- **재개 명령(Resume Command)**: 중단된 마이그레이션을 마지막으로 성공한 체크포인트에서 재개하기 위해 `--resume {job_id}` CLI 플래그를 도입합니다.
- **청크 쿼리(Chunking Queries)**: 데이터 전체를 다시 가져오지 않고 중단된 위치에서 올바르게 데이터 가져오기를 계속하기 위해 Oracle 데이터 추출용 PK 기반 청크/페이지 매김 쿼리를 구현합니다.
