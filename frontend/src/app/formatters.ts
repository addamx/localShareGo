export function formatRemaining(expiresAt: number | null) {
  if (!expiresAt) return "未提供";

  const remaining = expiresAt - Date.now();
  if (remaining <= 0) return "已过期";

  const seconds = Math.floor(remaining / 1000);
  const minutes = Math.floor(seconds / 60);

  return minutes > 0
    ? `${minutes}m ${String(seconds % 60).padStart(2, "0")}s`
    : `${seconds}s`;
}

export function formatDateTime(value: number | null | undefined) {
  if (!value) return "--";

  return new Intl.DateTimeFormat("zh-CN", {
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
}

export function formatSource(value: string) {
  if (value === "desktop_local") return "本机";
  if (value === "mobile_web") return "Web";
  return value;
}

export function describeError(error: unknown) {
  if (error instanceof Error) return error.message;
  if (typeof error === "string") return error;
  return String(error);
}

export function uniqueStrings(values: string[]) {
  return [...new Set(values)];
}
