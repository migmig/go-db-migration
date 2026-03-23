import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { MigrationOptionsPanel } from "./MigrationOptionsPanel";
import { MigrationOptions } from "../types";

const mockOptions: MigrationOptions = {
  objectGroup: "all",
  outFile: "output.sql",
  perTable: false,
  withDdl: true,
  withSequences: true,
  withIndexes: true,
  withConstraints: true,
  validate: true,
  truncate: false,
  upsert: false,
  oracleOwner: "",
  batchSize: 1000,
  workers: 4,
  copyBatch: 1000,
  dbMaxOpen: 10,
  dbMaxIdle: 5,
  dbMaxLife: 3600,
  logJson: false,
  dryRun: false,
};

const defaultProps = {
  tr: (en: string, _ko: string) => en,
  meta: { authEnabled: false, uiVersion: "v1" },
  targetMode: "direct" as const,
  objectGroupModeEnabled: true,
  effectiveObjectGroup: "all" as const,
  options: mockOptions,
  setOptions: vi.fn(),
  onApplyObjectGroupSelection: vi.fn(),
  precheckPolicy: "strict",
  setPrecheckPolicy: vi.fn(),
  precheckBusy: false,
  migrationBusy: false,
  selectedTablesCount: 5,
  onRunPrecheck: vi.fn(),
  precheckError: "",
  precheckSummary: null,
  precheckDecisionFilter: "all" as const,
  setPrecheckDecisionFilter: vi.fn(),
  precheckItems: [],
  onStartMigration: vi.fn(),
  migrationError: "",
};

describe("MigrationOptionsPanel", () => {
  it("renders basic migration options", () => {
    render(<MigrationOptionsPanel {...defaultProps} />);
    expect(screen.getByText("Migration Options")).toBeInTheDocument();
    expect(screen.getByLabelText("Include CREATE TABLE DDL")).toBeInTheDocument();
    expect(screen.getByLabelText("Include sequences")).toBeInTheDocument();
    expect(screen.getByLabelText("Include indexes")).toBeInTheDocument();
    expect(screen.getByLabelText("Include constraints")).toBeInTheDocument();
  });

  it("toggles advanced settings accordion", () => {
    render(<MigrationOptionsPanel {...defaultProps} />);
    const advancedBtn = screen.getByText("Advanced Settings");
    fireEvent.click(advancedBtn);
    expect(screen.getByLabelText("Workers")).toBeInTheDocument(); 
  });

  it("calls onApplyObjectGroupSelection when target select changes", () => {
    render(<MigrationOptionsPanel {...defaultProps} />);
    const select = screen.getByLabelText("Migration target");
    fireEvent.change(select, { target: { value: "tables" } });
    expect(defaultProps.onApplyObjectGroupSelection).toHaveBeenCalledWith("tables");
  });

  it("calls onRunPrecheck when button is clicked", () => {
    render(<MigrationOptionsPanel {...defaultProps} />);
    const precheckBtn = screen.getByRole("button", { name: "Run Pre-check" });
    fireEvent.click(precheckBtn);
    expect(defaultProps.onRunPrecheck).toHaveBeenCalled();
  });

  it("calls onStartMigration when button is clicked", () => {
    render(<MigrationOptionsPanel {...defaultProps} />);
    const startBtn = screen.getByRole("button", { name: "Start Migration" });
    fireEvent.click(startBtn);
    expect(defaultProps.onStartMigration).toHaveBeenCalled();
  });

  it("shows pre-check summary when provided", () => {
    const props = {
      ...defaultProps,
      precheckSummary: {
        total_tables: 10,
        transfer_required_count: 7,
        skip_candidate_count: 2,
        count_check_failed_count: 1,
      }
    };
    render(<MigrationOptionsPanel {...props} />);
    expect(screen.getByText("Total")).toBeInTheDocument();
    expect(screen.getByText("10")).toBeInTheDocument();
    expect(screen.getAllByText("Transfer Required").length).toBeGreaterThan(0);
    expect(screen.getByText("7")).toBeInTheDocument();
  });

  it("disables start button when migration is busy", () => {
    const props = {
      ...defaultProps,
      migrationBusy: true,
    };
    render(<MigrationOptionsPanel {...props} />);
    const startBtn = screen.getByRole("button", { name: "Migration running..." });
    expect(startBtn).toBeDisabled();
  });
});
