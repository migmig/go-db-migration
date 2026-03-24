# 작업 항목 - v27 (오류 수정 및 테스트 커버리지 +15% 향상)

- [x] 1. `internal/bus` 패키지 단위 테스트 보강
  - [x] 1.1 `internal/bus/bus_test.go` 작성 및 Publish/Subscribe 동작 검증
  - [x] 1.2 `go test ./... -coverprofile=coverage.out` 실행 후 `covdata` 오류 해소 확인
- [x] 2. `dbmigrator` 테스트 보강
  - [x] 2.1 마이그레이션 메인 진입점/인자/오류 경로에 대한 테스트 케이스 3개(정상/에러/엣지) 추가
- [x] 3. `internal/dialect` 테스트 보강
  - [x] 3.1 각 데이터베이스 방언 파서 및 타입 매핑 테스트 확대
- [x] 4. `internal/web` 및 `internal/web/ws` 테스트 보강
  - [x] 4.1 웹 계층 핸들러 예외 케이스 및 연결 종료 모의 테스트(Mock) 추가
  - [x] 4.2 WebSocket 이벤트 스트리밍 엣지 케이스 추가
- [x] 5. 커버리지 리포트 및 파이프라인 적용
  - [x] 5.1 커버리지 측정 스크립트 기반 `go tool cover -func=coverage.out` 표준화
  - [x] 5.2 전체 커버리지 71.5% 이상 달성 여부 점검
  - [x] 5.3 CI (또는 Makefile)에 최소 커버리지 점검 게이트 추가
  - [x] 5.4 Flaky 테스트(간헐 실패) 0건 확인
