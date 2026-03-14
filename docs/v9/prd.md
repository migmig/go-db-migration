# PRD (Product Requirements Document) - v9 안정성·관측성·데이터 무결성 강화

## 1. 개요 (Overview)

v8까지 엔터프라이즈 스케일업(멀티 세션, 제약조건, 커넥션 풀, 체크포인트)이 완료되고, ui-improve 브랜치에서 2컬럼 레이아웃·다크모드·접근성이 도입되었습니다.

이번 **v9 릴리스**는 실 운영 환경에서 드러나는 **데이터 무결성 검증, 마이그레이션 관측성(Observability), 에러 복구 개선, 그리고 코드 구조의 유지보수성**에 집중하는 **운영 안정화 릴리스**입니다.

---

## 2. 배경 및 문제 분석 (Background & Issues Found)

### 2.1. 데이터 무결성 검증 수단 부재

- **현상**: 마이그레이션 완료 후 소스(Oracle)와 타겟 DB 간의 데이터 정합성을 확인할 방법이 없습니다.
- **문제**: `Done` 상태가 되어도 실제 행 수 불일치, 컬럼 값 누락, 인코딩 불일치 등을 감지할 수 없어, DBA가 수동으로 `SELECT COUNT(*)` 비교나 샘플링 검증을 해야 합니다. 대규모 환경일수록 이 수동 검증 비용이 기하급수적으로 증가합니다.

### 2.2. 마이그레이션 감사 로그(Audit Trail) 부재

- **현상**: 마이그레이션 수행 이력(시작 시각, 종료 시각, 테이블별 행 수, 오류 내역, 수행자)이 체계적으로 기록되지 않습니다.
- **문제**: 운영 환경에서 "언제, 누가, 어떤 테이블을, 얼마나 이관했는가"를 사후 추적할 수 없습니다. 장애 발생 시 원인 분석이 어렵고, 컴플라이언스 요건(감사 추적)을 충족하지 못합니다.

### 2.3. 에러 발생 시 컨텍스트 손실

- **현상**: 에러 메시지가 `fmt.Errorf("failed to execute batch insert: %v", err)` 형태로 전파되어, 어느 배치의 몇 번째 행에서 어떤 컬럼 값으로 인해 실패했는지 알 수 없습니다.
- **문제**: 타입 불일치(예: Oracle `NUMBER`가 `NULL`인데 타겟에서 `NOT NULL`)나 인코딩 문제 발생 시 디버깅에 수십 분~수 시간이 소요됩니다. Web UI에서도 `"error"` 타입 메시지만 표시되어 사용자가 원인을 파악하기 어렵습니다.

### 2.4. PostgreSQL COPY 모드에서 진행률 미표시 및 부분 재개 불가

- **현상**: `pgx.CopyFrom`은 전체 데이터를 한 번에 스트리밍하므로, 진행 중간에 `tracker.Update()`가 호출되지 않습니다. 또한 COPY 중간에 실패하면 전체 테이블을 처음부터 다시 복사해야 합니다.
- **문제**: 대용량 테이블(수천만 건) 마이그레이션 시 Web UI가 0% → 100%로 갑자기 점프하여 사용자 경험이 나쁘고, 실패 시 체크포인트 기반 재개가 불가능합니다. v8에서 도입한 `MigrationState`가 COPY 모드에서는 무용지물입니다.

### 2.5. SQL Injection 취약점

- **현상**: `migration.go`에서 `fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)` 및 `fmt.Sprintf("SELECT * FROM %s", tableName)` 형태로 테이블명을 직접 쿼리에 삽입합니다.
- **문제**: Web UI를 통해 악의적인 테이블명이 전달될 경우 SQL Injection이 가능합니다. `validateMigrationRequest()`에서 테이블명 검증이 누락되어 있습니다.

### 2.6. 대형 소스 파일의 유지보수성 저하

- **현상**: Dialect 구현 파일들이 5,000~8,650줄에 달하며(`mssql.go` 8,650줄, `postgres.go` 7,522줄, `mysql.go` 7,161줄), `index.html`이 2,000줄 이상의 CSS/JS를 인라인으로 포함합니다.
- **문제**: 단일 파일이 커질수록 변경 시 merge conflict 확률 증가, IDE 성능 저하, 코드 리뷰 효율 감소가 발생합니다. 특히 dialect 파일은 DDL 생성과 INSERT 생성이 단일 파일에 혼재되어 있어 관심사 분리가 미흡합니다.

---

## 3. 목표 (Goals)

1. **데이터 무결성 보장**: 마이그레이션 후 소스-타겟 간 자동 검증(행 수 비교 + 체크섬 샘플링)을 수행하여 정합성을 보증한다.
2. **완전한 관측성**: 마이그레이션 전 과정을 구조화된 감사 로그로 기록하고, 결과 리포트를 생성한다.
3. **정밀한 에러 진단**: 실패 시 배치 번호, 행 오프셋, 문제 컬럼, 원인 분류를 포함한 상세 에러 컨텍스트를 제공한다.
4. **COPY 모드 개선**: PostgreSQL 대상 대용량 마이그레이션에서도 실시간 진행률 표시와 배치 단위 재개를 지원한다.
5. **보안 강화**: 모든 사용자 입력(테이블명, 스키마명)에 대한 검증을 강화하여 SQL Injection을 방지한다.
6. **코드 구조 개선**: Dialect 파일 분할 및 Web UI 에셋 분리로 유지보수성을 확보한다.

---

## 4. 기능 요구사항 (Functional Requirements)

### 4.1. 마이그레이션 후 데이터 검증 (Post-Migration Validation)

- **행 수(Row Count) 비교**:
  - 각 테이블에 대해 소스(`SELECT COUNT(*) FROM {table}`)와 타겟의 행 수를 비교한다.
  - 불일치 시 `warning` 레벨로 로그를 남기고, WebSocket으로 `validation_result` 메시지를 전송한다.

- **체크섬 샘플링 (선택적)**:
  - `--validate` 플래그(UI: "마이그레이션 후 검증") 활성화 시, 각 테이블에서 무작위 N개(기본 1000개) 행을 추출하여 소스/타겟의 해시 값을 비교한다.
  - 체크섬 알고리즘: 각 행의 모든 컬럼 값을 문자열로 직렬화 후 SHA-256 해시 비교.

- **검증 결과 리포트**:
  - JSON 형식의 검증 결과 파일(`validation_{job_id}.json`)을 생성한다.
  - Web UI에서 결과 단계(Step 4)에 검증 상태(Pass/Fail/Warning)를 테이블별로 표시한다.

- **WebSocket 프로토콜 확장**:
  ```json
  { "type": "validation_start", "table": "USERS" }
  { "type": "validation_result", "table": "USERS", "source_count": 50000, "target_count": 50000, "status": "pass" }
  { "type": "validation_result", "table": "ORDERS", "source_count": 120000, "target_count": 119998, "status": "mismatch", "detail": "2 rows missing" }
  ```

### 4.2. 구조화된 감사 로그 (Audit Trail)

- **마이그레이션 리포트 파일**:
  - 각 마이그레이션 수행 시 `.migration_state/{job_id}_report.json` 파일에 다음 정보를 기록한다:
    - `job_id`, 시작/종료 시각, 소스 DB 접속 정보(URL, 비밀번호 제외), 타겟 DB 종류 및 접속 정보
    - 테이블별: 행 수, 소요 시간(초), DDL 수행 여부, 오류 목록
    - 전체 요약: 총 행 수, 총 소요 시간, 성공/실패 테이블 수

- **Web UI 결과 대시보드 강화**:
  - Step 4(결과)에 테이블별 소요 시간, 처리 속도(rows/sec), 오류 수를 표 형태로 표시한다.
  - 리포트 JSON 다운로드 버튼을 추가한다.

- **CLI 출력**:
  - 마이그레이션 종료 시 요약 테이블을 터미널에 출력한다:
    ```
    ┌──────────┬────────┬──────────┬─────────┐
    │ Table    │ Rows   │ Duration │ Status  │
    ├──────────┼────────┼──────────┼─────────┤
    │ USERS    │ 50,000 │ 12.3s    │ OK      │
    │ ORDERS   │ 120K   │ 45.1s    │ OK      │
    │ LOGS     │ 2.1M   │ 5m23s    │ ERROR   │
    └──────────┴────────┴──────────┴─────────┘
    ```

### 4.3. 상세 에러 컨텍스트 (Rich Error Context)

- **구조화 에러 타입 도입**:
  ```go
  type MigrationError struct {
      Table      string
      Phase      string // "ddl", "data", "index", "constraint", "validation"
      BatchNum   int
      RowOffset  int
      Column     string // 문제 컬럼 (파악 가능한 경우)
      RootCause  error
      Suggestion string // 사용자에게 제안할 복구 방법
  }
  ```

- **에러 분류 체계**:
  - `TYPE_MISMATCH`: 데이터 타입 호환 불가 (예: Oracle CLOB → MySQL VARCHAR(255) 초과)
  - `NULL_VIOLATION`: NOT NULL 컬럼에 NULL 삽입 시도
  - `FK_VIOLATION`: 참조 무결성 위반
  - `CONNECTION_LOST`: 네트워크 단절
  - `TIMEOUT`: 쿼리 타임아웃
  - `PERMISSION_DENIED`: 권한 부족

- **WebSocket 에러 메시지 강화**:
  ```json
  {
    "type": "error",
    "table": "ORDERS",
    "error": "TYPE_MISMATCH: column DESCRIPTION (CLOB → VARCHAR(255)) exceeds max length at batch 42, row 41023",
    "suggestion": "Target column DESCRIPTION should be LONGTEXT or TEXT type",
    "phase": "data",
    "recoverable": true
  }
  ```

### 4.4. PostgreSQL COPY 모드 개선 — 배치 분할 COPY

- **배치 단위 COPY 전환**:
  - 기존: `CopyFrom`으로 전체 테이블을 단일 스트림 전송
  - 개선: `SELECT * FROM {table} ORDER BY ROWID OFFSET {n} ROWS FETCH NEXT {batch} ROWS ONLY` → 배치별 `CopyFrom` 실행
  - 각 배치 완료 시 `tracker.Update()` 호출 및 `MigrationState.UpdateOffset()` 저장

- **구현 전략**:
  - `--copy-batch` 플래그(기본값: 10000)로 COPY 배치 크기를 별도 제어한다.
  - 각 배치는 독립 트랜잭션으로 실행하되, 실패 시 해당 배치만 재시도한다.
  - 성능 저하를 최소화하기 위해 Oracle 측 `FETCH NEXT` 기반 페이징을 사용한다.

- **성능 벤치마크 기준**:
  - 배치 분할 COPY는 단일 COPY 대비 최대 20% 성능 저하를 허용한다.
  - 20% 이상 저하 시 기존 단일 COPY 모드를 `--copy-batch 0`으로 선택 가능하게 한다.

### 4.5. 입력 검증 및 보안 강화

- **테이블명 검증**:
  - Oracle 식별자 규칙에 따라 `^[A-Za-z_][A-Za-z0-9_$#]{0,127}$` 패턴으로 테이블명을 검증한다.
  - SQL 쿼리에서 테이블명 사용 시 반드시 `QuoteIdentifier()`를 통해 이스케이프한다.
  - `validateMigrationRequest()`에 테이블명 검증 로직을 추가한다.

- **쿼리 파라미터화**:
  - `fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)` → 식별자 검증 후 `QuoteIdentifier()` 적용
  - Oracle은 prepared statement에서 식별자를 파라미터로 받지 못하므로, 허용목록(allowlist) 기반 검증을 적용한다.

- **Oracle 식별자 Quoting 함수 추가**:
  - `dialect.QuoteOracleIdentifier(name string) string` 함수를 신설하여 Oracle 쿼리에서도 안전한 식별자 사용을 보장한다.

### 4.6. Dialect 코드 구조 개선

- **파일 분할 전략**:
  - 각 dialect 디렉토리를 하위 패키지로 분리하지 않고, 단일 패키지 내에서 파일만 분할한다:
    ```
    internal/dialect/
    ├── dialect.go              # 인터페이스 정의 (변경 없음)
    ├── postgres_ddl.go         # CreateTableDDL, CreateSequenceDDL, CreateIndexDDL, CreateConstraintDDL
    ├── postgres_insert.go      # InsertStatement, 값 직렬화
    ├── postgres_types.go       # MapOracleType, QuoteIdentifier, NormalizeURL
    ├── mysql_ddl.go
    ├── mysql_insert.go
    ├── mysql_types.go
    ├── mssql_ddl.go
    ├── mssql_insert.go
    ├── mssql_types.go
    ├── sqlite_ddl.go
    ├── sqlite_insert.go
    ├── sqlite_types.go
    ├── mariadb.go              # 얇은 래퍼이므로 단일 파일 유지
    └── *_test.go               # 기존 테스트 파일 유지 (분할 불필요)
    ```

- **분할 원칙**:
  - 구조체 정의와 `Name()`, `DriverName()` 등 메타 함수는 `*_types.go`에 배치
  - DDL 생성 로직은 `*_ddl.go`에 배치
  - INSERT/값 직렬화 로직은 `*_insert.go`에 배치
  - 기존 테스트는 변경하지 않음 (파일 분할은 리팩토링이지 기능 변경이 아님)

---

## 5. 아키텍처 및 인터페이스 변경 (Scope of Changes)

| 변경 영역 | 파일 / 모듈 | 상세 내용 |
|-----------|-------------|----------|
| **검증 엔진** | `migration/validation.go` (신규) | `ValidateTable()`: 소스-타겟 행 수 비교, 체크섬 샘플링 로직 |
| **감사 로그** | `migration/report.go` (신규) | `MigrationReport` 구조체, JSON 직렬화, CLI 테이블 출력 |
| **에러 타입** | `migration/errors.go` (신규) | `MigrationError` 구조화 에러, 에러 분류 상수 |
| **COPY 개선** | `migration/migration.go` | `MigrateTableDirect()` 내 PostgreSQL COPY를 배치 분할 루프로 교체 |
| **입력 검증** | `web/server.go`, `migration/migration.go` | 테이블명 검증 함수 추가, `QuoteOracleIdentifier()` 적용 |
| **Dialect 분할** | `dialect/*.go` | 기존 대형 파일을 `*_ddl.go`, `*_insert.go`, `*_types.go`로 분할 |
| **Config** | `config/config.go` | `--validate`, `--copy-batch` 플래그 추가 |
| **WebSocket** | `web/ws/tracker.go` | `ValidationStart()`, `ValidationResult()` 메서드 추가 |
| **Web UI** | `templates/index.html` | 검증 결과 표시, 리포트 다운로드, 상세 에러 패널 |
| **ProgressTracker** | `migration/migration.go` | `ValidationTracker` 인터페이스 추가 |

---

## 6. 비기능 요구사항 (Non-Functional Requirements)

- **성능**: 검증 단계는 마이그레이션 시간의 10% 이내로 완료되어야 한다. 체크섬 샘플링은 전수 검사가 아닌 통계적 샘플링이므로 대용량 테이블에서도 일정한 시간(최대 30초/테이블)을 유지한다.
- **하위 호환성**: `--validate`, `--copy-batch` 미지정 시 v8과 동일하게 동작한다. 기존 CLI/Web UI 워크플로우에 영향 없음.
- **보안**: 모든 사용자 입력 경로에 대한 식별자 검증을 적용하여, 알려진 SQL Injection 패턴을 차단한다.
- **코드 품질**: Dialect 파일 분할 후 기존 테스트가 100% 통과해야 한다. 새로운 기능에 대한 유닛 테스트 커버리지 80% 이상.

---

## 7. 마일스톤 (Milestones)

1. **1단계: 보안 강화** — 테이블명/스키마명 입력 검증, `QuoteOracleIdentifier()` 도입, SQL Injection 방어
2. **2단계: 에러 구조화** — `MigrationError` 타입 도입, 에러 분류 체계, WebSocket 에러 메시지 강화
3. **3단계: COPY 모드 개선** — 배치 분할 COPY, 진행률 표시, 체크포인트 연동
4. **4단계: 감사 로그 및 리포트** — `MigrationReport` 생성, CLI 요약 테이블, Web UI 결과 대시보드 강화
5. **5단계: 데이터 검증** — 행 수 비교, 체크섬 샘플링, 검증 결과 WebSocket 전송
6. **6단계: Dialect 리팩토링** — 대형 파일 분할, 기존 테스트 통과 확인
7. **7단계: 통합 테스트 및 QA** — 전체 시나리오 테스트, 성능 벤치마크, 보안 검증

---

## 8. 기대 효과 (Expected Outcomes)

v9 업데이트를 통해 **"마이그레이션을 수행했으나 결과를 신뢰할 수 없다"는 운영 현장의 핵심 불안 요소**를 제거합니다.

- **데이터 검증**: 자동화된 소스-타겟 정합성 검증으로 수동 검증 시간을 90% 이상 절감
- **감사 추적**: 구조화된 마이그레이션 리포트로 컴플라이언스 요건 충족 및 장애 시 원인 분석 시간 단축
- **에러 진단**: 상세 에러 컨텍스트로 디버깅 시간을 수십 분에서 수 분으로 단축
- **COPY 모드**: 대용량 PostgreSQL 마이그레이션에서도 실시간 진행률과 안정적 재개를 보장
- **보안**: SQL Injection 취약점 제거로 Web UI의 프로덕션 환경 배포 안전성 확보
- **유지보수**: 코드 구조 개선으로 향후 기능 추가 및 버그 수정 속도 향상
