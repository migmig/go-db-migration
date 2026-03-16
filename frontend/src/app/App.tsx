import { FormEvent, useEffect, useState } from "react";
import { apiRequest } from "../shared/api/client";
import {
  AuthUser,
  Credential,
  HistoryEntry,
  HistoryListResponse,
  RuntimeMeta,
} from "../shared/api/types";

type RoleFilter = "all" | "source" | "target";
type NoticeTone = "info" | "error";

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

const SOURCE_RECENT_KEY = "dbm:v16:source-recent";
const SOURCE_REMEMBER_KEY = "dbm:v16:source-remember-pass";

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

export function App() {
  const initialRememberPass = loadRememberPassword();
  const initialRecent = loadSourceRecent();

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
    mode: "file",
    targetDb: "postgres",
    targetUrl: "",
    schema: "public",
  });

  const [rememberSourcePassword, setRememberSourcePassword] =
    useState(initialRememberPass);
  const [sourceConnectBusy, setSourceConnectBusy] = useState(false);
  const [sourceConnectError, setSourceConnectError] = useState("");
  const [tableCount, setTableCount] = useState<number | null>(null);
  const [tablePreview, setTablePreview] = useState<string[]>([]);

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

  const [loginForm, setLoginForm] = useState({ username: "", password: "" });
  const [loginBusy, setLoginBusy] = useState(false);
  const [loginError, setLoginError] = useState("");

  const [notice, setNotice] = useState<{ text: string; tone: NoticeTone } | null>(
    null,
  );

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
    void boot();
  }, []);

  const filteredCredentials = credentials.filter((item) => {
    if (credentialFilter === "all") return true;
    if (credentialFilter === "source") return item.dbType === "oracle";
    return item.dbType !== "oracle";
  });

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
      const { response, data } = await apiRequest<HistoryListResponse>(
        "/api/history?page=1&pageSize=10",
      );
      if (!response.ok) {
        throw new Error("Failed to load migration history.");
      }
      setHistory(data.items ?? []);
    } catch (error) {
      setHistoryError(
        error instanceof Error ? error.message : "Failed to load migration history.",
      );
    } finally {
      setHistoryBusy(false);
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
      setNotice({ text: "History payload applied to forms.", tone: "info" });
    } catch (error) {
      setHistoryError(error instanceof Error ? error.message : "Replay failed.");
    }
  }

  function applyReplayPayload(payload: Record<string, unknown>) {
    const direct = Boolean(payload.direct);
    setSource((prev) => ({
      ...prev,
      oracleUrl: String(payload.oracleUrl ?? ""),
      username: String(payload.username ?? ""),
      password: "",
      like: "",
    }));
    setTarget((prev) => ({
      ...prev,
      mode: direct ? "direct" : "file",
      targetDb: String(payload.targetDb ?? prev.targetDb),
      targetUrl: String(payload.targetUrl ?? payload.pgUrl ?? ""),
      schema: String(payload.schema ?? ""),
    }));
  }

  async function connectSource() {
    setSourceConnectError("");
    setTableCount(null);
    setTablePreview([]);
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
      setTableCount(tables.length);
      setTablePreview(tables.slice(0, 8));
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

  if (booting) {
    return (
      <div className="flex min-h-screen items-center justify-center text-slate-700">
        Loading v16 preview...
      </div>
    );
  }

  if (bootError) {
    return (
      <div className="mx-auto flex min-h-screen max-w-3xl items-center px-6 py-12">
        <div className="card-surface w-full p-8">
          <h1 className="text-xl font-semibold text-slate-900">v16 boot failed</h1>
          <p className="mt-3 text-sm text-red-600">{bootError}</p>
          <button
            className="mt-5 rounded-lg bg-brand-600 px-4 py-2 text-sm font-semibold text-white"
            onClick={() => void boot()}
            type="button"
          >
            Retry
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
            <h1 className="text-2xl font-bold text-slate-900">v16 Connection Workspace</h1>
            <p className="mt-1 text-sm text-slate-600">
              Unified flow for source/target setup with saved connections.
            </p>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            {meta?.authEnabled ? (
              <>
                <span className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-semibold text-slate-700">
                  {user ? `User: ${user.username}` : "Auth enabled"}
                </span>
                <button
                  className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
                  onClick={() => void openCredentialsPanel("all")}
                  type="button"
                >
                  Saved Connections
                </button>
                <button
                  className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
                  onClick={() => void openHistoryPanel()}
                  type="button"
                >
                  My History
                </button>
                <button
                  className="rounded-lg bg-slate-900 px-3 py-2 text-sm font-semibold text-white hover:bg-slate-700"
                  onClick={() => void handleLogout()}
                  type="button"
                >
                  Logout
                </button>
              </>
            ) : (
              <span className="rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-700">
                Auth disabled
              </span>
            )}
          </div>
        </header>

        <details className="card-surface p-4">
          <summary className="cursor-pointer text-sm font-semibold text-slate-800">
            Recent source input (optional)
          </summary>
          <div className="mt-4 flex flex-wrap items-center gap-2">
            <label className="inline-flex items-center gap-2 text-sm text-slate-700">
              <input
                checked={rememberSourcePassword}
                onChange={(event) => setRememberSourcePassword(event.target.checked)}
                type="checkbox"
              />
              Remember source password
            </label>
            <button
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
              onClick={restoreRecentSource}
              type="button"
            >
              Restore
            </button>
            <button
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
              onClick={clearRecentSource}
              type="button"
            >
              Clear
            </button>
          </div>
        </details>

        <section className="grid gap-5 lg:grid-cols-2">
          <div className="card-surface p-5">
            <div className="mb-4 flex items-center justify-between gap-3">
              <h2 className="text-lg font-semibold text-slate-900">1. Source (Oracle)</h2>
              {meta?.authEnabled && (
                <button
                  className="rounded-lg border border-brand-300 bg-brand-50 px-3 py-2 text-xs font-semibold text-brand-700 hover:bg-brand-100"
                  onClick={() => void openCredentialsPanel("source")}
                  type="button"
                >
                  Load Saved Source
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
                <span className="mb-1 block text-slate-700">Username</span>
                <input
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                  onChange={(event) =>
                    setSource((prev) => ({ ...prev, username: event.target.value }))
                  }
                  value={source.username}
                />
              </label>
              <label className="block text-sm">
                <span className="mb-1 block text-slate-700">Password</span>
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
                {sourceConnectBusy ? "Loading tables..." : "Connect & Fetch Tables"}
              </button>
            </div>
            {sourceConnectError && (
              <p className="mt-3 text-sm font-medium text-red-600">{sourceConnectError}</p>
            )}
            {tableCount !== null && (
              <div className="mt-4 rounded-xl border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-800">
                <p className="font-semibold">Found {tableCount} table(s)</p>
                {tablePreview.length > 0 && (
                  <p className="mt-1 text-xs text-emerald-700">
                    Preview: {tablePreview.join(", ")}
                  </p>
                )}
              </div>
            )}
          </div>

          <div className="card-surface p-5">
            <div className="mb-4 flex items-center justify-between gap-3">
              <h2 className="text-lg font-semibold text-slate-900">2. Target</h2>
              {meta?.authEnabled && (
                <button
                  className="rounded-lg border border-brand-300 bg-brand-50 px-3 py-2 text-xs font-semibold text-brand-700 hover:bg-brand-100"
                  onClick={() => void openCredentialsPanel("target")}
                  type="button"
                >
                  Load Saved Target
                </button>
              )}
            </div>
            <div className="space-y-3">
              <label className="block text-sm">
                <span className="mb-1 block text-slate-700">Migration mode</span>
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
                  <option value="file">SQL file mode</option>
                  <option value="direct">Direct migration</option>
                </select>
              </label>
              <label className="block text-sm">
                <span className="mb-1 block text-slate-700">Target DB</span>
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
                <span className="mb-1 block text-slate-700">Target URL</span>
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
                <span className="mb-1 block text-slate-700">Schema</span>
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
                {targetTestBusy ? "Testing target..." : "Test Target Connection"}
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
            <p className="text-sm text-slate-500">No credentials found for this filter.</p>
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
