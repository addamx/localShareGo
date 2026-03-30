<template>
  <n-drawer :show="show" placement="left" :width="380" @update:show="emit('update:show', $event)">
    <div class="grid min-h-full content-start gap-3 bg-[#f5f2ea] p-3">
      <div class="flex items-center justify-between gap-3">
        <div
          class="inline-flex items-center gap-2 rounded-full bg-[rgba(31,122,90,0.08)] px-[0.56rem] py-[0.28rem] text-[0.78rem] font-bold uppercase tracking-[0.04em] text-[#1b3f34]"
        >
          <n-icon><GlobeIcon class="h-4 w-4" /></n-icon>
          <span>Web</span>
        </div>

        <n-button quaternary circle class="!rounded-[10px]" @click="emit('update:show', false)">
          <template #icon>
            <n-icon><CloseIcon class="h-[18px] w-[18px]" /></n-icon>
          </template>
        </n-button>
      </div>

      <div class="flex items-center gap-3">
        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button secondary circle class="!rounded-[10px]" @click="emit('rotate')">
              <template #icon>
                <n-icon><RotateIcon class="h-[18px] w-[18px]" /></n-icon>
              </template>
            </n-button>
          </template>
          轮换令牌
        </n-tooltip>

        <div class="grid gap-px">
          <span class="text-[0.76rem] text-[var(--text-muted)]">有效期</span>
          <strong class="text-base leading-none text-[var(--text-main)]">{{ tokenCountdown }}</strong>
        </div>
      </div>

      <div
        class="grid min-h-[248px] place-items-center rounded-[12px] border border-[rgba(20,33,27,0.08)] bg-[rgba(255,255,255,0.84)]"
      >
        <n-qr-code
          v-if="resolvedSessionUrl"
          :value="resolvedSessionUrl"
          :size="214"
          color="#18352c"
          background-color="#f7f5ef"
        />
        <span v-else class="text-[0.82rem] text-[var(--text-muted)]">链接不可用</span>
      </div>

      <div
        class="break-all rounded-[10px] border border-[rgba(20,33,27,0.08)] bg-[rgba(255,255,255,0.84)] px-[0.8rem] py-[0.72rem] font-mono text-[0.82rem] leading-[1.55]"
      >
        {{ resolvedSessionUrl || "暂无可用链接" }}
      </div>

      <div class="flex items-center gap-[0.32rem]">
        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button quaternary circle class="!rounded-[10px]" @click="emit('copy')">
              <template #icon>
                <n-icon><CopyIcon class="h-[18px] w-[18px]" /></n-icon>
              </template>
            </n-button>
          </template>
          复制链接
        </n-tooltip>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button quaternary circle class="!rounded-[10px]" @click="emit('open')">
              <template #icon>
                <n-icon><OpenIcon class="h-[18px] w-[18px]" /></n-icon>
              </template>
            </n-button>
          </template>
          打开链接
        </n-tooltip>

        <n-dropdown trigger="click" :options="candidateOptions" @select="emit('select-host', $event)">
          <n-button quaternary circle class="!rounded-[10px]">
            <template #icon>
              <n-icon><RouteIcon class="h-[18px] w-[18px]" /></n-icon>
            </template>
          </n-button>
        </n-dropdown>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button quaternary circle class="!rounded-[10px]" @click="emit('diagnostics')">
              <template #icon>
                <n-icon><DiagnosticsIcon class="h-[18px] w-[18px]" /></n-icon>
              </template>
            </n-button>
          </template>
          诊断
        </n-tooltip>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-2 text-[0.78rem] text-[var(--text-main)]">
        <span class="text-[0.76rem] text-[var(--text-muted)]">{{ activeHost || "--" }}</span>
        <span class="font-mono text-[0.74rem] text-[var(--text-muted)]">{{ bindAddress }}</span>
      </div>
    </div>
  </n-drawer>
</template>

<script setup lang="ts">
import { NButton, NDrawer, NDropdown, NIcon, NQrCode, NTooltip, type DropdownOption } from "naive-ui";

import {
  CloseIcon,
  CopyIcon,
  DiagnosticsIcon,
  GlobeIcon,
  OpenIcon,
  RotateIcon,
  RouteIcon,
} from "../../utils/desktopIcons";

defineProps<{
  activeHost: string;
  bindAddress: string;
  candidateOptions: DropdownOption[];
  resolvedSessionUrl: string;
  show: boolean;
  tokenCountdown: string;
}>();

const emit = defineEmits<{
  (e: "copy"): void;
  (e: "diagnostics"): void;
  (e: "open"): void;
  (e: "rotate"): void;
  (e: "select-host", key: string | number): void;
  (e: "update:show", value: boolean): void;
}>();
</script>
