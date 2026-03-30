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
  CopyText(text: string): Promise<void>;
  OpenURL(url: string): void;
}

interface Window {
  go?: {
    main?: {
      App?: DesktopAppApi;
    };
  };
}
