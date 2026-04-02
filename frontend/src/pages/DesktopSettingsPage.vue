<template>
  <div class="flex h-screen flex-col overflow-hidden p-2.5 md:p-3">
    <div class="h-2 shrink-0 [--wails-draggable:drag]" />

    <section
      class="flex min-h-0 flex-1 flex-col overflow-hidden rounded-[12px] border border-[rgba(20,33,27,0.12)] bg-[rgba(250,248,242,0.88)] shadow-[0_16px_48px_rgba(40,34,19,0.08)] backdrop-blur-[14px]"
    >
      <header class="flex items-center gap-3 border-b border-[rgba(20,33,27,0.08)] px-3 py-3">
        <n-button quaternary circle class="!rounded-[10px]" @click="void router.push('/desktop')">
          <template #icon>
            <n-icon>
              <BackIcon class="h-[18px] w-[18px]" />
            </n-icon>
          </template>
        </n-button>

        <div class="min-w-0">
          <p class="m-0 text-[0.95rem] font-semibold text-[var(--text-main)]">设置</p>
          <p class="m-0 text-[0.78rem] text-[var(--text-muted)]">仅保留桌面相关配置</p>
        </div>
      </header>

      <div class="min-h-0 flex-1 overflow-auto p-3">
        <n-tabs type="line" animated>
          <n-tab-pane name="shortcut" tab="快捷键">
            <section class="grid max-w-[38rem] gap-3">
              <div class="rounded-[12px] border border-[rgba(20,33,27,0.08)] bg-white/55 p-4">
                <div class="mb-3 flex items-start gap-3">
                  <div class="flex h-10 w-10 items-center justify-center rounded-[12px] bg-[rgba(31,122,90,0.09)] text-[#1f7a5a]">
                    <n-icon size="18">
                      <KeyboardIcon />
                    </n-icon>
                  </div>

                  <div class="min-w-0">
                    <p class="m-0 text-[0.92rem] font-semibold text-[var(--text-main)]">显示应用</p>
                    <p class="m-0 mt-1 text-[0.8rem] leading-6 text-[var(--text-muted)]">
                      应用隐藏到托盘后，可通过这个快捷键重新显示。
                    </p>
                  </div>
                </div>

                <button
                  ref="captureButtonRef"
                  type="button"
                  class="flex min-h-[3rem] w-full items-center justify-between rounded-[12px] border border-[rgba(20,33,27,0.1)] bg-[rgba(255,252,247,0.92)] px-3 py-2 text-left text-[0.92rem] text-[var(--text-main)] transition-colors duration-150 hover:border-[rgba(31,122,90,0.28)] focus:outline-none focus:ring-2 focus:ring-[rgba(31,122,90,0.18)]"
                  @click="startCapture"
                >
                  <span>{{ captureDisplay }}</span>
                  <small class="text-[0.76rem] text-[var(--text-muted)]">
                    {{ capturing ? "按下组合键" : "点击后录入" }}
                  </small>
                </button>

                <div class="mt-3 flex flex-wrap items-center gap-2">
                  <n-button secondary @click="startCapture">录入快捷键</n-button>
                  <n-button quaternary @click="clearHotkey">清空</n-button>
                  <n-button type="primary" :loading="saving" @click="void saveSettings()">保存</n-button>
                </div>
              </div>
            </section>
          </n-tab-pane>
        </n-tabs>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { NButton, NIcon, NTabPane, NTabs, useMessage } from "naive-ui";

import { describeError } from "../app/formatters";
import { desktopWorkbench } from "../services/desktopWorkbench";
import { BackIcon, KeyboardIcon } from "../utils/desktopIcons";
import { formatHotkeyFromKeyboardEvent } from "../utils/hotkeys";

const router = useRouter();
const message = useMessage();

const captureButtonRef = ref<HTMLButtonElement | null>(null);
const capturing = ref(false);
const capturePreview = ref("");
const draftHotkey = ref("");
const saving = ref(false);

const captureDisplay = computed(() => {
  if (capturing.value) {
    return capturePreview.value || "按下快捷键...";
  }
  return draftHotkey.value || "未设置";
});

onMounted(async () => {
  window.addEventListener("keydown", handleWindowKeyDown, true);
  window.addEventListener("blur", handleWindowBlur);

  try {
    const currentSettings = await desktopWorkbench.getDesktopSettings();
    draftHotkey.value = currentSettings.showAppHotkey;
  } catch (error) {
    message.error(describeError(error));
  }
});

onBeforeUnmount(() => {
  window.removeEventListener("keydown", handleWindowKeyDown, true);
  window.removeEventListener("blur", handleWindowBlur);
});

function startCapture() {
  capturing.value = true;
  capturePreview.value = "";
  void nextTick(() => {
    captureButtonRef.value?.focus();
  });
}

function clearHotkey() {
  capturing.value = false;
  capturePreview.value = "";
  draftHotkey.value = "";
}

function handleWindowKeyDown(event: KeyboardEvent) {
  if (capturing.value) {
    event.preventDefault();
    event.stopPropagation();
    applyCapturedHotkey(event);
    return;
  }

  if (event.key === "Escape") {
    void desktopWorkbench.hideDesktopApp();
  }
}

function applyCapturedHotkey(event: KeyboardEvent) {
  if (event.key === "Escape" && !event.ctrlKey && !event.altKey && !event.shiftKey && !event.metaKey) {
    capturing.value = false;
    capturePreview.value = "";
    return;
  }

  if (
    (event.key === "Backspace" || event.key === "Delete") &&
    !event.ctrlKey &&
    !event.altKey &&
    !event.shiftKey &&
    !event.metaKey
  ) {
    clearHotkey();
    return;
  }

  capturePreview.value = formatHotkeyFromKeyboardEvent(event, true);

  const nextHotkey = formatHotkeyFromKeyboardEvent(event);
  if (!nextHotkey) {
    return;
  }

  draftHotkey.value = nextHotkey;
  capturing.value = false;
  capturePreview.value = "";
}

function handleWindowBlur() {
  capturePreview.value = "";
  void desktopWorkbench.hideDesktopApp();
}

async function saveSettings() {
  saving.value = true;

  try {
    const saved = await desktopWorkbench.updateDesktopSettings({
      showAppHotkey: draftHotkey.value,
    });
    draftHotkey.value = saved.showAppHotkey;
    capturing.value = false;
    capturePreview.value = "";
    message.success("设置已保存");
  } catch (error) {
    message.error(describeError(error));
  } finally {
    saving.value = false;
  }
}
</script>
