import { useStorageAsync, type RemovableRef, type UseStorageAsyncOptions } from "@vueuse/core";
import localforage from "localforage";
import type { MaybeRefOrGetter } from "vue";

type LocalForageInstance = typeof localforage;
type LocalForageOptions = Parameters<LocalForageInstance["createInstance"]>[0];

export interface UseForgeStorageOptions<T> extends UseStorageAsyncOptions<T> {
  createInstance?: LocalForageOptions;
  instance?: LocalForageInstance;
}

export function useForgeStorage<T>(
  key: string,
  initialValue: MaybeRefOrGetter<T>,
  options: UseForgeStorageOptions<T> = {},
): RemovableRef<T> & Promise<RemovableRef<T>> {
  const { createInstance, instance, ...storageOptions } = options;
  const forge = instance ?? (createInstance ? localforage.createInstance(createInstance) : localforage);

  return useStorageAsync(
    key,
    initialValue,
    {
      getItem: (storageKey) => forge.getItem<string>(storageKey),
      setItem: async (storageKey, value) => {
        await forge.setItem(storageKey, value);
      },
      removeItem: (storageKey) => forge.removeItem(storageKey),
    },
    storageOptions,
  );
}
