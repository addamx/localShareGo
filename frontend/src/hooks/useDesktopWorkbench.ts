import { computed, h, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { useDebounceFn } from "@vueuse/core";
import { NIcon, useMessage, type DropdownMixedOption, type DropdownOption } from "naive-ui";

import { copyText } from "../app/env";
import { describeError, uniqueStrings } from "../app/formatters";
import { desktopWorkbench } from "../services/desktopWorkbench";
import type {
  AppBootstrap,
  ClipboardItemRecord,
  ConnectivityReport,
  OnlineDevice,
} from "../types/workbench";
import { GlobeIcon, SettingsIcon } from "../utils/desktopIcons";

const preferredHostStorageKey = "localsharego:web-entry-host";

export function useDesktopWorkbench() {
  const router = useRouter();
  const message = useMessage();
  const available = desktopWorkbench.isAvailable();

  const bootstrap = ref<AppBootstrap | null>(null);
  const items = ref<ClipboardItemRecord[]>([]);
  const selectedId = ref<string | null>(null);
  const detailItemId = ref<string | null>(null);
  const loading = ref(true);
  const refreshing = ref(false);
  const search = ref("");
  const pinnedOnly = ref(false);
  const webPanelOpen = ref(false);
  const detailPanelOpen = ref(false);
  const diagnosticsModalOpen = ref(false);
  const diagnosticsLoading = ref(false);
  const connectivity = ref<ConnectivityReport | null>(null);
  const selectedHost = ref("");
  const clock = ref(Date.now());
  const contextMenuOpen = ref(false);
  const contextMenuX = ref(0);
  const contextMenuY = ref(0);
  const contextMenuItem = ref<ClipboardItemRecord | null>(null);
  const onlineDevices = ref<OnlineDevice[]>([]);

  const detailItem = computed(() => items.value.find((item) => item.id === detailItemId.value) ?? null);
  const sessionPort = computed(() => {
    const server = bootstrap.value?.services.httpServer;
    return server ? (server.effectivePort ?? server.preferredPort) : 0;
  });
  const bindAddress = computed(() => {
    const server = bootstrap.value?.services.httpServer;
    if (!server) {
      return "--";
    }
    return `${server.bindHost}:${sessionPort.value}`;
  });
  const baseAccessUrl = computed(() => bootstrap.value?.services.session.accessUrl ?? "");
  const candidateHosts = computed(() => {
    if (!bootstrap.value || !baseAccessUrl.value) {
      return [];
    }

    const baseHost = new URL(baseAccessUrl.value).hostname;
    return uniqueStrings([baseHost, ...bootstrap.value.services.network.accessHosts]);
  });
  const activeHost = computed(() => {
    if (selectedHost.value) {
      return selectedHost.value;
    }
    return candidateHosts.value[0] ?? "";
  });
  const resolvedSessionUrl = computed(() => {
    if (!baseAccessUrl.value) {
      return "";
    }

    const url = new URL(baseAccessUrl.value);
    if (activeHost.value) {
      url.hostname = activeHost.value;
    }
    if (sessionPort.value > 0) {
      url.port = String(sessionPort.value);
    }
    return url.toString();
  });
  const tokenCountdown = computed(() =>
    formatTokenCountdown(bootstrap.value?.services.session.expiresAt ?? null, clock.value),
  );
  const candidateOptions = computed<DropdownOption[]>(() =>
    candidateHosts.value.map((host) => ({
      label: host === activeHost.value ? `${host} current` : host,
      key: host,
    })),
  );
  const moreOptions = computed<DropdownMixedOption[]>(() => [
    { label: "Web端", key: "web", renderIcon: renderOptionIcon(GlobeIcon) },
    { label: "设置", key: "settings", renderIcon: renderOptionIcon(SettingsIcon) },
    { label: "刷新", key: "refresh" },
    { label: "清空", key: "clear" },
    { label: pinnedOnly.value ? "查看全部" : "仅看置顶", key: "togglePinned" },
  ]);
  const contextMenuOptions = computed<DropdownOption[]>(() => {
    const item = contextMenuItem.value;
    if (!item) {
      return [];
    }

    const syncChildren: DropdownOption[] =
      onlineDevices.value.length === 0
        ? [{ label: "No online devices", key: "sync:none", disabled: true }]
        : [
            { label: "Sync to all", key: "sync:all" },
            ...onlineDevices.value.map((device) => ({
              label: device.name,
              key: `sync:${device.id}`,
            })),
          ];

    return [
      { label: item.pinned ? "Unpin" : "Pin", key: "pin" },
      { label: "Sync", key: "sync", children: syncChildren },
      { label: "View", key: "view" },
      { label: "Delete", key: "delete" },
    ];
  });

  const refreshBySearch = useDebounceFn(() => {
    void refreshHistory(true);
  }, 220);

  watch(search, () => {
    if (available) {
      refreshBySearch();
    }
  });

  let cleanup: (() => void) | null = null;
  let cleanupSession: (() => void) | null = null;
  let clockTimer: number | null = null;

  onMounted(async () => {
    window.addEventListener("keydown", handleWindowKeyDown);
    window.addEventListener("blur", handleWindowBlur);

    if (!available) {
      loading.value = false;
      return;
    }

    cleanup = desktopWorkbench.subscribeClipboardRefresh(() => {
      void refreshHistory(true);
    });
    cleanupSession = desktopWorkbench.subscribeSessionRefresh(() => {
      void loadBootstrap();
    });
    clockTimer = window.setInterval(() => {
      clock.value = Date.now();
    }, 1000);

    try {
      await loadBootstrap();
      await Promise.all([refreshHistory(true), refreshOnlineDevices()]);
    } finally {
      loading.value = false;
    }
  });

  onBeforeUnmount(() => {
    cleanup?.();
    cleanupSession?.();
    if (clockTimer !== null) {
      window.clearInterval(clockTimer);
    }
    window.removeEventListener("keydown", handleWindowKeyDown);
    window.removeEventListener("blur", handleWindowBlur);
  });

  async function loadBootstrap() {
    try {
      const nextBootstrap = await desktopWorkbench.getBootstrapContext();
      bootstrap.value = nextBootstrap;
      syncSelectedHost(nextBootstrap);
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function refreshHistory(silent: boolean) {
    if (!available) {
      return;
    }

    refreshing.value = true;
    closeContextMenu();

    try {
      items.value = await desktopWorkbench.listClipboardItems({
        search: search.value.trim() || null,
        pinnedOnly: pinnedOnly.value,
        includeDeleted: false,
        limit: 80,
      });

      const currentItemId = items.value.find((item) => item.isCurrent)?.id ?? items.value[0]?.id ?? null;
      if (!selectedId.value || !items.value.some((item) => item.id === selectedId.value)) {
        selectedId.value = currentItemId;
      }

      if (detailItemId.value && !items.value.some((item) => item.id === detailItemId.value)) {
        detailItemId.value = null;
        detailPanelOpen.value = false;
      }

      if (!silent) {
        message.success("Clipboard refreshed");
      }
    } catch (error) {
      message.error(describeError(error));
    } finally {
      refreshing.value = false;
    }
  }

  async function refreshOnlineDevices() {
    if (!available) {
      return;
    }

    try {
      onlineDevices.value = await desktopWorkbench.listOnlineDevices();
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handleRowClick(item: ClipboardItemRecord) {
    selectedId.value = item.id;
    await handleActivate(item.id, false);
  }

  async function openWebPanel() {
    await loadBootstrap();
    webPanelOpen.value = true;
  }

  async function hideDesktopApp() {
    if (!available) {
      return;
    }

    try {
      await desktopWorkbench.hideDesktopApp();
    } catch (error) {
      message.error(describeError(error));
    }
  }

  function handleWindowKeyDown(event: KeyboardEvent) {
    if (event.key === "Escape") {
      closeContextMenu();
      void hideDesktopApp();
    }
  }

  function handleWindowBlur() {
    closeContextMenu();
    void hideDesktopApp();
  }

  function openContextMenu(event: MouseEvent, item: ClipboardItemRecord) {
    void refreshOnlineDevices();
    selectedId.value = item.id;
    contextMenuItem.value = item;
    contextMenuX.value = event.clientX;
    contextMenuY.value = event.clientY;
    contextMenuOpen.value = true;
  }

  function closeContextMenu() {
    contextMenuOpen.value = false;
    contextMenuItem.value = null;
  }

  async function handleContextSelect(key: string | number) {
    const item = contextMenuItem.value;
    closeContextMenu();
    if (!item) {
      return;
    }

    if (key === "pin") {
      await handlePin(item);
      return;
    }
    if (key === "delete") {
      await handleDelete(item);
      return;
    }
    if (key === "view") {
      detailItemId.value = item.id;
      detailPanelOpen.value = true;
      return;
    }
    if (key === "sync:all") {
      await handleSync(item, [], true);
      return;
    }

    const textKey = String(key);
    if (textKey.startsWith("sync:")) {
      const targetDeviceID = textKey.slice(5);
      if (targetDeviceID) {
        await handleSync(item, [targetDeviceID], false);
      }
    }
  }

  async function handleActivate(itemId: string, notify: boolean) {
    try {
      await desktopWorkbench.activateClipboardItem(itemId);
      if (notify) {
        message.success("Copied to system clipboard");
      }
      await refreshHistory(true);
      await hideDesktopApp();
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handlePin(item: ClipboardItemRecord) {
    try {
      await desktopWorkbench.updateClipboardItemPin(item.id, !item.pinned);
      message.success(item.pinned ? "Unpinned" : "Pinned");
      await refreshHistory(true);
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handleDelete(item: ClipboardItemRecord) {
    if (!window.confirm(`Delete this record?\n\n${item.preview || item.content}`)) {
      return;
    }

    try {
      await desktopWorkbench.deleteClipboardItem(item.id);
      if (detailItemId.value === item.id) {
        detailItemId.value = null;
        detailPanelOpen.value = false;
      }
      message.success("Record deleted");
      await refreshHistory(true);
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handleSync(item: ClipboardItemRecord, targetDeviceIds: string[], syncAll: boolean) {
    try {
      const result = await desktopWorkbench.syncClipboardItem(item.id, targetDeviceIds, syncAll);
      if (result.deliveredDevices.length === 0) {
        message.warning("No available target devices");
        return;
      }
      message.success(`Synced to ${result.deliveredDevices.length} devices`);
      await refreshOnlineDevices();
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handleDeleteDetail() {
    if (!detailItem.value) {
      return;
    }
    await handleDelete(detailItem.value);
  }

  async function handleCopyDetail() {
    if (!detailItem.value) {
      return;
    }
    await copyText(detailItem.value.content);
    message.success("Content copied");
  }

  async function handleCopyEntry() {
    if (!resolvedSessionUrl.value) {
      return;
    }
    await copyText(resolvedSessionUrl.value);
    message.success("Link copied");
  }

  async function handleOpenEntry() {
    if (!resolvedSessionUrl.value) {
      return;
    }

    try {
      await desktopWorkbench.openURL(resolvedSessionUrl.value);
    } catch (error) {
      message.error(describeError(error));
    }
  }

  function handleCandidateSelect(key: string | number) {
    const host = String(key);
    selectedHost.value = host;
    window.localStorage.setItem(preferredHostStorageKey, host);
  }

  async function handleRefreshSession() {
    try {
      await desktopWorkbench.rotateSessionToken();
      await loadBootstrap();
      message.success("Session token rotated");
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function openDiagnostics() {
    diagnosticsModalOpen.value = true;
    await runConnectivityCheck();
  }

  async function runConnectivityCheck() {
    diagnosticsLoading.value = true;
    try {
      connectivity.value = await desktopWorkbench.getConnectivityReport();
    } catch (error) {
      message.error(describeError(error));
    } finally {
      diagnosticsLoading.value = false;
    }
  }

  async function handleMoreSelect(key: string | number) {
    if (key === "web") {
      await openWebPanel();
      return;
    }
    if (key === "settings") {
      await router.push("/desktop/settings");
      return;
    }
    if (key === "refresh") {
      await refreshHistory(false);
      return;
    }
    if (key === "togglePinned") {
      pinnedOnly.value = !pinnedOnly.value;
      await refreshHistory(true);
      return;
    }
    if (key === "clear") {
      await handleClearHistory();
    }
  }

  async function handleClearHistory() {
    if (!window.confirm("Clear all clipboard history?")) {
      return;
    }

    try {
      await desktopWorkbench.clearClipboardHistory();
      detailItemId.value = null;
      detailPanelOpen.value = false;
      message.success("History cleared");
      await refreshHistory(true);
    } catch (error) {
      message.error(describeError(error));
    }
  }

  function syncSelectedHost(appBootstrap: AppBootstrap) {
    const hosts = resolveCandidateHosts(appBootstrap);
    const storedHost = window.localStorage.getItem(preferredHostStorageKey)?.trim() ?? "";
    if (hosts.includes(selectedHost.value)) {
      return;
    }
    if (hosts.includes(storedHost)) {
      selectedHost.value = storedHost;
      return;
    }
    selectedHost.value = hosts[0] ?? "";
  }

  return {
    activeHost,
    available,
    bindAddress,
    bootstrap,
    candidateOptions,
    closeContextMenu,
    connectivity,
    contextMenuOpen,
    contextMenuOptions,
    contextMenuX,
    contextMenuY,
    detailItem,
    detailPanelOpen,
    diagnosticsLoading,
    diagnosticsModalOpen,
    handleCandidateSelect,
    handleContextSelect,
    handleCopyDetail,
    handleCopyEntry,
    handleDeleteDetail,
    handleMoreSelect,
    handleOpenEntry,
    handleRefreshSession,
    handleRowClick,
    items,
    loading,
    moreOptions,
    openContextMenu,
    openDiagnostics,
    refreshing,
    resolvedSessionUrl,
    search,
    selectedId,
    tokenCountdown,
    webPanelOpen,
  };
}

function resolveCandidateHosts(appBootstrap: AppBootstrap) {
  const baseHost = new URL(appBootstrap.services.session.accessUrl).hostname;
  return uniqueStrings([baseHost, ...appBootstrap.services.network.accessHosts]);
}

function formatTokenCountdown(expiresAt: number | null, now: number) {
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

function renderOptionIcon(icon: Parameters<typeof h>[0]) {
  return () =>
    h(
      NIcon,
      null,
      {
        default: () => h(icon),
      },
    );
}
