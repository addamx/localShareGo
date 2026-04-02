<template>
  <section class="panel-surface rounded-[26px] p-4 sm:p-[1rem]" @paste="handlePaste">
    <div class="mb-3 flex items-start justify-between gap-3">
      <div>
        <p class="panel-eyebrow">Compose</p>
        <h2 class="mt-0.5 text-[1.08rem]">Text / File</h2>
      </div>
      <span class="text-[0.82rem] leading-none text-[var(--text-muted)]">
        {{ draftBytes }} / {{ maxBytes }}
      </span>
    </div>

    <div
      class="mb-3 rounded-[18px] border border-dashed px-3 py-3 transition-colors duration-150"
      :class="dragActive ? 'border-[rgba(31,122,90,0.55)] bg-[rgba(31,122,90,0.08)]' : 'border-[rgba(20,33,27,0.12)] bg-[rgba(255,252,247,0.64)]'"
      @dragenter.prevent="dragActive = true"
      @dragover.prevent="dragActive = true"
      @dragleave.prevent="dragActive = false"
      @drop.prevent="handleDrop"
    >
      <input ref="fileInput" class="hidden" type="file" @change="handleFileInputChange" />

      <div class="flex flex-wrap items-center gap-2">
        <n-button secondary class="!rounded-[999px]" @click="openFilePicker">
          <template #icon>
            <n-icon><Upload class="h-[16px] w-[16px]" /></n-icon>
          </template>
          Pick File
        </n-button>

        <span class="text-[0.82rem] text-[var(--text-muted)]">Paste or drag a single file</span>
      </div>

      <div v-if="pendingFileLabel" class="mt-2 flex items-center gap-2 text-[0.84rem] text-[var(--text-main)]">
        <n-spin v-if="uploading" size="small" />
        <n-icon v-else size="15" class="text-[var(--text-muted)]">
          <Upload class="h-[15px] w-[15px]" />
        </n-icon>
        <span>
          {{ uploading ? "Sending" : "Ready to send" }}
          {{ pendingFileLabel }}{{ pendingFileSizeLabel ? ` - ${pendingFileSizeLabel}` : "" }}
        </span>
      </div>
    </div>

    <n-input
      :value="draft"
      type="textarea"
      placeholder="Type text, or pick / paste / drag a file, then click send"
      :autosize="{ minRows: 7, maxRows: 12 }"
      class="[&_.n-input-wrapper]:!rounded-[16px]"
      @update:value="emit('update:draft', $event)"
    />

    <div class="mt-[0.65rem] flex justify-end gap-1.5">
      <n-button circle secondary @click="emit('clear')">
        <template #icon>
          <n-icon><CloseIcon class="h-[18px] w-[18px]" /></n-icon>
        </template>
      </n-button>

      <n-dropdown trigger="click" placement="bottom-end" :options="syncOptions" @select="emit('sync-select', $event)">
        <n-button secondary class="gap-1.5 !rounded-[999px]" :disabled="!syncAvailable">
          <template #icon>
            <n-icon><Share2 class="h-[16px] w-[16px]" /></n-icon>
          </template>
          {{ syncLabel }}
        </n-button>
      </n-dropdown>

      <n-button circle type="primary" @click="emit('submit')">
        <template #icon>
          <n-icon><SendHorizontal class="h-[18px] w-[18px]" /></n-icon>
        </template>
      </n-button>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref } from "vue";
import { SendHorizontal, Share2, Upload } from "lucide-vue-next";
import { NButton, NDropdown, NIcon, NInput, NSpin, type DropdownOption } from "naive-ui";

import { CloseIcon } from "../../utils/desktopIcons";

defineProps<{
  draft: string;
  draftBytes: number;
  maxBytes: number;
  pendingFileLabel: string;
  pendingFileSizeLabel: string;
  syncAvailable: boolean;
  syncLabel: string;
  syncOptions: DropdownOption[];
  uploading: boolean;
}>();

const emit = defineEmits<{
  (e: "clear"): void;
  (e: "files-dropped", files: File[]): void;
  (e: "files-pasted", files: File[]): void;
  (e: "files-selected", files: File[]): void;
  (e: "submit"): void;
  (e: "sync-select", key: string | number): void;
  (e: "update:draft", value: string): void;
}>();

const fileInput = ref<HTMLInputElement | null>(null);
const dragActive = ref(false);

function openFilePicker() {
  fileInput.value?.click();
}

function handleFileInputChange(event: Event) {
  const input = event.target as HTMLInputElement;
  emit("files-selected", toFileArray(input.files));
  input.value = "";
}

function handleDrop(event: DragEvent) {
  dragActive.value = false;
  emit("files-dropped", toFileArray(event.dataTransfer?.files ?? null));
}

function handlePaste(event: ClipboardEvent) {
  const files = toFileArray(event.clipboardData?.files ?? null);
  if (files.length === 0) {
    return;
  }

  event.preventDefault();
  emit("files-pasted", files);
}

function toFileArray(files: FileList | File[] | null | undefined) {
  if (!files || files.length === 0) {
    return [];
  }

  return Array.from(files);
}
</script>
