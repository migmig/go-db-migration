type RecentSourcePanelProps = {
  rememberSourcePassword: boolean;
  t: (key: string) => string;
  onRememberSourcePasswordChange: (checked: boolean) => void;
  onRestore: () => void;
  onClear: () => void;
};

export function RecentSourcePanel({
  rememberSourcePassword,
  t,
  onRememberSourcePasswordChange,
  onRestore,
  onClear,
}: RecentSourcePanelProps) {
  return (
    <details className="card-surface p-4">
      <summary className="cursor-pointer text-sm font-semibold text-slate-800">
        {t("recentSourceOptional")}
      </summary>
      <div className="mt-4 flex flex-wrap items-center gap-2">
        <label className="inline-flex items-center gap-2 text-sm text-slate-700">
          <input
            checked={rememberSourcePassword}
            onChange={(event) => onRememberSourcePasswordChange(event.target.checked)}
            type="checkbox"
          />
          {t("rememberSourcePassword")}
        </label>
        <button
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
          onClick={onRestore}
          type="button"
        >
          {t("restore")}
        </button>
        <button
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
          onClick={onClear}
          type="button"
        >
          {t("clear")}
        </button>
      </div>
    </details>
  );
}
