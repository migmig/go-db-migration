# 기술 명세서: Oracle에서 PostgreSQL 마이그레이션 도구 v2

## 1. 소개
이 명세서는 직접 데이터베이스 마이그레이션, 스키마 검색 및 최적화된 리소스 관리에 중점을 둔 마이그레이션 도구 v2의 기술 아키텍처 및 설계를 정의합니다.

## 2. 업데이트된 기술 스택
- **Go 1.22 이상**
- **Oracle 드라이버:** `github.com/sijms/go-ora/v2`
- **PostgreSQL 드라이버:** `github.com/jackc/pgx/v5`
- **로깅:** `log/slog` (구조화된 로깅)

## 3. 아키텍처 개요

### 3.1 마이그레이션 모드
1. **파일 기반 (v1 레거시):** Oracle -> 메모리 -> SQL 파일.
2. **직접 마이그레이션 (v2):** Oracle -> 메모리 -> PostgreSQL (`COPY` 또는 배치 `INSERT`).

### 3.2 컴포넌트 설계
- **디스패처(Dispatcher):** 테이블 목록을 읽고, 작업을 생성하며, 작업자 풀을 관리합니다.
- **작업자(Worker):** 채널에서 작업(테이블 이름)을 소비하고, Oracle에서 추출을 수행하며, 파일 또는 PostgreSQL에 쓰는 작업을 처리합니다.
- **DDL 생성기(DDL Generator):** Oracle 메타데이터를 쿼리하여 호환되는 PostgreSQL `CREATE TABLE` 문을 구성합니다.

## 4. 구현 세부 사항

### 4.1 작업자 풀 (Worker Pool)
- 작업 채널과 `sync.WaitGroup`을 사용하여 작업자 풀을 구현합니다.
- `--workers` 플래그는 동시 테이블 처리기의 수를 결정합니다.
- 경합을 피하기 위해 각 작업자는 고유한 Oracle 및 (선택적으로) PostgreSQL 연결을 유지하거나, 스레드 안전 풀을 사용합니다.

### 4.2 직접 PostgreSQL 마이그레이션
- `pgx.Conn` 또는 `pgxpool.Pool`을 사용합니다.
- **기본 방법:** 고성능 대량 로드를 위해 `pgx.Conn.CopyFrom`을 통한 `COPY` 명령.
- **대체 방법:** 매개변수화된 쿼리를 사용한 배치 `INSERT` 문.

### 4.3 DDL 매핑 (Oracle에서 PostgreSQL로)
| Oracle 타입 | PostgreSQL 타입 | 참고 |
| --- | --- | --- |
| `NUMBER(*, 0)` | `bigint` / `integer` | 정밀도 기반. |
| `NUMBER(*, >0)` | `numeric` | |
| `VARCHAR2(n)`, `NVARCHAR2(n)` | `text` 또는 `varchar(n)` | |
| `DATE`, `TIMESTAMP` | `timestamp` | |
| `CLOB` | `text` | |
| `BLOB`, `RAW` | `bytea` | |

### 4.4 예행 연습 (Dry Run) 로직
- `--dry-run`이 활성화된 경우:
  - Oracle에 대한 연결을 설정합니다.
  - 각 테이블에 대해 `SELECT COUNT(*) FROM table`을 쿼리합니다.
  - 보고: "테이블 X: 약 Y개의 행이 마이그레이션됩니다."
  - 출력 파일을 열거나 대상에 삽입을 실행하지 않습니다.

### 4.5 구조화된 로깅 (`slog`)
- JSON 플래그가 설정된 경우 `slog.New(slog.NewJSONHandler(os.Stdout, nil))`로 전역 로거를 초기화하고, 그렇지 않은 경우 TextHandler를 사용합니다.
- 로그 컨텍스트: `slog.Info("processing table", "table", tableName, "status", "started")`.

## 5. 보안 및 안전성
- **PostgreSQL DSN:** `--pg-url` 플래그 또는 `PG_URL` 환경 변수를 통해 처리합니다.
- **트랜잭션 안전성:** 직접 삽입의 경우, 테이블당 원자적 결과를 보장하기 위해 각 테이블 마이그레이션을 트랜잭션으로 래핑하는 것을 고려하십시오.
