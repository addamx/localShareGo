interface DesktopAppApi {
  GetBootstrapContext(): Promise<any>;
  ListClipboardItems(query?: any): Promise<any[]>;
  GetClipboardItem(itemId: string): Promise<any>;
  ActivateClipboardItem(itemId: string): Promise<any>;
  UpdateClipboardItemPin(itemId: string, pinned: boolean): Promise<any>;
  DeleteClipboardItem(itemId: string): Promise<void>;
  ClearClipboardHistory(): Promise<number>;
  RotateSessionToken(): Promise<any>;
  GetConnectivityReport(): Promise<any>;
  ListOnlineDevices(): Promise<any[]>;
  ListLinkedWebDevices(): Promise<any[]>;
  RemoveLinkedWebDevice(deviceId: string): Promise<void>;
  ListPairRequests(): Promise<any[]>;
  ApprovePairRequest(requestId: string): Promise<any>;
  RejectPairRequest(requestId: string): Promise<any>;
  SyncClipboardItem(itemId: string, targetDeviceIds: string[], syncAll: boolean): Promise<any>;
  ReceiveClipboardFile(itemId: string): Promise<any>;
  CopyText(text: string): Promise<void>;
  HideDesktopApp(): Promise<void>;
  ShowDesktopApp(): Promise<void>;
  GetDesktopPinned(): Promise<boolean>;
  SetDesktopPinned(pinned: boolean): Promise<boolean>;
  GetDesktopSettings(): Promise<any>;
  UpdateDesktopSettings(input: any): Promise<any>;
  OpenURL(url: string): void;
}

interface Window {
  go?: {
    main?: {
      App?: DesktopAppApi;
    };
  };
}
