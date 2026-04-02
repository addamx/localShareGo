import { nextTick, onMounted, ref, unref, type Ref } from "vue";

type ScrollElement = HTMLElement | null | undefined;
type ScrollElementSource<T extends ScrollElement = HTMLElement> = T | Ref<T> | (() => T);

export const useRestorableScrollPosition = <T extends ScrollElement = HTMLElement>(
  el: ScrollElementSource<T>,
) => {
  const scrollPosition = ref(0);

  const getEl = (): T => {
    if (typeof el === "function") {
      return el();
    }

    return unref(el);
  };

  function recordPosition() {
    scrollPosition.value = getEl()?.scrollTop ?? 0;
  }

  function restorePosition() {
    console.log('Restoring scroll position:', scrollPosition.value);
    nextTick(() => {
      const element = getEl();
      if (!element) {
        return;
      }
      element.scrollTop = scrollPosition.value;
    });
  }

  onMounted(() => {
    scrollPosition.value = getEl()?.scrollTop ?? 0;
  });

  return {
    recordPosition,
    restorePosition,
    scrollPosition,
  };
};
