import { RuntimeMeta } from "../../shared/api/types";
import { CompareState, SourceState, TargetState } from "../types";

type ConnectionFormsProps = {
  meta: RuntimeMeta | null;
  source: SourceState;
  target: TargetState;
  sourceConnectBusy: boolean;
  sourceConnectError: string;
  allTablesCount: number;
  targetTestBusy: boolean;
  targetTestError: string;
  targetTestMessage: string;
  compareState: CompareState;
  migrationBusy: boolean;
  tr: (en: string, ko: string) => string;
  onOpenSourceCredentials: () => void;
  onOpenTargetCredentials: () => void;
  onSourceFieldChange: (field: keyof SourceState, value: string) => void;
  onTargetFieldChange: (field: keyof TargetState, value: string) => void;
  onConnectSource: () => void;
  onTestTarget: () => void;
  onFetchTargetTables: () => void;
};

export function ConnectionForms({
  meta,
  source,
  target,
  sourceConnectBusy,
  sourceConnectError,
  allTablesCount,
  targetTestBusy,
  targetTestError,
  targetTestMessage,
  compareState,
  migrationBusy,
  tr,
  onOpenSourceCredentials,
  onOpenTargetCredentials,
  onSourceFieldChange,
  onTargetFieldChange,
  onConnectSource,
  onTestTarget,
  onFetchTargetTables,
}: ConnectionFormsProps) {
  return (
    <section className="grid gap-5 lg:grid-cols-2">
      <div className="card-surface p-5">
        <div className="mb-4 flex items-center justify-between gap-3">
          <h2 className="text-lg font-semibold text-slate-900">
            {tr("1. Source (Oracle)", "1. 소스 (Oracle)")}
          </h2>
          {meta?.authEnabled && (
            <button
              className="rounded-lg border border-brand-300 bg-brand-50 px-3 py-2 text-xs font-semibold text-brand-700 hover:bg-brand-100"
              onClick={onOpenSourceCredentials}
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
              onChange={(event) => onSourceFieldChange("oracleUrl", event.target.value)}
              placeholder="localhost:1521/XE"
              value={source.oracleUrl}
            />
          </label>
          <label className="block text-sm">
            <span className="mb-1 block text-slate-700">{tr("Username", "사용자명")}</span>
            <input
              className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
              onChange={(event) => onSourceFieldChange("username", event.target.value)}
              value={source.username}
            />
          </label>
          <label className="block text-sm">
            <span className="mb-1 block text-slate-700">{tr("Password", "비밀번호")}</span>
            <input
              className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
              onChange={(event) => onSourceFieldChange("password", event.target.value)}
              type="password"
              value={source.password}
            />
          </label>
          <label className="block text-sm">
            <span className="mb-1 block text-slate-700">Table filter (LIKE)</span>
            <input
              className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
              onChange={(event) => onSourceFieldChange("like", event.target.value)}
              placeholder="USERS_%"
              value={source.like}
            />
          </label>
          <button
            className="w-full rounded-xl bg-brand-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-brand-700 disabled:cursor-not-allowed disabled:opacity-60"
            disabled={sourceConnectBusy}
            onClick={onConnectSource}
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
        {allTablesCount > 0 && (
          <div className="mt-4 rounded-xl border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-800">
            <p className="font-semibold">
              {tr("Found", "총")} {allTablesCount}
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
              onClick={onOpenTargetCredentials}
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
                onTargetFieldChange("mode", event.target.value as TargetState["mode"])
              }
              value={target.mode}
            >
              <option value="file">{tr("SQL file mode", "SQL 파일 모드")}</option>
              <option value="direct">{tr("Direct migration", "직접 마이그레이션")}</option>
            </select>
          </label>
          <div className="block text-sm">
            <span className="mb-1 block text-slate-700">{tr("Target DB", "타깃 DB")}</span>
            <span className="inline-block rounded-xl border border-slate-200 bg-slate-50 px-3 py-2 text-sm text-slate-700">
              PostgreSQL
            </span>
          </div>
          <label className="block text-sm">
            <span className="mb-1 block text-slate-700">{tr("Target URL", "타깃 URL")}</span>
            <input
              className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
              onChange={(event) => onTargetFieldChange("targetUrl", event.target.value)}
              placeholder="postgres://user:pass@host:5432/dbname"
              value={target.targetUrl}
            />
          </label>
          <label className="block text-sm">
            <span className="mb-1 block text-slate-700">{tr("Schema", "스키마")}</span>
            <input
              className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
              onChange={(event) => onTargetFieldChange("schema", event.target.value)}
              value={target.schema}
            />
          </label>
          <button
            className="w-full rounded-xl bg-slate-900 px-4 py-2.5 text-sm font-semibold text-white hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-60"
            disabled={targetTestBusy}
            onClick={onTestTarget}
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
        {target.mode === "direct" && (
          <div className="mt-3 flex flex-wrap items-center gap-2">
            <button
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-60"
              disabled={compareState.busy || migrationBusy || !target.targetUrl || !target.schema}
              onClick={onFetchTargetTables}
              type="button"
            >
              {compareState.busy
                ? tr("Fetching...", "조회 중...")
                : compareState.fetchedAt
                  ? tr("Refresh", "새로고침")
                  : tr("Fetch Target Tables", "타겟 테이블 조회")}
            </button>
            {compareState.targetTables.length > 0 && (
              <span className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-semibold text-slate-700">
                {compareState.targetTables.length} {tr("tables in target", "개 타겟 테이블")}
              </span>
            )}
            {compareState.fetchedAt && (
              <span className="text-xs text-slate-600 dark:text-slate-200">
                {tr("as of", "기준 시각")}{" "}
                {new Date(compareState.fetchedAt).toLocaleTimeString()}
              </span>
            )}
            {compareState.error && (
              <p className="w-full text-sm font-medium text-red-600">{compareState.error}</p>
            )}
          </div>
        )}
      </div>
    </section>
  );
}
