# 작업 목록 (Tasks) - v17

## 목표: 테이블 계열/시퀀스 계열 분리 조회 + 그룹 단위 실행/재시도 + 리포트 분리

### 1. 설계/문서화 (Design & Documentation)
- [x] `docs/v17/prd.md` 작성
- [x] `docs/v17/spec.md` 작성
- [x] `docs/v17/tasks.md` 작성
- [x] `README.md` 업데이트
  - [x] 객체 그룹 실행 모드(`all/tables/sequences`) 설명 추가
  - [x] 신규 옵션/요청 필드(`--object-group`, `object_group`) 문서화
  - [ ] Dry-run/리포트의 그룹별 출력 예시 추가

### 2. 도메인 모델/분류기 구현 (Domain & Classifier)
- [x] `ObjectGroup` 타입/상수 도입
  - [x] `all`, `tables`, `sequences` 정의
  - [x] 기본값 `all` 처리 경로 반영
- [ ] DDL 분류기 구현
  - [ ] table/constraint/index 등 `tables` 분류
  - [ ] sequence 관련 DDL `sequences` 분류
  - [ ] 분류 불가/애매 케이스 경고 로깅
- [ ] 그룹별 컨테이너 구조 도입
  - [ ] `GroupedMetadata`(조회 결과)
  - [ ] `GroupedScripts`(생성 SQL)
  - [ ] `GroupedStats`(집계 결과)

### 3. 조회/스크립트 생성 파이프라인 분리 (Discovery & Script Generation)
- [ ] 메타데이터 조회 결과를 그룹별로 분리 저장
- [ ] SQL 생성 시 그룹별 산출물 분리
  - [ ] `tables.sql` 성격의 SQL 묶음
  - [ ] `sequences.sql` 성격의 SQL 묶음
- [ ] Dry-run 출력 섹션 분리
  - [ ] `TABLES SQL`
  - [ ] `SEQUENCES SQL`

### 4. 실행기(Executor) 그룹 선택 로직 (Execution)
- [x] 실행 입력에 `object_group` 반영
- [ ] 실행 모드별 대상 선택 구현
  - [ ] `all`: `tables -> sequences` 순서 고정 
  - [ ] `tables`: tables만 실행
  - [ ] `sequences`: sequences만 실행
- [ ] 실패 격리 정책 반영
  - [ ] 기본 정책: `all`에서 tables 실패 시 sequences 미진입
  - [ ] 정책 이벤트 구조화 로그로 기록

### 5. API/웹 서버 계약 확장 (API Contract)
- [x] `POST /api/migrate` 요청 바디에 `object_group`(optional) 지원
- [x] `POST /api/migrate/retry` 요청 바디에 `object_group`(optional) 지원
- [x] 입력 검증
  - [x] 허용값 외 입력 시 400 반환
  - [x] 허용값 안내 메시지 포함
- [ ] 응답/이력 모델 확장
  - [ ] 그룹별 통계(`stats.tables`, `stats.sequences`) 노출
  - [ ] 이력에 `object_group` 저장(미존재 legacy는 `all`로 해석)

### 6. Web UI 반영 (Frontend)
- [ ] 실행 옵션에 `마이그레이션 대상` 선택 UI 추가
  - [ ] 전체
  - [ ] 테이블 계열만
  - [ ] 시퀀스만
- [ ] 조회 결과 패널 그룹 분리
  - [ ] 그룹별 카운트 표시
  - [ ] 그룹별 목록/접기-펼치기 UX 정리
- [ ] Dry-run/결과 요약 카드 분리
  - [ ] 그룹별 성공/실패/스킵 통계 표시

### 7. 로그/리포트/관측성 강화 (Observability)
- [ ] 주요 이벤트 로그에 `object_group` 필드 추가
  - [ ] `discovery.completed`
  - [ ] `script.generated`
  - [ ] `migration.started`
  - [ ] `migration.statement.failed`
  - [ ] `migration.completed`
- [ ] 최종 리포트 포맷 확장
  - [ ] 전체 요약 + 그룹별 상세 병기
- [ ] 운영 지표 수집 항목 추가
  - [ ] 모드별 사용률(`all/tables/sequences`)
  - [ ] 그룹별 실패율/재시도 성공률

### 8. 테스트 (Testing)
- [ ] 단위 테스트
  - [ ] 분류기: table/constraint/index/sequence 분류 정확성
  - [ ] 선택기: `all|tables|sequences`별 실행 목록 검증
  - [ ] 검증기: 잘못된 `object_group` 입력 검증
- [ ] 통합 테스트
  - [ ] `tables-only`에서 sequence SQL 미실행 검증
  - [ ] `sequences-only`에서 table SQL 미실행 검증
  - [ ] `all` 순서(`tables -> sequences`) 검증
- [ ] 회귀 테스트
  - [ ] 옵션 미지정(`all` 기본) 기존 동작 동일성
  - [ ] 기존 이력 replay 호환성
  - [ ] `go test ./...` 통과

### 9. 롤아웃/운영 가이드 (Rollout)
- [ ] 단계적 릴리즈 계획 반영
  - [ ] 내부 기능 플래그 기반 점진 활성화
  - [ ] 운영팀 대상 모드 선택 가이드 배포
- [ ] 장애 대응 플레이북 보강
  - [ ] `sequences-only` 복구 절차 문서화
  - [ ] 모드 오선택 방지 체크리스트 추가
- [ ] 최종 배포 체크
  - [ ] dry-run 검토 절차 준수 확인
  - [ ] 로그/리포트 대시보드 필드 누락 점검
