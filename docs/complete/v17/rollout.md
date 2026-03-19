# v17 롤아웃/운영 가이드

## 1. 단계적 릴리즈 계획

### 1-1. 제어 포인트

- `DBM_OBJECT_GROUP_UI_ENABLED=true`
  - v16 UI에 `Migration target` 선택기와 그룹별 조회/결과 패널을 노출한다.
- `DBM_OBJECT_GROUP_UI_ENABLED=false`
  - v16 UI를 legacy `all` 모드로 고정한다.
  - 백엔드의 `object_group` 호환성은 유지하지만 일반 운영자는 v17 분리 실행 UI를 보지 않는다.

### 1-2. 권장 활성화 순서

1. 스테이징
   - `DBM_OBJECT_GROUP_UI_ENABLED=true`
   - `all`, `tables`, `sequences` 3개 모드로 dry-run과 실제 실행을 각각 1회 이상 검증한다.
2. 내부 운영자 제한 오픈
   - 운영 담당 서버 또는 내부 접근 가능한 배포 슬롯에서만 `DBM_OBJECT_GROUP_UI_ENABLED=true`
   - 일반 사용자 서버는 `false`로 유지한다.
3. 전체 오픈
   - 운영 지표에서 모드별 실패율과 재시도 성공률이 허용 범위인지 확인한 뒤 전체 서버에 `true`를 반영한다.

## 2. 운영팀 모드 선택 가이드

- `all`
  - 기본값이다.
  - 테이블/데이터 이관과 시퀀스 반영을 한 번에 수행해야 할 때 사용한다.
  - 테이블 단계 실패 시 시퀀스 단계는 자동 스킵된다.
- `tables`
  - 테이블/데이터 경로만 먼저 검증하거나 시퀀스 변경을 의도적으로 제외해야 할 때 사용한다.
  - 대규모 데이터 적재 후 시퀀스는 별도 승인 절차로 분리하려는 운영 절차에 적합하다.
- `sequences`
  - 테이블/데이터 적재가 이미 완료되었고 시퀀스 보정만 따로 반영해야 할 때 사용한다.
  - 복구 작업이나 재시도 작업에 우선 적용한다.

## 3. 장애 대응 플레이북

### 3-1. `sequences-only` 복구 절차

1. 실패 이력에서 대상 작업의 `report_id`, `object_group`, `stats.tables`, `stats.sequences`를 확인한다.
2. 테이블 그룹이 정상 완료되었고 시퀀스 그룹만 실패했는지 판단한다.
3. 실패 원인이 권한, 이름 충돌, 대상 DB 접속 일시 장애인지 로그의 `migration.statement.failed` 필드로 확인한다.
4. 원인 제거 후 `object_group=sequences`로 재실행하거나 History replay로 `sequences` 모드를 재적용한다.
5. 완료 후 리포트에서 `stats.sequences.error_count == 0`인지 다시 확인한다.

### 3-2. 모드 오선택 방지 체크리스트

- 데이터와 DDL을 함께 반영해야 하면 `all`인지 확인한다.
- 시퀀스 반영을 제외하려는 명확한 운영 사유가 없으면 `tables`를 선택하지 않는다.
- 이미 데이터 적재가 끝난 복구 작업이 아니면 `sequences`를 선택하지 않는다.
- Dry-run 결과의 `TABLES SQL` / `SEQUENCES SQL` 섹션이 기대한 범위와 일치하는지 검토한다.

## 4. 최종 배포 체크

- dry-run 결과에서 대상 테이블 수와 시퀀스 수가 운영 변경 요청과 일치하는지 확인한다.
- 완료 리포트와 WebSocket 요약에 `object_group`, `stats.tables`, `stats.sequences`가 모두 노출되는지 확인한다.
- 운영 대시보드 또는 `/api/monitoring/metrics`에서 `migrations.all|tables|sequences` 지표가 수집되는지 확인한다.
- History replay 시 legacy 이력이 `all`로 복원되는지 샘플 1건 이상 확인한다.
