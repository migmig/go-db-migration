# 구현 작업: Oracle에서 PostgreSQL 마이그레이션 도구 v2

## 1단계: 종속성 관리 및 인프라
- [x] `go.mod`에서 Go를 1.22 이상으로 업그레이드.
- [x] PostgreSQL 드라이버 설치: `github.com/jackc/pgx/v5`.
- [x] 애플리케이션 전반에 걸쳐 구조화된 로깅을 위해 `log/slog` 구현.
- [x] 새로운 플래그(`--pg-url`, `--workers`, `--with-ddl`, `--dry-run`)를 처리하기 위한 구성(configuration) 구조체 구현.
- [x] 내부 패키지로 코드베이스 모듈화.

## 2단계: 작업자 풀 및 동시성
- [x] `n`개의 작업자 풀을 관리하기 위한 `Dispatcher` ( `Run` 함수 내) 구현.
- [x] `Job` 구조체 및 스레드 안전 작업자 메커니즘 구현.
- [x] 테이블 처리를 위한 단순한 `sync.WaitGroup` 루프를 작업자 풀로 교체.
- [x] 적절하고 우아한 종료 보장 (Run이 반환되기 전에 모든 작업자가 완료됨).

## 3단계: 직접 마이그레이션 구현
- [x] `pgxpool`을 사용한 PostgreSQL 연결 풀 관리 구현.
- [x] 고속 데이터 전송을 위해 `pgx.Conn.CopyFrom`을 사용하는 `DirectWriter` ( `MigrateTableDirect` 내) 구현.
- [ ] 호환성을 위해 매개변수화된 쿼리를 사용하는 대체(fallback) 배치 `INSERT` 메커니즘 구현.
- [x] 테이블 마이그레이션별 트랜잭션 지원 추가.

## 4단계: 스키마 및 DDL 생성
- [x] 정밀도, 스케일 및 제약 조건을 위한 Oracle 메타데이터 검색 ( `GetTableMetadata` 내) 구현.
- [x] Oracle 타입에서 PostgreSQL 타입으로의 매핑 함수 (`MapOracleToPostgres`) 구현.
- [x] `CREATE TABLE` 스크립트 생성 로직 (`GenerateCreateTableDDL`) 구현.
- [x] 데이터 삽입 전에 DDL을 실행하기 위한 `--with-ddl` 실행 흐름 추가.

## 5단계: 예행 연습 및 검증
- [x] 연결성을 검증하고 예상 행 수를 보고하기 위한 `--dry-run` 로직 ( `Run` 내) 구현.
- [x] 마이그레이션을 시작하기 전에 대상 테이블이 존재하는지 확인하는 검증 ( `MigrateTableDirect` 내) 구현.
- [x] 사전 점검 (풀 생성 및 dry-run 중 연결성 확인) 추가.

## 6단계: 테스트 및 품질 보증
- [x] `slog`를 사용하도록 단위 테스트 업데이트 (패키지 리팩토링을 통해 암시적으로 수행됨).
- [x] 작업자 풀 및 작업 디스패칭을 위한 새로운 단위 테스트 추가 (`worker_test.go`).
- [x] Oracle 및 PostgreSQL을 모두 시뮬레이션하기 위해 `pgx` 및 `sqlmock`을 사용하는 통합 테스트 추가 (`direct_test.go`).
- [ ] 파일 기반 마이그레이션과 직접 마이그레이션을 비교하는 성능 벤치마킹 수행.
- [x] `README.md`의 문서 및 예제 업데이트.
