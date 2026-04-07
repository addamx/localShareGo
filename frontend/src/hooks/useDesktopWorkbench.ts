import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRouter } from "vue-router";
import { useDebounceFn } from "@vueuse/core";
import { useMessage, type DropdownOption } from "naive-ui";

import { copyText } from "../app/env";
import { describeError } from "../app/formatters";
import { desktopWorkbench } from "../services/desktopWorkbench";
import type {
  AppBootstrap,
  ClipboardItemRecord,
  ConnectivityReport,
  LinkedWebDevice,
  OnlineDevice,
  PairRequestSummary,
} from "../types/workbench";
import { GlobeIcon, PinIcon, PinOffIcon, SettingsIcon } from "../utils/desktopIcons";
import {
  isDesktopFileItem,
  patchDesktopFileTransfer,
} from "./useDesktopFileTransfers";
import {
  formatTokenCountdown,
  resolveCandidateHosts,
  resolvePreferredHost,
  renderOptionIcon,
} from "./useDesktopWorkbenchHelpers";

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
  const desktopPinned = ref(false);
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
  const linkedDevices = ref<LinkedWebDevice[]>([]);
  const pairRequests = ref<PairRequestSummary[]>([]);
  const pairRequestModalOpen = ref(false);

  const activePairRequest = computed(() => pairRequests.value[0] ?? null);

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

    return resolveCandidateHosts(bootstrap.value);
  });
  const activeHost = computed(() => selectedHost.value || candidateHosts.value[0] || "");
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
  const moreOptions = computed<DropdownOption[]>(() => [
    { label: "关联设备", key: "web", renderIcon: renderOptionIcon(GlobeIcon) },
    { label: "设置", key: "settings", renderIcon: renderOptionIcon(SettingsIcon) },
    {
      label: desktopPinned.value ? "取消固定" : "固定窗口",
      key: "toggleDesktopPinned",
      renderIcon: renderOptionIcon(desktopPinned.value ? PinIcon : PinOffIcon),
    },
    { label: "刷新", key: "refresh" },
    { label: "清空", key: "clear" },
  ]);
  const contextMenuOptions = computed<DropdownOption[]>(() => {
    const item = contextMenuItem.value;
    if (!item) {
      return [];
    }

    const syncChildren: DropdownOption[] =
      onlineDevices.value.length === 0
        ? [{ label: "暂无在线设备", key: "sync:none", disabled: true }]
        : [
            { label: "同步到全部", key: "sync:all" },
            ...onlineDevices.value.map((device) => ({
              label: device.name,
              key: `sync:${device.id}`,
            })),
          ];

    return [
      { label: item.pinned ? "取消置顶" : "置顶", key: "pin" },
      { label: "同步", key: "sync", children: syncChildren },
      { label: "查看", key: "view" },
      { label: "删除", key: "delete" },
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

  let cleanupClipboard: (() => void) | null = null;
  let cleanupSession: (() => void) | null = null;
  let cleanupFileTransfer: (() => void) | null = null;
  let cleanupPairRequest: (() => void) | null = null;
  let clockTimer: number | null = null;

  onMounted(async () => {
    window.addEventListener("keydown", handleWindowKeyDown);
    window.addEventListener("blur", handleWindowBlur);

    if (!available) {
      loading.value = false;
      return;
    }

    cleanupClipboard = desktopWorkbench.subscribeClipboardRefresh(() => {
      void refreshHistory(true);
    });
    cleanupSession = desktopWorkbench.subscribeSessionRefresh(() => {
      void loadBootstrap();
      void refreshLinkedDevices();
    });
    cleanupFileTransfer = desktopWorkbench.subscribeFileTransferProgress((event) => {
      patchDesktopFileTransfer(items, event);
      if (event.status === "received" || event.status === "failed") {
        void refreshHistory(true);
      }
    });
    cleanupPairRequest = desktopWorkbench.subscribePairRequest(() => {
      void handleIncomingPairRequest();
    });
    clockTimer = window.setInterval(() => {
      clock.value = Date.now();
    }, 1000);

    try {
      await loadBootstrap();
      await Promise.all([refreshHistory(true), refreshOnlineDevices(), refreshLinkedDevices(), refreshPairRequests(), syncDesktopPinned()]);
    } finally {
      loading.value = false;
    }
  });

  onBeforeUnmount(() => {
    cleanupClipboard?.();
    cleanupSession?.();
    cleanupFileTransfer?.();
    cleanupPairRequest?.();
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
      selectedHost.value = resolvePreferredHost(
        nextBootstrap,
        selectedHost.value,
        window.localStorage.getItem(preferredHostStorageKey) ?? "",
      );
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function syncDesktopPinned() {
    try {
      desktopPinned.value = await desktopWorkbench.getDesktopPinned();
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
        pinnedOnly: false,
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
        message.success("剪贴板已刷新");
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

  async function refreshLinkedDevices() {
    if (!available) {
      return;
    }

    try {
      linkedDevices.value = await desktopWorkbench.listLinkedWebDevices();
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function refreshPairRequests() {
    if (!available) {
      return;
    }

    try {
      pairRequests.value = await desktopWorkbench.listPairRequests();
      pairRequestModalOpen.value = pairRequests.value.length > 0;
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handleRowClick(item: ClipboardItemRecord) {
    selectedId.value = item.id;
    if (isDesktopFileItem(item)) {
      await handleFilePrimaryAction(item);
      return;
    }
    await handleActivate(item.id, false);
  }

  async function openWebPanel() {
    await loadBootstrap();
    await refreshLinkedDevices();
    webPanelOpen.value = true;
  }

  async function handleIncomingPairRequest() {
    try {
      await desktopWorkbench.showDesktopApp();
    } catch {
      // The backend already attempts to show the app.
    }
    await refreshPairRequests();
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
      return;
    }

    if (event.altKey || event.ctrlKey || event.metaKey || isEditableTarget(event.target)) {
      return;
    }

    if (event.key === "ArrowDown") {
      event.preventDefault();
      moveSelection(1);
      return;
    }

    if (event.key === "ArrowUp") {
      event.preventDefault();
      moveSelection(-1);
      return;
    }

    if (event.key === "Enter" || event.key === " " || event.key === "Spacebar") {
      event.preventDefault();
      void activateSelectedItem();
      return;
    }
  }

  function moveSelection(delta: number) {
    if (items.value.length === 0) {
      return;
    }

    const currentIndex = selectedId.value ? items.value.findIndex((item) => item.id === selectedId.value) : -1;
    if (currentIndex < 0) {
      selectedId.value = items.value[0]?.id ?? null;
      return;
    }

    const nextIndex = Math.min(Math.max(currentIndex + delta, 0), items.value.length - 1);
    selectedId.value = items.value[nextIndex]?.id ?? selectedId.value;
  }

  async function activateSelectedItem() {
    const item = items.value.find((entry) => entry.id === selectedId.value) ?? items.value[0] ?? null;
    if (!item) {
      return;
    }

    await handleRowClick(item);
  }

  function selectFirstItem() {
    selectedId.value = items.value[0]?.id ?? null;
  }

  function isEditableTarget(target: EventTarget | null) {
    if (!(target instanceof HTMLElement)) {
      return false;
    }

    const tagName = target.tagName;
    if (tagName === "INPUT" || tagName === "TEXTAREA" || tagName === "SELECT") {
      return true;
    }

    return target.isContentEditable;
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
        message.success("已复制到系统剪贴板");
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
      message.success(item.pinned ? "已取消置顶" : "已置顶");
      await refreshHistory(true);
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handleDelete(item: ClipboardItemRecord) {
    const label = item.preview || item.content || item.fileMeta?.fileName || item.id;
    if (!window.confirm(`删除这条记录？\n\n${label}`)) {
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

  async function handleSync(item: ClipboardItemRecord, targetDeviceIds: string[], syncAll: boolean) {
    try {
      const result = await desktopWorkbench.syncClipboardItem(item.id, targetDeviceIds, syncAll);
      if (result.deliveredDevices.length === 0) {
        message.warning("没有可用的目标设备");
        return;
      }
      message.success(`已同步到 ${result.deliveredDevices.length} 台设备`);
      await refreshOnlineDevices();
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handleFilePrimaryAction(item: ClipboardItemRecord) {
    if (!isDesktopFileItem(item)) {
      return;
    }

    if (item.fileMeta?.transferState === "receiving" || item.fileMeta?.transferState === "metadata_only") {
      message.warning("文件仍在传输中");
      return;
    }

    if (item.fileMeta?.transferState === "failed") {
      message.warning("文件传输失败，请重新同步");
      return;
    }

    if (item.fileMeta?.transferState === "received") {
      await handleActivate(item.id, true);
      return;
    }

    message.warning("文件暂时不可用");
    return;
  }

  async function handleReceiveClipboardFile(
    item: ClipboardItemRecord,
    options: {
      hideAfterReceive?: boolean;
      openDetail?: boolean;
      successMessage?: string;
    } = {},
  ) {
    if (!isDesktopFileItem(item)) {
      return;
    }

    try {
      const received = await desktopWorkbench.receiveClipboardFile(item.id);
      items.value = items.value.map((entry) => (entry.id === received.id ? received : entry));
      if (options.openDetail !== false) {
        detailItemId.value = received.id;
        detailPanelOpen.value = true;
      }
      message.success(options.successMessage || "文件已接收");
      await refreshHistory(true);
      if (options.hideAfterReceive) {
        await hideDesktopApp();
      }
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
    if (!detailItem.value || isDesktopFileItem(detailItem.value)) {
      return;
    }
    await copyText(detailItem.value.content);
    message.success("内容已复制");
  }

  async function handleCopyEntry() {
    if (!resolvedSessionUrl.value) {
      return;
    }
    await copyText(resolvedSessionUrl.value);
    message.success("关联入口已复制");
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
      message.success("关联入口已刷新");
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handleRemoveLinkedDevice(deviceId: string) {
    try {
      await desktopWorkbench.removeLinkedWebDevice(deviceId);
      message.success("设备关联已移除");
      await refreshLinkedDevices();
      await refreshOnlineDevices();
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handleApprovePairRequest() {
    const request = activePairRequest.value;
    if (!request) {
      pairRequestModalOpen.value = false;
      return;
    }

    try {
      await desktopWorkbench.approvePairRequest(request.id);
      message.success("已同意设备关联");
      await refreshPairRequests();
      await refreshLinkedDevices();
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handleRejectPairRequest() {
    const request = activePairRequest.value;
    if (!request) {
      pairRequestModalOpen.value = false;
      return;
    }

    try {
      await desktopWorkbench.rejectPairRequest(request.id);
      message.success("已拒绝设备关联");
      await refreshPairRequests();
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
    if (key === "toggleDesktopPinned") {
      await handleDesktopPinnedToggle();
      return;
    }
    if (key === "refresh") {
      await refreshHistory(false);
      return;
    }
    if (key === "clear") {
      await handleClearHistory();
    }
  }

  async function handleDesktopPinnedToggle() {
    try {
      desktopPinned.value = await desktopWorkbench.setDesktopPinned(!desktopPinned.value);
      message.success(desktopPinned.value ? "窗口已固定" : "窗口已取消固定");
    } catch (error) {
      message.error(describeError(error));
    }
  }

  async function handleClearHistory() {
    if (!window.confirm("清空全部剪贴板历史？")) {
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
    activePairRequest,
    handleCandidateSelect,
    handleContextSelect,
    handleCopyDetail,
    handleCopyEntry,
    handleDeleteDetail,
    handleMoreSelect,
    handleOpenEntry,
    handleApprovePairRequest,
    handleRejectPairRequest,
    handleRemoveLinkedDevice,
    handleRefreshSession,
    handleRowClick,
    items,
    linkedDevices,
    loading,
    moreOptions,
    openContextMenu,
    openDiagnostics,
    pairRequestModalOpen,
    refreshing,
    resolvedSessionUrl,
    search,
    selectFirstItem,
    selectedId,
    tokenCountdown,
    webPanelOpen,
  };
}
