<template>
  <section
    class="flex min-h-0 flex-col overflow-hidden rounded-[12px] border border-[rgba(20,33,27,0.12)] bg-[rgba(250,248,242,0.88)] shadow-[0_16px_48px_rgba(40,34,19,0.08)] backdrop-blur-[14px]">
    <div class="flex items-center gap-[0.55rem] border-b border-[rgba(20,33,27,0.08)] px-1 py-[0.2rem]">
      <n-input :value="search" clearable placeholder="搜索" class="min-w-0 flex-1 [&_.n-input-wrapper]:!rounded-[10px]"
        @update:value="emit('update:search', $event)">
        <template #prefix>
          <n-icon>
            <SearchIcon class="h-[17px] w-[17px]" />
          </n-icon>
        </template>
      </n-input>

      <n-dropdown trigger="click" :options="moreOptions" @select="emit('more-select', $event)">
        <n-button quaternary circle class="!rounded-[10px]">
          <template #icon>
            <n-icon>
              <MoreIcon class="h-[18px] w-[18px]" />
            </n-icon>
          </template>
        </n-button>
      </n-dropdown>
    </div>

    <div ref="scroller" class="min-h-0 overflow-auto flex-1">
      <div v-if="loading || refreshing"
        class="flex items-center gap-[0.6rem] px-[1.1rem] py-4 text-[var(--text-muted)]">
        <n-spin size="small" />
        <span>同步中</span>
      </div>

      <n-empty v-else-if="items.length === 0" description="暂无记录" class="py-8" />

      <template v-else>
        <div>
          <button v-for="item in items" :key="item.id" type="button"
            class="flex w-full flex-col gap-[0.1rem] border-0 border-b border-[rgba(20,33,27,0.08)] px-[0.2rem] py-[0.5rem] text-left transition-colors duration-150 hover:bg-[rgba(31,122,90,0.06)]"
            :class="[
              item.id === selectedId || item.isCurrent
                ? 'bg-[rgba(31,122,90,0.08)]'
                : 'bg-transparent',
              item.isCurrent ? 'shadow-[inset_2px_0_0_#1f7a5a]' : '',
            ]" @mousedown.prevent @click.stop="emit('row-click', item)"
            @contextmenu.prevent.stop="emit('row-contextmenu', { event: $event, item })">
            <div class="flex flex-wrap items-center gap-[0.45rem] text-[0.77rem] text-[var(--text-muted)]">
              <span>{{ formatSource(item.sourceKind) }}</span>
              <span>{{ formatDateTime(item.createdAt) }}</span>
              <span v-if="item.pinned"
                class="rounded-full bg-[rgba(208,138,36,0.14)] px-[0.35rem] py-[0.08rem] text-[#9a691b]">
                置顶
              </span>
            </div>

            <p
              class="m-0 overflow-hidden text-[var(--text-main)] [display:-webkit-box] [-webkit-box-orient:vertical] [-webkit-line-clamp:2] [word-break:break-word] leading-[1.2]">
              {{ item.preview || item.content }}
            </p>
          </button>
        </div>
      </template>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { NButton, NDropdown, NEmpty, NIcon, NInput, NSpin, type DropdownMixedOption } from "naive-ui";

import { formatDateTime, formatSource } from "../../app/formatters";
import type { ClipboardItemRecord } from "../../types/workbench";
import { MoreIcon, SearchIcon } from "../../utils/desktopIcons";

defineProps<{
  items: ClipboardItemRecord[];
  loading: boolean;
  moreOptions: DropdownMixedOption[];
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

defineExpose({
  getScroller() {
    return scroller.value;
  },
});
</script>
