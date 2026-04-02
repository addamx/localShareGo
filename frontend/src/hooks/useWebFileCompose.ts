import { computed, ref, watch, type ComputedRef } from "vue";
import { useMessage, type DropdownOption } from "naive-ui";

import { useForgeStorage } from "./useForgeStorage";
import type { OnlineDevice } from "../types/workbench";

const webComposeSyncTargetsStorageKey = "localsharego:web:compose-sync-targets";

interface UseWebFileComposeOptions {
  maxTextBytes: ComputedRef<number>;
  onlineDevices: ComputedRef<OnlineDevice[]>;
  submitTextDraft(content: string, targetDeviceIds: string[], syncAll: boolean): Promise<boolean>;
  uploadFile(file: File, targetDeviceIds: string[]): Promise<boolean>;
}

export function useWebFileCompose(options: UseWebFileComposeOptions) {
  const message = useMessage();

  const draft = ref("");
  const uploading = ref(false);
  const pendingFile = ref<File | null>(null);
  const pendingFileLabel = ref("");
  const pendingFileSizeLabel = ref("");

  const storedComposeSyncTargetIDs = useForgeStorage<string[] | null>(webComposeSyncTargetsStorageKey, null);

  const draftBytes = computed(() => new TextEncoder().encode(draft.value).length);
  const selectedComposeDeviceIDs = computed(() => normalizeComposeSyncTargets(storedComposeSyncTargetIDs.value));
  const selectedComposeDevices = computed(() =>
    options.onlineDevices.value.filter((device) => selectedComposeDeviceIDs.value.includes(device.id)),
  );
  const syncAvailable = computed(() => options.onlineDevices.value.length > 0);
  const composeSyncOptions = computed<DropdownOption[]>(() => {
    if (options.onlineDevices.value.length === 0) {
      return [{ label: "No online devices", key: "compose-sync:none", disabled: true }];
    }

    return options.onlineDevices.value.map((device) => ({
      label: selectedComposeDeviceIDs.value.includes(device.id) ? `✓ ${device.name}` : device.name,
      key: device.id,
    }));
  });
  const composeSyncLabel = computed(() => {
    if (options.onlineDevices.value.length === 0) {
      return "No devices";
    }

    const devices = selectedComposeDevices.value;
    if (devices.length === options.onlineDevices.value.length) {
      return "Sync to all";
    }
    if (devices.length === 1) {
      return devices[0].name;
    }
    return `${devices[0].name} +${devices.length - 1}`;
  });

  watch(
    () => options.onlineDevices.value,
    () => {
      syncComposeTargets();
    },
    { immediate: true, deep: true },
  );

  async function submitDraft() {
    const file = pendingFile.value;
    if (file) {
      await uploadPendingFile(file);
      return;
    }

    const content = draft.value.trim();
    if (!content) {
      message.warning("Type some text first");
      return;
    }
    if (draftBytes.value > options.maxTextBytes.value) {
      message.warning("Text exceeds the allowed size");
      return;
    }

    const selectedTargets = selectedComposeDeviceIDs.value;
    const ok = await options.submitTextDraft(content, selectedTargets, false);
    if (ok) {
      draft.value = "";
    }
  }

  function clearCompose() {
    draft.value = "";
    pendingFile.value = null;
    pendingFileLabel.value = "";
    pendingFileSizeLabel.value = "";
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

  function syncComposeTargets() {
    if (options.onlineDevices.value.length === 0) {
      if (storedComposeSyncTargetIDs.value === null) {
        storedComposeSyncTargetIDs.value = [];
      }
      return;
    }

    const onlineDeviceIDs = new Set(options.onlineDevices.value.map((device) => device.id));
    const currentTargets = normalizeComposeSyncTargets(storedComposeSyncTargetIDs.value).filter((deviceID) =>
      onlineDeviceIDs.has(deviceID),
    );
    if (storedComposeSyncTargetIDs.value === null || currentTargets.length === 0) {
      storedComposeSyncTargetIDs.value = options.onlineDevices.value.map((device) => device.id);
      return;
    }
    storedComposeSyncTargetIDs.value = currentTargets;
  }

  async function handleFilesSelected(files: File[] | FileList | null | undefined) {
    const normalizedFiles = toFiles(files);
    if (normalizedFiles.length === 0) {
      return;
    }
    if (normalizedFiles.length > 1) {
      message.warning("Only one file can be sent at a time");
    }

    const file = normalizedFiles[0] ?? null;
    if (!file) {
      return;
    }

    pendingFile.value = file;
    pendingFileLabel.value = file.name;
    pendingFileSizeLabel.value = formatFileSize(file.size);
  }

  async function handleFilesDropped(files: File[] | FileList | null | undefined) {
    await handleFilesSelected(files);
  }

  async function handleFilesPasted(files: File[] | FileList | null | undefined) {
    await handleFilesSelected(files);
  }

  async function uploadPendingFile(file: File) {
    uploading.value = true;

    try {
      const ok = await options.uploadFile(file, selectedComposeDeviceIDs.value);
      if (!ok) {
        message.error("File upload failed");
        return;
      }

      pendingFile.value = null;
      pendingFileLabel.value = "";
      pendingFileSizeLabel.value = "";
    } finally {
      uploading.value = false;
    }
  }

  return {
    clearCompose,
    composeSyncLabel,
    composeSyncOptions,
    draft,
    draftBytes,
    handleComposeSyncSelect,
    handleFilesDropped,
    handleFilesPasted,
    handleFilesSelected,
    pendingFileLabel,
    pendingFileSizeLabel,
    selectedComposeDeviceIDs,
    syncAvailable,
    submitDraft,
    uploading,
  };
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

function toFiles(files: File[] | FileList | null | undefined) {
  if (!files || files.length === 0) {
    return [];
  }

  if (Array.isArray(files)) {
    return files;
  }

  return Array.from(files);
}

function formatFileSize(bytes: number) {
  if (bytes < 1024) {
    return `${bytes} B`;
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`;
  }
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
