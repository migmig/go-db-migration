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
import {
  DEFAULT_OPTIONS,
  SOURCE_RECENT_KEY,
  SOURCE_REMEMBER_KEY,
  TARGET_RECENT_KEY,
  UI_LOCALE_KEY,
  UI_TEXT,
} from "./constants";
import {
  CompareFilter,
  CompareState,
  DdlEvent,
  DiscoverySummary,
  Locale,
  MetricsState,
  MigrationOptions,
  NoticeTone,
  ObjectGroup,
  PrecheckDecisionFilter,
  ReportSummary,
  RoleFilter,
  SourceState,
  TableHistoryDetail,
  TableHistoryState,
  TableHistoryStatusFilter,
  TableRunState,
  TableRunStatus,
  TableSortOption,
  TargetState,
  TargetTableEntry,
  ValidationState,
  WsProgressMsg,
  WsStatus,
} from "./types";
import {
  createSessionId,
  isObjectGroupModeEnabled,
  loadLocale,
  loadRememberPassword,
  loadSourceRecent,
  loadTargetRecent,
  normalizeTableKey,
  parseReplayedTables,
  toBool,
  toNumber,
  toObjectGroup,
  toString,
  toStringArray,
  wsStatusLabel,
} from "./utils";
import { HeaderBar } from "./components/HeaderBar";
import { LoginModal } from "./components/LoginModal";
import { RecentSourcePanel } from "./components/RecentSourcePanel";
import { CredentialsPanel } from "./components/CredentialsPanel";
import { HistoryPanel } from "./components/HistoryPanel";
import { ConnectionForms } from "./components/ConnectionForms";
import { RunStatusPanel } from "./components/RunStatusPanel";
import { MigrationOptionsPanel } from "./components/MigrationOptionsPanel";
import { TableSelection } from "./components/TableSelection";

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
  const [compareState, setCompareState] = useState<CompareState>({
    targetTables: [],
    fetchedAt: null,
    busy: false,
    error: null,
  });
  const [compareFilter, setCompareFilter] = useState<CompareFilter>("all");
  const [compareSearch, setCompareSearch] = useState("");

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
      const searchTerms = tableSearch.toLowerCase().split(',').map(t => t.trim()).filter(Boolean);
      if (searchTerms.length > 0 && !searchTerms.some(term => table.toLowerCase().includes(term))) {
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
  const compareEntries = useMemo((): TargetTableEntry[] => {
    if (allTables.length === 0 || compareState.targetTables.length === 0) return [];
    const sourceSet = new Set(allTables.map((t) => t.toLowerCase()));
    const targetSet = new Set(compareState.targetTables.map((t) => t.toLowerCase()));
    const allNames: Set<string> = new Set([...sourceSet, ...targetSet]);
    const catOrder: Record<TargetTableEntry["category"], number> = {
      source_only: 0,
      both: 1,
      target_only: 2,
    };
    return Array.from(allNames)
      .map((name): TargetTableEntry => {
        const inSource = sourceSet.has(name);
        const inTarget = targetSet.has(name);
        const category: TargetTableEntry["category"] =
          inSource && inTarget ? "both" : inSource ? "source_only" : "target_only";
        const precheckRow = precheckItems.find(
          (r) => r.table_name.toLowerCase() === name,
        );
        return {
          name,
          inSource,
          inTarget,
          category,
          sourceRowCount: precheckRow?.source_row_count ?? null,
          targetRowCount: precheckRow?.target_row_count ?? null,
        };
      })
      .sort((a, b) => {
        const diff = catOrder[a.category] - catOrder[b.category];
        return diff !== 0 ? diff : a.name.localeCompare(b.name);
      });
  }, [allTables, compareState.targetTables, precheckItems]);

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
  }, [target.mode, target.targetUrl, target.schema]);

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

  async function fetchTargetTables() {
    if (!target.targetUrl || !target.schema) return;
    setCompareState((prev) => ({ ...prev, busy: true, error: null }));
    try {
      const { response, data } = await apiRequest<{
        tables?: string[];
        fetchedAt?: string;
        error?: string;
      }>("/api/target-tables", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ targetUrl: target.targetUrl, schema: target.schema }),
      });
      if (!response.ok) throw new Error(data.error ?? tr("Fetch failed", "조회 실패"));
      setCompareState({
        targetTables: data.tables ?? [],
        fetchedAt: data.fetchedAt ?? null,
        busy: false,
        error: null,
      });
    } catch (e) {
      setCompareState((prev) => ({
        ...prev,
        busy: false,
        error: e instanceof Error ? e.message : tr("Unknown error", "알 수 없는 오류"),
      }));
    }
  }

  function selectByCategory(category: TargetTableEntry["category"]) {
    const names = new Set(
      compareEntries
        .filter((e) => e.category === category)
        .map((e) => e.name.toUpperCase()),
    );
    setSelectedTables((prev) => {
      const next = new Set(prev);
      allTables.forEach((t) => {
        if (names.has(t.toUpperCase())) next.add(t);
      });
      return Array.from(next);
    });
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
        <HeaderBar
          authMeta={meta}
          locale={locale}
          onLogout={() => void handleLogout()}
          onOpenCredentials={() => void openCredentialsPanel("all")}
          onOpenHistory={() => void openHistoryPanel()}
          onToggleLocale={() => setLocale((prev) => (prev === "en" ? "ko" : "en"))}
          t={t}
          user={user}
        />

        <RecentSourcePanel
          onClear={clearRecentSource}
          onRememberSourcePasswordChange={setRememberSourcePassword}
          onRestore={restoreRecentSource}
          rememberSourcePassword={rememberSourcePassword}
          t={t}
        />

        <ConnectionForms
          allTablesCount={allTables.length}
          compareState={compareState}
          meta={meta}
          migrationBusy={migrationBusy}
          onConnectSource={() => void connectSource()}
          onFetchTargetTables={() => void fetchTargetTables()}
          onOpenSourceCredentials={() => void openCredentialsPanel("source")}
          onOpenTargetCredentials={() => void openCredentialsPanel("target")}
          onSourceFieldChange={(field, value) =>
            setSource((prev) => ({ ...prev, [field]: value }))
          }
          onTargetFieldChange={(field, value) =>
            setTarget((prev) => ({ ...prev, [field]: value }))
          }
          onTestTarget={() => void testTarget()}
          source={source}
          sourceConnectBusy={sourceConnectBusy}
          sourceConnectError={sourceConnectError}
          target={target}
          targetTestBusy={targetTestBusy}
          targetTestError={targetTestError}
          targetTestMessage={targetTestMessage}
          tr={tr}
        />

        {allTables.length > 0 && (
          <section className="grid gap-5 xl:grid-cols-[1.2fr_1fr]">
            <TableSelection
              activeHistoryDetail={activeHistoryDetail}
              activeTableHistory={activeTableHistory}
              allTables={allTables}
              allVisibleSelected={allVisibleSelected}
              compareEntries={compareEntries}
              compareFilter={compareFilter}
              compareSearch={compareSearch}
              deselectAllVisibleTables={deselectAllVisibleTables}
              discoverySummary={discoverySummary}
              excludeMigratedSuccess={excludeMigratedSuccess}
              filteredTables={filteredTables}
              historyByTable={historyByTable}
              locale={locale}
              migrationBusy={migrationBusy}
              objectGroupModeEnabled={objectGroupModeEnabled}
              openTableHistory={openTableHistory}
              previewObjectGroup={previewObjectGroup}
              previewSequences={previewSequences}
              previewTables={previewTables}
              replayHistory={replayHistory}
              selectAllVisibleTables={selectAllVisibleTables}
              selectByCategory={selectByCategory}
              selectedTableSet={selectedTableSet}
              selectedTables={selectedTables}
              setActiveTableHistory={setActiveTableHistory}
              setCompareFilter={setCompareFilter}
              setCompareSearch={setCompareSearch}
              setExcludeMigratedSuccess={setExcludeMigratedSuccess}
              setTableSearch={setTableSearch}
              setTableSort={setTableSort}
              setTableStatusFilter={setTableStatusFilter}
              tableHistoryBusy={tableHistoryBusy}
              tableHistoryError={tableHistoryError}
              tableProgress={tableProgress}
              tableSearch={tableSearch}
              tableSort={tableSort}
              tableStatusFilter={tableStatusFilter}
              toggleTable={toggleTable}
              tr={tr}
            />

            

            <MigrationOptionsPanel
              effectiveObjectGroup={effectiveObjectGroup}
              meta={meta}
              migrationBusy={migrationBusy}
              migrationError={migrationError}
              objectGroupModeEnabled={objectGroupModeEnabled}
              onApplyObjectGroupSelection={applyObjectGroupSelection}
              onRunPrecheck={() => void runPrecheck()}
              onStartMigration={() => void startMigration()}
              options={options}
              precheckBusy={precheckBusy}
              precheckDecisionFilter={precheckDecisionFilter}
              precheckError={precheckError}
              precheckItems={precheckItems}
              precheckPolicy={precheckPolicy}
              precheckSummary={precheckSummary}
              selectedTablesCount={selectedTables.length}
              setOptions={setOptions}
              setPrecheckDecisionFilter={setPrecheckDecisionFilter}
              setPrecheckPolicy={setPrecheckPolicy}
              targetMode={target.mode}
              tr={tr}
            />
          </section>
        )}

        {runReadyToShow && (
          <RunStatusPanel
            ddlEvents={ddlEvents}
            effectiveObjectGroup={effectiveObjectGroup}
            elapsedSeconds={elapsedSeconds}
            etaSeconds={etaSeconds}
            groupSummary={groupSummary}
            locale={locale}
            metrics={metrics}
            migrationBusy={migrationBusy}
            objectGroupModeEnabled={objectGroupModeEnabled}
            onResetRunState={resetRunState}
            overallPercent={overallPercent}
            processedRows={processedRows}
            reportSummary={reportSummary}
            rowsPerSecond={rowsPerSecond}
            runDoneTables={runDoneTables}
            runDryRun={runDryRun}
            runEntries={runEntries}
            runFailCount={runFailCount}
            runSessionId={runSessionId}
            runStartedAt={runStartedAt}
            runSuccessCount={runSuccessCount}
            runTotalTables={runTotalTables}
            tr={tr}
            validation={validation}
            warnings={warnings}
            wsStatusText={wsStatusLabel(wsStatus, locale)}
            zipFileId={zipFileId}
          />
        )}
      </div>

      {credentialsPanelOpen && (
        <CredentialsPanel
          credentialFilter={credentialFilter}
          credentialsBusy={credentialsBusy}
          credentialsError={credentialsError}
          filteredCredentials={filteredCredentials}
          onApply={applyCredential}
          onClose={() => setCredentialsPanelOpen(false)}
          onFilterChange={setCredentialFilter}
          tr={tr}
        />
      )}

      {historyPanelOpen && (
        <HistoryPanel
          history={history}
          historyBusy={historyBusy}
          historyError={historyError}
          onClose={() => setHistoryPanelOpen(false)}
          onReplay={(id) => void replayHistory(id)}
          tr={tr}
        />
      )}

      {meta?.authEnabled && !user && (
        <LoginModal
          loginBusy={loginBusy}
          loginError={loginError}
          loginForm={loginForm}
          onPasswordChange={(value) =>
            setLoginForm((prev) => ({ ...prev, password: value }))
          }
          onSubmit={handleLogin}
          onUsernameChange={(value) =>
            setLoginForm((prev) => ({ ...prev, username: value }))
          }
          tr={tr}
        />
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
