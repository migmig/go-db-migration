# go-db-migration

Oracle에서 PostgreSQL로 데이터를 마이그레이션하는 Go 기반 CLI / Web UI 도구입니다.

이 프로젝트는 다음 시나리오를 지원합니다.

- Oracle → PostgreSQL 직접 마이그레이션
- Oracle 메타데이터 기반 DDL 생성
- SQL 파일 생성 방식의 오프라인 이관
- 대용량 테이블 배치 처리 및 병렬 처리
- 진행률 모니터링, 재시도, 재개(resume), 검증
- Web UI 기반 실행 및 이력 관리

> **v22+ 변경사항**
> 타겟 데이터베이스는 **PostgreSQL만 지원**합니다. 기존 MariaDB / MySQL / MSSQL / SQLite 타겟 지원은 제거되었습니다.

---

## 주요 기능

- **순수 Go 드라이버 사용**
  - Oracle Instant Client나 CGO 의존 없이 실행 가능
- **PostgreSQL 직접 이관**
  - `pgx` 기반 `COPY` 프로토콜 사용
- **Web UI 제공**
  - WebSocket 기반 실시간 진행률 모니터링
- **대용량 테이블 처리**
  - 배치 처리, 병렬 워커, 청크 기반 진행
- **자동 복구 / 재시도**
  - 일시적 오류 발생 시 재시도 가능
- **Resume 지원**
  - 중단된 작업을 Job ID 기준으로 재개
- **DDL 생성 지원**
  - 테이블, 인덱스, 시퀀스, 제약조건 생성 가능
- **사전 점검(Pre-check)**
  - 원본/타겟 row count 비교 후 전송 대상 선별 가능
- **검증(Validation)**
  - 마이그레이션 후 row count 비교 가능
- **인증 기반 멀티유저 Web UI**
  - 로그인, 세션, 사용자별 credential / history 관리
- **쉘 자동완성 지원**
  - Bash / Zsh / Fish / PowerShell

---

## 지원 대상

### Source
- Oracle

### Target
- PostgreSQL

---

## 프로젝트 구조

```text
.
├── main.go                  # CLI 진입점
├── internal/
│   ├── config/              # 플래그/설정 처리
│   ├── db/                  # DB 연결 및 메타데이터 조회
│   ├── dialect/             # SQL 방언 처리
│   ├── migration/           # 마이그레이션 핵심 로직
│   ├── security/            # 인증/암호화
│   └── web/                 # Web API / Web UI 서버
├── frontend/                # Vite + React 프론트엔드
├── docs/                    # 버전별 문서
└── scripts/                 # 빌드/보조 스크립트
```

---

## 설치

```bash
go build -o dbmigrator main.go
```

### 크로스 컴파일 예시

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o dbmigrator-linux main.go

# Windows
GOOS=windows GOARCH=amd64 go build -o dbmigrator.exe main.go

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o dbmigrator-mac main.go
```

---

## 빠른 시작

### 1) Web UI 실행

```bash
./dbmigrator -web
```

기본 주소:
- `http://localhost:8080`

---

### 2) Oracle → PostgreSQL 직접 이관

```bash
./dbmigrator \
  -url "localhost:1521/ORCL" \
  -user "scott" \
  -password "tiger" \
  -tables "USERS" \
  -pg-url "postgres://pguser:pgpass@localhost:5432/mydb"
```

---

### 3) DDL 포함 이관

```bash
./dbmigrator \
  -url "localhost:1521/ORCL" \
  -user "scott" \
  -password "tiger" \
  -tables "USERS" \
  -pg-url "postgres://pguser:pgpass@localhost:5432/mydb" \
  -schema "myschema" \
  -with-ddl
```

---

### 4) SQL 파일 생성

```bash
./dbmigrator \
  -url "localhost:1521/ORCL" \
  -user "scott" \
  -password "tiger" \
  -tables "USERS,ORDERS" \
  -out "export.sql"
```

---

### 5) Dry-run

```bash
./dbmigrator \
  -url "localhost:1521/ORCL" \
  -user "scott" \
  -password "tiger" \
  -tables "USERS,ORDERS" \
  -dry-run
```

---

## 주요 사용 예시

### 인덱스 / 시퀀스 / 제약조건 포함

```bash
./dbmigrator \
  -url "localhost:1521/ORCL" \
  -user "scott" \
  -password "tiger" \
  -tables "USERS" \
  -with-ddl \
  -with-sequences \
  -with-indexes \
  -with-constraints
```

### Oracle owner 지정 및 시퀀스 명시

```bash
./dbmigrator \
  -url "localhost:1521/ORCL" \
  -user "scott" \
  -password "tiger" \
  -tables "USERS" \
  -with-ddl \
  -with-sequences \
  -oracle-owner "HR" \
  -sequences "SEQ_USERS,SEQ_ORDERS"
```

### 테이블별 개별 SQL 파일 생성

```bash
./dbmigrator \
  -url "localhost:1521/ORCL" \
  -user "scott" \
  -password "tiger" \
  -tables "USERS,ORDERS" \
  -per-table \
  -out "export.sql"
```

### Resume

```bash
./dbmigrator \
  -url "localhost:1521/ORCL" \
  -user "scott" \
  -password "tiger" \
  -resume "20260313150405"
```

---

## Web UI

Web UI에서는 다음 기능을 사용할 수 있습니다.

- Oracle 테이블 목록 조회
- 타겟 PostgreSQL 연결 테스트
- 마이그레이션 시작 / 재시도
- 실시간 진행률 확인
- 결과 ZIP / 리포트 다운로드
- row count pre-check
- 사용자별 credential / history 관리 (인증 모드)

### 최신 프런트엔드 실행 흐름

```bash
cd frontend
npm install
npm run build

cd ..
./dbmigrator -web
```

### 오프라인 단일 바이너리 빌드

```bash
make offline
```

이 타깃은 프런트 자산을 바이너리에 임베드합니다.

---

## 인증 모드

인증 기반 멀티유저 모드를 사용하려면:

- `-auth-enabled`
- `DBM_MASTER_KEY`

가 필요합니다.

```bash
export DBM_MASTER_KEY="change-me-32-bytes-or-more"
./dbmigrator -web -auth-enabled
```

### 사용자 관리 CLI

```bash
./dbmigrator users list
./dbmigrator users add <username> <password>
./dbmigrator users add <username> <password> --admin
./dbmigrator users reset-password <username> <new_password>
./dbmigrator users delete <username>
```

기본 사용자 DB 경로:
- `.migration_state/auth.db`

필요 시:

```bash
export DBM_AUTH_DB_PATH=./my-auth.db
```

---

## Object Group 모드

`-object-group`으로 실행 대상을 나눌 수 있습니다.

- `all`: 기본값, 테이블 + 시퀀스
- `tables`: 테이블/데이터만 실행
- `sequences`: 시퀀스 DDL만 실행

예시:

```bash
./dbmigrator -url "localhost:1521/ORCL" -user "scott" -password "tiger" -tables "USERS" -with-ddl -object-group all
./dbmigrator -url "localhost:1521/ORCL" -user "scott" -password "tiger" -tables "USERS" -with-ddl -object-group tables
./dbmigrator -url "localhost:1521/ORCL" -user "scott" -password "tiger" -tables "USERS" -with-ddl -object-group sequences
```

---

## Pre-check Row Count

마이그레이션 전에 원본/타겟 row count를 비교할 수 있습니다.

```bash
./dbmigrator \
  -url "localhost:1521/ORCL" \
  -user "scott" \
  -password "tiger" \
  -tables "USERS,ORDERS" \
  -pg-url "postgres://pguser:pgpass@localhost:5432/mydb" \
  -precheck-row-count \
  -precheck-policy skip_equal_rows
```

정책:
- `strict`
- `best_effort`
- `skip_equal_rows`

---

## 주요 플래그

| Flag | 설명 | 기본값 |
| --- | --- | --- |
| `-web` | Web UI 모드 실행 | `false` |
| `-url` | Oracle URL | 없음 |
| `-user` | Oracle 사용자명 | 없음 |
| `-password` | Oracle 비밀번호 | 없음 |
| `-tables` | 대상 테이블 목록 | 없음 |
| `-target-db` | 타겟 DB (`postgres` only) | `postgres` |
| `-target-url` | PostgreSQL 연결 URL | 없음 |
| `-pg-url` | PostgreSQL 연결 URL (legacy) | 없음 |
| `-schema` | PostgreSQL 스키마 | 없음 |
| `-with-ddl` | DDL 생성/실행 포함 | `false` |
| `-with-sequences` | 시퀀스 포함 | `false` |
| `-with-indexes` | 인덱스 포함 | `false` |
| `-with-constraints` | 제약조건 포함 | `false` |
| `-parallel` | 병렬 처리 | `false` |
| `-workers` | 워커 수 | `4` |
| `-batch` | INSERT 배치 크기 | `1000` |
| `-copy-batch` | PostgreSQL COPY 배치 크기 | `10000` |
| `-dry-run` | 실제 이관 없이 점검만 수행 | `false` |
| `-resume` | Job ID 기준 재개 | 없음 |
| `-validate` | 이관 후 검증 수행 | `false` |
| `-truncate` | 이관 전 타겟 TRUNCATE | `false` |
| `-upsert` | PK 기준 upsert | `false` |
| `-precheck-row-count` | 사전 row count 비교 | `false` |
| `-precheck-policy` | pre-check 정책 | `strict` |
| `-auth-enabled` | 인증 모드 활성화 | `false` |
| `-object-group` | 실행 그룹 선택 | `all` |
| `-on-error` | 배치 오류 처리 정책 | `fail_fast` |
| `-completion` | 쉘 자동완성 출력 | 없음 |

---

## 환경 변수

| 변수 | 설명 |
| --- | --- |
| `DBM_MASTER_KEY` | 인증/credential 암호화 마스터 키 |
| `DBM_AUTH_DB_PATH` | 인증 DB 경로 |
| `DBM_MAX_SESSIONS` | 최대 세션 수 |
| `DBM_SESSION_CLEANUP_INTERVAL` | 세션 정리 주기 |
| `DBM_V19_PRECHECK` | pre-check 기능 on/off |
| `DBM_MAX_RETRIES` | 재시도 횟수 override |
| `DBM_RETRY_INITIAL_WAIT` | 초기 재시도 대기시간 override |

---

## 개발

### 테스트

```bash
go test -v ./...
```

### 프런트엔드 검증

```bash
cd frontend
npm run test
npm run typecheck
npm run build
```

---

## 주의사항

- Oracle source / PostgreSQL target 구조 차이를 반드시 확인하세요.
- 대용량 이관 전에는 `-dry-run` 또는 `-precheck-row-count`를 먼저 권장합니다.
- 운영 반영 전에는 테스트 환경에서 DDL / 제약조건 / 시퀀스 동작을 검증하세요.
- 인증 모드 사용 시 `DBM_MASTER_KEY`를 안전하게 관리하세요.

---

## 라이선스

라이선스 정보가 별도로 명시되어 있지 않다면 저장소 정책을 확인하세요.
