import { EventsOn } from "../../wailsjs/runtime/runtime";

import { desktopApp } from "../app/env";
import type {
  AppBootstrap,
  ClipboardItemRecord,
  ClipboardListQuery,
  ConnectivityReport,
} from "../types/workbench";

export const desktopWorkbench = {
  isAvailable() {
    return Boolean(desktopApp?.GetBootstrapContext);
  },
  getBootstrapContext() {
    return desktopApp!.GetBootstrapContext() as Promise<AppBootstrap>;
  },
  listClipboardItems(query: ClipboardListQuery) {
    return desktopApp!.ListClipboardItems(query) as Promise<ClipboardItemRecord[]>;
  },
  activateClipboardItem(itemId: string) {
    return desktopApp!.ActivateClipboardItem(itemId) as Promise<ClipboardItemRecord>;
  },
  updateClipboardItemPin(itemId: string, pinned: boolean) {
    return desktopApp!.UpdateClipboardItemPin(itemId, pinned) as Promise<ClipboardItemRecord>;
  },
  deleteClipboardItem(itemId: string) {
    return desktopApp!.DeleteClipboardItem(itemId);
  },
  clearClipboardHistory() {
    return desktopApp!.ClearClipboardHistory();
  },
  rotateSessionToken() {
    return desktopApp!.RotateSessionToken();
  },
  getConnectivityReport() {
    return desktopApp!.GetConnectivityReport() as Promise<ConnectivityReport>;
  },
  openURL(url: string) {
    return Promise.resolve(desktopApp!.OpenURL(url));
  },
  subscribeClipboardRefresh(handler: () => void) {
    const off = EventsOn("localshare://clipboard/refresh", handler);
    return () => off();
  },
};
