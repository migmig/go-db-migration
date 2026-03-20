import { MetricsState, TableRunState, ValidationState } from "../types";
import { DdlEvent, ReportSummary } from "../types";
import { tableStatusLabel } from "../utils";

type RunStatusPanelProps = {
  tr: (en: string, ko: string) => string;
  locale: "en" | "ko";
  runDryRun: boolean;
  runSessionId: string;
  wsStatusText: string;
  objectGroupModeEnabled: boolean;
  effectiveObjectGroup: string;
  reportSummary: ReportSummary | null;
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
  groupSummary: ReportSummary["stats"] | null;
  ddlEvents: DdlEvent[];
  validation: Record<string, ValidationState>;
  runEntries: Array<[string, TableRunState]>;
  migrationBusy: boolean;
  runStartedAt: number | null;
  zipFileId: string;
  onResetRunState: () => void;
};

export function RunStatusPanel({
  tr,
  locale,
  runDryRun,
  runSessionId,
  wsStatusText,
  objectGroupModeEnabled,
  effectiveObjectGroup,
  reportSummary,
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
  groupSummary,
  ddlEvents,
  validation,
  runEntries,
  migrationBusy,
  runStartedAt,
  zipFileId,
  onResetRunState,
}: RunStatusPanelProps) {
  return (
    <section className="card-surface p-5">
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-lg font-semibold text-slate-900">
            {tr("5. Run Status", "5. 실행 상태")}{" "}
            {runDryRun ? tr("(Dry-run)", "(드라이런)") : ""}
          </h2>
          <p className="mt-1 text-sm text-slate-600">
            {tr("Session:", "세션:")} {runSessionId || tr("untracked", "미추적")} ·{" "}
            {wsStatusText} · {tr("Target", "대상")}{" "}
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
            {tr("Success", "성공")}
          </p>
          <p className="mt-1 text-xl font-bold text-emerald-700">{runSuccessCount}</p>
        </div>
        <div className="rounded-xl border border-slate-200 bg-white p-3">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            {tr("Failed", "실패")}
          </p>
          <p className="mt-1 text-xl font-bold text-red-700">{runFailCount}</p>
        </div>
        <div className="rounded-xl border border-slate-200 bg-white p-3">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            {tr("Warnings", "경고")}
          </p>
          <p className="mt-1 text-xl font-bold text-amber-700">{warnings.length}</p>
        </div>
        <div className="rounded-xl border border-slate-200 bg-white p-3">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            {tr("Rows", "행 수")}
          </p>
          <p className="mt-1 text-xl font-bold text-slate-900">
            {processedRows.toLocaleString()}
          </p>
        </div>
      </div>

      <div className="mt-3 grid gap-3 md:grid-cols-4">
        <div className="rounded-xl border border-slate-200 bg-white p-3">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            {tr("Elapsed", "경과 시간")}
          </p>
          <p className="mt-1 text-base font-bold text-slate-900">{elapsedSeconds}s</p>
        </div>
        <div className="rounded-xl border border-slate-200 bg-white p-3">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            {tr("Speed", "처리 속도")}
          </p>
          <p className="mt-1 text-base font-bold text-slate-900">
            {rowsPerSecond.toLocaleString()} rows/s
          </p>
        </div>
        <div className="rounded-xl border border-slate-200 bg-white p-3">
          <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
            {tr("ETA", "예상 완료")}
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
              {tr("Tables Group", "테이블 그룹")}
            </p>
            <p className="mt-1 text-sm text-slate-600">
              {groupSummary.tables.success_count} {tr("ok", "성공")} · {groupSummary.tables.error_count} {tr("error", "오류")}
              {groupSummary.tables.skipped_count > 0
                ? ` · ${groupSummary.tables.skipped_count} ${tr("skipped", "건너뜀")}`
                : ""}
            </p>
            <p className="mt-2 text-xl font-bold text-slate-900">
              {groupSummary.tables.total_rows?.toLocaleString() ?? "0"} {tr("rows", "행")}
            </p>
          </div>
          <div className="rounded-xl border border-slate-200 bg-white p-4">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">
              {tr("Sequences Group", "시퀀스 그룹")}
            </p>
            <p className="mt-1 text-sm text-slate-600">
              {groupSummary.sequences.success_count} {tr("ok", "성공")} · {groupSummary.sequences.error_count} {tr("error", "오류")}
              {groupSummary.sequences.skipped_count > 0
                ? ` · ${groupSummary.sequences.skipped_count} ${tr("skipped", "건너뜀")}`
                : ""}
            </p>
            <p className="mt-2 text-xl font-bold text-slate-900">
              {groupSummary.sequences.total_items.toLocaleString()} {tr("objects", "객체")}
            </p>
          </div>
        </div>
      )}

      {warnings.length > 0 && (
        <div className="mt-4 rounded-xl border border-amber-200 bg-amber-50 p-3">
          <p className="text-sm font-semibold text-amber-800">{tr("Warnings", "경고")}</p>
          <ul className="mt-2 list-disc space-y-1 pl-5 text-sm text-amber-900">
            {warnings.slice(0, 8).map((warning) => (
              <li key={warning}>{warning}</li>
            ))}
          </ul>
        </div>
      )}

      {ddlEvents.length > 0 && (
        <div className="mt-4 rounded-xl border border-slate-200 bg-white p-3">
          <p className="text-sm font-semibold text-slate-800">{tr("DDL Events", "DDL 이벤트")}</p>
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
          <p className="text-sm font-semibold text-slate-800">{tr("Validation", "검증")}</p>
          <div className="mt-2 overflow-auto">
            <table className="w-full border-collapse text-sm">
              <thead>
                <tr>
                  <th className="border-b border-slate-200 px-2 py-1 text-left">{tr("Table", "테이블")}</th>
                  <th className="border-b border-slate-200 px-2 py-1 text-right">{tr("Source", "소스")}</th>
                  <th className="border-b border-slate-200 px-2 py-1 text-right">{tr("Target", "타깃")}</th>
                  <th className="border-b border-slate-200 px-2 py-1 text-left">{tr("Status", "상태")}</th>
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
              {tr("Download ZIP", "ZIP 다운로드")}
            </a>
          )}
          {reportSummary?.report_id && (
            <a
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-100"
              href={`/api/report/${reportSummary.report_id}`}
            >
              {tr("Download Report", "리포트 다운로드")}
            </a>
          )}
          <button
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-semibold text-slate-700 hover:bg-slate-100"
            onClick={onResetRunState}
            type="button"
          >
            {tr("Clear Run Board", "실행 현황 초기화")}
          </button>
        </div>
      )}
    </section>
  );
}
