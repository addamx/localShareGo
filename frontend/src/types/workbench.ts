export type NoticeKind = "success" | "info" | "warning" | "error";

export interface Notice {
  kind: NoticeKind;
  message: string;
}

export interface AppBootstrap {
  appName: string;
  paths: { databasePath: string };
  runtimeConfig: { maxTextBytes: number; webRoute: string };
  services: {
    clipboard: { pollIntervalMs: number };
    httpServer: {
      bindHost: string;
      preferredPort: number;
      effectivePort: number | null;
      state: string;
      lastError: string | null;
      webBasePath: string;
    };
    network: { deviceName: string; accessHost: string; accessHosts: string[] };
    session: {
      accessUrl: string;
      expiresAt: number;
      tokenQueryKey: string;
    };
  };
}

export interface ClipboardListQuery {
  search?: string | null;
  pinnedOnly: boolean;
  includeDeleted?: boolean;
  limit: number;
}

export interface ClipboardItemRecord {
  id: string;
  content: string;
  preview: string;
  charCount: number;
  sourceKind: string;
  sourceDeviceId: string | null;
  pinned: boolean;
  isCurrent: boolean;
  createdAt: number;
  updatedAt: number;
}

export interface ClipboardItemSummary {
  id: string;
  preview: string;
  charCount: number;
  sourceKind: string;
  sourceDeviceId: string | null;
  pinned: boolean;
  isCurrent: boolean;
  createdAt: number;
  updatedAt: number;
}

export interface ClipboardListResponse {
  items: ClipboardItemSummary[];
}

export interface ClipboardWriteResponse {
  item: ClipboardItemRecord;
  created: boolean;
  reusedExisting: boolean;
}

export interface HealthResponse {
  status: string;
}

export interface SessionResponse {
  deviceName: string;
  accessUrl: string;
  expiresAt: number;
  maxTextBytes: number;
}

export interface OnlineDevice {
  id: string;
  name: string;
  kind: string;
}

export interface DevicePresenceResponse {
  self: OnlineDevice;
  devices: OnlineDevice[];
}

export interface SyncClipboardRequest {
  content: string;
  sourceDeviceId: string;
  targetDeviceIds: string[];
  syncAll: boolean;
}

export interface SyncClipboardResponse {
  deliveredDevices: OnlineDevice[];
}

export interface DesktopSettings {
  showAppHotkey: string;
}

export interface ClipboardRefreshEvent {
  itemId: string;
}

export interface SyncClipboardEvent {
  targetDeviceIds: string[];
  content: string;
  sourceKind: string;
  createdAt: number;
}

export interface ServerEvent {
  kind: string;
  scope: string;
  itemId: string | null;
  sync: SyncClipboardEvent | null;
  ts: number;
}

export interface ConnectivityCheck {
  host: string;
  url: string;
  tcpOk: boolean;
  httpOk: boolean;
  httpStatusLine: string | null;
  error: string | null;
}

export interface ConnectivityReport {
  serverState: string;
  effectivePort: number;
  accessUrl: string;
  checks: ConnectivityCheck[];
}

export interface ApiErrorPayload {
  code: string;
  message: string;
}

export interface ApiEnvelope<T> {
  ok: boolean;
  data: T | null;
  error: ApiErrorPayload | null;
}
