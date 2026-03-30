<template>
  <header
    class="flex h-9 items-center border border-[rgba(20,33,27,0.12)] bg-[rgba(250,248,242,0.96)] px-1.5 shadow-[0_10px_24px_rgba(40,34,19,0.08)]"
  >
    <nav class="flex items-center gap-0.5">
      <n-button
        quaternary
        class="!h-7 !rounded-[6px] !px-2 text-[var(--text-main)]"
        :disabled="!available"
        @click="emit('open-web')"
      >
        <span class="flex items-center gap-1.5 text-[0.82rem] leading-none">
          <n-icon size="16"><GlobeIcon /></n-icon>
          <span>Web端</span>
        </span>
      </n-button>

      <n-button quaternary class="!h-7 !rounded-[6px] !px-2 text-[var(--text-main)]">
        <span class="flex items-center gap-1.5 text-[0.82rem] leading-none">
          <n-icon size="16"><HelpIcon /></n-icon>
          <span>帮助</span>
        </span>
      </n-button>
    </nav>
  </header>
</template>

<script setup lang="ts">
import { NButton, NIcon } from "naive-ui";

import { GlobeIcon, HelpIcon } from "../../utils/desktopIcons";

defineProps<{
  available: boolean;
}>();

const emit = defineEmits<{
  (e: "open-web"): void;
}>();
</script>
