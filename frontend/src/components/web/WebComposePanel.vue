<template>
  <section class="panel-surface rounded-[26px] p-4 sm:p-[1rem]">
    <div class="mb-3 flex items-start justify-between gap-3">
      <div>
        <p class="panel-eyebrow">Compose</p>
        <h2 class="mt-0.5 text-[1.08rem]">文本</h2>
      </div>
      <span class="text-[0.82rem] leading-none text-[var(--text-muted)]">
        {{ draftBytes }} / {{ maxBytes }}
      </span>
    </div>

    <n-input
      :value="draft"
      type="textarea"
      placeholder="输入文本后，点击发送可保存到当前浏览器并同步到已配置设备"
      :autosize="{ minRows: 8, maxRows: 12 }"
      class="[&_.n-input-wrapper]:!rounded-[16px]"
      @update:value="emit('update:draft', $event)"
    />

    <div class="mt-[0.65rem] flex justify-end gap-1.5">
      <n-button circle secondary @click="emit('clear')">
        <template #icon>
          <n-icon><CloseIcon class="h-[18px] w-[18px]" /></n-icon>
        </template>
      </n-button>

      <n-dropdown trigger="click" placement="bottom-end" :options="syncOptions" @select="emit('sync-select', $event)">
        <n-button secondary class="gap-1.5 !rounded-[999px]" :disabled="!syncAvailable">
          <template #icon>
            <n-icon><Share2 class="h-[16px] w-[16px]" /></n-icon>
          </template>
          {{ syncLabel }}
        </n-button>
      </n-dropdown>

      <n-button circle type="primary" @click="emit('submit')">
        <template #icon>
          <n-icon><SendHorizontal class="h-[18px] w-[18px]" /></n-icon>
        </template>
      </n-button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { SendHorizontal, Share2 } from "lucide-vue-next";
import { NButton, NDropdown, NIcon, NInput, type DropdownOption } from "naive-ui";

import { CloseIcon } from "../../utils/desktopIcons";

defineProps<{
  draft: string;
  draftBytes: number;
  maxBytes: number;
  syncAvailable: boolean;
  syncLabel: string;
  syncOptions: DropdownOption[];
}>();

const emit = defineEmits<{
  (e: "clear"): void;
  (e: "submit"): void;
  (e: "sync-select", key: string | number): void;
  (e: "update:draft", value: string): void;
}>();
</script>
