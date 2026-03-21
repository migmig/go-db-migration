import { renderHook, act, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { useAuth } from "./useAuth";
import * as client from "../../shared/api/client";

vi.mock("../../shared/api/client", () => ({
  apiRequest: vi.fn(),
}));

describe("useAuth", () => {
  const mockProps = {
    resetRunState: vi.fn(),
    setCredentialsPanelOpen: vi.fn(),
    setHistoryPanelOpen: vi.fn(),
    setNotice: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("boots and fetches metadata and current user", async () => {
    vi.mocked(client.apiRequest).mockImplementation(async (url) => {
      if (url === "/api/meta") return { response: { ok: true }, data: { authEnabled: true } } as any;
      if (url === "/api/auth/me") return { response: { ok: true }, data: { userId: 1, username: "alice" } } as any;
      return { response: { ok: false } } as any;
    });

    const { result } = renderHook(() => useAuth(mockProps));

    await waitFor(() => expect(result.current.booting).toBe(false));
    expect(result.current.meta?.authEnabled).toBe(true);
    expect(result.current.user?.username).toBe("alice");
  });

  it("handles login flow", async () => {
    // Initial boot mocks
    vi.mocked(client.apiRequest).mockImplementation(async (url) => {
      if (url === "/api/meta") return { response: { ok: true }, data: { authEnabled: true } } as any;
      if (url === "/api/auth/me") return { response: { ok: false } } as any;
      return { response: { ok: false } } as any;
    });

    const { result } = renderHook(() => useAuth(mockProps));
    await waitFor(() => expect(result.current.booting).toBe(false));
    
    // Set form data
    act(() => {
      result.current.setLoginForm({ username: "bob", password: "password" });
    });

    // Mock login call
    vi.mocked(client.apiRequest).mockResolvedValueOnce({ 
      response: { ok: true }, 
      data: { userId: 1, username: "bob" } 
    } as any);

    await act(async () => {
      await result.current.handleLogin(new Event("submit") as any);
    });

    expect(result.current.user?.username).toBe("bob");
    expect(mockProps.setNotice).toHaveBeenCalledWith(expect.objectContaining({ text: "Logged in successfully." }));
  });

  it("handles login error", async () => {
    vi.mocked(client.apiRequest).mockImplementation(async (url) => {
      if (url === "/api/meta") return { response: { ok: true }, data: { authEnabled: true } } as any;
      if (url === "/api/auth/me") return { response: { ok: false } } as any;
      return { response: { ok: false } } as any;
    });

    const { result } = renderHook(() => useAuth(mockProps));
    await waitFor(() => expect(result.current.booting).toBe(false));

    vi.mocked(client.apiRequest).mockResolvedValueOnce({ 
      response: { ok: false }, 
      data: { error: "Invalid credentials" } 
    } as any);

    await act(async () => {
      await result.current.handleLogin(new Event("submit") as any);
    });

    expect(result.current.loginError).toBe("Invalid credentials");
    expect(result.current.user).toBeNull();
  });
});
