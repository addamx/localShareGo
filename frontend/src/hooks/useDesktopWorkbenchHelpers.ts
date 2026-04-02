import { h } from "vue";
import { NIcon } from "naive-ui";

import type { AppBootstrap } from "../types/workbench";
import { uniqueStrings } from "../app/formatters";

export function resolveCandidateHosts(appBootstrap: AppBootstrap) {
  const baseHost = new URL(appBootstrap.services.session.accessUrl).hostname;
  return uniqueStrings([baseHost, ...appBootstrap.services.network.accessHosts]);
}

export function resolvePreferredHost(appBootstrap: AppBootstrap, currentHost: string, storedHost: string) {
  const hosts = resolveCandidateHosts(appBootstrap);
  const normalizedStoredHost = storedHost.trim();
  if (hosts.includes(currentHost)) {
    return currentHost;
  }
  if (hosts.includes(normalizedStoredHost)) {
    return normalizedStoredHost;
  }
  return hosts[0] ?? "";
}

export function formatTokenCountdown(expiresAt: number | null, now: number) {
  if (!expiresAt) {
    return "Pending";
  }

  const remaining = expiresAt - now;
  if (remaining <= 0) {
    return "Expired";
  }
  if (remaining > 60_000) {
    return `${Math.max(1, Math.floor(remaining / 60_000))}m`;
  }

  return `${Math.floor(remaining / 1000)}s`;
}

export function renderOptionIcon(icon: Parameters<typeof h>[0]) {
  return () =>
    h(
      NIcon,
      null,
      {
        default: () => h(icon),
      },
    );
}
