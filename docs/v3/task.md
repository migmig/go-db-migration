# Implementation Tasks: Web UI Addition (v3)

## Phase 1: Web Server Infrastructure
- [x] Gin 프레임워크 의존성 추가 및 초기화
- [x] `--web` 플래그 및 실행 모드 분기 로직 구현
- [x] HTML 템플릿 및 정적 파일 라우팅 설정
- [x] 프로젝트 레이아웃 구성 (`web/templates`, `internal/web`)

## Phase 2: Core Table Management API
- [x] Oracle DB 연결 및 테이블 목록 조회 API (`POST /api/tables`)
- [x] `LIKE` 필터를 통한 테이블 검색 기능 구현
- [x] 프론트엔드 테이블 리스트 렌더링 및 체크박스 선택 로직

## Phase 3: Real-time Progress Tracking (WebSocket)
- [x] WebSocket Tracker 구현 (`internal/web/ws`)
- [x] `migration.Run`에 `ProgressTracker` 인터페이스 도입 및 연동
- [x] 프론트엔드 WebSocket 클라이언트 구현 및 프로그레스 바 시각화
- [x] 실시간 처리 건수 표시 기능

## Phase 4: Output Management & ZIP
- [x] ZIP 압축 유틸리티 구현 (`internal/web/ziputil`)
- [x] 작업 완료 후 자동 ZIP 압축 및 임시 SQL 파일 정리 로직
- [x] ZIP 파일 다운로드 API 구현 (`GET /api/download/:id`)
- [x] 다운로드 후 일정 시간 뒤 ZIP 파일 자동 삭제 (Cleanup)

## Phase 5: UI/UX Refinement & Next Steps
- [x] Vanilla JS 기반의 현대적 UI 디자인 (Glassmorphism, Inter font)
- [x] 에러 핸들링 및 사용자 알림 UI
- [x] Web UI에서 PostgreSQL 직접 마이그레이션(Direct Copy) 옵션 추가
- [x] UI에서 Batch Size 및 Worker 수 설정 기능 추가
- [x] 대용량 테이블 처리 시 WebSocket 이벤트 Throttling 최적화
