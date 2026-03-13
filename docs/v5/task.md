# Implementation Tasks: Sequence / Index 마이그레이션 지원 (v5)

## Phase 1: Config 확장

### 1-1. `Config` 구조체 필드 추가
- [x] `internal/config/config.go`의 `Config`에 4개 필드 추가
  - `WithSequences bool`
  - `WithIndexes bool`
  - `Sequences string` (쉼표 구분 명시 지정용)
  - `OracleOwner string`

### 1-2. CLI 플래그 등록
- [x] `--with-sequences` bool 플래그 추가 (`"연관 Sequence DDL 포함"`)
- [x] `--with-indexes` bool 플래그 추가 (`"연관 Index DDL 포함"`)
- [x] `--sequences` string 플래그 추가 (`"추가 포함할 Sequence 이름 목록 (쉼표 구분)"`)
- [x] `--oracle-owner` string 플래그 추가 (`"Oracle 스키마 소유자 (미지정 시 -user 값 사용)"`)

### 1-3. 사용 예시 업데이트
- [x] `flag.Usage` 예시 블록에 `--with-sequences`, `--with-indexes` 조합 예시 추가

---

## Phase 2: DDL 로직 구현 (`ddl.go`)

### 2-1. Sequence 메타데이터 조회
- [x] `SequenceMetadata` 구조체 정의
  - `Name string`
  - `MinValue int64`
  - `MaxValue string` (Oracle 최대값은 28자리 → string)
  - `IncrementBy int64`
  - `CycleFlag string` (`"Y"` / `"N"`)
  - `LastNumber int64`
- [x] `GetSequenceMetadata(db *sql.DB, tableName, owner string, extraNames []string) ([]SequenceMetadata, error)` 구현
  - `ALL_TAB_COLUMNS.DATA_DEFAULT`에서 `.NEXTVAL` 포함 컬럼으로 연관 Sequence 이름 추출
  - 이름 패턴(`<TABLE>_SEQ`, `<TABLE>_ID_SEQ`, `SEQ_<TABLE>`) 조회
  - `extraNames`(명시 지정 목록) 병합, 중복 제거
  - `ALL_SEQUENCES` 에서 메타데이터 조회

### 2-2. Sequence DDL 생성
- [x] `GenerateSequenceDDL(seq SequenceMetadata, schema string) string` 구현
  - `CREATE SEQUENCE IF NOT EXISTS {schema.}name` 출력
  - `START WITH last_number`
  - `INCREMENT BY increment_by`
  - `MINVALUE min_value`
  - Oracle 기본 MAXVALUE(28자리 9) 이상이면 `MAXVALUE` 절 생략
  - `CYCLE` / `NO CYCLE` 조건 처리

### 2-3. Index 메타데이터 조회
- [x] `IndexMetadata` 구조체 정의
  - `Name string`
  - `Uniqueness string` (`"UNIQUE"` / `"NONUNIQUE"`)
  - `IndexType string` (`"NORMAL"` / `"FUNCTION-BASED NORMAL"`)
  - `IsPK bool` (index_name이 `SYS_C%` 패턴)
  - `Columns []IndexColumn`
- [x] `IndexColumn` 구조체 정의
  - `Name string`
  - `Position int`
  - `Descend string` (`"ASC"` / `"DESC"`)
- [x] `GetIndexMetadata(db *sql.DB, tableName, owner string) ([]IndexMetadata, error)` 구현
  - `ALL_INDEXES` + `ALL_IND_COLUMNS` JOIN 조회
  - `index_type IN ('NORMAL', 'FUNCTION-BASED NORMAL')` 필터
  - `LOB` 인덱스 제외
  - `SYS_C%` 패턴 → `IsPK = true` 표시

### 2-4. Index DDL 생성
- [x] `GenerateIndexDDL(idx IndexMetadata, tableName, schema string) string` 구현
  - `IsPK=true`: `ALTER TABLE {schema.}table ADD PRIMARY KEY (col)` 출력
  - `Uniqueness="UNIQUE"`: `CREATE UNIQUE INDEX IF NOT EXISTS` 출력
  - 일반: `CREATE INDEX IF NOT EXISTS` 출력
  - `Descend="DESC"` 컬럼은 `col DESC` 로 표현

---

## Phase 3: 마이그레이션 연동 (`migration.go`)

### 3-1. `MigrateTableToFile` 확장
- [x] `cfg.WithSequences` true 시 `GetSequenceMetadata` 호출 후 `GenerateSequenceDDL` 결과를 `CREATE TABLE` **이전**에 출력
- [x] `cfg.WithIndexes` true 시 `GetIndexMetadata` 호출 후 `GenerateIndexDDL` 결과를 `CREATE TABLE` **이후**, INSERT **이전**에 출력
- [x] Sequence/Index 조회 실패 시 경고 로그 후 해당 객체 스킵 (전체 중단 없음)

### 3-2. `MigrateTableDirect` 확장
- [x] `cfg.WithSequences` true 시 Sequence DDL을 `pgPool.Exec`으로 실행
- [x] `cfg.WithIndexes` true 시 Index DDL을 `pgPool.Exec`으로 실행 (COPY 완료 후)
- [x] 각 DDL 실행 후 tracker가 있으면 `DDLProgress` 메시지 전송

### 3-3. OracleOwner 기본값 처리
- [x] `cfg.OracleOwner`가 빈 문자열이면 `cfg.User`를 대문자 변환하여 사용

---

## Phase 4: WebSocket 프로토콜 확장

### 4-1. 메시지 타입 및 구조체 추가
- [x] `internal/web/ws/tracker.go`에 `MsgDDLProgress MsgType = "ddl_progress"` 상수 추가
- [x] `ProgressMsg`에 다음 필드 추가
  - `Object string json:"object,omitempty"` (`"sequence"` / `"index"`)
  - `ObjectName string json:"object_name,omitempty"`
  - `Status string json:"status,omitempty"` (`"ok"` / `"error"`)

### 4-2. `DDLProgress` 메서드 구현
- [x] `WebSocketTracker`에 `DDLProgress(object, name, status string, err error)` 메서드 추가
- [x] `status="error"` 시 `ErrorMsg` 필드에 에러 내용 포함

---

## Phase 5: Web API 확장 (`server.go`)

### 5-1. `startMigrationRequest` 구조체 확장
- [x] `WithSequences bool json:"withSequences"` 필드 추가
- [x] `WithIndexes bool json:"withIndexes"` 필드 추가
- [x] `OracleOwner string json:"oracleOwner"` 필드 추가

### 5-2. Config 매핑
- [x] `startMigration` 핸들러에서 `cfg.WithSequences`, `cfg.WithIndexes`, `cfg.OracleOwner` 매핑 추가

---

## Phase 6: Web UI 확장 (`index.html`)

### 6-1. 고급 설정 섹션 체크박스 추가
- [x] `--with-ddl` 체크박스 아래에 구분 없이 연속 배치
  - `Sequence 포함 (--with-sequences)` 체크박스 추가 (id: `withSequences`, 기본값: unchecked)
  - `Index 포함 (--with-indexes)` 체크박스 추가 (id: `withIndexes`, 기본값: unchecked)
- [x] `Oracle 소유자` 텍스트 입력 추가 (id: `oracleOwner`, placeholder: `접속 계정과 동일`)

### 6-2. 조건부 활성화 JavaScript 로직
- [x] `withDdl` 체크박스 change 이벤트: unchecked 시 `withSequences`, `withIndexes` 체크 해제 및 비활성화
- [x] `withDdl` checked 시 두 체크박스 활성화

### 6-3. 요청 payload 확장
- [x] `btnMigrate` 클릭 핸들러에서 `withSequences`, `withIndexes`, `oracleOwner` 값 수집 및 전송

### 6-4. DDL Progress WebSocket 메시지 처리
- [x] `handleProgressMessage`에 `ddl_progress` 타입 분기 추가
- [x] 진행 목록에 `[sequence] SEQ_USERS ✓` / `[index] IDX_USERS_EMAIL ✓` 형태로 표시
- [x] `status="error"` 시 경고 아이콘 + 에러 메시지 표시

---

## Phase 7: 테스트

### 7-1. 단위 테스트 (`ddl_test.go` 확장)
- [ ] `TestGenerateSequenceDDL_Basic` - 기본 Sequence DDL 생성 검증
- [ ] `TestGenerateSequenceDDL_MaxValueOmit` - Oracle 기본 MAXVALUE 생략 검증
- [ ] `TestGenerateSequenceDDL_Cycle` - CYCLE 옵션 반영 검증
- [ ] `TestGenerateSequenceDDL_WithSchema` - 스키마 접두사 포함 검증
- [ ] `TestGenerateIndexDDL_Normal` - 일반 Index DDL 검증
- [ ] `TestGenerateIndexDDL_Unique` - Unique Index DDL 검증
- [ ] `TestGenerateIndexDDL_PrimaryKey` - PK → ALTER TABLE 변환 검증
- [ ] `TestGenerateIndexDDL_Descend` - DESC 컬럼 표현 검증
- [ ] `TestDDLProgress_Broadcast` - DDLProgress WebSocket 메시지 검증

### 7-2. 통합 테스트 (`v5_integration_test.go` 신규)
- [ ] `WithSequences=true`: Sequence DDL이 CREATE TABLE 이전에 출력되는지 검증
- [ ] `WithIndexes=true`: Index DDL이 CREATE TABLE 이후에 출력되는지 검증
- [ ] `WithSequences=false, WithIndexes=false`: 기존 동작 완전 유지 검증
- [ ] OracleOwner 기본값: 빈 문자열 시 User 값 대문자로 대체 검증

### 7-3. 프론트엔드 수동 테스트
- [ ] `withDdl` 미체크 시 `withSequences`, `withIndexes` 비활성화 확인
- [ ] `ddl_progress` 메시지 UI 렌더링 확인

---

## Phase 8: 동기화 및 마무리

### 8-1. 템플릿 동기화
- [ ] `internal/web/templates/index.html` 변경사항을 `web/templates/index.html`에 동기화

### 8-2. 빌드 및 검증
- [ ] `go build` 성공 확인
- [ ] `go test ./...` 전체 테스트 통과 확인
- [ ] `go vet ./...` 정적 분석 통과 확인

### 8-3. 문서 정리
- [ ] `docs/v5/prd.md` 최종 확인
- [ ] `docs/v5/task.md` 완료 항목 체크
