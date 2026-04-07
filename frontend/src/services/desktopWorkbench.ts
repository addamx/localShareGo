import { EventsOn } from "../../wailsjs/runtime/runtime";

import { desktopApp } from "../app/env";
import type {
  AppBootstrap,
  ClipboardItemRecord,
  ClipboardListQuery,
  ConnectivityReport,
  DesktopSettings,
  FileTransferEvent,
  LinkedWebDevice,
  OnlineDevice,
  PairRequestSummary,
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
  listLinkedWebDevices() {
    return desktopApp!.ListLinkedWebDevices() as Promise<LinkedWebDevice[]>;
  },
  removeLinkedWebDevice(deviceId: string) {
    return desktopApp!.RemoveLinkedWebDevice(deviceId) as Promise<void>;
  },
  listPairRequests() {
    return desktopApp!.ListPairRequests() as Promise<PairRequestSummary[]>;
  },
  approvePairRequest(requestId: string) {
    return desktopApp!.ApprovePairRequest(requestId) as Promise<PairRequestSummary>;
  },
  rejectPairRequest(requestId: string) {
    return desktopApp!.RejectPairRequest(requestId) as Promise<PairRequestSummary>;
  },
  syncClipboardItem(itemId: string, targetDeviceIds: string[], syncAll: boolean) {
    return desktopApp!.SyncClipboardItem(itemId, targetDeviceIds, syncAll) as Promise<SyncClipboardResponse>;
  },
  receiveClipboardFile(itemId: string) {
    return desktopApp!.ReceiveClipboardFile(itemId) as Promise<ClipboardItemRecord>;
  },
  openURL(url: string) {
    return Promise.resolve(desktopApp!.OpenURL(url));
  },
  hideDesktopApp() {
    return desktopApp!.HideDesktopApp() as Promise<void>;
  },
  showDesktopApp() {
    return desktopApp!.ShowDesktopApp() as Promise<void>;
  },
  getDesktopPinned() {
    return desktopApp!.GetDesktopPinned() as Promise<boolean>;
  },
  setDesktopPinned(pinned: boolean) {
    return desktopApp!.SetDesktopPinned(pinned) as Promise<boolean>;
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
  subscribeFileTransferProgress(handler: (event: FileTransferEvent) => void) {
    const off = EventsOn("localshare://file-transfer/progress", handler);
    return () => off();
  },
  subscribePairRequest(handler: (event: PairRequestSummary) => void) {
    const off = EventsOn("localshare://pair-request/pending", handler);
    return () => off();
  },
};
