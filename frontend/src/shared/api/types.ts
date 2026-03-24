export type RuntimeMeta = {
  authEnabled: boolean;
  uiVersion: string;
  googleClientId?: string;
  features?: {
    objectGroupMode?: boolean;
    tableHistory?: boolean;
    precheckRowCount?: boolean;
  };
};

export type PrecheckTableResult = {
  table_name: string;
  source_row_count: number;
  target_row_count: number;
  diff: number;
  decision: "transfer_required" | "skip_candidate" | "count_check_failed";
  policy?: string;
  reason?: string;
  transfer_planned?: boolean;
  checked_at?: string;
};

export type PrecheckSummary = {
  total_tables: number;
  transfer_required_count: number;
  skip_candidate_count: number;
  count_check_failed_count: number;
};

export type PrecheckResponse = {
  summary: PrecheckSummary;
  items: PrecheckTableResult[];
  error?: string;
};

export type AuthUser = {
  userId: number;
  username: string;
};

export type Credential = {
  id: number;
  userId: number;
  alias: string;
  dbType: string;
  host: string;
  port?: number;
  databaseName?: string;
  username?: string;
  password?: string;
  createdAt: string;
  updatedAt: string;
};

export type HistoryEntry = {
  id: number;
  userId: number;
  status: string;
  sourceSummary: string;
  targetSummary: string;
  optionsJson: string;
  logSummary?: string;
  createdAt: string;
};

export type HistoryListResponse = {
  items: HistoryEntry[];
  page: number;
  pageSize: number;
  total: number;
};
