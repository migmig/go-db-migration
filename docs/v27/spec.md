# 기술 명세서 - v27 (오류 수정 및 테스트 커버리지 +15% 향상 계획)

## 1. 개요
본 명세서는 전체 테스트 커버리지 수집 과정에서 발생하는 `covdata` 오류를 해소하고, 주요 핵심 패키지에 대한 테스트 보강을 통해 커버리지를 56.5%에서 71.5% 이상으로 +15%p 끌어올리기 위한 기술적 접근 방안을 정의한다.

## 2. coverage 수집 오류 해소
### 2.1 `internal/bus` 테스트 보강
- `internal/bus` 패키지 내 이벤트 publish/subscribe 기본 동작을 검증하는 단위 테스트(`bus_test.go`)를 확인/추가하여, Go 툴체인이 커버리지 프로파일을 비정상 종료 없이 생성하도록 한다.
- 테스트 파일 부재 또는 빈 패키지로 인해 발생하는 `go: no such tool "covdata"` 에러 경로를 원천 차단한다.

## 3. 핵심 모듈 테스트 범위 확대
### 3.1 `dbmigrator` (마이그레이터 진입) 패키지
- 메인 엔트리 포인트 및 주요 분기(정상, 오류, 엣지 케이스)에 대한 테스트 보강.
- 인자/오류 경로 테스트 커버리지를 높인다.

### 3.2 `internal/dialect` 패키지
- 각 데이터베이스 방언(MySQL, PostgreSQL, MSSQL, Oracle, SQLite 등)별 파서 및 분기 로직 테스트를 확대한다.

### 3.3 `internal/web` 및 `internal/web/ws` 패키지
- HTTP 핸들러 및 WebSocket 이벤트 스트리밍의 예외 케이스(연결 끊김, 잘못된 페이로드 등) 모의(Mock) 테스트 추가.

## 4. 커버리지 게이트 설정 및 표준화
### 4.1 측정 커맨드 표준화
- 로컬 및 CI 환경에서 `go test ./... -coverprofile=coverage.out` 실행을 표준화한다.
- `go tool cover -func=coverage.out` 결과 수집을 확인한다.

### 4.2 CI/CD 파이프라인 연동 (단계별)
- 점진적 커버리지 게이트웨이 도입: Phase 1 (60%) -> Phase 2 (66%) -> Phase 3 (71.5% 이상)
