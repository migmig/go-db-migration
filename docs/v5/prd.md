# PRD (Product Requirements Document) - Sequence / Index 마이그레이션 지원 (v5)

## 1. 개요 (Overview)

현재 `--with-ddl` 옵션은 Oracle 테이블의 컬럼 정의(`CREATE TABLE`)만 PostgreSQL로 변환합니다.
실제 운영 DB 이전에는 **Sequence**(자동 증가 값 생성기)와 **Index**(검색 성능 구조체)도 함께 옮겨야 합니다.
이 버전에서는 `--with-sequences`, `--with-indexes` 두 옵션을 추가하여 테이블 DDL과 함께
관련 Sequence·Index를 Oracle에서 읽어 PostgreSQL 호환 DDL로 변환·출력합니다.

---

## 2. 배경 (Background)

### 2.1. 현재 한계

| 객체 | `--with-ddl` 지원 | 비고 |
|------|:-----------------:|------|
| 테이블 컬럼 정의 | O | `CREATE TABLE IF NOT EXISTS` 생성 |
| Sequence | X | Oracle `CREATE SEQUENCE` → PG 미변환 |
| 일반 Index | X | Oracle `CREATE INDEX` → PG 미변환 |
| Unique Index | X | Oracle `UNIQUE` → PG 미변환 |
| Primary Key | X | Oracle `PRIMARY KEY` 제약 → PG 미변환 |
| Foreign Key | △ | 범위 외 (v5 미포함) |

### 2.2. Oracle ↔ PostgreSQL 객체 대응

| Oracle | PostgreSQL |
|--------|-----------|
| `CREATE SEQUENCE s START WITH n INCREMENT BY m MINVALUE a MAXVALUE b CYCLE/NOCYCLE` | `CREATE SEQUENCE IF NOT EXISTS s START n INCREMENT m MINVALUE a MAXVALUE b [CYCLE]` |
| `CREATE INDEX i ON t (col1, col2)` | `CREATE INDEX IF NOT EXISTS i ON t (col1, col2)` |
| `CREATE UNIQUE INDEX i ON t (col)` | `CREATE UNIQUE INDEX IF NOT EXISTS i ON t (col)` |
| `PRIMARY KEY (col)` 제약 (테이블 단위) | `ALTER TABLE t ADD PRIMARY KEY (col)` |

---

## 3. 목표 (Goals)

- Oracle DB의 지정된 테이블과 **연관된 Sequence·Index**를 자동으로 추출해 PostgreSQL DDL로 변환합니다.
- 기존 `--with-ddl` 플래그와 **독립적으로** 선택 가능하게 합니다 (조합 자유).
- SQL File 출력 모드와 Direct(PG 직접 실행) 모드 모두 지원합니다.
- Web UI에서도 동일하게 제어할 수 있도록 확장합니다.
- 하위 호환성을 유지합니다 (기존 옵션 미설정 시 동작 불변).

---

## 4. 기능 요구사항 (Functional Requirements)

### 4.1. Sequence 마이그레이션 (`--with-sequences`)

#### 4.1.1. Oracle 메타데이터 조회

`ALL_SEQUENCES` 뷰에서 지정 테이블과 **이름이 연관된** Sequence를 조회합니다.

연관 판별 기준 (우선순위 순):
1. 테이블의 컬럼 중 `DEFAULT` 값에 해당 Sequence의 `.NEXTVAL`이 포함된 경우 (`ALL_TAB_COLUMNS.DATA_DEFAULT`)
2. Sequence 이름이 `<TABLE_NAME>_SEQ`, `<TABLE_NAME>_ID_SEQ`, `SEQ_<TABLE_NAME>` 패턴인 경우
3. `--sequences` 플래그로 명시적으로 지정한 Sequence 이름

조회 쿼리 (안):
```sql
SELECT sequence_name, min_value, max_value, increment_by,
       cycle_flag, last_number
FROM   all_sequences
WHERE  sequence_owner = :owner
  AND  sequence_name IN (
         SELECT REGEXP_SUBSTR(data_default, '[A-Z0-9_$#]+', 1, 1)
         FROM   all_tab_columns
         WHERE  table_name = :table
           AND  data_default LIKE '%.NEXTVAL%'
         UNION ALL
         SELECT sequence_name
         FROM   all_sequences
         WHERE  sequence_name IN (
                  :table || '_SEQ',
                  :table || '_ID_SEQ',
                  'SEQ_' || :table
                )
       )
```

#### 4.1.2. PostgreSQL DDL 생성

```sql
CREATE SEQUENCE IF NOT EXISTS {schema.}seq_name
    START WITH {last_number}
    INCREMENT BY {increment_by}
    MINVALUE {min_value}
    MAXVALUE {max_value}
    [CYCLE | NO CYCLE];
```

- `last_number`는 Oracle의 현재 Sequence 값을 반영해 충돌 없이 이어받음
- `MAXVALUE`가 Oracle 기본값(`9999999999999999999999999999`) 이상이면 PostgreSQL 기본값으로 생략
- DDL은 `CREATE TABLE` **이전**에 출력되어 `DEFAULT nextval(...)` 컬럼 정의보다 앞에 위치

#### 4.1.3. 컬럼 기본값 연동

`--with-ddl`과 함께 사용할 경우, `DEFAULT` 값에 `.NEXTVAL`이 있는 컬럼은 PostgreSQL DDL에서:

```sql
column_name bigint DEFAULT nextval('schema.seq_name')
```

로 자동 변환합니다.

---

### 4.2. Index 마이그레이션 (`--with-indexes`)

#### 4.2.1. Oracle 메타데이터 조회

`ALL_INDEXES`, `ALL_IND_COLUMNS` 뷰에서 테이블의 Index 목록과 컬럼 정보를 조회합니다.

```sql
SELECT i.index_name,
       i.uniqueness,
       i.index_type,
       c.column_name,
       c.column_position,
       c.descend
FROM   all_indexes i
JOIN   all_ind_columns c
  ON   c.index_name  = i.index_name
 AND   c.table_owner = i.owner
WHERE  i.table_name  = :table
  AND  i.owner       = :owner
  AND  i.index_type IN ('NORMAL', 'FUNCTION-BASED NORMAL')
ORDER  BY i.index_name, c.column_position
```

제외 대상:
- `index_type = 'LOB'` (BLOB/CLOB 관리용 내부 인덱스)
- Oracle 내부 PK 인덱스(`index_name` 패턴 `SYS_C%`)는 `ALTER TABLE ADD PRIMARY KEY`로 대체

#### 4.2.2. PostgreSQL DDL 생성

```sql
-- 일반 인덱스
CREATE INDEX IF NOT EXISTS {index_name} ON {schema.}{table_name} ({col1} [DESC], {col2});

-- Unique 인덱스
CREATE UNIQUE INDEX IF NOT EXISTS {index_name} ON {schema.}{table_name} ({col});

-- Primary Key (SYS_C% 계열)
ALTER TABLE {schema.}{table_name} ADD PRIMARY KEY ({pk_col});
```

- `DESCEND = 'DESC'`인 컬럼은 `col DESC` 으로 표현
- Function-based index (`FUNCTION-BASED NORMAL`)의 경우 표현식 그대로 사용
- DDL은 `INSERT` 문 **이후** (또는 `CREATE TABLE` 바로 다음) 에 출력

#### 4.2.3. 출력 순서

```
-- Sequence DDL (--with-sequences)
CREATE SEQUENCE IF NOT EXISTS ...;

-- Table DDL (--with-ddl)
CREATE TABLE IF NOT EXISTS ...;

-- Index DDL (--with-indexes)
CREATE INDEX IF NOT EXISTS ...;
ALTER TABLE ... ADD PRIMARY KEY (...);

-- Data INSERT
INSERT INTO ... VALUES ...;
```

---

### 4.3. CLI 플래그

| 플래그 | 타입 | 기본값 | 설명 |
|--------|------|--------|------|
| `--with-sequences` | bool | `false` | 연관 Sequence DDL 포함 |
| `--with-indexes` | bool | `false` | 연관 Index DDL 포함 |
| `--sequences` | string | `""` | 추가로 포함할 Sequence 이름 목록 (쉼표 구분) |
| `--oracle-owner` | string | `""` | Oracle 스키마(소유자) 이름. 미지정 시 `-user` 값 사용 |

#### 조합 예시

```bash
# 테이블 DDL + Sequence + Index 모두 포함하여 SQL 파일로 출력
dbmigrator -url localhost:1521/ORCL -user scott -password tiger \
  -tables USERS,ORDERS \
  -with-ddl -with-sequences -with-indexes

# 직접 마이그레이션 + Index만
dbmigrator -url localhost:1521/ORCL -user scott -password tiger \
  -tables USERS \
  -pg-url postgres://pguser:pgpass@localhost:5432/mydb \
  -with-ddl -with-indexes -schema myschema

# Sequence 이름 명시 지정
dbmigrator ... -with-sequences -sequences SEQ_USERS,SEQ_ORDERS
```

---

### 4.4. Web UI 확장

#### 4.4.1. 고급 설정 섹션에 추가

기존 `--with-ddl` 체크박스 아래에 다음을 추가합니다:

```
[고급 설정] ▾
┌──────────────────────────────────────────────────────┐
│  ...                                                 │
│  ☑ DDL 생성 (CREATE TABLE)         ← 기존            │
│  ☐ Sequence 포함 (--with-sequences)  ← NEW          │
│  ☐ Index 포함   (--with-indexes)     ← NEW          │
│  Oracle 소유자: [           ]         ← NEW          │
└──────────────────────────────────────────────────────┘
```

- `Sequence 포함`, `Index 포함` 체크박스는 `DDL 생성`이 체크된 경우에만 활성화(enable) 처리
- `Oracle 소유자` 입력은 비어있으면 연결 시 사용한 Username으로 서버 측 대체

#### 4.4.2. API 요청 필드 추가

```json
{
  "withSequences": true,
  "withIndexes": true,
  "oracleOwner": "SCOTT"
}
```

#### 4.4.3. WebSocket 진행 메시지 확장

Sequence/Index DDL 실행 중 진행 상황을 알리는 새 메시지 타입:

```json
{ "type": "ddl_progress", "object": "sequence", "name": "SEQ_USERS", "status": "ok" }
{ "type": "ddl_progress", "object": "index",    "name": "IDX_USERS_EMAIL", "status": "ok" }
{ "type": "ddl_progress", "object": "index",    "name": "IDX_ORDERS_DATE", "status": "error", "error": "..." }
```

---

## 5. 비기능 요구사항 (Non-Functional Requirements)

- **멱등성**: `IF NOT EXISTS` 사용으로 재실행 시 오류 없이 스킵
- **오류 격리**: 특정 Sequence/Index 변환 실패 시 해당 객체만 경고 로그 후 계속 진행 (전체 중단 없음)
- **권한 최소화**: `ALL_SEQUENCES`, `ALL_INDEXES`, `ALL_IND_COLUMNS`, `ALL_TAB_COLUMNS` 조회 권한만 필요
- **성능**: 메타데이터 조회는 마이그레이션 시작 전 1회만 수행, 데이터 이전 성능에 영향 없음
- **하위 호환성**: `--with-sequences`, `--with-indexes` 미지정 시 기존 동작 완전 유지

---

## 6. 영향 범위 (Scope of Changes)

| 파일 | 변경 내용 |
|------|----------|
| `internal/config/config.go` | `WithSequences`, `WithIndexes`, `Sequences`, `OracleOwner` 필드 및 플래그 추가 |
| `internal/migration/ddl.go` | `GetSequenceMetadata`, `GenerateSequenceDDL`, `GetIndexMetadata`, `GenerateIndexDDL` 함수 추가 |
| `internal/migration/migration.go` | `MigrateTable` 내에서 Sequence/Index DDL 출력 로직 추가 |
| `internal/web/server.go` | `startMigrationRequest`에 `WithSequences`, `WithIndexes`, `OracleOwner` 필드 추가 |
| `internal/web/ws/tracker.go` | `MsgDDLProgress` 메시지 타입 및 `DDLProgress()` 메서드 추가 |
| `internal/web/templates/index.html` | 고급 설정 체크박스 2개 + Oracle 소유자 입력 추가 |

---

## 7. 마일스톤 (Milestones)

1. **PRD 확정**: `docs/v5/prd.md` 작성 ✅
2. **Config 확장**: 4개 필드 및 플래그 추가
3. **DDL 로직 구현**: `ddl.go`에 Sequence/Index 메타조회·변환 함수 구현
4. **마이그레이션 연동**: `migration.go` 출력 순서 통합
5. **WebSocket 확장**: `ddl_progress` 메시지 타입 추가
6. **Web UI 확장**: 체크박스·입력 필드 추가, 조건부 활성화 로직
7. **테스트**: 단위 테스트(메타조회 mock, DDL 생성 검증) + 통합 테스트
8. **최종 리뷰 및 문서 정리**

---

## 8. 향후 확장 고려사항 (Future Considerations)

- **Foreign Key** 마이그레이션 (`--with-fk`): 참조 무결성 제약 변환
- **Trigger** 마이그레이션 (`--with-triggers`): Oracle PL/SQL → PostgreSQL PL/pgSQL 변환 (복잡도 높음)
- **View** 마이그레이션 (`--with-views`): 연관 뷰 DDL 변환
- **파티션** 지원: Oracle Range/List 파티션 → PostgreSQL 파티션 테이블 변환
