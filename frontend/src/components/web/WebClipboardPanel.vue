<template>
  <section class="panel-surface rounded-[26px] p-4 sm:p-[1rem]">
    <div class="flex items-center gap-[0.55rem] border-b border-[rgba(20,33,27,0.08)] pb-[0.65rem]">
      <n-input
        :value="search"
        clearable
        placeholder="Search text or filename"
        class="min-w-0 flex-1 [&_.n-input-wrapper]:!rounded-[10px]"
        @update:value="emit('update:search', $event)"
      >
        <template #prefix>
          <n-icon><Search class="h-[17px] w-[17px]" /></n-icon>
        </template>
      </n-input>

      <n-dropdown
        trigger="click"
        :size="webDropdownSize"
        :theme-overrides="webDropdownThemeOverrides"
        :options="moreOptions"
        @select="emit('more-select', $event)"
      >
        <n-button quaternary circle class="!rounded-[10px]">
          <template #icon>
            <n-icon><MoreIcon class="h-[18px] w-[18px]" /></n-icon>
          </template>
        </n-button>
      </n-dropdown>
    </div>

    <div class="mt-[0.2rem] min-h-[22rem] overflow-auto md:max-h-[calc(100vh-18rem)]">
      <div v-if="initializing" class="flex items-center gap-[0.65rem] px-[0.35rem] py-4 text-[var(--text-muted)]">
        <n-spin size="small" />
        <span>Connecting...</span>
      </div>

      <n-empty v-else-if="items.length === 0" description="No clipboard entries" class="py-8" />

      <template v-else>
        <div
          v-for="item in items"
          :key="item.id"
          role="button"
          tabindex="0"
          class="w-full cursor-pointer border-0 border-b border-[rgba(20,33,27,0.08)] bg-transparent px-[0.75rem] py-[0.82rem] text-left text-inherit transition-colors duration-150 hover:bg-[rgba(31,122,90,0.05)]"
          :class="[
            item.id === selectedId || item.isCurrent ? 'bg-[rgba(31,122,90,0.07)]' : '',
            item.isCurrent ? 'shadow-[inset_2px_0_0_#1f7a5a]' : '',
          ]"
          @click="emit('row-click', item)"
          @contextmenu.prevent.stop="emit('row-contextmenu', { event: $event, item })"
        >
          <div v-if="item.itemKind === 'file'" class="flex min-w-0 gap-3">
            <div
              class="flex h-[3.55rem] w-[3.55rem] shrink-0 overflow-hidden rounded-[14px] border border-[rgba(20,33,27,0.08)] bg-[rgba(255,252,247,0.9)]"
            >
              <img
                v-if="item.fileMeta?.thumbnailDataUrl"
                :src="item.fileMeta.thumbnailDataUrl"
                :alt="item.fileMeta.fileName"
                class="h-full w-full object-cover"
              />
              <div v-else class="flex h-full w-full items-center justify-center">
                <n-icon size="22">
                  <component :is="fileIcon(item)" class="h-[22px] w-[22px]" />
                </n-icon>
              </div>
            </div>

            <div class="min-w-0 flex-1">
              <div class="flex items-start justify-between gap-2">
                <div class="min-w-0">
                  <p class="m-0 truncate text-[0.92rem] font-semibold text-[var(--text-main)]">
                    {{ item.fileMeta?.fileName || item.preview || "Untitled file" }}
                  </p>
                  <div class="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-[0.77rem] text-[var(--text-muted)]">
                    <span>{{ formatSource(item.sourceKind) }}</span>
                    <span>{{ formatDateTime(item.createdAt) }}</span>
                    <span>{{ formatFileSize(item.fileMeta?.sizeBytes ?? 0) }}</span>
                    <span>{{ formatTransferState(item.fileMeta?.transferState ?? "metadata_only") }}</span>
                    <span v-if="item.fileMeta?.extension">{{ item.fileMeta.extension }}</span>
                  </div>
                </div>

                <div class="flex shrink-0 items-center gap-1.5">
                  <n-icon v-if="item.pinned" size="14" class="text-[#9a691b]">
                    <Pin class="h-[14px] w-[14px]" />
                  </n-icon>

                  <n-button
                    quaternary
                    circle
                    class="!h-10 !w-10 !min-w-[2.5rem] !rounded-[12px] text-[var(--text-muted)] hover:!text-[var(--text-main)]"
                    @click.stop="emit('row-contextmenu', { event: $event, item })"
                  >
                    <template #icon>
                      <n-icon><MoreIcon class="h-[18px] w-[18px]" /></n-icon>
                    </template>
                  </n-button>
                </div>
              </div>

              <div v-if="item.fileMeta?.transferState === 'receiving'" class="mt-2">
                <div class="h-1.5 overflow-hidden rounded-full bg-[rgba(20,33,27,0.08)]">
                  <div
                    class="h-full rounded-full bg-[#1f7a5a] transition-[width] duration-150"
                    :style="{ width: `${item.fileMeta.progressPercent}%` }"
                  />
                </div>
                <div class="mt-1 text-[0.74rem] text-[var(--text-muted)]">
                  {{ item.fileMeta.progressPercent }}%
                </div>
              </div>
            </div>
          </div>

          <div v-else class="grid min-w-0 gap-[0.36rem]">
            <div class="flex items-start justify-between gap-3">
              <div class="flex min-w-0 flex-1 flex-wrap items-center gap-[0.45rem] pr-1 text-[0.77rem] text-[var(--text-muted)]">
                <span>{{ formatSource(item.sourceKind) }}</span>
                <span>{{ formatDateTime(item.createdAt) }}</span>
                <span>{{ item.charCount }} chars</span>
                <n-icon v-if="item.pinned" size="14" class="text-[#9a691b]">
                  <Pin class="h-[14px] w-[14px]" />
                </n-icon>
              </div>

              <n-button
                quaternary
                circle
                class="!mt-[-0.35rem] !h-10 !w-10 !min-w-[2.5rem] shrink-0 !rounded-[12px] text-[var(--text-muted)] hover:!text-[var(--text-main)]"
                @click.stop="emit('row-contextmenu', { event: $event, item })"
              >
                <template #icon>
                  <n-icon><MoreIcon class="h-[18px] w-[18px]" /></n-icon>
                </template>
              </n-button>
            </div>

            <p
              class="m-0 overflow-hidden text-[var(--text-main)] leading-[1.58] [display:-webkit-box] [-webkit-box-orient:vertical] [-webkit-line-clamp:2] [word-break:break-word]"
            >
              {{ item.preview || "Empty clipboard entry" }}
            </p>
          </div>
        </div>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { File, FileArchive, FileImage, FileText, Pin, Search } from "lucide-vue-next";
import { NButton, NDropdown, NEmpty, NIcon, NInput, NSpin, type DropdownOption } from "naive-ui";

import { formatDateTime, formatSource } from "../../app/formatters";
import type { ClipboardItemRecord } from "../../types/workbench";
import { MoreIcon } from "../../utils/desktopIcons";
import { webDropdownSize, webDropdownThemeOverrides } from "../../utils/dropdown";

defineProps<{
  initializing: boolean;
  items: ClipboardItemRecord[];
  moreOptions: DropdownOption[];
  search: string;
  selectedId: string | null;
}>();

const emit = defineEmits<{
  (e: "more-select", key: string | number): void;
  (e: "row-click", item: ClipboardItemRecord): void;
  (e: "row-contextmenu", payload: { event: MouseEvent; item: ClipboardItemRecord }): void;
  (e: "update:search", value: string): void;
}>();

function fileIcon(item: ClipboardItemRecord) {
  const mime = item.fileMeta?.mimeType.toLowerCase() ?? "";
  const extension = item.fileMeta?.extension.toLowerCase() ?? "";
  if (mime.startsWith("image/") || ["png", "jpg", "jpeg", "gif", "webp", "bmp", "svg"].includes(extension)) {
    return FileImage;
  }
  if (
    mime.startsWith("text/") ||
    ["txt", "md", "json", "csv", "xml", "log", "yaml", "yml"].includes(extension)
  ) {
    return FileText;
  }
  if (["zip", "rar", "7z", "gz", "tar"].includes(extension)) {
    return FileArchive;
  }
  return File;
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

function formatTransferState(value: string) {
  if (value === "receiving") {
    return "Receiving";
  }
  if (value === "received") {
    return "Received";
  }
  if (value === "failed") {
    return "Failed";
  }
  return "Pending";
}
</script>
