import { HistoryEntry } from "../../shared/api/types";

type HistoryPanelProps = {
  historyBusy: boolean;
  historyError: string;
  history: HistoryEntry[];
  tr: (en: string, ko: string) => string;
  onClose: () => void;
  onReplay: (id: number) => void;
};

export function HistoryPanel({
  historyBusy,
  historyError,
  history,
  tr,
  onClose,
  onReplay,
}: HistoryPanelProps) {
  return (
    <aside className="fixed inset-y-0 right-0 z-30 w-full max-w-md border-l border-slate-200 bg-white p-5 shadow-2xl">
      <div className="mb-3 flex items-center justify-between">
        <h3 className="text-lg font-semibold text-slate-900">{tr("My History", "내 작업 이력")}</h3>
        <button
          className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-100"
          onClick={onClose}
          type="button"
        >
          {tr("Close", "닫기")}
        </button>
      </div>
      {historyBusy && <p className="text-sm text-slate-600">{tr("Loading history...", "이력 불러오는 중...")}</p>}
      {historyError && <p className="text-sm text-red-600">{historyError}</p>}
      {!historyBusy && !historyError && history.length === 0 && (
        <p className="text-sm text-slate-500">{tr("No migration history yet.", "아직 마이그레이션 이력이 없습니다.")}</p>
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
              onClick={() => onReplay(entry.id)}
              type="button"
            >
              {tr("Replay into form", "폼에 재적용")}
            </button>
          </div>
        ))}
      </div>
    </aside>
  );
}
