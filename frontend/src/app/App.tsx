import { FormEvent, useEffect, useMemo, useRef, useState } from "react";
import { apiRequest } from "../shared/api/client";
import {
  AuthUser,
  Credential,
  HistoryEntry,
  HistoryListResponse,
  PrecheckResponse,
  PrecheckSummary,
  PrecheckTableResult,
  RuntimeMeta,
} from "../shared/api/types";

type RoleFilter = "all" | "source" | "target";
type NoticeTone = "info" | "error";
type PrecheckDecisionFilter = "all" | "transfer_required" | "skip_candidate" | "count_check_failed";
type WsStatus = "idle" | "connecting" | "connected" | "closed" | "error";
type TableRunStatus = "pending" | "running" | "completed" | "error";
type TableHistoryStatusFilter = "all" | "not_started" | "success" | "failed";
type TableSortOption = "table_asc" | "table_desc" | "recent_desc" | "runs_desc" | "history_status";
type ObjectGroup = "all" | "tables" | "sequences";

type SourceState = {
  oracleUrl: string;
  username: string;
  password: string;
  like: string;
};

type TargetState = {
  mode: "file" | "direct";
  targetDb: string;
  targetUrl: string;
  schema: string;
};

type SourceRecent = {
  oracleUrl: string;
  username: string;
  password: string;
};

type MigrationOptions = {
  objectGroup: ObjectGroup;
  outFile: string;
  perTable: boolean;
  withDdl: boolean;
  withSequences: boolean;
  withIndexes: boolean;
  withConstraints: boolean;
  validate: boolean;
  truncate: boolean;
  upsert: boolean;
  oracleOwner: string;
  batchSize: number;
  workers: number;
  copyBatch: number;
  dbMaxOpen: number;
  dbMaxIdle: number;
  dbMaxLife: number;
  logJson: boolean;
  dryRun: boolean;
};

type TableRunState = {
  total: number;
  count: number;
  status: TableRunStatus;
  error?: string;
  details?: string;
};

type TableHistoryState = {
  status: "success" | "failed";
  runCount: number;
  lastRunAt: string;
};

type TableHistoryDetail = {
  tableName: string;
  entries: HistoryEntry[];
};

function normalizeTableKey(tableName: string): string {
  return tableName.trim().toUpperCase();
}

function formatHistoryTime(value: string): string {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) {
    return "-";
  }
  return parsed.toLocaleString();
}

type ValidationState = {
  sourceCount: number;
  targetCount: number;
  status: string;
  message: string;
};

type DdlEvent = {
  key: string;
  object: string;
  name: string;
  status: string;
  error?: string;
};

type DiscoverySummary = {
  objectGroup: ObjectGroup;
  tables: string[];
  sequences: string[];
};

type ReportSummary = {
  total_rows: number;
  success_count: number;
  error_count: number;
  duration: string;
  report_id: string;
  object_group: ObjectGroup;
  stats: GroupedStats;
};

type GroupStats = {
  total_items: number;
  success_count: number;
  error_count: number;
  skipped_count: number;
  total_rows?: number;
};

type GroupedStats = {
  tables: GroupStats;
  sequences: GroupStats;
};

type WsProgressMsg = {
  type: string;
  table?: string;
  count?: number;
  total?: number;
  error?: string;
  message?: string;
  zip_file_id?: string;
  connection_ok?: boolean;
  object?: string;
  object_name?: string;
  status?: string;
  object_group?: ObjectGroup;
  tables?: string[];
  sequences?: string[];
  phase?: string;
  category?: string;
  suggestion?: string;
  recoverable?: boolean;
  batch_num?: number;
  row_offset?: number;
  report_summary?: ReportSummary;
};

type MetricsState = {
  cpu: string;
  mem: string;
};

type Locale = "en" | "ko";

const SOURCE_RECENT_KEY = "dbm:v16:source-recent";
const SOURCE_REMEMBER_KEY = "dbm:v16:source-remember-pass";
const TARGET_RECENT_KEY = "dbm:v16:target-recent";
const UI_LOCALE_KEY = "dbm:v18:ui-locale";

const UI_TEXT: Record<Locale, Record<string, string>> = {
  en: {
    loading: "Loading v18 preview...",
    bootFailed: "v18 boot failed",
    retry: "Retry",
    workspaceTitle: "v18 Migration Workspace",
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
    loading: "v18 미리보기를 불러오는 중...",
    bootFailed: "v18 부팅 실패",
    retry: "다시 시도",
    workspaceTitle: "v18 마이그레이션 작업공간",
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

const DEFAULT_OPTIONS: MigrationOptions = {
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

function loadRememberPassword(): boolean {
  try {
    return localStorage.getItem(SOURCE_REMEMBER_KEY) === "true";
  } catch {
    return false;
  }
}

function loadSourceRecent(): SourceRecent {
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

function loadTargetRecent(): Partial<TargetState> {
  try {
    const raw = localStorage.getItem(TARGET_RECENT_KEY);
    if (!raw) return {};
    return JSON.parse(raw) as Partial<TargetState>;
  } catch {
    return {};
  }
}

function loadLocale(): Locale {
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

function toBool(value: unknown, fallback: boolean): boolean {
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

function toNumber(value: unknown, fallback: number): number {
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

function toString(value: unknown, fallback = ""): string {
  if (typeof value === "string") {
    return value;
  }
  return fallback;
}

function toObjectGroup(value: unknown, fallback: ObjectGroup): ObjectGroup {
  const normalized = toString(value, fallback).trim().toLowerCase();
  if (normalized === "tables" || normalized === "sequences") {
    return normalized;
  }
  return "all";
}

function toStringArray(value: unknown): string[] {
  if (!Array.isArray(value)) {
    return [];
  }
  return value.filter((item): item is string => typeof item === "string");
}

function isObjectGroupModeEnabled(meta: RuntimeMeta | null): boolean {
  return meta?.features?.objectGroupMode ?? true;
}

function createSessionId(): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return crypto.randomUUID();
  }
  return `v16-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

function wsStatusLabel(status: WsStatus, locale: Locale): string {
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

function tableStatusLabel(status: TableRunStatus, locale: Locale): string {
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

function tableStatusBadgeClass(status: TableRunStatus): string {
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

function historyStatusLabel(status: TableHistoryState["status"], locale: Locale): string {
  return status === "success"
    ? locale === "ko"
      ? "이력 성공"
      : "History success"
    : locale === "ko"
      ? "이력 실패"
      : "History failed";
}

function historyStatusBadgeClass(status: TableHistoryState["status"]): string {
  return status === "success"
    ? "border-emerald-300 bg-emerald-100 text-emerald-900"
    : "border-red-300 bg-red-100 text-red-900";
}

function parseReplayedTables(optionsJson: string): string[] {
  if (!optionsJson) return [];
  try {
    const parsed = JSON.parse(optionsJson) as { tables?: unknown };
    if (!Array.isArray(parsed.tables)) return [];
    return parsed.tables.filter((item): item is string => typeof item === "string");
  } catch {
    return [];
  }
}

export function App() {
  const [locale, setLocale] = useState<Locale>(() => loadLocale());
  const initialRememberPass = loadRememberPassword();
  const initialRecent = loadSourceRecent();
  const initialTarget = loadTargetRecent();

  const [meta, setMeta] = useState<RuntimeMeta | null>(null);
  const [user, setUser] = useState<AuthUser | null>(null);
  const [booting, setBooting] = useState(true);
  const [bootError, setBootError] = useState("");

  const [source, setSource] = useState<SourceState>({
    oracleUrl: initialRecent.oracleUrl,
    username: initialRecent.username,
    password: initialRecent.password,
    like: "",
  });
  const [target, setTarget] = useState<TargetState>({
    mode: initialTarget.mode ?? "file",
    targetDb: initialTarget.targetDb ?? "postgres",
    targetUrl: initialTarget.targetUrl ?? "",
    schema: initialTarget.schema ?? "public",
  });

  const [rememberSourcePassword, setRememberSourcePassword] =
    useState(initialRememberPass);

  const [sourceConnectBusy, setSourceConnectBusy] = useState(false);
  const [sourceConnectError, setSourceConnectError] = useState("");
  const [allTables, setAllTables] = useState<string[]>([]);
  const [tableSearch, setTableSearch] = useState("");
  const [tableStatusFilter, setTableStatusFilter] = useState<TableHistoryStatusFilter>("all");
  const [tableSort, setTableSort] = useState<TableSortOption>("table_asc");
  const [excludeMigratedSuccess, setExcludeMigratedSuccess] = useState(false);
  const [selectedTables, setSelectedTables] = useState<string[]>([]);

  const [options, setOptions] = useState<MigrationOptions>(DEFAULT_OPTIONS);

  const [targetTestBusy, setTargetTestBusy] = useState(false);
  const [targetTestError, setTargetTestError] = useState("");
  const [targetTestMessage, setTargetTestMessage] = useState("");

  const [tableProgress, setTableProgress] = useState<Record<string, TableRunState>>(
    {},
  );
  const [validation, setValidation] = useState<Record<string, ValidationState>>({});
  const [ddlEvents, setDdlEvents] = useState<DdlEvent[]>([]);
  const [warnings, setWarnings] = useState<string[]>([]);
  const [discoverySummary, setDiscoverySummary] = useState<DiscoverySummary | null>(
    null,
  );
  const [metrics, setMetrics] = useState<MetricsState>({ cpu: "-", mem: "-" });
  const [migrationBusy, setMigrationBusy] = useState(false);
  const [migrationError, setMigrationError] = useState("");
  const [wsStatus, setWsStatus] = useState<WsStatus>("idle");
  const [runSessionId, setRunSessionId] = useState("");
  const [runStartedAt, setRunStartedAt] = useState<number | null>(null);
  const [runEndedAt, setRunEndedAt] = useState<number | null>(null);
  const [runDryRun, setRunDryRun] = useState(false);
  const [zipFileId, setZipFileId] = useState("");
  const [reportSummary, setReportSummary] = useState<ReportSummary | null>(null);
  const [clock, setClock] = useState(Date.now());

  const [credentialsPanelOpen, setCredentialsPanelOpen] = useState(false);
  const [credentialFilter, setCredentialFilter] = useState<RoleFilter>("all");
  const [credentialsBusy, setCredentialsBusy] = useState(false);
  const [credentialsError, setCredentialsError] = useState("");
  const [credentials, setCredentials] = useState<Credential[]>([]);

  const [historyPanelOpen, setHistoryPanelOpen] = useState(false);
  const [historyBusy, setHistoryBusy] = useState(false);
  const [historyError, setHistoryError] = useState("");
  const [history, setHistory] = useState<HistoryEntry[]>([]);
  const [activeTableHistory, setActiveTableHistory] = useState<string | null>(null);
  const [tableHistoryBusy, setTableHistoryBusy] = useState(false);
  const [tableHistoryError, setTableHistoryError] = useState("");

  // v19: precheck state
  const [precheckBusy, setPrecheckBusy] = useState(false);
  const [precheckError, setPrecheckError] = useState("");
  const [precheckSummary, setPrecheckSummary] = useState<PrecheckSummary | null>(null);
  const [precheckItems, setPrecheckItems] = useState<PrecheckTableResult[]>([]);
  const [precheckDecisionFilter, setPrecheckDecisionFilter] = useState<PrecheckDecisionFilter>("all");
  const [precheckPolicy, setPrecheckPolicy] = useState("strict");

  const [loginForm, setLoginForm] = useState({ username: "", password: "" });
  const [loginBusy, setLoginBusy] = useState(false);
  const [loginError, setLoginError] = useState("");

  const [notice, setNotice] = useState<{ text: string; tone: NoticeTone } | null>(
    null,
  );

  const wsRef = useRef<WebSocket | null>(null);
  const warningSetRef = useRef<Set<string>>(new Set());
  const migrationActiveRef = useRef(false);
  const runDryRunRef = useRef(false);

  const filteredCredentials = credentials.filter((item) => {
    if (credentialFilter === "all") return true;
    if (credentialFilter === "source") return item.dbType === "oracle";
    return item.dbType !== "oracle";
  });

  const historyByTable = useMemo<Record<string, TableHistoryState>>(() => {
    const next: Record<string, TableHistoryState> = {};
    for (const entry of history) {
      const tables = parseReplayedTables(entry.optionsJson);
      if (tables.length === 0) continue;
      for (const tableName of tables) {
        const normalized = normalizeTableKey(tableName);
        const current = next[normalized];
        if (!current) {
          next[normalized] = {
            status: entry.status === "success" ? "success" : "failed",
            runCount: 1,
            lastRunAt: entry.createdAt,
          };
          continue;
        }
        current.runCount += 1;
        if (new Date(entry.createdAt).getTime() > new Date(current.lastRunAt).getTime()) {
          current.lastRunAt = entry.createdAt;
          current.status = entry.status === "success" ? "success" : "failed";
        }
      }
    }
    return next;
  }, [history]);

  const tableHistoryDetails = useMemo<Record<string, TableHistoryDetail>>(() => {
    const next: Record<string, TableHistoryDetail> = {};
    for (const entry of history) {
      const tables = parseReplayedTables(entry.optionsJson);
      for (const tableName of tables) {
        const normalized = normalizeTableKey(tableName);
        if (!next[normalized]) {
          next[normalized] = { tableName: normalized, entries: [] };
        }
        next[normalized].entries.push(entry);
      }
    }

    Object.values(next).forEach((detail) => {
      detail.entries.sort(
        (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
      );
    });

    return next;
  }, [history]);

  const filteredTables = useMemo(() => {
    const filtered = allTables.filter((table) => {
      if (!table.toLowerCase().includes(tableSearch.toLowerCase())) {
        return false;
      }
      const historyState = historyByTable[normalizeTableKey(table)];
      if (excludeMigratedSuccess && historyState?.status === "success") {
        return false;
      }
      if (tableStatusFilter === "not_started") {
        return !historyState;
      }
      if (tableStatusFilter === "success") {
        return historyState?.status === "success";
      }
      if (tableStatusFilter === "failed") {
        return historyState?.status === "failed";
      }
      return true;
    });

    const historyRank = (tableName: string): number => {
      const state = historyByTable[normalizeTableKey(tableName)];
      if (!state) return 0;
      if (state.status === "failed") return 1;
      return 2;
    };

    filtered.sort((a, b) => {
      if (tableSort === "table_desc") {
        return b.localeCompare(a);
      }
      if (tableSort === "recent_desc") {
        const aTime = historyByTable[normalizeTableKey(a)]?.lastRunAt ?? "";
        const bTime = historyByTable[normalizeTableKey(b)]?.lastRunAt ?? "";
        const diff = new Date(bTime).getTime() - new Date(aTime).getTime();
        if (diff !== 0) return diff;
        return a.localeCompare(b);
      }
      if (tableSort === "runs_desc") {
        const aCount = historyByTable[normalizeTableKey(a)]?.runCount ?? 0;
        const bCount = historyByTable[normalizeTableKey(b)]?.runCount ?? 0;
        if (bCount !== aCount) return bCount - aCount;
        return a.localeCompare(b);
      }
      if (tableSort === "history_status") {
        const rankDiff = historyRank(a) - historyRank(b);
        if (rankDiff !== 0) return rankDiff;
        return a.localeCompare(b);
      }
      return a.localeCompare(b);
    });

    return filtered;
  }, [allTables, excludeMigratedSuccess, historyByTable, tableSearch, tableSort, tableStatusFilter]);
  const objectGroupModeEnabled = isObjectGroupModeEnabled(meta);
  const selectedTableSet = new Set(selectedTables);
  const activeHistoryDetail = activeTableHistory
    ? tableHistoryDetails[normalizeTableKey(activeTableHistory)]
    : undefined;
  const allVisibleSelected =
    filteredTables.length > 0 &&
    filteredTables.every((table) => selectedTableSet.has(table));

  const runEntries = Object.entries(tableProgress).sort((a, b) => {
    const rank: Record<TableRunStatus, number> = {
      running: 0,
      pending: 1,
      error: 2,
      completed: 3,
    };
    const statusDiff = rank[a[1].status] - rank[b[1].status];
    if (statusDiff !== 0) return statusDiff;
    return a[0].localeCompare(b[0]);
  });

  const runTotalTables = runEntries.length;
  const runDoneTables = runEntries.filter(
    ([, item]) => item.status === "completed" || item.status === "error",
  ).length;
  const runSuccessCount = runEntries.filter(
    ([, item]) => item.status === "completed",
  ).length;
  const runFailCount = runEntries.filter(([, item]) => item.status === "error").length;
  const processedRows = runEntries.reduce((sum, [, item]) => sum + item.count, 0);
  const expectedRows = runEntries.reduce((sum, [, item]) => {
    if (item.total > 0) return sum + item.total;
    return sum + item.count;
  }, 0);
  const overallPercent =
    runTotalTables > 0 ? Math.floor((runDoneTables / runTotalTables) * 100) : 0;
  const elapsedSeconds = runStartedAt
    ? Math.max(
        1,
        Math.floor(((runEndedAt ?? clock) - runStartedAt) / 1000),
      )
    : 0;
  const rowsPerSecond =
    elapsedSeconds > 0 ? Math.floor(processedRows / elapsedSeconds) : 0;
  const etaSeconds =
    rowsPerSecond > 0 && expectedRows > processedRows
      ? Math.floor((expectedRows - processedRows) / rowsPerSecond)
      : null;
  const runReadyToShow = runStartedAt !== null || runEntries.length > 0;
  const groupSummary = reportSummary?.stats ?? null;
  const effectiveObjectGroup = objectGroupModeEnabled ? options.objectGroup : "all";
  const previewObjectGroup = objectGroupModeEnabled
    ? discoverySummary?.objectGroup ?? effectiveObjectGroup
    : "all";
  const previewTables = discoverySummary?.tables ?? selectedTables;
  const previewSequences = objectGroupModeEnabled ? discoverySummary?.sequences ?? [] : [];

  const t = (key: string): string => UI_TEXT[locale][key] ?? UI_TEXT.en[key] ?? key;
  const tr = (en: string, ko: string): string => (locale === "ko" ? ko : en);

  useEffect(() => {
    try {
      localStorage.setItem(UI_LOCALE_KEY, locale);
    } catch {
      // no-op
    }
  }, [locale]);

  useEffect(() => {
    migrationActiveRef.current = migrationBusy;
  }, [migrationBusy]);

  useEffect(() => {
    runDryRunRef.current = runDryRun;
  }, [runDryRun]);

  useEffect(() => {
    const timeout = setTimeout(() => {
      if (notice) {
        setNotice(null);
      }
    }, 2400);
    return () => clearTimeout(timeout);
  }, [notice]);

  useEffect(() => {
    if (!migrationBusy) {
      return;
    }
    const id = window.setInterval(() => {
      setClock(Date.now());
    }, 1000);
    return () => window.clearInterval(id);
  }, [migrationBusy]);

  useEffect(() => {
    try {
      localStorage.setItem(SOURCE_REMEMBER_KEY, String(rememberSourcePassword));
    } catch {
      // Ignore storage errors in restricted browser environments.
    }
  }, [rememberSourcePassword]);

  useEffect(() => {
    try {
      localStorage.setItem(
        SOURCE_RECENT_KEY,
        JSON.stringify({
          oracleUrl: source.oracleUrl,
          username: source.username,
          password: rememberSourcePassword ? source.password : "",
        }),
      );
    } catch {
      // Ignore storage errors in restricted browser environments.
    }
  }, [source.oracleUrl, source.username, source.password, rememberSourcePassword]);

  useEffect(() => {
    try {
      localStorage.setItem(TARGET_RECENT_KEY, JSON.stringify(target));
    } catch {
      // Ignore storage errors in restricted browser environments.
    }
  }, [target.mode, target.targetDb, target.targetUrl, target.schema]);

  useEffect(() => {
    void boot();
  }, []);

  useEffect(() => {
    return () => {
      closeWebSocket();
    };
  }, []);

  function closeWebSocket() {
    const socket = wsRef.current;
    if (!socket) {
      return;
    }
    wsRef.current = null;
    socket.onopen = null;
    socket.onclose = null;
    socket.onerror = null;
    socket.onmessage = null;
    if (socket.readyState === WebSocket.OPEN || socket.readyState === WebSocket.CONNECTING) {
      socket.close();
    }
  }

  function resetRunState() {
    closeWebSocket();
    warningSetRef.current = new Set();
    setWarnings([]);
    setValidation({});
    setDdlEvents([]);
    setDiscoverySummary(null);
    setMetrics({ cpu: "-", mem: "-" });
    setTableProgress({});
    setMigrationError("");
    setMigrationBusy(false);
    setWsStatus("idle");
    setRunSessionId("");
    setRunStartedAt(null);
    setRunEndedAt(null);
    setRunDryRun(false);
    setZipFileId("");
    setReportSummary(null);
  }

  async function openWebSocket(sessionId: string): Promise<boolean> {
    closeWebSocket();
    setWsStatus("connecting");

    return await new Promise((resolve) => {
      const protocol = window.location.protocol === "https:" ? "wss" : "ws";
      const socket = new WebSocket(
        `${protocol}://${window.location.host}/api/ws?sessionId=${encodeURIComponent(sessionId)}`,
      );
      wsRef.current = socket;

      let settled = false;
      const finish = (value: boolean) => {
        if (settled) return;
        settled = true;
        resolve(value);
      };

      const timer = window.setTimeout(() => {
        finish(false);
      }, 3000);

      socket.onopen = () => {
        window.clearTimeout(timer);
        setWsStatus("connected");
        finish(true);
      };

      socket.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data) as WsProgressMsg;
          handleProgressMessage(msg);
        } catch {
          // Ignore malformed websocket payloads.
        }
      };

      socket.onerror = () => {
        setWsStatus("error");
      };

      socket.onclose = () => {
        window.clearTimeout(timer);
        if (wsRef.current === socket) {
          wsRef.current = null;
        }
        if (migrationActiveRef.current) {
          setWsStatus("closed");
        }
        finish(false);
      };
    });
  }

  function handleProgressMessage(msg: WsProgressMsg) {
    if (msg.type === "metrics") {
      if (!msg.message) return;
      try {
        const payload = JSON.parse(msg.message) as {
          cpu_usage_pct?: string;
          mem_usage_mb?: string;
        };
        setMetrics({
          cpu: payload.cpu_usage_pct ? `${payload.cpu_usage_pct}%` : "-",
          mem: payload.mem_usage_mb ? `${payload.mem_usage_mb} MB` : "-",
        });
      } catch {
        // Ignore malformed metrics payloads.
      }
      return;
    }

    if (msg.type === "warning") {
      const warningText = msg.message ?? "Unknown warning";
      if (warningSetRef.current.has(warningText)) {
        return;
      }
      warningSetRef.current.add(warningText);
      setWarnings((prev) => [warningText, ...prev]);
      return;
    }

    if (msg.type === "ddl_progress") {
      const object = msg.object ?? "";
      const objectName = msg.object_name ?? "";
      const key = `${object}:${objectName}`;
      setDdlEvents((prev) => {
        const updated = [
          {
            key,
            object,
            name: objectName,
            status: msg.status ?? "unknown",
            error: msg.error,
          },
          ...prev.filter((item) => item.key !== key),
        ];
        return updated.slice(0, 20);
      });
      return;
    }

    if (msg.type === "discovery_summary") {
      setDiscoverySummary({
        objectGroup: msg.object_group ?? options.objectGroup,
        tables: msg.tables ?? [],
        sequences: msg.sequences ?? [],
      });
      return;
    }

    if (msg.type === "validation_start") {
      if (!msg.table) return;
      const table = msg.table;
      setValidation((prev) => {
        if (prev[table]) {
          return prev;
        }
        return {
          ...prev,
          [table]: {
            sourceCount: 0,
            targetCount: 0,
            status: "running",
            message: "",
          },
        };
      });
      return;
    }

    if (msg.type === "validation_result") {
      if (!msg.table) return;
      setValidation((prev) => ({
        ...prev,
        [msg.table!]: {
          sourceCount: msg.total ?? 0,
          targetCount: msg.count ?? 0,
          status: msg.status ?? "unknown",
          message: msg.message ?? "",
        },
      }));
      return;
    }

    if (msg.type === "all_done") {
      setMigrationBusy(false);
      setRunEndedAt(Date.now());
      setZipFileId(msg.zip_file_id ?? "");
      setReportSummary(msg.report_summary ?? null);
      setNotice({
        text: runDryRunRef.current
          ? "Verification completed."
          : "Migration completed.",
        tone: "info",
      });
      closeWebSocket();
      setWsStatus("closed");
      return;
    }

    if (!msg.table) {
      return;
    }

    if (msg.type === "init") {
      setTableProgress((prev) => {
        const current = prev[msg.table!] ?? { total: 0, count: 0, status: "pending" };
        return {
          ...prev,
          [msg.table!]: {
            ...current,
            total: msg.total ?? current.total,
            status: "running",
          },
        };
      });
      return;
    }

    if (msg.type === "update") {
      setTableProgress((prev) => {
        const current = prev[msg.table!] ?? { total: 0, count: 0, status: "pending" };
        return {
          ...prev,
          [msg.table!]: {
            ...current,
            count: msg.count ?? current.count,
            status: "running",
          },
        };
      });
      return;
    }

    if (msg.type === "done") {
      setTableProgress((prev) => {
        const current = prev[msg.table!] ?? { total: 0, count: 0, status: "pending" };
        return {
          ...prev,
          [msg.table!]: {
            ...current,
            count: current.total > 0 ? current.total : current.count,
            status: "completed",
            error: undefined,
            details: undefined,
          },
        };
      });
      return;
    }

    if (msg.type === "dry_run_result") {
      const ok = msg.connection_ok ?? false;
      setTableProgress((prev) => {
        const current = prev[msg.table!] ?? { total: 0, count: 0, status: "pending" };
        const rowCount = msg.total ?? current.total ?? current.count;
        return {
          ...prev,
          [msg.table!]: {
            ...current,
            total: rowCount,
            count: rowCount,
            status: ok ? "completed" : "error",
            error: ok ? undefined : "Target connection failed",
            details: ok ? undefined : "Target connection failed in dry-run.",
          },
        };
      });
      return;
    }

    if (msg.type === "error") {
      const detailParts: string[] = [];
      if (msg.phase) detailParts.push(`phase=${msg.phase}`);
      if (msg.category) detailParts.push(`category=${msg.category}`);
      if (msg.batch_num) detailParts.push(`batch=${msg.batch_num}`);
      if (msg.row_offset) detailParts.push(`offset=${msg.row_offset}`);
      if (msg.suggestion) detailParts.push(`suggestion=${msg.suggestion}`);
      if (typeof msg.recoverable === "boolean") {
        detailParts.push(`recoverable=${String(msg.recoverable)}`);
      }

      setTableProgress((prev) => {
        const current = prev[msg.table!] ?? { total: 0, count: 0, status: "pending" };
        return {
          ...prev,
          [msg.table!]: {
            ...current,
            status: "error",
            error: msg.error ?? "Unknown error",
            details: detailParts.join(" · "),
          },
        };
      });
    }
  }

  async function boot() {
    setBooting(true);
    setBootError("");
    try {
      const { response, data } = await apiRequest<RuntimeMeta>("/api/meta", {}, {
        allowUnauthorized: true,
      });
      if (!response.ok) {
        throw new Error("Failed to load runtime metadata.");
      }
      setMeta(data);

      if (!data.authEnabled) {
        setUser(null);
        return;
      }

      const me = await apiRequest<AuthUser | { error: string }>("/api/auth/me", {}, {
        allowUnauthorized: true,
      });
      if (me.response.ok) {
        setUser(me.data as AuthUser);
      } else {
        setUser(null);
      }
    } catch (error) {
      setBootError(error instanceof Error ? error.message : "Unknown boot error");
    } finally {
      setBooting(false);
    }
  }

  async function handleLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoginError("");
    setLoginBusy(true);
    try {
      const { response, data } = await apiRequest<AuthUser | { error: string }>(
        "/api/auth/login",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(loginForm),
        },
        { allowUnauthorized: true },
      );
      if (!response.ok) {
        const message = (data as { error?: string }).error ?? "Login failed";
        throw new Error(message);
      }
      setUser(data as AuthUser);
      setLoginForm((prev) => ({ ...prev, password: "" }));
      setNotice({ text: "Logged in successfully.", tone: "info" });
    } catch (error) {
      setLoginError(error instanceof Error ? error.message : "Login failed");
    } finally {
      setLoginBusy(false);
    }
  }

  async function handleLogout() {
    await apiRequest("/api/auth/logout", { method: "POST" }, { allowUnauthorized: true });
    resetRunState();
    setUser(null);
    setCredentialsPanelOpen(false);
    setHistoryPanelOpen(false);
    setNotice({ text: "Logged out.", tone: "info" });
  }

  async function loadCredentials() {
    if (!meta?.authEnabled || !user) {
      return;
    }
    setCredentialsBusy(true);
    setCredentialsError("");
    try {
      const { response, data } = await apiRequest<{ items: Credential[] }>("/api/credentials");
      if (!response.ok) {
        throw new Error("Failed to load credentials.");
      }
      setCredentials(data.items ?? []);
    } catch (error) {
      setCredentialsError(
        error instanceof Error ? error.message : "Failed to load credentials.",
      );
    } finally {
      setCredentialsBusy(false);
    }
  }

  async function openCredentialsPanel(filter: RoleFilter) {
    if (!meta?.authEnabled) {
      return;
    }
    if (!user) {
      setNotice({ text: "Please log in first.", tone: "error" });
      return;
    }
    if (filter === "target") {
      setTarget((prev) => ({ ...prev, mode: "direct" }));
    }
    setCredentialFilter(filter);
    setCredentialsPanelOpen(true);
    await loadCredentials();
  }

  function applyCredential(item: Credential) {
    if (item.dbType === "oracle") {
      setSource((prev) => ({
        ...prev,
        oracleUrl: item.host ?? "",
        username: item.username ?? "",
        password: item.password ?? "",
      }));
      setNotice({ text: `${item.alias} applied to source form.`, tone: "info" });
    } else {
      setTarget((prev) => ({
        ...prev,
        mode: "direct",
        targetDb: item.dbType || "postgres",
        targetUrl: item.host ?? "",
      }));
      setNotice({ text: `${item.alias} applied to target form.`, tone: "info" });
    }
    setCredentialsPanelOpen(false);
  }

  async function fetchHistoryEntries(): Promise<HistoryEntry[]> {
    const { response, data } = await apiRequest<HistoryListResponse>(
      "/api/history?page=1&pageSize=10",
    );
    if (!response.ok) {
      throw new Error("Failed to load migration history.");
    }
    return data.items ?? [];
  }

  async function openHistoryPanel() {
    if (!meta?.authEnabled || !user) {
      return;
    }
    setHistoryPanelOpen(true);
    await loadHistory();
  }

  async function loadHistory() {
    if (!meta?.authEnabled || !user) {
      return;
    }
    setHistoryBusy(true);
    setHistoryError("");
    try {
      setHistory(await fetchHistoryEntries());
    } catch (error) {
      setHistoryError(
        error instanceof Error ? error.message : "Failed to load migration history.",
      );
    } finally {
      setHistoryBusy(false);
    }
  }

  async function openTableHistory(table: string) {
    setActiveTableHistory(table);
    setTableHistoryError("");

    if (!meta?.authEnabled || !user) {
      return;
    }

    setTableHistoryBusy(true);
    try {
      setHistory(await fetchHistoryEntries());
    } catch (error) {
      setTableHistoryError(
        error instanceof Error ? error.message : "Failed to load migration history.",
      );
    } finally {
      setTableHistoryBusy(false);
    }
  }

  async function replayHistory(id: number) {
    try {
      const { response, data } = await apiRequest<{ payload: Record<string, unknown> }>(
        `/api/history/${id}/replay`,
        { method: "POST" },
      );
      if (!response.ok) {
        throw new Error("Failed to replay history.");
      }
      applyReplayPayload(data.payload ?? {});
      setHistoryPanelOpen(false);
      setNotice({ text: "History payload applied to form.", tone: "info" });
    } catch (error) {
      setHistoryError(error instanceof Error ? error.message : "Replay failed.");
    }
  }

  function applyReplayPayload(payload: Record<string, unknown>) {
    const direct = Boolean(payload.direct);
    const replayObjectGroup = objectGroupModeEnabled
      ? toObjectGroup(payload.objectGroup, options.objectGroup)
      : "all";
    const replayOptions = {
      objectGroup: replayObjectGroup,
      outFile: toString(payload.outFile, options.outFile),
      perTable: toBool(payload.perTable, options.perTable),
      withDdl: toBool(payload.withDdl, options.withDdl),
      withSequences: toBool(payload.withSequences, options.withSequences),
      withIndexes: toBool(payload.withIndexes, options.withIndexes),
      withConstraints: toBool(payload.withConstraints, options.withConstraints),
      validate: toBool(payload.validate, options.validate),
      truncate: toBool(payload.truncate, options.truncate),
      upsert: toBool(payload.upsert, options.upsert),
      oracleOwner: toString(payload.oracleOwner, options.oracleOwner),
      batchSize: toNumber(payload.batchSize, options.batchSize),
      workers: toNumber(payload.workers, options.workers),
      copyBatch: toNumber(payload.copyBatch, options.copyBatch),
      dbMaxOpen: toNumber(payload.dbMaxOpen, options.dbMaxOpen),
      dbMaxIdle: toNumber(payload.dbMaxIdle, options.dbMaxIdle),
      dbMaxLife: toNumber(payload.dbMaxLife, options.dbMaxLife),
      logJson: toBool(payload.logJson, options.logJson),
      dryRun: toBool(payload.dryRun, options.dryRun),
    };
    const replayTables = toStringArray(payload.tables);

    setSource((prev) => ({
      ...prev,
      oracleUrl: toString(payload.oracleUrl, ""),
      username: toString(payload.username, ""),
      password: "",
      like: "",
    }));
    setTarget((prev) => ({
      ...prev,
      mode: direct ? "direct" : "file",
      targetDb: toString(payload.targetDb, prev.targetDb),
      targetUrl: toString(payload.targetUrl ?? payload.pgUrl, ""),
      schema: toString(payload.schema, ""),
    }));
    setOptions((prev) => {
      const next = { ...prev, ...replayOptions };
      if (replayObjectGroup === "tables") {
        next.withSequences = false;
      }
      if (replayObjectGroup === "sequences") {
        next.withDdl = true;
        next.withSequences = true;
      }
      return next;
    });
    if (replayTables.length > 0) {
      setSelectedTables(replayTables);
    }
  }

  async function connectSource() {
    setSourceConnectError("");
    setAllTables([]);
    setSelectedTables([]);
    setDiscoverySummary(null);
    if (!source.oracleUrl || !source.username || !source.password) {
      setSourceConnectError("Oracle URL, username and password are required.");
      return;
    }

    setSourceConnectBusy(true);
    try {
      const { response, data } = await apiRequest<{ tables: string[]; error?: string }>(
        "/api/tables",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            oracleUrl: source.oracleUrl,
            username: source.username,
            password: source.password,
            like: source.like,
          }),
        },
      );
      if (!response.ok) {
        throw new Error(data.error ?? "Failed to load Oracle tables.");
      }
      const tables = data.tables ?? [];
      setAllTables(tables);
      setTableSearch("");
      if (meta?.authEnabled && user) {
        await loadHistory();
      }
      setNotice({ text: `Loaded ${tables.length} table(s).`, tone: "info" });
    } catch (error) {
      setSourceConnectError(
        error instanceof Error ? error.message : "Failed to load Oracle tables.",
      );
    } finally {
      setSourceConnectBusy(false);
    }
  }

  async function testTarget() {
    setTargetTestError("");
    setTargetTestMessage("");
    if (target.mode !== "direct") {
      setTargetTestMessage("File mode selected. Target DB test is skipped.");
      return;
    }
    if (!target.targetUrl) {
      setTargetTestError("Target URL is required in direct mode.");
      return;
    }

    setTargetTestBusy(true);
    try {
      const { response, data } = await apiRequest<{ message?: string; error?: string }>(
        "/api/test-target",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            targetDb: target.targetDb,
            targetUrl: target.targetUrl,
          }),
        },
      );
      if (!response.ok) {
        throw new Error(data.error ?? "Target connection test failed.");
      }
      setTargetTestMessage(data.message ?? "Target DB connection verified.");
    } catch (error) {
      setTargetTestError(
        error instanceof Error ? error.message : "Target connection test failed.",
      );
    } finally {
      setTargetTestBusy(false);
    }
  }

  function clearRecentSource() {
    setSource((prev) => ({ ...prev, oracleUrl: "", username: "", password: "" }));
    try {
      localStorage.removeItem(SOURCE_RECENT_KEY);
    } catch {
      // Ignore storage errors in restricted browser environments.
    }
  }

  function restoreRecentSource() {
    const recent = loadSourceRecent();
    setSource((prev) => ({
      ...prev,
      oracleUrl: recent.oracleUrl,
      username: recent.username,
      password: recent.password,
    }));
    setNotice({ text: "Recent source values restored.", tone: "info" });
  }

  function toggleTable(table: string, checked: boolean) {
    setSelectedTables((prev) => {
      if (checked) {
        if (prev.includes(table)) return prev;
        return [...prev, table];
      }
      return prev.filter((item) => item !== table);
    });
  }

  function applyObjectGroupSelection(nextGroup: ObjectGroup) {
    setOptions((prev) => {
      if (nextGroup === "tables") {
        return {
          ...prev,
          objectGroup: nextGroup,
          withSequences: false,
        };
      }
      if (nextGroup === "sequences") {
        return {
          ...prev,
          objectGroup: nextGroup,
          withDdl: true,
          withSequences: true,
        };
      }
      return {
        ...prev,
        objectGroup: nextGroup,
      };
    });
  }

  function selectAllVisibleTables() {
    setSelectedTables((prev) => {
      const merged = new Set(prev);
      filteredTables.forEach((table) => merged.add(table));
      return Array.from(merged);
    });
  }

  function deselectAllVisibleTables() {
    const hiddenSet = new Set(filteredTables);
    setSelectedTables((prev) => prev.filter((table) => !hiddenSet.has(table)));
  }

  async function runPrecheck() {
    setPrecheckError("");
    if (!source.oracleUrl || !source.username || !source.password) {
      setPrecheckError("Source connection fields are required.");
      return;
    }
    if (selectedTables.length === 0) {
      setPrecheckError("Select at least one table.");
      return;
    }
    setPrecheckBusy(true);
    setPrecheckSummary(null);
    setPrecheckItems([]);
    try {
      const { response, data } = await apiRequest<PrecheckResponse>(
        "/api/migrations/precheck",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            oracleUrl: source.oracleUrl,
            username: source.username,
            password: source.password,
            tables: selectedTables,
            targetDb: target.targetDb,
            targetUrl: target.mode === "direct" ? target.targetUrl.trim() : "",
            policy: precheckPolicy,
          }),
        },
      );
      if (!response.ok || data.error) {
        setPrecheckError(data.error ?? "Pre-check failed.");
      } else {
        setPrecheckSummary(data.summary);
        setPrecheckItems(data.items ?? []);
      }
    } catch (err) {
      setPrecheckError(err instanceof Error ? err.message : "Pre-check failed.");
    } finally {
      setPrecheckBusy(false);
    }
  }

  async function startMigration() {
    setMigrationError("");

    if (!source.oracleUrl || !source.username || !source.password) {
      setMigrationError("Source connection fields are required.");
      return;
    }
    if (selectedTables.length === 0) {
      setMigrationError("Select at least one table.");
      return;
    }
    if (target.mode === "direct" && !target.targetUrl.trim()) {
      setMigrationError("Target URL is required in direct mode.");
      return;
    }

    warningSetRef.current = new Set();
    setWarnings([]);
    setValidation({});
    setDdlEvents([]);
    setDiscoverySummary(null);
    setMetrics({ cpu: "-", mem: "-" });
    setZipFileId("");
    setReportSummary(null);
    setRunDryRun(options.dryRun);
    runDryRunRef.current = options.dryRun;
    setRunStartedAt(Date.now());
    setRunEndedAt(null);
    setClock(Date.now());

    const initialState: Record<string, TableRunState> = {};
    selectedTables.forEach((table) => {
      initialState[table] = {
        total: 0,
        count: 0,
        status: "pending",
      };
    });
    setTableProgress(initialState);

    setMigrationBusy(true);
    const sessionId = createSessionId();
    setRunSessionId(sessionId);

    const wsConnected = await openWebSocket(sessionId);
    const payloadSessionId = wsConnected ? sessionId : "";
    if (!wsConnected) {
      setNotice({
        text: "WebSocket unavailable. Real-time progress might be limited.",
        tone: "error",
      });
    }

    try {
      const { response, data } = await apiRequest<{ error?: string; message?: string }>(
        "/api/migrate",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            sessionId: payloadSessionId,
            oracleUrl: source.oracleUrl,
            username: source.username,
            password: source.password,
            tables: selectedTables,
            direct: target.mode === "direct",
            targetDb: target.targetDb,
            targetUrl: target.targetUrl.trim(),
            pgUrl: target.targetUrl.trim(),
            withDdl: options.withDdl,
            batchSize: options.batchSize,
            workers: options.workers,
            outFile: options.outFile,
            perTable: options.perTable,
            schema: target.schema,
            dryRun: options.dryRun,
            logJson: options.logJson,
            withSequences: options.withSequences,
            withIndexes: options.withIndexes,
            withConstraints: options.withConstraints,
            validate: options.validate,
            oracleOwner: options.oracleOwner,
            dbMaxOpen: options.dbMaxOpen,
            dbMaxIdle: options.dbMaxIdle,
            dbMaxLife: options.dbMaxLife,
            copyBatch: options.copyBatch,
            objectGroup: effectiveObjectGroup,
            truncate: options.truncate,
            upsert: options.upsert,
          }),
        },
      );
      if (!response.ok) {
        throw new Error(data.error ?? "Failed to start migration.");
      }
      setNotice({
        text: options.dryRun ? "Verification started." : "Migration started.",
        tone: "info",
      });
    } catch (error) {
      setMigrationError(
        error instanceof Error ? error.message : "Failed to start migration.",
      );
      setMigrationBusy(false);
      setRunEndedAt(Date.now());
      closeWebSocket();
      setWsStatus("error");
    }
  }

  if (booting) {
    return (
      <div className="flex min-h-screen items-center justify-center text-slate-700">
        {t("loading")}
      </div>
    );
  }

  if (bootError) {
    return (
      <div className="mx-auto flex min-h-screen max-w-3xl items-center px-6 py-12">
        <div className="card-surface w-full p-8">
          <h1 className="text-xl font-semibold text-slate-900">{t("bootFailed")}</h1>
          <p className="mt-3 text-sm text-red-600">{bootError}</p>
          <button
            className="mt-5 rounded-lg bg-brand-600 px-4 py-2 text-sm font-semibold text-white"
            onClick={() => void boot()}
            type="button"
          >
            {t("retry")}
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="relative min-h-screen px-4 pb-16 pt-8 sm:px-6 lg:px-10">
      <div className="mx-auto flex max-w-7xl flex-col gap-6">
        <header className="card-surface flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.18em] text-brand-700">
              DBMigrator
            </p>
            <h1 className="text-2xl font-bold text-slate-900">{t("workspaceTitle")}</h1>
            <p className="mt-1 text-sm text-slate-600">
              {t("workspaceDesc")}
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <button
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
              onClick={() => setLocale((prev) => (prev === "en" ? "ko" : "en"))}
              type="button"
            >
              {locale === "en" ? t("switchToKorean") : t("switchToEnglish")}
            </button>
            {meta?.authEnabled ? (
              <>
                <span className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-semibold text-slate-700">
                  {user ? `User: ${user.username}` : t("authEnabled")}
                </span>
                <button
                  className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
                  onClick={() => void openCredentialsPanel("all")}
                  type="button"
                >
                  {t("savedConnections")}
                </button>
                <button
                  className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
                  onClick={() => void openHistoryPanel()}
                  type="button"
                >
                  {t("myHistory")}
                </button>
                <button
                  className="rounded-lg bg-slate-900 px-3 py-2 text-sm font-semibold text-white hover:bg-slate-700"
                  onClick={() => void handleLogout()}
                  type="button"
                >
                  {t("logout")}
                </button>
              </>
            ) : (
              <span className="rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-700">
                {t("authDisabled")}
              </span>
            )}
          </div>
        </header>

        <details className="card-surface p-4">
          <summary className="cursor-pointer text-sm font-semibold text-slate-800">
            {t("recentSourceOptional")}
          </summary>
          <div className="mt-4 flex flex-wrap items-center gap-2">
            <label className="inline-flex items-center gap-2 text-sm text-slate-700">
              <input
                checked={rememberSourcePassword}
                onChange={(event) => setRememberSourcePassword(event.target.checked)}
                type="checkbox"
              />
              {t("rememberSourcePassword")}
            </label>
            <button
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
              onClick={restoreRecentSource}
              type="button"
            >
              {t("restore")}
            </button>
            <button
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
              onClick={clearRecentSource}
              type="button"
            >
              {t("clear")}
            </button>
          </div>
        </details>

        <section className="grid gap-5 lg:grid-cols-2">
          <div className="card-surface p-5">
            <div className="mb-4 flex items-center justify-between gap-3">
              <h2 className="text-lg font-semibold text-slate-900">
                {tr("1. Source (Oracle)", "1. 소스 (Oracle)")}
              </h2>
              {meta?.authEnabled && (
                <button
                  className="rounded-lg border border-brand-300 bg-brand-50 px-3 py-2 text-xs font-semibold text-brand-700 hover:bg-brand-100"
                  onClick={() => void openCredentialsPanel("source")}
                  type="button"
                >
                  {tr("Load Saved Source", "저장된 소스 불러오기")}
                </button>
              )}
            </div>
            <div className="space-y-3">
              <label className="block text-sm">
                <span className="mb-1 block text-slate-700">Oracle URL</span>
                <input
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                  onChange={(event) =>
                    setSource((prev) => ({ ...prev, oracleUrl: event.target.value }))
                  }
                  placeholder="localhost:1521/XE"
                  value={source.oracleUrl}
                />
              </label>
              <label className="block text-sm">
                <span className="mb-1 block text-slate-700">{tr("Username", "사용자명")}</span>
                <input
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                  onChange={(event) =>
                    setSource((prev) => ({ ...prev, username: event.target.value }))
                  }
                  value={source.username}
                />
              </label>
              <label className="block text-sm">
                <span className="mb-1 block text-slate-700">{tr("Password", "비밀번호")}</span>
                <input
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                  onChange={(event) =>
                    setSource((prev) => ({ ...prev, password: event.target.value }))
                  }
                  type="password"
                  value={source.password}
                />
              </label>
              <label className="block text-sm">
                <span className="mb-1 block text-slate-700">Table filter (LIKE)</span>
                <input
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                  onChange={(event) =>
                    setSource((prev) => ({ ...prev, like: event.target.value }))
                  }
                  placeholder="USERS_%"
                  value={source.like}
                />
              </label>
              <button
                className="w-full rounded-xl bg-brand-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-brand-700 disabled:cursor-not-allowed disabled:opacity-60"
                disabled={sourceConnectBusy}
                onClick={() => void connectSource()}
                type="button"
              >
                {sourceConnectBusy
                  ? tr("Loading tables...", "테이블 불러오는 중...")
                  : tr("Connect & Fetch Tables", "연결 후 테이블 조회")}
              </button>
            </div>
            {sourceConnectError && (
              <p className="mt-3 text-sm font-medium text-red-600">{sourceConnectError}</p>
            )}
            {allTables.length > 0 && (
              <div className="mt-4 rounded-xl border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-800">
                <p className="font-semibold">
                  {tr("Found", "총")} {allTables.length}
                  {tr(" table(s)", "개 테이블 발견")}
                </p>
                <p className="mt-1 text-xs text-emerald-700">
                  {tr(
                    "Step 2 is ready. Select tables and options below.",
                    "2단계 준비 완료. 아래에서 테이블과 옵션을 선택하세요.",
                  )}
                </p>
              </div>
            )}
          </div>

          <div className="card-surface p-5">
            <div className="mb-4 flex items-center justify-between gap-3">
              <h2 className="text-lg font-semibold text-slate-900">
                {tr("2. Target", "2. 타깃")}
              </h2>
              {meta?.authEnabled && (
                <button
                  className="rounded-lg border border-brand-300 bg-brand-50 px-3 py-2 text-xs font-semibold text-brand-700 hover:bg-brand-100"
                  onClick={() => void openCredentialsPanel("target")}
                  type="button"
                >
                  {tr("Load Saved Target", "저장된 타깃 불러오기")}
                </button>
              )}
            </div>
            <div className="space-y-3">
              <label className="block text-sm">
                <span className="mb-1 block text-slate-700">
                  {tr("Migration mode", "마이그레이션 모드")}
                </span>
                <select
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                  onChange={(event) =>
                    setTarget((prev) => ({
                      ...prev,
                      mode: event.target.value as TargetState["mode"],
                    }))
                  }
                  value={target.mode}
                >
                  <option value="file">{tr("SQL file mode", "SQL 파일 모드")}</option>
                  <option value="direct">{tr("Direct migration", "직접 마이그레이션")}</option>
                </select>
              </label>
              <label className="block text-sm">
                <span className="mb-1 block text-slate-700">{tr("Target DB", "타깃 DB")}</span>
                <select
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                  onChange={(event) =>
                    setTarget((prev) => ({ ...prev, targetDb: event.target.value }))
                  }
                  value={target.targetDb}
                >
                  <option value="postgres">PostgreSQL</option>
                  <option value="mysql">MySQL</option>
                  <option value="mariadb">MariaDB</option>
                  <option value="sqlite">SQLite</option>
                  <option value="mssql">MSSQL</option>
                </select>
              </label>
              <label className="block text-sm">
                <span className="mb-1 block text-slate-700">{tr("Target URL", "타깃 URL")}</span>
                <input
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                  onChange={(event) =>
                    setTarget((prev) => ({ ...prev, targetUrl: event.target.value }))
                  }
                  placeholder="postgres://user:pass@host:5432/dbname"
                  value={target.targetUrl}
                />
              </label>
              <label className="block text-sm">
                <span className="mb-1 block text-slate-700">{tr("Schema", "스키마")}</span>
                <input
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                  onChange={(event) =>
                    setTarget((prev) => ({ ...prev, schema: event.target.value }))
                  }
                  value={target.schema}
                />
              </label>
              <button
                className="w-full rounded-xl bg-slate-900 px-4 py-2.5 text-sm font-semibold text-white hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-60"
                disabled={targetTestBusy}
                onClick={() => void testTarget()}
                type="button"
              >
                {targetTestBusy
                  ? tr("Testing target...", "타깃 연결 확인 중...")
                  : tr("Test Target Connection", "타깃 연결 테스트")}
              </button>
            </div>
            {targetTestError && (
              <p className="mt-3 text-sm font-medium text-red-600">{targetTestError}</p>
            )}
            {targetTestMessage && (
              <p className="mt-3 text-sm font-medium text-emerald-700">{targetTestMessage}</p>
            )}
          </div>
        </section>

        {allTables.length > 0 && (
          <section className="grid gap-5 xl:grid-cols-[1.2fr_1fr]">
            <div className="card-surface p-5">
              <div className="mb-3 flex items-center justify-between gap-3">
                <h2 className="text-lg font-semibold text-slate-900">
                  {tr("3. Table Selection", "3. 테이블 선택")}
                </h2>
                <span className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-semibold text-slate-700">
                  {selectedTables.length} / {allTables.length} {tr("selected", "선택됨")}
                </span>
              </div>
              {objectGroupModeEnabled && (
                <div className="mb-4 grid gap-3 lg:grid-cols-2">
                  <details className="rounded-xl border border-slate-200 bg-slate-50 p-3" open>
                    <summary className="cursor-pointer text-sm font-semibold text-slate-800">
                      Tables Group
                      <span className="ml-2 rounded-full bg-white px-2 py-0.5 text-xs text-slate-600">
                        {previewTables.length}
                      </span>
                    </summary>
                    <p className="mt-2 text-xs text-slate-500">
                      {discoverySummary
                        ? "Oracle discovery completed for tables group."
                        : "Selected tables to be migrated."}
                    </p>
                    <div className="mt-2 max-h-32 overflow-auto rounded-lg border border-slate-200 bg-white p-2">
                      {previewTables.length > 0 ? (
                        <ul className="space-y-1 text-sm text-slate-700">
                          {previewTables.map((table) => (
                            <li key={`preview-table-${table}`}>{table}</li>
                          ))}
                        </ul>
                      ) : (
                        <p className="text-sm text-slate-500">No tables selected.</p>
                      )}
                    </div>
                  </details>
                  <details className="rounded-xl border border-slate-200 bg-slate-50 p-3">
                    <summary className="cursor-pointer text-sm font-semibold text-slate-800">
                      Sequences Group
                      <span className="ml-2 rounded-full bg-white px-2 py-0.5 text-xs text-slate-600">
                        {previewObjectGroup === "tables" ? 0 : previewSequences.length}
                      </span>
                    </summary>
                    <p className="mt-2 text-xs text-slate-500">
                      {previewObjectGroup === "tables"
                        ? "Tables-only mode disables sequence discovery."
                        : discoverySummary
                          ? "Discovered from Oracle metadata at run start."
                          : "Sequence discovery runs automatically when migration starts."}
                    </p>
                    <div className="mt-2 max-h-32 overflow-auto rounded-lg border border-slate-200 bg-white p-2">
                      {previewObjectGroup === "tables" ? (
                        <p className="text-sm text-slate-500">Sequence group is disabled.</p>
                      ) : previewSequences.length > 0 ? (
                        <ul className="space-y-1 text-sm text-slate-700">
                          {previewSequences.map((sequence) => (
                            <li key={`preview-sequence-${sequence}`}>{sequence}</li>
                          ))}
                        </ul>
                      ) : (
                        <p className="text-sm text-slate-500">No sequences discovered yet.</p>
                      )}
                    </div>
                  </details>
                </div>
              )}
              <div className="mb-3 flex flex-wrap gap-2">
                <input
                  className="min-w-[220px] flex-1 rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                  onChange={(event) => setTableSearch(event.target.value)}
                  placeholder={tr("Search table...", "테이블 검색...")}
                  value={tableSearch}
                />
                <select
                  aria-label="Table history status filter"
                  className="rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                  onChange={(event) =>
                    setTableStatusFilter(event.target.value as TableHistoryStatusFilter)
                  }
                  value={tableStatusFilter}
                >
                  <option value="all">All history status</option>
                  <option value="not_started">Not started</option>
                  <option value="success">Migrated (success)</option>
                  <option value="failed">Migrated (failed)</option>
                </select>
                <select
                  aria-label="Table sort"
                  className="rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                  onChange={(event) => setTableSort(event.target.value as TableSortOption)}
                  value={tableSort}
                >
                  <option value="table_asc">Sort: Table name (A-Z)</option>
                  <option value="table_desc">Sort: Table name (Z-A)</option>
                  <option value="recent_desc">Sort: Latest history</option>
                  <option value="runs_desc">Sort: Run count</option>
                  <option value="history_status">Sort: History status</option>
                </select>
                <label className="inline-flex items-center gap-2 rounded-xl border border-slate-300 px-3 py-2 text-sm text-slate-700">
                  <input
                    checked={excludeMigratedSuccess}
                    onChange={(event) => setExcludeMigratedSuccess(event.target.checked)}
                    type="checkbox"
                  />
                  {tr("Exclude migrated success", "성공 이관 테이블 제외")}
                </label>
                <button
                  className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100 disabled:opacity-60"
                  disabled={migrationBusy}
                  onClick={selectAllVisibleTables}
                  type="button"
                >
                  {tr("Select visible", "현재 목록 전체 선택")}
                </button>
                <button
                  className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100 disabled:opacity-60"
                  disabled={migrationBusy}
                  onClick={deselectAllVisibleTables}
                  type="button"
                >
                  {tr("Clear visible", "현재 목록 선택 해제")}
                </button>
              </div>
              <div className="max-h-[420px] overflow-auto rounded-xl border border-slate-200 bg-white">
                <table className="w-full border-collapse text-sm">
                  <thead className="sticky top-0 bg-slate-50">
                    <tr>
                      <th className="w-12 border-b border-slate-200 px-3 py-2 text-center">
                        <input
                          checked={allVisibleSelected}
                          disabled={migrationBusy || filteredTables.length === 0}
                          onChange={(event) => {
                            if (event.target.checked) {
                              selectAllVisibleTables();
                            } else {
                              deselectAllVisibleTables();
                            }
                          }}
                          type="checkbox"
                        />
                      </th>
                      <th className="border-b border-slate-200 px-3 py-2 text-left text-xs uppercase tracking-wide text-slate-500">
                        {tr("Table", "테이블")}
                      </th>
                      <th className="border-b border-slate-200 px-3 py-2 text-left text-xs uppercase tracking-wide text-slate-500">
                        {tr("Status", "상태")}
                      </th>
                      <th className="border-b border-slate-200 px-3 py-2 text-left text-xs uppercase tracking-wide text-slate-500">
                        {tr("History", "이력")}
                      </th>
                      <th className="border-b border-slate-200 px-3 py-2 text-left text-xs uppercase tracking-wide text-slate-500">
                        {tr("Actions", "작업")}
                      </th>
                    </tr>
                  </thead>
                  <tbody>
                    {filteredTables.length === 0 && (
                      <tr>
                        <td
                          className="px-3 py-6 text-center text-slate-500"
                          colSpan={5}
                        >
                          {tr("No tables match your filter.", "필터에 맞는 테이블이 없습니다.")}
                        </td>
                      </tr>
                    )}
                    {filteredTables.map((table) => {
                      const item = tableProgress[table];
                      const historyState = historyByTable[normalizeTableKey(table)];
                      const status = item?.status ?? "pending";
                      const statusLabel = tableStatusLabel(status, locale);
                      const badgeClass = tableStatusBadgeClass(status);

                      return (
                        <tr className="border-b border-slate-100 last:border-b-0" key={table}>
                          <td className="px-3 py-2 text-center">
                            <input
                              checked={selectedTableSet.has(table)}
                              disabled={migrationBusy}
                              onChange={(event) => toggleTable(table, event.target.checked)}
                              type="checkbox"
                            />
                          </td>
                          <td className="px-3 py-2 font-medium text-slate-800">{table}</td>
                          <td className="px-3 py-2">
                            <span
                              aria-label={`Table status: ${statusLabel}`}
                              className={`inline-flex items-center gap-1 rounded-full border px-2.5 py-1 text-xs font-semibold ${badgeClass}`}
                              role="status"
                            >
                              <span aria-hidden="true">●</span>
                              {statusLabel}
                            </span>
                          </td>
                          <td className="px-3 py-2 text-xs text-slate-600">
                            {historyState ? (
                              <div className="flex flex-wrap items-center gap-2">
                                <span
                                  aria-label={historyStatusLabel(historyState.status, locale)}
                                  className={`inline-flex items-center gap-1 rounded-full border px-2 py-0.5 font-semibold ${historyStatusBadgeClass(historyState.status)}`}
                                  role="status"
                                >
                                  <span aria-hidden="true">●</span>
                                  {historyState.status === "success"
                                    ? tr("Success", "성공")
                                    : tr("Failed", "실패")}
                                </span>
                                <span>{historyState.runCount}{tr(" run(s)", "회 실행")}</span>
                                <span>{formatHistoryTime(historyState.lastRunAt)}</span>
                              </div>
                            ) : (
                              <span
                                aria-label={tr("History not started", "이력 미시작")}
                                className="inline-flex items-center gap-1 rounded-full border border-slate-300 bg-slate-100 px-2 py-0.5 font-semibold text-slate-800"
                                role="status"
                              >
                                <span aria-hidden="true">●</span>
                                {tr("Not started", "미시작")}
                              </span>
                            )}
                          </td>

                          <td className="px-3 py-2 text-xs text-slate-600">
                            <button
                              className="rounded border border-slate-300 px-2 py-1 text-xs font-medium hover:bg-slate-100"
                              onClick={() => void openTableHistory(table)}
                              type="button"
                            >
                              {tr("View history", "이력 보기")}
                            </button>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
              {activeTableHistory && (
                <div className="mt-4 rounded-xl border border-slate-200 bg-slate-50 p-4">
                  <div className="mb-3 flex items-center justify-between gap-3">
                    <h3 className="text-sm font-semibold text-slate-900">
                      {tr("Table history:", "테이블 이력:")} {activeTableHistory}
                    </h3>
                    <button
                      className="rounded border border-slate-300 px-2 py-1 text-xs font-medium hover:bg-slate-100"
                      onClick={() => setActiveTableHistory(null)}
                      type="button"
                    >
                      {tr("Close", "닫기")}
                    </button>
                  </div>
                  {tableHistoryBusy ? (
                    <div
                      className="space-y-2"
                      role="status"
                      aria-label={tr("Loading table history", "테이블 이력 로딩 중")}
                    >
                      <div className="h-11 animate-pulse rounded border border-slate-200 bg-slate-100" />
                      <div className="h-11 animate-pulse rounded border border-slate-200 bg-slate-100" />
                      <div className="h-11 animate-pulse rounded border border-slate-200 bg-slate-100" />
                    </div>
                  ) : tableHistoryError ? (
                    <div className="rounded border border-red-200 bg-red-50 p-3">
                      <p className="text-xs text-red-700">{tableHistoryError}</p>
                      <button
                        className="mt-2 rounded border border-red-300 bg-white px-2 py-1 text-xs font-semibold text-red-700 hover:bg-red-100"
                        onClick={() => activeTableHistory && void openTableHistory(activeTableHistory)}
                        type="button"
                      >
                        {tr("Retry", "재시도")}
                      </button>
                    </div>
                  ) : activeHistoryDetail && activeHistoryDetail.entries.length > 0 ? (
                    <ul className="space-y-2 text-xs text-slate-700">
                      {activeHistoryDetail.entries.slice(0, 5).map((entry) => {
                        const failed = entry.status !== "success";
                        return (
                          <li
                            className="rounded border border-slate-200 bg-white px-3 py-2"
                            key={`table-history-${entry.id}`}
                          >
                            <div className="flex flex-wrap items-center gap-2">
                              <span
                                className={`inline-flex items-center gap-1 rounded-full border px-2 py-0.5 font-semibold ${failed ? "border-red-300 bg-red-100 text-red-900" : "border-emerald-300 bg-emerald-100 text-emerald-900"}`}
                              >
                                <span aria-hidden="true">●</span>
                                {failed ? tr("Failed", "실패") : tr("Success", "성공")}
                              </span>
                              <span>{formatHistoryTime(entry.createdAt)}</span>
                              {failed && (
                                <button
                                  className="rounded border border-red-300 bg-red-50 px-2 py-0.5 font-semibold text-red-700 hover:bg-red-100"
                                  onClick={() => void replayHistory(entry.id)}
                                  type="button"
                                >
                                  {tr("Retry settings", "설정 다시 적용")}
                                </button>
                              )}
                            </div>
                            {failed && entry.logSummary && (
                              <p className="mt-1 text-red-700">{entry.logSummary}</p>
                            )}
                          </li>
                        );
                      })}
                    </ul>
                  ) : (
                    <p className="text-xs text-slate-500">
                      {tr("No history found for this table.", "이 테이블의 이력이 없습니다.")}
                    </p>
                  )}
                </div>
              )}
            </div>

            <div className="card-surface p-5">
              <h2 className="mb-4 text-lg font-semibold text-slate-900">
                {tr("4. Migration Options", "4. 마이그레이션 옵션")}
              </h2>
              <div className="space-y-3">
                {target.mode === "file" && (
                  <>
                    <label className="block text-sm">
                      <span className="mb-1 block text-slate-700">Output file</span>
                      <input
                        className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                        onChange={(event) =>
                          setOptions((prev) => ({ ...prev, outFile: event.target.value }))
                        }
                        value={options.outFile}
                      />
                    </label>
                    <label className="inline-flex items-center gap-2 text-sm text-slate-700">
                      <input
                        checked={options.perTable}
                        onChange={(event) =>
                          setOptions((prev) => ({ ...prev, perTable: event.target.checked }))
                        }
                        type="checkbox"
                      />
                      Per-table output files
                    </label>
                  </>
                )}

                {objectGroupModeEnabled && (
                  <label className="block text-sm">
                    <span className="mb-1 block text-slate-700">Migration target</span>
                    <select
                      className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                      onChange={(event) =>
                        applyObjectGroupSelection(event.target.value as ObjectGroup)
                      }
                      value={options.objectGroup}
                    >
                      <option value="all">All objects</option>
                      <option value="tables">Tables only</option>
                      <option value="sequences">Sequences only</option>
                    </select>
                  </label>
                )}

                <label className="inline-flex items-center gap-2 text-sm text-slate-700">
                  <input
                    checked={options.withDdl}
                    disabled={effectiveObjectGroup === "sequences"}
                    onChange={(event) =>
                      setOptions((prev) => ({ ...prev, withDdl: event.target.checked }))
                    }
                    type="checkbox"
                  />
                  Include CREATE TABLE DDL
                </label>
                <label className="inline-flex items-center gap-2 text-sm text-slate-700">
                  <input
                    checked={options.withSequences}
                    disabled={effectiveObjectGroup !== "all"}
                    onChange={(event) =>
                      setOptions((prev) => ({ ...prev, withSequences: event.target.checked }))
                    }
                    type="checkbox"
                  />
                  Include sequences
                </label>
                {objectGroupModeEnabled && effectiveObjectGroup !== "all" && (
                  <p className="text-xs text-slate-500">
                    {effectiveObjectGroup === "tables"
                      ? "Tables-only mode disables sequence DDL automatically."
                      : "Sequences-only mode forces DDL + sequence generation automatically."}
                  </p>
                )}
                <label className="inline-flex items-center gap-2 text-sm text-slate-700">
                  <input
                    checked={options.withIndexes}
                    onChange={(event) =>
                      setOptions((prev) => ({ ...prev, withIndexes: event.target.checked }))
                    }
                    type="checkbox"
                  />
                  Include indexes
                </label>
                <label className="inline-flex items-center gap-2 text-sm text-slate-700">
                  <input
                    checked={options.withConstraints}
                    onChange={(event) =>
                      setOptions((prev) => ({
                        ...prev,
                        withConstraints: event.target.checked,
                      }))
                    }
                    type="checkbox"
                  />
                  Include constraints
                </label>
                <label className="inline-flex items-center gap-2 text-sm text-slate-700">
                  <input
                    checked={options.validate}
                    onChange={(event) =>
                      setOptions((prev) => ({ ...prev, validate: event.target.checked }))
                    }
                    type="checkbox"
                  />
                  Validate row counts after migration
                </label>
                <label className="inline-flex items-center gap-2 text-sm text-slate-700">
                  <input
                    checked={options.truncate}
                    onChange={(event) =>
                      setOptions((prev) => ({ ...prev, truncate: event.target.checked }))
                    }
                    type="checkbox"
                  />
                  Truncate target tables before migration (prevents duplicates)
                </label>
                <label className="inline-flex items-center gap-2 text-sm text-slate-700">
                  <input
                    checked={options.upsert}
                    onChange={(event) =>
                      setOptions((prev) => ({ ...prev, upsert: event.target.checked }))
                    }
                    type="checkbox"
                  />
                  Upsert mode — skip duplicate rows by PK (table must have PK)
                </label>
                <label className="block text-sm">
                  <span className="mb-1 block text-slate-700">Oracle owner (optional)</span>
                  <input
                    className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                    onChange={(event) =>
                      setOptions((prev) => ({ ...prev, oracleOwner: event.target.value }))
                    }
                    placeholder="defaults to connected account"
                    value={options.oracleOwner}
                  />
                </label>

                <details className="rounded-xl border border-slate-200 bg-slate-50 px-3 py-2">
                  <summary className="cursor-pointer text-sm font-semibold text-slate-700">
                    Advanced
                  </summary>
                  <div className="mt-3 grid gap-3 sm:grid-cols-2">
                    <label className="block text-sm">
                      <span className="mb-1 block text-slate-700">Batch size</span>
                      <input
                        className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                        onChange={(event) =>
                          setOptions((prev) => ({
                            ...prev,
                            batchSize: toNumber(event.target.value, prev.batchSize),
                          }))
                        }
                        type="number"
                        value={options.batchSize}
                      />
                    </label>
                    <label className="block text-sm">
                      <span className="mb-1 block text-slate-700">Workers</span>
                      <input
                        className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                        onChange={(event) =>
                          setOptions((prev) => ({
                            ...prev,
                            workers: toNumber(event.target.value, prev.workers),
                          }))
                        }
                        type="number"
                        value={options.workers}
                      />
                    </label>
                    <label className="block text-sm">
                      <span className="mb-1 block text-slate-700">COPY batch</span>
                      <input
                        className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                        onChange={(event) =>
                          setOptions((prev) => ({
                            ...prev,
                            copyBatch: toNumber(event.target.value, prev.copyBatch),
                          }))
                        }
                        type="number"
                        value={options.copyBatch}
                      />
                    </label>
                    <label className="block text-sm">
                      <span className="mb-1 block text-slate-700">DB max open</span>
                      <input
                        className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                        onChange={(event) =>
                          setOptions((prev) => ({
                            ...prev,
                            dbMaxOpen: toNumber(event.target.value, prev.dbMaxOpen),
                          }))
                        }
                        type="number"
                        value={options.dbMaxOpen}
                      />
                    </label>
                    <label className="block text-sm">
                      <span className="mb-1 block text-slate-700">DB max idle</span>
                      <input
                        className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                        onChange={(event) =>
                          setOptions((prev) => ({
                            ...prev,
                            dbMaxIdle: toNumber(event.target.value, prev.dbMaxIdle),
                          }))
                        }
                        type="number"
                        value={options.dbMaxIdle}
                      />
                    </label>
                    <label className="block text-sm">
                      <span className="mb-1 block text-slate-700">DB max life (sec)</span>
                      <input
                        className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                        onChange={(event) =>
                          setOptions((prev) => ({
                            ...prev,
                            dbMaxLife: toNumber(event.target.value, prev.dbMaxLife),
                          }))
                        }
                        type="number"
                        value={options.dbMaxLife}
                      />
                    </label>
                  </div>
                  <div className="mt-3 flex flex-wrap gap-4">
                    <label className="inline-flex items-center gap-2 text-sm text-slate-700">
                      <input
                        checked={options.logJson}
                        onChange={(event) =>
                          setOptions((prev) => ({ ...prev, logJson: event.target.checked }))
                        }
                        type="checkbox"
                      />
                      JSON logging
                    </label>
                    <label className="inline-flex items-center gap-2 text-sm text-slate-700">
                      <input
                        checked={options.dryRun}
                        onChange={(event) =>
                          setOptions((prev) => ({ ...prev, dryRun: event.target.checked }))
                        }
                        type="checkbox"
                      />
                      Dry-run mode
                    </label>
                  </div>
                </details>

                {/* v19: Pre-check row count section */}
                {meta?.features?.precheckRowCount !== false && (
                  <div className="mt-4 rounded-xl border border-slate-200 bg-slate-50 p-4">
                    <div className="mb-3 flex flex-wrap items-center gap-3">
                      <h3 className="text-sm font-semibold text-slate-800">Pre-check Row Count</h3>
                      <select
                        className="rounded-lg border border-slate-300 px-2 py-1 text-xs text-slate-700"
                        value={precheckPolicy}
                        onChange={(e) => setPrecheckPolicy(e.target.value)}
                      >
                        <option value="strict">strict</option>
                        <option value="best_effort">best_effort</option>
                        <option value="skip_equal_rows">skip_equal_rows</option>
                      </select>
                      <button
                        className="rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
                        disabled={precheckBusy || migrationBusy || selectedTables.length === 0}
                        onClick={() => void runPrecheck()}
                        type="button"
                      >
                        {precheckBusy ? "Checking..." : "Run Pre-check"}
                      </button>
                    </div>
                    {precheckError && (
                      <p className="mb-2 text-xs font-medium text-red-600">{precheckError}</p>
                    )}
                    {precheckSummary && (
                      <>
                        <div className="mb-3 grid grid-cols-4 gap-2">
                          {(
                            [
                              { label: "Total", value: precheckSummary.total_tables, cls: "bg-slate-100 text-slate-800" },
                              { label: "Transfer Required", value: precheckSummary.transfer_required_count, cls: "bg-amber-100 text-amber-800" },
                              { label: "Skip Candidate", value: precheckSummary.skip_candidate_count, cls: "bg-emerald-100 text-emerald-800" },
                              { label: "Check Failed", value: precheckSummary.count_check_failed_count, cls: "bg-red-100 text-red-800" },
                            ] as { label: string; value: number; cls: string }[]
                          ).map(({ label, value, cls }) => (
                            <div className={`rounded-lg p-2 text-center ${cls}`} key={label}>
                              <p className="text-lg font-bold">{value}</p>
                              <p className="text-xs">{label}</p>
                            </div>
                          ))}
                        </div>
                        <div className="mb-2 flex gap-1">
                          {(["all", "transfer_required", "skip_candidate", "count_check_failed"] as PrecheckDecisionFilter[]).map((f) => (
                            <button
                              className={`rounded-lg px-2 py-1 text-xs font-medium ${precheckDecisionFilter === f ? "bg-blue-600 text-white" : "bg-white border border-slate-300 text-slate-700 hover:bg-slate-100"}`}
                              key={f}
                              onClick={() => setPrecheckDecisionFilter(f)}
                              type="button"
                            >
                              {f === "all" ? "All" : f === "transfer_required" ? "Transfer Required" : f === "skip_candidate" ? "Skip" : "Failed"}
                            </button>
                          ))}
                        </div>
                        <div className="max-h-48 overflow-auto rounded-lg border border-slate-200">
                          <table className="w-full text-xs">
                            <thead className="bg-slate-100 text-slate-600">
                              <tr>
                                <th className="px-2 py-1.5 text-left">Table</th>
                                <th className="px-2 py-1.5 text-right">Source</th>
                                <th className="px-2 py-1.5 text-right">Target</th>
                                <th className="px-2 py-1.5 text-right">Diff</th>
                                <th className="px-2 py-1.5 text-left">Decision</th>
                              </tr>
                            </thead>
                            <tbody>
                              {precheckItems
                                .filter((r) => precheckDecisionFilter === "all" || r.decision === precheckDecisionFilter)
                                .map((r) => (
                                  <tr className="border-t border-slate-100 hover:bg-slate-50" key={r.table_name}>
                                    <td className="px-2 py-1.5 font-mono">{r.table_name}</td>
                                    <td className="px-2 py-1.5 text-right">{r.source_row_count.toLocaleString()}</td>
                                    <td className="px-2 py-1.5 text-right">{r.target_row_count.toLocaleString()}</td>
                                    <td className={`px-2 py-1.5 text-right ${r.diff !== 0 ? "font-semibold text-amber-700" : "text-slate-500"}`}>{r.diff > 0 ? "+" : ""}{r.diff.toLocaleString()}</td>
                                    <td className="px-2 py-1.5">
                                      <span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${
                                        r.decision === "transfer_required" ? "bg-amber-100 text-amber-800" :
                                        r.decision === "skip_candidate" ? "bg-emerald-100 text-emerald-800" :
                                        "bg-red-100 text-red-800"
                                      }`}>
                                        {r.decision === "count_check_failed" && <span title={r.reason}>⚠</span>}
                                        {r.decision}
                                      </span>
                                    </td>
                                  </tr>
                                ))}
                            </tbody>
                          </table>
                        </div>
                      </>
                    )}
                  </div>
                )}

                <button
                  className="w-full rounded-xl bg-emerald-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-60"
                  disabled={migrationBusy || selectedTables.length === 0}
                  onClick={() => void startMigration()}
                  type="button"
                >
                  {migrationBusy
                    ? options.dryRun
                      ? "Verification running..."
                      : "Migration running..."
                    : options.dryRun
                      ? "Run Verification"
                      : "Start Migration"}
                </button>
              </div>
              {migrationError && (
                <p className="mt-3 text-sm font-medium text-red-600">{migrationError}</p>
              )}
            </div>
          </section>
        )}

        {runReadyToShow && (
          <section className="card-surface p-5">
            <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
              <div>
                <h2 className="text-lg font-semibold text-slate-900">
                  {tr("5. Run Status", "5. 실행 상태")}{" "}
                  {runDryRun ? tr("(Dry-run)", "(드라이런)") : ""}
                </h2>
                <p className="mt-1 text-sm text-slate-600">
                  {tr("Session:", "세션:")} {runSessionId || tr("untracked", "미추적")} ·{" "}
                  {wsStatusLabel(wsStatus, locale)} · {tr("Target", "대상")}{" "}
                  {objectGroupModeEnabled
                    ? reportSummary?.object_group ?? effectiveObjectGroup
                    : "all"}
                </p>
              </div>
              <span className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-semibold text-slate-700">
                {runDoneTables} / {runTotalTables} {tr("done", "완료")}
              </span>
            </div>

            <div className="mb-4 rounded-xl border border-slate-200 bg-slate-50 p-3">
              <div className="mb-1 flex items-center justify-between text-xs font-semibold text-slate-600">
                <span>{tr("Overall progress", "전체 진행률")}</span>
                <span>{overallPercent}%</span>
              </div>
              <div className="h-3 rounded-full bg-slate-200">
                <div
                  className="h-3 rounded-full bg-brand-600 transition-all"
                  style={{ width: `${overallPercent}%` }}
                />
              </div>
            </div>

            <div className="grid gap-3 md:grid-cols-4">
              <div className="rounded-xl border border-slate-200 bg-white p-3">
                <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                  Success
                </p>
                <p className="mt-1 text-xl font-bold text-emerald-700">{runSuccessCount}</p>
              </div>
              <div className="rounded-xl border border-slate-200 bg-white p-3">
                <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                  Failed
                </p>
                <p className="mt-1 text-xl font-bold text-red-700">{runFailCount}</p>
              </div>
              <div className="rounded-xl border border-slate-200 bg-white p-3">
                <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                  Warnings
                </p>
                <p className="mt-1 text-xl font-bold text-amber-700">{warnings.length}</p>
              </div>
              <div className="rounded-xl border border-slate-200 bg-white p-3">
                <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                  Rows
                </p>
                <p className="mt-1 text-xl font-bold text-slate-900">
                  {processedRows.toLocaleString()}
                </p>
              </div>
            </div>

            <div className="mt-3 grid gap-3 md:grid-cols-4">
              <div className="rounded-xl border border-slate-200 bg-white p-3">
                <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                  Elapsed
                </p>
                <p className="mt-1 text-base font-bold text-slate-900">{elapsedSeconds}s</p>
              </div>
              <div className="rounded-xl border border-slate-200 bg-white p-3">
                <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                  Speed
                </p>
                <p className="mt-1 text-base font-bold text-slate-900">
                  {rowsPerSecond.toLocaleString()} rows/s
                </p>
              </div>
              <div className="rounded-xl border border-slate-200 bg-white p-3">
                <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                  ETA
                </p>
                <p className="mt-1 text-base font-bold text-slate-900">
                  {etaSeconds === null ? "-" : `${etaSeconds}s`}
                </p>
              </div>
              <div className="rounded-xl border border-slate-200 bg-white p-3">
                <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                  CPU / MEM
                </p>
                <p className="mt-1 text-base font-bold text-slate-900">
                  {metrics.cpu} / {metrics.mem}
                </p>
              </div>
            </div>

            {objectGroupModeEnabled && groupSummary && (
              <div className="mt-4 grid gap-3 lg:grid-cols-2">
                <div className="rounded-xl border border-slate-200 bg-white p-4">
                  <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                    Tables Group
                  </p>
                  <p className="mt-1 text-sm text-slate-600">
                    {groupSummary.tables.success_count} ok · {groupSummary.tables.error_count} error
                    {groupSummary.tables.skipped_count > 0
                      ? ` · ${groupSummary.tables.skipped_count} skipped`
                      : ""}
                  </p>
                  <p className="mt-2 text-xl font-bold text-slate-900">
                    {groupSummary.tables.total_rows?.toLocaleString() ?? "0"} rows
                  </p>
                </div>
                <div className="rounded-xl border border-slate-200 bg-white p-4">
                  <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
                    Sequences Group
                  </p>
                  <p className="mt-1 text-sm text-slate-600">
                    {groupSummary.sequences.success_count} ok · {groupSummary.sequences.error_count} error
                    {groupSummary.sequences.skipped_count > 0
                      ? ` · ${groupSummary.sequences.skipped_count} skipped`
                      : ""}
                  </p>
                  <p className="mt-2 text-xl font-bold text-slate-900">
                    {groupSummary.sequences.total_items.toLocaleString()} objects
                  </p>
                </div>
              </div>
            )}

            {warnings.length > 0 && (
              <div className="mt-4 rounded-xl border border-amber-200 bg-amber-50 p-3">
                <p className="text-sm font-semibold text-amber-800">Warnings</p>
                <ul className="mt-2 list-disc space-y-1 pl-5 text-sm text-amber-900">
                  {warnings.slice(0, 8).map((warning) => (
                    <li key={warning}>{warning}</li>
                  ))}
                </ul>
              </div>
            )}

            {ddlEvents.length > 0 && (
              <div className="mt-4 rounded-xl border border-slate-200 bg-white p-3">
                <p className="text-sm font-semibold text-slate-800">DDL Events</p>
                <div className="mt-2 max-h-40 space-y-1 overflow-auto text-xs">
                  {ddlEvents.map((event) => (
                    <p key={event.key}>
                      <span className="font-semibold">[{event.object}]</span> {event.name} ·{" "}
                      {event.status}
                      {event.error ? ` · ${event.error}` : ""}
                    </p>
                  ))}
                </div>
              </div>
            )}

            {Object.keys(validation).length > 0 && (
              <div className="mt-4 rounded-xl border border-slate-200 bg-white p-3">
                <p className="text-sm font-semibold text-slate-800">Validation</p>
                <div className="mt-2 overflow-auto">
                  <table className="w-full border-collapse text-sm">
                    <thead>
                      <tr>
                        <th className="border-b border-slate-200 px-2 py-1 text-left">Table</th>
                        <th className="border-b border-slate-200 px-2 py-1 text-right">Source</th>
                        <th className="border-b border-slate-200 px-2 py-1 text-right">Target</th>
                        <th className="border-b border-slate-200 px-2 py-1 text-left">Status</th>
                      </tr>
                    </thead>
                    <tbody>
                      {Object.entries(validation).map(([table, item]) => (
                        <tr className="border-b border-slate-100 last:border-b-0" key={table}>
                          <td className="px-2 py-1">{table}</td>
                          <td className="px-2 py-1 text-right">
                            {item.sourceCount.toLocaleString()}
                          </td>
                          <td className="px-2 py-1 text-right">
                            {item.targetCount.toLocaleString()}
                          </td>
                          <td className="px-2 py-1">
                            {item.status}
                            {item.message ? ` · ${item.message}` : ""}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}

            <div className="mt-4 space-y-2">
              {runEntries.map(([table, item]) => {
                const pct =
                  item.total > 0
                    ? Math.min(100, Math.floor((item.count / item.total) * 100))
                    : item.status === "completed" || item.status === "error"
                      ? 100
                      : 0;
                const barClass =
                  item.status === "completed"
                    ? "bg-emerald-500"
                    : item.status === "error"
                      ? "bg-red-500"
                      : "bg-brand-600";

                return (
                  <div className="rounded-xl border border-slate-200 bg-white p-3" key={table}>
                    <div className="mb-1 flex items-center justify-between gap-2 text-sm">
                      <span className="font-semibold text-slate-800">{table}</span>
                      <span
                        className={`rounded-full px-2 py-0.5 text-xs font-semibold ${
                          item.status === "completed"
                            ? "bg-emerald-100 text-emerald-700"
                            : item.status === "error"
                              ? "bg-red-100 text-red-700"
                              : item.status === "running"
                                ? "bg-blue-100 text-blue-700"
                                : "bg-slate-100 text-slate-600"
                        }`}
                      >
                        {tableStatusLabel(item.status, locale)}
                      </span>
                    </div>
                    <div className="h-2 rounded-full bg-slate-200">
                      <div
                        className={`h-2 rounded-full transition-all ${barClass}`}
                        style={{ width: `${pct}%` }}
                      />
                    </div>
                    <p className="mt-1 text-xs text-slate-600">
                      {pct}% · {item.count.toLocaleString()} / {item.total.toLocaleString()} rows
                    </p>
                    {item.error && (
                      <p className="mt-1 text-xs font-medium text-red-600">
                        {item.error}
                        {item.details ? ` · ${item.details}` : ""}
                      </p>
                    )}
                  </div>
                );
              })}
            </div>

            {!migrationBusy && runStartedAt !== null && (
              <div className="mt-5 flex flex-wrap gap-2">
                {!runDryRun && zipFileId && (
                  <a
                    className="rounded-lg bg-brand-600 px-3 py-2 text-sm font-semibold text-white hover:bg-brand-700"
                    href={`/api/download/${zipFileId}`}
                  >
                    Download ZIP
                  </a>
                )}
                {reportSummary?.report_id && (
                  <a
                    className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-100"
                    href={`/api/report/${reportSummary.report_id}`}
                  >
                    Download Report
                  </a>
                )}
                <button
                  className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-100"
                  onClick={resetRunState}
                  type="button"
                >
                  Clear Run Board
                </button>
              </div>
            )}
          </section>
        )}
      </div>

      {credentialsPanelOpen && (
        <aside className="fixed inset-y-0 right-0 z-30 w-full max-w-md border-l border-slate-200 bg-white p-5 shadow-2xl">
          <div className="mb-3 flex items-center justify-between">
            <h3 className="text-lg font-semibold text-slate-900">Saved Connections</h3>
            <button
              className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-100"
              onClick={() => setCredentialsPanelOpen(false)}
              type="button"
            >
              Close
            </button>
          </div>
          <div className="mb-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-600">
            Filter:{" "}
            {credentialFilter === "all"
              ? "All"
              : credentialFilter === "source"
                ? "Source only (Oracle)"
                : "Target only (non-Oracle)"}
          </div>
          <div className="mb-4 flex gap-2">
            {(["all", "source", "target"] as RoleFilter[]).map((role) => (
              <button
                className={`rounded-lg px-3 py-1.5 text-xs font-semibold ${
                  credentialFilter === role
                    ? "bg-brand-600 text-white"
                    : "border border-slate-300 text-slate-700 hover:bg-slate-100"
                }`}
                key={role}
                onClick={() => setCredentialFilter(role)}
                type="button"
              >
                {role === "all" ? "All" : role === "source" ? "Source" : "Target"}
              </button>
            ))}
          </div>
          {credentialsBusy && <p className="text-sm text-slate-600">Loading...</p>}
          {credentialsError && <p className="text-sm text-red-600">{credentialsError}</p>}
          {!credentialsBusy && !credentialsError && filteredCredentials.length === 0 && (
            <p className="text-sm text-slate-500">
              {credentialFilter === "source"
                ? "No saved source connections found."
                : credentialFilter === "target"
                  ? "No saved target connections found."
                  : "No saved connections found."}
            </p>
          )}
          <div className="space-y-3">
            {filteredCredentials.map((item) => (
              <div className="rounded-xl border border-slate-200 p-3" key={item.id}>
                <p className="text-sm font-semibold text-slate-900">{item.alias}</p>
                <p className="text-xs text-slate-500">
                  {item.dbType === "oracle" ? "Source" : "Target"} · {item.dbType}
                </p>
                <p className="mt-1 break-all text-xs text-slate-700">{item.host}</p>
                <button
                  className="mt-3 rounded-lg bg-brand-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-brand-700"
                  onClick={() => applyCredential(item)}
                  type="button"
                >
                  Apply to form
                </button>
              </div>
            ))}
          </div>
        </aside>
      )}

      {historyPanelOpen && (
        <aside className="fixed inset-y-0 right-0 z-30 w-full max-w-md border-l border-slate-200 bg-white p-5 shadow-2xl">
          <div className="mb-3 flex items-center justify-between">
            <h3 className="text-lg font-semibold text-slate-900">My History</h3>
            <button
              className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-100"
              onClick={() => setHistoryPanelOpen(false)}
              type="button"
            >
              Close
            </button>
          </div>
          {historyBusy && <p className="text-sm text-slate-600">Loading history...</p>}
          {historyError && <p className="text-sm text-red-600">{historyError}</p>}
          {!historyBusy && !historyError && history.length === 0 && (
            <p className="text-sm text-slate-500">No migration history yet.</p>
          )}
          <div className="space-y-3">
            {history.map((entry) => (
              <div className="rounded-xl border border-slate-200 p-3" key={entry.id}>
                <p className="text-sm font-semibold text-slate-900">{entry.status}</p>
                <p className="text-xs text-slate-500">
                  {new Date(entry.createdAt).toLocaleString()}
                </p>
                <p className="mt-1 text-xs text-slate-700">{entry.sourceSummary}</p>
                <p className="text-xs text-slate-700">{entry.targetSummary}</p>
                <button
                  className="mt-3 rounded-lg bg-brand-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-brand-700"
                  onClick={() => void replayHistory(entry.id)}
                  type="button"
                >
                  Replay into form
                </button>
              </div>
            ))}
          </div>
        </aside>
      )}

      {meta?.authEnabled && !user && (
        <div className="fixed inset-0 z-40 flex items-center justify-center bg-slate-900/45 px-4">
          <form className="card-surface w-full max-w-sm p-6" onSubmit={handleLogin}>
            <h3 className="text-xl font-semibold text-slate-900">Sign in</h3>
            <p className="mt-1 text-sm text-slate-600">
              Auth mode is enabled. Log in to use saved connections and history.
            </p>
            <label className="mt-4 block text-sm">
              <span className="mb-1 block text-slate-700">Username</span>
              <input
                className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                onChange={(event) =>
                  setLoginForm((prev) => ({ ...prev, username: event.target.value }))
                }
                required
                value={loginForm.username}
              />
            </label>
            <label className="mt-3 block text-sm">
              <span className="mb-1 block text-slate-700">Password</span>
              <input
                className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                onChange={(event) =>
                  setLoginForm((prev) => ({ ...prev, password: event.target.value }))
                }
                required
                type="password"
                value={loginForm.password}
              />
            </label>
            {loginError && <p className="mt-3 text-sm font-medium text-red-600">{loginError}</p>}
            <button
              className="mt-4 w-full rounded-xl bg-brand-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-brand-700 disabled:cursor-not-allowed disabled:opacity-60"
              disabled={loginBusy}
              type="submit"
            >
              {loginBusy ? "Signing in..." : "Sign in"}
            </button>
          </form>
        </div>
      )}

      {notice && (
        <div className="fixed bottom-4 right-4 z-50">
          <div
            className={`rounded-xl px-4 py-3 text-sm font-semibold shadow-lg ${
              notice.tone === "error"
                ? "bg-red-600 text-white"
                : "bg-slate-900 text-white"
            }`}
          >
            {notice.text}
          </div>
        </div>
      )}
    </div>
  );
}
