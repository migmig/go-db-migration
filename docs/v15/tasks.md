# 작업 목록 (Tasks) - v15

## 목표: 인증 기반 멀티유저 Web UI + 사용자별 접속정보/이력 분리 + 관리자 CLI

### 1. 설계/문서화 (Design & Documentation)
- [x] `docs/v15/prd.md` 작성 및 보완
- [x] `docs/v15/spec.md` 작성
- [x] `docs/v15/tasks.md` 작성
- [x] `README.md` 기능 섹션/플래그/운영 가이드 업데이트
  - [x] 인증 모드 개요 및 로그인 흐름 추가
  - [x] 관리자 CLI(`users list/add/reset-password/delete`) 사용 예시 추가
  - [x] 비밀키(`DBM_MASTER_KEY`) 설정 가이드 추가

### 2. DB 스키마/저장소 구현 (Database & Repository)
- [ ] `internal/db/db.go`
  - [ ] `users` 테이블 생성/마이그레이션 추가
  - [ ] `db_credentials` 테이블 생성/마이그레이션 추가
  - [ ] `migration_history.user_id` 스키마 확장 및 인덱스 추가
  - [ ] 업그레이드 백필 로직(user_id 매핑) 구현
- [ ] 사용자 저장소 메서드 구현
  - [ ] `CreateUser`, `GetUserByUsername`, `ListUsers`, `DeleteUser`, `ResetPassword`
- [ ] 접속정보 저장소 메서드 구현
  - [ ] `CreateCredential`, `ListCredentialsByUser`, `UpdateCredential`, `DeleteCredential`
- [ ] 이력 저장소 메서드 구현
  - [ ] `InsertHistory(userID, ...)`, `ListHistoryByUser(page,pageSize)`, `GetHistoryByID(userID,id)`

### 3. 보안 유틸 구현 (Security)
- [x] 비밀번호 해시 유틸 추가
  - [x] bcrypt 해시/검증 함수
  - [x] 최소 길이/정책 검증
- [x] Credential 비밀번호 암복호화 유틸 추가
  - [x] AES-GCM 암호화/복호화 구현
  - [x] nonce/포맷 직렬화 규칙 정의
  - [x] 키 누락 시 fail-fast 처리
- [x] 민감정보 로깅 차단
  - [x] 구조화 로그 필드 점검(원문 비밀번호/암호문 미노출)

### 4. 인증/인가 서버 구현 (Web Backend)
- [ ] `internal/web/server.go`
  - [ ] `POST /api/auth/login`
  - [ ] `POST /api/auth/logout`
  - [ ] `GET /api/auth/me`
  - [ ] 인증 미들웨어(`requireAuth`) 적용
- [ ] 보호 API 사용자 스코프 강제
  - [ ] credentials CRUD에서 `user_id` 소유권 체크
  - [ ] history 조회/상세/재실행에서 `user_id` 소유권 체크
- [ ] 세션 정책 적용
  - [ ] 쿠키(HttpOnly, SameSite, Secure) 설정
  - [ ] idle/absolute timeout 처리

### 5. 마이그레이션 실행 경로 연동 (Migration Flow)
- [ ] `internal/migration` 연동
  - [ ] 실행 컨텍스트에 `user_id` 주입
  - [ ] 실행 완료 시 사용자 귀속 이력 저장
  - [ ] retry/replay 경로의 사용자 권한 검증

### 6. 프론트엔드 구현 (Web UI)
- [ ] `internal/web/templates/index.html`
  - [ ] 로그인 화면/폼 추가
  - [ ] 인증 후 GNB(내 정보/내 작업 내역/로그아웃) 추가
  - [ ] 저장된 접속정보 불러오기 UI(드롭다운/모달) 추가
  - [ ] 내 작업 내역 섹션(목록/상태/재실행 버튼) 추가
- [ ] `internal/web/templates/chart.js`
  - [ ] 인증 상태 확인/세션 만료 처리
  - [ ] credentials API 연동(조회/생성/수정/삭제/불러오기)
  - [ ] history API 연동(페이지네이션/상세/재실행)
  - [ ] 로그인/로그아웃 이벤트 핸들링

### 7. 관리자 CLI 구현 (Admin Shell)
- [ ] `main.go` 커맨드 엔트리 추가
  - [ ] `users list`
  - [ ] `users add <username> <password>`
  - [ ] `users reset-password <username> <newPassword>`
  - [ ] `users delete <username>`
- [ ] 에러 처리/종료코드 표준화
  - [ ] 성공 시 0, 실패 시 non-zero
  - [ ] 오류 메시지 표준에러 출력

### 8. 테스트 (Testing)
- [ ] 단위 테스트
  - [x] bcrypt 해시/검증
  - [x] AES-GCM 암복호화/키 오류
  - [ ] user_id 스코프 필터(권한) 검증
- [ ] 통합 테스트
  - [ ] 로그인/로그아웃/세션 만료 시나리오
  - [ ] 사용자 A/B 데이터 격리 검증(credentials/history)
  - [ ] history pagination/replay 검증
  - [ ] 관리자 CLI 명령 정상/예외 케이스
- [ ] 회귀 테스트
  - [ ] `go test ./...` 실행

### 9. 운영/릴리즈 준비 (Rollout)
- [ ] 배포 전 체크리스트
  - [ ] `DBM_MASTER_KEY` 주입 확인
  - [ ] 초기 관리자 계정 생성/비밀번호 변경 절차 점검
  - [x] 인증 기능 플래그(`auth-enabled`) 기본값 확정 (`false`)
- [ ] 모니터링 지표 추가
  - [ ] 로그인 실패율, 세션 만료율
  - [ ] credentials/history API 오류율
