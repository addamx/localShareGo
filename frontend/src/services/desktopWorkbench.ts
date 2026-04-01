import { EventsOn } from "../../wailsjs/runtime/runtime";

import { desktopApp } from "../app/env";
import type {
  AppBootstrap,
  ClipboardItemRecord,
  ClipboardListQuery,
  ConnectivityReport,
  DesktopSettings,
  OnlineDevice,
  SyncClipboardResponse,
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
  listOnlineDevices() {
    return desktopApp!.ListOnlineDevices() as Promise<OnlineDevice[]>;
  },
  syncClipboardItem(itemId: string, targetDeviceIds: string[], syncAll: boolean) {
    return desktopApp!.SyncClipboardItem(itemId, targetDeviceIds, syncAll) as Promise<SyncClipboardResponse>;
  },
  openURL(url: string) {
    return Promise.resolve(desktopApp!.OpenURL(url));
  },
  hideDesktopApp() {
    return desktopApp!.HideDesktopApp() as Promise<void>;
  },
  getDesktopSettings() {
    return desktopApp!.GetDesktopSettings() as Promise<DesktopSettings>;
  },
  updateDesktopSettings(input: DesktopSettings) {
    return desktopApp!.UpdateDesktopSettings(input) as Promise<DesktopSettings>;
  },
  subscribeClipboardRefresh(handler: () => void) {
    const off = EventsOn("localshare://clipboard/refresh", handler);
    return () => off();
  },
  subscribeSessionRefresh(handler: () => void) {
    const off = EventsOn("localshare://session/refresh", handler);
    return () => off();
  },
};
