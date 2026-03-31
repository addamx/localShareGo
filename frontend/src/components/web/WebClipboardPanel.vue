<template>
  <section class="panel-surface rounded-[26px] p-4 sm:p-[1rem]">
    <div class="flex items-center gap-[0.55rem] border-b border-[rgba(20,33,27,0.08)] pb-[0.65rem]">
      <n-input
        :value="search"
        clearable
        placeholder="搜索剪贴板"
        class="min-w-0 flex-1 [&_.n-input-wrapper]:!rounded-[10px]"
        @update:value="emit('update:search', $event)"
      >
        <template #prefix>
          <n-icon><SearchIcon class="h-[17px] w-[17px]" /></n-icon>
        </template>
      </n-input>

      <n-button
        quaternary
        circle
        :class="pinnedOnly ? 'text-[#1f7a5a] bg-[rgba(31,122,90,0.08)]' : ''"
        @click="emit('toggle-pinned')"
      >
        <template #icon>
          <n-icon>
            <Pin v-if="pinnedOnly" class="h-[18px] w-[18px]" />
            <PinOff v-else class="h-[18px] w-[18px]" />
          </n-icon>
        </template>
      </n-button>
    </div>

    <div class="mt-[0.2rem] min-h-[22rem] overflow-auto md:max-h-[calc(100vh-18rem)]">
      <div
        v-if="initializing"
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
          @click="emit('activate', item.id)"
          @contextmenu.prevent.stop="emit('row-contextmenu', { event: $event, item })"
        >
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
        </button>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { Pin, PinOff } from "lucide-vue-next";
import { NButton, NEmpty, NIcon, NInput, NSpin } from "naive-ui";

import { formatDateTime, formatSource } from "../../app/formatters";
import type { ClipboardItemRecord } from "../../types/workbench";
import { SearchIcon } from "../../utils/desktopIcons";

defineProps<{
  initializing: boolean;
  items: ClipboardItemRecord[];
  pinnedOnly: boolean;
  search: string;
  selectedId: string | null;
}>();

const emit = defineEmits<{
  (e: "activate", itemId: string): void;
  (e: "row-contextmenu", payload: { event: MouseEvent; item: ClipboardItemRecord }): void;
  (e: "toggle-pinned"): void;
  (e: "update:search", value: string): void;
}>();
</script>
