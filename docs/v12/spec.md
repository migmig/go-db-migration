# Go DB Migration v12 Technical Spec

## 설계
- `internal/config/config.go`에 `CompletionShell` 필드를 추가한다.
- `generateCompletionScript(shell string)` 헬퍼로 쉘별 스크립트를 생성한다.
- `ParseFlags()`에서 `-completion`이 설정되면 스크립트를 stdout으로 출력하고 조기 반환한다.

## 호환성
- 기존 플래그 동작은 유지한다.
- completion 모드 외에는 기존 필수 플래그 검증 로직을 그대로 적용한다.

## 테스트
- `-completion=bash` 입력 시 필수 플래그 없이도 성공하는지 검증.
- completion 출력에 핵심 스니펫이 포함되는지 검증.
- 미지원 쉘 입력 시 에러를 반환하는지 검증.
