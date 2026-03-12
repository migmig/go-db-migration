# Technical Specification: Web UI Addition (v3)

## 1. 개요 (Introduction)
v3에서는 기존 CLI 기반의 dbmigrator에 사용자 친화적인 웹 인터페이스를 추가합니다. Go의 Gin 프레임워크를 사용하여 경량 웹 서버를 구축하고, WebSocket을 통해 실시간 마이그레이션 진행 상태를 시각화합니다.

## 2. 기술 스택 (Technical Stack)
- **Backend**: Go 1.21+, Gin Web Framework
- **Frontend**: HTML5, Vanilla JavaScript, CSS3 (Google Fonts 'Inter')
- **Real-time Communication**: WebSockets (`github.com/gorilla/websocket`)
- **Zip Generation**: Standard `archive/zip` library

## 3. 상세 설계 (Detailed Design)

### 3.1. Web Server Architecture
- **Framework**: `gin-gonic/gin`을 사용하여 라우팅 및 미들웨어를 관리합니다.
- **Static Assets**: HTML 템플릿은 `web/templates`에, 정적 자산은 `web/static`에 위치합니다. (현재는 `index.html`에 스타일과 스크립트가 내장된 형태)
- **Concurrency**: 각 마이그레이션 요청은 별도의 고루틴에서 실행되어 서버의 응답성을 유지합니다.

### 3.2. API 엔드포인트
- `POST /api/tables`: Oracle DB 연결 정보를 받아 테이블 목록을 반환합니다. (`LIKE` 검색 지원)
- `POST /api/migrate`: 선택된 테이블들에 대해 마이그레이션을 시작합니다. (비동기 처리)
- `GET /api/progress`: WebSocket 연결을 통해 실시간 이벤트를 전송합니다.
- `GET /api/download/:id`: 생성된 ZIP 파일을 다운로드합니다.

### 3.3. WebSocket 프로토콜 명세 (JSON)
- **Init**: `{"type": "init", "table": "NAME", "total": 1000}`
- **Update**: `{"type": "update", "table": "NAME", "count": 500}`
- **Done**: `{"type": "done", "table": "NAME"}`
- **Error**: `{"type": "error", "table": "NAME", "error": "MSG"}`
- **All Done**: `{"type": "all_done", "zip_file_id": "FILENAME.zip"}`

### 3.4. 마이그레이션 및 ZIP 처리
- 웹 모드에서의 마이그레이션은 항상 `PerTable: true` 옵션을 사용합니다.
- 결과물은 `os.TempDir()` 하위의 임시 디렉토리에 생성됩니다.
- 모든 테이블 작업 완료 후 `ziputil`을 통해 디렉토리를 압축하고, 원본 SQL 파일들은 즉시 삭제합니다.
- ZIP 파일은 다운로드 후 약 5분 뒤에 자동 삭제되도록 스케줄링됩니다.

## 4. 보안 고려사항
- **경로 트래버스 방지**: 다운로드 API에서 `filepath.Base()`를 사용하여 권한 없는 파일 접근을 차단합니다.
- **제한된 로컬 환경**: 본 도구는 로컬 구동용이며, 외부 노출 시 추가적인 인증 레이어가 필요합니다.
