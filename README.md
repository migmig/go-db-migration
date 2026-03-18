# Oracle to Multi-Target Data Migration CLI (v20)

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
- **인증 기반 멀티유저 (v15):** `-auth-enabled` 플래그로 Web UI의 로그인/로그아웃, 세션 기반 접근 제어, 사용자별 접속정보 저장, 사용자별 작업 이력 조회를 활성화할 수 있습니다.
- **객체 그룹 선택 실행 (v17):** `-object-group` 플래그로 `all|tables|sequences` 실행 그룹을 선택할 수 있습니다. `sequences` 모드는 시퀀스 DDL 전용 경로를 사용합니다.
- **테이블 이력 기반 목록 UX (v18):** v16 UI의 테이블 선택 화면에서 이력 상태 필터(미실행/성공/실패), `성공 제외` 토글, 정렬 컨트롤을 제공하며 상태/이력 뱃지에 텍스트+색상을 함께 적용해 접근성을 높였습니다.
- **테이블 상세 이력/재시도 UX (v18):** 목록에서 테이블별 최근 이력을 바로 열어 실패 요약을 확인하고, 실패 항목의 설정을 즉시 재적용(`Retry settings`)해 재시도 준비를 단축할 수 있으며, 빈 상태/오류 상태/로딩 스켈레톤을 제공합니다.
- **사전 행 수 비교 (Pre-check Row Count) (v19):** 마이그레이션 전 Oracle 원본과 대상 DB의 테이블별 행 수를 병렬로 비교하여 전송 필요 여부를 자동 판정합니다. `transfer_required`, `skip_candidate`, `count_check_failed` 세 가지 decision과 `strict`, `best_effort`, `skip_equal_rows` 세 가지 policy를 지원합니다. Web UI의 Migration Options에서 "Run Pre-check" 버튼으로 즉시 사용할 수 있으며, CLI에서는 `-precheck-row-count` 플래그로 활성화합니다.
- **CLI 수치형 입력 검증 (v20):** `-batch`, `-workers`, `-db-max-open`, `-db-max-idle` 값에 범위 검증이 적용되어 잘못된 파라미터를 실행 전에 즉시 차단합니다.

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
- v18 추가: 인증 모드에서 `/api/history` 기반으로 테이블별 이력 상태/실행 횟수를 표시하고, 성공 이력 제외 필터로 미완료 대상 위주로 선택할 수 있습니다.
- v14 추가: 상단 Quick Shared Connection에서 DB URL/ID/PASS를 공통 관리할 수 있으며, `비밀번호 기억` 체크 시 PASS까지 재접속 후 복원됩니다(공용 PC 비권장).


### 인증 모드 및 키 설정 (v15)

v15 멀티유저 인증 기능을 활성화할 수 있습니다.

- `-auth-enabled`: 인증 기반 멀티유저 모드 활성화 플래그 (기본값 `false`)
- `DBM_MASTER_KEY`: DB 접속정보 암호화에 사용할 마스터 키 환경변수 (`-auth-enabled` 사용 시 필요)

예시:
```bash
export DBM_MASTER_KEY="change-me-32-bytes-or-more"
./dbmigrator -web -auth-enabled
```

> 참고: `-auth-enabled` 사용 시 `/api/tables`, `/api/migrate`, `/api/credentials`, `/api/history` 등 주요 API는 로그인 세션이 있어야 접근할 수 있습니다.
> 인증 세션은 `SameSite=Lax`, `HttpOnly` 쿠키를 사용하며 idle timeout 30분, absolute timeout 24시간 정책을 따릅니다. HTTPS 환경에서는 `Secure` 쿠키가 적용됩니다.
> 운영 모니터링은 로그인 후 `GET /api/monitoring/metrics`에서 확인할 수 있으며, 로그인 실패율/세션 만료율, `credentials`/`history` API 오류율, `all|tables|sequences` 모드별 실행 수/실패율/재시도 성공률을 제공합니다.

### v18 프런트 미리보기 (Vite + React + Tailwind)

v18 UI는 별도 프런트 빌드 산출물을 서버가 정적으로 제공하는 방식입니다.

```bash
cd frontend
npm install
npm run build

cd ..
export DBM_MASTER_KEY="replace-with-16-24-or-32-byte-key"
./dbmigrator -web -auth-enabled
```

- 기본 접속 URL: `http://localhost:8080/` -> 자동으로 `http://localhost:8080/app`로 이동합니다.
- 구 화면(legacy): `http://localhost:8080/legacy`
- 런타임에는 Node/Vite dev server가 필요하지 않습니다(빌드 결과물만 사용).
- `go build`만 수행한 바이너리에는 placeholder 프런트 페이지가 포함되고, `make offline`으로 빌드한 바이너리에는 실제 프런트 번들이 포함됩니다.

단일 오프라인 바이너리로 묶으려면 아래처럼 한 번에 빌드할 수 있습니다.

```bash
make offline
```

- 이 타깃은 `frontend` 검증/빌드 후, 프런트 자산을 Go 바이너리에 임베드해서 `./dbmigrator` 하나만 생성합니다.
- 생성된 바이너리는 런타임에 `frontend/dist`, Node, npm, 네트워크 연결이 필요하지 않습니다.
- 다른 출력 파일명을 쓰려면 `make offline OUTPUT=./build/dbmigrator` 형식으로 실행하면 됩니다.

프런트 개발 시 권장 체크 명령:

```bash
cd frontend
npm run test
npm run typecheck
npm run build
```

`tsgo`를 설치해두면 더 빠른 타입체크 경로를 사용할 수 있습니다(선택):

```bash
go install github.com/microsoft/typescript-go/cmd/tsgo@latest
cd frontend
npm run typecheck:fast
npm run verify:fast
```

- `typecheck:fast`는 `tsgo`가 있으면 `tsgo -b`, 없으면 자동으로 `tsc -b`를 사용합니다.

### 관리자 CLI (v15)

아래 계정 관리 커맨드를 통해 로컬 인증 사용자 계정을 관리할 수 있습니다.

```bash
./dbmigrator users list
./dbmigrator users add <username> <password>
./dbmigrator users reset-password <username> <new_password>
./dbmigrator users delete <username>
```

기본적으로 사용자 정보는 `.migration_state/auth.db`에 저장됩니다.
필요하면 `DBM_AUTH_DB_PATH` 환경변수로 경로를 변경할 수 있습니다.

```bash
export DBM_AUTH_DB_PATH=./my-auth.db
./dbmigrator users list
```

### 배포 전 체크리스트 (v15)

인증 모드 배포 전에는 아래 항목을 확인하세요.

1. `DBM_MASTER_KEY`가 운영 환경에 주입되어 있는지 확인합니다.
2. 인증 DB 경로(`DBM_AUTH_DB_PATH`)가 운영 서버에서 지속 저장되는 위치인지 확인합니다.
3. 초기 관리자 계정을 생성합니다.
4. 초기 비밀번호로 로그인 확인 후 즉시 비밀번호를 변경합니다.

예시:
```bash
export DBM_MASTER_KEY="replace-with-strong-32-byte-secret"
export DBM_AUTH_DB_PATH=/var/lib/dbmigrator/auth.db

./dbmigrator users add admin temporary123 --admin
./dbmigrator -web -auth-enabled

# 로그인 확인 후 즉시 변경
./dbmigrator users reset-password admin stronger-password-123
```

권장사항:
- `DBM_MASTER_KEY`는 소스 코드나 셸 이력에 남기지 말고 비밀 저장소나 배포 환경 변수로 관리합니다.
- 초기 관리자 계정은 작업 완료 후 `users list`로 생성 여부를 점검합니다.
- 공용 또는 테스트용 임시 비밀번호는 운영 반영 전에 반드시 교체합니다.

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

### 객체 그룹 실행 모드 (v17)

`-object-group`으로 실행 대상을 분리할 수 있습니다.

- `all`: 테이블/데이터를 먼저 처리한 뒤 시퀀스를 후속 단계로 실행합니다.
- `tables`: 테이블 계열만 실행하며 sequence DDL은 자동 비활성화됩니다.
- `sequences`: sequence DDL만 생성/실행하며 `-with-ddl`, `-with-sequences`가 자동 활성화됩니다.
- SQL 파일 모드에서 `all|tables`는 tables 계열 묶음을 `tables.sql`로 보관하고, `all|sequences`는 시퀀스 산출물을 `sequences.sql`로 분리합니다.
- 완료 리포트(`Download Report`)와 WebSocket 완료 요약에는 `stats.tables`, `stats.sequences` 그룹별 통계가 포함됩니다.
- 운영 중 단계적 오픈이 필요하면 `DBM_OBJECT_GROUP_UI_ENABLED=false`로 v16 UI의 객체 그룹 선택기를 숨기고 legacy `all` 모드로 고정할 수 있습니다.
- 운영 절차와 복구 플로우는 [docs/v17/rollout.md](/Users/migmig/gowork/go-db-migration/docs/v17/rollout.md)에서 관리합니다.

```bash
# 기본값: all
./dbmigrator -url "localhost:1521/ORCL" -user "scott" -password "tiger" -tables "USERS" -with-ddl -object-group all

# tables 전용: 테이블/데이터 중심 경로 (sequence DDL 비활성)
./dbmigrator -url "localhost:1521/ORCL" -user "scott" -password "tiger" -tables "USERS" -with-ddl -with-sequences -object-group tables

# sequences 전용: 시퀀스 DDL만 생성/실행
./dbmigrator -url "localhost:1521/ORCL" -user "scott" -password "tiger" -tables "USERS" -with-ddl -object-group sequences
```

예시: Dry-run / 완료 리포트의 그룹별 요약

```text
TABLES SQL
count: 12
estimated_tables: 12
estimated_rows: 184320
SEQUENCES SQL
count: 3

Run Status
Session: ... · WS closed · Target all
Tables Group: 12 ok · 0 error · 184,320 rows
Sequences Group: 3 ok · 0 error · 3 objects
```

```json
{
  "report_id": "job_20260317083000",
  "object_group": "all",
  "stats": {
    "tables": {
      "total_items": 12,
      "success_count": 12,
      "error_count": 0,
      "skipped_count": 0,
      "total_rows": 184320
    },
    "sequences": {
      "total_items": 3,
      "success_count": 3,
      "error_count": 0,
      "skipped_count": 0
    }
  }
}
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
| `-batch` | INSERT 배치당 행(Row) 수 (1~100000) | `1000` | 아니오 |
| `-schema` | PostgreSQL 스키마 이름 (선택) | 없음 | 아니오 |
| `-per-table` | 테이블별 별도 SQL 파일로 출력 | `false` | 아니오 |
| `-parallel` | 테이블 병렬 처리 | `false` | 아니오 |
| `-workers` | 병렬 처리 워커 수 (1~64) | `4` | 아니오 |
| `-with-ddl` | CREATE TABLE DDL 생성/실행 포함 | `false` | 아니오 |
| `-with-sequences` | 연관 Sequence DDL 포함 | `false` | 아니오 |
| `-with-indexes` | 연관 Index DDL 포함 | `false` | 아니오 |
| `-with-constraints` | 제약조건(Default, FK, Check) 마이그레이션 포함 | `false` | 아니오 |
| `-sequences` | 추가 포함할 Sequence 이름 목록 (쉼표 구분) | 없음 | 아니오 |
| `-oracle-owner` | Oracle 스키마 소유자 (미지정 시 `-user` 값 사용) | 없음 | 아니오 |
| `-db-max-open` | DB 커넥션 풀 최대 활성 연결 수 (1~1000) | `10` | 아니오 |
| `-db-max-idle` | DB 커넥션 풀 최대 유휴 연결 수 (0~1000) | `2` | 아니오 |
| `-db-max-life` | DB 커넥션 풀 최대 유지 시간(초) (0: 무제한) | `0` | 아니오 |
| `-validate` | 마이그레이션 후 소스-타겟 행 수 검증 수행 | `false` | 아니오 |
| `-copy-batch` | PostgreSQL COPY 배치 크기 (0: 단일 COPY 모드) | `10000` | 아니오 |
| `-resume` | 재개할 Job ID | 없음 | 아니오 |
| `-completion` | 쉘 자동완성 스크립트 출력 (`bash`, `zsh`, `fish`, `powershell`) | 없음 | 아니오 |
| `-auth-enabled` | 인증 기반 멀티유저 모드 활성화 (로그인/세션 접근제어) | `false` | 아니오 |
| `-object-group` | 마이그레이션 객체 그룹 선택 (`all`, `tables`, `sequences`) | `all` | 아니오 |
| `-dry-run` | 연결 확인 및 예상 행 수만 조회 (실제 이관 없음) | `false` | 아니오 |
| `-log-json` | JSON 구조화 로그 활성화 | `false` | 아니오 |
| `-truncate` | 마이그레이션 전 대상 테이블 TRUNCATE | `false` | 아니오 |
| `-upsert` | PK 기준 Upsert (중복 행 건너뜀, PK 필수) | `false` | 아니오 |
| `-precheck-row-count` | 마이그레이션 전 원본/대상 행 수 사전 점검 수행 (v19) | `false` | 아니오 |
| `-precheck-policy` | pre-check 정책 (`strict`\|`best_effort`\|`skip_equal_rows`) (v19) | `strict` | 아니오 |
| `-precheck-filter` | pre-check 결과 출력 필터 (`all`\|`transfer_required`\|`skip_candidate`\|`count_check_failed`) (v19) | `all` | 아니오 |

*\* CLI 모드에서는 필수 항목입니다. (Web 모드 시 UI에서 입력)*

## 환경 변수 (Environment Variables)

| 변수 | 설명 | 필수 여부 |
| --- | --- | --- |
| `DBM_MASTER_KEY` | v15 인증/접속정보 암호화 기능에서 사용할 마스터 키. 운영 환경에서는 반드시 강한 비밀값을 사용하세요. | `-auth-enabled` 사용 시 필요 |
| `DBM_V19_PRECHECK` | `false`로 설정 시 v19 pre-check 기능 비활성화 (기본값: `true`) | 아니오 |

## 개발 (Development)

테스트 코드 실행:
```bash
go test -v ./...
```
