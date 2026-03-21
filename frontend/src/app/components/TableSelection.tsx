import { useState, useMemo } from "react";
import { HistoryEntry } from "../../shared/api/types";
import {
  DiscoverySummary,
  TableHistoryDetail,
  TableHistoryState,
  TableRunState,
  TargetTableEntry,
} from "../types";
import {
  formatHistoryTime,
  normalizeTableKey,
  parseReplayedTables,
} from "../utils";

interface TableSelectionProps {
  tr: (en: string, ko: string) => string;
  allTables: string[];
  selectedTables: string[];
  setSelectedTables: React.Dispatch<React.SetStateAction<string[]>>;
  objectGroupModeEnabled: boolean;
  previewTables: string[];
  discoverySummary: DiscoverySummary | null;
  previewObjectGroup: string | null;
  previewSequences: string[];
  compareEntries: TargetTableEntry[];
  migrationBusy: boolean;
  selectByCategory: (category: "source_only" | "both" | "target_only") => void;
  tableProgress: Record<string, TableRunState>;
  historyByTable: Record<string, TableHistoryState>;
  history: HistoryEntry[];
  openTableHistory: (table: string) => Promise<void>;
  activeTableHistory: string | null;
  setActiveTableHistory: (val: string | null) => void;
  tableHistoryBusy: boolean;
  tableHistoryError: string | null;
  replayHistory: (id: number) => Promise<void>;
}

export function TableSelection({
  tr,
  allTables,
  selectedTables,
  setSelectedTables,
  objectGroupModeEnabled,
  previewTables,
  discoverySummary,
  previewObjectGroup,
  previewSequences,
  compareEntries,
  migrationBusy,
  selectByCategory,
  tableProgress,
  history,
  openTableHistory,
  activeTableHistory,
  setActiveTableHistory,
  tableHistoryBusy,
  tableHistoryError,
  replayHistory,
}: TableSelectionProps) {
  const [leftSearch, setLeftSearch] = useState("");
  const [rightSearch, setRightSearch] = useState("");
  const [leftChecked, setLeftChecked] = useState<Set<string>>(new Set());
  const [rightChecked, setRightChecked] = useState<Set<string>>(new Set());

  // Derived states
  const selectedTableSet = useMemo(() => new Set(selectedTables), [selectedTables]);
  const availableTables = useMemo(() => allTables.filter(t => !selectedTableSet.has(t)), [allTables, selectedTableSet]);

  const filteredAvailable = useMemo(() => {
    if (!leftSearch) return availableTables;
    const lower = leftSearch.toLowerCase();
    return availableTables.filter(t => t.toLowerCase().includes(lower));
  }, [availableTables, leftSearch]);

  const filteredSelected = useMemo(() => {
    if (!rightSearch) return selectedTables;
    const lower = rightSearch.toLowerCase();
    return selectedTables.filter(t => t.toLowerCase().includes(lower));
  }, [selectedTables, rightSearch]);

  const activeHistoryDetail = useMemo<TableHistoryDetail | null>(() => {
    if (!activeTableHistory) return null;
    const normalized = normalizeTableKey(activeTableHistory);
    const entries: HistoryEntry[] = [];
    for (const entry of history) {
      const tables = parseReplayedTables(entry.optionsJson);
      if (tables.some(t => normalizeTableKey(t) === normalized)) {
        entries.push(entry);
      }
    }
    entries.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
    return { tableName: normalized, entries };
  }, [activeTableHistory, history]);

  const handleToggleLeft = (table: string) => {
    setLeftChecked(prev => {
      const next = new Set(prev);
      if (next.has(table)) next.delete(table);
      else next.add(table);
      return next;
    });
  };

  const handleToggleRight = (table: string) => {
    setRightChecked(prev => {
      const next = new Set(prev);
      if (next.has(table)) next.delete(table);
      else next.add(table);
      return next;
    });
  };

  const moveRight = () => {
    setSelectedTables(prev => [...prev, ...Array.from(leftChecked)]);
    setLeftChecked(new Set());
  };

  const moveLeft = () => {
    setSelectedTables(prev => prev.filter(t => !rightChecked.has(t)));
    setRightChecked(new Set());
  };

  const moveAllRight = () => {
    setSelectedTables(prev => [...prev, ...filteredAvailable]);
    setLeftChecked(new Set());
  };

  const moveAllLeft = () => {
    const toRemove = new Set(filteredSelected);
    setSelectedTables(prev => prev.filter(t => !toRemove.has(t)));
    setRightChecked(new Set());
  };

  return (
    <div className="card-surface p-5 dark:bg-slate-800 dark:border-slate-700">
      <div className="mb-3 flex items-center justify-between gap-3">
        <h2 className="text-lg font-semibold text-slate-900 dark:text-slate-100">
          {tr("Table Selection", "테이블 선택")}
        </h2>
        <span className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-semibold text-slate-700 dark:border-slate-600 dark:bg-slate-700 dark:text-slate-300">
          {selectedTables.length} / {allTables.length} {tr("selected", "선택됨")}
        </span>
      </div>

      {objectGroupModeEnabled && (
        <div className="mb-4 grid gap-3 lg:grid-cols-2">
          <details className="rounded-xl border border-slate-200 bg-slate-50 p-3 dark:border-slate-600 dark:bg-slate-700/50" open>
            <summary className="cursor-pointer text-sm font-semibold text-slate-800 dark:text-slate-200">
              {tr("Tables Group", "테이블 그룹")}
              <span className="ml-2 rounded-full bg-white px-2 py-0.5 text-xs text-slate-600 dark:bg-slate-600 dark:text-slate-300">
                {previewTables.length}
              </span>
            </summary>
            <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
              {discoverySummary
                ? tr("Oracle discovery completed for tables group.", "테이블 그룹 Oracle 탐색이 완료되었습니다.")
                : tr("Selected tables to be migrated.", "마이그레이션할 테이블을 선택하세요.")}
            </p>
          </details>
          <details className="rounded-xl border border-slate-200 bg-slate-50 p-3 dark:border-slate-600 dark:bg-slate-700/50">
            <summary className="cursor-pointer text-sm font-semibold text-slate-800 dark:text-slate-200">
              {tr("Sequences Group", "시퀀스 그룹")}
              <span className="ml-2 rounded-full bg-white px-2 py-0.5 text-xs text-slate-600 dark:bg-slate-600 dark:text-slate-300">
                {previewObjectGroup === "tables" ? 0 : previewSequences.length}
              </span>
            </summary>
            <p className="mt-2 text-xs text-slate-500 dark:text-slate-400">
              {previewObjectGroup === "tables"
                ? tr("Tables-only mode disables sequence discovery.", "테이블 전용 모드에서는 시퀀스 탐색이 비활성화됩니다.")
                : discoverySummary
                  ? tr("Discovered from Oracle metadata at run start.", "실행 시작 시 Oracle 메타데이터에서 탐색됩니다.")
                  : tr("Sequence discovery runs automatically when migration starts.", "마이그레이션 시작 시 시퀀스 탐색이 자동으로 실행됩니다.")}
            </p>
          </details>
        </div>
      )}

      {compareEntries.length > 0 && (
        <details className="mb-4 rounded-xl border border-slate-200 bg-slate-50 dark:border-slate-700 dark:bg-slate-800/50">
          <summary className="cursor-pointer px-4 py-3 text-sm font-semibold text-slate-800 dark:text-slate-200">
            {tr("Source vs Target Comparison", "소스-타겟 비교")}
            <span className="ml-2 rounded-full bg-white px-2 py-0.5 text-xs text-slate-600 dark:bg-slate-700 dark:text-slate-300">
              {compareEntries.length}
            </span>
          </summary>
          <div className="border-t border-slate-200 p-4 dark:border-slate-700">
             <div className="mb-3 flex flex-wrap gap-2">
                <button className="rounded-lg border border-blue-300 bg-blue-50 px-3 py-2 text-sm font-medium text-blue-800 hover:bg-blue-100 disabled:opacity-60 dark:border-blue-800 dark:bg-blue-900/30 dark:text-blue-300 dark:hover:bg-blue-900/50" disabled={migrationBusy} onClick={() => selectByCategory("source_only")} type="button">
                  {tr("Select source-only", "소스에만 있는 테이블 선택")}
                </button>
                <button className="rounded-lg border border-emerald-300 bg-emerald-50 px-3 py-2 text-sm font-medium text-emerald-800 hover:bg-emerald-100 disabled:opacity-60 dark:border-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-300 dark:hover:bg-emerald-900/50" disabled={migrationBusy} onClick={() => selectByCategory("both")} type="button">
                  {tr("Select both", "양쪽에 있는 테이블 선택")}
                </button>
             </div>
          </div>
        </details>
      )}

      <div className="flex flex-col gap-4 lg:flex-row lg:items-stretch">
        {/* Left Panel: Available */}
        <div className="flex flex-1 flex-col rounded-xl border border-slate-200 bg-white p-3 dark:border-slate-700 dark:bg-slate-900">
          <div className="mb-2">
            <input
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200"
              placeholder={tr("Search available...", "추가할 테이블 검색...")}
              value={leftSearch}
              onChange={(e) => setLeftSearch(e.target.value)}
            />
          </div>
          <div className="flex items-center justify-between px-2 pb-2 text-xs font-semibold text-slate-500 dark:text-slate-400">
            <span>{tr("Available Tables", "선택 가능")} ({filteredAvailable.length})</span>
            {leftChecked.size > 0 && <span>{leftChecked.size} {tr("selected", "선택됨")}</span>}
          </div>
          <div className="h-64 flex-1 overflow-auto rounded-lg border border-slate-100 bg-slate-50 p-2 dark:border-slate-700/50 dark:bg-slate-800/50">
            {filteredAvailable.length === 0 ? (
              <div className="py-8 text-center text-sm text-slate-400">
                {tr("No tables match", "테이블이 없습니다")}
              </div>
            ) : (
              <ul className="space-y-1">
                {filteredAvailable.map(t => (
                  <li
                    key={t}
                    onClick={() => handleToggleLeft(t)}
                    className={`cursor-pointer rounded px-2 py-1 text-sm transition-colors ${
                      leftChecked.has(t) 
                        ? "bg-brand-100 text-brand-900 dark:bg-brand-900/40 dark:text-brand-100" 
                        : "text-slate-700 hover:bg-slate-200 dark:text-slate-300 dark:hover:bg-slate-700"
                    }`}
                  >
                    {t}
                  </li>
                ))}
              </ul>
            )}
          </div>
        </div>

        {/* Center Controls */}
        <div className="flex items-center justify-center gap-2 lg:flex-col lg:gap-3">
          <button
            onClick={moveAllRight}
            disabled={filteredAvailable.length === 0 || migrationBusy}
            className="rounded-lg bg-slate-100 p-2 text-slate-600 transition-colors hover:bg-brand-100 hover:text-brand-700 disabled:opacity-50 dark:bg-slate-800 dark:text-slate-400 dark:hover:bg-brand-900/50 dark:hover:text-brand-300"
            title={tr("Add all", "전체 추가")}
          >
            <span className="hidden lg:inline">{">>"}</span>
            <span className="inline lg:hidden">{"\u2193\u2193"}</span>
          </button>
          <button
            onClick={moveRight}
            disabled={leftChecked.size === 0 || migrationBusy}
            className="rounded-lg bg-slate-100 p-2 text-slate-600 transition-colors hover:bg-brand-100 hover:text-brand-700 disabled:opacity-50 dark:bg-slate-800 dark:text-slate-400 dark:hover:bg-brand-900/50 dark:hover:text-brand-300"
            title={tr("Add selected", "선택 추가")}
          >
            <span className="hidden lg:inline">{">"}</span>
            <span className="inline lg:hidden">{"\u2193"}</span>
          </button>
          <button
            onClick={moveLeft}
            disabled={rightChecked.size === 0 || migrationBusy}
            className="rounded-lg bg-slate-100 p-2 text-slate-600 transition-colors hover:bg-red-100 hover:text-red-700 disabled:opacity-50 dark:bg-slate-800 dark:text-slate-400 dark:hover:bg-brand-900/50 dark:hover:text-brand-300"
            title={tr("Remove selected", "선택 제거")}
          >
            <span className="hidden lg:inline">{"<"}</span>
            <span className="inline lg:hidden">{"\u2191"}</span>
          </button>
          <button
            onClick={moveAllLeft}
            disabled={filteredSelected.length === 0 || migrationBusy}
            className="rounded-lg bg-slate-100 p-2 text-slate-600 transition-colors hover:bg-red-100 hover:text-red-700 disabled:opacity-50 dark:bg-slate-800 dark:text-slate-400 dark:hover:bg-brand-900/50 dark:hover:text-brand-300"
            title={tr("Remove all", "전체 제거")}
          >
            <span className="hidden lg:inline">{"<<"}</span>
            <span className="inline lg:hidden">{"\u2191\u2191"}</span>
          </button>
        </div>

        {/* Right Panel: Selected */}
        <div className="flex flex-1 flex-col rounded-xl border border-brand-200 bg-brand-50/30 p-3 dark:border-brand-900/50 dark:bg-brand-900/10">
          <div className="mb-2">
            <input
              className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200"
              placeholder={tr("Search selected...", "선택된 테이블 검색...")}
              value={rightSearch}
              onChange={(e) => setRightSearch(e.target.value)}
            />
          </div>
          <div className="flex items-center justify-between px-2 pb-2 text-xs font-semibold text-brand-700 dark:text-brand-400">
            <span>{tr("Selected Tables", "선택된 테이블")} ({filteredSelected.length})</span>
            {rightChecked.size > 0 && <span>{rightChecked.size} {tr("selected", "선택됨")}</span>}
          </div>
          <div className="h-64 flex-1 overflow-auto rounded-lg border border-brand-100 bg-white p-2 dark:border-brand-900/30 dark:bg-slate-800/80">
            {filteredSelected.length === 0 ? (
              <div className="py-8 text-center text-sm text-slate-400">
                {tr("No tables selected", "선택된 테이블이 없습니다")}
              </div>
            ) : (
              <ul className="space-y-1">
                {filteredSelected.map(t => {
                  const item = tableProgress[t];
                  const status = item?.status ?? "pending";
                  
                  return (
                  <li
                    key={t}
                    onClick={() => handleToggleRight(t)}
                    className={`flex cursor-pointer items-center justify-between rounded px-2 py-1 text-sm transition-colors ${
                      rightChecked.has(t) 
                        ? "bg-red-100 text-red-900 dark:bg-red-900/40 dark:text-red-100" 
                        : "text-slate-800 hover:bg-slate-100 dark:text-slate-200 dark:hover:bg-slate-700"
                    }`}
                  >
                    <div className="flex items-center gap-2 overflow-hidden">
                      <span className="truncate">{t}</span>
                      {status !== "pending" && (
                        <span className="h-2 w-2 flex-shrink-0 rounded-full bg-brand-500"></span>
                      )}
                    </div>
                    <button
                      className="text-xs text-brand-600 hover:underline dark:text-brand-400 flex-shrink-0"
                      onClick={(e) => { e.stopPropagation(); openTableHistory(t); }}
                    >
                      {tr("History", "이력")}
                    </button>
                  </li>
                )})}
              </ul>
            )}
          </div>
        </div>
      </div>
      
      {/* Table History View */}
      {activeTableHistory && (
        <div className="mt-4 rounded-xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-800/50">
          <div className="mb-3 flex items-center justify-between gap-3">
            <h3 className="text-sm font-semibold text-slate-900 dark:text-slate-100">
              {tr("Table history:", "테이블 이력:")} {activeTableHistory}
            </h3>
            <button
              className="rounded border border-slate-300 px-2 py-1 text-xs font-medium hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
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
              <div className="h-11 animate-pulse rounded border border-slate-200 bg-slate-100 dark:border-slate-700 dark:bg-slate-800" />
            </div>
          ) : tableHistoryError ? (
            <div className="rounded border border-red-200 bg-red-50 p-3 dark:border-red-900/50 dark:bg-red-900/20">
              <p className="text-xs text-red-700 dark:text-red-400">{tableHistoryError}</p>
            </div>
          ) : activeHistoryDetail && activeHistoryDetail.entries.length > 0 ? (
            <ul className="space-y-2 text-xs text-slate-700 dark:text-slate-300">
              {activeHistoryDetail.entries.slice(0, 5).map((entry) => {
                const failed = entry.status !== "success";
                return (
                  <li
                    className="rounded border border-slate-200 bg-white px-3 py-2 dark:border-slate-700 dark:bg-slate-800"
                    key={`table-history-${entry.id}`}
                  >
                    <div className="flex flex-wrap items-center gap-2">
                      <span
                        className={`inline-flex items-center gap-1 rounded-full border px-2 py-0.5 font-semibold ${failed ? "border-red-300 bg-red-100 text-red-900 dark:border-red-800 dark:bg-red-900/50 dark:text-red-300" : "border-emerald-300 bg-emerald-100 text-emerald-900 dark:border-emerald-800 dark:bg-emerald-900/50 dark:text-emerald-300"}`}
                      >
                        <span aria-hidden="true">●</span>
                        {failed ? tr("Failed", "실패") : tr("Success", "성공")}
                      </span>
                      <span>{formatHistoryTime(entry.createdAt)}</span>
                      {failed && (
                        <button
                          className="rounded border border-red-300 bg-red-50 px-2 py-0.5 font-semibold text-red-700 hover:bg-red-100 dark:border-red-800 dark:bg-red-900/30 dark:text-red-400 dark:hover:bg-red-900/50"
                          onClick={() => void replayHistory(entry.id)}
                          type="button"
                        >
                          {tr("Retry settings", "설정 다시 적용")}
                        </button>
                      )}
                    </div>
                    {failed && entry.logSummary && (
                      <p className="mt-1 text-red-700 dark:text-red-400">{entry.logSummary}</p>
                    )}
                  </li>
                );
              })}
            </ul>
          ) : (
            <p className="text-xs text-slate-500 dark:text-slate-400">
              {tr("No history found for this table.", "이 테이블의 이력이 없습니다.")}
            </p>
          )}
        </div>
      )}
    </div>
  );
}
