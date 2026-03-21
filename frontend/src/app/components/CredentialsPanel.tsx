import { Credential } from "../../shared/api/types";
import { RoleFilter } from "../types";

type CredentialsPanelProps = {
  credentialFilter: RoleFilter;
  credentialsBusy: boolean;
  credentialsError: string;
  filteredCredentials: Credential[];
  tr: (en: string, ko: string) => string;
  onClose: () => void;
  onFilterChange: (role: RoleFilter) => void;
  onApply: (item: Credential) => void;
};

export function CredentialsPanel({
  credentialFilter,
  credentialsBusy,
  credentialsError,
  filteredCredentials,
  tr,
  onClose,
  onFilterChange,
  onApply,
}: CredentialsPanelProps) {
  return (
    <aside className="fixed inset-y-0 right-0 z-30 w-full max-w-md border-l border-slate-200 bg-white p-5 shadow-2xl">
      <div className="mb-3 flex items-center justify-between">
        <h3 className="text-lg font-semibold text-slate-900">{tr("Saved Connections", "저장된 연결")}</h3>
        <button
          className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm text-slate-600 hover:bg-slate-100"
          onClick={onClose}
          type="button"
        >
          {tr("Close", "닫기")}
        </button>
      </div>
      <div className="mb-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-600">
        {tr("Filter:", "필터:")}{" "}
        {credentialFilter === "all"
          ? tr("All", "전체")
          : credentialFilter === "source"
            ? tr("Source only (Oracle)", "소스만 (Oracle)")
            : tr("Target only (non-Oracle)", "타깃만 (Oracle 제외)")}
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
            onClick={() => onFilterChange(role)}
            type="button"
          >
            {role === "all" ? tr("All", "전체") : role === "source" ? tr("Source", "소스") : tr("Target", "타깃")}
          </button>
        ))}
      </div>
      {credentialsBusy && <p className="text-sm text-slate-600">{tr("Loading...", "불러오는 중...")}</p>}
      {credentialsError && <p className="text-sm text-red-600">{credentialsError}</p>}
      {!credentialsBusy && !credentialsError && filteredCredentials.length === 0 && (
        <p className="text-sm text-slate-700 dark:text-slate-300">
          {credentialFilter === "source"
            ? tr("No saved source connections found.", "저장된 소스 연결이 없습니다.")
            : credentialFilter === "target"
              ? tr("No saved target connections found.", "저장된 타깃 연결이 없습니다.")
              : tr("No saved connections found.", "저장된 연결이 없습니다.")}
        </p>
      )}
      <div className="space-y-3">
        {filteredCredentials.map((item) => (
          <div className="rounded-xl border border-slate-200 p-3" key={item.id}>
            <p className="text-sm font-semibold text-slate-900">{item.alias}</p>
            <p className="text-xs text-slate-700 dark:text-slate-300">
              {item.dbType === "oracle" ? tr("Source", "소스") : tr("Target", "타깃")} · {item.dbType}
            </p>
            <p className="mt-1 break-all text-xs text-slate-700">{item.host}</p>
            <button
              className="mt-3 rounded-lg bg-brand-600 px-3 py-1.5 text-xs font-semibold text-white hover:bg-brand-700"
              onClick={() => onApply(item)}
              type="button"
            >
              {tr("Apply to form", "폼에 적용")}
            </button>
          </div>
        ))}
      </div>
    </aside>
  );
}
