import { AuthUser, RuntimeMeta } from "../../shared/api/types";
import { Locale } from "../types";
import { Theme } from "../hooks/useTheme";

type HeaderBarProps = {
  locale: Locale;
  theme: Theme;
  authMeta: RuntimeMeta | null;
  user: AuthUser | null;
  t: (key: string) => string;
  onToggleLocale: () => void;
  onToggleTheme: () => void;
  onOpenCredentials: () => void;
  onOpenHistory: () => void;
  onLogout: () => void;
};

export function HeaderBar({
  locale,
  theme,
  authMeta,
  user,
  t,
  onToggleLocale,
  onToggleTheme,
  onOpenCredentials,
  onOpenHistory,
  onLogout,
}: HeaderBarProps) {
  return (
    <header className="card-surface flex flex-col gap-4 p-5 sm:flex-row sm:items-center sm:justify-between dark:bg-slate-800 dark:border-slate-700">
      <div>
        <p className="text-xs font-semibold uppercase tracking-[0.18em] text-brand-700 dark:text-brand-400">
          DBMigrator
        </p>
        <h1 className="text-2xl font-bold text-slate-900 dark:text-slate-100">{t("workspaceTitle")}</h1>
        <p className="mt-1 text-sm text-slate-600 dark:text-slate-400">{t("workspaceDesc")}</p>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <button
          className="flex h-9 w-9 items-center justify-center rounded-lg border border-slate-300 text-slate-700 hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
          onClick={onToggleTheme}
          type="button"
          aria-label="Toggle Dark Mode"
        >
          {theme === "light" ? "🌙" : "☀️"}
        </button>
        <button
          className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
          onClick={onToggleLocale}
          type="button"
        >
          {locale === "en" ? t("switchToKorean") : t("switchToEnglish")}
        </button>
        {authMeta?.authEnabled ? (
          <>
            <span className="rounded-full border border-slate-200 bg-slate-50 px-3 py-1 text-xs font-semibold text-slate-700 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-300">
              {user ? `User: ${user.username}` : t("authEnabled")}
            </span>
            <button
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
              onClick={onOpenCredentials}
              type="button"
            >
              {t("savedConnections")}
            </button>
            <button
              className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-700 hover:bg-slate-100 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700"
              onClick={onOpenHistory}
              type="button"
            >
              {t("myHistory")}
            </button>
            <button
              className="rounded-lg bg-slate-900 px-3 py-2 text-sm font-semibold text-white hover:bg-slate-700 dark:bg-slate-100 dark:text-slate-900 dark:hover:bg-slate-200"
              onClick={onLogout}
              type="button"
            >
              {t("logout")}
            </button>
          </>
        ) : (
          <span className="rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs font-semibold text-emerald-700 dark:border-emerald-900 dark:bg-emerald-900/30 dark:text-emerald-400">
            {t("authDisabled")}
          </span>
        )}
      </div>
    </header>
  );
}
