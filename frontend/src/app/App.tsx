import { useEffect, useMemo, useRef, useState } from "react";
import { apiRequest } from "../shared/api/client";
import {
  Credential,
  HistoryEntry,
  HistoryListResponse,
  PrecheckResponse,
  PrecheckSummary,
  PrecheckTableResult,
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
  CompareState,
  Locale,
  MigrationOptions,
  NoticeTone,
  ObjectGroup,
  PrecheckDecisionFilter,
  RoleFilter,
  SourceState,
  TableHistoryState,
  TableRunStatus,
  TargetState,
  TargetTableEntry,
} from "./types";
import {
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
import { useAuth } from "./hooks/useAuth";
import { useMigrationRun } from "./hooks/useMigrationRun";
import { useTheme } from "./hooks/useTheme";
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
  const { theme, toggleTheme } = useTheme();
  const [locale, setLocale] = useState<Locale>(() => loadLocale());
  const [currentStep, setCurrentStep] = useState<1 | 2 | 3>(1);
  const resetRunStateRef = useRef<() => void>(() => {});
  const initialRememberPass = loadRememberPassword();
  const initialRecent = loadSourceRecent();
  const initialTarget = loadTargetRecent();


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
  const [selectedTables, setSelectedTables] = useState<string[]>([]);

  const [options, setOptions] = useState<MigrationOptions>(DEFAULT_OPTIONS);

  const [targetTestBusy, setTargetTestBusy] = useState(false);
  const [targetTestError, setTargetTestError] = useState("");
  const [targetTestMessage, setTargetTestMessage] = useState("");


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

  const [precheckBusy, setPrecheckBusy] = useState(false);
  const [precheckError, setPrecheckError] = useState("");
  const [precheckSummary, setPrecheckSummary] = useState<PrecheckSummary | null>(null);
  const [precheckItems, setPrecheckItems] = useState<PrecheckTableResult[]>([]);
  const [precheckDecisionFilter, setPrecheckDecisionFilter] = useState<PrecheckDecisionFilter>("all");
  const [precheckPolicy, setPrecheckPolicy] = useState("strict");


  const [notice, setNotice] = useState<{ text: string; tone: NoticeTone } | null>(
    null,
  );




  const {
    meta,
    user,
    booting,
    bootError,
    loginForm,
    loginBusy,
    loginError,
    setLoginForm,
    boot,
    handleLogin,
    handleLogout,
  } = useAuth({
    resetRunState: () => resetRunStateRef.current(),
    setCredentialsPanelOpen,
    setHistoryPanelOpen,
    setNotice,
  });

  const {
    tableProgress,
    validation,
    ddlEvents,
    warnings,
    discoverySummary,
    metrics,
    migrationBusy,
    migrationError,
    wsStatus,
    runSessionId,
    runStartedAt,
    runEndedAt,
    runDryRun,
    zipFileId,
    reportSummary,
    clock,
    startMigration,
    resetRunState,
  } = useMigrationRun({
    options,
    source,
    target,
    selectedTables,
    effectiveObjectGroup: isObjectGroupModeEnabled(meta) ? options.objectGroup : "all",
    setNotice,
  });
  resetRunStateRef.current = resetRunState;



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


  useEffect(() => {
    if (migrationBusy || runStartedAt !== null) {
      setCurrentStep(3);
    }
  }, [migrationBusy, runStartedAt]);

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
    const timeout = setTimeout(() => {
      if (notice) {
        setNotice(null);
      }
    }, 2400);
    return () => clearTimeout(timeout);
  }, [notice]);


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
    <div className="relative min-h-screen px-4 pb-16 pt-8 sm:px-6 lg:px-10 dark:bg-slate-900 dark:text-slate-100 transition-colors duration-200">
      <div className="mx-auto flex max-w-7xl flex-col gap-6">
        <HeaderBar
          authMeta={meta}
          locale={locale}
          theme={theme}
          onLogout={() => void handleLogout()}
          onOpenCredentials={() => void openCredentialsPanel("all")}
          onOpenHistory={() => void openHistoryPanel()}
          onToggleLocale={() => setLocale((prev) => (prev === "en" ? "ko" : "en"))}
          onToggleTheme={toggleTheme}
          t={t}
          user={user}
        />

        {/* Step Indicator */}
        <div className="mb-8 mt-2 flex items-center justify-center gap-2 sm:gap-4">
          {[
            { step: 1, label: t("Connections") ?? tr("Connections", "연결 설정") },
            { step: 2, label: t("Configuration") ?? tr("Configuration", "테이블 & 옵션") },
            { step: 3, label: t("Execution") ?? tr("Execution", "실행 및 상태") },
          ].map((s, idx) => (
            <div key={s.step} className="flex items-center gap-2 sm:gap-4">
              <div
                className={`flex items-center gap-2 rounded-full px-4 py-2 text-sm font-bold transition-all ${
                  currentStep === s.step
                    ? "bg-brand-600 text-white shadow-md ring-2 ring-brand-600/30 ring-offset-2"
                    : currentStep > s.step
                    ? "bg-brand-100 text-brand-700"
                    : "bg-slate-100 text-slate-400"
                }`}
              >
                <span className={`flex h-6 w-6 items-center justify-center rounded-full text-xs ${
                  currentStep === s.step ? "bg-white/20" : "bg-transparent"
                }`}>
                  {currentStep > s.step ? "✓" : s.step}
                </span>
                <span className="hidden sm:inline-block">{s.label}</span>
              </div>
              {idx < 2 && <div className={`h-px w-6 sm:w-12 ${currentStep > s.step ? "bg-brand-300" : "bg-slate-200"}`} />}
            </div>
          ))}
        </div>

        {/* Step 1: Connections */}
        {currentStep === 1 && (
          <div className="flex animate-fade-in flex-col gap-6">
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
            <div className="mt-2 flex justify-end">
              <button
                type="button"
                onClick={() => setCurrentStep(2)}
                disabled={allTables.length === 0}
                className="rounded-xl bg-brand-600 px-8 py-3 text-sm font-bold text-white shadow-sm transition-all hover:bg-brand-700 disabled:cursor-not-allowed disabled:opacity-50"
              >
                {tr("Next: Select Tables", "다음: 테이블 선택")} →
              </button>
            </div>
          </div>
        )}

        {/* Step 2: Configuration */}
        {currentStep === 2 && (
          <div className="flex animate-fade-in flex-col gap-6">
            <section className="grid gap-5 xl:grid-cols-[1.2fr_1fr]">
              <TableSelection
                tr={tr}
                allTables={allTables}
                selectedTables={selectedTables}
                setSelectedTables={setSelectedTables}
                objectGroupModeEnabled={objectGroupModeEnabled}
                previewTables={previewTables}
                discoverySummary={discoverySummary}
                previewObjectGroup={previewObjectGroup}
                previewSequences={previewSequences}
                compareEntries={compareEntries}
                migrationBusy={migrationBusy}
                selectByCategory={selectByCategory}
                tableProgress={tableProgress}
                historyByTable={historyByTable}
                history={history}
                openTableHistory={openTableHistory}
                activeTableHistory={activeTableHistory}
                setActiveTableHistory={setActiveTableHistory}
                tableHistoryBusy={tableHistoryBusy}
                tableHistoryError={tableHistoryError}
                replayHistory={replayHistory}
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
            <div className="mt-2 flex justify-between">
              <button
                type="button"
                onClick={() => setCurrentStep(1)}
                className="rounded-xl bg-slate-200 px-8 py-3 text-sm font-bold text-slate-700 transition-all hover:bg-slate-300"
              >
                ← {tr("Back", "이전 단계로")}
              </button>
            </div>
          </div>
        )}

        {/* Step 3: Execution */}
        {currentStep === 3 && runReadyToShow && (
          <div className="flex animate-fade-in flex-col gap-6">
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
            {!migrationBusy && runEndedAt !== null && (
              <div className="mt-4 flex justify-center">
                <button
                  type="button"
                  onClick={() => {
                    resetRunState();
                    setCurrentStep(2);
                  }}
                  className="rounded-xl bg-slate-800 px-8 py-3 text-sm font-bold text-white transition-all hover:bg-slate-700 shadow-md"
                >
                  {tr("Configure New Migration", "새 마이그레이션 설정")}
                </button>
              </div>
            )}
          </div>
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
