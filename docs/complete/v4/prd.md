# PRD (Product Requirements Document) - Web UI CLI 파라미터 통합 (v4)

## 1. 개요 (Overview)

현재 Web UI에서는 CLI 모드에서 제공하는 일부 파라미터만 제어 가능합니다. CLI 전용으로 남아있는 5개 파라미터(`-out`, `-per-table`, `-schema`, `-dry-run`, `-log-json`)를 Web UI에서도 표시 및 제어할 수 있도록 확장합니다.

## 2. 배경 (Background)

### 2.1. 현재 상태

| 파라미터 | CLI 플래그 | Web UI 지원 | 비고 |
|----------|-----------|-------------|------|
| Oracle URL | `-url` | O | 연결 폼 |
| Username | `-user` | O | 연결 폼 |
| Password | `-password` | O | 연결 폼 |
| Tables | `-tables` | O | 체크박스 선택 |
| Batch Size | `-batch` | O | 고급 설정 |
| Workers | `-workers` | O | 고급 설정 |
| PostgreSQL URL | `-pg-url` | O | Direct 모드 |
| DDL 생성 | `-with-ddl` | O | Direct 모드 체크박스 |
| **Output 파일명** | `-out` | **X** | 기본값 `migration.sql` |
| **테이블별 파일 분리** | `-per-table` | **X** | 기본값 `false` |
| **PostgreSQL 스키마** | `-schema` | **X** | 기본값 빈 문자열 |
| **Dry-Run 모드** | `-dry-run` | **X** | 기본값 `false` |
| **JSON 로깅** | `-log-json` | **X** | 기본값 `false` |

### 2.2. 문제점
- Web UI 사용자는 출력 파일명을 지정하거나 테이블별 파일 분리를 선택할 수 없음
- PostgreSQL 스키마를 지정하려면 CLI로 전환해야 함
- 실제 마이그레이션 전 사전 검증(Dry-Run)을 Web UI에서 수행할 수 없음
- 디버깅용 JSON 로그 출력을 Web UI에서 활성화할 수 없음

## 3. 목표 (Goals)

- **CLI-Web 기능 동등성(Feature Parity)**: CLI에서 가능한 모든 설정을 Web UI에서도 동일하게 제어 가능하게 합니다.
- **직관적인 UI 배치**: 기존 고급 설정(Advanced Settings) 섹션을 확장하여 새 파라미터를 자연스럽게 배치합니다.
- **하위 호환성 유지**: 기존 API 요청에 새 필드가 없는 경우 기본값으로 동작합니다.

## 4. 기능 요구사항 (Functional Requirements)

### 4.1. 출력 파일명 설정 (`-out`)

- **UI 컨트롤**: 텍스트 입력 필드
- **위치**: 고급 설정(Advanced Settings) 섹션 내
- **기본값**: `migration.sql`
- **적용 범위**: SQL File 모드에서만 활성화 (Direct 모드에서는 비활성화/숨김)
- **동작**: ZIP 파일 내 생성되는 SQL 파일의 기본 이름으로 사용됨. `-per-table` 활성화 시에는 접두사(prefix)로 활용되거나 무시될 수 있음

### 4.2. 테이블별 파일 분리 (`-per-table`)

- **UI 컨트롤**: 토글 스위치 또는 체크박스
- **위치**: 고급 설정 섹션 내
- **기본값**: `false` (비활성화)
- **적용 범위**: SQL File 모드에서만 활성화
- **동작**:
  - 비활성화 시: 모든 테이블의 INSERT 문이 단일 SQL 파일에 기록됨
  - 활성화 시: 각 테이블별로 별도의 `.sql` 파일이 생성됨 (예: `USERS.sql`, `ORDERS.sql`)
- **참고**: 현재 Web UI는 내부적으로 항상 `PerTable=true`로 동작 중. 이 옵션을 노출하여 사용자가 단일 파일 출력도 선택 가능하게 함

### 4.3. PostgreSQL 스키마 지정 (`-schema`)

- **UI 컨트롤**: 텍스트 입력 필드
- **위치**: 고급 설정 섹션 내
- **기본값**: 빈 문자열 (미지정)
- **Placeholder**: `public`
- **적용 범위**: SQL File 모드, Direct 모드 모두 적용
- **동작**:
  - 값이 지정되면 생성되는 DDL 및 INSERT 문에서 테이블명 앞에 스키마가 붙음 (예: `myschema.USERS`)
  - `-with-ddl` 활성화 시 `CREATE TABLE` 문에도 스키마가 반영됨
  - 빈 값인 경우 기존 동작과 동일 (스키마 미지정)

### 4.4. Dry-Run 모드 (`-dry-run`)

- **UI 컨트롤**: 토글 스위치 또는 체크박스
- **위치**: 마이그레이션 시작 버튼 근처 (잘 보이는 위치)
- **기본값**: `false` (비활성화)
- **동작**:
  - 활성화 시: 실제 데이터 마이그레이션 없이 다음을 수행
    - Oracle DB 연결 확인
    - PostgreSQL DB 연결 확인 (Direct 모드 시)
    - 각 테이블의 총 레코드 수(Row Count) 추정
  - 결과를 WebSocket을 통해 실시간으로 UI에 표시
  - **UI 변화**:
    - Dry-Run 활성화 시 "마이그레이션 시작" 버튼 텍스트가 "사전 검증 실행"으로 변경
    - 진행 상황 섹션에 "Dry-Run 모드" 라벨 표시
    - 완료 후 다운로드 버튼 대신 검증 결과 요약 표시

### 4.5. JSON 로깅 (`-log-json`)

- **UI 컨트롤**: 토글 스위치 또는 체크박스
- **위치**: 고급 설정 섹션 내
- **기본값**: `false` (비활성화)
- **동작**:
  - 활성화 시: 서버 측 로그 출력이 구조화된 JSON 형식으로 전환됨
  - 디버깅 및 로그 수집 시스템 연동에 유용
- **선택적 확장**: 향후 Web UI에 로그 뷰어 패널을 추가하여 JSON 로그를 실시간으로 표시할 수 있음

## 5. API 변경사항 (API Changes)

### 5.1. `POST /api/migrate` 요청 구조 확장

```json
{
  "oracleUrl": "localhost:1521/XE",
  "username": "user",
  "password": "pass",
  "tables": ["USERS", "ORDERS"],
  "direct": false,
  "pgUrl": "",
  "withDdl": true,
  "batchSize": 1000,
  "workers": 4,
  "outFile": "migration.sql",
  "perTable": true,
  "schema": "public",
  "dryRun": false,
  "logJson": false
}
```

### 5.2. 서버 측 `startMigrationRequest` 구조체 변경

```go
type startMigrationRequest struct {
    OracleURL string   `json:"oracleUrl" binding:"required"`
    Username  string   `json:"username" binding:"required"`
    Password  string   `json:"password" binding:"required"`
    Tables    []string `json:"tables" binding:"required"`
    Direct    bool     `json:"direct"`
    PGURL     string   `json:"pgUrl"`
    WithDDL   bool     `json:"withDdl"`
    BatchSize int      `json:"batchSize"`
    Workers   int      `json:"workers"`
    // v4 추가 필드
    OutFile   string   `json:"outFile"`
    PerTable  bool     `json:"perTable"`
    Schema    string   `json:"schema"`
    DryRun    bool     `json:"dryRun"`
    LogJSON   bool     `json:"logJson"`
}
```

### 5.3. WebSocket 메시지 확장 (Dry-Run 모드)

Dry-Run 모드 전용 메시지 타입 추가:

```json
{
  "type": "dry_run_result",
  "table": "USERS",
  "total": 150000,
  "estimatedTime": "~2m30s",
  "connectionOk": true
}
```

## 6. UI/UX 설계 (UI/UX Design)

### 6.1. 고급 설정 섹션 확장

기존 고급 설정(Advanced Settings) `<details>` 요소 내부에 다음 항목을 추가합니다:

```
[고급 설정] ▾
┌─────────────────────────────────────────────────┐
│  Batch Size:     [1000        ]                 │
│  Workers:        [4           ]                 │
│  ─────────────────────────────────               │
│  Output 파일명:  [migration.sql]  ← NEW         │
│  PG 스키마:      [public      ]  ← NEW         │
│  ─────────────────────────────────               │
│  ☐ 테이블별 파일 분리              ← NEW         │
│  ☐ JSON 로깅                      ← NEW         │
└─────────────────────────────────────────────────┘
```

### 6.2. Dry-Run 컨트롤

마이그레이션 시작 버튼 영역에 Dry-Run 토글을 배치합니다:

```
┌─────────────────────────────────────────────────┐
│  ☐ Dry-Run (사전 검증만 수행)      ← NEW        │
│                                                 │
│  [ 🚀 마이그레이션 시작 ]                         │
│  ──── Dry-Run 활성화 시 ────                     │
│  [ 🔍 사전 검증 실행 ]                            │
└─────────────────────────────────────────────────┘
```

### 6.3. 조건부 표시 규칙

| 컨트롤 | SQL File 모드 | Direct 모드 |
|--------|--------------|-------------|
| Output 파일명 | 표시 | 숨김 |
| 테이블별 파일 분리 | 표시 | 숨김 |
| PG 스키마 | 표시 | 표시 |
| Dry-Run | 표시 | 표시 |
| JSON 로깅 | 표시 | 표시 |

## 7. 비기능 요구사항 (Non-Functional Requirements)

- **하위 호환성**: 새 필드가 요청에 포함되지 않을 경우 기존 기본값으로 동작해야 합니다.
- **유효성 검증**:
  - `outFile`에 경로 구분자(`/`, `\`)가 포함된 경우 거부
  - `schema`에 SQL 인젝션 가능한 문자가 포함된 경우 거부
  - `batchSize`, `workers`는 양의 정수만 허용
- **반응형 UI**: 새 컨트롤들이 모바일/작은 화면에서도 정상 표시되어야 합니다.
- **접근성**: 모든 입력 필드에 적절한 `<label>` 및 `aria-` 속성을 부여합니다.

## 8. 영향 범위 (Scope of Changes)

### 8.1. 변경 파일 목록

| 파일 | 변경 내용 |
|------|----------|
| `internal/web/server.go` | `startMigrationRequest` 구조체 확장, 핸들러 로직 수정 |
| `internal/web/templates/index.html` | 고급 설정 UI 확장, Dry-Run 토글, 조건부 표시 로직 |
| `internal/web/ws/tracker.go` | Dry-Run 결과 메시지 타입 추가 |
| `internal/migration/migration.go` | Dry-Run 로직 분기 처리 |
| `internal/logger/logger.go` | 런타임 로그 모드 전환 지원 (선택적) |

### 8.2. 변경하지 않는 파일

| 파일 | 사유 |
|------|------|
| `internal/config/config.go` | CLI 플래그 파싱은 기존 유지, Web API에서 직접 매핑 |
| `main.go` | 진입점 변경 없음 |

## 9. 마일스톤 (Milestones)

1. **PRD 확정**: `docs/v4/prd.md` 작성 및 리뷰
2. **API 확장**: `startMigrationRequest` 구조체 및 핸들러 수정
3. **UI 확장**: 고급 설정 섹션에 새 컨트롤 추가
4. **Dry-Run 구현**: 사전 검증 로직 및 WebSocket 메시지 확장
5. **조건부 표시**: SQL File / Direct 모드에 따른 UI 토글 로직
6. **유효성 검증**: 입력값 검증 로직 추가 (프론트엔드 + 백엔드)
7. **테스트**: 단위 테스트 및 통합 테스트 작성
8. **최종 리뷰**: 코드 리뷰 및 문서 정리

## 10. 향후 확장 고려사항 (Future Considerations)

- **로그 뷰어**: Web UI에 실시간 로그 패널을 추가하여 JSON 로그를 시각적으로 확인
- **설정 프리셋**: 자주 사용하는 설정 조합을 저장/로드하는 기능
- **웹 서버 포트 설정**: 현재 하드코딩된 `8080` 포트를 Web UI에서 설정 가능하게 확장
