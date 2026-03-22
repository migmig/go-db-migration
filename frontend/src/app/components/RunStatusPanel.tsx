import { useState, useMemo } from "react";
import { MetricsState, TableRunState, ValidationState } from "../types";
import { DdlEvent, ReportSummary } from "../types";
import { tableStatusLabel } from "../utils";

type RunStatusPanelProps = {
  tr: (en: string, ko: string) => string;
  locale: "en" | "ko";
  runDryRun: boolean;
  runSessionId: string;
  wsStatusText: string;
  runDoneTables: number;
  runTotalTables: number;
  overallPercent: number;
  runSuccessCount: number;
  runFailCount: number;
  warnings: string[];
  processedRows: number;
  elapsedSeconds: number;
  rowsPerSecond: number;
  etaSeconds: number | null;
  metrics: MetricsState;
  ddlEvents: DdlEvent[];
  runEntries: Array<[string, TableRunState]>;
  migrationBusy: boolean;
  runStartedAt: number | null;
  zipFileId: string;
  onResetRunState: () => void;
  // Keep these but make optional/unused if we don't display them yet to satisfy TS
  objectGroupModeEnabled?: boolean;
  effectiveObjectGroup?: string;
  reportSummary?: ReportSummary | null;
  groupSummary?: ReportSummary["stats"] | null;
  validation?: Record<string, ValidationState>;
};

function CircularProgress({ percent, size = 120, strokeWidth = 10 }: { percent: number; size?: number; strokeWidth?: number }) {
  const radius = (size - strokeWidth) / 2;
  const circumference = radius * 2 * Math.PI;
  const offset = circumference - (percent / 100) * circumference;

  return (
    <div className="relative flex items-center justify-center" style={{ width: size, height: size }}>
      <svg width={size} height={size} className="rotate-[-90deg]">
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          stroke="currentColor"
          strokeWidth={strokeWidth}
          fill="transparent"
          className="text-slate-200 dark:text-slate-700"
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          stroke="currentColor"
          strokeWidth={strokeWidth}
          fill="transparent"
          strokeDasharray={circumference}
          style={{ strokeDashoffset: offset, transition: "stroke-dashoffset 0.5s ease" }}
          strokeLinecap="round"
          className="text-brand-600 dark:text-brand-500"
        />
      </svg>
      <div className="absolute inset-0 flex flex-col items-center justify-center">
        <span className="text-2xl font-bold text-slate-900 dark:text-slate-100">{percent}%</span>
        <span className="text-[10px] font-bold uppercase text-slate-700 dark:text-slate-300">Done</span>
      </div>
    </div>
  );
}

export function RunStatusPanel({
  tr,
  locale,
  runDryRun,
  runSessionId,
  wsStatusText,
  runDoneTables,
  runTotalTables,
  overallPercent,
  runSuccessCount,
  runFailCount,
  warnings,
  processedRows,
  elapsedSeconds,
  rowsPerSecond,
  etaSeconds,
  metrics,
  ddlEvents,
  runEntries,
  migrationBusy,
  runStartedAt,
  zipFileId,
  onResetRunState,
}: RunStatusPanelProps) {
  const [searchTerm, setSearchTerm] = useState("");

  const filteredEntries = useMemo(() => {
    if (!searchTerm) return runEntries;
    const lower = searchTerm.toLowerCase();
    return runEntries.filter(([table]) => table.toLowerCase().includes(lower));
  }, [runEntries, searchTerm]);

  const formatTime = (s: number | null) => {
    if (s === null) return "--:--";
    const mins = Math.floor(s / 60);
    const secs = s % 60;
    return `${mins}:${secs.toString().padStart(2, "0")}`;
  };

  return (
    <section className="card-surface p-6 dark:bg-slate-800 dark:border-slate-700">
      {/* Header Info */}
      <div className="mb-6 flex flex-wrap items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h2 className="text-xl font-bold text-slate-900 dark:text-slate-100">
              {tr("Migration Monitor", "마이그레이션 모니터링")}
            </h2>
            {runDryRun && (
              <span className="rounded-md bg-amber-100 px-2 py-0.5 text-xs font-bold text-amber-700 dark:bg-amber-900/30 dark:text-amber-400">
                DRY-RUN
              </span>
            )}
          </div>
          <p className="mt-1 text-sm text-slate-700 dark:text-slate-300">
            {tr("ID:", "세션 ID:")} <span className="font-mono text-slate-700 dark:text-slate-300">{runSessionId || "N/A"}</span> ·{" "}
            <span className="inline-flex items-center gap-1.5">
              <span className={`h-2 w-2 rounded-full ${wsStatusText.includes("connected") || wsStatusText.includes("연결됨") ? "bg-emerald-500 animate-pulse" : "bg-slate-400"}`}></span>
              {wsStatusText}
            </span>
          </p>
        </div>
        <div className="flex gap-2">
           <div className="rounded-lg bg-slate-100 px-3 py-2 text-center dark:bg-slate-700/50">
              <p className="text-[10px] font-bold uppercase text-slate-700 dark:text-slate-300">CPU</p>
              <p className="text-sm font-bold text-slate-700 dark:text-slate-200">{metrics?.cpu ?? 0}%</p>
           </div>
           <div className="rounded-lg bg-slate-100 px-3 py-2 text-center dark:bg-slate-700/50">
              <p className="text-[10px] font-bold uppercase text-slate-700 dark:text-slate-300">MEM</p>
              <p className="text-sm font-bold text-slate-700 dark:text-slate-200">{metrics?.mem ?? 0}MB</p>
           </div>
        </div>
      </div>

      {/* Main Dashboard Grid */}
      <div className="grid gap-6 lg:grid-cols-[auto_1fr]">
        {/* Left: Overall Chart */}
        <div className="flex flex-col items-center justify-center rounded-2xl border border-slate-200 bg-slate-50 p-6 dark:border-slate-700 dark:bg-slate-900/30">
          <CircularProgress percent={overallPercent} size={140} strokeWidth={12} />
          <div className="mt-4 text-center">
            <p className="text-sm font-bold text-slate-700 dark:text-slate-300">
              {runDoneTables} / {runTotalTables} {tr("Tables", "테이블 완료")}
            </p>
            <p className="text-xs text-slate-700 dark:text-slate-300">
              {runSuccessCount} ok · {runFailCount} failed
            </p>
          </div>
        </div>

        {/* Right: Stat Cards */}
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="rounded-2xl border border-slate-200 p-4 dark:border-slate-700">
            <p className="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300">{tr("Remaining Time", "예상 남은 시간")}</p>
            <p className="mt-2 text-3xl font-black text-slate-900 dark:text-slate-100">{formatTime(etaSeconds)}</p>
            <p className="mt-1 text-xs text-slate-700 dark:text-slate-300">{tr("Elapsed:", "경과:")} {elapsedSeconds}s</p>
          </div>
          <div className="rounded-2xl border border-slate-200 p-4 dark:border-slate-700">
            <p className="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300">{tr("Processing Speed", "현재 처리 속도")}</p>
            <p className="mt-2 text-3xl font-black text-slate-900 dark:text-slate-100">{rowsPerSecond.toLocaleString()}</p>
            <p className="mt-1 text-xs text-slate-700 dark:text-slate-300">rows / second</p>
          </div>
          <div className="rounded-2xl border border-slate-200 p-4 dark:border-slate-700">
            <p className="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300">{tr("Total Rows", "처리된 행 수")}</p>
            <p className="mt-2 text-3xl font-black text-slate-900 dark:text-slate-100">{processedRows.toLocaleString()}</p>
            <p className="mt-1 text-xs text-slate-700 dark:text-slate-300">{tr("Processed rows", "현재까지 완료된 데이터 행")}</p>
          </div>
          <div className="rounded-2xl border border-slate-200 p-4 dark:border-slate-700">
            <p className="text-xs font-bold uppercase tracking-wider text-slate-700 dark:text-slate-300">{tr("Status Summary", "실행 요약")}</p>
            <div className="mt-2 flex items-baseline gap-2">
               <span className="text-3xl font-black text-emerald-600 dark:text-emerald-500">{runSuccessCount}</span>
               <span className="text-sm font-bold text-slate-600 dark:text-slate-200">/ {runTotalTables}</span>
            </div>
            <p className="mt-1 text-xs text-slate-700 dark:text-slate-300">{runFailCount} {tr("errors encountered", "건의 오류 발생")}</p>
          </div>
        </div>
      </div>

      {/* Warnings & Events */}
      {(warnings.length > 0 || ddlEvents.length > 0) && (
        <div className="mt-6 grid gap-4 lg:grid-cols-2">
          {warnings.length > 0 && (
            <div className="rounded-xl bg-amber-50 p-4 dark:bg-amber-900/20">
              <p className="flex items-center gap-2 text-sm font-bold text-amber-800 dark:text-amber-400">
                <span>⚠️</span> {tr("Warnings", "경고")} ({warnings.length})
              </p>
              <ul className="mt-2 max-h-32 space-y-1 overflow-auto text-xs text-amber-900 dark:text-amber-300">
                {warnings.map((w, idx) => <li key={idx}>• {w}</li>)}
              </ul>
            </div>
          )}
          {ddlEvents.length > 0 && (
            <div className="rounded-xl bg-slate-50 p-4 dark:bg-slate-900/50">
              <p className="flex items-center gap-2 text-sm font-bold text-slate-800 dark:text-slate-200">
                <span>📜</span> {tr("DDL Events", "DDL 이벤트")}
              </p>
              <div className="mt-2 max-h-32 space-y-1 overflow-auto font-mono text-[10px] text-slate-600 dark:text-slate-400">
                {ddlEvents.map((e) => (
                  <div key={e.key}>[{e.object}] {e.name} → {e.status}</div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Detail Table List */}
      <div className="mt-8">
        <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
          <h3 className="text-sm font-bold text-slate-900 dark:text-slate-100">{tr("Detailed Table Progress", "상세 테이블 진행 상태")}</h3>
          <div className="w-full sm:w-64">
            <input
              type="text"
              className="w-full rounded-lg border border-slate-300 px-3 py-1.5 text-xs outline-none focus:border-brand-500 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200"
              placeholder={tr("Filter tables...", "테이블 검색...")}
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
            />
          </div>
        </div>
        <div className="max-h-[60vh] overflow-auto rounded-xl border border-slate-100 bg-slate-50/30 p-4 dark:border-slate-700/30 dark:bg-slate-900/10">
          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4 3xl:grid-cols-6">
            {filteredEntries.map(([table, item]) => {
            const pct = item.total > 0 ? Math.min(100, Math.floor((item.count / item.total) * 100)) : (item.status === "completed" ? 100 : 0);
            const isError = item.status === "error";
            const isDone = item.status === "completed";
            const isRunning = item.status === "running";

            return (
              <div key={table} className={`rounded-xl border p-3 transition-all ${
                isError ? "border-red-200 bg-red-50 dark:border-red-900/50 dark:bg-red-900/10" :
                isDone ? "border-emerald-100 bg-emerald-50/30 dark:border-emerald-900/30 dark:bg-emerald-900/5" :
                "border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-800"
              }`}>
                <div className="mb-2 flex items-center justify-between gap-2">
                  <span className="truncate text-xs font-bold text-slate-800 dark:text-slate-200" title={table}>{table}</span>
                  <span className={`rounded-full px-2 py-0.5 text-[10px] font-bold uppercase ${
                    isDone ? "bg-emerald-500 text-white" :
                    isError ? "bg-red-500 text-white" :
                    isRunning ? "bg-blue-500 text-white animate-pulse" :
                    "bg-slate-200 text-slate-600 dark:bg-slate-700 dark:text-slate-600 dark:text-slate-200"
                  }`}>
                    {tableStatusLabel(item.status, locale)}
                  </span>
                </div>
                <div className="h-1.5 rounded-full bg-slate-100 dark:bg-slate-700">
                  <div
                    className={`h-1.5 rounded-full transition-all duration-500 ${
                      isError ? "bg-red-500" : isDone ? "bg-emerald-500" : "bg-brand-500"
                    }`}
                    style={{ width: `${pct}%` }}
                  />
                </div>
                <div className="mt-2 flex items-center justify-between text-[10px] font-medium text-slate-700 dark:text-slate-300">
                  <span>{pct}%</span>
                  <span>{item.count.toLocaleString()} / {item.total.toLocaleString()}</span>
                </div>
                {item.error && (
                  <p className="mt-1 truncate text-[10px] font-semibold text-red-600 dark:text-red-400" title={item.error}>{item.error}</p>
                )}
              </div>
            );
          })}
          </div>
        </div>
      </div>

      {/* Footer Actions */}
      {!migrationBusy && runStartedAt !== null && (
        <div className="mt-8 flex flex-wrap items-center justify-center gap-3 border-t border-slate-100 pt-6 dark:border-slate-700">
          {!runDryRun && zipFileId && (
            <a
              className="rounded-xl bg-brand-600 px-6 py-2.5 text-sm font-bold text-white shadow-lg shadow-brand-600/20 transition-all hover:bg-brand-700 hover:scale-[1.02]"
              href={`/api/download/${zipFileId}`}
            >
              {tr("Download Exported SQL (ZIP)", "내보낸 SQL 다운로드")}
            </a>
          )}
          <button
            className="rounded-xl border border-slate-300 bg-white px-6 py-2.5 text-sm font-bold text-slate-700 transition-all hover:bg-slate-50 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200 dark:hover:bg-slate-700"
            onClick={onResetRunState}
            type="button"
          >
            {tr("Close Monitoring", "모니터링 종료")}
          </button>
        </div>
      )}
    </section>
  );
}
