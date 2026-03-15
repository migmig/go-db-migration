# 기술 사양서 (Technical Specifications) - v13

## 1. 아키텍처 및 구현 방향
Go 표준 라이브러리의 `flag` 패키지는 `StringVar` 플래그에 인자가 주어지지 않았을 때 에러 메시지를 출력하고 즉시 `os.Exit(2)`를 호출합니다. 이를 방지하고 사용자 친화적인 메시지를 제공 및 쉘을 자동 감지하기 위해서는 `flag.Parse()`가 실행되기 전에 `os.Args`를 전처리(pre-process)하는 방법이 가장 직관적이고 안정적입니다.

### 1.1. `os.Args` 전처리 및 `$SHELL` 자동 감지 방식 도입
`internal/config/config.go`의 `ParseFlags()` 함수 시작 부분에 인자 검사 로직을 추가합니다.

- 사용자가 입력한 명령어 인자(`os.Args[1:]`)를 순회합니다.
- `-completion` 인자가 발견되었을 때, 다음 조건 중 하나에 해당하면(단독으로 쓰인 경우) 쉘 감지를 시도합니다:
  1. `-completion`이 마지막 인자인 경우.
  2. `-completion` 바로 다음 인자가 `-` 문자로 시작하는 경우 (즉, 다른 플래그인 경우).

**쉘 자동 감지 로직:**
- `os.Getenv("SHELL")`을 통해 현재 쉘 경로를 가져옵니다.
- 경로에 `bash`, `zsh`, `fish`, `pwsh` (또는 `powershell`) 문자열이 포함되어 있는지 확인합니다.
- 매칭되는 쉘이 지원 목록(`bash, zsh, fish, powershell`)에 있다면, `generateCompletionScript()`를 호출하여 스크립트를 출력하고 프로그램 실행을 정상 종료(`os.Exit(0)`)합니다.
- 매칭되지 않거나 빈 값이라면 사용법을 출력하고 프로그램 실행을 에러 상태로 중단(`os.Exit(1)`)합니다.

### 1.2. 사용법 출력 함수 추가
`printCompletionUsage()` 헬퍼 함수를 추가하여 표준 출력(혹은 표준 에러)에 아래 정보를 제공합니다.
- 올바른 사용법
- 지원 가능한 쉘 목록 (bash, zsh, fish, powershell)
- 간단한 실행 예시

### 1.3. 코드 구현 (예시)
```go
func detectShell() string {
    shellEnv := strings.ToLower(os.Getenv("SHELL"))
    if strings.Contains(shellEnv, "bash") { return "bash" }
    if strings.Contains(shellEnv, "zsh") { return "zsh" }
    if strings.Contains(shellEnv, "fish") { return "fish" }
    if strings.Contains(shellEnv, "pwsh") || strings.Contains(shellEnv, "powershell") { return "powershell" }
    return ""
}

func ParseFlags() (*Config, error) {
    // flag.Parse() 호출 전, -completion 단독 사용 예외 처리
    args := os.Args[1:]
    for i, arg := range args {
        if arg == "-completion" || arg == "--completion" {
            // 마지막 인자이거나 다음 인자가 또 다른 플래그일 때 (인자 없음)
            if i+1 == len(args) || strings.HasPrefix(args[i+1], "-") {
                detected := detectShell()
                if detected != "" {
                    script, _ := generateCompletionScript(detected)
                    fmt.Println(script)
                    os.Exit(0)
                } else {
                    printCompletionUsage()
                    os.Exit(1)
                }
            }
        }
    }

    // 기존 flag 초기화 및 파싱 로직
    // ...
}
```

## 2. 테스트 방안
- **통합/단위 테스트 (Unit Test):** 
  - `config_test.go` 파일 내에서, `os.Args`를 임의로 조작(`[]string{"cmd", "-completion"}`)하고 임시로 `os.Setenv("SHELL", "/bin/zsh")`를 설정한 뒤 `ParseFlags()`(또는 분리된 검증 로직)를 호출했을 때, 기존처럼 `flag needs an argument` 패닉/에러가 아니라 해당 쉘의 스크립트가 잘 반환(또는 출력)되는지 검증합니다.
  - 지원하지 않는 쉘의 경우 적절한 사용법 안내 텍스트가 출력되는지 확인합니다.
  - (주의: `os.Exit`을 피하기 위해, 테스트가 용이하도록 `ParseFlags()` 내부 로직을 리팩토링하여 에러나 값을 반환하게 하는 것이 필요할 수 있습니다.)
- **기존 테스트 호환성:** 기존에 `-completion=bash` 등의 정상 케이스에 대한 테스트가 깨지지 않는지 확인합니다.