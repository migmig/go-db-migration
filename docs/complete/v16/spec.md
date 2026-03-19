# 기술 사양서 (Technical Specifications) - v16

## 1. 아키텍처 개요
v16은 v15의 인증/권한/이력 백엔드 구조를 유지하면서, 프런트엔드를 `Vite + React + Tailwind CSS`로 전환한다.
핵심 목표는 다음 3가지다.

1. 연결 UX 단일화(`저장된 연결` 중심)
2. 프런트 코드베이스 모듈화(컴포넌트/상태 분리)
3. 런타임 오프라인 동작 보장(Node 미의존)

---

## 2. 컴포넌트 설계

### 2.1 서버 컴포넌트
- `internal/web/server.go`
  - `/api/*` 기존 계약 유지
  - SPA 정적 파일 서빙 및 라우팅 fallback 처리
- `internal/web/assets` (신규, go:embed 대상)
  - `frontend/dist` 산출물 포함

### 2.2 클라이언트 컴포넌트
- `frontend/` (신규)
  - `React 19 + TypeScript + Vite`
  - `Tailwind CSS` 기반 UI 레이어
  - 인증 상태, credential, history, migration 설정 상태를 컴포넌트 단위로 분리

---

## 3. 디렉터리 구조(안)

```text
frontend/
  index.html
  package.json
  vite.config.ts
  tailwind.config.ts
  postcss.config.js
  src/
    main.tsx
    app/
      App.tsx
      routes.tsx
    features/
      auth/
      credentials/
      history/
      migration/
    shared/
      api/
      ui/
      hooks/
      styles/
```

---

## 4. 빌드/배포/오프라인 전략

### 4.1 빌드 단계
1. `frontend` 의존성 설치 (`npm ci` 혹은 `pnpm i --frozen-lockfile`)
2. `vite build` 수행
3. 생성된 `frontend/dist`를 Go에서 정적 서빙

### 4.2 런타임 단계(오프라인)
- `./dbmigrator -web` 실행 시 내장 정적 파일만으로 UI 제공
- Node, npm, Vite 프로세스가 런타임에 필요하지 않음
- 외부 CDN 의존 없음(폰트/아이콘/스크립트 로컬 번들)

### 4.3 fallback 라우팅
- 브라우저 직접 진입(` / `, `/history` 등) 시 `index.html`로 fallback
- `/api/*`, `/static/*`는 기존 API/정적 라우팅 우선

---

## 5. UI/상태 모델

### 5.1 도메인 상태
- `auth`: 로그인 사용자, 세션 만료 상태
- `connections`: source/target 현재 입력, 최근 입력, 저장된 연결 목록/필터
- `history`: 내 작업 이력 목록/페이지네이션/replay payload
- `migration`: stepper 단계, 선택 테이블, 실행 옵션, 실시간 진행 상태

### 5.2 핵심 UX 규칙
- `저장된 연결`이 연결 데이터의 단일 관리 지점
- Source/Target 필드 옆 `저장된 연결 불러오기`는 역할 기반 필터(`source/target`)를 강제
- `최근 입력(선택)`은 보조 기능이며 기본 접힘 상태

---

## 6. API 계약

v16에서 API 스펙은 v15와 동일하게 유지한다.

- 인증: `/api/auth/login`, `/api/auth/logout`, `/api/auth/me`
- 접속정보: `/api/credentials` CRUD
- 이력: `/api/history`, `/api/history/:id`, `/api/history/:id/replay`
- 마이그레이션: `/api/tables`, `/api/migrate`, `/api/migrate/retry`, `/api/test-target`
- 모니터링: `/api/monitoring/metrics`

---

## 7. 보안/운영 고려사항

1. **민감정보 처리**
   - 브라우저 저장소(localStorage)는 비밀번호 저장을 opt-in으로 제한
   - 저장된 연결 비밀번호는 서버측 암호화 저장 정책(v15) 유지
2. **세션 정책**
   - 기존 쿠키 정책(HttpOnly/SameSite/idle+absolute timeout) 유지
3. **오프라인 운영**
   - 사내망/폐쇄망 환경에서도 바이너리 단독 실행 가능해야 함

---

## 8. 테스트 전략

### 8.1 프런트
- 단위 테스트: 역할 필터, 폼 상태 동기화, replay payload 적용
- 컴포넌트 테스트: Source/Target 불러오기 플로우, 세션 만료 UI

### 8.2 서버/통합
- 기존 `go test ./...` 회귀 유지
- 정적 파일 서빙 및 SPA fallback 라우팅 테스트 추가

### 8.3 E2E(선택)
- 로그인 → 저장된 Source 불러오기 → 테이블 조회
- Target 불러오기 → 연결 테스트 → 마이그레이션 시작

---

## 9. 마이그레이션 전략

1. **Phase 1**
   - `frontend` 스캐폴딩, 빌드 파이프라인, Go 정적 서빙 연결
2. **Phase 2**
   - Step 1(연결 화면) React 전환
3. **Phase 3**
   - Step 2/3, history/monitoring 전환
4. **Phase 4**
   - 기존 템플릿 잔존 코드 정리

---

## 10. 결정 필요사항

- 패키지 매니저 표준(`npm` vs `pnpm`)
- 프런트 테스트 도구(Vitest + Testing Library) 도입 범위
- 기존 `index.html`을 병행 유지할 기간(rollback window)
