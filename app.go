package main

import (
	"context"
	"fmt"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"localShareGo/internal/apierr"
	"localShareGo/internal/auth"
	"localShareGo/internal/clipboard"
	"localShareGo/internal/runtimeapp"
	"localShareGo/internal/store"
)

type App struct {
	ctx     context.Context
	runtime *runtimeapp.AppRuntime
}

func NewApp() *App {
	return &App{}
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
}

func (a *App) shutdown(context.Context) {
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
	return a.mustRuntime().Store().ListClipboardItems(*query)
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
	if err := a.mustRuntime().Clipboard().WriteText(item.Content); err != nil {
		return store.ClipboardItemRecord{}, err
	}
	activated, err := a.mustRuntime().Store().ActivateClipboardItem(itemID)
	if err != nil {
		return store.ClipboardItemRecord{}, err
	}
	a.emitClipboardEvent(clipboard.RefreshEvent{
		ItemID:       itemID,
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

func (a *App) CopyText(text string) error {
	return a.mustRuntime().Clipboard().WriteText(text)
}

func (a *App) OpenURL(url string) {
	if a.ctx != nil {
		wruntime.BrowserOpenURL(a.ctx, url)
	}
}

func (a *App) mustRuntime() *runtimeapp.AppRuntime {
	if a.runtime == nil {
		panic("application runtime is not ready")
	}
	return a.runtime
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
