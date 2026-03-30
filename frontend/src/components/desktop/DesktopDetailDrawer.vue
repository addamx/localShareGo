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

      <n-empty v-else description="无内容" />
    </div>
  </n-drawer>
</template>

<script setup lang="ts">
import { NButton, NDrawer, NEmpty, NIcon, NTooltip } from "naive-ui";

import { formatDateTime, formatSource } from "../../app/formatters";
import type { ClipboardItemRecord } from "../../types/workbench";
import { CloseIcon, CopyIcon, DeleteIcon, EyeIcon } from "../../utils/desktopIcons";

defineProps<{
  item: ClipboardItemRecord | null;
  show: boolean;
}>();

const emit = defineEmits<{
  (e: "copy"): void;
  (e: "delete"): void;
  (e: "update:show", value: boolean): void;
}>();
</script>
