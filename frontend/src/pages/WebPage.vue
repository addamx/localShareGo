<template>
  <div class="page-shell">
    <section class="hero-banner hero-banner--web">
      <div class="hero-copy">
        <p class="eyebrow">Web Route</p>
        <h1>LocalShareGo Web</h1>
        <p class="hero-text">
          通过 NaiveDesktop 分享的链接，在桌面、平板和手机浏览器中浏览历史并投递新的文本内容。
        </p>
        <n-space>
          <n-tag round type="success">
            设备 {{ session?.deviceName ?? "未连接" }}
          </n-tag>
          <n-tag round :type="serviceTagType">
            服务 {{ serviceState }}
          </n-tag>
          <n-tag round :type="streamTagType">
            流 {{ streamState }}
          </n-tag>
          <n-tag round type="warning">
            到期 {{ expiresIn }}
          </n-tag>
        </n-space>
      </div>
      <section class="panel-surface hero-panel">
        <div class="panel-heading">
          <div>
            <p class="panel-eyebrow">Session</p>
            <h2>当前会话</h2>
          </div>
          <n-button
            @click="refreshHistory(false)"
            :loading="loadingItems"
            :disabled="authState !== 'valid'"
          >
            刷新
          </n-button>
        </div>
        <div class="mono-block">{{ accessUrl }}</div>
        <p class="hero-note">
          路由固定为 <strong>/web</strong>，鉴权依赖链接中的 <code>token</code>。
        </p>
      </section>
    </section>

    <section v-if="authState === 'missing'" class="panel-surface">
      <n-result
        status="warning"
        title="缺少 token"
        description="请从 NaiveDesktop 复制 Web 访问链接，使用 /web?token=... 进入。"
      />
    </section>

    <section v-else-if="authState === 'invalid'" class="panel-surface">
      <n-result
        status="error"
        title="token 无效或已过期"
        description="请重新从 NaiveDesktop 复制新的 Web 访问链接。"
      />
    </section>

    <template v-else>
      <section class="workspace-grid">
        <div class="stack-column">
          <section class="panel-surface">
            <div class="panel-heading">
              <div>
                <p class="panel-eyebrow">Compose</p>
                <h2>提交新文本</h2>
              </div>
              <span class="byte-count">
                {{ draftBytes }} / {{ session?.maxTextBytes ?? 65536 }} bytes
              </span>
            </div>
            <n-input
              v-model:value="draft"
              type="textarea"
              placeholder="输入要发送到 NaiveDesktop 的文本"
              :autosize="{ minRows: 7, maxRows: 10 }"
            />
            <n-space>
              <n-button type="primary" @click="submitDraft">
                提交文本
              </n-button>
              <n-button secondary @click="draft = ''">
                清空输入
              </n-button>
            </n-space>
          </section>

          <section class="panel-surface">
            <div class="panel-heading">
              <div>
                <p class="panel-eyebrow">History</p>
                <h2>Web 历史视图</h2>
              </div>
              <n-button secondary @click="togglePinnedOnly">
                {{ pinnedOnly ? "查看全部" : "仅看置顶" }}
              </n-button>
            </div>
            <n-input
              v-model:value="search"
              placeholder="搜索历史内容"
              clearable
            />
            <div class="list-shell">
              <div v-if="initializing || loadingItems" class="state-shell">
                <n-spin size="small" /> 正在同步历史...
              </div>
              <n-empty
                v-else-if="items.length === 0"
                description="暂无可展示的剪贴板历史"
              />
              <article
                v-for="item in items"
                :key="item.id"
                class="history-card"
                :class="{ 'history-card--active': item.id === selectedId }"
              >
                <button class="history-select" @click="selectItem(item.id)">
                  <div class="history-meta">
                    <span>{{ formatSource(item.sourceKind) }}</span>
                    <span>{{ formatDateTime(item.createdAt) }}</span>
                    <span>{{ item.charCount }} chars</span>
                    <n-tag v-if="item.isCurrent" size="small" round type="success">
                      当前
                    </n-tag>
                    <n-tag v-if="item.pinned" size="small" round type="warning">
                      置顶
                    </n-tag>
                  </div>
                  <p>{{ item.preview || "空内容" }}</p>
                </button>
                <n-button size="small" @click="activateItem(item.id)">
                  激活到 NaiveDesktop
                </n-button>
              </article>
            </div>
          </section>
        </div>

        <div class="stack-column">
          <section class="panel-surface">
            <div class="panel-heading">
              <div>
                <p class="panel-eyebrow">Detail</p>
                <h2>选中详情</h2>
              </div>
            </div>
            <template v-if="selectedItem">
              <div class="detail-tags">
                <n-tag round>{{ formatSource(selectedItem.sourceKind) }}</n-tag>
                <n-tag round>{{ formatDateTime(selectedItem.createdAt) }}</n-tag>
                <n-tag round>{{ formatDateTime(selectedItem.updatedAt) }}</n-tag>
                <n-tag round>{{ selectedItem.charCount }} chars</n-tag>
              </div>
              <pre class="clipboard-pre">{{ selectedItemContent }}</pre>
            </template>
            <n-empty v-else description="请选择一条历史记录" />
            <p class="sync-note">
              {{ lastSyncedAt ? `最近同步于 ${formatDateTime(lastSyncedAt)}` : "尚未完成同步" }}
            </p>
          </section>
        </div>
      </section>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from "vue";
import { useDebounceFn } from "@vueuse/core";
import {
  NButton,
  NEmpty,
  NInput,
  NResult,
  NSpace,
  NSpin,
  NTag,
  useMessage,
} from "naive-ui";
import { useRoute } from "vue-router";

import {
  describeError,
  formatDateTime,
  formatRemaining,
  formatSource,
} from "../app/formatters";
import {
  createWorkbenchApiClient,
  WorkbenchApiError,
  type WorkbenchApiClient,
} from "../services/workbenchApi";
import type {
  ClipboardItemRecord,
  ClipboardItemSummary,
  SessionResponse,
} from "../types/workbench";

const route = useRoute();
const message = useMessage();

const initializing = ref(true);
const loadingItems = ref(false);
const authState = ref<"idle" | "missing" | "valid" | "invalid">("idle");
const serviceState = ref<"unknown" | "online" | "offline" | "failed">("unknown");
const streamState = ref<"idle" | "connecting" | "live" | "reconnecting">("idle");
const draft = ref("");
const search = ref("");
const pinnedOnly = ref(false);
const items = ref<ClipboardItemSummary[]>([]);
const selectedId = ref<string | null>(null);
const selectedDetail = ref<ClipboardItemRecord | null>(null);
const session = ref<SessionResponse | null>(null);
const api = ref<WorkbenchApiClient | null>(null);
const lastSyncedAt = ref<number | null>(null);

const token = computed(() => {
  const value = route.query.token;
  return typeof value === "string" ? value.trim() : "";
});
const accessUrl = computed(() => session.value?.accessUrl || window.location.href);
const expiresIn = computed(() =>
  formatRemaining(session.value?.expiresAt ?? null),
);
const draftBytes = computed(() => new TextEncoder().encode(draft.value).length);
const selectedItem = computed(
  () => selectedDetail.value ?? items.value.find((item) => item.id === selectedId.value) ?? null,
);
const selectedItemContent = computed(
  () =>
    (selectedItem.value && "content" in selectedItem.value
      ? selectedItem.value.content
      : selectedItem.value?.preview) || "",
);
const serviceTagType = computed(() => {
  if (serviceState.value === "online") return "success";
  if (serviceState.value === "failed") return "error";
  return "warning";
});
const streamTagType = computed(() => {
  if (streamState.value === "live") return "success";
  if (streamState.value === "reconnecting") return "warning";
  return "default";
});

const refreshBySearch = useDebounceFn(() => {
  void refreshHistory(true);
}, 280);

watch(search, () => {
  if (authState.value === "valid") {
    refreshBySearch();
  }
});

watch(
  token,
  () => {
    void initializePage();
  },
  { immediate: true },
);

onBeforeUnmount(() => {
  closeEvents();
});

let eventSource: EventSource | null = null;

async function initializePage() {
  closeEvents();

  initializing.value = true;
  authState.value = "idle";
  serviceState.value = "unknown";
  streamState.value = "idle";
  items.value = [];
  selectedId.value = null;
  selectedDetail.value = null;
  session.value = null;
  lastSyncedAt.value = null;

  if (!token.value) {
    authState.value = "missing";
    api.value = null;
    initializing.value = false;
    return;
  }

  api.value = createWorkbenchApiClient(window.location.origin, token.value);

  try {
    const health = await api.value.health();
    serviceState.value = health.status === "failed" ? "failed" : "online";
    session.value = await api.value.session();
    authState.value = "valid";
    await refreshHistory(true);
    connectEvents();
  } catch (error) {
    if (error instanceof WorkbenchApiError && error.status === 401) {
      authState.value = "invalid";
    } else {
      serviceState.value = "offline";
      message.error(describeError(error));
    }
  } finally {
    initializing.value = false;
  }
}

async function refreshHistory(silent: boolean) {
  if (!api.value || authState.value !== "valid") {
    return;
  }

  loadingItems.value = true;
  try {
    const response = await api.value.listClipboardItems({
      search: search.value.trim() || null,
      pinnedOnly: pinnedOnly.value,
      limit: 80,
    });

    items.value = response.items;
    if (!selectedId.value || !items.value.some((item) => item.id === selectedId.value)) {
      selectedId.value = items.value[0]?.id ?? null;
    }

    if (selectedId.value) {
      selectedDetail.value = await api.value.getClipboardItem(selectedId.value);
    } else {
      selectedDetail.value = null;
    }

    lastSyncedAt.value = Date.now();
    if (!silent) {
      message.success("历史已刷新");
    }
  } catch (error) {
    if (error instanceof WorkbenchApiError && error.status === 401) {
      authState.value = "invalid";
      closeEvents();
    }
    message.error(describeError(error));
  } finally {
    loadingItems.value = false;
  }
}

function togglePinnedOnly() {
  pinnedOnly.value = !pinnedOnly.value;
  void refreshHistory(true);
}

async function selectItem(itemId: string) {
  if (!api.value) return;

  selectedId.value = itemId;
  try {
    selectedDetail.value = await api.value.getClipboardItem(itemId);
  } catch (error) {
    message.error(describeError(error));
  }
}

async function submitDraft() {
  if (!api.value) return;
  if (!draft.value.trim()) {
    message.warning("请输入要提交的文本");
    return;
  }

  try {
    await api.value.submitClipboardItem(draft.value.trim());
    draft.value = "";
    message.success("文本已提交");
    await refreshHistory(true);
  } catch (error) {
    if (error instanceof WorkbenchApiError && error.status === 401) {
      authState.value = "invalid";
      closeEvents();
    }
    message.error(describeError(error));
  }
}

async function activateItem(itemId: string) {
  if (!api.value) return;

  try {
    await api.value.activateClipboardItem(itemId);
    message.success("已请求 NaiveDesktop 激活该条内容");
    await refreshHistory(true);
  } catch (error) {
    if (error instanceof WorkbenchApiError && error.status === 401) {
      authState.value = "invalid";
      closeEvents();
    }
    message.error(describeError(error));
  }
}

function connectEvents() {
  if (!api.value) return;

  closeEvents();
  streamState.value = "connecting";

  eventSource = new EventSource(api.value.eventsUrl());
  eventSource.onopen = () => {
    streamState.value = "live";
  };
  eventSource.addEventListener("refresh", async () => {
    await refreshHistory(true);
  });
  eventSource.onerror = () => {
    streamState.value = "reconnecting";
  };
}

function closeEvents() {
  eventSource?.close();
  eventSource = null;
}
</script>

<style scoped>
.hero-banner--web {
  grid-template-columns: minmax(0, 1.25fr) minmax(320px, 0.85fr);
}

.hero-panel {
  padding: 1.5rem;
}

.hero-note {
  margin: 1rem 0 0;
  color: var(--text-muted);
  line-height: 1.65;
}

code {
  padding: 0.12rem 0.45rem;
  border-radius: 999px;
  background: rgba(24, 37, 31, 0.08);
  font-family: "Cascadia Code", "SFMono-Regular", Consolas, monospace;
}

.byte-count {
  color: var(--text-muted);
  font-size: 0.9rem;
}

.list-shell {
  display: grid;
  gap: 0.9rem;
  margin-top: 1rem;
}

.history-card {
  display: grid;
  gap: 0.9rem;
  padding: 1rem 1.05rem;
  border-radius: 20px;
  border: 1px solid rgba(24, 37, 31, 0.08);
  background: rgba(255, 255, 255, 0.72);
}

.history-card--active {
  border-color: rgba(31, 122, 90, 0.35);
  background: rgba(240, 248, 243, 0.95);
}

.history-select {
  display: grid;
  gap: 0.75rem;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
  text-align: left;
}

.history-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 0.55rem;
  align-items: center;
  color: var(--text-muted);
  font-size: 0.82rem;
}

.history-select p {
  margin: 0;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.detail-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 0.6rem;
  margin-bottom: 1rem;
}

.sync-note {
  margin: 1rem 0 0;
  color: var(--text-muted);
  font-size: 0.9rem;
}

.state-shell {
  display: flex;
  gap: 0.75rem;
  align-items: center;
  padding: 1rem 0;
  color: var(--text-muted);
}
</style>
