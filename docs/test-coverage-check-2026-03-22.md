# 테스트 커버리지 점검 (2026-03-22)

## 실행 커맨드

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -n 1
```

## 결과 요약

- 전체 statement 커버리지: **56.6%**
- `internal/bus` 패키지 커버리지 수집 정상화: **100.0%**

### 패키지별 커버리지

- `dbmigrator`: 32.2%
- `internal/bus`: 100.0%
- `internal/config`: 89.0%
- `internal/db`: 56.5%
- `internal/dialect`: 48.1%
- `internal/logger`: 100.0%
- `internal/migration`: 57.7%
- `internal/security`: 86.7%
- `internal/web`: 55.4%
- `internal/web/ws`: 54.9%
- `internal/web/ziputil`: 82.1%

## 비고

- `internal/bus` 테스트 추가 후 기존 `covdata` 오류가 재현되지 않았습니다.
- 다음 커버리지 우선 개선 영역: `dbmigrator`, `internal/dialect`, `internal/web/ws`.
