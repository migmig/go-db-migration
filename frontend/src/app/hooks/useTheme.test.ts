import { renderHook, act } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { useTheme } from "./useTheme";

describe("useTheme", () => {
  const mockClassList = {
    add: vi.fn(),
    remove: vi.fn(),
  };

  beforeEach(() => {
    vi.stubGlobal("window", {
      ...window,
      matchMedia: vi.fn().mockReturnValue({ 
        matches: false,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
      }),
      localStorage: {
        getItem: vi.fn(),
        setItem: vi.fn(),
      },
      document: {
        documentElement: {
          classList: mockClassList,
        }
      }
    });
    vi.clearAllMocks();
  });

  it("defaults to light theme", () => {
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe("light");
  });

  it("toggles theme", () => {
    const { result } = renderHook(() => useTheme());
    act(() => {
      result.current.toggleTheme();
    });
    expect(result.current.theme).toBe("dark");
    expect(mockClassList.add).toHaveBeenCalledWith("dark");
    
    act(() => {
      result.current.toggleTheme();
    });
    expect(result.current.theme).toBe("light");
    expect(mockClassList.remove).toHaveBeenCalledWith("dark");
  });

  it("loads theme from localStorage", () => {
    vi.mocked(window.localStorage.getItem).mockReturnValue("dark");
    const { result } = renderHook(() => useTheme());
    expect(result.current.theme).toBe("dark");
  });
});
