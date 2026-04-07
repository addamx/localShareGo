<template>
  <div class="grid gap-[0.9rem]" @click="closeContextMenu">
    <section class="flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-end">
      <div>
        <h1 class="m-0 text-[clamp(2rem,4vw,3.6rem)] leading-[0.94] tracking-[-0.06em]">
          LocalShareGo Web
        </h1>
        <p class="mt-[0.55rem] max-w-[34rem] text-[0.95rem] leading-[1.6] text-[var(--text-muted)]">
          浏览器会保留自己的剪贴板缓存，并与选中的设备同步。
        </p>
      </div>

      <div
        v-if="authState === 'ready'"
        class="inline-flex items-center gap-[0.45rem] whitespace-nowrap rounded-full border border-[rgba(20,33,27,0.12)] bg-[rgba(255,252,247,0.84)] px-[0.8rem] py-[0.55rem] text-[var(--text-main)]"
      >
        <n-icon size="16"><Clock3 class="h-4 w-4" /></n-icon>
        <span>设备关联剩余 {{ expiresIn }}</span>
        <n-button v-if="renewAvailable" size="tiny" secondary @click="void renewDeviceAssociation()">延期</n-button>
      </div>
    </section>

    <section v-if="authState === 'no-link'" class="panel-surface">
      <n-result status="warning" title="无可用关联" description="当前浏览器还没有可用的设备关联。">
        <template #footer>
          <n-button type="primary" @click="void requestDeviceAssociation()">申请关联</n-button>
        </template>
      </n-result>
    </section>

    <section v-else-if="authState === 'expired'" class="panel-surface">
      <n-result status="error" title="关联已失效" description="当前浏览器的设备关联已经失效。">
        <template #footer>
          <n-button type="primary" @click="void requestDeviceAssociation()">申请关联</n-button>
        </template>
      </n-result>
    </section>

    <section v-else-if="authState === 'requesting' || authState === 'waiting-approval'" class="panel-surface">
      <n-result
        status="info"
        title="等待桌面端确认"
        description="已发起设备关联申请，请到桌面端确认是否允许关联。"
      />
    </section>

    <template v-else>
      <section class="grid items-start gap-[0.9rem] md:grid-cols-[minmax(280px,360px)_minmax(0,1fr)]">
        <WebComposePanel
          :draft="draft"
          :draft-bytes="draftBytes"
          :max-bytes="session?.maxTextBytes ?? 65536"
          :pending-file-label="pendingFileLabel"
          :pending-file-size-label="pendingFileSizeLabel"
          :sync-available="syncAvailable"
          :sync-label="composeSyncLabel"
          :sync-options="composeSyncOptions"
          :uploading="uploading"
          @clear="clearCompose"
          @files-dropped="handleFilesDropped"
          @files-pasted="handleFilesPasted"
          @files-selected="handleFilesSelected"
          @submit="submitDraft"
          @sync-select="handleComposeSyncSelect"
          @update:draft="draft = $event"
        />

        <WebClipboardPanel
          :initializing="initializing"
          :items="items"
          :more-options="moreOptions"
          :search="search"
          :selected-id="selectedId"
          @more-select="handleMoreSelect"
          @row-click="handleRowClick"
          @row-contextmenu="openContextMenu($event.event, $event.item)"
          @update:search="search = $event"
        />
      </section>
    </template>

    <n-dropdown
      trigger="manual"
      :size="webDropdownSize"
      placement="bottom-start"
      :theme-overrides="webDropdownThemeOverrides"
      :x="contextMenuX"
      :y="contextMenuY"
      :options="contextMenuOptions"
      :show="contextMenuOpen"
      @select="handleContextSelect"
      @clickoutside="closeContextMenu"
    />
  </div>
</template>

<script setup lang="ts">
import { Clock3 } from "lucide-vue-next";
import { NButton, NDropdown, NIcon, NResult } from "naive-ui";

import WebClipboardPanel from "../components/web/WebClipboardPanel.vue";
import WebComposePanel from "../components/web/WebComposePanel.vue";
import { useWebWorkbench } from "../hooks/useWebWorkbench";
import { webDropdownSize, webDropdownThemeOverrides } from "../utils/dropdown";

const {
  authState,
  closeContextMenu,
  clearCompose,
  composeSyncLabel,
  composeSyncOptions,
  contextMenuOpen,
  contextMenuOptions,
  contextMenuX,
  contextMenuY,
  draft,
  draftBytes,
  expiresIn,
  handleMoreSelect,
  handleComposeSyncSelect,
  handleFilesDropped,
  handleFilesPasted,
  handleFilesSelected,
  handleContextSelect,
  handleRowClick,
  initializing,
  items,
  moreOptions,
  openContextMenu,
  pendingFileLabel,
  pendingFileSizeLabel,
  renewAvailable,
  renewDeviceAssociation,
  requestDeviceAssociation,
  search,
  selectedId,
  session,
  submitDraft,
  syncAvailable,
  uploading,
} = useWebWorkbench();
</script>
