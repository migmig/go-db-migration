# 작업 목록 (Tasks)

## 목표: App.tsx 리팩토링 분리 + 완료 문서 정리

### 1) 문서 정리
- [x] `docs/v23-prd.md` → `docs/v23/prd.md`로 이동
- [x] `docs/v22/*` → `docs/complete/v22/*`로 정리
- [x] `docs/v23/spec.md` 작성
- [x] `docs/v23/tasks.md` 작성

---

### 2) 타입/상수/유틸 분리
- [x] `frontend/src/app/types.ts` 생성 및 타입 이동
- [x] `frontend/src/app/constants.ts` 생성 및 상수 이동
- [x] `frontend/src/app/utils.ts` 생성 및 유틸 이동
- [x] `App.tsx` 내부 중복 타입/상수/유틸 제거

---

### 3) UI 컴포넌트 추출
- [x] `LoginModal.tsx` 추출
- [x] `HeaderBar.tsx` 추출
- [x] `RecentSource.tsx` 추출
- [x] `ConnectionForms.tsx` 추출
- [x] `TableSelection.tsx` 추출
- [x] `MigrationOptionsPanel.tsx` 추출
- [x] `RunStatus.tsx` 추출
- [x] `CredentialsPanel.tsx` 추출
- [x] `HistoryPanel.tsx` 추출

---

### 4) App.tsx Orchestrator 정리
- [x] 하위 컴포넌트 import/props 연결
- [x] 핸들러/useEffect/useMemo 의존성 점검
- [x] dead code 및 사용하지 않는 import 제거

---

### 5) 검증
- [x] `cd frontend && npm run test`
- [x] `cd frontend && npm run typecheck`
- [x] `cd frontend && npm run build`
- [x] 주요 수동 시나리오 회귀 확인 (로그인/연결/pre-check/실행/히스토리)
