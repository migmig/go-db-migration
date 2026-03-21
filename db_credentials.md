# Database & Application Credentials

## 1. Oracle Database (Source - Docker)
- **Container Name:** `magical_maxwell`
- **URL / DSN:** `localhost:1521/XE`
- **Username:** `SYSTEM` (또는 `SYS`)
- **Password:** `YmY2NzVm`

## 2. PostgreSQL Database (Target - Docker)
- **Container Name:** `my-postgres`
- **URL / DSN:** `postgres://postgres:mysecretpassword@localhost:5432/postgres`
- **Username:** `postgres`
- **Password:** `mysecretpassword`
- **Database Name:** `postgres`
- **Default Schema:** `public`

## 3. Web UI / Admin CLI Login
- **Username:** `admin`
- **Password:** `stronger-password-123` (또는 프로젝트 초기화 시 설정한 비밀번호)

> **참고:** 위 정보는 현재 실행 중인 로컬 Docker 컨테이너에서 추출한 접속 정보입니다.
