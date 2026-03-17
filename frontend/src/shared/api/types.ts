export type RuntimeMeta = {
  authEnabled: boolean;
  uiVersion: string;
  features?: {
    objectGroupMode?: boolean;
  };
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
