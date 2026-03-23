import { Locale, MigrationOptions } from "./types";

export const SOURCE_RECENT_KEY = "dbm:v16:source-recent";
export const SOURCE_REMEMBER_KEY = "dbm:v16:source-remember-pass";
export const TARGET_RECENT_KEY = "dbm:v16:target-recent";
export const UI_LOCALE_KEY = "dbm:v18:ui-locale";

export const UI_TEXT: Record<Locale, Record<string, string>> = {
  en: {
    loading: "Loading v21 preview...",
    bootFailed: "v21 boot failed",
    retry: "Retry",
    workspaceTitle: "v21 Migration Workspace",
    workspaceDesc: "Source/target setup, migration options, and real-time run status.",
    authEnabled: "Auth enabled",
    authDisabled: "Auth disabled",
    savedConnections: "Saved Connections",
    myHistory: "My History",
    logout: "Logout",
    recentSourceOptional: "Recent source input (optional)",
    rememberSourcePassword: "Remember source password",
    restore: "Restore",
    clear: "Clear",
    switchToKorean: "한국어",
    switchToEnglish: "English",
  },
  ko: {
    loading: "v21 미리보기를 불러오는 중...",
    bootFailed: "v21 부팅 실패",
    retry: "다시 시도",
    workspaceTitle: "v21 마이그레이션 작업공간",
    workspaceDesc: "소스/타깃 설정, 마이그레이션 옵션, 실시간 실행 상태를 관리합니다.",
    authEnabled: "인증 사용 중",
    authDisabled: "인증 비활성화",
    savedConnections: "저장된 연결",
    myHistory: "내 작업 이력",
    logout: "로그아웃",
    recentSourceOptional: "최근 소스 입력값 (선택)",
    rememberSourcePassword: "소스 비밀번호 기억",
    restore: "복원",
    clear: "지우기",
    switchToKorean: "한국어",
    switchToEnglish: "English",
  },
};

export const DEFAULT_OPTIONS: MigrationOptions = {
  objectGroup: "all",
  outFile: "migration.sql",
  perTable: true,
  withDdl: true,
  withSequences: false,
  withIndexes: false,
  withConstraints: false,
  validate: false,
  truncate: false,
  upsert: false,
  oracleOwner: "",
  batchSize: 1000,
  workers: 4,
  copyBatch: 10000,
  dbMaxOpen: 0,
  dbMaxIdle: 2,
  dbMaxLife: 0,
  logJson: false,
  dryRun: false,
};
