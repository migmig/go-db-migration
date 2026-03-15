# Oracle to Multi-Target Data Migration CLI (v15)

Oracle 데이터베이스에서 다양한 대상 데이터베이스(PostgreSQL, MySQL, MariaDB, SQLite, MSSQL)로 데이터를 마이그레이션하기 위해 설계된 고성능 Go 기반 CLI 애플리케이션입니다. 실시간 모니터링, 자동 복구(Auto-healing), 대용량 테이블 청크(Chunking) 처리가 가능한 고급 웹 UI를 제공합니다.

## 주요 기능 (Features)

- **순수 Go 드라이버 (Pure Go Drivers):** Oracle Instant Client나 CGO 설치가 필요하지 않습니다.
- **다중 대상 데이터베이스 지원:** PostgreSQL, MySQL, MariaDB, SQLite, MSSQL로 직접 데이터를 마이그레이션 할 수 있습니다.
- **고급 웹 UI (Advanced Web UI) (v11):** WebSocket 기반의 실시간 진행률 추적, 대시보드 모니터링 및 스키마 토폴로지를 제공하는 대화형 웹 인터페이스입니다.
- **테이블 청크 분할 (Table Chunking) (v11):** 대용량 테이블을 자동으로 분할하여 테이블 내 병렬 처리(Intra-table parallel migration)를 지원함으로써 처리 속도를 극대화합니다.
- **자동 복구 (Auto-Healing) (v11):** 네트워크 타임아웃이나 일시적 오류 발생 시 스마트 자동 재시도 메커니즘을 통해 안정적인 마이그레이션을 보장합니다.
- **직접 마이그레이션 (Direct Migration):** Oracle에서 대상 DB로 데이터를 직접 스트리밍합니다. PostgreSQL의 경우 고성능 `COPY` 프로토콜을 사용합니다.
- **대량 SQL 생성 (Bulk SQL Generation):** 직접 연결 대신 타겟 호환용 대량 `INSERT` SQL 스크립트를 파일로 생성할 수 있습니다.
- **DDL 자동 생성 (DDL Generation):** Oracle 메타데이터를 기반으로 대상 DB에 맞는 `CREATE TABLE` 문, 인덱스, 시퀀스, 제약조건 등을 자동으로 생성하고 실행합니다.
- **워커 풀 병렬 처리 (Worker Pool Parallelism):** 설정 가능한 워커 풀을 통해 여러 테이블을 효율적으로 동시 처리합니다.
- **데이터 검증 (Validation):** 마이그레이션 완료 후 소스 DB와 대상 DB 간의 행(Row) 수를 비교하여 검증합니다.
- **예행 연습 모드 (Dry Run Mode):** 실제 마이그레이션을 수행하지 않고 데이터베이스 연결을 확인하고 예상 데이터 볼륨을 산출합니다.
- **마이그레이션 재개 (Resume) 및 히스토리 (v11):** 로컬 SQLite에 마이그레이션 이력과 로그를 저장하여 중단된 작업을 쉽게 재개(Resume)하고 감사를 수행할 수 있습니다.
- **구조화된 로깅 (Structured Logging):** `log/slog`를 활용한 JSON 또는 Text 기반의 구조화된 로깅을 지원합니다.
- **데이터 타입 매핑 (Data Type Mapping):** VARCHAR2, CLOB, BLOB, RAW, DATE, TIMESTAMP, NUMBER(정밀도 포함) 등 복잡한 타입을 안전하게 매핑합니다.
- **쉘 자동완성 (Shell Completion) (v12, v13):** `-completion` 플래그로 Bash/Zsh/Fish/PowerShell 자동완성 스크립트를 생성할 수 있습니다. 단독으로 입력 시 현재 쉘을 자동 감지합니다.
- **Web UI 입력 자동완성/기억 (v14):** 최근 입력값 자동완성과 상단 공통 DB URL/ID/PASS(비밀번호 기억 옵트인) 복원을 지원하여 재접속 후에도 빠르게 작업을 이어갈 수 있습니다.
- **인증 기반 멀티유저 준비 (v15):** `-auth-enabled` 플래그와 `DBM_MASTER_KEY` 환경변수 기반 설정을 추가하여 인증/암호화 기능 도입을 위한 실행 옵션을 제공합니다.

## 설치 (Installation)

```bash
go build -o dbmigrator main.go
```

### 크로스 컴파일 빌드 (Cross-Platform Build)

Go의 크로스 컴파일 기능을 사용하여 각 OS용 실행 파일을 쉽게 빌드할 수 있습니다:

**Linux:**
```bash
GOOS=linux GOARCH=amd64 go build -o dbmigrator-linux main.go
```

**Windows:**
```bash
GOOS=windows GOARCH=amd64 go build -o dbmigrator.exe main.go
```

**macOS (Apple Silicon):**
```bash
GOOS=darwin GOARCH=arm64 go build -o dbmigrator-mac main.go
```

## 사용법 (Usage)

### Web UI 모드

브라우저 기반의 인터페이스를 사용하려면 웹 모드로 실행하세요:

```bash
./dbmigrator -web
```
- 기본 접속 URL: `http://localhost:8080`
- 기능: 테이블 검색(LIKE), 실시간 마이그레이션 진행 상황 추적, 생성된 SQL 파일 ZIP 다운로드 등.
- v14 추가: 상단 Quick Shared Connection에서 DB URL/ID/PASS를 공통 관리할 수 있으며, `비밀번호 기억` 체크 시 PASS까지 재접속 후 복원됩니다(공용 PC 비권장).

![Web UI Screenshot](docs/web-ui.png)

### 인증 모드 및 키 설정 (v15 준비)

v15 멀티유저 인증 기능 도입을 위한 실행 옵션이 추가되었습니다.

- `-auth-enabled`: 인증 기반 멀티유저 모드 활성화 플래그 (기본값 `false`)
- `DBM_MASTER_KEY`: DB 접속정보 암호화에 사용할 마스터 키 환경변수

예시:
```bash
export DBM_MASTER_KEY="change-me-32-bytes-or-more"
./dbmigrator -web -auth-enabled
```

> 참고: 현재 릴리즈에서는 인증 상세 기능이 단계적으로 구현 중이며, 위 옵션은 v15 구현 경로를 위한 준비 설정입니다.

### 관리자 CLI (v15 예정)

v15 스펙 기준으로 아래 계정 관리 커맨드를 사용할 예정입니다.

```bash
./dbmigrator users list
./dbmigrator users add <username> <password>
./dbmigrator users reset-password <username> <new_password>
./dbmigrator users delete <username>
```

### 쉘 자동완성 스크립트 생성 (v12, v13)

`-completion` 플래그를 단독으로 사용하면 현재 사용 중인 쉘(Bash, Zsh, Fish 등)을 자동으로 감지하여 적절한 스크립트를 출력합니다.

**현재 쉘에 즉시 적용 (권장):**
```bash
eval "$(./dbmigrator -completion)"
```

수동으로 쉘을 지정하여 파일로 저장할 수도 있습니다:

**Bash:**
```bash
./dbmigrator -completion bash > /etc/bash_completion.d/dbmigrator
source /etc/bash_completion.d/dbmigrator
```

**Zsh:**
```bash
./dbmigrator -completion zsh > ~/.zsh/completions/_dbmigrator
autoload -U compinit && compinit
```

**Fish:**
```bash
./dbmigrator -completion fish > ~/.config/fish/completions/dbmigrator.fish
```

**PowerShell:**
```powershell
./dbmigrator -completion powershell | Out-String | Invoke-Expression
```

### 직접 마이그레이션 (Direct Migration)

**PostgreSQL로 직접 이관:**
```bash
./dbmigrator -url "localhost:1521/ORCL" \
             -user "scott" \
             -password "tiger" \
             -tables "USERS" \
             -pg-url "postgres://pguser:pgpass@localhost:5432/mydb"
```

**스키마 지정 및 DDL 포함 (PostgreSQL):**
```bash
./dbmigrator -url "localhost:1521/ORCL" \
             -user "scott" \
             -password "tiger" \
             -tables "USERS" \
             -pg-url "postgres://pguser:pgpass@localhost:5432/mydb" \
             -schema "myschema" \
             -with-ddl
```

**Sequence 및 Index DDL 포함:**
```bash
./dbmigrator -url "localhost:1521/ORCL" \
             -user "scott" \
             -password "tiger" \
             -tables "USERS" \
             -with-ddl -with-sequences -with-indexes
```

**제약조건(Default, FK, Check) 포함:**
```bash
./dbmigrator -url "localhost:1521/ORCL" \
             -user "scott" \
             -password "tiger" \
             -tables "USERS" \
             -with-ddl -with-constraints
```

**소유자 명시 및 개별 Sequence 지정:**
```bash
./dbmigrator -url "localhost:1521/ORCL" \
             -user "scott" \
             -password "tiger" \
             -tables "USERS" \
             -with-ddl -with-sequences \
             -oracle-owner "HR" -sequences "SEQ_USERS,SEQ_ORDERS"
```

### 파일 기반 마이그레이션 (SQL File Generation)

**SQL 단일 파일 생성:**
```bash
./dbmigrator -url "localhost:1521/ORCL" \
             -user "scott" \
             -password "tiger" \
             -tables "USERS,ORDERS" \
             -out "export.sql" \
             -batch 1000
```

**테이블별 개별 SQL 파일 생성:**
```bash
./dbmigrator -url "localhost:1521/ORCL" \
             -user "scott" \
             -password "tiger" \
             -tables "USERS,ORDERS" \
             -per-table -out "export.sql"
```

### 마이그레이션 재개 (Resume)

실패하거나 중단된 마이그레이션 작업을 Job ID를 사용하여 이어서 실행합니다:
```bash
./dbmigrator -url "localhost:1521/ORCL" \
             -user "scott" \
             -password "tiger" \
             -resume "20260313150405"
```

### 예행 연습 (Dry Run)

실제 데이터를 이관하지 않고 연결 테스트 및 처리 예상 행(Row) 수만 확인합니다:
```bash
./dbmigrator -url "localhost:1521/ORCL" -user "scott" -password "tiger" -tables "USERS,ORDERS" -dry-run
```

## 플래그 (Flags)

| 플래그 | 설명 | 기본값 | 필수 여부 |
| --- | --- | --- | --- |
| `-web` | Web UI 모드로 실행 | `false` | 아니오 |
| `-url` | Oracle 데이터베이스 URL (예: host:port/service_name) | 없음 | 예* |
| `-user` | Oracle 데이터베이스 사용자명 | 없음 | 예* |
| `-password` | Oracle 데이터베이스 비밀번호 | 없음 | 예* |
| `-tables` | 마이그레이션할 테이블 목록 (쉼표로 구분) | 없음 | 예* |
| `-target-db` | 출력 대상 DB 종류 (`postgres`, `mysql`, `mariadb`, `sqlite`, `mssql`) | `postgres` | 아니오 |
| `-target-url` | 대상 DB 연결 URL (PostgreSQL 외 Direct 마이그레이션 시) | 없음 | 아니오 |
| `-pg-url` | PostgreSQL 연결 URL (Legacy) | 없음 | 아니오 |
| `-out` | 출력 SQL 파일명 | `migration.sql` | 아니오 |
| `-batch` | INSERT 배치당 행(Row) 수 | `1000` | 아니오 |
| `-schema` | PostgreSQL 스키마 이름 (선택) | 없음 | 아니오 |
| `-per-table` | 테이블별 별도 SQL 파일로 출력 | `false` | 아니오 |
| `-parallel` | 테이블 병렬 처리 | `false` | 아니오 |
| `-workers` | 병렬 처리 워커 수 | `4` | 아니오 |
| `-with-ddl` | CREATE TABLE DDL 생성/실행 포함 | `false` | 아니오 |
| `-with-sequences` | 연관 Sequence DDL 포함 | `false` | 아니오 |
| `-with-indexes` | 연관 Index DDL 포함 | `false` | 아니오 |
| `-with-constraints` | 제약조건(Default, FK, Check) 마이그레이션 포함 | `false` | 아니오 |
| `-sequences` | 추가 포함할 Sequence 이름 목록 (쉼표 구분) | 없음 | 아니오 |
| `-oracle-owner` | Oracle 스키마 소유자 (미지정 시 `-user` 값 사용) | 없음 | 아니오 |
| `-db-max-open` | DB 커넥션 풀 최대 활성 연결 수 (0: 무제한) | `0` | 아니오 |
| `-db-max-idle` | DB 커넥션 풀 최대 유휴 연결 수 | `2` | 아니오 |
| `-db-max-life` | DB 커넥션 풀 최대 유지 시간(초) (0: 무제한) | `0` | 아니오 |
| `-validate` | 마이그레이션 후 소스-타겟 행 수 검증 수행 | `false` | 아니오 |
| `-copy-batch` | PostgreSQL COPY 배치 크기 (0: 단일 COPY 모드) | `10000` | 아니오 |
| `-resume` | 재개할 Job ID | 없음 | 아니오 |
| `-completion` | 쉘 자동완성 스크립트 출력 (`bash`, `zsh`, `fish`, `powershell`) | 없음 | 아니오 |
| `-auth-enabled` | 인증 기반 멀티유저 모드 활성화 (v15 준비) | `false` | 아니오 |
| `-dry-run` | 연결 확인 및 예상 행 수만 조회 (실제 이관 없음) | `false` | 아니오 |
| `-log-json` | JSON 구조화 로그 활성화 | `false` | 아니오 |

*\* CLI 모드에서는 필수 항목입니다. (Web 모드 시 UI에서 입력)*

## 환경 변수 (Environment Variables)

| 변수 | 설명 | 필수 여부 |
| --- | --- | --- |
| `DBM_MASTER_KEY` | v15 인증/접속정보 암호화 기능에서 사용할 마스터 키. 운영 환경에서는 반드시 강한 비밀값을 사용하세요. | v15 인증모드 사용 시 권장 |

## 개발 (Development)

테스트 코드 실행:
```bash
go test -v ./...
```
