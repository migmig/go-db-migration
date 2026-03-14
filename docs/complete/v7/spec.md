# 기술 명세서 - v7 (품질 개선 및 버그 수정)

## 1. 타입 매핑 정밀도 개선
### 1.1. MySQL `MapOracleType`
- `VARCHAR2(n)`: `VARCHAR(n)`으로 매핑. 만약 `n > 16383`인 경우, `LONGTEXT`로 매핑.
- `CHAR(n)`: `CHAR(n)`으로 매핑 (정밀도 포함).

### 1.2. MSSQL `MapOracleType`
- `VARCHAR2(n)`: 만약 `n <= 4000`인 경우, `NVARCHAR(n)`으로 매핑. 만약 `n > 4000`인 경우, `NVARCHAR(MAX)`로 매핑.
- `CHAR(n)`: `NCHAR(n)`으로 매핑 (정밀도 포함, 최대 4000).
- `NUMBER` (정밀도 없음): `NUMERIC`으로 매핑.

## 2. MSSQL DDL 조건 확인
### 2.1. `CreateTableDDL`
- 다른 스키마에 동일한 이름의 테이블이 존재할 때 발생하는 오탐(false positive)을 방지하기 위해 `IF NOT EXISTS` 검사에 `TABLE_SCHEMA` 조건을 추가합니다.
- 스키마가 지정되지 않은 경우 기본값을 `dbo`로 설정합니다.

### 2.2. `CreateIndexDDL`
- `sys.objects`와 `sys.indexes`를 `object_id` 기준으로 조인하여 테이블 이름과 스키마로 엄격하게 필터링함으로써, 다른 테이블 간의 중복된 인덱스 이름 충돌을 방지합니다.

## 3. 웹 UI 개선
### 3.1. DDL 옵션 가시성
- DDL 관련 옵션(`--with-ddl`, `--with-sequences`, `--with-indexes`, `oracleOwner`)을 "Direct Migration" 토글 섹션에서 분리합니다.
- 이를 "고급 설정(Advanced Settings)" 아래의 공통 "DDL 설정" 섹션에 표시하여, 파일 출력(File Output)과 직접 마이그레이션(Direct Migration) 모드 모두에 적용 가능하고 눈에 띄게 만듭니다.

### 3.2. 레이블 및 제목 수정 (PostgreSQL 종속성 제거)
- `server.go` HTML 제목을 `"Oracle to PostgreSQL Migrator"`에서 `"Oracle DB Migrator"`로 업데이트합니다.
- `index.html`에서 Schema 입력 레이블을 `"PG Schema"`에서 `"Schema"`로 업데이트합니다.

## 4. WebSocket 경고 메시지 구현
- `ws/tracker.go`에 `MsgWarning MsgType = "warning"`을 추가합니다.
- `WebSocketTracker`에 `Warning(message string)` 메서드를 구현합니다.
- `ProgressTracker` 인터페이스를 `WarningTracker` 인터페이스로 확장합니다.
- 특정 방언(dialect)이 시퀀스 DDL을 지원하지 않을 때(예: MySQL) 마이그레이션 중에 경고 메시지를 발생시킵니다.
- 웹 UI(`index.html`)의 `handleProgressMessage`를 업데이트하여 `warning` 이벤트를 수신할 때 진행 컨테이너 상단에 경고 배너(노란색)를 표시합니다.

## 5. 예행 연습(Dry-Run) 대상 DB 연결 검증
- 예행 연습 모드(`cfg.DryRun == true`)에서 `--target-url`이 지정된 경우 대상 데이터베이스 연결을 시도합니다.
- `DryRunResult.ConnectionOk`를 통해 연결 성공/실패 결과를 전송합니다.
- 대상 DB 연결 상태를 웹 UI에 반영합니다.

## 6. 단위 테스트 추가
- `internal/dialect/`에 테스트 파일을 생성합니다:
  - `mysql_test.go`
  - `mariadb_test.go`
  - `sqlite_test.go`
  - `mssql_test.go`
- 테스트 케이스는 다음을 포괄해야 합니다:
  1. `TestMapOracleType_*`
  2. `TestCreateTableDDL_*`
  3. `TestCreateIndexDDL_*`
  4. `TestInsertStatement_*`
