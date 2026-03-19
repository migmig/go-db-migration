# 기술 사양서 (Technical Specifications) - v17

## 1. 아키텍처 개요
v17은 v16의 마이그레이션 파이프라인을 확장하여 DDL 처리 대상을 `tables` 그룹과 `sequences` 그룹으로 분리한다.
핵심 목표는 다음 3가지다.

1. 조회(Discovery) 결과의 객체 유형 분리
2. 실행(Execute) 단계의 그룹 단위 선택/재시도
3. Dry-run/리포트/로그의 그룹별 관찰성 강화

---

## 2. 도메인 모델 및 분류 규칙

### 2.1 ObjectGroup 정의
- `all`: 전체(기본값)
- `tables`: 테이블 계열
- `sequences`: 시퀀스 계열

### 2.2 분류 규칙
- `tables`
  - table DDL
  - pk/fk/uk/check constraint DDL
  - index DDL
  - 테이블 종속 부가 DDL
- `sequences`
  - sequence 생성 DDL
  - increment/start/cache/order 관련 DDL

### 2.3 내부 데이터 구조(안)
```go
type ObjectGroup string

const (
    ObjectGroupAll       ObjectGroup = "all"
    ObjectGroupTables    ObjectGroup = "tables"
    ObjectGroupSequences ObjectGroup = "sequences"
)

type GroupedScripts struct {
    TablesSQL    []string
    SequencesSQL []string
}

type GroupedStats struct {
    Tables    GroupStats
    Sequences GroupStats
}
```

---

## 3. 컴포넌트 설계

### 3.1 Migration 파이프라인 변경
- 스키마 조회 결과를 `GroupedMetadata` 형태로 저장
- SQL 생성 단계에서 `GroupedScripts` 생성
- 실행기(Executor)는 전달받은 `ObjectGroup`에 따라 수행 대상을 제한

### 3.2 실행 순서 정책
- `all`: `tables -> sequences` 고정
- `tables`: tables 그룹만 실행
- `sequences`: sequences 그룹만 실행

### 3.3 재시도 정책
- 실패 이력에 `object_group` 필드를 저장
- 재시도 시 `retry_group` 미지정이면 원래 그룹 유지
- `all` 실패 후 `sequences`만 선택 재시도 가능

---

## 4. 인터페이스 계약

### 4.1 CLI 플래그
- 신규 플래그(안): `--object-group`
  - 허용값: `all|tables|sequences`
  - 기본값: `all`
- 기존 명령/플래그와 호환되도록 미지정 시 기존 동작 유지

### 4.2 Web UI
- 실행 옵션에 `마이그레이션 대상` 선택 UI 추가
  - 전체
  - 테이블 계열만
  - 시퀀스만
- 조회 결과 패널에서 그룹별 카운트/목록을 분리 표시

### 4.3 API 스키마 확장(안)
- `POST /api/migrate`
  - 요청 필드: `object_group`(optional, default=`all`)
- `POST /api/migrate/retry`
  - 요청 필드: `object_group`(optional)
- 응답/이력에 그룹별 통계 포함
  - `stats.tables.success|failed|skipped`
  - `stats.sequences.success|failed|skipped`

---

## 5. Dry-run / 리포트 / 로깅

### 5.1 Dry-run 출력 규칙
- 출력 섹션을 `TABLES SQL` / `SEQUENCES SQL`로 분리
- 그룹별 SQL 건수 및 예상 영향도 표시

### 5.2 실행 리포트
- 그룹별 처리 건수(success/failed/skipped) 집계
- 최종 요약은 전체 + 그룹별 상세를 함께 표기

### 5.3 구조화 로그
- 모든 주요 이벤트에 `object_group` 필드 포함
- 예시 이벤트
  - `discovery.completed`
  - `script.generated`
  - `migration.started`
  - `migration.statement.failed`
  - `migration.completed`

---

## 6. 호환성 및 마이그레이션 전략

### 6.1 하위 호환성
- `object_group` 미지정 시 기존과 동일한 전체 실행
- 기존 이력 데이터(그룹 필드 없음) 조회 시 `all`로 간주

### 6.2 단계적 적용
1. 내부 파이프라인 그룹 분리(기능 플래그 off)
2. CLI/API에 `object_group` 노출
3. UI 분리 노출 및 리포트 강화
4. 회귀 테스트 완료 후 기본 활성화

---

## 7. 오류 처리/검증

### 7.1 입력 검증
- 허용되지 않은 `object_group` 값은 400 에러
- 에러 메시지에 허용값 명시

### 7.2 의존성 경고
- `sequences-only` 실행 시 대상 시퀀스의 참조 테이블 미존재 가능성 경고 로그 출력
- dry-run 단계에서 경고 목록에 포함

### 7.3 실패 격리
- 한 그룹 실패가 다른 그룹 실행 여부에 영향을 주는 정책을 명확화
  - 기본: `all`에서 tables 실패 시 sequences 단계 진입하지 않음
  - 옵션 정책은 후속 버전에서 확장

---

## 8. 테스트 전략

### 8.1 단위 테스트
- 분류기(Classifier)
  - table/constraint/index/sequence 분류 정확성
- 실행 선택기(Selector)
  - `all|tables|sequences`별 실행 목록 검증

### 8.2 통합 테스트
- `tables-only` 실행 시 sequence SQL 미실행 보장
- `sequences-only` 실행 시 table SQL 미실행 보장
- `all` 실행 시 순서 `tables -> sequences` 보장

### 8.3 회귀 테스트
- `object_group` 미지정 경로에서 기존 결과 동일성 검증
- 기존 이력 replay 시 동작 호환성 검증

---

## 9. 운영 관측 지표
- 실행 모드 사용률: `all/tables/sequences`
- 그룹별 실패율 및 재시도 성공률
- `sequences-only` 복구 시나리오 MTTR
- Dry-run 검토 소요 시간 변화

---

## 10. 결정 필요사항
- `--object-group` 네이밍 확정 여부(`--target-group` 대안)
- `all` 모드에서 tables 실패 시 sequences 계속 진행 옵션 제공 여부
- 유형별 SQL 아티팩트(`tables.sql`, `sequences.sql`) 기본 보관 정책
