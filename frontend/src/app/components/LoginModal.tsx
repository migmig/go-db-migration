import { FormEvent } from "react";

type LoginModalProps = {
  loginForm: { username: string; password: string };
  loginBusy: boolean;
  loginError: string;
  tr: (en: string, ko: string) => string;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onUsernameChange: (value: string) => void;
  onPasswordChange: (value: string) => void;
  onGoogleLogin: (credential: string) => void;
};

export function LoginModal({
  loginForm,
  loginBusy,
  loginError,
  tr,
  onSubmit,
  onUsernameChange,
  onPasswordChange,
  onGoogleLogin,
}: LoginModalProps) {
  return (
    <div className="fixed inset-0 z-40 flex items-center justify-center bg-slate-900/45 px-4">
      <form className="card-surface w-full max-w-sm p-6" onSubmit={onSubmit}>
        <h3 className="text-xl font-semibold text-slate-900">{tr("Sign in", "로그인")}</h3>
        <p className="mt-1 text-sm text-slate-600">
          {tr(
            "Auth mode is enabled. Log in to use saved connections and history.",
            "인증 모드가 활성화되어 있습니다. 저장된 연결과 이력을 사용하려면 로그인하세요.",
          )}
        </p>
        <label className="mt-4 block text-sm">
          <span className="mb-1 block text-slate-700">{tr("Username", "사용자명")}</span>
          <input
            className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
            onChange={(event) => onUsernameChange(event.target.value)}
            required
            value={loginForm.username}
          />
        </label>
        <label className="mt-3 block text-sm">
          <span className="mb-1 block text-slate-700">{tr("Password", "비밀번호")}</span>
          <input
            className="w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 focus:ring-2 focus:ring-brand-200"
            onChange={(event) => onPasswordChange(event.target.value)}
            required
            type="password"
            value={loginForm.password}
          />
        </label>
        {loginError && <p className="mt-3 text-sm font-medium text-red-600">{loginError}</p>}
        <button
          className="mt-4 w-full rounded-xl bg-brand-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-brand-700 disabled:cursor-not-allowed disabled:opacity-60"
          disabled={loginBusy}
          type="submit"
        >
          {loginBusy ? tr("Signing in...", "로그인 중...") : tr("Sign in", "로그인")}
        </button>

        <div className="my-4 flex items-center gap-3">
          <div className="h-px flex-1 bg-slate-200"></div>
          <span className="text-xs font-medium text-slate-400 uppercase">{tr("OR", "또는")}</span>
          <div className="h-px flex-1 bg-slate-200"></div>
        </div>

        <div id="google-login-btn" className="flex justify-center"></div>
      </form>
    </div>
  );
}
