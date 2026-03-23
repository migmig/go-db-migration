import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { RunStatusPanel } from "./RunStatusPanel";

const defaultProps = {
  ddlEvents: [],
  effectiveObjectGroup: "all" as const,
  elapsedSeconds: 60,
  etaSeconds: 120,
  groupSummary: null,
  locale: "en" as const,
  metrics: { cpu: "15", mem: "256" },
  migrationBusy: true,
  objectGroupModeEnabled: true,
  onResetRunState: vi.fn(),
  overallPercent: 33,
  processedRows: 5000,
  reportSummary: null,
  rowsPerSecond: 83,
  runDoneTables: 2,
  runDryRun: false,
  runEntries: [
    ["USERS", { status: "completed" as const, count: 3000, total: 3000 }],
    ["ORDERS", { status: "running" as const, count: 2000, total: 5000 }],
  ] as [string, any][],
  runFailCount: 0,
  runSessionId: "session-123",
  runStartedAt: Date.now() - 60000,
  runSuccessCount: 1,
  runTotalTables: 3,
  tr: (en: string, _ko: string) => en,
  validation: {},
  warnings: [],
  wsStatusText: "Connected",
  zipFileId: "",
};

describe("RunStatusPanel", () => {
  it("renders migration monitor header", () => {
    render(<RunStatusPanel {...defaultProps} />);
    expect(screen.getByText("Migration Monitor")).toBeInTheDocument();
    expect(screen.getByText("33%")).toBeInTheDocument();
    expect(screen.getByText("5,000")).toBeInTheDocument(); // processed rows
    expect(screen.getByText("83")).toBeInTheDocument(); // rows per second value
  });

  it("renders table list entries", () => {
    render(<RunStatusPanel {...defaultProps} />);
    expect(screen.getByTitle("USERS")).toBeInTheDocument();
    expect(screen.getByTitle("ORDERS")).toBeInTheDocument();
    expect(screen.getByText("3,000 / 3,000")).toBeInTheDocument();
  });

  it("shows dry-run badge when active", () => {
    render(<RunStatusPanel {...defaultProps} runDryRun={true} />);
    expect(screen.getByText("DRY-RUN")).toBeInTheDocument();
  });

  it("renders circular progress SVG", () => {
    const { container } = render(<RunStatusPanel {...defaultProps} />);
    const svg = container.querySelector("svg");
    expect(svg).toBeInTheDocument();
    const circles = container.querySelectorAll("circle");
    expect(circles.length).toBeGreaterThan(0);
  });

  it("renders CPU and Memory metrics", () => {
    render(<RunStatusPanel {...defaultProps} />);
    expect(screen.getByText("15%")).toBeInTheDocument();
    expect(screen.getByText("256MB")).toBeInTheDocument();
  });

  it("calls onResetRunState when Close button is clicked", () => {
    render(<RunStatusPanel {...defaultProps} migrationBusy={false} />);
    const closeBtn = screen.getByRole("button", { name: "Close Monitoring" });
    fireEvent.click(closeBtn);
    expect(defaultProps.onResetRunState).toHaveBeenCalled();
  });
});
