import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useDebounceFn } from "@vueuse/core";
import { useMessage, type DropdownOption } from "naive-ui";

import { copyText } from "../app/env";
import { describeError, uniqueStrings } from "../app/formatters";
import { desktopWorkbench } from "../services/desktopWorkbench";
import type {
  AppBootstrap,
  ClipboardItemRecord,
  ConnectivityReport,
} from "../types/workbench";

const preferredHostStorageKey = "localsharego:web-entry-host";

export function useDesktopWorkbench() {
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
  const menuBarVisible = ref(false);
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

  const detailItem = computed(
    () => items.value.find((item) => item.id === detailItemId.value) ?? null,
  );
  const sessionPort = computed(() => {
    const server = bootstrap.value?.services.httpServer;
    return server ? (server.effectivePort ?? server.preferredPort) : 0;
  });
  const bindAddress = computed(() => {
    const server = bootstrap.value?.services.httpServer;
    if (!server) return "--";
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
    if (selectedHost.value) return selectedHost.value;
    return candidateHosts.value[0] ?? "";
  });
  const resolvedSessionUrl = computed(() => {
    if (!baseAccessUrl.value) return "";

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
      label: host === activeHost.value ? `${host} · current` : host,
      key: host,
    })),
  );
  const moreOptions = computed<DropdownOption[]>(() => [
    { label: "刷新", key: "refresh" },
    { label: "清空", key: "clear" },
    { label: pinnedOnly.value ? "查看全部" : "仅看置顶", key: "togglePinned" },
  ]);
  const contextMenuOptions = computed<DropdownOption[]>(() => {
    const item = contextMenuItem.value;
    if (!item) return [];

    return [
      { label: item.pinned ? "取消置顶" : "置顶", key: "pin" },
      { label: "删除", key: "delete" },
      { label: "查看", key: "view" },
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
  let clockTimer: number | null = null;
  let altToggleArmed = false;

  onMounted(async () => {
    window.addEventListener("keydown", handleWindowKeyDown);
    window.addEventListener("keyup", handleWindowKeyUp);
    window.addEventListener("blur", handleWindowBlur);

    if (!available) {
      loading.value = false;
      return;
    }

    cleanup = desktopWorkbench.subscribeClipboardRefresh(() => {
      void refreshHistory(true);
    });
    clockTimer = window.setInterval(() => {
      clock.value = Date.now();
    }, 1000);

    try {
      await loadBootstrap();
      await refreshHistory(true);
    } finally {
      loading.value = false;
    }
  });

  onBeforeUnmount(() => {
    cleanup?.();
    if (clockTimer !== null) {
      window.clearInterval(clockTimer);
    }
    window.removeEventListener("keydown", handleWindowKeyDown);
    window.removeEventListener("keyup", handleWindowKeyUp);
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
    if (!available) return;

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
        message.success("历史已刷新");
      }
    } catch (error) {
      message.error(describeError(error));
    } finally {
      refreshing.value = false;
    }
  }

  async function handleRowClick(item: ClipboardItemRecord) {
    selectedId.value = item.id;
    await handleActivate(item.id, false);
  }

  function openWebPanel() {
    hideMenuBar();
    webPanelOpen.value = true;
  }

  function hideMenuBar() {
    menuBarVisible.value = false;
    altToggleArmed = false;
  }

  function handleWindowKeyDown(event: KeyboardEvent) {
    if (event.key === "Alt") {
      if (!event.repeat) {
        altToggleArmed = true;
      }
      return;
    }

    altToggleArmed = false;

    if (event.key === "Escape") {
      hideMenuBar();
    }
  }

  function handleWindowKeyUp(event: KeyboardEvent) {
    if (event.key !== "Alt") {
      return;
    }

    if (altToggleArmed) {
      menuBarVisible.value = !menuBarVisible.value;
    }

    altToggleArmed = false;
  }

  function handleWindowBlur() {
    hideMenuBar();
  }

  function openContextMenu(event: MouseEvent, item: ClipboardItemRecord) {
    selectedId.value = item.id;
    contextMenuItem.value = item;
    contextMenuX.value = event.clientX;
    contextMenuY.value = event.clientY;
    contextMenuOpen.value = true;
  }

  function closeContextMenu() {
    contextMenuOpen.value = false;
  }

  async function handleContextSelect(key: string | number) {
    const item = contextMenuItem.value;
    closeContextMenu();
    if (!item) return;

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
    }
  }

  async function handleActivate(itemId: string, notify: boolean) {
    try {
      await desktopWorkbench.activateClipboardItem(itemId);
      if (notify) {
        message.success("已写回系统剪贴板");
      }
      await refreshHistory(true);
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handlePin(item: ClipboardItemRecord) {
    try {
      await desktopWorkbench.updateClipboardItemPin(item.id, !item.pinned);
      message.success(item.pinned ? "已取消置顶" : "已置顶");
      await refreshHistory(true);
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handleDelete(item: ClipboardItemRecord) {
    if (!window.confirm(`确认删除这条历史记录？\n\n${item.preview || item.content}`)) {
      return;
    }

    try {
      await desktopWorkbench.deleteClipboardItem(item.id);
      if (detailItemId.value === item.id) {
        detailItemId.value = null;
        detailPanelOpen.value = false;
      }
      message.success("记录已删除");
      await refreshHistory(true);
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handleDeleteDetail() {
    if (!detailItem.value) return;
    await handleDelete(detailItem.value);
  }

  async function handleCopyDetail() {
    if (!detailItem.value) return;
    await copyText(detailItem.value.content);
    message.success("内容已复制");
  }

  async function handleCopyEntry() {
    if (!resolvedSessionUrl.value) return;
    await copyText(resolvedSessionUrl.value);
    message.success("链接已复制");
  }

  async function handleOpenEntry() {
    if (!resolvedSessionUrl.value) return;

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

  async function handleRotateSession() {
    try {
      await desktopWorkbench.rotateSessionToken();
      await loadBootstrap();
      message.success("令牌已轮换");
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
    if (!window.confirm("确认清空全部剪贴板历史？")) {
      return;
    }

    try {
      await desktopWorkbench.clearClipboardHistory();
      detailItemId.value = null;
      detailPanelOpen.value = false;
      message.success("历史已清空");
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
    handleRotateSession,
    handleRowClick,
    hideMenuBar,
    items,
    loading,
    menuBarVisible,
    moreOptions,
    openContextMenu,
    openDiagnostics,
    openWebPanel,
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
  if (!expiresAt) return "--";

  const remaining = expiresAt - now;
  if (remaining <= 0) return "已过期";

  if (remaining > 60_000) {
    return `${Math.max(1, Math.floor(remaining / 60_000))}m`;
  }

  return `${Math.floor(remaining / 1000)}s`;
}
