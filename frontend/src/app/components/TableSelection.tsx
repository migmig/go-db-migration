import {
  CompareFilter,
  DiscoverySummary,
  Locale,
  TableHistoryDetail,
  TableHistoryState,
  TableHistoryStatusFilter,
  TableRunState,
  TableSortOption,
  TargetTableEntry,
} from "../types";
import {
  formatHistoryTime,
  historyStatusBadgeClass,
  historyStatusLabel,
  normalizeTableKey,
  tableStatusBadgeClass,
  tableStatusLabel,
} from "../utils";

interface TableSelectionProps {
  tr: (en: string, ko: string) => string;
  allTables: string[];
  selectedTables: string[];
  objectGroupModeEnabled: boolean;
  previewTables: string[];
  discoverySummary: DiscoverySummary | null;
  previewObjectGroup: string | null;
  previewSequences: string[];
  compareEntries: TargetTableEntry[];
  compareFilter: CompareFilter;
  setCompareFilter: (val: CompareFilter) => void;
  compareSearch: string;
  setCompareSearch: (val: string) => void;
  tableSearch: string;
  setTableSearch: (val: string) => void;
  tableStatusFilter: TableHistoryStatusFilter;
  setTableStatusFilter: (val: TableHistoryStatusFilter) => void;
  tableSort: TableSortOption;
  setTableSort: (val: TableSortOption) => void;
  excludeMigratedSuccess: boolean;
  setExcludeMigratedSuccess: (val: boolean) => void;
  migrationBusy: boolean;
  selectAllVisibleTables: () => void;
  deselectAllVisibleTables: () => void;
  selectByCategory: (category: "source_only" | "both" | "target_only") => void;
  allVisibleSelected: boolean;
  filteredTables: string[];
  toggleTable: (table: string, checked: boolean) => void;
  selectedTableSet: Set<string>;
  tableProgress: Record<string, TableRunState>;
  historyByTable: Record<string, TableHistoryState>;
  locale: Locale;
  openTableHistory: (table: string) => Promise<void>;
  activeTableHistory: string | null;
  setActiveTableHistory: (val: string | null) => void;
  tableHistoryBusy: boolean;
  tableHistoryError: string | null;
  activeHistoryDetail: TableHistoryDetail | null | undefined;
  replayHistory: (id: number) => Promise<void>;
}

export function TableSelection({
  tr,
  allTables,
  selectedTables,
  objectGroupModeEnabled,
  previewTables,
  discoverySummary,
  previewObjectGroup,
  previewSequences,
  compareEntries,
  compareFilter,
  setCompareFilter,
  compareSearch,
  setCompareSearch,
  tableSearch,
  setTableSearch,
  tableStatusFilter,
  setTableStatusFilter,
  tableSort,
  setTableSort,
  excludeMigratedSuccess,
  setExcludeMigratedSuccess,
  migrationBusy,
  selectAllVisibleTables,
  deselectAllVisibleTables,
  selectByCategory,
  allVisibleSelected,
  filteredTables,
  toggleTable,
  selectedTableSet,
  tableProgress,
  historyByTable,
  locale,
  openTableHistory,
  activeTableHistory,
  setActiveTableHistory,
  tableHistoryBusy,
  tableHistoryError,
  activeHistoryDetail,
  replayHistory,
}: TableSelectionProps) {
  return (
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
                      {tr("Tables Group", "테이블 그룹")}
                      <span className="ml-2 rounded-full bg-white px-2 py-0.5 text-xs text-slate-600">
                        {previewTables.length}
                      </span>
                    </summary>
                    <p className="mt-2 text-xs text-slate-500">
                      {discoverySummary
                        ? tr("Oracle discovery completed for tables group.", "테이블 그룹 Oracle 탐색이 완료되었습니다.")
                        : tr("Selected tables to be migrated.", "마이그레이션할 테이블을 선택하세요.")}
                    </p>
                    <div className="mt-2 max-h-32 overflow-auto rounded-lg border border-slate-200 bg-white p-2">
                      {previewTables.length > 0 ? (
                        <ul className="space-y-1 text-sm text-slate-700">
                          {previewTables.map((table) => (
                            <li key={`preview-table-${table}`}>{table}</li>
                          ))}
                        </ul>
                      ) : (
                        <p className="text-sm text-slate-500">{tr("No tables selected.", "선택된 테이블이 없습니다.")}</p>
                      )}
                    </div>
                  </details>
                  <details className="rounded-xl border border-slate-200 bg-slate-50 p-3">
                    <summary className="cursor-pointer text-sm font-semibold text-slate-800">
                      {tr("Sequences Group", "시퀀스 그룹")}
                      <span className="ml-2 rounded-full bg-white px-2 py-0.5 text-xs text-slate-600">
                        {previewObjectGroup === "tables" ? 0 : previewSequences.length}
                      </span>
                    </summary>
                    <p className="mt-2 text-xs text-slate-500">
                      {previewObjectGroup === "tables"
                        ? tr("Tables-only mode disables sequence discovery.", "테이블 전용 모드에서는 시퀀스 탐색이 비활성화됩니다.")
                        : discoverySummary
                          ? tr("Discovered from Oracle metadata at run start.", "실행 시작 시 Oracle 메타데이터에서 탐색됩니다.")
                          : tr("Sequence discovery runs automatically when migration starts.", "마이그레이션 시작 시 시퀀스 탐색이 자동으로 실행됩니다.")}
                    </p>
                    <div className="mt-2 max-h-32 overflow-auto rounded-lg border border-slate-200 bg-white p-2">
                      {previewObjectGroup === "tables" ? (
                        <p className="text-sm text-slate-500">{tr("Sequence group is disabled.", "시퀀스 그룹이 비활성화되어 있습니다.")}</p>
                      ) : previewSequences.length > 0 ? (
                        <ul className="space-y-1 text-sm text-slate-700">
                          {previewSequences.map((sequence) => (
                            <li key={`preview-sequence-${sequence}`}>{sequence}</li>
                          ))}
                        </ul>
                      ) : (
                        <p className="text-sm text-slate-500">{tr("No sequences discovered yet.", "아직 탐색된 시퀀스가 없습니다.")}</p>
                      )}
                    </div>
                  </details>
                </div>
              )}
              {compareEntries.length > 0 && (() => {
                const sourceOnlyCount = compareEntries.filter((e) => e.category === "source_only").length;
                const bothCount = compareEntries.filter((e) => e.category === "both").length;
                const targetOnlyCount = compareEntries.filter((e) => e.category === "target_only").length;
                const filteredCompare = compareEntries.filter((e) => {
                  if (compareFilter !== "all" && e.category !== compareFilter) return false;
                  if (compareSearch && !e.name.toLowerCase().includes(compareSearch.toLowerCase())) return false;
                  return true;
                });
                return (
                  <details className="mb-4 rounded-xl border border-slate-200 bg-slate-50">
                    <summary className="cursor-pointer px-4 py-3 text-sm font-semibold text-slate-800">
                      {tr("Source vs Target Comparison", "소스-타겟 비교")}
                      <span className="ml-2 rounded-full bg-white px-2 py-0.5 text-xs text-slate-600">
                        {compareEntries.length}
                      </span>
                    </summary>
                    <div className="border-t border-slate-200 p-4">
                      <div className="mb-3 flex flex-wrap gap-2">
                        <span className="rounded-full border border-blue-300 bg-blue-100 px-3 py-1 text-xs font-semibold text-blue-800">
                          {tr("Source only", "소스에만")} {sourceOnlyCount}
                        </span>
                        <span className="rounded-full border border-emerald-300 bg-emerald-100 px-3 py-1 text-xs font-semibold text-emerald-800">
                          {tr("Both", "양쪽")} {bothCount}
                        </span>
                        <span className="rounded-full border border-amber-300 bg-amber-100 px-3 py-1 text-xs font-semibold text-amber-800">
                          {tr("Target only", "타겟에만")} {targetOnlyCount}
                        </span>
                      </div>
                      <div className="mb-2 flex flex-wrap gap-2">
                        {(["all", "source_only", "both", "target_only"] as CompareFilter[]).map((f) => (
                          <button
                            key={f}
                            className={`rounded-lg border px-3 py-1 text-xs font-medium ${compareFilter === f ? "border-brand-500 bg-brand-100 text-brand-800" : "border-slate-300 bg-white text-slate-600 hover:bg-slate-50"}`}
                            onClick={() => setCompareFilter(f)}
                            type="button"
                          >
                            {f === "all" ? tr("All", "전체") : f === "source_only" ? tr("Source only", "소스만") : f === "both" ? tr("Both", "양쪽") : tr("Target only", "타겟만")}
                          </button>
                        ))}
                        <input
                          className="ml-auto rounded-lg border border-slate-300 px-3 py-1 text-xs outline-none focus:border-brand-500"
                          onChange={(e) => setCompareSearch(e.target.value)}
                          placeholder={tr("Search...", "검색...")}
                          value={compareSearch}
                        />
                      </div>
                      <div className="overflow-x-auto">
                        <table className="w-full border-collapse text-xs">
                          <thead>
                            <tr className="bg-slate-100 text-left text-slate-600">
                              <th className="px-2 py-1.5">{tr("Table", "테이블명")}</th>
                              <th className="px-2 py-1.5 text-center">{tr("Source", "소스")}</th>
                              <th className="px-2 py-1.5 text-center">{tr("Target", "타겟")}</th>
                              <th className="px-2 py-1.5 text-right">{tr("Src rows", "소스 행 수")}</th>
                              <th className="px-2 py-1.5 text-right">{tr("Tgt rows", "타겟 행 수")}</th>
                              <th className="px-2 py-1.5">{tr("Status", "상태")}</th>
                            </tr>
                          </thead>
                          <tbody>
                            {filteredCompare.map((e) => {
                              const isRowDiff =
                                e.category === "both" &&
                                e.sourceRowCount !== null &&
                                e.targetRowCount !== null &&
                                e.sourceRowCount !== e.targetRowCount;
                              return (
                                <tr key={e.name} className="border-t border-slate-100 hover:bg-white">
                                  <td className="px-2 py-1.5 font-mono">{e.name}</td>
                                  <td className="px-2 py-1.5 text-center">{e.inSource ? "✓" : "—"}</td>
                                  <td className="px-2 py-1.5 text-center">{e.inTarget ? "✓" : "—"}</td>
                                  <td className="px-2 py-1.5 text-right text-slate-600">
                                    {e.sourceRowCount !== null ? e.sourceRowCount.toLocaleString() : "—"}
                                  </td>
                                  <td className="px-2 py-1.5 text-right text-slate-600">
                                    {e.targetRowCount !== null ? e.targetRowCount.toLocaleString() : "—"}
                                  </td>
                                  <td className="px-2 py-1.5">
                                    <span className={`rounded-full border px-2 py-0.5 text-xs font-medium ${
                                      e.category === "source_only"
                                        ? "border-blue-300 bg-blue-100 text-blue-800"
                                        : e.category === "both"
                                          ? "border-emerald-300 bg-emerald-100 text-emerald-800"
                                          : "border-amber-300 bg-amber-100 text-amber-800"
                                    }`}>
                                      {e.category === "source_only" ? tr("Source only", "소스만") : e.category === "both" ? tr("Both", "양쪽") : tr("Target only", "타겟만")}
                                    </span>
                                    {isRowDiff && (
                                      <span className="ml-1 rounded-full border border-orange-300 bg-orange-100 px-2 py-0.5 text-xs font-medium text-orange-800">
                                        {tr("Row diff", "행 수 불일치")}
                                      </span>
                                    )}
                                    {e.category === "both" && e.sourceRowCount === null && (
                                      <span className="ml-1 text-xs text-slate-400">
                                        {tr("Run pre-check to see row counts", "Pre-check 실행 후 행 수가 표시됩니다")}
                                      </span>
                                    )}
                                  </td>
                                </tr>
                              );
                            })}
                          </tbody>
                        </table>
                        {filteredCompare.length === 0 && (
                          <p className="py-4 text-center text-xs text-slate-400">
                            {tr("No tables match the filter.", "필터 조건에 맞는 테이블이 없습니다.")}
                          </p>
                        )}
                      </div>
                    </div>
                  </details>
                );
              })()}
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
                  <option value="all">{tr("All history status", "전체 이력 상태")}</option>
                  <option value="not_started">{tr("Not started", "미시작")}</option>
                  <option value="success">{tr("Migrated (success)", "이관 완료 (성공)")}</option>
                  <option value="failed">{tr("Migrated (failed)", "이관 완료 (실패)")}</option>
                </select>
                <select
                  aria-label="Table sort"
                  className="rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
                  onChange={(event) => setTableSort(event.target.value as TableSortOption)}
                  value={tableSort}
                >
                  <option value="table_asc">{tr("Sort: Table name (A-Z)", "정렬: 테이블명 (A-Z)")}</option>
                  <option value="table_desc">{tr("Sort: Table name (Z-A)", "정렬: 테이블명 (Z-A)")}</option>
                  <option value="recent_desc">{tr("Sort: Latest history", "정렬: 최근 이력순")}</option>
                  <option value="runs_desc">{tr("Sort: Run count", "정렬: 실행 횟수순")}</option>
                  <option value="history_status">{tr("Sort: History status", "정렬: 이력 상태순")}</option>
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
                {compareEntries.length > 0 && (
                  <>
                    <button
                      className="rounded-lg border border-blue-300 bg-blue-50 px-3 py-2 text-sm font-medium text-blue-800 hover:bg-blue-100 disabled:opacity-60"
                      disabled={migrationBusy}
                      onClick={() => selectByCategory("source_only")}
                      type="button"
                    >
                      {tr("Select source-only", "소스에만 있는 테이블 선택")}
                    </button>
                    <button
                      className="rounded-lg border border-emerald-300 bg-emerald-50 px-3 py-2 text-sm font-medium text-emerald-800 hover:bg-emerald-100 disabled:opacity-60"
                      disabled={migrationBusy}
                      onClick={() => selectByCategory("both")}
                      type="button"
                    >
                      {tr("Select both", "양쪽에 있는 테이블 선택")}
                    </button>
                  </>
                )}
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
            </div>  );
}
