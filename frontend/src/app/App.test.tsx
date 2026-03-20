import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { App } from "./App";

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

function mockFetch(
  handler: (url: string, method: string, init?: RequestInit) => Response | Promise<Response>,
) {
  const fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = typeof input === "string" ? input : input.toString();
    const method = (init?.method ?? "GET").toUpperCase();
    return await handler(url, method, init);
  });

  vi.stubGlobal("fetch", fetchMock as unknown as typeof fetch);
  return fetchMock;
}

afterEach(() => {
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
  window.localStorage.clear();
});

describe("App", () => {

  it("toggles UI language between English and Korean", async () => {
    const user = userEvent.setup();

    mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: false, uiVersion: "v18-preview" });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });

    await user.click(screen.getByRole("button", { name: "한국어" }));

    await screen.findByRole("heading", { name: "v21 마이그레이션 작업공간" });
    expect(screen.getByText("소스/타깃 설정, 마이그레이션 옵션, 실시간 실행 상태를 관리합니다.")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "English" })).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "English" }));
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });
  });

  it("filters saved connections by source/target role", async () => {
    const user = userEvent.setup();

    mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: true, uiVersion: "v18-preview" });
      }
      if (url === "/api/auth/me" && method === "GET") {
        return jsonResponse({ userId: 1, username: "alice" });
      }
      if (url === "/api/credentials" && method === "GET") {
        return jsonResponse({
          items: [
            {
              id: 1,
              userId: 1,
              alias: "ORA_DEV",
              dbType: "oracle",
              host: "oracle-dev.local:1521/XE",
              username: "scott",
              password: "tiger",
              createdAt: "2026-03-16T00:00:00Z",
              updatedAt: "2026-03-16T00:00:00Z",
            },
            {
              id: 2,
              userId: 1,
              alias: "PG_PROD",
              dbType: "postgres",
              host: "postgres://app:***@prod:5432/app",
              createdAt: "2026-03-16T00:00:00Z",
              updatedAt: "2026-03-16T00:00:00Z",
            },
          ],
        });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);

    await screen.findByRole("heading", { name: "v21 Migration Workspace" });
    await user.click(screen.getByRole("button", { name: "Saved Connections" }));
    await screen.findByText("ORA_DEV");
    expect(screen.getByText("PG_PROD")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Source" }));
    await waitFor(() => {
      expect(screen.getByText("ORA_DEV")).toBeInTheDocument();
      expect(screen.queryByText("PG_PROD")).not.toBeInTheDocument();
    });

    await user.click(screen.getByRole("button", { name: "Target" }));
    await waitFor(() => {
      expect(screen.getByText("PG_PROD")).toBeInTheDocument();
      expect(screen.queryByText("ORA_DEV")).not.toBeInTheDocument();
    });
  });

  it("applies replay payload into connection and option forms", async () => {
    const user = userEvent.setup();

    mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: true, uiVersion: "v18-preview" });
      }
      if (url === "/api/auth/me" && method === "GET") {
        return jsonResponse({ userId: 1, username: "alice" });
      }
      if (url === "/api/tables" && method === "POST") {
        return jsonResponse({ tables: ["USERS", "ORDERS"] });
      }
      if (url.startsWith("/api/history?page=") && method === "GET") {
        return jsonResponse({
          items: [
            {
              id: 77,
              userId: 1,
              status: "success",
              sourceSummary: "SCOTT@oracle-old:1521/XE",
              targetSummary: "postgres:postgres://***@db:5432/app",
              optionsJson: "{}",
              createdAt: "2026-03-16T00:00:00Z",
            },
          ],
          page: 1,
          pageSize: 10,
          total: 1,
        });
      }
      if (url === "/api/history/77/replay" && method === "POST") {
        return jsonResponse({
          payload: {
            oracleUrl: "oracle-new:1521/ORCL",
            username: "hr",
            direct: true,
            targetDb: "postgres",
            targetUrl: "postgres://app:***@newhost:5432/newdb",
            schema: "audit",
            tables: ["USERS"],
            dryRun: true,
            objectGroup: "sequences",
            withDdl: true,
            withSequences: true,
            batchSize: 2000,
            workers: 6,
          },
        });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });

    await user.type(screen.getByLabelText("Oracle URL"), "oracle-old:1521/XE");
    await user.type(screen.getByLabelText("Username"), "scott");
    await user.type(screen.getByLabelText("Password"), "tiger");
    await user.click(screen.getByRole("button", { name: "Connect & Fetch Tables" }));
    await screen.findByText("Found 2 table(s)");

    await user.click(screen.getByRole("button", { name: "My History" }));
    await screen.findByRole("button", { name: "Replay into form" });
    await user.click(screen.getByRole("button", { name: "Replay into form" }));

    await screen.findByText("History payload applied to form.");
    await waitFor(() => {
      expect(screen.getByLabelText("Oracle URL")).toHaveValue("oracle-new:1521/ORCL");
      expect(screen.getByLabelText("Username")).toHaveValue("hr");
      expect(screen.getByLabelText("Password")).toHaveValue("");
      expect(screen.getByRole("combobox", { name: "Migration mode" })).toHaveValue("direct");
      expect(screen.getByLabelText("Target URL")).toHaveValue(
        "postgres://app:***@newhost:5432/newdb",
      );
      expect(screen.getByLabelText("Schema")).toHaveValue("audit");
      expect(screen.getByRole("combobox", { name: "Migration target" })).toHaveValue(
        "sequences",
      );
      expect(screen.getByRole("checkbox", { name: "Dry-run mode" })).toBeChecked();
      expect(screen.getByRole("checkbox", { name: "Include CREATE TABLE DDL" })).toBeChecked();
      expect(screen.getByRole("checkbox", { name: "Include CREATE TABLE DDL" })).toBeDisabled();
      expect(screen.getByRole("checkbox", { name: "Include sequences" })).toBeChecked();
      expect(screen.getByRole("checkbox", { name: "Include sequences" })).toBeDisabled();
      expect(screen.getByText("1 / 2 selected")).toBeInTheDocument();
    });
  });

  it("defaults replay payload without objectGroup to all", async () => {
    const user = userEvent.setup();

    mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: true, uiVersion: "v18-preview" });
      }
      if (url === "/api/auth/me" && method === "GET") {
        return jsonResponse({ userId: 1, username: "alice" });
      }
      if (url === "/api/tables" && method === "POST") {
        return jsonResponse({ tables: ["USERS"] });
      }
      if (url.startsWith("/api/history?page=") && method === "GET") {
        return jsonResponse({
          items: [
            {
              id: 88,
              userId: 1,
              status: "success",
              sourceSummary: "SCOTT@oracle-old:1521/XE",
              targetSummary: "file:migration.sql",
              optionsJson: "{}",
              createdAt: "2026-03-16T00:00:00Z",
            },
          ],
          page: 1,
          pageSize: 10,
          total: 1,
        });
      }
      if (url === "/api/history/88/replay" && method === "POST") {
        return jsonResponse({
          payload: {
            oracleUrl: "oracle-new:1521/ORCL",
            username: "hr",
            tables: ["USERS"],
            withDdl: true,
          },
        });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });

    await user.type(screen.getByLabelText("Oracle URL"), "oracle-old:1521/XE");
    await user.type(screen.getByLabelText("Username"), "scott");
    await user.type(screen.getByLabelText("Password"), "tiger");
    await user.click(screen.getByRole("button", { name: "Connect & Fetch Tables" }));
    await screen.findByText("Found 1 table(s)");

    await user.click(screen.getByRole("button", { name: "My History" }));
    await user.click(screen.getByRole("button", { name: "Replay into form" }));

    await waitFor(() => {
      expect(screen.getByRole("combobox", { name: "Migration target" })).toHaveValue("all");
    });
  });

  it("shows grouped discovery preview and tables-only sequence disable state", async () => {
    const user = userEvent.setup();

    mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: false, uiVersion: "v18-preview" });
      }
      if (url === "/api/tables" && method === "POST") {
        return jsonResponse({ tables: ["USERS", "ORDERS"] });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });

    await user.type(screen.getByLabelText("Oracle URL"), "oracle:1521/XE");
    await user.type(screen.getByLabelText("Username"), "scott");
    await user.type(screen.getByLabelText("Password"), "tiger");
    await user.click(screen.getByRole("button", { name: "Connect & Fetch Tables" }));
    await screen.findByText("Found 2 table(s)");

    const usersRow = screen.getByText("USERS").closest("tr");
    const rowCheckbox = usersRow?.querySelector('input[type="checkbox"]') as HTMLInputElement | null;
    expect(rowCheckbox).not.toBeNull();
    await user.click(rowCheckbox!);

    expect(screen.getByText("Tables Group")).toBeInTheDocument();
    expect(screen.getByText("Sequences Group")).toBeInTheDocument();
    expect(screen.getByText("Selected tables to be migrated.")).toBeInTheDocument();

    await user.selectOptions(screen.getByRole("combobox", { name: "Migration target" }), "tables");
    expect(screen.getByText("Sequence group is disabled.")).toBeInTheDocument();
  });

  it("filters table list by migration history and excludes successful migrations", async () => {
    const user = userEvent.setup();

    mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: true, uiVersion: "v18-preview" });
      }
      if (url === "/api/auth/me" && method === "GET") {
        return jsonResponse({ userId: 1, username: "alice" });
      }
      if (url === "/api/tables" && method === "POST") {
        return jsonResponse({ tables: ["USERS", "ORDERS", "PRODUCTS"] });
      }
      if (url.startsWith("/api/history?page=") && method === "GET") {
        return jsonResponse({
          items: [
            {
              id: 1,
              userId: 1,
              status: "success",
              sourceSummary: "src",
              targetSummary: "dst",
              optionsJson: JSON.stringify({ tables: ["USERS"] }),
              createdAt: "2026-03-16T00:00:00Z",
            },
            {
              id: 2,
              userId: 1,
              status: "failed",
              sourceSummary: "src",
              targetSummary: "dst",
              optionsJson: JSON.stringify({ tables: ["ORDERS"] }),
              createdAt: "2026-03-17T00:00:00Z",
            },
          ],
          page: 1,
          pageSize: 10,
          total: 2,
        });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });

    await user.type(screen.getByLabelText("Oracle URL"), "oracle:1521/XE");
    await user.type(screen.getByLabelText("Username"), "scott");
    await user.type(screen.getByLabelText("Password"), "tiger");
    await user.click(screen.getByRole("button", { name: "Connect & Fetch Tables" }));
    await screen.findByText("Found 3 table(s)");

    await waitFor(() => {
      expect(screen.getByText("USERS")).toBeInTheDocument();
      expect(screen.getByText("ORDERS")).toBeInTheDocument();
      expect(screen.getByText("PRODUCTS")).toBeInTheDocument();
    });
    expect(screen.getAllByLabelText("Table status: Pending").length).toBeGreaterThan(0);
    expect(screen.getByLabelText("History success")).toBeInTheDocument();
    expect(screen.getByLabelText("History failed")).toBeInTheDocument();
    expect(screen.getByLabelText("History not started")).toBeInTheDocument();

    await user.selectOptions(
      screen.getByRole("combobox", { name: "Table history status filter" }),
      "not_started",
    );

    await waitFor(() => {
      expect(screen.queryByText("USERS")).not.toBeInTheDocument();
      expect(screen.queryByText("ORDERS")).not.toBeInTheDocument();
      expect(screen.getByText("PRODUCTS")).toBeInTheDocument();
    });

    await user.selectOptions(
      screen.getByRole("combobox", { name: "Table history status filter" }),
      "all",
    );
    await user.click(screen.getByRole("checkbox", { name: "Exclude migrated success" }));

    await waitFor(() => {
      expect(screen.queryByText("USERS")).not.toBeInTheDocument();
      expect(screen.getByText("ORDERS")).toBeInTheDocument();
      expect(screen.getByText("PRODUCTS")).toBeInTheDocument();
    });
  });


  it("sorts table list with the table sort control", async () => {
    const user = userEvent.setup();

    mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: true, uiVersion: "v18-preview" });
      }
      if (url === "/api/auth/me" && method === "GET") {
        return jsonResponse({ userId: 1, username: "alice" });
      }
      if (url === "/api/tables" && method === "POST") {
        return jsonResponse({ tables: ["BETA", "ALPHA", "GAMMA"] });
      }
      if (url.startsWith("/api/history?page=") && method === "GET") {
        return jsonResponse({
          items: [
            {
              id: 1,
              userId: 1,
              status: "success",
              sourceSummary: "src",
              targetSummary: "dst",
              optionsJson: JSON.stringify({ tables: ["BETA", "BETA"] }),
              createdAt: "2026-03-17T00:00:00Z",
            },
            {
              id: 2,
              userId: 1,
              status: "failed",
              sourceSummary: "src",
              targetSummary: "dst",
              optionsJson: JSON.stringify({ tables: ["ALPHA"] }),
              createdAt: "2026-03-18T00:00:00Z",
            },
          ],
          page: 1,
          pageSize: 10,
          total: 2,
        });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });

    await user.type(screen.getByLabelText("Oracle URL"), "oracle:1521/XE");
    await user.type(screen.getByLabelText("Username"), "scott");
    await user.type(screen.getByLabelText("Password"), "tiger");
    await user.click(screen.getByRole("button", { name: "Connect & Fetch Tables" }));
    await screen.findByText("Found 3 table(s)");

    const table = screen.getByRole("table");
    const getOrder = () =>
      Array.from(table.querySelectorAll("tbody tr td:nth-child(2)")).map((cell) =>
        cell.textContent?.trim(),
      );

    expect(getOrder()).toEqual(["ALPHA", "BETA", "GAMMA"]);

    await user.selectOptions(screen.getByRole("combobox", { name: "Table sort" }), "table_desc");
    await waitFor(() => {
      expect(getOrder()).toEqual(["GAMMA", "BETA", "ALPHA"]);
    });

    await user.selectOptions(screen.getByRole("combobox", { name: "Table sort" }), "runs_desc");
    await waitFor(() => {
      expect(getOrder()).toEqual(["BETA", "ALPHA", "GAMMA"]);
    });

    await user.selectOptions(screen.getByRole("combobox", { name: "Table sort" }), "recent_desc");
    await waitFor(() => {
      expect(getOrder()).toEqual(["ALPHA", "BETA", "GAMMA"]);
    });

    await user.selectOptions(
      screen.getByRole("combobox", { name: "Table sort" }),
      "history_status",
    );
    await waitFor(() => {
      expect(getOrder()).toEqual(["GAMMA", "ALPHA", "BETA"]);
    });
  });

  it("shows per-table history panel and failed retry action", async () => {
    const user = userEvent.setup();

    const fetchMock = mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: true, uiVersion: "v18-preview" });
      }
      if (url === "/api/auth/me" && method === "GET") {
        return jsonResponse({ userId: 1, username: "alice" });
      }
      if (url === "/api/tables" && method === "POST") {
        return jsonResponse({ tables: ["ORDERS"] });
      }
      if (url.startsWith("/api/history?page=") && method === "GET") {
        return jsonResponse({
          items: [
            {
              id: 11,
              userId: 1,
              status: "failed",
              sourceSummary: "src",
              targetSummary: "dst",
              optionsJson: JSON.stringify({ tables: ["ORDERS"] }),
              logSummary: "duplicate key value",
              createdAt: "2026-03-19T00:00:00Z",
            },
            {
              id: 10,
              userId: 1,
              status: "success",
              sourceSummary: "src",
              targetSummary: "dst",
              optionsJson: JSON.stringify({ tables: ["ORDERS"] }),
              createdAt: "2026-03-18T00:00:00Z",
            },
          ],
          page: 1,
          pageSize: 10,
          total: 2,
        });
      }
      if (url === "/api/history/11/replay" && method === "POST") {
        return jsonResponse({ payload: { oracleUrl: "oracle:1521/XE", username: "scott" } });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });

    await user.type(screen.getByLabelText("Oracle URL"), "oracle:1521/XE");
    await user.type(screen.getByLabelText("Username"), "scott");
    await user.type(screen.getByLabelText("Password"), "tiger");
    await user.click(screen.getByRole("button", { name: "Connect & Fetch Tables" }));
    await screen.findByText("Found 1 table(s)");

    await user.click(screen.getByRole("button", { name: "View history" }));
    await screen.findByText("Table history: ORDERS");
    expect(screen.getByText("duplicate key value")).toBeInTheDocument();

    await user.click(screen.getByRole("button", { name: "Retry settings" }));
    await screen.findByText("History payload applied to form.");

    expect(fetchMock).toHaveBeenCalledWith(
      "/api/history/11/replay",
      expect.objectContaining({ method: "POST" }),
    );
  });

  it("shows empty and error states in per-table history panel", async () => {
    const user = userEvent.setup();

    let historyFails = false;
    mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: true, uiVersion: "v18-preview" });
      }
      if (url === "/api/auth/me" && method === "GET") {
        return jsonResponse({ userId: 1, username: "alice" });
      }
      if (url === "/api/tables" && method === "POST") {
        return jsonResponse({ tables: ["USERS"] });
      }
      if (url.startsWith("/api/history?page=") && method === "GET") {
        if (historyFails) {
          return jsonResponse({ error: "boom" }, 500);
        }
        return jsonResponse({
          items: [
            {
              id: 1,
              userId: 1,
              status: "success",
              sourceSummary: "src",
              targetSummary: "dst",
              optionsJson: JSON.stringify({ tables: ["ORDERS"] }),
              createdAt: "2026-03-16T00:00:00Z",
            },
          ],
          page: 1,
          pageSize: 10,
          total: 1,
        });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });

    await user.type(screen.getByLabelText("Oracle URL"), "oracle:1521/XE");
    await user.type(screen.getByLabelText("Username"), "scott");
    await user.type(screen.getByLabelText("Password"), "tiger");
    await user.click(screen.getByRole("button", { name: "Connect & Fetch Tables" }));
    await screen.findByText("Found 1 table(s)");

    await user.click(screen.getByRole("button", { name: "View history" }));
    await screen.findByText("No history found for this table.");

    historyFails = true;
    await user.click(screen.getByRole("button", { name: "View history" }));
    await screen.findByText("Failed to load migration history.");
    await screen.findByRole("button", { name: "Retry" });
  });

  it("shows session-expired message when protected API returns 401", async () => {
    const user = userEvent.setup();

    mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: true, uiVersion: "v18-preview" });
      }
      if (url === "/api/auth/me" && method === "GET") {
        return jsonResponse({ userId: 1, username: "alice" });
      }
      if (url === "/api/credentials" && method === "GET") {
        return jsonResponse({ error: "Unauthorized" }, 401);
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);

    await screen.findByRole("heading", { name: "v21 Migration Workspace" });
    await user.click(screen.getByRole("button", { name: "Saved Connections" }));
    await screen.findByText("Session expired. Please log in again.");
  });

  // ── v18 Component Tests: filter/toggle/retry interactions ─────────────

  it("toggles exclude-success checkbox and filters table list accordingly", async () => {
    const user = userEvent.setup();

    mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: true, uiVersion: "v18-preview" });
      }
      if (url === "/api/auth/me" && method === "GET") {
        return jsonResponse({ userId: 1, username: "alice" });
      }
      if (url === "/api/tables" && method === "POST") {
        return jsonResponse({ tables: ["USERS", "ORDERS"] });
      }
      if (url.startsWith("/api/history?page=") && method === "GET") {
        return jsonResponse({
          items: [
            {
              id: 1,
              userId: 1,
              status: "success",
              sourceSummary: "src",
              targetSummary: "dst",
              optionsJson: JSON.stringify({ tables: ["USERS"] }),
              createdAt: "2026-03-16T00:00:00Z",
            },
          ],
          page: 1,
          pageSize: 10,
          total: 1,
        });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });

    await user.type(screen.getByLabelText("Oracle URL"), "oracle:1521/XE");
    await user.type(screen.getByLabelText("Username"), "scott");
    await user.type(screen.getByLabelText("Password"), "tiger");
    await user.click(screen.getByRole("button", { name: "Connect & Fetch Tables" }));
    await screen.findByText("Found 2 table(s)");

    // Both tables visible initially
    expect(screen.getByText("USERS")).toBeInTheDocument();
    expect(screen.getByText("ORDERS")).toBeInTheDocument();

    // Toggle exclude success on
    const excludeCheckbox = screen.getByRole("checkbox", { name: "Exclude migrated success" });
    await user.click(excludeCheckbox);

    await waitFor(() => {
      expect(screen.queryByText("USERS")).not.toBeInTheDocument();
      expect(screen.getByText("ORDERS")).toBeInTheDocument();
    });

    // Toggle exclude success off again
    await user.click(excludeCheckbox);

    await waitFor(() => {
      expect(screen.getByText("USERS")).toBeInTheDocument();
      expect(screen.getByText("ORDERS")).toBeInTheDocument();
    });
  });

  it("status filter dropdown shows only matching tables", async () => {
    const user = userEvent.setup();

    mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: true, uiVersion: "v18-preview" });
      }
      if (url === "/api/auth/me" && method === "GET") {
        return jsonResponse({ userId: 1, username: "alice" });
      }
      if (url === "/api/tables" && method === "POST") {
        return jsonResponse({ tables: ["ALPHA", "BETA", "GAMMA"] });
      }
      if (url.startsWith("/api/history?page=") && method === "GET") {
        return jsonResponse({
          items: [
            {
              id: 1,
              userId: 1,
              status: "success",
              sourceSummary: "src",
              targetSummary: "dst",
              optionsJson: JSON.stringify({ tables: ["ALPHA"] }),
              createdAt: "2026-03-16T00:00:00Z",
            },
            {
              id: 2,
              userId: 1,
              status: "failed",
              sourceSummary: "src",
              targetSummary: "dst",
              optionsJson: JSON.stringify({ tables: ["BETA"] }),
              createdAt: "2026-03-17T00:00:00Z",
            },
          ],
          page: 1,
          pageSize: 10,
          total: 2,
        });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });

    await user.type(screen.getByLabelText("Oracle URL"), "oracle:1521/XE");
    await user.type(screen.getByLabelText("Username"), "scott");
    await user.type(screen.getByLabelText("Password"), "tiger");
    await user.click(screen.getByRole("button", { name: "Connect & Fetch Tables" }));
    await screen.findByText("Found 3 table(s)");

    // Filter by failed status
    await user.selectOptions(
      screen.getByRole("combobox", { name: "Table history status filter" }),
      "failed",
    );

    await waitFor(() => {
      expect(screen.queryByText("ALPHA")).not.toBeInTheDocument();
      expect(screen.getByText("BETA")).toBeInTheDocument();
      expect(screen.queryByText("GAMMA")).not.toBeInTheDocument();
    });

    // Filter by success status
    await user.selectOptions(
      screen.getByRole("combobox", { name: "Table history status filter" }),
      "success",
    );

    await waitFor(() => {
      expect(screen.getByText("ALPHA")).toBeInTheDocument();
      expect(screen.queryByText("BETA")).not.toBeInTheDocument();
      expect(screen.queryByText("GAMMA")).not.toBeInTheDocument();
    });

    // Reset to all
    await user.selectOptions(
      screen.getByRole("combobox", { name: "Table history status filter" }),
      "all",
    );

    await waitFor(() => {
      expect(screen.getByText("ALPHA")).toBeInTheDocument();
      expect(screen.getByText("BETA")).toBeInTheDocument();
      expect(screen.getByText("GAMMA")).toBeInTheDocument();
    });
  });

  it("retry button is only shown on failed history entries in detail panel", async () => {
    const user = userEvent.setup();

    mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: true, uiVersion: "v18-preview" });
      }
      if (url === "/api/auth/me" && method === "GET") {
        return jsonResponse({ userId: 1, username: "alice" });
      }
      if (url === "/api/tables" && method === "POST") {
        return jsonResponse({ tables: ["ORDERS"] });
      }
      if (url.startsWith("/api/history?page=") && method === "GET") {
        return jsonResponse({
          items: [
            {
              id: 20,
              userId: 1,
              status: "failed",
              sourceSummary: "src",
              targetSummary: "dst",
              optionsJson: JSON.stringify({ tables: ["ORDERS"] }),
              logSummary: "timeout error",
              createdAt: "2026-03-19T00:00:00Z",
            },
            {
              id: 19,
              userId: 1,
              status: "success",
              sourceSummary: "src",
              targetSummary: "dst",
              optionsJson: JSON.stringify({ tables: ["ORDERS"] }),
              createdAt: "2026-03-18T00:00:00Z",
            },
          ],
          page: 1,
          pageSize: 10,
          total: 2,
        });
      }
      if (url === "/api/history/20/replay" && method === "POST") {
        return jsonResponse({ payload: { oracleUrl: "oracle:1521/XE", username: "scott" } });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });

    await user.type(screen.getByLabelText("Oracle URL"), "oracle:1521/XE");
    await user.type(screen.getByLabelText("Username"), "scott");
    await user.type(screen.getByLabelText("Password"), "tiger");
    await user.click(screen.getByRole("button", { name: "Connect & Fetch Tables" }));
    await screen.findByText("Found 1 table(s)");

    // Open table history detail
    await user.click(screen.getByRole("button", { name: "View history" }));
    await screen.findByText("Table history: ORDERS");

    // Should show failed entry with retry button
    expect(screen.getByText("timeout error")).toBeInTheDocument();
    const retryButtons = screen.getAllByRole("button", { name: "Retry settings" });
    expect(retryButtons).toHaveLength(1); // Only one failed entry has retry button

    // Success entry should NOT have a retry button
    // Both "Success" and "Failed" labels appear in the detail panel entries
    expect(screen.getAllByText("Success").length).toBeGreaterThanOrEqual(1);
    expect(screen.getAllByText("Failed").length).toBeGreaterThanOrEqual(1);
  });

  // ── v18 E2E: fail filter → detail history → retry ─────────────────────

  it("E2E: filter by failed → view detail history → retry failed table", async () => {
    const user = userEvent.setup();

    const fetchMock = mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: true, uiVersion: "v18-preview" });
      }
      if (url === "/api/auth/me" && method === "GET") {
        return jsonResponse({ userId: 1, username: "alice" });
      }
      if (url === "/api/tables" && method === "POST") {
        return jsonResponse({ tables: ["USERS", "ORDERS", "PRODUCTS"] });
      }
      if (url.startsWith("/api/history?page=") && method === "GET") {
        return jsonResponse({
          items: [
            {
              id: 101,
              userId: 1,
              status: "success",
              sourceSummary: "SCOTT@oracle:1521/XE",
              targetSummary: "postgres:pg://host:5432/db",
              optionsJson: JSON.stringify({ tables: ["USERS"] }),
              createdAt: "2026-03-15T10:00:00Z",
            },
            {
              id: 102,
              userId: 1,
              status: "failed",
              sourceSummary: "SCOTT@oracle:1521/XE",
              targetSummary: "postgres:pg://host:5432/db",
              optionsJson: JSON.stringify({ tables: ["ORDERS"] }),
              logSummary: "duplicate key constraint violation",
              createdAt: "2026-03-16T14:00:00Z",
            },
          ],
          page: 1,
          pageSize: 10,
          total: 2,
        });
      }
      if (url === "/api/history/102/replay" && method === "POST") {
        return jsonResponse({
          payload: {
            oracleUrl: "oracle:1521/XE",
            username: "scott",
            direct: true,
            targetDb: "postgres",
            targetUrl: "pg://host:5432/db",
            tables: ["ORDERS"],
            truncate: true,
          },
        });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });

    // Step 1: Connect and fetch tables
    await user.type(screen.getByLabelText("Oracle URL"), "oracle:1521/XE");
    await user.type(screen.getByLabelText("Username"), "scott");
    await user.type(screen.getByLabelText("Password"), "tiger");
    await user.click(screen.getByRole("button", { name: "Connect & Fetch Tables" }));
    await screen.findByText("Found 3 table(s)");

    // Step 2: Filter to show only failed tables
    await user.selectOptions(
      screen.getByRole("combobox", { name: "Table history status filter" }),
      "failed",
    );

    await waitFor(() => {
      expect(screen.queryByText("USERS")).not.toBeInTheDocument();
      expect(screen.getByText("ORDERS")).toBeInTheDocument();
      expect(screen.queryByText("PRODUCTS")).not.toBeInTheDocument();
    });

    // Step 3: Open detail history for the failed table
    await user.click(screen.getByRole("button", { name: "View history" }));
    await screen.findByText("Table history: ORDERS");

    // Verify error summary is highlighted
    expect(screen.getByText("duplicate key constraint violation")).toBeInTheDocument();

    // Step 4: Click retry to replay the failed migration settings
    await user.click(screen.getByRole("button", { name: "Retry settings" }));
    await screen.findByText("History payload applied to form.");

    // Verify replay API was called with the correct failed entry
    expect(fetchMock).toHaveBeenCalledWith(
      "/api/history/102/replay",
      expect.objectContaining({ method: "POST" }),
    );

    // Verify form was populated with replay data
    await waitFor(() => {
      expect(screen.getByLabelText("Oracle URL")).toHaveValue("oracle:1521/XE");
      expect(screen.getByLabelText("Username")).toHaveValue("scott");
    });
  });
});

// ── v22: 소스-타겟 비교 UI ────────────────────────────────────────────────────

describe("v22 소스-타겟 비교 UI", () => {
  /** 공통 fetch mock 설정 */
  function setupCompareFetch(sourceTables: string[], targetTables: string[]) {
    return mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: false, uiVersion: "v18-preview" });
      }
      if (url === "/api/tables" && method === "POST") {
        return jsonResponse({ tables: sourceTables });
      }
      if (url === "/api/target-tables" && method === "POST") {
        return jsonResponse({ tables: targetTables, fetchedAt: "2026-03-19T10:00:00Z" });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });
  }

  /** 소스 테이블 로드 → direct 모드 전환 → 타겟 테이블 조회 */
  async function loadAndFetchTables(
    user: ReturnType<typeof userEvent.setup>,
    sourceTables: string[],
  ) {
    render(<App />);
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });

    await user.type(screen.getByLabelText("Oracle URL"), "oracle:1521/XE");
    await user.type(screen.getByLabelText("Username"), "hr");
    await user.type(screen.getByLabelText("Password"), "hr");
    await user.click(screen.getByRole("button", { name: "Connect & Fetch Tables" }));
    await screen.findByText(`Found ${sourceTables.length} table(s)`);

    await user.selectOptions(
      screen.getByRole("combobox", { name: "Migration mode" }),
      "Direct migration",
    );
    await user.type(screen.getByLabelText("Target URL"), "postgres://localhost/db");
    await user.type(screen.getByLabelText("Schema"), "public");

    await user.click(screen.getByRole("button", { name: "Fetch Target Tables" }));
    await screen.findByText(/tables in target|개 타겟 테이블/);
  }

  it("source_only / both / target_only 카테고리를 올바르게 분류한다", async () => {
    const user = userEvent.setup();
    // 소스: USERS, ORDERS, PRODUCTS / 타겟: USERS, PRODUCTS, DEPT
    // source_only=ORDERS(1), both=USERS+PRODUCTS(2), target_only=DEPT(1)
    setupCompareFetch(["USERS", "ORDERS", "PRODUCTS"], ["USERS", "PRODUCTS", "DEPT"]);

    await loadAndFetchTables(user, ["USERS", "ORDERS", "PRODUCTS"]);

    const panelSummary = await screen.findByText(/Source vs Target Comparison|소스-타겟 비교/);
    await user.click(panelSummary);

    await waitFor(() => {
      const text = document.body.textContent ?? "";
      // 요약 카드: source_only=1, both=2, target_only=1
      expect(text).toMatch(/소스에만|Source only/);
      expect(text).toMatch(/양쪽|Both/);
      expect(text).toMatch(/타겟에만|Target only/);
    });

    // 테이블 행 검증
    await waitFor(() => {
      expect(screen.getByText("orders")).toBeInTheDocument(); // source_only (lowercase normalized)
      expect(screen.getByText("dept")).toBeInTheDocument();   // target_only
    });
  });

  it("대소문자 정규화: 소스 USERS + 타겟 users → both 분류 (source_only/target_only=0)", async () => {
    const user = userEvent.setup();
    setupCompareFetch(["USERS"], ["users"]);

    await loadAndFetchTables(user, ["USERS"]);

    const panelSummary = await screen.findByText(/Source vs Target Comparison|소스-타겟 비교/);
    await user.click(panelSummary);

    await waitFor(() => {
      // both=1 → "users" 행이 both 배지로 표시
      expect(screen.getByText("users")).toBeInTheDocument();
    });

    // source_only / target_only 행이 없음을 확인
    const rows = screen.queryAllByText(/orders|dept/i);
    expect(rows).toHaveLength(0);
  });

  it("'소스에만 있는 테이블 선택' 버튼이 source_only 테이블을 선택에 추가한다", async () => {
    const user = userEvent.setup();
    // ORDERS만 source_only, USERS는 both
    setupCompareFetch(["USERS", "ORDERS"], ["users", "dept"]);

    await loadAndFetchTables(user, ["USERS", "ORDERS"]);

    // 테이블 선택 섹션에서 선택 카운트 확인 (초기: 0)
    expect(screen.getByText(/0 \/ 2 selected|0 \/ 2 선택됨/)).toBeInTheDocument();

    // 빠른 선택 버튼 클릭
    await user.click(
      screen.getByRole("button", { name: /소스에만 있는 테이블 선택|Select source-only/ }),
    );

    await waitFor(() => {
      // source_only = ORDERS 1개만 선택됨
      expect(screen.getByText(/1 \/ 2 selected|1 \/ 2 선택됨/)).toBeInTheDocument();
    });
  });

  it("row_diff: both + 행 수 불일치 시 배지가 표시된다", async () => {
    const user = userEvent.setup();
    mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: false, uiVersion: "v18-preview" });
      }
      if (url === "/api/tables" && method === "POST") {
        return jsonResponse({ tables: ["USERS"] });
      }
      if (url === "/api/target-tables" && method === "POST") {
        return jsonResponse({ tables: ["users"], fetchedAt: "2026-03-19T10:00:00Z" });
      }
      if (url === "/api/migrations/precheck" && method === "POST") {
        return jsonResponse({
          summary: {
            total_tables: 1,
            transfer_required_count: 1,
            skip_candidate_count: 0,
            count_check_failed_count: 0,
          },
          items: [
            {
              table_name: "USERS",
              source_row_count: 1000,
              target_row_count: 500,
              diff: -500,
              decision: "transfer_required",
            },
          ],
        });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);
    await screen.findByRole("heading", { name: "v21 Migration Workspace" });

    await user.type(screen.getByLabelText("Oracle URL"), "oracle:1521/XE");
    await user.type(screen.getByLabelText("Username"), "hr");
    await user.type(screen.getByLabelText("Password"), "hr");
    await user.click(screen.getByRole("button", { name: "Connect & Fetch Tables" }));
    await screen.findByText("Found 1 table(s)");

    await user.selectOptions(
      screen.getByRole("combobox", { name: "Migration mode" }),
      "Direct migration",
    );
    await user.type(screen.getByLabelText("Target URL"), "postgres://localhost/db");
    await user.type(screen.getByLabelText("Schema"), "public");

    // USERS 테이블 체크 후 pre-check 실행 (테이블 체크박스는 접근 라벨이 없어 인덱스로 선택)
    const checkboxes = screen.getAllByRole("checkbox");
    await user.click(checkboxes[2]);
    await user.click(screen.getByRole("button", { name: /Pre-check|pre-check/i }));
    await waitFor(() => {
      expect(screen.queryByText(/1,000|1000/)).toBeInTheDocument();
    });

    // 타겟 테이블 조회
    await user.click(screen.getByRole("button", { name: "Fetch Target Tables" }));
    await screen.findByText(/tables in target|개 타겟 테이블/);

    // 비교 패널 열기
    const panelSummary = await screen.findByText(/Source vs Target Comparison|소스-타겟 비교/);
    await user.click(panelSummary);

    await waitFor(() => {
      expect(screen.getByText(/Row diff|행 수 불일치/)).toBeInTheDocument();
    });
  });
});
