<template>
  <n-modal :show="show" @update:show="emit('update:show', $event)">
    <div
      class="grid max-h-[calc(100vh-1.5rem)] w-[min(720px,calc(100vw-1.5rem))] content-start gap-3 overflow-auto rounded-[14px] border border-[rgba(20,33,27,0.12)] bg-[#f5f2ea] p-3 shadow-[0_24px_80px_rgba(38,30,15,0.18)]"
      @click.stop
    >
      <div class="flex items-center justify-between gap-3">
        <div
          class="inline-flex items-center gap-2 rounded-full bg-[rgba(31,122,90,0.08)] px-[0.56rem] py-[0.28rem] text-[0.78rem] font-bold uppercase tracking-[0.04em] text-[#1b3f34]"
        >
          <n-icon><DiagnosticsIcon class="h-4 w-4" /></n-icon>
          <span>诊断</span>
        </div>

        <n-button quaternary circle class="!rounded-[10px]" @click="emit('update:show', false)">
          <template #icon>
            <n-icon><CloseIcon class="h-[18px] w-[18px]" /></n-icon>
          </template>
        </n-button>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-2 text-[0.78rem] text-[var(--text-main)]">
        <span>{{ serverState || "--" }}</span>
        <span>{{ bindAddress }}</span>
        <span class="font-mono text-[0.74rem] text-[var(--text-muted)]">{{ resolvedSessionUrl || "--" }}</span>
      </div>

      <div
        v-if="loading"
        class="flex items-center gap-[0.6rem] py-1 text-[var(--text-muted)]"
      >
        <n-spin size="small" />
        <span>检测中</span>
      </div>

      <div v-else-if="connectivity" class="grid gap-[0.55rem]">
        <article
          v-for="check in connectivity.checks"
          :key="check.url"
          class="rounded-[10px] border border-[rgba(20,33,27,0.08)] bg-[rgba(255,255,255,0.84)] px-[0.78rem] py-[0.72rem]"
        >
          <div class="flex flex-wrap items-center gap-[0.45rem] text-[0.78rem]">
            <strong>{{ check.host }}</strong>
            <span>{{ check.tcpOk ? "TCP ok" : "TCP fail" }}</span>
            <span>{{ check.httpOk ? "HTTP ok" : "HTTP fail" }}</span>
          </div>

          <p class="my-[0.35rem] break-all font-mono text-[0.75rem] leading-[1.5] text-[var(--text-muted)]">
            {{ check.url }}
          </p>

          <span class="text-[0.78rem] text-[var(--text-muted)]">
            {{ check.httpStatusLine || check.error || "no response" }}
          </span>
        </article>
      </div>

      <n-empty v-else description="暂无结果" />
    </div>
  </n-modal>
</template>

<script setup lang="ts">
import { NButton, NEmpty, NIcon, NModal, NSpin } from "naive-ui";

import type { ConnectivityReport } from "../../types/workbench";
import { CloseIcon, DiagnosticsIcon } from "../../utils/desktopIcons";

defineProps<{
  bindAddress: string;
  connectivity: ConnectivityReport | null;
  loading: boolean;
  resolvedSessionUrl: string;
  serverState?: string | null;
  show: boolean;
}>();

const emit = defineEmits<{
  (e: "update:show", value: boolean): void;
}>();
</script>
