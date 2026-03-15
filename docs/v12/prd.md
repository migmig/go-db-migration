# Go DB Migration v12 PRD

## 배경
CLI 사용자가 매번 긴 플래그를 기억해야 해서 사용성이 떨어집니다.

## 목표
- Bash, Zsh, Fish, PowerShell 환경에서 dbmigrator 플래그 자동완성을 제공한다.
- 기존 마이그레이션 실행 경로와 하위 호환성을 유지한다.

## 요구사항
1. `-completion <shell>` 플래그를 제공한다.
2. 지원 쉘: `bash`, `zsh`, `fish`, `powershell`.
3. 자동완성 모드에서는 DB 접속 필수 플래그 검증을 건너뛴다.
4. 지원하지 않는 쉘 입력 시 명확한 에러를 반환한다.
