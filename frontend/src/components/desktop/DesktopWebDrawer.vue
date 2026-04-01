<template>
  <n-drawer :show="show" placement="left" :width="220" @update:show="emit('update:show', $event)">
    <div class="grid min-h-full content-start gap-3 bg-[#f5f2ea] p-3">

      <div class="relative rounded-xl border border-[rgba(20,33,27,0.08)] bg-[rgba(255,255,255,0.84)]">
        <template v-if="resolvedSessionUrl">
          <n-qr-code :value="resolvedSessionUrl" :size="150" color="#18352c" background-color="#f7f5ef" />
          <n-icon class="absolute! right-1 top-1 cursor-pointer" @click="emit('refresh')">
            <RotateIcon class="h-3! w-3!" />
          </n-icon>
        </template>

        <span v-else class="text-[0.82rem] text-[var(--text-muted)]">链接不可用</span>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-2 text-[0.78rem] text-[var(--text-main)]">
        <span class="text-[0.76rem] text-[var(--text-muted)]">{{ activeHost || "--" }}</span>
      </div>

      <div class="flex items-center gap-[0.32rem]">
        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button quaternary circle class="!rounded-[10px]" @click="emit('copy')">
              <template #icon>
                <n-icon>
                  <CopyIcon class="h-[18px] w-[18px]" />
                </n-icon>
              </template>
            </n-button>
          </template>
          复制链接
        </n-tooltip>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button quaternary circle class="!rounded-[10px]" @click="emit('open')">
              <template #icon>
                <n-icon>
                  <OpenIcon class="h-[18px] w-[18px]" />
                </n-icon>
              </template>
            </n-button>
          </template>
          打开链接
        </n-tooltip>

        <n-dropdown trigger="click" :options="candidateOptions" @select="emit('select-host', $event)">
          <n-button quaternary circle class="!rounded-[10px]">
            <template #icon>
              <n-icon>
                <RouteIcon class="h-[18px] w-[18px]" />
              </n-icon>
            </template>
          </n-button>
        </n-dropdown>

        <n-tooltip trigger="hover">
          <template #trigger>
            <n-button quaternary circle class="!rounded-[10px]" @click="emit('diagnostics')">
              <template #icon>
                <n-icon>
                  <DiagnosticsIcon class="h-[18px] w-[18px]" />
                </n-icon>
              </template>
            </n-button>
          </template>
          诊断
        </n-tooltip>
      </div>


    </div>
  </n-drawer>
</template>

<script setup lang="ts">
import { NButton, NDrawer, NDropdown, NIcon, NQrCode, NTooltip, type DropdownOption } from "naive-ui";

import {
  CopyIcon,
  DiagnosticsIcon,
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
  (e: "refresh"): void;
  (e: "select-host", key: string | number): void;
  (e: "update:show", value: boolean): void;
}>();
</script>
