# 작업 목록 (Tasks) - v14

## 목표: Web UI 최근 입력 자동완성 + 상단 DB URL/ID/PASS 영속 복원

### 1. 설계/문서화 (Design & Documentation)
- [x] `docs/v14/prd.md` 작성 및 보완
- [x] `docs/v14/spec.md` 작성
- [x] `docs/v14/tasks.md` 작성

### 2. 프론트엔드 구현 (Implementation - Web UI)
- [x] `internal/web/templates/index.html`
  - [x] 상단 공통 연결 정보 섹션(DB URL/ID/PASS, 비밀번호 기억, 초기화 버튼) 추가
  - [x] 자동완성 제안 UI 컨테이너(또는 datalist) 배치
  - [x] 필드별 이력 삭제/전체 삭제 액션 UI 추가
- [x] `internal/web/templates/chart.js`
  - [x] localStorage 키/모델 정의 (`history:*`, `sharedConnection`)
  - [x] 필드별 최근 이력 save/load/clear 유틸 구현 (중복 제거, 최근순, N개 제한)
  - [x] 민감 필드 제외 로직 구현 (`password/token/secret`)
  - [x] 상단 공통 DB URL/ID/PASS 복원 및 저장 로직 구현
  - [x] PASS 저장 옵트인(`rememberPass`) 분기 처리
  - [x] 자동완성 노출/선택/키보드 내비게이션 구현
  - [x] localStorage 예외 처리(graceful fallback) 추가

### 3. 백엔드/호환성 점검 (Compatibility)
- [x] `internal/web/server.go` 및 기존 API 계약 점검
  - [x] 프론트 변경 후도 기존 요청 JSON 필드와 바인딩이 깨지지 않는지 확인
  - [x] 세션/웹소켓/다운로드 흐름 영향 없음 확인

### 4. 테스트 (Testing)
- [x] 단위 테스트
  - [x] 이력 정렬/중복 제거/개수 제한 검증
  - [x] 민감정보 미저장 및 rememberPass 분기 검증
- [x] UI/통합 테스트
  - [x] 새로고침/재접속 시 상단 DB URL/ID 복원 검증
  - [x] rememberPass true 시 PASS 복원, false 시 미복원 검증
  - [x] 자동완성 표시/선택/삭제 시나리오 검증
  - [x] localStorage 미지원 환경 fallback 검증
- [x] 회귀 테스트
  - [x] `go test ./...` 실행

### 5. 문서 업데이트 (Documentation)
- [x] `README.md` 업데이트
  - [x] Web UI 입력 이력 자동완성 기능 설명 추가
  - [x] 상단 DB URL/ID/PASS 복원 및 비밀번호 기억(옵트인) 동작 안내
  - [x] 개인정보/보안 주의사항(공용 PC) 명시
