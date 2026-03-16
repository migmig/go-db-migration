# 작업 목록 (Tasks) - v16

## 목표: Vite/React/Tailwind 전환 + 오프라인 런타임 보장 + 연결 UX 단일화

### 1. 설계/문서화 (Design & Documentation)
- [x] `docs/v16/prd.md` 작성
- [x] `docs/v16/spec.md` 작성
- [x] `docs/v16/tasks.md` 작성
- [ ] `README.md` 개발/빌드/오프라인 실행 가이드 보강
  - [ ] 프런트 빌드 명령 및 산출물 위치
  - [ ] 런타임 오프라인 동작 방식(노드 미의존) 설명

### 2. 프런트 프로젝트 스캐폴딩 (Frontend Scaffolding)
- [ ] `frontend/` 초기화
  - [ ] Vite + React + TypeScript 설정
  - [ ] Tailwind CSS + PostCSS 설정
  - [ ] ESLint/기본 스크립트(`dev/build/test`) 정리
- [ ] 공통 앱 골격 구성
  - [ ] `src/app/App.tsx`
  - [ ] `src/shared/api/client.ts` (fetch 래퍼)
  - [ ] 전역 스타일/테마 변수 정리

### 3. 오프라인 서빙 경로 구현 (Offline Runtime Path)
- [ ] Go 정적 파일 서빙 경로 추가
  - [ ] `frontend/dist` 자산 embed 또는 배포 경로 연결
  - [ ] `/api/*` 우선 라우팅 + SPA fallback 적용
- [ ] 외부 CDN 의존 제거
  - [ ] 폰트/아이콘/스크립트 로컬 번들화 확인

### 4. Step 1 UI 전환 (Connection UX)
- [ ] Source/Target 연결 화면 React 컴포넌트 전환
  - [ ] Source URL 옆 `저장된 연결 불러오기`
  - [ ] Target URL 옆 `저장된 연결 불러오기`
  - [ ] `최근 입력(선택)` 기본 접힘
- [ ] 저장된 연결 패널 UX 정리
  - [ ] 필터 상태(`all/source/target`) 명시
  - [ ] 역할별 빈 상태 메시지 반영

### 5. 인증/세션/이력 연동 (Auth & History)
- [ ] 인증 상태 게이트 연동(`/api/auth/me`)
- [ ] 로그인/로그아웃 플로우 React 상태로 이전
- [ ] 내 작업 이력 조회/재실행 플로우 이전

### 6. 마이그레이션 실행 영역 이전 (Step 2/3)
- [ ] 테이블 선택/옵션 폼 React 전환
- [ ] 실행/진행률/요약 카드 React 전환
- [ ] WebSocket 수신 및 상태 갱신 로직 이전

### 7. 테스트 (Testing)
- [ ] 프런트 단위/컴포넌트 테스트
  - [ ] 역할별 불러오기 필터
  - [ ] replay payload 폼 반영
  - [ ] 세션 만료 처리 UI
- [ ] 서버 회귀 테스트
  - [ ] `go test ./...`
  - [ ] 정적 파일 서빙 + SPA fallback 경로

### 8. 롤아웃/전환 (Rollout)
- [ ] 병행 운영 전략 확정
  - [ ] 구 UI fallback 유지 기간 정의
  - [ ] 롤백 절차 문서화
- [ ] 최종 전환 체크리스트
  - [ ] 오프라인 환경 실행 검증
  - [ ] 인증 모드(`-auth-enabled`) 동작 검증
  - [ ] 주요 사용자 시나리오 E2E 확인
