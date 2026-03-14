# 기술 명세서: Oracle에서 PostgreSQL 데이터 마이그레이션 CLI

## 1. 소개
이 문서는 Oracle 데이터베이스에서 PostgreSQL 호환 SQL 스크립트로 데이터를 마이그레이션하는 Go 기반 CLI 도구의 기술적 설계 및 구현 세부 사항을 정의합니다.

## 2. 기술 스택
- **언어:** Go 1.21 이상
- **Oracle 드라이버:** `github.com/sijms/go-ora/v2` (순수 Go 드라이버, Oracle Instant Client 불필요)
- **동시성:** Go 표준 라이브러리 `sync` (WaitGroups, Mutex) 및 `channels`

## 3. 아키텍처 및 설계
이 도구는 단일 바이너리 CLI로 작동합니다. 테이블 크기에 관계없이 적은 메모리 사용량을 보장하기 위해 스트리밍 아키텍처를 따릅니다.

### 3.1 데이터 흐름
1. **연결:** 제공된 DSN을 사용하여 소스 Oracle DB에 대한 연결을 설정합니다.
2. **메타데이터 발견:** 각 테이블에 대해 열 이름과 타입을 쿼리합니다.
3. **스트리밍 추출:** `SELECT * FROM <table>`을 실행하고 `sql.Rows`를 사용하여 행을 반복합니다.
4. **변환:** Oracle 타입을 PostgreSQL 호환 리터럴로 변환합니다.
5. **일괄 처리(Batching):** `batch` 크기에 도달할 때까지 행을 메모리에 누적합니다.
6. **쓰기:** `INSERT INTO` 문을 대상 파일에 형식화하고 씁니다.

## 4. 데이터 타입 매핑
이 도구는 다음 변환을 처리해야 합니다:

| Oracle 타입 | PostgreSQL 타입 | 변환 로직 |
| --- | --- | --- |
| `VARCHAR2`, `CHAR`, `NVARCHAR2` | `text` / `varchar` | 작은따옴표(`'`)를 두 개(`''`)로 만들어 이스케이프합니다. |
| `NUMBER` | `numeric` / `int` / `float` | 직접적인 문자열 표현. |
| `DATE`, `TIMESTAMP` | `timestamp` | `YYYY-MM-DD HH24:MI:SS.FF` 형식으로 포맷합니다. |
| `CLOB` | `text` | 큰 문자열로 취급하고 이스케이프 처리를 합니다. |
| `BLOB`, `RAW` | `bytea` | `\x...` 16진수 형식으로 변환합니다. |
| `NULL` | `NULL` | 명시적으로 `NULL`을 씁니다. |

## 5. CLI 인터페이스
애플리케이션은 표준 `flag` 패키지 또는 `cobra`와 같은 라이브러리를 사용합니다.

| 플래그 | 타입 | 설명 |
| --- | --- | --- |
| `-url` | string | Oracle DSN (예: `oracle://user:pass@host:port/service`) |
| `-user` | string | DB 사용자명 |
| `-password` | string | DB 비밀번호 |
| `-tables` | string | 쉼표로 구분된 테이블 목록 (예: `USERS,ORDERS`) |
| `-out` | string | 출력 파일명 (기본값: `migration.sql`) |
| `-batch` | int | `INSERT` 문당 행 수 (기본값: `1000`) |
| `-per-table` | bool | 별도의 파일 생성: `<TABLE>_migration.sql` |
| `-parallel` | bool | 여러 테이블을 동시에 처리 |

## 6. 구현 세부 사항

### 6.1 병렬 처리 (`--parallel`)
- 테이블 처리 루틴의 완료를 추적하기 위해 `sync.WaitGroup`을 사용합니다.
- 필요한 경우 동시성을 제한합니다 (PRD에서는 제한을 명시하지 않았지만, 작업자 풀(Worker Pool)은 향후 개선 사항이 될 수 있습니다).

### 6.2 파일 처리 및 동기화
- **테이블별 모드(Per-Table Mode):** 각 고루틴이 자체 파일을 열고 씁니다. 테이블 작업자 간의 동기화가 필요하지 않습니다.
- **단일 파일 모드 + 병렬 처리:** 서로 다른 테이블의 일괄 삽입(Bulk Insert) 블록이 섞이지 않도록 공유 `io.Writer` 또는 `*os.File`을 `sync.Mutex`로 보호해야 합니다.

### 6.3 성능 최적화
- **버퍼링된 I/O:** 시스템 호출을 최소화하기 위해 모든 파일 작업에 `bufio.Writer`를 사용합니다.
- **메모리 관리:** 행은 하나씩 처리되며, 디스크에 플러시하기 전에 현재 행 배치의 내용만 메모리에 유지됩니다.

## 7. 보안 및 안전성
- **SQL 인젝션:** 이 도구는 수동 실행을 위한 스크립트를 생성하므로 문자열 값을 이스케이프해야 합니다. Oracle 드라이버는 값을 `interface{}`로 반환해야 하며, 그런 다음 안전하게 캐스팅하고 형식화합니다.
- **자격 증명 처리:** 비밀번호는 플래그를 통해 허용되어야 하지만, 향후 버전에서는 더 나은 보안을 위해 환경 변수를 지원해야 합니다.
