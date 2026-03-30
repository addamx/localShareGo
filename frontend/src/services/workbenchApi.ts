import type {
  ApiEnvelope,
  ClipboardItemRecord,
  ClipboardListQuery,
  ClipboardListResponse,
  ClipboardWriteResponse,
  HealthResponse,
  SessionResponse,
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
    const response = await fetch(withToken(path), {
      ...init,
      headers: {
        Accept: "application/json",
        ...(init.body ? { "Content-Type": "application/json" } : {}),
      },
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
    activateClipboardItem: (itemId: string) =>
      request<ClipboardItemRecord>(
        `/api/v1/clipboard-items/${encodeURIComponent(itemId)}/activate`,
        { method: "POST" },
      ),
    eventsUrl: () => withToken("/api/v1/events").toString(),
  };
}

export type WorkbenchApiClient = ReturnType<typeof createWorkbenchApiClient>;
