# 기술 사양서 (Technical Specifications) - v15

## 1. 아키텍처 개요
v15는 기존 단일 사용자 기반 Web UI를 **사용자 인증 기반 멀티 사용자 구조**로 확장한다. 핵심 목표는 다음 4가지다.

1. 로그인/로그아웃 및 세션 검증
2. 사용자별 DB 접속정보(Credential) 안전 저장/조회
3. 사용자별 마이그레이션 이력 조회/재실행
4. 관리자 CLI를 통한 계정 라이프사이클 관리

구현은 기존 구조를 최대한 유지하며 `internal/web`, `internal/db`, `internal/migration`, `internal/config`에 인증/권한 계층을 추가하는 방향으로 진행한다.

---

## 2. 컴포넌트 설계

### 2.1 서버 컴포넌트
- `internal/web/server.go`
  - 로그인/로그아웃/세션 체크 API 추가
  - 인증 미들웨어(`requireAuth`) 도입
  - 사용자별 Credential/History API 엔드포인트 추가
- `internal/db/db.go`
  - 사용자, 자격증명, 이력 테이블 생성/마이그레이션 로직 추가
  - 사용자 스코프 CRUD 메서드 추가
- `internal/migration`
  - 작업 실행 시 `user_id` 컨텍스트를 수집하여 history 저장
- `main.go`
  - Admin CLI 서브커맨드 진입점 추가 (`users list/add/reset-password/delete`)

### 2.2 클라이언트 컴포넌트
- `internal/web/templates/index.html`
  - 비인증 상태에서 로그인 UI 렌더링
  - 인증 후 GNB: 내 정보, 내 작업 내역, 로그아웃
  - 저장된 접속 정보 불러오기 드롭다운/모달
- `internal/web/templates/chart.js`
  - 인증 상태 확인 및 세션 만료 처리
  - Credential 목록 조회/적용/저장/삭제
  - My History 조회/페이지네이션/재실행 입력값 반영

---

## 3. 데이터 모델 및 스키마

### 3.1 `users`
- `id` INTEGER PK AUTOINCREMENT
- `username` TEXT UNIQUE NOT NULL
- `password_hash` TEXT NOT NULL (bcrypt)
- `is_admin` BOOLEAN NOT NULL DEFAULT 0
- `created_at` DATETIME NOT NULL
- `updated_at` DATETIME NOT NULL

인덱스:
- `ux_users_username (username)`

### 3.2 `db_credentials`
- `id` INTEGER PK AUTOINCREMENT
- `user_id` INTEGER NOT NULL FK -> `users(id)` ON DELETE CASCADE
- `alias` TEXT NOT NULL
- `db_type` TEXT NOT NULL
- `host` TEXT NOT NULL
- `port` INTEGER NULL
- `database_name` TEXT NULL
- `username` TEXT NOT NULL
- `password_enc` TEXT NOT NULL (AES-GCM 암호문 + nonce 포함 포맷)
- `created_at` DATETIME NOT NULL
- `updated_at` DATETIME NOT NULL

인덱스:
- `ix_db_credentials_user_id (user_id)`
- `ux_db_credentials_user_alias (user_id, alias)`

### 3.3 `migration_history`
- 기존 컬럼 + `user_id` INTEGER NOT NULL FK -> `users(id)` ON DELETE CASCADE
- `status` TEXT (`success`/`failed`)
- `source_summary`, `target_summary`, `options_json`, `log_summary`
- `created_at` DATETIME NOT NULL

인덱스:
- `ix_migration_history_user_created_at (user_id, created_at DESC)`

### 3.4 마이그레이션 전략
1. 신규 설치: 위 3개 테이블을 최신 스키마로 생성
2. 기존 설치 업그레이드:
   - `users` 생성 + 기본 관리자 계정 정책 적용(초기 비밀번호 강제 변경 권장)
   - `migration_history`에 `user_id` nullable로 추가 후 데이터 백필
   - 백필 이후 `NOT NULL` 제약 적용
3. 롤백: 스키마 롤백보다는 forward-only 정책 권장

---

## 4. 인증/인가 설계

### 4.1 인증 방식
- 기본: 서버 세션 쿠키 기반
- 쿠키 속성:
  - `HttpOnly=true`
  - `Secure=true` (TLS 환경)
  - `SameSite=Lax`
  - 만료: idle timeout (예: 30분) + absolute timeout (예: 24시간)

### 4.2 패스워드 정책
- 저장: bcrypt(hash cost 기본 10~12)
- 최소 길이(예: 8자 이상) 검증
- 관리자 CLI reset 시 임시 비밀번호 발급 가능

### 4.3 인가 규칙
- 모든 business API는 `requireAuth` 필수
- Credential/History 조회/수정/삭제는 `WHERE user_id = session.user_id` 강제
- Admin CLI는 웹 인증과 별개이며 로컬 실행 권한을 전제로 함

---

## 5. 암호화 및 비밀정보 처리

### 5.1 DB 비밀번호 저장
- `db_credentials.password_enc`는 AES-GCM으로 암호화 저장
- 마스터 키는 환경 변수(`DBM_MASTER_KEY`) 또는 설정 파일에서 주입
- 키가 없으면 서버 시작 실패(fail-fast)

### 5.2 메모리 처리
- 복호화된 비밀번호는 연결 직전 최소 범위에서만 사용
- 로깅/에러 메시지에 비밀번호 출력 금지

### 5.3 감사 로그
- 로그인 성공/실패, 비밀번호 초기화, 사용자 삭제는 구조화 로그로 기록
- 로그에 민감정보(원문 비밀번호, 암호문 전문) 포함 금지

---

## 6. API 사양

### 6.1 인증 API
- `POST /api/auth/login`
  - req: `{ "username": "...", "password": "..." }`
  - res: `200 { "ok": true, "user": {"id":1,"username":"..."} }`
- `POST /api/auth/logout`
  - 세션 무효화
- `GET /api/auth/me`
  - 로그인 사용자 정보 반환

### 6.2 Credential API
- `GET /api/credentials`
- `POST /api/credentials`
- `PUT /api/credentials/:id`
- `DELETE /api/credentials/:id`
- 공통: 본인 소유 데이터만 접근 가능

### 6.3 My History API
- `GET /api/history?page=1&pageSize=20`
  - 사용자 본인 이력만 반환
- `GET /api/history/:id`
  - 상세 조회 (권한 체크)
- `POST /api/history/:id/replay`
  - 이력의 설정값을 현재 입력 폼에 복원할 수 있는 payload 반환

---

## 7. UI/UX 사양

### 7.1 로그인 게이트
- 미인증 시 메인 마이그레이션 화면 대신 로그인 폼 노출
- 로그인 성공 시 기존 메인 화면 렌더
- 세션 만료 시 토스트 + 로그인 화면 리다이렉트

### 7.2 상단 네비게이션(GNB)
- 메뉴: `내 정보(접속정보)`, `내 작업 내역`, `로그아웃`
- 현재 사용자명 표시

### 7.3 접속정보 관리
- 별칭(alias) 기반 목록
- 생성/수정/삭제 모달 제공
- "불러오기" 클릭 시 소스/타겟 폼 자동 채움

### 7.4 내 작업 내역
- 열: 실행일시, 소스/타겟 요약, 결과 상태
- 페이지네이션
- "이 설정으로 다시 실행" 액션 제공

---

## 8. 관리자 CLI 사양

### 8.1 커맨드 구조
- `go-db-migration users list`
- `go-db-migration users add <username> <password>`
- `go-db-migration users reset-password <username> <newPassword>`
- `go-db-migration users delete <username>`

### 8.2 동작 규칙
- 출력은 표준출력(성공), 오류는 표준에러 + non-zero exit code
- `delete`는 연관 Credential/History cascade 삭제
- `reset-password`는 대상 유저 존재 여부 검증 후 해시 갱신

---

## 9. 테스트 전략

### 9.1 단위 테스트
- 비밀번호 해시/검증 유틸
- AES-GCM 암복호화 유틸
- 권한 필터(`user_id` 스코프) 검증

### 9.2 통합 테스트
- 로그인/로그아웃/세션 만료 시나리오
- 사용자 A/B 분리(credential, history 상호 비가시성)
- history pagination 및 replay payload 검증
- admin CLI 명령 정상/예외 케이스

### 9.3 보안 회귀
- 인증 없이 보호 API 접근 시 401
- 타 사용자 리소스 접근 시 403/404
- 민감정보 로그 노출 여부 점검

---

## 10. 롤아웃/운영
1. DB 스키마 마이그레이션 적용
2. 마스터 키 설정 및 배포
3. 초기 관리자 계정 생성
4. 기능 플래그(`auth-enabled`)로 단계적 오픈 고려
5. 모니터링:
   - 로그인 실패율
   - 세션 만료율
   - credential/historical API 오류율

---

## 11. 오픈 이슈 및 결정 필요사항
- 로컬 단일 사용자 모드 유지 여부(`--auth-enabled=false`) 확정 필요
- 초기 관리자 계정 생성 정책(환경변수 vs 첫 실행 인터랙션) 확정 필요
- 마스터 키 로테이션 전략(다중 키 버전 관리) 설계 필요
