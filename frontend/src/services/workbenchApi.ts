import type {
  ApiEnvelope,
  ClipboardItemRecord,
  ClipboardListQuery,
  ClipboardListResponse,
  ClipboardWriteResponse,
  DevicePresenceResponse,
  EntryActivationResponse,
  HealthResponse,
  PairRequestStatus,
  PairRequestSummary,
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

async function requestEnvelope<T>(input: URL | string, init: RequestInit = {}) {
  const headers = new Headers(init.headers);
  headers.set("Accept", "application/json");
  if (init.body && !(init.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }

  const response = await fetch(input, {
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
}

export function createWorkbenchApiClient(origin: string, token: string) {
  const withToken = (path: string) => {
    const url = new URL(path, origin);
    url.searchParams.set("token", token);
    return url;
  };

  const request = <T,>(path: string, init: RequestInit = {}) => requestEnvelope<T>(withToken(path), init);

  return {
    health: () => request<HealthResponse>("/api/v1/health"),
    session: () => request<SessionResponse>("/api/v1/session"),
    renewSession: () => request<SessionResponse>("/api/v1/session/renew", { method: "POST" }),
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

export const anonymousWorkbenchApi = {
  activateEntry(origin: string, token: string, deviceId: string, deviceName: string) {
    return requestEnvelope<EntryActivationResponse>(new URL("/api/v1/session/activate-entry", origin), {
      method: "POST",
      body: JSON.stringify({ token, deviceId, deviceName }),
    });
  },
  createPairRequest(origin: string, deviceId: string, deviceName: string) {
    return requestEnvelope<{ request: PairRequestSummary }>(new URL("/api/v1/pair-requests", origin), {
      method: "POST",
      body: JSON.stringify({ deviceId, deviceName }),
    });
  },
  getPairRequest(origin: string, requestId: string) {
    return requestEnvelope<{ request: PairRequestStatus }>(
      new URL(`/api/v1/pair-requests/${encodeURIComponent(requestId)}`, origin),
    );
  },
};

export type WorkbenchApiClient = ReturnType<typeof createWorkbenchApiClient>;
