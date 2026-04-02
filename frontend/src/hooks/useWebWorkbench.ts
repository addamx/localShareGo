import { computed } from "vue";
import { useRoute } from "vue-router";

import { useWebClipboardItems } from "./useWebClipboardItems";
import { useWebFileCompose } from "./useWebFileCompose";

export function useWebWorkbench() {
  const route = useRoute();
  const token = computed(() => {
    const value = route.query.token;
    return typeof value === "string" ? value.trim() : "";
  });

  const clipboard = useWebClipboardItems(token);
  const fileCompose = useWebFileCompose({
    maxTextBytes: computed(() => clipboard.session.value?.maxTextBytes ?? 65_536),
    onlineDevices: computed(() => clipboard.onlineDevices.value),
    submitTextDraft: clipboard.submitTextDraft,
    uploadFile: clipboard.uploadFile,
  });

  return {
    ...clipboard,
    ...fileCompose,
  };
}
