# 기술 사양서 (Technical Specifications) - v14

## 1. 아키텍처 및 구현 방향
v14 기능은 서버 저장 없이 브라우저 측 상태(localStorage)를 활용해 구현합니다. 핵심은 (1) 필드별 최근 입력 자동완성, (2) 상단 공통 DB URL/ID/PASS 복원, (3) 민감정보 보호(기본 비저장 + PASS 옵트인)입니다.

구현은 `internal/web/templates/index.html`(폼 구조/UI) + `internal/web/templates/chart.js`(클라이언트 로직) 조합으로 진행하며, 기존 API 스키마 및 백엔드 핸들러(`internal/web/server.go`)는 변경 없이 호환됩니다.

---

## 2. 데이터 모델 (Client-side)

### 2.1 localStorage 키 설계
- 접두사: `dbmigrator:webui:v14:`
- 필드별 최근 이력: `dbmigrator:webui:v14:history:<fieldKey>`
  - 예: `history:sourceHost`, `history:schema`, `history:oracleOwner`
- 상단 공통 접속정보: `dbmigrator:webui:v14:sharedConnection`
  - JSON 구조:
    ```json
    {
      "dbUrl": "...",
      "dbId": "...",
      "dbPass": "...",
      "rememberPass": false,
      "updatedAt": "2026-01-01T00:00:00Z"
    }
    ```

### 2.2 저장 규칙
- trim 후 빈 문자열은 저장하지 않음
- 필드별 중복값 제거 후 최신값을 배열 선두로 이동
- 필드별 최대 N개(기본 10개) 유지
- `password`, `token`, `secret` 계열 필드는 이력 저장 대상에서 제외
- 상단 `dbPass`는 `rememberPass=true`일 때만 저장

---

## 3. UI/UX 사양

### 3.1 상단 공통 연결 영역
- 배치: 페이지 상단(현재 헤더/폼 시작 영역 인접)에 `DB URL`, `DB ID`, `DB PASS`, `비밀번호 기억` 체크박스 제공
- 동작:
  1. 진입 시 `sharedConnection` 읽어서 입력값 자동 복원
  2. 사용자가 수정 시 디바운스(예: 300ms) 또는 submit 시 저장
  3. "초기화" 버튼으로 URL/ID/PASS + rememberPass 일괄 삭제

### 3.2 필드 자동완성
- 대상: 텍스트/검색형 입력 필드 중 민감정보 제외
- 노출:
  - 포커스 시 최근값 목록 노출
  - 입력 중 prefix 매칭 필터
- 선택:
  - 클릭 또는 Enter로 값 반영
- 접근성:
  - ↑/↓로 항목 이동, Enter 선택, Esc 닫기
  - blur 시 드롭다운 닫기

### 3.3 이력 삭제 UX
- 필드 단위 삭제: 입력 우측 또는 설정 메뉴에서 "최근값 지우기"
- 전체 삭제: "입력 이력 전체 삭제" 액션 제공
- 삭제 직후 UI 및 localStorage 즉시 동기화

---

## 4. 클라이언트 로직 세부

### 4.1 모듈화 함수(예시)
- `loadHistory(fieldKey): string[]`
- `saveHistory(fieldKey, value): void`
- `clearHistory(fieldKey?): void` (없으면 전체 삭제)
- `loadSharedConnection(): SharedConnection`
- `saveSharedConnection(model): void`
- `bindAutocomplete(inputEl, fieldKey, options): void`
- `isSensitiveField(fieldKey|inputName): boolean`

### 4.2 예외 처리
- `localStorage` 접근 시 `try/catch`로 감싸고 실패 시 no-op 처리
- JSON 파싱 실패 시 해당 키 삭제 후 기본값 복구
- quota 초과 시 오래된 이력부터 정리 후 재시도

### 4.3 기존 동작 호환성
- localStorage 사용 불가 시 자동완성/복원만 비활성화되고,
  - 입력
  - 테이블 조회
  - 마이그레이션 실행
  - 결과 다운로드
  의 기존 핵심 플로우는 그대로 동작해야 함

---

## 5. 보안/프라이버시 가이드
- 기본 정책: 민감정보 자동 저장 금지
- PASS 저장은 옵트인(`rememberPass=true`)일 때만 허용
- 공유 기기 사용 경고 문구를 상단 영역 근처에 배치
- 가능하면 PASS는 마스킹 상태 유지, 토글 보기 기능은 별도 검토

---

## 6. 테스트 방안

### 6.1 단위 테스트 (JS)
- 중복 제거/최신순 정렬 검증
- 최대 N개 제한 검증
- 민감필드 저장 제외 검증
- rememberPass false 시 PASS 미저장 검증

### 6.2 통합/UI 테스트
- 페이지 재로드 시 상단 URL/ID 복원 검증
- rememberPass true/false에 따른 PASS 복원 분기 검증
- 자동완성 포커스/타이핑/선택/삭제 동작 검증
- localStorage 비활성화 환경에서 graceful fallback 검증

### 6.3 회귀 테스트
- 기존 `/api/tables`, `/api/start`, `/api/retry` 요청 payload가 변경되지 않는지 확인
- 다크모드/반응형 레이아웃에서 상단 공통 영역 UI 깨짐 여부 확인

---

## 7. 단계별 롤아웃 제안
1. Phase 1: 필드별 최근 이력 자동완성(민감정보 제외)
2. Phase 2: 상단 공통 DB URL/ID/PASS + rememberPass
3. Phase 3: 삭제 UX 고도화 및 접근성 개선(키보드 내비게이션)
