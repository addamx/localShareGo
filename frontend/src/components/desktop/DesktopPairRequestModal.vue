<template>
  <n-modal :show="show" @update:show="emit('update:show', $event)">
    <div
      class="grid w-[min(420px,calc(100vw-1.5rem))] gap-4 rounded-[14px] border border-[rgba(20,33,27,0.12)] bg-[#f5f2ea] p-4 shadow-[0_24px_80px_rgba(38,30,15,0.18)]"
      @click.stop
    >
      <div class="grid gap-1">
        <div class="text-[0.88rem] font-bold text-[var(--text-main)]">新的设备关联申请</div>
        <div class="text-[0.78rem] text-[var(--text-muted)]">确认是否允许该浏览器关联当前设备。</div>
      </div>

      <div v-if="request" class="rounded-xl border border-[rgba(20,33,27,0.08)] bg-[rgba(255,255,255,0.84)] px-3 py-3">
        <div class="text-[0.86rem] font-bold text-[var(--text-main)]">{{ request.deviceName }}</div>
        <div class="mt-1 text-[0.74rem] text-[var(--text-muted)]">{{ request.deviceId }}</div>
      </div>

      <div class="flex justify-end gap-2">
        <n-button secondary @click="emit('reject')">拒绝</n-button>
        <n-button type="primary" @click="emit('approve')">同意</n-button>
      </div>
    </div>
  </n-modal>
</template>

<script setup lang="ts">
import { NButton, NModal } from "naive-ui";

import type { PairRequestSummary } from "../../types/workbench";

defineProps<{
  request: PairRequestSummary | null;
  show: boolean;
}>();

const emit = defineEmits<{
  (e: "approve"): void;
  (e: "reject"): void;
  (e: "update:show", value: boolean): void;
}>();
</script>
