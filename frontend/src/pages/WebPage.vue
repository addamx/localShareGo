<template>
  <div class="grid gap-[0.9rem]" @click="closeContextMenu">
    <section class="flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-end">
      <div>
        <h1 class="m-0 text-[clamp(2rem,4vw,3.6rem)] leading-[0.94] tracking-[-0.06em]">
          LocalShareGo Web
        </h1>
        <p class="mt-[0.55rem] max-w-[34rem] text-[0.95rem] leading-[1.6] text-[var(--text-muted)]">
          当前浏览器保存自己的剪贴板数据，发送时会按已配置设备执行同步。
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
      <n-result status="warning" title="缺少 token" description="请从桌面端复制 Web 链接后再打开。" />
    </section>

    <section v-else-if="authState === 'invalid'" class="panel-surface">
      <n-result status="error" title="会话已过期" description="请从桌面端重新复制新的 Web 链接。" />
    </section>

    <template v-else>
      <section class="grid items-start gap-[0.9rem] md:grid-cols-[minmax(280px,360px)_minmax(0,1fr)]">
        <WebComposePanel
          :draft="draft"
          :draft-bytes="draftBytes"
          :max-bytes="session?.maxTextBytes ?? 65536"
          :sync-available="syncAvailable"
          :sync-label="composeSyncLabel"
          :sync-options="composeSyncOptions"
          @update:draft="draft = $event"
          @clear="draft = ''"
          @submit="submitDraft"
          @sync-select="handleComposeSyncSelect"
        />

        <WebClipboardPanel
          :initializing="initializing"
          :items="items"
          :pinned-only="pinnedOnly"
          :search="search"
          :selected-id="selectedId"
          @activate="activateItem"
          @row-contextmenu="openContextMenu($event.event, $event.item)"
          @toggle-pinned="togglePinnedOnly"
          @update:search="search = $event"
        />
      </section>
    </template>

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
import { Clock3 } from "lucide-vue-next";
import { NDropdown, NIcon, NResult } from "naive-ui";

import WebClipboardPanel from "../components/web/WebClipboardPanel.vue";
import WebComposePanel from "../components/web/WebComposePanel.vue";
import { useWebWorkbench } from "../hooks/useWebWorkbench";

const {
  activateItem,
  authState,
  closeContextMenu,
  composeSyncLabel,
  composeSyncOptions,
  contextMenuOpen,
  contextMenuOptions,
  contextMenuX,
  contextMenuY,
  draft,
  draftBytes,
  expiresIn,
  handleComposeSyncSelect,
  handleContextSelect,
  initializing,
  items,
  openContextMenu,
  pinnedOnly,
  search,
  selectedId,
  session,
  submitDraft,
  syncAvailable,
  togglePinnedOnly,
} = useWebWorkbench();
</script>
