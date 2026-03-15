# 작업 목록 (Tasks) - v13

## 목표: `-completion` 플래그 단독 사용성 개선 및 쉘 자동 감지

### 1. 설계 (Design & Documentation)
- [x] `docs/v13/prd.md` 작성 (요구사항 문서)
- [x] `docs/v13/spec.md` 작성 (기술 사양서)
- [x] `docs/v13/tasks.md` 작성 (작업 목록)

### 2. 구현 (Implementation)
- [x] `internal/config/config.go`에 `detectShell()` 헬퍼 함수 추가 (`$SHELL` 환경 변수 기반 감지)
- [x] `internal/config/config.go`에 `printCompletionUsage()` 헬퍼 함수 추가
  - [x] `bash, zsh, fish, powershell`에 대한 사용법 및 자동 감지 실패 안내 텍스트 작성
- [x] `internal/config/config.go`의 `ParseFlags()` 함수 시작 부분에 `os.Args` 사전 검사 로직 추가
  - [x] `-completion`이 인자 없이 단독으로 쓰인 경우 `detectShell()` 호출
  - [x] 쉘이 정상적으로 감지되면 `generateCompletionScript()`를 통해 출력 후 정상 종료(`os.Exit(0)`)
  - [x] 쉘을 감지하지 못하면 `printCompletionUsage()` 호출 후 에러 종료(`os.Exit(1)`)

### 3. 테스트 (Testing)
- [x] `internal/config/config_test.go`에 새로운 시나리오 테스트 추가
  - [x] `-completion` 단독 실행 및 `$SHELL` 환경 변수에 따른 정상 스크립트 반환 검증 (`os.Exit` 문제 우회를 위한 테스트 로직 구성)
  - [x] `-completion` 단독 실행 시 `$SHELL` 감지 실패할 경우 사용법 텍스트 출력 검증
- [x] 기존 통합/단위 테스트(`.go` 테스트들) 실행 및 통과 여부 확인 (`go test ./...`)

### 4. 문서 업데이트 (Documentation)
- [x] `README.md` 내용 수정
  - [x] 기능 설명 부분에 `-completion` 단독 실행 시 현재 쉘 자동 감지 기능이 지원된다는 문구 추가