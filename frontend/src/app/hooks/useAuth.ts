import { FormEvent, useEffect, useState } from "react";
import { apiRequest } from "../../shared/api/client";
import { AuthUser, RuntimeMeta } from "../../shared/api/types";
import { NoticeTone } from "../types";

export interface UseAuthProps {
  resetRunState: () => void;
  setCredentialsPanelOpen: (open: boolean) => void;
  setHistoryPanelOpen: (open: boolean) => void;
  setNotice: (notice: { text: string; tone: NoticeTone } | null) => void;
}

export function useAuth({
  resetRunState,
  setCredentialsPanelOpen,
  setHistoryPanelOpen,
  setNotice,
}: UseAuthProps) {
  const [meta, setMeta] = useState<RuntimeMeta | null>(null);
  const [user, setUser] = useState<AuthUser | null>(null);
  const [booting, setBooting] = useState(true);
  const [bootError, setBootError] = useState("");

  const [loginForm, setLoginForm] = useState({ username: "", password: "" });
  const [loginBusy, setLoginBusy] = useState(false);
  const [loginError, setLoginError] = useState("");

  useEffect(() => {
    void boot();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  async function boot() {
    setBooting(true);
    setBootError("");
    try {
      const { response, data } = await apiRequest<RuntimeMeta>("/api/meta", {}, {
        allowUnauthorized: true,
      });
      if (!response.ok) {
        throw new Error("Failed to load runtime metadata.");
      }
      setMeta(data);

      if (!data.authEnabled) {
        setUser(null);
        return;
      }

      const me = await apiRequest<AuthUser | { error: string }>("/api/auth/me", {}, {
        allowUnauthorized: true,
      });
      if (me.response.ok) {
        setUser(me.data as AuthUser);
      } else {
        setUser(null);
      }
    } catch (error) {
      setBootError(error instanceof Error ? error.message : "Unknown boot error");
    } finally {
      setBooting(false);
    }
  }

  async function handleLogin(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoginError("");
    setLoginBusy(true);
    try {
      const { response, data } = await apiRequest<AuthUser | { error: string }>(
        "/api/auth/login",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(loginForm),
        },
        { allowUnauthorized: true },
      );
      if (!response.ok) {
        const message = (data as { error?: string }).error ?? "Login failed";
        throw new Error(message);
      }
      setUser(data as AuthUser);
      setLoginForm((prev) => ({ ...prev, password: "" }));
      setNotice({ text: "Logged in successfully.", tone: "info" });
    } catch (error) {
      setLoginError(error instanceof Error ? error.message : "Login failed");
    } finally {
      setLoginBusy(false);
    }
  }

  async function handleLogout() {
    await apiRequest("/api/auth/logout", { method: "POST" }, { allowUnauthorized: true });
    resetRunState();
    setUser(null);
    setCredentialsPanelOpen(false);
    setHistoryPanelOpen(false);
    setNotice({ text: "Logged out.", tone: "info" });
  }

  async function handleGoogleLogin(credential: string) {
    setLoginError("");
    setLoginBusy(true);
    try {
      const { response, data } = await apiRequest<AuthUser | { error: string }>(
        "/api/auth/google",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ credential }),
        },
        { allowUnauthorized: true },
      );
      if (!response.ok) {
        const message = (data as { error?: string }).error ?? "Google login failed";
        throw new Error(message);
      }
      setUser(data as AuthUser);
      setNotice({ text: "Logged in with Google successfully.", tone: "info" });
    } catch (error) {
      setLoginError(error instanceof Error ? error.message : "Google login failed");
    } finally {
      setLoginBusy(false);
    }
  }

  return {
    meta,
    user,
    booting,
    bootError,
    loginForm,
    loginBusy,
    loginError,
    setLoginForm,
    boot,
    handleLogin,
    handleLogout,
    handleGoogleLogin,
  };
}
