<template>
  <div class="grid gap-[0.9rem]" @click="closeContextMenu">
    <section class="flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-end">
      <div>
        <h1 class="m-0 text-[clamp(2rem,4vw,3.6rem)] leading-[0.94] tracking-[-0.06em]">
          LocalShareGo Web
        </h1>
        <p class="mt-[0.55rem] max-w-[34rem] text-[0.95rem] leading-[1.6] text-[var(--text-muted)]">
          The browser keeps its own clipboard cache and syncs it to selected devices.
        </p>
      </div>

      <div
        class="inline-flex items-center gap-[0.45rem] whitespace-nowrap rounded-full border border-[rgba(20,33,27,0.12)] bg-[rgba(255,252,247,0.84)] px-[0.8rem] py-[0.55rem] text-[var(--text-main)]"
      >
        <n-icon size="16"><Clock3 class="h-4 w-4" /></n-icon>
        <span>{{ expiresIn }}</span>
      </div>
    </section>

    <section v-if="authState === 'missing'" class="panel-surface">
      <n-result status="warning" title="Missing token" description="Copy the Web link from the desktop app and reopen it." />
    </section>

    <section v-else-if="authState === 'invalid'" class="panel-surface">
      <n-result status="error" title="Session expired" description="Copy a fresh Web link from the desktop app." />
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
import { NDropdown, NIcon, NResult } from "naive-ui";

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
  search,
  selectedId,
  session,
  submitDraft,
  syncAvailable,
  uploading,
} = useWebWorkbench();
</script>
