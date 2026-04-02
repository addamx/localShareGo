import type { Ref } from "vue";

import type {
  ClipboardItemRecord,
  ClipboardTransferState,
  FileTransferEvent,
} from "../types/workbench";

const fileStateLabels: Record<ClipboardTransferState, string> = {
  metadata_only: "元数据",
  receiving: "接收中",
  received: "已接收",
  failed: "接收失败",
};

export function isDesktopFileItem(item: ClipboardItemRecord) {
  return item.itemKind === "file";
}

export function isDesktopImageFile(item: ClipboardItemRecord) {
  return Boolean(item.fileMeta?.mimeType.toLowerCase().startsWith("image/"));
}

export function formatDesktopFileSize(bytes: number) {
  if (bytes < 1024) {
    return `${bytes} B`;
  }

  const kilobytes = bytes / 1024;
  if (kilobytes < 1024) {
    return `${kilobytes.toFixed(kilobytes >= 10 ? 1 : 2)} KB`;
  }

  const megabytes = kilobytes / 1024;
  return `${megabytes.toFixed(megabytes >= 10 ? 1 : 2)} MB`;
}

export function formatDesktopTransferState(state: ClipboardTransferState) {
  return fileStateLabels[state];
}

export function patchDesktopFileTransfer(items: Ref<ClipboardItemRecord[]>, event: FileTransferEvent) {
  const now = Date.now();
  items.value = items.value.map((item) => {
    if (item.id !== event.itemId || item.itemKind !== "file" || !item.fileMeta) {
      return item;
    }

    return {
      ...item,
      fileMeta: {
        ...item.fileMeta,
        transferState: event.status,
        progressPercent: event.progressPercent,
      },
      updatedAt: now,
    };
  });
}
