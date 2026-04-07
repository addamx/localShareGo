export type NoticeKind = "success" | "info" | "warning" | "error";

export interface Notice {
  kind: NoticeKind;
  message: string;
}

export interface AppBootstrap {
  appName: string;
  paths: {
    databasePath: string;
    fileStagingDir: string;
    desktopReceiveDir: string;
  };
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

export type ClipboardItemKind = "text" | "file";
export type ClipboardTransferState = "metadata_only" | "receiving" | "received" | "failed";

export interface ClipboardFileMeta {
  fileName: string;
  extension: string;
  mimeType: string;
  sizeBytes: number;
  thumbnailDataUrl: string | null;
  transferState: ClipboardTransferState;
  progressPercent: number;
  localPath: string | null;
  downloadedAt: number | null;
}

export interface ClipboardItemRecord {
  itemKind: ClipboardItemKind;
  id: string;
  content: string;
  contentType: string;
  hash: string;
  preview: string;
  charCount: number;
  fileMeta: ClipboardFileMeta | null;
  sourceKind: string;
  sourceDeviceId: string | null;
  pinned: boolean;
  isCurrent: boolean;
  createdAt: number;
  updatedAt: number;
}

export interface ClipboardItemSummary {
  itemKind: ClipboardItemKind;
  id: string;
  preview: string;
  charCount: number;
  contentType: string;
  fileMeta: ClipboardFileMeta | null;
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
  selfDeviceId: string;
  accessUrl: string;
  sessionId: string;
  sessionKind: string;
  sessionStatus: string;
  expiresAt: number;
  tokenTtlMinutes: number;
  rotationEnabled: boolean;
  maxTextBytes: number;
}

export interface EntryActivationResponse {
  session: SessionResponse;
  credential: string;
}

export interface OnlineDevice {
  id: string;
  name: string;
  kind: string;
}

export interface LinkedWebDevice {
  id: string;
  name: string;
  lastKnownIp: string;
  lastSeenAt: number;
  online: boolean;
  expiresAt: number;
}

export interface PairRequestSummary {
  id: string;
  deviceId: string;
  deviceName: string;
  status: string;
  createdAt: number;
  updatedAt: number;
  expiresAt: number;
}

export interface PairRequestStatus {
  id: string;
  deviceId: string;
  deviceName: string;
  status: string;
  accessUrl: string;
  credential: string;
  createdAt: number;
  updatedAt: number;
  expiresAt: number;
}

export interface DevicePresenceResponse {
  self: OnlineDevice;
  devices: OnlineDevice[];
}

export interface SyncClipboardRequest {
  itemId?: string | null;
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
  item: ClipboardItemRecord;
  createdAt: number;
}

export interface FileTransferEvent {
  itemId: string;
  status: ClipboardTransferState;
  progressPercent: number;
  bytesTransferred: number;
  bytesTotal: number;
  errorMessage: string | null;
}

export interface RevokedEvent {
  deviceId: string;
  sessionId: string;
  reason: string;
}

export interface ServerEvent {
  kind: string;
  scope: string;
  itemId: string | null;
  sync: SyncClipboardEvent | null;
  fileTransfer: FileTransferEvent | null;
  revoked: RevokedEvent | null;
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
