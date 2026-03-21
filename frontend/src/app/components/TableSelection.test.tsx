import { render, screen, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi } from "vitest";
import { TableSelection } from "./TableSelection";

const defaultProps = {
  tr: (en: string, _ko: string) => en,
  allTables: ["USERS", "ORDERS", "PRODUCTS", "DEPT"],
  selectedTables: ["USERS"],
  setSelectedTables: vi.fn(),
  objectGroupModeEnabled: true,
  previewTables: ["USERS"],
  discoverySummary: null,
  previewObjectGroup: "all",
  previewSequences: [],
  compareEntries: [],
  migrationBusy: false,
  selectByCategory: vi.fn(),
  tableProgress: {},
  historyByTable: {},
  history: [],
  openTableHistory: vi.fn(),
  activeTableHistory: null,
  setActiveTableHistory: vi.fn(),
  tableHistoryBusy: false,
  tableHistoryError: null,
  replayHistory: vi.fn(),
};

describe("TableSelection", () => {
  it("renders available and selected tables", () => {
    render(<TableSelection {...defaultProps} />);
    expect(screen.getByText("Available Tables (3)")).toBeInTheDocument();
    expect(screen.getByText("Selected Tables (1)")).toBeInTheDocument();
    expect(screen.getByText("ORDERS")).toBeInTheDocument();
    // USERS is in selected list
    const selectedList = screen.getByText("Selected Tables (1)").closest('div')?.parentElement;
    expect(selectedList?.textContent).toContain("USERS");
  });

  it("filters available tables by search", () => {
    render(<TableSelection {...defaultProps} />);
    const searchInput = screen.getByPlaceholderText("Search...");
    fireEvent.change(searchInput, { target: { value: "ORD" } });
    
    expect(screen.getByText("ORDERS")).toBeInTheDocument();
    expect(screen.queryByText("PRODUCTS")).not.toBeInTheDocument();
  });

  it("calls setSelectedTables when moving all right", () => {
    render(<TableSelection {...defaultProps} />);
    const moveAllBtn = screen.getByTitle("Add all");
    fireEvent.click(moveAllBtn);
    
    expect(defaultProps.setSelectedTables).toHaveBeenCalled();
  });

  it("selects a table when clicked in the list", () => {
    render(<TableSelection {...defaultProps} />);
    const ordersRow = screen.getByText("ORDERS").closest("tr");
    fireEvent.click(ordersRow!);
    
    // Checked state is internal, but we can verify it by clicking 'Add selected'
    const addSelectedBtn = screen.getByTitle("Add selected");
    fireEvent.click(addSelectedBtn);
    expect(defaultProps.setSelectedTables).toHaveBeenCalled();
  });
});
