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
});

describe("App", () => {
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

    await screen.findByRole("heading", { name: "v18 Migration Workspace" });
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
    await screen.findByRole("heading", { name: "v18 Migration Workspace" });

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
    await screen.findByRole("heading", { name: "v18 Migration Workspace" });

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
    await screen.findByRole("heading", { name: "v18 Migration Workspace" });

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
    await screen.findByRole("heading", { name: "v18 Migration Workspace" });

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
    await screen.findByRole("heading", { name: "v18 Migration Workspace" });

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
    await screen.findByRole("heading", { name: "v18 Migration Workspace" });

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
    await screen.findByRole("heading", { name: "v18 Migration Workspace" });

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

    await screen.findByRole("heading", { name: "v18 Migration Workspace" });
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
    await screen.findByRole("heading", { name: "v18 Migration Workspace" });

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
    await screen.findByRole("heading", { name: "v18 Migration Workspace" });

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
    await screen.findByRole("heading", { name: "v18 Migration Workspace" });

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
    await screen.findByRole("heading", { name: "v18 Migration Workspace" });

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
