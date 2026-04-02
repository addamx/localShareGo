<template>
  <n-drawer :show="show" placement="right" :width="360" @update:show="emit('update:show', $event)">
    <div class="grid min-h-full content-start gap-3 bg-[#f5f2ea] p-3">
      <div class="flex items-center justify-between gap-3">
        <div
          class="inline-flex items-center gap-2 rounded-full bg-[rgba(31,122,90,0.08)] px-[0.56rem] py-[0.28rem] text-[0.78rem] font-bold uppercase tracking-[0.04em] text-[#1b3f34]"
        >
          <n-icon><EyeIcon class="h-4 w-4" /></n-icon>
          <span>查看</span>
        </div>

        <n-button quaternary circle class="!rounded-[10px]" @click="emit('update:show', false)">
          <template #icon>
            <n-icon><CloseIcon class="h-[18px] w-[18px]" /></n-icon>
          </template>
        </n-button>
      </div>

      <template v-if="item">
        <template v-if="isDesktopFileItem(item)">
          <div class="grid gap-3">
            <div class="overflow-hidden rounded-[14px] border border-[rgba(20,33,27,0.08)] bg-[rgba(255,255,255,0.85)]">
              <div class="aspect-[4/3] bg-[rgba(31,122,90,0.06)]">
                <img
                  v-if="isDesktopImageFile(item) && item.fileMeta?.thumbnailDataUrl"
                  :src="item.fileMeta.thumbnailDataUrl"
                  :alt="item.fileMeta.fileName"
                  class="h-full w-full object-cover"
                />
                <div v-else class="grid h-full w-full place-items-center text-[#1f7a5a]">
                  <n-icon size="44">
                    <component :is="fileIcon(item)" />
                  </n-icon>
                </div>
              </div>
              <div class="border-t border-[rgba(20,33,27,0.08)] px-4 py-3">
                <p class="m-0 truncate text-[0.98rem] font-semibold text-[var(--text-main)]">
                  {{ item.fileMeta?.fileName || item.preview || "Untitled file" }}
                </p>
                <div class="mt-[0.2rem] flex flex-wrap items-center gap-[0.42rem] text-[0.77rem] text-[var(--text-muted)]">
                  <span>{{ formatFileSize(item.fileMeta?.sizeBytes ?? 0) }}</span>
                  <span v-if="item.fileMeta?.extension">{{ item.fileMeta.extension }}</span>
                  <span>{{ formatFileState(item.fileMeta?.transferState) }}</span>
                </div>
              </div>
            </div>

            <dl class="m-0 grid gap-[0.65rem]">
              <div class="grid gap-[0.22rem]">
                <dt class="text-[0.76rem] text-[var(--text-muted)]">来源</dt>
                <dd class="m-0 text-[0.86rem] leading-[1.45] text-[var(--text-main)]">
                  {{ formatSource(item.sourceKind) }}
                </dd>
              </div>
              <div class="grid gap-[0.22rem]">
                <dt class="text-[0.76rem] text-[var(--text-muted)]">时间</dt>
                <dd class="m-0 text-[0.86rem] leading-[1.45] text-[var(--text-main)]">
                  {{ formatDateTime(item.createdAt) }}
                </dd>
              </div>
              <div class="grid gap-[0.22rem]">
                <dt class="text-[0.76rem] text-[var(--text-muted)]">类型</dt>
                <dd class="m-0 text-[0.86rem] leading-[1.45] text-[var(--text-main)]">
                  {{ item.fileMeta?.mimeType || "application/octet-stream" }}
                </dd>
              </div>
            </dl>

            <div v-if="item.fileMeta?.transferState === 'receiving'" class="grid gap-[0.24rem]">
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

            <div class="flex gap-[0.38rem]">
              <n-tooltip v-if="false" trigger="hover">
                <template #trigger>
                  <n-button
                    quaternary
                    circle
                    class="!rounded-[10px]"
                    :disabled="item.fileMeta?.transferState === 'receiving'"
                    @click="emit('receive', item)"
                  >
                    <template #icon>
                      <n-icon>
                        <LoaderCircle
                          v-if="item.fileMeta?.transferState === 'receiving'"
                          class="h-[18px] w-[18px] animate-spin"
                        />
                        <Check v-else-if="item.fileMeta?.transferState === 'received'" class="h-[18px] w-[18px]" />
                        <Download v-else class="h-[18px] w-[18px]" />
                      </n-icon>
                    </template>
                  </n-button>
                </template>
                接收
              </n-tooltip>

              <n-tooltip trigger="hover">
                <template #trigger>
                  <n-button quaternary circle class="!rounded-[10px] text-[#b54d4d]" @click="emit('delete')">
                    <template #icon>
                      <n-icon><DeleteIcon class="h-[18px] w-[18px]" /></n-icon>
                    </template>
                  </n-button>
                </template>
                删除
              </n-tooltip>
            </div>
          </div>
        </template>

        <template v-else>
          <dl class="m-0 grid gap-[0.65rem]">
            <div class="grid gap-[0.22rem]">
              <dt class="text-[0.76rem] text-[var(--text-muted)]">来源</dt>
              <dd class="m-0 text-[0.86rem] leading-[1.45] text-[var(--text-main)]">
                {{ formatSource(item.sourceKind) }}
              </dd>
            </div>
            <div class="grid gap-[0.22rem]">
              <dt class="text-[0.76rem] text-[var(--text-muted)]">时间</dt>
              <dd class="m-0 text-[0.86rem] leading-[1.45] text-[var(--text-main)]">
                {{ formatDateTime(item.createdAt) }}
              </dd>
            </div>
          </dl>

          <pre
            class="m-0 max-h-[calc(100vh-250px)] min-h-[240px] overflow-auto rounded-[10px] border border-[rgba(20,33,27,0.08)] bg-[rgba(255,255,255,0.84)] p-[0.8rem] whitespace-pre-wrap break-words leading-[1.6] text-[var(--text-main)]"
          >{{ item.content }}</pre>

          <div class="flex gap-[0.38rem]">
            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button quaternary circle class="!rounded-[10px]" @click="emit('copy')">
                  <template #icon>
                    <n-icon><CopyIcon class="h-[18px] w-[18px]" /></n-icon>
                  </template>
                </n-button>
              </template>
              复制
            </n-tooltip>

            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button quaternary circle class="!rounded-[10px] text-[#b54d4d]" @click="emit('delete')">
                  <template #icon>
                    <n-icon><DeleteIcon class="h-[18px] w-[18px]" /></n-icon>
                  </template>
                </n-button>
              </template>
              删除
            </n-tooltip>
          </div>
        </template>
      </template>

      <n-empty v-else description="无内容" />
    </div>
  </n-drawer>
</template>

<script setup lang="ts">
import { Check, Download, FileImage, FileText, LoaderCircle } from "lucide-vue-next";
import { NButton, NDrawer, NEmpty, NIcon, NTooltip } from "naive-ui";

import { formatDateTime, formatSource } from "../../app/formatters";
import type { ClipboardItemRecord, ClipboardTransferState } from "../../types/workbench";
import {
  formatDesktopFileSize,
  formatDesktopTransferState,
  isDesktopFileItem,
  isDesktopImageFile,
} from "../../hooks/useDesktopFileTransfers";
import { CloseIcon, CopyIcon, DeleteIcon, EyeIcon } from "../../utils/desktopIcons";

defineProps<{
  item: ClipboardItemRecord | null;
  show: boolean;
}>();

const emit = defineEmits<{
  (e: "copy"): void;
  (e: "delete"): void;
  (e: "receive", item: ClipboardItemRecord): void;
  (e: "update:show", value: boolean): void;
}>();

function fileIcon(item: ClipboardItemRecord) {
  return isDesktopImageFile(item) ? FileImage : FileText;
}

function formatFileSize(bytes: number) {
  return formatDesktopFileSize(bytes);
}

function formatFileState(state: ClipboardTransferState | undefined | null) {
  return formatDesktopTransferState(state ?? "metadata_only");
}
</script>
