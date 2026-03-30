<template>
  <div
    class="grid min-h-screen p-2.5 md:p-3"
    :class="menuBarVisible ? 'grid-rows-[auto,minmax(0,1fr)] gap-3' : 'grid-rows-[minmax(0,1fr)] gap-0'"
    @click="handleSurfaceClick"
  >
    <DesktopMenuBar
      v-if="menuBarVisible"
      :available="available"
      @open-web="openWebPanel"
    />

    <section
      v-if="!available"
      class="grid min-h-0 place-items-center rounded-[12px] border border-[rgba(20,33,27,0.12)] bg-[rgba(250,248,242,0.88)] shadow-[0_16px_48px_rgba(40,34,19,0.08)] backdrop-blur-[14px]"
    >
      <div class="grid gap-1 text-center">
        <span class="text-[0.92rem] font-bold text-[var(--text-main)]">NaiveDesktop API 不可用</span>
        <small class="text-[var(--text-muted)]">当前环境没有 Wails 绑定。</small>
      </div>
    </section>

    <DesktopClipboardList
      v-else
      :items="items"
      :loading="loading"
      :more-options="moreOptions"
      :refreshing="refreshing"
      :search="search"
      :selected-id="selectedId"
      @update:search="search = $event"
      @more-select="handleMoreSelect"
      @row-click="handleRowClick"
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
      @rotate="handleRotateSession"
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
import { NDropdown } from "naive-ui";

import DesktopClipboardList from "../components/desktop/DesktopClipboardList.vue";
import DesktopDetailDrawer from "../components/desktop/DesktopDetailDrawer.vue";
import DesktopDiagnosticsModal from "../components/desktop/DesktopDiagnosticsModal.vue";
import DesktopMenuBar from "../components/desktop/DesktopMenuBar.vue";
import DesktopWebDrawer from "../components/desktop/DesktopWebDrawer.vue";
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
  handleRotateSession,
  handleRowClick,
  hideMenuBar,
  items,
  loading,
  menuBarVisible,
  moreOptions,
  openContextMenu,
  openDiagnostics,
  openWebPanel,
  refreshing,
  resolvedSessionUrl,
  search,
  selectedId,
  tokenCountdown,
  webPanelOpen,
} = useDesktopWorkbench();

function handleSurfaceClick() {
  closeContextMenu();
  hideMenuBar();
}
</script>
