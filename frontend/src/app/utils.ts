import { RuntimeMeta } from "../shared/api/types";
import {
  Locale,
  ObjectGroup,
  SourceRecent,
  TableHistoryState,
  TableRunStatus,
  TargetState,
  WsStatus,
} from "./types";
import { SOURCE_RECENT_KEY, SOURCE_REMEMBER_KEY, TARGET_RECENT_KEY, UI_LOCALE_KEY } from "./constants";

export function normalizeTableKey(tableName: string): string {
  return tableName.trim().toUpperCase();
}

export function formatHistoryTime(value: string): string {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return "-";
  }
  return parsed.toLocaleString();
}

export function loadRememberPassword(): boolean {
  try {
    return localStorage.getItem(SOURCE_REMEMBER_KEY) === "true";
  } catch {
    return false;
  }
}

export function loadSourceRecent(): SourceRecent {
  try {
    const raw = localStorage.getItem(SOURCE_RECENT_KEY);
    if (!raw) {
      return { oracleUrl: "", username: "", password: "" };
    }
    const parsed = JSON.parse(raw) as Partial<SourceRecent>;
    return {
      oracleUrl: parsed.oracleUrl ?? "",
      username: parsed.username ?? "",
      password: parsed.password ?? "",
    };
  } catch {
    return { oracleUrl: "", username: "", password: "" };
  }
}

export function loadTargetRecent(): Partial<TargetState> {
  try {
    const raw = localStorage.getItem(TARGET_RECENT_KEY);
    if (!raw) return {};
    return JSON.parse(raw) as Partial<TargetState>;
  } catch {
    return {};
  }
}

export function loadLocale(): Locale {
  try {
    const raw = localStorage.getItem(UI_LOCALE_KEY);
    if (raw === "ko") {
      return "ko";
    }
  } catch {
    // no-op
  }
  return "en";
}

export function toBool(value: unknown, fallback: boolean): boolean {
  if (typeof value === "boolean") {
    return value;
  }
  if (typeof value === "string") {
    const normalized = value.trim().toLowerCase();
    if (normalized === "true") return true;
    if (normalized === "false") return false;
  }
  return fallback;
}

export function toNumber(value: unknown, fallback: number): number {
  if (typeof value === "number" && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === "string" && value.trim() !== "") {
    const parsed = Number(value);
    if (Number.isFinite(parsed)) {
      return parsed;
    }
  }
  return fallback;
}

export function toString(value: unknown, fallback = ""): string {
  if (typeof value === "string") {
    return value;
  }
  return fallback;
}

export function toObjectGroup(value: unknown, fallback: ObjectGroup): ObjectGroup {
  const normalized = toString(value, fallback).trim().toLowerCase();
  if (normalized === "tables" || normalized === "sequences") {
    return normalized;
  }
  return "all";
}

export function toStringArray(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return [];
  }
  return value.filter((item): item is string => typeof item === "string");
}

export function isObjectGroupModeEnabled(meta: RuntimeMeta | null): boolean {
  return meta?.features?.objectGroupMode ?? true;
}

export function createSessionId(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  return `v16-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

export function wsStatusLabel(status: WsStatus, locale: Locale): string {
  const tr = (en: string, ko: string): string => (locale === "ko" ? ko : en);
  switch (status) {
    case "connecting":
      return tr("WS connecting", "WS 연결 중");
    case "connected":
      return tr("WS connected", "WS 연결됨");
    case "closed":
      return tr("WS disconnected", "WS 연결 종료");
    case "error":
      return tr("WS error", "WS 오류");
    default:
      return tr("WS idle", "WS 대기");
  }
}

export function tableStatusLabel(status: TableRunStatus, locale: Locale): string {
  const tr = (en: string, ko: string): string => (locale === "ko" ? ko : en);
  switch (status) {
    case "running":
      return tr("Running", "실행 중");
    case "completed":
      return tr("Completed", "완료");
    case "error":
      return tr("Error", "오류");
    default:
      return tr("Pending", "대기");
  }
}

export function tableStatusBadgeClass(status: TableRunStatus): string {
  switch (status) {
    case "completed":
      return "border-emerald-300 bg-emerald-100 text-emerald-900";
    case "error":
      return "border-red-300 bg-red-100 text-red-900";
    case "running":
      return "border-blue-300 bg-blue-100 text-blue-900";
    default:
      return "border-slate-300 bg-slate-100 text-slate-800";
  }
}

export function historyStatusLabel(status: TableHistoryState["status"], locale: Locale): string {
  return status === "success"
    ? locale === "ko"
      ? "이력 성공"
      : "History success"
    : locale === "ko"
      ? "이력 실패"
      : "History failed";
}

export function historyStatusBadgeClass(status: TableHistoryState["status"]): string {
  return status === "success"
    ? "border-emerald-300 bg-emerald-100 text-emerald-900"
    : "border-red-300 bg-red-100 text-red-900";
}

export function parseReplayedTables(optionsJson: string): string[] {
  if (!optionsJson) return [];
  try {
    const parsed = JSON.parse(optionsJson) as { tables?: unknown };
    if (!Array.isArray(parsed.tables)) return [];
    return parsed.tables.filter((item): item is string => typeof item === "string");
  } catch {
    return [];
  }
}
