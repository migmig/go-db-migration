# 작업 목록 (Tasks) - v19

## 목표: Direct Insert 사전 행 수 비교 기반 전송 필요성 판단 체계 도입

### 1. 문서화
- [x] `docs/v19/prd.md` 작성
- [x] `docs/v19/spec.md` 작성
- [x] `docs/v19/tasks.md` 작성

### 2. 백엔드: pre-check 엔진
- [x] 테이블별 source/target COUNT 수집 모듈 추가 (`internal/migration/precheck_engine.go`)
- [x] decision 판정 로직 구현 (`transfer_required`, `skip_candidate`, `count_check_failed`)
- [x] policy 적용기 구현 (`strict`, `best_effort`, `skip_equal_rows`)
- [x] 병렬 처리/타임아웃/재시도(필요시) 설정 반영 (worker pool, context timeout)

### 3. 백엔드: API/연계
- [x] `POST /api/migrations/precheck` 구현
- [x] pre-check 결과 필터 조회 API 구현 (`GET /api/migrations/precheck/results`)
- [x] 기존 마이그레이션 실행 API와 `use_precheck` 연계 (`usePrecheckResults` 파라미터)
- [x] 입력 검증(정책/필터 enum) 및 에러 코드 표준화

### 4. 프론트엔드: pre-check UX
- [x] pre-check 실행 버튼 및 요약 카드 추가
- [x] 결과 테이블 컬럼(원본/대상/차이/판정) 추가
- [x] decision 필터 탭 추가 (`all`, `transfer_required`, `skip_candidate`, `count_check_failed`)
- [x] 실패 항목 reason 툴팁/경고 표시

### 5. CLI
- [x] `--precheck-row-count` 플래그 추가
- [x] `--precheck-policy` 플래그 추가
- [x] `--precheck-filter` 플래그 추가
- [x] dry-run 출력 포맷 확장

### 6. 관측성
- [x] 구조화 로그 필드 추가 (`table_name`, `decision`, `policy`, `reason` 등)
- [x] pre-check 관련 메트릭 추가 (`precheckRunTotal`, `precheckTablesTotal` 등)
- [ ] 모니터링 대시보드/알림 규칙 업데이트

### 7. 테스트
- [x] 단위 테스트: decision/policy 판정
- [ ] 통합 테스트: pre-check API 및 필터링
- [ ] 실행 연계 테스트: pre-check 기반 실제 전송 대상 축소
- [ ] 성능 테스트: 대량 테이블 pre-check 처리 시간

### 8. 운영 가이드/릴리즈
- [x] README에 신규 플래그 및 pre-check 모드 설명 반영
- [x] feature flag(`DBM_V19_PRECHECK`)로 점진 배포
- [ ] 운영자 가이드(정책 선택 기준, 실패 대응) 추가
