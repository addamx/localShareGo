import { computed } from "vue";

import { useWebClipboardItems } from "./useWebClipboardItems";
import { useWebFileCompose } from "./useWebFileCompose";

export function useWebWorkbench() {
  const clipboard = useWebClipboardItems();
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
