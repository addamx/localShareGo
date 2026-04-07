package main

import (
	"context"
	"fmt"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"localShareGo/internal/apierr"
	"localShareGo/internal/auth"
	"localShareGo/internal/clipboard"
	"localShareGo/internal/desktopshell"
	"localShareGo/internal/httpserver"
	"localShareGo/internal/runtimeapp"
	"localShareGo/internal/settings"
	"localShareGo/internal/store"
)

type App struct {
	ctx      context.Context
	runtime  *runtimeapp.AppRuntime
	shell    *desktopshell.Manager
	trayIcon []byte
}

func NewApp(trayIcon []byte) *App {
	return &App{trayIcon: trayIcon}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	runtime, err := newAppRuntime(ctx)
	if err != nil {
		panic(err)
	}
	if err := runtime.Start(); err != nil {
		panic(err)
	}

	a.runtime = runtime

	shell, err := desktopshell.New(runtime.Paths(), a.trayIcon)
	if err != nil {
		panic(err)
	}
	if err := shell.Start(ctx); err != nil {
		panic(err)
	}
	a.shell = shell
	a.runtime.HTTP().SetPairRequestHandler(func(request auth.PairRequestSummary) {
		if a.shell != nil {
			_ = a.shell.Show()
		}
		wruntime.EventsEmit(a.ctx, auth.PairRequestEventName, request)
	})
}

func (a *App) shutdown(context.Context) {
	if a.shell != nil {
		a.shell.Stop()
	}
	if a.runtime != nil {
		a.runtime.Stop()
	}
}

func (a *App) GetBootstrapContext() (runtimeapp.AppBootstrap, error) {
	return a.mustRuntime().Bootstrap()
}

func (a *App) ListClipboardItems(query *store.ClipboardListQuery) ([]store.ClipboardItemRecord, error) {
	if query == nil {
		query = &store.ClipboardListQuery{Limit: 50}
	}
	items, err := a.mustRuntime().Store().ListClipboardItems(*query)
	if err != nil {
		return nil, err
	}
	filtered := make([]store.ClipboardItemRecord, 0, len(items))
	for _, item := range items {
		if item.SourceKind == "mobile_web" {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered, nil
}

func (a *App) GetClipboardItem(itemID string) (store.ClipboardItemRecord, error) {
	item, err := a.mustRuntime().Store().GetClipboardItem(itemID)
	if err != nil {
		return store.ClipboardItemRecord{}, err
	}
	if item == nil {
		return store.ClipboardItemRecord{}, apierr.NotFound(fmt.Sprintf("clipboard item `%s` not found", itemID))
	}
	return *item, nil
}

func (a *App) ActivateClipboardItem(itemID string) (store.ClipboardItemRecord, error) {
	item, err := a.mustRuntime().Store().GetClipboardItem(itemID)
	if err != nil {
		return store.ClipboardItemRecord{}, err
	}
	if item == nil {
		return store.ClipboardItemRecord{}, apierr.NotFound(fmt.Sprintf("clipboard item `%s` not found", itemID))
	}
	if item.ItemKind == store.ClipboardItemKindFile {
		if item.FileMeta == nil || item.FileMeta.LocalPath == nil {
			return store.ClipboardItemRecord{}, apierr.State("file clipboard content is unavailable")
		}
		if err := a.mustRuntime().Clipboard().WriteFile(*item.FileMeta.LocalPath); err != nil {
			return store.ClipboardItemRecord{}, err
		}
	} else {
		if err := a.mustRuntime().Clipboard().WriteText(item.Content); err != nil {
			return store.ClipboardItemRecord{}, err
		}
	}
	activated, err := a.mustRuntime().Store().ReplaceClipboardItemWithCurrent(
		itemID,
		"desktop_local",
		optionalString(a.mustRuntime().OnlineDeviceID()),
	)
	if err != nil {
		return store.ClipboardItemRecord{}, err
	}
	a.emitClipboardEvent(clipboard.RefreshEvent{
		ItemID:       activated.ID,
		IsCurrent:    true,
		SourceKind:   activated.SourceKind,
		ObservedAtMs: nowMs(),
	})
	return activated, nil
}

func (a *App) UpdateClipboardItemPin(itemID string, pinned bool) (store.ClipboardItemRecord, error) {
	item, err := a.mustRuntime().Store().UpdateClipboardItemPin(itemID, pinned)
	if err != nil {
		return store.ClipboardItemRecord{}, err
	}
	a.emitClipboardEvent(clipboard.RefreshEvent{
		ItemID:       itemID,
		IsCurrent:    item.IsCurrent,
		SourceKind:   item.SourceKind,
		ObservedAtMs: nowMs(),
	})
	return item, nil
}

func (a *App) DeleteClipboardItem(itemID string) error {
	if err := a.mustRuntime().Store().SoftDeleteClipboardItem(itemID); err != nil {
		return err
	}
	a.emitClipboardEvent(clipboard.RefreshEvent{
		ItemID:       itemID,
		ObservedAtMs: nowMs(),
	})
	return nil
}

func (a *App) ClearClipboardHistory() (int, error) {
	count, err := a.mustRuntime().Store().ClearClipboardHistory()
	if err != nil {
		return 0, err
	}
	a.emitClipboardEvent(clipboard.RefreshEvent{
		ObservedAtMs: nowMs(),
	})
	return count, nil
}

func (a *App) RotateSessionToken() (auth.SessionSnapshot, error) {
	return a.mustRuntime().RotateSession()
}

func (a *App) GetConnectivityReport() (runtimeapp.ConnectivityReport, error) {
	return a.mustRuntime().GetConnectivityReport()
}

func (a *App) ListOnlineDevices() []httpserver.OnlineDevice {
	return a.mustRuntime().HTTP().ListOnlineDevices(a.mustRuntime().OnlineDeviceID())
}

func (a *App) ListLinkedWebDevices() ([]auth.LinkedDeviceSummary, error) {
	return a.mustRuntime().HTTP().ListLinkedDevices()
}

func (a *App) RemoveLinkedWebDevice(deviceID string) error {
	return a.mustRuntime().HTTP().RemoveLinkedDevice(deviceID)
}

func (a *App) ListPairRequests() []auth.PairRequestSummary {
	return a.mustRuntime().HTTP().ListPairRequests()
}

func (a *App) ApprovePairRequest(requestID string) (auth.PairRequestSummary, error) {
	return a.mustRuntime().HTTP().ApprovePairRequest(requestID)
}

func (a *App) RejectPairRequest(requestID string) (auth.PairRequestSummary, error) {
	return a.mustRuntime().HTTP().RejectPairRequest(requestID)
}

func (a *App) SyncClipboardItem(itemID string, targetDeviceIDs []string, syncAll bool) (httpserver.SyncClipboardResponse, error) {
	item, err := a.mustRuntime().Store().GetClipboardItem(itemID)
	if err != nil {
		return httpserver.SyncClipboardResponse{}, err
	}
	if item == nil {
		return httpserver.SyncClipboardResponse{}, apierr.NotFound(fmt.Sprintf("clipboard item `%s` not found", itemID))
	}

	return a.mustRuntime().HTTP().SyncClipboardItem(*item, a.mustRuntime().OnlineDeviceID(), targetDeviceIDs, syncAll)
}

func (a *App) ReceiveClipboardFile(itemID string) (store.ClipboardItemRecord, error) {
	item, err := a.mustRuntime().HTTP().ReceiveClipboardFile(itemID, a.mustRuntime().Paths().DesktopReceiveDir)
	if err != nil {
		return store.ClipboardItemRecord{}, err
	}
	if item.FileMeta == nil || item.FileMeta.LocalPath == nil {
		return store.ClipboardItemRecord{}, apierr.State("received file is unavailable")
	}
	if err := a.mustRuntime().Clipboard().WriteFile(*item.FileMeta.LocalPath); err != nil {
		return store.ClipboardItemRecord{}, err
	}
	a.emitClipboardEvent(clipboard.RefreshEvent{
		ItemID:       item.ID,
		IsCurrent:    item.IsCurrent,
		SourceKind:   item.SourceKind,
		ObservedAtMs: nowMs(),
	})
	return item, nil
}

func (a *App) CopyText(text string) error {
	return a.mustRuntime().Clipboard().WriteText(text)
}

func (a *App) OpenURL(url string) {
	if a.ctx != nil {
		wruntime.BrowserOpenURL(a.ctx, url)
	}
}

func (a *App) HideDesktopApp() error {
	return a.mustShell().Hide()
}

func (a *App) ShowDesktopApp() error {
	return a.mustShell().Show()
}

func (a *App) GetDesktopPinned() bool {
	return a.mustShell().IsPinned()
}

func (a *App) SetDesktopPinned(pinned bool) bool {
	return a.mustShell().SetPinned(pinned)
}

func (a *App) GetDesktopSettings() (settings.DesktopSettings, error) {
	return a.mustShell().Settings()
}

func (a *App) UpdateDesktopSettings(input settings.DesktopSettings) (settings.DesktopSettings, error) {
	return a.mustShell().UpdateSettings(input)
}

func (a *App) mustRuntime() *runtimeapp.AppRuntime {
	if a.runtime == nil {
		panic("application runtime is not ready")
	}
	return a.runtime
}

func (a *App) mustShell() *desktopshell.Manager {
	if a.shell == nil {
		panic("desktop shell is not ready")
	}
	return a.shell
}

func (a *App) emitClipboardEvent(event clipboard.RefreshEvent) {
	a.mustRuntime().HTTP().PublishRefresh("clipboard", optionalString(event.ItemID))
	if a.ctx != nil {
		wruntime.EventsEmit(a.ctx, clipboard.EventName, event)
	}
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}
	copy := value
	return &copy
}

func nowMs() int64 {
	return time.Now().UnixMilli()
}
