# Product Requirements Document (PRD) - v13

## 1. 개요 (Overview)
현재 DB 마이그레이션 도구는 `-completion` 플래그를 통해 다양한 쉘(bash, zsh, fish, powershell)에 대한 자동완성 스크립트를 제공하고 있습니다. 하지만 사용자가 쉘 종류를 명시하지 않고 `-completion` 플래그만 단독으로 입력할 경우, Go 표준 `flag` 패키지의 기본 에러 메시지("flag needs an argument")가 출력되어 사용자 경험이 다소 불친절한 문제가 있습니다. 
본 기능 개선은 `-completion` 플래그만 입력되었을 때, **현재 사용 중인 쉘 환경 변수(`$SHELL`)를 자동으로 감지하여 해당 쉘에 맞는 스크립트를 출력**하고, 만약 감지할 수 없거나 지원하지 않는 쉘인 경우에는 올바른 사용법을 친절하게 안내하도록 개선하는 것을 목표로 합니다.

## 2. 문제점 (Problem Statement)
- 사용자가 `./dbmigrator -completion`을 실행하면 `flag needs an argument: -completion` 이라는 단순 에러만 발생합니다.
- 사용자는 자신의 쉘 이름(bash, zsh 등)을 매번 명시적으로 입력해야 하는 번거로움이 있습니다.

## 3. 목표 (Goals)
- `-completion` 플래그만 입력 시, `$SHELL` 환경 변수를 통해 현재 쉘을 자동으로 감지하여 해당 자동완성 스크립트를 출력합니다.
- 자동 감지가 불가능하거나 지원되지 않는 쉘일 경우, 사용법(예시 및 지원 쉘 목록)을 화면에 출력합니다.
- 기존의 정상적인 사용법(`-completion=bash`, `-completion zsh` 등)은 아무런 영향 없이 기존대로 동작해야 합니다.

## 4. 상세 요구사항 (Detailed Requirements)
1. **현재 쉘 자동 감지 및 스크립트 출력**
   - `-completion` 플래그에 인자가 주어지지 않은 상태로 프로그램이 실행될 경우, `os.Getenv("SHELL")` 등을 통해 쉘을 감지합니다.
   - 감지된 쉘 이름(예: `bash`, `zsh`, `fish`, `powershell`)이 지원하는 쉘 목록에 포함되어 있다면 해당 스크립트를 출력하고 프로그램이 정상(0) 종료됩니다.
2. **사용자 친화적 안내 메시지 출력**
   - 쉘을 감지하지 못했거나 지원하지 않는 쉘일 경우, 아래와 같은 형태의 안내 메시지를 표준 출력(혹은 표준 에러)으로 제공합니다.
     ```text
     자동 감지된 쉘이 지원되지 않거나 알 수 없습니다.

     사용법:
       -completion <shell>

     지원하는 쉘(shell):
       bash, zsh, fish, powershell

     사용 예시:
       ./dbmigrator -completion bash > /etc/bash_completion.d/dbmigrator
       ./dbmigrator -completion zsh > ~/.zsh/completions/_dbmigrator
     ```
3. **기존 동작 유지**
   - 지원하는 쉘 이름이 정상적으로 주어졌을 때는 안내 메시지 없이 기존처럼 해당 쉘의 자동완성 스크립트만 출력하고 종료되어야 합니다.
4. **오류 처리 및 종료 코드**
   - 인자 없이 `-completion`만 입력되고 자동 감지마저 실패하여 사용법을 출력한 경우, 프로그램은 적절한 에러 종료 코드(예: `1`)와 함께 종료되어야 합니다.

## 5. 비기능 요구사항 (Non-functional Requirements)
- Go 표준 라이브러리의 `flag` 패키지 한계를 극복하기 위해 `os.Args`를 사전 검사하는 등 부작용이 없는 깔끔한 방식으로 구현되어야 합니다.
- 기존의 테스트 코드들을 깨뜨리지 않아야 하며, 새로운 동작에 대한 단위 테스트가 추가되어야 합니다.