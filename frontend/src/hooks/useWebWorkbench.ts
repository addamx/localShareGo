import { computed, onBeforeUnmount, ref, watch } from "vue";
import UAParser from "ua-parser-js";
import { useMessage, type DropdownOption } from "naive-ui";
import { useRoute } from "vue-router";

import { copyText } from "../app/env";
import { useForgeStorage } from "./useForgeStorage";
import { createWorkbenchApiClient, WorkbenchApiError, type WorkbenchApiClient } from "../services/workbenchApi";
import type { ClipboardItemRecord, OnlineDevice, ServerEvent, SessionResponse } from "../types/workbench";

const forgeOptions = {
  createInstance: {
    name: "localShareGo",
    storeName: "webClipboard",
  },
} as const;

const webClipboardStorageKey = "localsharego:web:clipboard-items";
const webDeviceIDStorageKey = "localsharego:web:device-id";
const webComposeSyncTargetsStorageKey = "localsharego:web:compose-sync-targets";

export function useWebWorkbench() {
  const route = useRoute();
  const message = useMessage();

  const initializing = ref(true);
  const authState = ref<"idle" | "missing" | "valid" | "invalid">("idle");
  const draft = ref("");
  const search = ref("");
  const pinnedOnly = ref(false);
  const selectedId = ref<string | null>(null);
  const session = ref<SessionResponse | null>(null);
  const api = ref<WorkbenchApiClient | null>(null);
  const localDevice = ref<OnlineDevice | null>(null);
  const onlineDevices = ref<OnlineDevice[]>([]);
  const clock = ref(Date.now());
  const contextMenuOpen = ref(false);
  const contextMenuX = ref(0);
  const contextMenuY = ref(0);
  const contextMenuItem = ref<ClipboardItemRecord | null>(null);

  const storedItems = useForgeStorage<ClipboardItemRecord[]>(webClipboardStorageKey, [], forgeOptions);
  const storedComposeSyncTargetIDs = useForgeStorage<string[] | null>(webComposeSyncTargetsStorageKey, null, forgeOptions);

  const token = computed(() => {
    const value = route.query.token;
    return typeof value === "string" ? value.trim() : "";
  });
  const items = computed(() => {
    const keyword = search.value.trim().toLowerCase();
    return storedItems.value.filter((item) => {
      if (pinnedOnly.value && !item.pinned) {
        return false;
      }
      if (!keyword) {
        return true;
      }
      return item.preview.toLowerCase().includes(keyword) || item.content.toLowerCase().includes(keyword);
    });
  });
  const draftBytes = computed(() => new TextEncoder().encode(draft.value).length);
  const expiresIn = computed(() => formatCountdown(session.value?.expiresAt ?? null, clock.value));
  const selectedComposeDeviceIDs = computed(() => normalizeComposeSyncTargets(storedComposeSyncTargetIDs.value));
  const selectedComposeDevices = computed(() =>
    onlineDevices.value.filter((device) => selectedComposeDeviceIDs.value.includes(device.id)),
  );
  const syncAvailable = computed(() => onlineDevices.value.length > 0);
  const composeSyncOptions = computed<DropdownOption[]>(() => {
    if (onlineDevices.value.length === 0) {
      return [{ label: "暂无可同步设备", key: "compose-sync:none", disabled: true }];
    }

    return onlineDevices.value.map((device) => ({
      label: selectedComposeDeviceIDs.value.includes(device.id) ? `✓ ${device.name}` : device.name,
      key: device.id,
    }));
  });
  const composeSyncLabel = computed(() => {
    const devices = selectedComposeDevices.value;
    if (devices.length === 0) {
      return "不同步";
    }
    if (devices.length === 1) {
      return devices[0].name;
    }
    return `${devices[0].name} +${devices.length - 1}`;
  });
  const contextMenuOptions = computed<DropdownOption[]>(() => {
    if (!contextMenuItem.value) {
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
      { label: contextMenuItem.value.pinned ? "取消置顶" : "置顶", key: "pin" },
      { label: "同步", key: "sync", children: syncChildren },
    ];
  });

  let eventSource: EventSource | null = null;
  let heartbeatTimer: number | null = null;
  let clockTimer: number | null = null;

  void storedItems.then(() => {
    syncSelectedItem();
  });
  void storedComposeSyncTargetIDs;

  watch(
    token,
    () => {
      void initializePage();
    },
    { immediate: true },
  );

  onBeforeUnmount(() => {
    closeContextMenu();
    stopHeartbeat();
    stopClock();
    closeEvents();
  });

  async function initializePage() {
    stopHeartbeat();
    stopClock();
    closeEvents();
    closeContextMenu();

    initializing.value = true;
    authState.value = "idle";
    session.value = null;
    api.value = null;
    localDevice.value = null;
    onlineDevices.value = [];

    if (!token.value) {
      authState.value = "missing";
      initializing.value = false;
      return;
    }

    api.value = createWorkbenchApiClient(window.location.origin, token.value);

    try {
      await api.value.health();
      session.value = await api.value.session();
      startClock();
      authState.value = "valid";
      await initializePresence();
    } catch (error) {
      if (error instanceof WorkbenchApiError && error.status === 401) {
        authState.value = "invalid";
      } else {
        message.error(describeError(error));
      }
    } finally {
      initializing.value = false;
    }
  }

  async function initializePresence() {
    try {
      await registerDevice();
      connectEvents();
      startHeartbeat();
    } catch (error) {
      if (error instanceof WorkbenchApiError && error.status === 401) {
        authState.value = "invalid";
        stopClock();
        return;
      }
      message.error(describeError(error));
    }
  }

  async function registerDevice() {
    if (!api.value) {
      return;
    }

    const deviceID = getOrCreateLocalDeviceID();
    const response = await api.value.registerWebDevice(buildWebDeviceName(), deviceID);
    localDevice.value = response.self;
    persistLocalDeviceID(response.self.id);
    onlineDevices.value = response.devices;
    syncComposeTargets();
    hydrateOwnItems();
  }

  async function refreshPresence() {
    if (!api.value || !localDevice.value) {
      return;
    }

    try {
      const response = await api.value.heartbeatWebDevice(localDevice.value.id);
      localDevice.value = response.self;
      onlineDevices.value = response.devices;
      syncComposeTargets();
    } catch (error) {
      if (error instanceof WorkbenchApiError && error.status === 404) {
        await registerDevice();
        return;
      }
      if (error instanceof WorkbenchApiError && error.status === 401) {
        authState.value = "invalid";
        stopHeartbeat();
        stopClock();
        closeEvents();
        return;
      }
      message.error(describeError(error));
    }
  }

  async function submitDraft() {
    const content = draft.value.trim();
    if (!content) {
      message.warning("先输入文本");
      return;
    }

    insertLocalItem(content, localDevice.value?.name ?? "Browser", true, localDevice.value?.id ?? null);
    if (selectedComposeDeviceIDs.value.length > 0) {
      await syncClipboardContent(content, selectedComposeDeviceIDs.value, false);
    } else {
      message.success("已保存到当前浏览器");
    }
    draft.value = "";
  }

  function handleComposeSyncSelect(key: string | number) {
    const targetDeviceID = String(key);
    if (targetDeviceID === "compose-sync:none") {
      return;
    }

    const currentTargets = normalizeComposeSyncTargets(storedComposeSyncTargetIDs.value);
    if (currentTargets.includes(targetDeviceID)) {
      storedComposeSyncTargetIDs.value = currentTargets.filter((deviceID) => deviceID !== targetDeviceID);
      return;
    }

    storedComposeSyncTargetIDs.value = [...currentTargets, targetDeviceID];
  }

  async function activateItem(itemId: string) {
    const item = storedItems.value.find((entry) => entry.id === itemId);
    if (!item) {
      return;
    }

    await copyText(item.content);
    selectedId.value = itemId;
    const now = Date.now();
    const updatedItems = storedItems.value.map((entry) => ({
      ...entry,
      isCurrent: entry.id === itemId,
      updatedAt: entry.id === itemId ? now : entry.updatedAt,
    }));
    const currentItem = updatedItems.find((entry) => entry.id === itemId);
    if (!currentItem) {
      return;
    }

    const pinnedItems = updatedItems.filter((entry) => entry.pinned && entry.id !== itemId);
    const unpinnedItems = updatedItems.filter((entry) => !entry.pinned && entry.id !== itemId);
    storedItems.value = currentItem.pinned
      ? [currentItem, ...pinnedItems, ...unpinnedItems]
      : [...pinnedItems, currentItem, ...unpinnedItems];
    message.success("已设为当前并复制到剪贴板");
  }

  function togglePinnedOnly() {
    pinnedOnly.value = !pinnedOnly.value;
  }

  function openContextMenu(event: MouseEvent, item: ClipboardItemRecord) {
    void refreshPresence();
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
      toggleItemPin(item.id);
      return;
    }
    if (key === "sync:all") {
      await syncClipboardContent(item.content, [], true);
      return;
    }

    const textKey = String(key);
    if (textKey.startsWith("sync:")) {
      const targetDeviceID = textKey.slice(5);
      if (targetDeviceID) {
        await syncClipboardContent(item.content, [targetDeviceID], false);
      }
    }
  }

  async function syncClipboardContent(content: string, targetDeviceIds: string[], syncAll: boolean) {
    if (!api.value || !localDevice.value) {
      return false;
    }

    try {
      const result = await api.value.syncClipboard({
        content,
        sourceDeviceId: localDevice.value.id,
        targetDeviceIds,
        syncAll,
      });
      if (result.deliveredDevices.length === 0) {
        message.warning("没有可同步的目标设备");
        return false;
      }

      message.success(`已同步到 ${result.deliveredDevices.length} 个设备`);
      await refreshPresence();
      return true;
    } catch (error) {
      if (error instanceof WorkbenchApiError && error.status === 401) {
        authState.value = "invalid";
        stopHeartbeat();
        stopClock();
        closeEvents();
        return false;
      }
      message.error(describeError(error));
      return false;
    }
  }

  function connectEvents() {
    if (!api.value) {
      return;
    }

    closeEvents();
    eventSource = new EventSource(api.value.eventsUrl());
    eventSource.addEventListener("sync", (event) => {
      const payload = JSON.parse((event as MessageEvent).data) as ServerEvent;
      const sync = payload.sync;
      if (!sync || !localDevice.value) {
        return;
      }
      if (!sync.targetDeviceIds.includes(localDevice.value.id)) {
        return;
      }

      insertLocalItem(sync.content, sync.sourceKind, false, null);
      message.success(`收到来自 ${sync.sourceKind} 的同步`);
    });
  }

  function closeEvents() {
    eventSource?.close();
    eventSource = null;
  }

  function startHeartbeat() {
    stopHeartbeat();
    heartbeatTimer = window.setInterval(() => {
      void refreshPresence();
    }, 20_000);
  }

  function stopHeartbeat() {
    if (heartbeatTimer !== null) {
      window.clearInterval(heartbeatTimer);
      heartbeatTimer = null;
    }
  }

  function startClock() {
    stopClock();
    clockTimer = window.setInterval(() => {
      clock.value = Date.now();
    }, 1000);
  }

  function stopClock() {
    if (clockTimer !== null) {
      window.clearInterval(clockTimer);
      clockTimer = null;
    }
  }

  function hydrateOwnItems() {
    const sourceName = localDevice.value?.name;
    if (!sourceName) {
      return;
    }

    storedItems.value = storedItems.value.map((item) =>
      item.sourceKind === "Browser" ? { ...item, sourceKind: sourceName } : item,
    );
    syncSelectedItem();
  }

  function insertLocalItem(content: string, sourceKind: string, markCurrent: boolean, sourceDeviceID: string | null) {
    const trimmed = content.trim();
    if (!trimmed) {
      return;
    }

    const now = Date.now();
    const nextItem: ClipboardItemRecord = {
      id: createLocalID(),
      content: trimmed,
      preview: buildPreview(trimmed),
      charCount: [...trimmed].length,
      sourceKind,
      sourceDeviceId: sourceDeviceID,
      pinned: false,
      isCurrent: markCurrent,
      createdAt: now,
      updatedAt: now,
    };

    storedItems.value = [
      nextItem,
      ...storedItems.value.map((item) =>
        markCurrent && item.isCurrent ? { ...item, isCurrent: false, updatedAt: now } : item,
      ),
    ];
    selectedId.value = nextItem.id;
  }

  function toggleItemPin(itemId: string) {
    storedItems.value = storedItems.value.map((item) =>
      item.id === itemId ? { ...item, pinned: !item.pinned, updatedAt: Date.now() } : item,
    );
  }

  function syncSelectedItem() {
    if (!selectedId.value || !storedItems.value.some((item) => item.id === selectedId.value)) {
      selectedId.value = storedItems.value.find((item) => item.isCurrent)?.id ?? storedItems.value[0]?.id ?? null;
    }
  }

  function syncComposeTargets() {
    if (onlineDevices.value.length === 0) {
      if (storedComposeSyncTargetIDs.value === null) {
        storedComposeSyncTargetIDs.value = [];
      }
      return;
    }

    const onlineDeviceIDs = new Set(onlineDevices.value.map((device) => device.id));
    const currentTargets = normalizeComposeSyncTargets(storedComposeSyncTargetIDs.value).filter((deviceID) =>
      onlineDeviceIDs.has(deviceID),
    );
    if (storedComposeSyncTargetIDs.value === null && currentTargets.length === 0) {
      currentTargets.push(onlineDevices.value[0].id);
    }
    storedComposeSyncTargetIDs.value = currentTargets;
  }

  return {
    activateItem,
    authState,
    closeContextMenu,
    composeSyncLabel,
    composeSyncOptions,
    contextMenuOpen,
    contextMenuOptions,
    contextMenuX,
    contextMenuY,
    draft,
    draftBytes,
    expiresIn,
    handleComposeSyncSelect,
    handleContextSelect,
    initializing,
    items,
    openContextMenu,
    pinnedOnly,
    search,
    selectedId,
    session,
    submitDraft,
    syncAvailable,
    togglePinnedOnly,
  };
}

function buildWebDeviceName() {
  const parser = new UAParser(window.navigator.userAgent);
  const browser = parser.getBrowser().name?.trim() || "Browser";
  const os = parser.getOS().name?.trim() || "OS";
  return `${browser}-${os}`;
}

function buildPreview(content: string) {
  const normalized = content.split(/\s+/).filter(Boolean).join(" ");
  const limit = 120;
  return normalized.length <= limit ? normalized : `${normalized.slice(0, limit)}...`;
}

function getOrCreateLocalDeviceID() {
  const stored = window.localStorage.getItem(webDeviceIDStorageKey)?.trim() ?? "";
  if (stored) {
    return stored;
  }

  const deviceID = createLocalID();
  window.localStorage.setItem(webDeviceIDStorageKey, deviceID);
  return deviceID;
}

function persistLocalDeviceID(deviceID: string) {
  window.localStorage.setItem(webDeviceIDStorageKey, deviceID.trim());
}

function createLocalID() {
  const cryptoObject = window.crypto;
  if (typeof cryptoObject?.randomUUID === "function") {
    return cryptoObject.randomUUID();
  }

  if (typeof cryptoObject?.getRandomValues === "function") {
    const buffer = new Uint8Array(16);
    cryptoObject.getRandomValues(buffer);
    buffer[6] = (buffer[6] & 0x0f) | 0x40;
    buffer[8] = (buffer[8] & 0x3f) | 0x80;
    const hex = Array.from(buffer, (value) => value.toString(16).padStart(2, "0"));
    return `${hex.slice(0, 4).join("")}-${hex.slice(4, 6).join("")}-${hex.slice(6, 8).join("")}-${hex.slice(8, 10).join("")}-${hex.slice(10, 16).join("")}`;
  }

  const seed = `${Date.now()}-${Math.random().toString(16).slice(2)}`;
  return `web-${seed}`;
}

function formatCountdown(expiresAt: number | null, now: number) {
  if (!expiresAt) {
    return "待访问";
  }

  const remaining = expiresAt - now;
  if (remaining <= 0) {
    return "已过期";
  }

  const seconds = Math.floor(remaining / 1000);
  const minutes = Math.floor(seconds / 60);
  return minutes > 0 ? `${minutes}m ${String(seconds % 60).padStart(2, "0")}s` : `${seconds}s`;
}

function describeError(error: unknown) {
  if (error instanceof Error) {
    return error.message;
  }
  if (typeof error === "string") {
    return error;
  }
  return String(error);
}

function normalizeComposeSyncTargets(value: string[] | string | null | undefined) {
  if (Array.isArray(value)) {
    return value.filter((item) => typeof item === "string" && item.trim() !== "");
  }
  if (typeof value === "string" && value.trim() !== "") {
    return [value.trim()];
  }
  return [];
}
