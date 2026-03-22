# DBMigrator Project History (Summary)

이 문서는 Oracle에서 PostgreSQL로의 데이터 마이그레이션 도구인 DBMigrator의 발전 과정을 요약한 기록입니다.

## 1. 프로젝트 진화 개요
- **초기 (v01~v02):** 데이터 추출 및 SQL 파일 생성을 위한 기본 CLI 도구.
- **웹 도입 (v03~v04):** Gin 프레임워크 기반의 Web UI 및 실시간 모니터링 도입.
- **기능 고도화 (v05~v20):** LOB 스트리밍, 제약 조건 처리, 마이그레이션 재개(Resume) 등 안정성 강화.
- **현대화 (v21~v25):** React 전환, Tailwind CSS 기반 디자인 개편, 상세 이력 관리 기능.
- **엔터프라이즈 (v26+):** 대규모 테이블 최적화, 고해상도 지원, 구글 OAuth2 인증 통합.

---

## 2. 주요 버전별 마일스톤

### Phase 1: CLI & Core Engine (v01 - v02)
- **핵심:** 고성능 Oracle 데이터 추출 엔진 구축.
- **주요 기능:** 
  - `INSERT` SQL 파일 생성 및 병렬 처리 (`--parallel`).
  - 중간 파일 없는 **직접 마이그레이션(Direct Migration)** 지원.
  - Oracle 메타데이터 기반 `CREATE TABLE` DDL 자동 생성.

### Phase 2: Web Interface & Real-time Tracking (v03 - v04)
- **핵심:** 사용자 편의를 위한 웹 대시보드 도입.
- **주요 기능:**
  - Gin 웹 서버 내장 및 WebSocket 기반 실시간 진행률 표시.
  - 마이그레이션 결과물 ZIP 다운로드 제공.
  - CLI의 모든 파라미터를 Web UI 설정으로 통합.

### Phase 3: Resilience & Enterprise Features (v05 - v20)
- **핵심:** 실제 운영 환경에서의 복원력 및 보안 강화.
- **주요 기능:**
  - **데이터:** BLOB/CLOB 스트리밍 처리, 인덱스 및 FK 제약조건 마이그레이션.
  - **운영:** 실패 지점부터 다시 시작하는 **Resume** 기능, 상세 에러 로그.
  - **보안:** 사용자 로그인 시스템, 접속 정보(Credential) 암호화 저장.

### Phase 4: UI/UX Revolution (ui-improve - v25)
- **핵심:** 현대적인 웹 애플리케이션으로의 탈바꿈.
- **주요 기능:**
  - **UX:** 3단계 위저드(Stepper) 방식 도입, React 기반 SPA 전환.
  - **디자인:** 다크 모드 지원, 반응형 레이아웃 적용.
  - **관리:** 테이블별 상세 실행 이력 조회 및 설정값 재사용(Replay).

### Phase 5: Scalability & Modern Auth (v26 - Present)
- **핵심:** 대규모 환경 대응 및 클라우드 친화적 인증.
- **주요 기능:**
  - **최적화:** 100+ 테이블 대응을 위한 리스트 스크롤 및 콤마(,) 구분 검색.
  - **레이아웃:** 와이드 모니터(1600px+) 대응 그리드 최적화.
  - **인증:** **구글 로그인(OAuth2)** 연동 및 접속 이력 자동 저장.

---

## 3. 기술 스택 진화
- **Language:** Go 1.21+ (Backend), TypeScript (Frontend)
- **Backend:** Gin, WebSocket, `go-ora`, `pgx/v5`
- **Frontend:** React, Tailwind CSS, Vite, Vitest
- **Database:** SQLite (Auth & History storage)
