<template>
  <n-drawer :show="show" placement="left" :width="480" @update:show="emit('update:show', $event)">
    <div class="flex min-h-full gap-1 bg-[#f5f2ea] p-2">
      <section class="flex shrink-0 flex-col gap-3">
        <div class="text-[0.78rem] font-bold text-[var(--text-main)]">关联设备</div>

        <div class="relative rounded-xl border border-[rgba(20,33,27,0.08)] bg-[rgba(255,255,255,0.84)] p-2">
          <template v-if="resolvedSessionUrl">
            <n-qr-code :value="resolvedSessionUrl" :size="150" color="#18352c" background-color="#f7f5ef" />
            <n-icon class="absolute! right-2 top-2 cursor-pointer" @click="emit('refresh')">
              <RotateIcon class="h-3! w-3!" />
            </n-icon>
          </template>

          <span v-else class="text-[0.82rem] text-[var(--text-muted)]">入口不可用</span>
        </div>

        <div class="grid gap-1 rounded-xl border border-[rgba(20,33,27,0.08)] bg-[rgba(255,255,255,0.72)] px-3 py-2 text-[0.76rem]">
          <div class="flex items-center justify-between gap-3">
            <span class="text-[var(--text-muted)]">当前地址</span>
            <span class="truncate text-[var(--text-main)]">{{ activeHost || "--" }}</span>
          </div>
          <div class="flex items-center justify-between gap-3">
            <span class="text-[var(--text-muted)]">监听地址</span>
            <span class="truncate text-[var(--text-main)]">{{ bindAddress }}</span>
          </div>
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
            复制入口
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
            打开入口
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
      </section>

      <section class="grid min-h-0 min-w-0 flex-1 content-start gap-3 rounded-xl border border-[rgba(20,33,27,0.08)] bg-[rgba(255,255,255,0.72)] p-3">
        <div class="flex items-center justify-between gap-2">
          <div class="text-[0.78rem] font-bold text-[var(--text-main)]">已关联设备</div>
          <span class="text-[0.74rem] text-[var(--text-muted)]">{{ linkedDevices.length }} 台</span>
        </div>

        <n-empty v-if="linkedDevices.length === 0" description="暂无已关联设备" class="py-6" />

        <div v-else class="grid gap-2">
          <article
            v-for="device in linkedDevices"
            :key="device.id"
            class="grid grid-cols-[minmax(0,1fr)_auto] items-start gap-2 rounded-xl border border-[rgba(20,33,27,0.08)] bg-[rgba(255,255,255,0.92)] px-3 py-2"
          >
            <div class="min-w-0">
              <div class="truncate text-[0.82rem] font-bold text-[var(--text-main)]">{{ device.name }}</div>
              <div class="mt-1 flex flex-wrap items-center gap-2 text-[0.72rem] text-[var(--text-muted)]">
                <span>{{ device.lastKnownIp || "--" }}</span>
                <span>最近活跃 {{ formatDateTime(device.lastSeenAt || null) }}</span>
              </div>
            </div>

            <n-button quaternary circle class="!rounded-[10px]" @click="emit('remove-device', device.id)">
              <template #icon>
                <n-icon>
                  <DeleteIcon class="h-[16px] w-[16px]" />
                </n-icon>
              </template>
            </n-button>
          </article>
        </div>
      </section>
    </div>
  </n-drawer>
</template>

<script setup lang="ts">
import { NButton, NDrawer, NDropdown, NEmpty, NIcon, NQrCode, NTooltip, type DropdownOption } from "naive-ui";

import { formatDateTime } from "../../app/formatters";
import type { LinkedWebDevice } from "../../types/workbench";
import { CopyIcon, DeleteIcon, DiagnosticsIcon, OpenIcon, RotateIcon, RouteIcon } from "../../utils/desktopIcons";

defineProps<{
  activeHost: string;
  bindAddress: string;
  candidateOptions: DropdownOption[];
  linkedDevices: LinkedWebDevice[];
  resolvedSessionUrl: string;
  show: boolean;
  tokenCountdown: string;
}>();

const emit = defineEmits<{
  (e: "copy"): void;
  (e: "diagnostics"): void;
  (e: "open"): void;
  (e: "refresh"): void;
  (e: "remove-device", deviceId: string): void;
  (e: "select-host", key: string | number): void;
  (e: "update:show", value: boolean): void;
}>();
</script>
