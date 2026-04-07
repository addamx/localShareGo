<template>
  <div class="flex h-screen flex-col overflow-hidden" @click="handleSurfaceClick">
    <div class="h-3 shrink-0 bg-amber-100 cursor-grab [--wails-draggable:drag]" />

    <section
      v-if="!available"
      class="grid min-h-0 flex-1 place-items-center rounded-[12px] border border-[rgba(20,33,27,0.12)] bg-[rgba(250,248,242,0.88)] shadow-[0_16px_48px_rgba(40,34,19,0.08)] backdrop-blur-[14px]"
    >
      <div class="grid gap-1 text-center">
        <span class="text-[0.92rem] font-bold text-[var(--text-main)]">Desktop API unavailable</span>
        <small class="text-[var(--text-muted)]">This environment does not have Wails bindings.</small>
      </div>
    </section>

    <DesktopClipboardList
      v-else
      ref="clipboardListRef"
      class="min-h-0 flex-1"
      :items="items"
      :loading="loading"
      :more-options="moreOptions"
      :refreshing="refreshing"
      :search="search"
      :selected-id="selectedId"
      @search-arrow-down="selectFirstItem"
      @update:search="search = $event"
      @more-select="handleMoreSelect"
      @row-click="handleClipboardRowClick"
      @row-contextmenu="openContextMenu($event.event, $event.item)"
    />

    <DesktopWebDrawer
      v-model:show="webPanelOpen"
      :active-host="activeHost"
      :bind-address="bindAddress"
      :candidate-options="candidateOptions"
      :resolved-session-url="resolvedSessionUrl"
      :token-countdown="tokenCountdown"
      @copy="handleCopyEntry"
      @diagnostics="openDiagnostics"
      @open="handleOpenEntry"
      @refresh="handleRefreshSession"
      @select-host="handleCandidateSelect"
    />

    <DesktopDetailDrawer
      v-model:show="detailPanelOpen"
      :item="detailItem"
      @copy="handleCopyDetail"
      @delete="handleDeleteDetail"
    />

    <DesktopDiagnosticsModal
      v-model:show="diagnosticsModalOpen"
      :bind-address="bindAddress"
      :connectivity="connectivity"
      :loading="diagnosticsLoading"
      :resolved-session-url="resolvedSessionUrl"
      :server-state="bootstrap?.services.httpServer.state"
    />

    <n-dropdown
      trigger="manual"
      placement="bottom-start"
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
import { ref } from "vue";
import { NDropdown } from "naive-ui";

import DesktopClipboardList from "../components/desktop/DesktopClipboardList.vue";
import DesktopDetailDrawer from "../components/desktop/DesktopDetailDrawer.vue";
import DesktopDiagnosticsModal from "../components/desktop/DesktopDiagnosticsModal.vue";
import DesktopWebDrawer from "../components/desktop/DesktopWebDrawer.vue";
import { useRestorableScrollPosition } from "../hooks/useRestorableScrollPosition";
import { useDesktopWorkbench } from "../hooks/useDesktopWorkbench";

const {
  activeHost,
  available,
  bindAddress,
  bootstrap,
  candidateOptions,
  closeContextMenu,
  connectivity,
  contextMenuOpen,
  contextMenuOptions,
  contextMenuX,
  contextMenuY,
  detailItem,
  detailPanelOpen,
  diagnosticsLoading,
  diagnosticsModalOpen,
  handleCandidateSelect,
  handleContextSelect,
  handleCopyDetail,
  handleCopyEntry,
  handleDeleteDetail,
  handleMoreSelect,
  handleOpenEntry,
  handleRefreshSession,
  handleRowClick,
  items,
  loading,
  moreOptions,
  openContextMenu,
  openDiagnostics,
  refreshing,
  resolvedSessionUrl,
  search,
  selectFirstItem,
  selectedId,
  tokenCountdown,
  webPanelOpen,
} = useDesktopWorkbench();

type DesktopClipboardListExpose = {
  getScroller: () => HTMLElement | null;
};

const clipboardListRef = ref<DesktopClipboardListExpose | null>(null);
const { recordPosition, restorePosition } = useRestorableScrollPosition(() => clipboardListRef.value?.getScroller());

async function handleClipboardRowClick(item: (typeof items.value)[number]) {
  recordPosition();
  await handleRowClick(item);
  restorePosition();
}

function handleSurfaceClick() {
  closeContextMenu();
}
</script>
