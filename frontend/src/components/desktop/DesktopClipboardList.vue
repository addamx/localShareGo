<template>
  <section class="panel-surface flex min-h-0 flex-col rounded-[26px] p-4 sm:p-[1rem]">
    <div class="flex items-center gap-[0.55rem] border-b border-[rgba(20,33,27,0.08)] pb-[0.65rem]">
      <n-input
        :value="search"
        clearable
        placeholder="Search clipboard"
        class="min-w-0 flex-1 [&_.n-input-wrapper]:!rounded-[10px]"
        @update:value="emit('update:search', $event)"
      >
        <template #prefix>
          <n-icon><SearchIcon class="h-[17px] w-[17px]" /></n-icon>
        </template>
      </n-input>

      <n-dropdown trigger="click" :options="moreOptions" @select="emit('more-select', $event)">
        <n-button quaternary circle class="!rounded-[10px]">
          <template #icon>
            <n-icon><MoreIcon class="h-[18px] w-[18px]" /></n-icon>
          </template>
        </n-button>
      </n-dropdown>
    </div>

    <div ref="scroller" class="mt-[0.2rem] min-h-0 flex-1 overflow-auto md:max-h-[calc(100vh-18rem)]">
      <div
        v-if="loading || refreshing"
        class="flex items-center gap-[0.65rem] px-[0.35rem] py-4 text-[var(--text-muted)]"
      >
        <n-spin size="small" />
        <span>正在连接设备...</span>
      </div>

      <n-empty v-else-if="items.length === 0" description="暂无剪贴板内容" class="py-8" />

      <template v-else>
        <button
          v-for="item in items"
          :key="item.id"
          type="button"
          class="flex w-full items-start gap-3 border-0 border-b border-[rgba(20,33,27,0.08)] bg-transparent px-[0.75rem] py-[0.82rem] text-left text-inherit transition-colors duration-150 hover:bg-[rgba(31,122,90,0.05)]"
          :class="[
            item.id === selectedId || item.isCurrent ? 'bg-[rgba(31,122,90,0.07)]' : '',
            item.isCurrent ? 'shadow-[inset_2px_0_0_#1f7a5a]' : '',
          ]"
          @click="emit('row-click', item)"
          @contextmenu.prevent.stop="emit('row-contextmenu', { event: $event, item })"
        >
          <template v-if="isDesktopFileItem(item)">
            <div class="flex min-w-0 flex-1 items-start gap-3">
              <div class="relative h-[56px] w-[56px] shrink-0 overflow-hidden rounded-[14px] border border-[rgba(20,33,27,0.08)] bg-[rgba(31,122,90,0.08)]">
                <img
                  v-if="isDesktopImageFile(item) && item.fileMeta?.thumbnailDataUrl"
                  :src="item.fileMeta.thumbnailDataUrl"
                  :alt="item.fileMeta.fileName"
                  class="h-full w-full object-cover"
                />
                <div v-else class="grid h-full w-full place-items-center text-[#1f7a5a]">
                  <n-icon size="22">
                    <component :is="fileIcon(item)" />
                  </n-icon>
                </div>
              </div>

              <div class="grid min-w-0 flex-1 gap-[0.3rem]">
                <div class="flex items-start justify-between gap-3">
                  <div class="min-w-0">
                    <p class="m-0 truncate text-[0.95rem] font-semibold text-[var(--text-main)]">
                      {{ item.fileMeta?.fileName || item.preview || "Untitled file" }}
                    </p>
                    <div class="mt-[0.12rem] flex flex-wrap items-center gap-[0.42rem] text-[0.77rem] text-[var(--text-muted)]">
                      <span>{{ formatFileSize(item.fileMeta?.sizeBytes ?? 0) }}</span>
                      <span>{{ formatFileState(item.fileMeta?.transferState) }}</span>
                      <span v-if="item.fileMeta?.extension">{{ item.fileMeta.extension }}</span>
                    </div>
                  </div>

                  <n-icon class="text-[var(--text-muted)]">
                    <LoaderCircle
                      v-if="item.fileMeta?.transferState === 'receiving'"
                      class="h-[18px] w-[18px] animate-spin"
                    />
                    <Check v-else-if="item.fileMeta?.transferState === 'received'" class="h-[18px] w-[18px]" />
                    <FileText v-else class="h-[18px] w-[18px]" />
                  </n-icon>
                </div>

                <p
                  class="m-0 overflow-hidden text-[var(--text-main)] leading-[1.5] [display:-webkit-box] [-webkit-box-orient:vertical] [-webkit-line-clamp:2] [word-break:break-word]"
                >
                  {{ item.preview || item.fileMeta?.fileName || "File item" }}
                </p>

                <div v-if="item.fileMeta?.transferState === 'receiving'" class="grid gap-[0.22rem]">
                  <div class="h-[6px] overflow-hidden rounded-full bg-[rgba(20,33,27,0.08)]">
                    <div
                      class="h-full rounded-full bg-[#1f7a5a] transition-[width] duration-150"
                      :style="{ width: `${item.fileMeta?.progressPercent ?? 0}%` }"
                    />
                  </div>
                  <div class="text-[0.74rem] text-[var(--text-muted)]">
                    {{ item.fileMeta?.progressPercent ?? 0 }}%
                  </div>
                </div>
              </div>
            </div>
          </template>

          <template v-else>
            <div class="grid min-w-0 flex-1 gap-[0.36rem]">
              <div class="flex flex-wrap items-center gap-[0.45rem] text-[0.77rem] text-[var(--text-muted)]">
                <span>{{ formatSource(item.sourceKind) }}</span>
                <span>{{ formatDateTime(item.createdAt) }}</span>
                <span>{{ item.charCount }} chars</span>
                <n-icon v-if="item.pinned" size="14" class="text-[#9a691b]">
                  <Pin class="h-[14px] w-[14px]" />
                </n-icon>
              </div>

              <p
                class="m-0 overflow-hidden text-[var(--text-main)] leading-[1.58] [display:-webkit-box] [-webkit-box-orient:vertical] [-webkit-line-clamp:2] [word-break:break-word]"
              >
                {{ item.preview || "Empty clipboard entry" }}
              </p>
            </div>
          </template>
        </button>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { Check, FileImage, FileText, LoaderCircle, Pin } from "lucide-vue-next";
import { ref } from "vue";
import { NButton, NDropdown, NEmpty, NIcon, NInput, NSpin, type DropdownOption } from "naive-ui";

import { formatDateTime, formatSource } from "../../app/formatters";
import type { ClipboardItemRecord, ClipboardTransferState } from "../../types/workbench";
import {
  formatDesktopFileSize,
  formatDesktopTransferState,
  isDesktopFileItem,
  isDesktopImageFile,
} from "../../hooks/useDesktopFileTransfers";
import { MoreIcon, SearchIcon } from "../../utils/desktopIcons";

defineProps<{
  items: ClipboardItemRecord[];
  loading: boolean;
  moreOptions: DropdownOption[];
  refreshing: boolean;
  search: string;
  selectedId: string | null;
}>();

const emit = defineEmits<{
  (e: "more-select", key: string | number): void;
  (e: "row-click", item: ClipboardItemRecord): void;
  (e: "row-contextmenu", payload: { event: MouseEvent; item: ClipboardItemRecord }): void;
  (e: "update:search", value: string): void;
}>();

const scroller = ref<HTMLElement | null>(null);

function fileIcon(item: ClipboardItemRecord) {
  return isDesktopImageFile(item) ? FileImage : FileText;
}

function formatFileSize(bytes: number) {
  return formatDesktopFileSize(bytes);
}

function formatFileState(state: ClipboardTransferState | undefined | null) {
  return formatDesktopTransferState(state ?? "metadata_only");
}

defineExpose({
  getScroller() {
    return scroller.value;
  },
});
</script>
