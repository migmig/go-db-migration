import { AuthUser, RuntimeMeta } from "../../shared/api/types";
import { Locale } from "../types";

type HeaderBarProps = {
  locale: Locale;
  authMeta: RuntimeMeta | null;
  user: AuthUser | null;
  t: (key: string) => string;
  onToggleLocale: () => void;
  onOpenCredentials: () => void;
  onOpenHistory: () => void;
  onLogout: () => void;
};

export function HeaderBar({
  locale,
  authMeta,
  user,
  t,
  onToggleLocale,
  onOpenCredentials,
  onOpenHistory,
  onLogout,
}: HeaderBarProps) {
  return (
    <header className="card-surface flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <p className="text-xs font-semibold uppercase tracking-[0.18em] text-brand-700">
          DBMigrator
        </p>
        <h1 className="text-2xl font-bold text-slate-900">{t("workspaceTitle")}</h1>
        <p className="mt-1 text-sm text-slate-600">{t("workspaceDesc")}</p>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <button
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
          onClick={onToggleLocale}
          type="button"
        >
          {locale === "en" ? t("switchToKorean") : t("switchToEnglish")}
        </button>
        {authMeta?.authEnabled ? (
          <>
            <span className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-semibold text-slate-700">
              {user ? `User: ${user.username}` : t("authEnabled")}
            </span>
            <button
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
              onClick={onOpenCredentials}
              type="button"
            >
              {t("savedConnections")}
            </button>
            <button
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100"
              onClick={onOpenHistory}
              type="button"
            >
              {t("myHistory")}
            </button>
            <button
              className="rounded-lg bg-slate-900 px-3 py-2 text-sm font-semibold text-white hover:bg-slate-700"
              onClick={onLogout}
              type="button"
            >
              {t("logout")}
            </button>
          </>
        ) : (
          <span className="rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-700">
            {t("authDisabled")}
          </span>
        )}
      </div>
    </header>
  );
}
