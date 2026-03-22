import { HistoryEntry } from "../shared/api/types";

export type RoleFilter = "all" | "source" | "target";
export type NoticeTone = "info" | "error";
export type PrecheckDecisionFilter = "all" | "transfer_required" | "skip_candidate" | "count_check_failed";
export type WsStatus = "idle" | "connecting" | "connected" | "closed" | "error";
export type TableRunStatus = "pending" | "running" | "completed" | "error";
export type TableHistoryStatusFilter = "all" | "not_started" | "success" | "failed";
export type TableSortOption = "table_asc" | "table_desc" | "recent_desc" | "runs_desc" | "history_status";
export type ObjectGroup = "all" | "tables" | "sequences";

export type SourceState = {
  oracleUrl: string;
  username: string;
  password: string;
  like: string;
  saveCredential?: boolean;
  alias?: string;
};

export type TargetState = {
  mode: "file" | "direct";
  targetUrl: string;
  schema: string;
  saveCredential?: boolean;
  alias?: string;
};

export type TargetTableEntry = {
  name: string;
  inSource: boolean;
  inTarget: boolean;
  category: "source_only" | "both" | "target_only";
  sourceRowCount: number | null;
  targetRowCount: number | null;
};

export type CompareState = {
  targetTables: string[];
  fetchedAt: string | null;
  busy: boolean;
  error: string | null;
};

export type CompareFilter = "all" | "source_only" | "both" | "target_only";

export type SourceRecent = {
  oracleUrl: string;
  username: string;
  password: string;
};

export type MigrationOptions = {
  objectGroup: ObjectGroup;
  outFile: string;
  perTable: boolean;
  withDdl: boolean;
  withSequences: boolean;
  withIndexes: boolean;
  withConstraints: boolean;
  validate: boolean;
  truncate: boolean;
  upsert: boolean;
  oracleOwner: string;
  batchSize: number;
  workers: number;
  copyBatch: number;
  dbMaxOpen: number;
  dbMaxIdle: number;
  dbMaxLife: number;
  logJson: boolean;
  dryRun: boolean;
};

export type TableRunState = {
  total: number;
  count: number;
  status: TableRunStatus;
  error?: string;
  details?: string;
};

export type TableHistoryState = {
  status: "success" | "failed";
  runCount: number;
  lastRunAt: string;
};

export type TableHistoryDetail = {
  tableName: string;
  entries: HistoryEntry[];
};

export type ValidationState = {
  sourceCount: number;
  targetCount: number;
  status: string;
  message: string;
};

export type DdlEvent = {
  key: string;
  object: string;
  name: string;
  status: string;
  error?: string;
};

export type DiscoverySummary = {
  objectGroup: ObjectGroup;
  tables: string[];
  sequences: string[];
};

export type ReportSummary = {
  total_rows: number;
  success_count: number;
  error_count: number;
  duration: string;
  report_id: string;
  object_group: ObjectGroup;
  stats: GroupedStats;
};

export type GroupStats = {
  total_items: number;
  success_count: number;
  error_count: number;
  skipped_count: number;
  total_rows?: number;
};

export type GroupedStats = {
  tables: GroupStats;
  sequences: GroupStats;
};

export type WsProgressMsg = {
  type: string;
  table?: string;
  count?: number;
  total?: number;
  error?: string;
  message?: string;
  zip_file_id?: string;
  connection_ok?: boolean;
  object?: string;
  object_name?: string;
  status?: string;
  object_group?: ObjectGroup;
  tables?: string[];
  sequences?: string[];
  phase?: string;
  category?: string;
  suggestion?: string;
  recoverable?: boolean;
  batch_num?: number;
  row_offset?: number;
  report_summary?: ReportSummary;
};

export type MetricsState = {
  cpu: string;
  mem: string;
};

export type Locale = "en" | "ko";
