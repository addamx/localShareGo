<template>
  <header
    class="flex min-h-[54px] items-center justify-between gap-4 rounded-[12px] border border-[rgba(20,33,27,0.12)] bg-[rgba(250,248,242,0.88)] px-3 py-2 shadow-[0_16px_48px_rgba(40,34,19,0.08)] backdrop-blur-[14px]"
  >
    <div class="flex min-w-0 items-center gap-3">
      <div class="grid h-[34px] w-[34px] place-items-center rounded-[10px] bg-[rgba(31,122,90,0.12)] text-[#1b3f34]">
        <ShareLogoIcon class="h-5 w-5" />
      </div>

      <div class="grid min-w-0 gap-px">
        <strong class="text-[0.95rem] font-bold leading-none text-[var(--text-main)]">
          {{ appName || "LocalShareGo" }}
        </strong>
        <span class="truncate text-[0.78rem] leading-none text-[var(--text-muted)]">
          {{ deviceName || "NaiveDesktop" }}
        </span>
      </div>
    </div>

    <div class="flex items-center gap-1">
      <span
        class="h-[7px] w-[7px] rounded-full"
        :class="{
          'bg-[rgba(90,102,94,0.4)]': serverStateTone === 'idle',
          'bg-[#1f7a5a]': serverStateTone === 'running',
          'bg-[#d65f5f]': serverStateTone === 'failed',
        }"
      ></span>

      <n-tooltip trigger="hover">
        <template #trigger>
          <n-button
            quaternary
            circle
            class="!rounded-[10px]"
            :disabled="!available"
            @click="emit('open-web')"
          >
            <template #icon>
              <n-icon><GlobeIcon class="h-[18px] w-[18px]" /></n-icon>
            </template>
          </n-button>
        </template>
        Web
      </n-tooltip>

      <n-tooltip trigger="hover">
        <template #trigger>
          <n-button quaternary circle class="!rounded-[10px]">
            <template #icon>
              <n-icon><HelpIcon class="h-[18px] w-[18px]" /></n-icon>
            </template>
          </n-button>
        </template>
        帮助
      </n-tooltip>
    </div>
  </header>
</template>

<script setup lang="ts">
import { NButton, NIcon, NTooltip } from "naive-ui";

import { GlobeIcon, HelpIcon, ShareLogoIcon } from "../../utils/desktopIcons";

defineProps<{
  appName?: string | null;
  available: boolean;
  deviceName?: string | null;
  serverStateTone: "idle" | "running" | "failed";
}>();

const emit = defineEmits<{
  (e: "open-web"): void;
}>();
</script>
