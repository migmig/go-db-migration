import { Dispatch, SetStateAction, useState } from "react";
import { PrecheckSummary, PrecheckTableResult, RuntimeMeta } from "../../shared/api/types";
import {
  MigrationOptions,
  ObjectGroup,
  PrecheckDecisionFilter,
  TargetState,
} from "../types";
import { toNumber } from "../utils";

type MigrationOptionsPanelProps = {
  tr: (en: string, ko: string) => string;
  meta: RuntimeMeta | null;
  targetMode: TargetState["mode"];
  objectGroupModeEnabled: boolean;
  effectiveObjectGroup: ObjectGroup;
  options: MigrationOptions;
  setOptions: Dispatch<SetStateAction<MigrationOptions>>;
  onApplyObjectGroupSelection: (nextGroup: ObjectGroup) => void;
  precheckPolicy: string;
  setPrecheckPolicy: Dispatch<SetStateAction<string>>;
  precheckBusy: boolean;
  migrationBusy: boolean;
  selectedTablesCount: number;
  onRunPrecheck: () => void;
  precheckError: string;
  precheckSummary: PrecheckSummary | null;
  precheckDecisionFilter: PrecheckDecisionFilter;
  setPrecheckDecisionFilter: Dispatch<SetStateAction<PrecheckDecisionFilter>>;
  precheckItems: PrecheckTableResult[];
  onStartMigration: () => void;
  migrationError: string;
};

export function MigrationOptionsPanel({
  tr,
  meta,
  targetMode,
  objectGroupModeEnabled,
  effectiveObjectGroup,
  options,
  setOptions,
  onApplyObjectGroupSelection,
  precheckPolicy,
  setPrecheckPolicy,
  precheckBusy,
  migrationBusy,
  selectedTablesCount,
  onRunPrecheck,
  precheckError,
  precheckSummary,
  precheckDecisionFilter,
  setPrecheckDecisionFilter,
  precheckItems,
  onStartMigration,
  migrationError,
}: MigrationOptionsPanelProps) {
  const [isAdvancedOpen, setIsAdvancedOpen] = useState(false);

  return (
    <div className="card-surface p-5 dark:bg-slate-800 dark:border-slate-700">
      <h2 className="mb-4 text-lg font-semibold text-slate-900 dark:text-slate-100">
        {tr("Migration Options", "마이그레이션 옵션")}
      </h2>
      <div className="space-y-4">
        {/* Basic Options */}
        <div className="space-y-3">
          {objectGroupModeEnabled && (
            <div className="block text-sm">
              <label htmlFor="migration-target-select" className="mb-1 block text-slate-700 dark:text-slate-300">{tr("Migration target", "마이그레이션 대상")}</label>
              <select
                id="migration-target-select"
                className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-200"
                onChange={(event) =>
                  onApplyObjectGroupSelection(event.target.value as ObjectGroup)
                }
                value={options.objectGroup}
              >
                <option value="all">{tr("All objects", "전체 객체")}</option>
                <option value="tables">{tr("Tables only", "테이블만")}</option>
                <option value="sequences">{tr("Sequences only", "시퀀스만")}</option>
              </select>
            </div>
          )}

          <div className="grid gap-2 sm:grid-cols-2">
            <label className="inline-flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300">
              <input
                checked={options.withDdl}
                disabled={effectiveObjectGroup === "sequences"}
                onChange={(event) =>
                  setOptions((prev) => ({ ...prev, withDdl: event.target.checked }))
                }
                type="checkbox"
                className="rounded border-slate-300 dark:border-slate-600 dark:bg-slate-700"
              />
              {tr("Include CREATE TABLE DDL", "테이블 DDL 포함")}
            </label>
            <label className="inline-flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300">
              <input
                checked={options.withSequences}
                disabled={effectiveObjectGroup !== "all"}
                onChange={(event) =>
                  setOptions((prev) => ({ ...prev, withSequences: event.target.checked }))
                }
                type="checkbox"
                className="rounded border-slate-300 dark:border-slate-600 dark:bg-slate-700"
              />
              {tr("Include sequences", "시퀀스 포함")}
            </label>
            <label className="inline-flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300">
              <input
                checked={options.withIndexes}
                onChange={(event) =>
                  setOptions((prev) => ({ ...prev, withIndexes: event.target.checked }))
                }
                type="checkbox"
                className="rounded border-slate-300 dark:border-slate-600 dark:bg-slate-700"
              />
              {tr("Include indexes", "인덱스 포함")}
            </label>
            <label className="inline-flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300">
              <input
                checked={options.withConstraints}
                onChange={(event) =>
                  setOptions((prev) => ({
                    ...prev,
                    withConstraints: event.target.checked,
                  }))
                }
                type="checkbox"
                className="rounded border-slate-300 dark:border-slate-600 dark:bg-slate-700"
              />
              {tr("Include constraints", "제약 조건 포함")}
            </label>
          </div>

          {objectGroupModeEnabled && effectiveObjectGroup !== "all" && (
            <p className="text-xs text-slate-700 dark:text-slate-300">
              {effectiveObjectGroup === "tables"
                ? tr("Tables-only mode disables sequence DDL automatically.", "테이블 전용 모드에서는 시퀀스 DDL이 자동으로 비활성화됩니다.")
                : tr("Sequences-only mode forces DDL + sequence generation automatically.", "시퀀스 전용 모드에서는 DDL + 시퀀스 생성이 자동으로 활성화됩니다.")}
            </p>
          )}

          <div className="mt-2 space-y-2">
            <label className="inline-flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300">
              <input
                checked={options.validate}
                onChange={(event) =>
                  setOptions((prev) => ({ ...prev, validate: event.target.checked }))
                }
                type="checkbox"
                className="rounded border-slate-300 dark:border-slate-600 dark:bg-slate-700"
              />
              {tr("Validate row counts after migration", "마이그레이션 후 행 수 검증")}
            </label>
            <label className="inline-flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300">
              <input
                checked={options.truncate}
                onChange={(event) =>
                  setOptions((prev) => ({ ...prev, truncate: event.target.checked }))
                }
                type="checkbox"
                className="rounded border-slate-300 dark:border-slate-600 dark:bg-slate-700"
              />
              {tr("Truncate target tables before migration", "마이그레이션 전 타깃 테이블 초기화 (중복 방지)")}
            </label>
            <label className="inline-flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300">
              <input
                checked={options.upsert}
                onChange={(event) =>
                  setOptions((prev) => ({ ...prev, upsert: event.target.checked }))
                }
                type="checkbox"
                className="rounded border-slate-300 dark:border-slate-600 dark:bg-slate-700"
              />
              {tr("Upsert mode — skip duplicate rows by PK", "Upsert 모드 — PK 기준 중복 행 건너뛰기")}
            </label>
            <label className="inline-flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300">
              <input
                checked={options.dryRun}
                onChange={(event) =>
                  setOptions((prev) => ({ ...prev, dryRun: event.target.checked }))
                }
                type="checkbox"
                className="rounded border-slate-300 dark:border-slate-600 dark:bg-slate-700"
              />
              <span className="font-semibold">{tr("Dry-run mode", "드라이런 모드 (실제 전송 없음)")}</span>
            </label>
          </div>
        </div>

        <div className="border-t border-slate-200 pt-4 dark:border-slate-700">
          <button
            type="button"
            onClick={() => setIsAdvancedOpen(!isAdvancedOpen)}
            className="flex w-full items-center justify-between text-sm font-semibold text-slate-700 hover:text-brand-600 focus:outline-none dark:text-slate-300 dark:hover:text-brand-400"
          >
            <span>{tr("Advanced Settings", "고급 설정")}</span>
            <span className={`transition-transform duration-200 ${isAdvancedOpen ? "rotate-180" : ""}`}>
              ▼
            </span>
          </button>
          
          <div className={`mt-3 overflow-hidden transition-all duration-300 ${isAdvancedOpen ? "max-h-[1000px] opacity-100" : "max-h-0 opacity-0"}`}>
            <div className="grid gap-3 rounded-xl border border-slate-200 bg-slate-50 p-4 sm:grid-cols-2 dark:border-slate-700 dark:bg-slate-800/50">
              {targetMode === "file" && (
                <>
                  <div className="col-span-full block text-sm">
                    <label htmlFor="output-file-input" className="mb-1 block text-slate-700 dark:text-slate-300">{tr("Output file", "출력 파일")}</label>
                    <input
                      id="output-file-input"
                      className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-200"
                      onChange={(event) =>
                        setOptions((prev) => ({ ...prev, outFile: event.target.value }))
                      }
                      value={options.outFile}
                    />
                  </div>
                  <label className="col-span-full inline-flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300">
                    <input
                      checked={options.perTable}
                      onChange={(event) =>
                        setOptions((prev) => ({ ...prev, perTable: event.target.checked }))
                      }
                      type="checkbox"
                      className="rounded border-slate-300 dark:border-slate-600 dark:bg-slate-700"
                    />
                    {tr("Per-table output files", "테이블별 출력 파일")}
                  </label>
                </>
              )}
              <div className="col-span-full block text-sm">
                <label htmlFor="oracle-owner-input" className="mb-1 block text-slate-700 dark:text-slate-300">{tr("Oracle owner (optional)", "Oracle 소유자 (선택)")}</label>
                <input
                  id="oracle-owner-input"
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-200"
                  onChange={(event) =>
                    setOptions((prev) => ({ ...prev, oracleOwner: event.target.value }))
                  }
                  placeholder={tr("defaults to connected account", "연결 계정 기본값 사용")}
                  value={options.oracleOwner}
                />
              </div>

              <div className="block text-sm">
                <label htmlFor="batch-size-input" className="mb-1 block text-slate-700 dark:text-slate-300">{tr("Batch size", "배치 크기")}</label>
                <input
                  id="batch-size-input"
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-200"
                  onChange={(event) =>
                    setOptions((prev) => ({
                      ...prev,
                      batchSize: toNumber(event.target.value, prev.batchSize),
                    }))
                  }
                  type="number"
                  value={options.batchSize}
                />
              </div>
              <div className="block text-sm">
                <label htmlFor="workers-input" className="mb-1 block text-slate-700 dark:text-slate-300">{tr("Workers", "워커 수")}</label>
                <input
                  id="workers-input"
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-200"
                  onChange={(event) =>
                    setOptions((prev) => ({
                      ...prev,
                      workers: toNumber(event.target.value, prev.workers),
                    }))
                  }
                  type="number"
                  value={options.workers}
                />
              </div>
              <div className="block text-sm">
                <label htmlFor="copy-batch-input" className="mb-1 block text-slate-700 dark:text-slate-300">{tr("COPY batch", "COPY 배치")}</label>
                <input
                  id="copy-batch-input"
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-200"
                  onChange={(event) =>
                    setOptions((prev) => ({
                      ...prev,
                      copyBatch: toNumber(event.target.value, prev.copyBatch),
                    }))
                  }
                  type="number"
                  value={options.copyBatch}
                />
              </div>
              <div className="block text-sm">
                <label htmlFor="db-max-open-input" className="mb-1 block text-slate-700 dark:text-slate-300">{tr("DB max open", "DB 최대 연결")}</label>
                <input
                  id="db-max-open-input"
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-200"
                  onChange={(event) =>
                    setOptions((prev) => ({
                      ...prev,
                      dbMaxOpen: toNumber(event.target.value, prev.dbMaxOpen),
                    }))
                  }
                  type="number"
                  value={options.dbMaxOpen}
                />
              </div>
              <div className="block text-sm">
                <label htmlFor="db-max-idle-input" className="mb-1 block text-slate-700 dark:text-slate-300">{tr("DB max idle", "DB 최대 유휴")}</label>
                <input
                  id="db-max-idle-input"
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-200"
                  onChange={(event) =>
                    setOptions((prev) => ({
                      ...prev,
                      dbMaxIdle: toNumber(event.target.value, prev.dbMaxIdle),
                    }))
                  }
                  type="number"
                  value={options.dbMaxIdle}
                />
              </div>
              <div className="block text-sm">
                <label htmlFor="db-max-life-input" className="mb-1 block text-slate-700 dark:text-slate-300">{tr("DB max life (sec)", "DB 최대 수명 (초)")}</label>
                <input
                  id="db-max-life-input"
                  className="w-full rounded-xl border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-200"
                  onChange={(event) =>
                    setOptions((prev) => ({
                      ...prev,
                      dbMaxLife: toNumber(event.target.value, prev.dbMaxLife),
                    }))
                  }
                  type="number"
                  value={options.dbMaxLife}
                />
              </div>
              <label className="col-span-full inline-flex items-center gap-2 text-sm text-slate-700 dark:text-slate-300">
                <input
                  checked={options.logJson}
                  onChange={(event) =>
                    setOptions((prev) => ({ ...prev, logJson: event.target.checked }))
                  }
                  type="checkbox"
                  className="rounded border-slate-300 dark:border-slate-600 dark:bg-slate-700"
                />
                {tr("JSON logging", "JSON 로깅")}
              </label>
            </div>
          </div>
        </div>

        {meta?.features?.precheckRowCount !== false && (
          <div className="mt-4 rounded-xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-800/50">
            <div className="mb-3 flex flex-wrap items-center gap-3">
              <h3 className="text-sm font-semibold text-slate-800 dark:text-slate-200">{tr("Pre-check Row Count", "사전 행 수 점검")}</h3>
              <label htmlFor="precheck-policy-select" className="sr-only">Pre-check Policy</label>
              <select
                id="precheck-policy-select"
                className="rounded-lg border border-slate-300 px-2 py-1 text-xs text-slate-700 outline-none focus:border-brand-500 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-200"
                value={precheckPolicy}
                onChange={(e) => setPrecheckPolicy(e.target.value)}
              >
                <option value="strict">strict</option>
                <option value="best_effort">best_effort</option>
                <option value="skip_equal_rows">skip_equal_rows</option>
              </select>
              <button
                className="rounded-lg bg-blue-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
                disabled={precheckBusy || migrationBusy || selectedTablesCount === 0}
                onClick={onRunPrecheck}
                type="button"
              >
                {precheckBusy ? tr("Checking...", "점검 중...") : tr("Run Pre-check", "사전 점검 실행")}
              </button>
            </div>
            {precheckError && (
              <p className="mb-2 text-xs font-medium text-red-600 dark:text-red-400">{precheckError}</p>
            )}
            {precheckSummary && (
              <>
                <div className="mb-3 grid grid-cols-4 gap-2">
                  {(
                    [
                      { label: tr("Total", "전체"), value: precheckSummary.total_tables, cls: "bg-slate-100 text-slate-800 dark:bg-slate-700 dark:text-slate-200" },
                      { label: tr("Transfer Required", "이관 필요"), value: precheckSummary.transfer_required_count, cls: "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300" },
                      { label: tr("Skip Candidate", "건너뛰기 후보"), value: precheckSummary.skip_candidate_count, cls: "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300" },
                      { label: tr("Check Failed", "점검 실패"), value: precheckSummary.count_check_failed_count, cls: "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300" },
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
                      className={`rounded-lg px-2 py-1 text-xs font-medium ${
                        precheckDecisionFilter === f 
                          ? "bg-blue-600 text-white" 
                          : "bg-white border border-slate-300 text-slate-700 hover:bg-slate-100 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-300 dark:hover:bg-slate-700"
                      }`}
                      key={f}
                      onClick={() => setPrecheckDecisionFilter(f)}
                      type="button"
                    >
                      {f === "all" ? tr("All", "전체") : f === "transfer_required" ? tr("Transfer Required", "이관 필요") : f === "skip_candidate" ? tr("Skip", "건너뛰기") : tr("Failed", "실패")}
                    </button>
                  ))}
                </div>
                <div className="max-h-48 overflow-auto rounded-lg border border-slate-200 dark:border-slate-700">
                  <div className="w-full text-xs" role="none">
                    <div className="bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-600 dark:text-slate-200 flex font-bold" role="none">
                      <div className="px-2 py-1.5 text-left flex-1">{tr("Table", "테이블")}</div>
                      <div className="px-2 py-1.5 text-right w-20">{tr("Source", "소스")}</div>
                      <div className="px-2 py-1.5 text-right w-20">{tr("Target", "타깃")}</div>
                      <div className="px-2 py-1.5 text-right w-20">{tr("Diff", "차이")}</div>
                      <div className="px-2 py-1.5 text-left w-24">{tr("Decision", "결정")}</div>
                    </div>
                    <div className="dark:bg-slate-800/50" role="none">
                      {(precheckItems || [])
                        .filter((r) => precheckDecisionFilter === "all" || r.decision === precheckDecisionFilter)
                        .map((r) => (
                          <div className="border-t border-slate-100 hover:bg-slate-50 dark:border-slate-700 dark:hover:bg-slate-700/50 flex" key={r.table_name} role="none">
                            <div className="px-2 py-1.5 font-mono dark:text-slate-300 flex-1">{r.table_name}</div>
                            <div className="px-2 py-1.5 text-right dark:text-slate-300 w-20">{r.source_row_count.toLocaleString()}</div>
                            <div className="px-2 py-1.5 text-right dark:text-slate-300 w-20">{r.target_row_count.toLocaleString()}</div>
                            <div className={`px-2 py-1.5 text-right w-20 ${r.diff !== 0 ? "font-semibold text-amber-700 dark:text-amber-400" : "text-slate-700 dark:text-slate-300"}`}>{r.diff > 0 ? "+" : ""}{r.diff.toLocaleString()}</div>
                            <div className="px-2 py-1.5 w-24">
                              <span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${
                                r.decision === "transfer_required" ? "bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300" :
                                r.decision === "skip_candidate" ? "bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300" :
                                "bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300"
                              }`}>
                                {r.decision === "count_check_failed" && <span title={r.reason}>⚠</span>}
                                {r.decision}
                              </span>
                            </div>
                          </div>
                        ))}
                    </div>
                  </div>
                </div>
              </>
            )}
          </div>
        )}

        <button
          className="mt-2 w-full rounded-xl bg-emerald-600 px-4 py-3 text-sm font-semibold text-white shadow-sm transition-all hover:bg-emerald-700 disabled:cursor-not-allowed disabled:opacity-60"
          disabled={migrationBusy || selectedTablesCount === 0}
          onClick={onStartMigration}
          type="button"
        >
          {migrationBusy
            ? options.dryRun
              ? tr("Verification running...", "검증 실행 중...")
              : tr("Migration running...", "마이그레이션 실행 중...")
            : options.dryRun
              ? tr("Run Verification", "검증 실행")
              : tr("Start Migration", "마이그레이션 시작")}
        </button>
      </div>
      {migrationError && (
        <p className="mt-3 text-sm font-medium text-red-600 dark:text-red-400">{migrationError}</p>
      )}
    </div>
  );
}
