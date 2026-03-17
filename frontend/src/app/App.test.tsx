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
        return jsonResponse({ authEnabled: true, uiVersion: "v16-preview" });
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

    await screen.findByRole("heading", { name: "v16 Migration Workspace" });
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
        return jsonResponse({ authEnabled: true, uiVersion: "v16-preview" });
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
    await screen.findByRole("heading", { name: "v16 Migration Workspace" });

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
        return jsonResponse({ authEnabled: true, uiVersion: "v16-preview" });
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
    await screen.findByRole("heading", { name: "v16 Migration Workspace" });

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
        return jsonResponse({ authEnabled: false, uiVersion: "v16-preview" });
      }
      if (url === "/api/tables" && method === "POST") {
        return jsonResponse({ tables: ["USERS", "ORDERS"] });
      }
      return jsonResponse({ error: `Unhandled: ${method} ${url}` }, 500);
    });

    render(<App />);
    await screen.findByRole("heading", { name: "v16 Migration Workspace" });

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

  it("shows session-expired message when protected API returns 401", async () => {
    const user = userEvent.setup();

    mockFetch((url, method) => {
      if (url === "/api/meta" && method === "GET") {
        return jsonResponse({ authEnabled: true, uiVersion: "v16-preview" });
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

    await screen.findByRole("heading", { name: "v16 Migration Workspace" });
    await user.click(screen.getByRole("button", { name: "Saved Connections" }));
    await screen.findByText("Session expired. Please log in again.");
  });
});
