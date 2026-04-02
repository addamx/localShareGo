import type {
  ApiEnvelope,
  ClipboardItemRecord,
  ClipboardListQuery,
  ClipboardListResponse,
  ClipboardWriteResponse,
  DevicePresenceResponse,
  HealthResponse,
  SessionResponse,
  SyncClipboardRequest,
  SyncClipboardResponse,
} from "../types/workbench";

export class WorkbenchApiError extends Error {
  code: string;
  status: number;

  constructor(code: string, message: string, status: number) {
    super(message);
    this.name = "WorkbenchApiError";
    this.code = code;
    this.status = status;
  }
}

export function createWorkbenchApiClient(origin: string, token: string) {
  const withToken = (path: string) => {
    const url = new URL(path, origin);
    url.searchParams.set("token", token);
    return url;
  };

  const request = async <T,>(path: string, init: RequestInit = {}): Promise<T> => {
    const headers = new Headers(init.headers);
    headers.set("Accept", "application/json");
    if (init.body && !(init.body instanceof FormData) && !headers.has("Content-Type")) {
      headers.set("Content-Type", "application/json");
    }
    const response = await fetch(withToken(path), {
      ...init,
      headers,
    });
    const payload = (await response.json()) as ApiEnvelope<T>;
    if (!response.ok || !payload.ok || !payload.data) {
      throw new WorkbenchApiError(
        payload.error?.code ?? "HTTP_ERROR",
        payload.error?.message ?? `HTTP ${response.status}`,
        response.status,
      );
    }
    return payload.data;
  };

  return {
    health: () => request<HealthResponse>("/api/v1/health"),
    session: () => request<SessionResponse>("/api/v1/session"),
    registerWebDevice: (name: string, deviceId: string) =>
      request<DevicePresenceResponse>("/api/v1/devices/register", {
        method: "POST",
        body: JSON.stringify({ deviceId, name }),
      }),
    heartbeatWebDevice: (deviceId: string) =>
      request<DevicePresenceResponse>("/api/v1/devices/heartbeat", {
        method: "POST",
        body: JSON.stringify({ deviceId }),
      }),
    listClipboardItems: (query: ClipboardListQuery) => {
      const url = withToken("/api/v1/clipboard-items");
      if (query.search?.trim()) {
        url.searchParams.set("search", query.search.trim());
      }
      if (query.pinnedOnly) {
        url.searchParams.set("pinnedOnly", "true");
      }
      url.searchParams.set("limit", String(query.limit));
      return request<ClipboardListResponse>(url.pathname + url.search);
    },
    getClipboardItem: (itemId: string) =>
      request<ClipboardItemRecord>(
        `/api/v1/clipboard-items/${encodeURIComponent(itemId)}`,
      ),
    submitClipboardItem: (content: string) =>
      request<ClipboardWriteResponse>("/api/v1/clipboard-items", {
        method: "POST",
        body: JSON.stringify({ content, pinned: false, activate: false }),
      }),
    createFileItem: (file: File) => {
      const body = new FormData();
      body.set("file", file);
      return request<ClipboardWriteResponse>("/api/v1/file-items", {
        method: "POST",
        body,
      });
    },
    activateClipboardItem: (itemId: string) =>
      request<ClipboardItemRecord>(
        `/api/v1/clipboard-items/${encodeURIComponent(itemId)}/activate`,
        { method: "POST" },
      ),
    syncClipboard: (payload: SyncClipboardRequest) =>
      request<SyncClipboardResponse>("/api/v1/clipboard-sync", {
        method: "POST",
        body: JSON.stringify(payload),
      }),
    receiveFileItem: (itemId: string) =>
      request<ClipboardItemRecord>(`/api/v1/file-items/${encodeURIComponent(itemId)}/receive`, {
        method: "POST",
      }),
    fileContentUrl: (itemId: string) =>
      withToken(`/api/v1/file-items/${encodeURIComponent(itemId)}/content`).toString(),
    eventsUrl: () => withToken("/api/v1/events").toString(),
  };
}

export type WorkbenchApiClient = ReturnType<typeof createWorkbenchApiClient>;
