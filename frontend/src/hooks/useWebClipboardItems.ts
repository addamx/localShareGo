import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import localforage from "localforage";
import UAParser from "ua-parser-js";
import { useMessage, type DropdownOption } from "naive-ui";

import { copyText } from "../app/env";
import { anonymousWorkbenchApi, createWorkbenchApiClient, WorkbenchApiError, type WorkbenchApiClient } from "../services/workbenchApi";
import type {
  ClipboardFileMeta,
  ClipboardItemRecord,
  ClipboardTransferState,
  FileTransferEvent,
  OnlineDevice,
  PairRequestStatus,
  ServerEvent,
  SessionResponse,
} from "../types/workbench";
import { useForgeStorage } from "./useForgeStorage";

const forgeOptions = {
  createInstance: {
    name: "localShareGo",
    storeName: "webClipboard",
  },
} as const;

const fileForge = localforage.createInstance({
  name: "localShareGo",
  storeName: "webClipboardFiles",
});

const webClipboardStorageKey = "localsharego:web:clipboard-items";
const webCredentialStorageKey = "localsharego:web:device-credential";
const webDeviceIDStorageKey = "localsharego:web:device-id";

export function useWebClipboardItems() {
  const route = useRoute();
  const router = useRouter();
  const message = useMessage();

  const initializing = ref(true);
  const authState = ref<"idle" | "no-link" | "expired" | "requesting" | "waiting-approval" | "ready">("idle");
  const search = ref("");
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
  const activeCredential = ref("");
  const pendingPairRequestID = ref("");

  const storedItems = useForgeStorage<ClipboardItemRecord[]>(webClipboardStorageKey, [], forgeOptions);
  const routeToken = computed(() => {
    const value = route.query.token;
    return typeof value === "string" ? value.trim() : "";
  });

  const items = computed(() => {
    const keyword = search.value.trim().toLowerCase();
    return storedItems.value.filter((item) => {
      if (!keyword) {
        return true;
      }

      const searchable = [
        item.preview,
        item.content,
        item.sourceKind,
        item.fileMeta?.fileName ?? "",
        item.fileMeta?.extension ?? "",
        item.fileMeta?.mimeType ?? "",
      ];
      return searchable.some((value) => value.toLowerCase().includes(keyword));
    });
  });

  const expiresIn = computed(() => formatCountdown(session.value?.expiresAt ?? null, clock.value));
  const renewAvailable = computed(() => {
    if (!session.value?.expiresAt) {
      return false;
    }
    const remaining = session.value.expiresAt - clock.value;
    return remaining > 0 && remaining <= session.value.tokenTtlMinutes * 60_000 / 2;
  });
  const moreOptions = computed<DropdownOption[]>(() => [
    { label: "Refresh", key: "refresh" },
    { label: "Clear", key: "clear" },
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

    const options: DropdownOption[] = [];
    options.push({ label: item.pinned ? "Unpin" : "Pin", key: "pin" });
    options.push({ label: "Sync", key: "sync", children: syncChildren });
    options.push({ label: "Delete", key: "delete" });
    return options;
  });

  let eventSource: EventSource | null = null;
  let heartbeatTimer: number | null = null;
  let clockTimer: number | null = null;
  let pairRequestTimer: number | null = null;
  const activeFileTransfers = new Map<string, Promise<ClipboardItemRecord | null>>();

  void storedItems.then(() => {
    syncSelectedItem();
  });

  watch(
    routeToken,
    () => {
      if (routeToken.value === activeCredential.value && authState.value === "ready" && api.value) {
        return;
      }
      void initializePage();
    },
    { immediate: true },
  );

  onBeforeUnmount(() => {
    closeContextMenu();
    stopHeartbeat();
    stopClock();
    stopPairRequestPolling();
    closeEvents();
  });

  async function initializePage() {
    stopHeartbeat();
    stopClock();
    stopPairRequestPolling();
    closeEvents();
    closeContextMenu();

    initializing.value = true;
    authState.value = "idle";
    session.value = null;
    api.value = null;
    localDevice.value = null;
    onlineDevices.value = [];
    activeCredential.value = "";

    const urlToken = routeToken.value;
    if (urlToken) {
      const activatedCredential = await tryActivateEntry(urlToken);
      if (activatedCredential) {
        if (await initializeAuthorizedSession(activatedCredential, true)) {
          await replaceCredentialInUrl(activatedCredential);
          initializing.value = false;
          return;
        }

        clearCredential();
        await replaceCredentialInUrl(null);
        authState.value = "expired";
        initializing.value = false;
        return;
      }

      if (await initializeAuthorizedSession(urlToken, true)) {
        initializing.value = false;
        return;
      }

      clearCredential();
      await replaceCredentialInUrl(null);
      authState.value = "expired";
      initializing.value = false;
      return;
    }

    const storedCredential = readStoredCredential();
    if (!storedCredential) {
      authState.value = "no-link";
      initializing.value = false;
      return;
    }

    if (await initializeAuthorizedSession(storedCredential, true)) {
      await replaceCredentialInUrl(storedCredential);
      initializing.value = false;
      return;
    }

    clearCredential();
    authState.value = "expired";
    initializing.value = false;
  }

  async function tryActivateEntry(entryToken: string) {
    try {
      const response = await anonymousWorkbenchApi.activateEntry(
        window.location.origin,
        entryToken,
        getOrCreateLocalDeviceID(),
        buildWebDeviceName(),
      );
      return response.credential.trim();
    } catch (error) {
      if (error instanceof WorkbenchApiError && (error.status === 400 || error.status === 401 || error.status === 404)) {
        return "";
      }
      message.error(describeError(error));
      return "";
    }
  }

  async function initializeAuthorizedSession(credential: string, persist: boolean) {
    api.value = createWorkbenchApiClient(window.location.origin, credential);

    try {
      session.value = await api.value.session();
      if (persist) {
        persistCredential(credential);
      }
      activeCredential.value = credential;
      if (!(await initializePresence())) {
        return false;
      }
      startClock();
      authState.value = "ready";
      return true;
    } catch (error) {
      if (error instanceof WorkbenchApiError && error.status === 401) {
        return false;
      }
      message.error(describeError(error));
      return false;
    }
  }

  async function initializePresence() {
    try {
      await registerDevice();
      connectEvents();
      startHeartbeat();
      return true;
    } catch (error) {
      if (error instanceof WorkbenchApiError && error.status === 401) {
        expireAssociation();
        return false;
      }
      message.error(describeError(error));
      return true;
    }
  }

  async function registerDevice() {
    if (!api.value || !session.value) {
      return;
    }

    const deviceID = session.value.selfDeviceId || getOrCreateLocalDeviceID();
    const response = await api.value.registerWebDevice(buildWebDeviceName(), deviceID);
    localDevice.value = response.self;
    persistLocalDeviceID(response.self.id);
    onlineDevices.value = response.devices;
    hydrateOwnItems();
    syncSelectedItem();
    queuePendingFileTransfers();
  }

  async function refreshPresence() {
    if (!api.value || !localDevice.value) {
      return;
    }

    try {
      const response = await api.value.heartbeatWebDevice(localDevice.value.id);
      localDevice.value = response.self;
      onlineDevices.value = response.devices;
      hydrateOwnItems();
    } catch (error) {
      if (error instanceof WorkbenchApiError && error.status === 404) {
        await registerDevice();
        return;
      }
      if (error instanceof WorkbenchApiError && error.status === 401) {
        expireAssociation();
        return;
      }
      message.error(describeError(error));
    }
  }

  async function submitTextDraft(content: string, targetDeviceIds: string[], syncAll: boolean) {
    if (!api.value || !localDevice.value) {
      return false;
    }

    const trimmed = content.trim();
    if (!trimmed) {
      return false;
    }

    insertLocalText(trimmed, localDevice.value.name, true, localDevice.value.id);

    if (targetDeviceIds.length === 0 && !syncAll) {
      message.success("Saved to the current browser");
      return true;
    }

    try {
      const result = await api.value.syncClipboard({
        itemId: null,
        content: trimmed,
        sourceDeviceId: localDevice.value.id,
        targetDeviceIds,
        syncAll,
      });
      if (result.deliveredDevices.length === 0) {
        message.warning("No online target devices");
        return false;
      }

      message.success(`Synced to ${result.deliveredDevices.length} device(s)`);
      await refreshPresence();
      return true;
    } catch (error) {
      if (error instanceof WorkbenchApiError && error.status === 401) {
        expireAssociation();
        return false;
      }
      message.error(describeError(error));
      return false;
    }
  }

  async function uploadFile(file: File, targetDeviceIds: string[]) {
    if (!api.value || !localDevice.value) {
      return false;
    }

    try {
      const result = await api.value.createFileItem(file);
      const nextItem = markFileCached(
        normalizeClipboardItemRecord({
          ...result.item,
          sourceKind: localDevice.value.name,
          sourceDeviceId: localDevice.value.id,
        }),
        file.name,
      );
      await storeFileBlob(nextItem.id, file);
      upsertItem(nextItem);
      selectedId.value = nextItem.id;
      let synced = true;
      if (targetDeviceIds.length > 0) {
        synced = await syncClipboardItem(nextItem, targetDeviceIds, false);
      }
      await refreshPresence();
      if (synced) {
        message.success(`Uploaded ${file.name}`);
      }
      return synced;
    } catch (error) {
      if (error instanceof WorkbenchApiError && error.status === 401) {
        expireAssociation();
        return false;
      }
      message.error(describeError(error));
      return false;
    }
  }

  async function activateItem(itemId: string) {
    const item = storedItems.value.find((entry) => entry.id === itemId);
    if (!item) {
      return;
    }

    selectedId.value = itemId;
    if (item.itemKind === "file") {
      return;
    }

    await copyText(item.content);
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

    storedItems.value = sortClipboardItems([currentItem, ...updatedItems.filter((entry) => entry.id !== itemId)]);
    message.success("Copied to clipboard");
  }

  async function handleRowClick(item: ClipboardItemRecord) {
    if (item.itemKind === "file") {
      await copyFileItem(item);
      return;
    }

    await activateItem(item.id);
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

  async function handleMoreSelect(key: string | number) {
    const textKey = String(key);
    if (textKey === "refresh") {
      await handleRefreshList();
      return;
    }
    if (textKey === "clear") {
      await handleClearHistory();
    }
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
    if (key === "delete") {
      deleteItem(item.id);
      return;
    }
    if (key === "sync:all") {
      await syncClipboardItem(item, [], true);
      return;
    }

    const textKey = String(key);
    if (textKey.startsWith("sync:")) {
      const targetDeviceID = textKey.slice(5);
      if (targetDeviceID) {
        await syncClipboardItem(item, [targetDeviceID], false);
      }
    }
  }

  async function syncClipboardItem(item: ClipboardItemRecord, targetDeviceIds: string[], syncAll: boolean) {
    if (!api.value || !localDevice.value) {
      return false;
    }

    try {
      const result = await api.value.syncClipboard({
        itemId: item.itemKind === "file" ? item.id : null,
        content: item.itemKind === "file" ? "" : item.content,
        sourceDeviceId: localDevice.value.id,
        targetDeviceIds,
        syncAll,
      });
      if (result.deliveredDevices.length === 0) {
        message.warning("No online target devices");
        return false;
      }

      message.success(`Synced to ${result.deliveredDevices.length} device(s)`);
      await refreshPresence();
      return true;
    } catch (error) {
      if (error instanceof WorkbenchApiError && error.status === 401) {
        expireAssociation();
        return false;
      }
      message.error(describeError(error));
      return false;
    }
  }

  async function receiveFileItem(item: ClipboardItemRecord) {
    if (!api.value || item.itemKind !== "file") {
      return false;
    }

    try {
      const updated = await prepareFileItem(item);
      if (!updated) {
        return false;
      }
      triggerFileDownload(api.value.fileContentUrl(item.id), updated.fileMeta?.fileName ?? updated.preview ?? updated.id);
      return true;
    } catch (error) {
      markFileTransfer(item.id, "failed", 0, describeError(error));
      if (error instanceof WorkbenchApiError && error.status === 401) {
        expireAssociation();
        return false;
      }
      message.error(describeError(error));
      return false;
    }
  }

  function receiveFileItemById(itemId: string) {
    const item = storedItems.value.find((entry) => entry.id === itemId);
    if (!item || item.itemKind !== "file") {
      return Promise.resolve(false);
    }

    return receiveFileItem(item);
  }

  async function copyFileItem(item: ClipboardItemRecord) {
    if (!api.value || item.itemKind !== "file") {
      return false;
    }

    selectedId.value = item.id;

    if (item.fileMeta?.transferState === "receiving" || activeFileTransfers.has(item.id)) {
      message.warning("File is still transferring");
      return false;
    }

    try {
      let currentItem = item;
      let blob = await loadFileBlob(item.id);
      if (!blob) {
        void ensureFileAvailable(item);
        message.warning("File is still transferring");
        return false;
      }

      currentItem = markFileCached(currentItem, currentItem.fileMeta?.fileName);
      upsertItem(currentItem);
      const copied = await writeFileBlobToClipboard(
        blob,
        currentItem.fileMeta?.mimeType || blob.type || "application/octet-stream",
      );
      if (copied) {
        message.success("Copied to clipboard");
        return true;
      }

      triggerBlobDownload(blob, currentItem.fileMeta?.fileName ?? currentItem.preview ?? currentItem.id);
      message.warning("Browser clipboard does not reliably support arbitrary files, downloaded instead");
      return false;
    } catch (error) {
      message.error(describeError(error));
      return false;
    }
  }

  async function prepareFileItem(item: ClipboardItemRecord) {
    if (!api.value || item.itemKind !== "file") {
      return null
    }

    markFileTransfer(item.id, "receiving", 0, null);
    const updated = await api.value.receiveFileItem(item.id);
    const normalized = normalizeClipboardItemRecord(updated);
    upsertItem(normalized);
    return normalized;
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

      const nextItem = normalizeClipboardItemRecord(sync.item);
      upsertItem(nextItem);
      if (nextItem.itemKind === "file") {
        void ensureFileAvailable(nextItem);
      }
      message.success(`Received sync from ${sync.item.sourceKind}`);
    });
    eventSource.addEventListener("file-transfer", (event) => {
      const payload = JSON.parse((event as MessageEvent).data) as ServerEvent;
      if (!payload.fileTransfer) {
        return;
      }
      applyFileTransferEvent(payload.fileTransfer);
    });
    eventSource.addEventListener("refresh", (event) => {
      const payload = JSON.parse((event as MessageEvent).data) as ServerEvent;
      if (payload.scope !== "clipboard") {
        return;
      }
      queuePendingFileTransfers();
    });
    eventSource.addEventListener("revoked", (event) => {
      const payload = JSON.parse((event as MessageEvent).data) as ServerEvent;
      if (!payload.revoked || !session.value || payload.revoked.sessionId !== session.value.sessionId) {
        return;
      }
      expireAssociation();
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
      const now = Date.now();
      clock.value = now;
      if (session.value?.expiresAt && now >= session.value.expiresAt) {
        expireAssociation();
      }
    }, 1000);
  }

  function stopClock() {
    if (clockTimer !== null) {
      window.clearInterval(clockTimer);
      clockTimer = null;
    }
  }

  function startPairRequestPolling() {
    stopPairRequestPolling();
    pairRequestTimer = window.setInterval(() => {
      void pollPairRequestStatus();
    }, 2_000);
  }

  function stopPairRequestPolling() {
    if (pairRequestTimer !== null) {
      window.clearInterval(pairRequestTimer);
      pairRequestTimer = null;
    }
  }

  function expireAssociation() {
    clearCredential();
    stopHeartbeat();
    stopClock();
    stopPairRequestPolling();
    closeEvents();
    api.value = null;
    localDevice.value = null;
    onlineDevices.value = [];
    session.value = null;
    activeCredential.value = "";
    pendingPairRequestID.value = "";
    authState.value = "expired";
    void replaceCredentialInUrl(null);
  }

  function hydrateOwnItems() {
    const sourceName = localDevice.value?.name;
    const sourceDeviceId = localDevice.value?.id;
    if (!sourceName) {
      return;
    }

    storedItems.value = storedItems.value
      .filter((item) => item.sourceKind !== "desktop_local")
      .map((item) => {
        if (item.sourceKind === "Browser" || item.sourceKind === "mobile_web") {
          return {
            ...item,
            sourceKind: sourceName,
            sourceDeviceId: sourceDeviceId ?? item.sourceDeviceId,
          };
        }
        return item;
      });
    syncSelectedItem();
  }

  function insertLocalText(content: string, sourceKind: string, markCurrent: boolean, sourceDeviceID: string | null) {
    const trimmed = content.trim();
    if (!trimmed) {
      return;
    }

    const now = Date.now();
    const nextItem = normalizeClipboardItemRecord({
      itemKind: "text",
      id: createLocalID(),
      content: trimmed,
      contentType: "text/plain",
      hash: hashText(trimmed),
      preview: buildPreview(trimmed),
      charCount: [...trimmed].length,
      fileMeta: null,
      sourceKind,
      sourceDeviceId: sourceDeviceID,
      pinned: false,
      isCurrent: markCurrent,
      createdAt: now,
      updatedAt: now,
    });

    const currentItems = storedItems.value.map((item) =>
      markCurrent && item.isCurrent ? { ...item, isCurrent: false, updatedAt: now } : item,
    );
    storedItems.value = sortClipboardItems([nextItem, ...currentItems]);
    selectedId.value = nextItem.id;
  }

  function upsertItem(next: ClipboardItemRecord) {
    const normalized = normalizeClipboardItemRecord(next);
    const existingIndex = storedItems.value.findIndex((item) => item.id === normalized.id);
    if (existingIndex >= 0) {
      const current = storedItems.value[existingIndex];
      const merged = mergeClipboardItem(current, normalized);
      const nextItems = [...storedItems.value];
      nextItems[existingIndex] = merged;
      storedItems.value = sortClipboardItems(nextItems);
      syncSelectedItem();
      return;
    }

    storedItems.value = sortClipboardItems([normalized, ...storedItems.value]);
    syncSelectedItem();
  }

  function mergeClipboardItem(existing: ClipboardItemRecord, incoming: ClipboardItemRecord) {
    const merged = normalizeClipboardItemRecord({
      ...existing,
      ...incoming,
      content: incoming.content !== "" ? incoming.content : existing.content,
      fileMeta: mergeFileMeta(existing.fileMeta, incoming.fileMeta),
    });
    if (merged.itemKind === "file") {
      merged.content = "";
      merged.charCount = 0;
    }
    return merged;
  }

  function mergeFileMeta(existing: ClipboardFileMeta | null, incoming: ClipboardFileMeta | null) {
    if (!existing && !incoming) {
      return null;
    }
    if (!existing) {
      return cloneFileMeta(incoming);
    }
    if (!incoming) {
      return cloneFileMeta(existing);
    }

    return {
      ...existing,
      ...incoming,
      transferState: resolveTransferState(existing.transferState, incoming.transferState),
      progressPercent:
        resolveTransferState(existing.transferState, incoming.transferState) === existing.transferState
          ? existing.progressPercent
          : incoming.progressPercent,
      thumbnailDataUrl: incoming.thumbnailDataUrl ?? existing.thumbnailDataUrl,
      localPath: incoming.localPath ?? existing.localPath,
      downloadedAt: incoming.downloadedAt ?? existing.downloadedAt,
    };
  }

  function applyFileTransferEvent(event: FileTransferEvent) {
    const current = storedItems.value.find((item) => item.id === event.itemId);
    if (!current || current.itemKind !== "file") {
      return;
    }

    const next = normalizeClipboardItemRecord({
      ...current,
      fileMeta: {
        ...(current.fileMeta ?? defaultFileMeta()),
        transferState: event.status,
        progressPercent: event.progressPercent,
        localPath: current.fileMeta?.localPath ?? null,
        downloadedAt:
          event.status === "received"
            ? current.fileMeta?.downloadedAt ?? Date.now()
            : current.fileMeta?.downloadedAt ?? null,
        thumbnailDataUrl: current.fileMeta?.thumbnailDataUrl ?? null,
        fileName: current.fileMeta?.fileName ?? "",
        extension: current.fileMeta?.extension ?? "",
        mimeType: current.fileMeta?.mimeType ?? "",
        sizeBytes: current.fileMeta?.sizeBytes ?? event.bytesTotal,
      },
    });
    upsertItem(next);
  }

  function markFileTransfer(
    itemId: string,
    status: ClipboardTransferState,
    progressPercent: number,
    errorMessage: string | null,
  ) {
    const current = storedItems.value.find((item) => item.id === itemId);
    if (!current || current.itemKind !== "file") {
      return;
    }

    const next = normalizeClipboardItemRecord({
      ...current,
      fileMeta: {
        ...(current.fileMeta ?? defaultFileMeta()),
        transferState: status,
        progressPercent,
        localPath: current.fileMeta?.localPath ?? null,
        downloadedAt: current.fileMeta?.downloadedAt ?? null,
        thumbnailDataUrl: current.fileMeta?.thumbnailDataUrl ?? null,
        fileName: current.fileMeta?.fileName ?? "",
        extension: current.fileMeta?.extension ?? "",
        mimeType: current.fileMeta?.mimeType ?? "",
        sizeBytes: current.fileMeta?.sizeBytes ?? 0,
      },
    });
    upsertItem(next);
    if (errorMessage) {
      message.error(errorMessage);
    }
  }

  function toggleItemPin(itemId: string) {
    storedItems.value = storedItems.value.map((item) =>
      item.id === itemId ? { ...item, pinned: !item.pinned, updatedAt: Date.now() } : item,
    );
  }

  function deleteItem(itemId: string) {
    void fileForge.removeItem(itemId);
    storedItems.value = storedItems.value.filter((item) => item.id !== itemId);
    if (selectedId.value === itemId) {
      selectedId.value = storedItems.value[0]?.id ?? null;
    }
  }

  function syncSelectedItem() {
    if (!selectedId.value || !storedItems.value.some((item) => item.id === selectedId.value)) {
      selectedId.value = storedItems.value.find((item) => item.isCurrent)?.id ?? storedItems.value[0]?.id ?? null;
    }
  }

  async function handleRefreshList() {
    await refreshPresence();
    queuePendingFileTransfers();
    syncSelectedItem();
    message.success("Refreshed");
  }

  async function handleClearHistory() {
    if (!window.confirm("Clear all local clipboard history?")) {
      return;
    }

    storedItems.value = [];
    selectedId.value = null;
    await fileForge.clear();
    message.success("History cleared");
  }

  async function requestDeviceAssociation() {
    pendingPairRequestID.value = "";
    authState.value = "requesting";
    try {
      const response = await anonymousWorkbenchApi.createPairRequest(
        window.location.origin,
        getOrCreateLocalDeviceID(),
        buildWebDeviceName(),
      );
      pendingPairRequestID.value = response.request.id;
      authState.value = "waiting-approval";
      startPairRequestPolling();
    } catch (error) {
      pendingPairRequestID.value = "";
      authState.value = "expired";
      message.error(describeError(error));
    }
  }

  async function pollPairRequestStatus() {
    if (!pendingPairRequestID.value) {
      return;
    }

    try {
      const response = await anonymousWorkbenchApi.getPairRequest(window.location.origin, pendingPairRequestID.value);
      const request = response.request;
      if (request.status === "pending") {
        return;
      }

      stopPairRequestPolling();
      pendingPairRequestID.value = "";
      await handlePairRequestResolution(request);
    } catch (error) {
      stopPairRequestPolling();
      pendingPairRequestID.value = "";
      authState.value = "expired";
      message.error(describeError(error));
    }
  }

  async function handlePairRequestResolution(request: PairRequestStatus) {
    if (request.status === "approved" && request.credential.trim()) {
      if (await initializeAuthorizedSession(request.credential, true)) {
        await replaceCredentialInUrl(request.credential);
        return;
      }
    }

    clearCredential();
    pendingPairRequestID.value = "";
    authState.value = "expired";
    await replaceCredentialInUrl(null);
  }

  async function renewDeviceAssociation() {
    if (!api.value) {
      return false;
    }

    try {
      session.value = await api.value.renewSession();
      startClock();
      message.success("设备关联已延期");
      return true;
    } catch (error) {
      if (error instanceof WorkbenchApiError && error.status === 401) {
        expireAssociation();
        return false;
      }
      message.error(describeError(error));
      return false;
    }
  }

  async function replaceCredentialInUrl(token: string | null) {
    const nextQuery = { ...route.query };
    const current = typeof nextQuery.token === "string" ? nextQuery.token.trim() : "";
    if (token) {
      if (current === token) {
        return;
      }
      nextQuery.token = token;
    } else {
      if (!current) {
        return;
      }
      delete nextQuery.token;
    }
    await router.replace({
      path: route.path,
      query: nextQuery,
    });
  }

  function persistCredential(token: string) {
    window.localStorage.setItem(webCredentialStorageKey, token.trim());
  }

  function readStoredCredential() {
    return window.localStorage.getItem(webCredentialStorageKey)?.trim() ?? "";
  }

  function clearCredential() {
    window.localStorage.removeItem(webCredentialStorageKey);
  }

  function triggerFileDownload(url: string, fileName: string) {
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = fileName;
    anchor.rel = "noopener";
    anchor.click();
  }

  function triggerBlobDownload(blob: Blob, fileName: string) {
    const url = URL.createObjectURL(blob);
    triggerFileDownload(url, fileName);
    window.setTimeout(() => {
      URL.revokeObjectURL(url);
    }, 0);
  }

  function queuePendingFileTransfers() {
    for (const item of storedItems.value) {
      if (item.itemKind === "file" && item.fileMeta?.transferState === "receiving") {
        void ensureFileAvailable(item);
      }
    }
  }

  async function ensureFileAvailable(item: ClipboardItemRecord) {
    if (!api.value || item.itemKind !== "file") {
      return null;
    }

    if (activeFileTransfers.has(item.id)) {
      return activeFileTransfers.get(item.id)!;
    }

    const cached = await loadFileBlob(item.id);
    if (cached) {
      const ready = markFileCached(item, item.fileMeta?.fileName);
      upsertItem(ready);
      return ready;
    }

    const task = downloadFileToCache(item).finally(() => {
      activeFileTransfers.delete(item.id);
    });
    activeFileTransfers.set(item.id, task);
    return task;
  }

  async function downloadFileToCache(item: ClipboardItemRecord) {
    if (!api.value || item.itemKind !== "file") {
      return null;
    }

    markFileTransfer(item.id, "receiving", 0, null);

    try {
      const response = await fetch(api.value.fileContentUrl(item.id));
      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const blob = await readResponseBlobWithProgress(response, item);
      await storeFileBlob(item.id, blob);
      const current = storedItems.value.find((entry) => entry.id === item.id) ?? item;
      const ready = markFileCached(current, current.fileMeta?.fileName);
      upsertItem(ready);
      return ready;
    } catch (error) {
      markFileTransfer(item.id, "failed", 0, describeError(error));
      return null;
    }
  }

  async function readResponseBlobWithProgress(response: Response, item: ClipboardItemRecord) {
    const body = response.body;
    if (!body) {
      return response.blob();
    }

    const reader = body.getReader();
    const chunks: BlobPart[] = [];
    const total = Number(response.headers.get("Content-Length") ?? item.fileMeta?.sizeBytes ?? 0);
    let transferred = 0;

    for (;;) {
      const { done, value } = await reader.read();
      if (done) {
        break;
      }
      if (!value) {
        continue;
      }
      const chunk = new Uint8Array(value.byteLength);
      chunk.set(value);
      chunks.push(chunk);
      transferred += value.byteLength;
      markFileTransfer(item.id, "receiving", clampPercent(progressPercent(transferred, total)), null);
    }

    return new Blob(chunks, {
      type: item.fileMeta?.mimeType || response.headers.get("Content-Type") || "application/octet-stream",
    });
  }

  async function loadFileBlob(itemId: string) {
    return (await fileForge.getItem<Blob>(itemId)) ?? null;
  }

  async function storeFileBlob(itemId: string, blob: Blob) {
    await fileForge.setItem(itemId, blob);
  }

  function markFileCached(item: ClipboardItemRecord, localPath: string | null | undefined) {
    return normalizeClipboardItemRecord({
      ...item,
      fileMeta: {
        ...(item.fileMeta ?? defaultFileMeta()),
        transferState: "received",
        progressPercent: 100,
        localPath: localPath ?? item.fileMeta?.localPath ?? null,
        downloadedAt: Date.now(),
        thumbnailDataUrl: item.fileMeta?.thumbnailDataUrl ?? null,
        fileName: item.fileMeta?.fileName ?? "",
        extension: item.fileMeta?.extension ?? "",
        mimeType: item.fileMeta?.mimeType ?? "",
        sizeBytes: item.fileMeta?.sizeBytes ?? 0,
      },
    });
  }

  return {
    activateItem,
    authState,
    closeContextMenu,
    contextMenuItem,
    contextMenuOpen,
    contextMenuOptions,
    contextMenuX,
    contextMenuY,
    deleteItem,
    expiresIn,
    handleMoreSelect,
    handleContextSelect,
    handleRowClick,
    initializing,
    items,
    localDevice,
    markFileTransfer,
    moreOptions,
    onlineDevices,
    openContextMenu,
    refreshPresence,
    renewAvailable,
    renewDeviceAssociation,
    requestDeviceAssociation,
    search,
    selectedId,
    session,
    submitTextDraft,
    syncClipboardItem,
    toggleItemPin,
    uploadFile,
  };
}

function normalizeClipboardItemRecord(item: ClipboardItemRecord): ClipboardItemRecord {
  const fileMeta = item.fileMeta ? normalizeFileMeta(item.fileMeta) : null;
  const itemKind = normalizeItemKind(item.itemKind, fileMeta);

  if (itemKind === "file") {
    return {
      ...item,
      itemKind,
      content: "",
      contentType: item.contentType || fileContentType(fileMeta),
      charCount: 0,
      preview: item.preview || fileMeta?.fileName || "",
      fileMeta,
    };
  }

  return {
    ...item,
    itemKind,
    contentType: item.contentType || "text/plain",
    charCount: item.charCount || [...item.content].length,
    preview: item.preview || buildPreview(item.content),
    fileMeta: null,
  };
}

function normalizeFileMeta(meta: ClipboardFileMeta): ClipboardFileMeta {
  return {
    ...meta,
    fileName: meta.fileName.trim(),
    extension: meta.extension.trim(),
    mimeType: meta.mimeType.trim(),
    thumbnailDataUrl: meta.thumbnailDataUrl ?? null,
    transferState: normalizeTransferState(meta.transferState),
    progressPercent: clampPercent(meta.progressPercent),
    localPath: meta.localPath ?? null,
    downloadedAt: meta.downloadedAt ?? null,
  };
}

function normalizeTransferState(value: string): ClipboardTransferState {
  if (value === "receiving" || value === "received" || value === "failed") {
    return value;
  }
  return "metadata_only";
}

function normalizeItemKind(value: string, fileMeta: ClipboardFileMeta | null) {
  if (value === "file" || fileMeta) {
    return "file";
  }
  return "text";
}

function fileContentType(meta: ClipboardFileMeta | null) {
  if (meta?.mimeType) {
    return meta.mimeType;
  }
  return "application/octet-stream";
}

function defaultFileMeta(): ClipboardFileMeta {
  return {
    fileName: "",
    extension: "",
    mimeType: "",
    sizeBytes: 0,
    thumbnailDataUrl: null,
    transferState: "metadata_only",
    progressPercent: 0,
    localPath: null,
    downloadedAt: null,
  };
}

function cloneFileMeta(meta: ClipboardFileMeta | null) {
  if (!meta) {
    return null;
  }
  return { ...meta };
}

function mergeClipboardItems(current: ClipboardItemRecord[], next: ClipboardItemRecord[]) {
  const map = new Map<string, ClipboardItemRecord>();
  for (const item of current) {
    map.set(item.id, item);
  }
  for (const item of next) {
    const existing = map.get(item.id);
    map.set(item.id, existing ? mergeClipboardRecord(existing, item) : item);
  }
  return sortClipboardItems([...map.values()]);
}

function mergeClipboardRecord(existing: ClipboardItemRecord, incoming: ClipboardItemRecord) {
  const merged = normalizeClipboardItemRecord({
    ...existing,
    ...incoming,
    content: incoming.content !== "" ? incoming.content : existing.content,
    fileMeta: mergeFileMeta(existing.fileMeta, incoming.fileMeta),
  });
  if (merged.itemKind === "file") {
    merged.content = "";
    merged.charCount = 0;
  }
  return merged;
}

function mergeFileMeta(existing: ClipboardFileMeta | null, incoming: ClipboardFileMeta | null) {
  if (!existing && !incoming) {
    return null;
  }
  if (!existing) {
    return cloneFileMeta(incoming);
  }
  if (!incoming) {
    return cloneFileMeta(existing);
  }

  return normalizeFileMeta({
    ...existing,
    ...incoming,
    thumbnailDataUrl: incoming.thumbnailDataUrl ?? existing.thumbnailDataUrl,
    localPath: incoming.localPath ?? existing.localPath,
    downloadedAt: incoming.downloadedAt ?? existing.downloadedAt,
  });
}

function sortClipboardItems(items: ClipboardItemRecord[]) {
  return [...items].sort((left, right) => {
    if (left.pinned !== right.pinned) {
      return left.pinned ? -1 : 1;
    }
    if (left.createdAt !== right.createdAt) {
      return right.createdAt - left.createdAt;
    }
    return right.id.localeCompare(left.id);
  });
}

function hashText(value: string) {
  const bytes = new TextEncoder().encode(value);
  let hash = 0;
  for (const byte of bytes) {
    hash = (hash << 5) - hash + byte;
    hash |= 0;
  }
  return `${hash}`;
}

function buildPreview(content: string) {
  const normalized = content.split(/\s+/).filter(Boolean).join(" ");
  const limit = 120;
  return normalized.length <= limit ? normalized : `${normalized.slice(0, limit)}...`;
}

async function writeFileBlobToClipboard(blob: Blob, mimeType: string) {
  if (typeof window.ClipboardItem !== "function" || typeof navigator.clipboard?.write !== "function") {
    return false;
  }

  const type = mimeType || blob.type || "application/octet-stream";
  try {
    await navigator.clipboard.write([
      new window.ClipboardItem({
        [type]: blob,
      }),
    ]);
    return true;
  } catch {
    return false;
  }
}

function resolveTransferState(current: ClipboardTransferState, incoming: ClipboardTransferState) {
  const rank: Record<ClipboardTransferState, number> = {
    metadata_only: 0,
    failed: 1,
    receiving: 2,
    received: 3,
  };
  return rank[current] >= rank[incoming] ? current : incoming;
}

function progressPercent(transferred: number, total: number) {
  if (total <= 0) {
    return transferred > 0 ? 100 : 0;
  }
  return Math.floor((transferred * 100) / total);
}

function clampPercent(value: number) {
  if (value < 0) {
    return 0;
  }
  if (value > 100) {
    return 100;
  }
  return value;
}

function formatCountdown(expiresAt: number | null, now: number) {
  if (!expiresAt) {
    return "Waiting";
  }

  const remaining = expiresAt - now;
  if (remaining <= 0) {
    return "Expired";
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

function buildWebDeviceName() {
  const parser = new UAParser(window.navigator.userAgent);
  const browser = parser.getBrowser().name?.trim() || "Browser";
  const os = parser.getOS().name?.trim() || "OS";
  return `${browser}-${os}`;
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
