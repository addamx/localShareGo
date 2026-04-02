package clipboard

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"localShareGo/internal/apierr"
	"localShareGo/internal/fileclip"
	"localShareGo/internal/store"
)

const defaultDedupWindowMs = 1600

type Service struct {
	store            *store.Store
	pollIntervalMs   int
	maxTextBytes     int
	dedupWindowMs    int
	sourceDeviceID   *string
	onRefresh        func(RefreshEvent)
	mu               sync.RWMutex
	running          bool
	stop             chan struct{}
	lastProcessedKey string
	lastFailedKey    string
	lastFailedAtMs   int64
}

func New(store *store.Store, pollIntervalMs, maxTextBytes int, sourceDeviceID *string, onRefresh func(RefreshEvent)) *Service {
	return &Service{
		store:          store,
		pollIntervalMs: pollIntervalMs,
		maxTextBytes:   maxTextBytes,
		dedupWindowMs:  maxInt(defaultDedupWindowMs, pollIntervalMs*2),
		sourceDeviceID: sourceDeviceID,
		onRefresh:      onRefresh,
	}
}

func (c *Service) Status() ClipboardStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return ClipboardStatus{
		Mode:                "polling",
		PollIntervalMs:      c.pollIntervalMs,
		DedupWindowMs:       c.dedupWindowMs,
		MaxTextBytes:        c.maxTextBytes,
		CurrentItemTracking: true,
		Running:             c.running,
		SubscriberCount:     0,
		RefreshEventTopic:   EventName,
	}
}

func (c *Service) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.running {
		return nil
	}
	c.stop = make(chan struct{})
	c.running = true

	go c.runLoop(c.stop)
	return nil
}

func (c *Service) StopLoop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.running {
		return
	}
	close(c.stop)
	c.running = false
}

func (c *Service) WriteText(text string) error {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return apierr.InvalidArgument("clipboard content cannot be empty")
	}
	if err := writeClipboardText(trimmed); err != nil {
		return apierr.State(err.Error())
	}

	c.mu.Lock()
	c.lastProcessedKey = textClipboardKey(trimmed)
	c.lastFailedKey = ""
	c.lastFailedAtMs = 0
	c.mu.Unlock()
	return nil
}

func (c *Service) WriteFile(path string) error {
	cleanPath := filepath.Clean(path)
	meta, err := fileclip.InspectPath(cleanPath)
	if err != nil {
		return apierr.State(err.Error())
	}
	if err := fileclip.WriteClipboardFile(cleanPath); err != nil {
		return apierr.State(err.Error())
	}

	c.mu.Lock()
	c.lastProcessedKey = fileClipboardKey(cleanPath, meta)
	c.lastFailedKey = ""
	c.lastFailedAtMs = 0
	c.mu.Unlock()
	return nil
}

func (c *Service) runLoop(stop <-chan struct{}) {
	ticker := time.NewTicker(time.Duration(maxInt(c.pollIntervalMs, 150)) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			c.tick()
		}
	}
}

func (c *Service) tick() {
	if fileItem, ok, err := fileclip.ReadClipboardFile(); err == nil && ok {
		c.tickFile(fileItem)
		return
	}

	text, err := readClipboardText()
	if err != nil || strings.TrimSpace(text) == "" {
		return
	}

	key := textClipboardKey(text)
	now := nowMs()

	c.mu.RLock()
	shouldSkip := c.lastProcessedKey == key || (c.lastFailedKey == key && now-c.lastFailedAtMs < int64(maxInt(c.dedupWindowMs*2, 3000)))
	c.mu.RUnlock()
	if shouldSkip {
		return
	}

	if len([]byte(text)) > c.maxTextBytes {
		c.mu.Lock()
		c.lastFailedKey = key
		c.lastFailedAtMs = now
		c.mu.Unlock()
		return
	}

	result, err := c.store.SaveClipboardItem(store.SaveClipboardInput{
		ItemKind:       store.ClipboardItemKindText,
		Content:        text,
		ContentType:    "text/plain",
		Hash:           key,
		Preview:        buildTextPreview(text),
		CharCount:      len([]rune(text)),
		SourceKind:     "desktop_local",
		SourceDeviceID: c.sourceDeviceID,
		Pinned:         false,
		MarkCurrent:    true,
	}, c.dedupWindowMs, c.maxTextBytes)
	if err != nil {
		c.mu.Lock()
		c.lastFailedKey = key
		c.lastFailedAtMs = now
		c.mu.Unlock()
		return
	}

	c.mu.Lock()
	c.lastProcessedKey = key
	c.lastFailedKey = ""
	c.lastFailedAtMs = 0
	c.mu.Unlock()

	if c.onRefresh != nil {
		c.onRefresh(RefreshEvent{
			ItemID:         result.Item.ID,
			Created:        result.Created,
			ReusedExisting: result.ReusedExisting,
			IsCurrent:      result.Item.IsCurrent,
			SourceKind:     result.Item.SourceKind,
			ObservedAtMs:   nowMs(),
		})
	}
}

func (c *Service) tickFile(item fileclip.ClipboardFile) {
	key := fileClipboardKey(item.Path, item.Metadata)
	now := nowMs()

	c.mu.RLock()
	shouldSkip := c.lastProcessedKey == key || (c.lastFailedKey == key && now-c.lastFailedAtMs < int64(maxInt(c.dedupWindowMs*2, 3000)))
	c.mu.RUnlock()
	if shouldSkip {
		return
	}

	result, err := c.store.SaveClipboardItem(store.SaveClipboardInput{
		ItemKind:    store.ClipboardItemKindFile,
		Content:     "",
		ContentType: item.Metadata.MIMEType,
		Hash:        key,
		Preview:     item.Metadata.FileName,
		CharCount:   0,
		FileMeta: &store.ClipboardFileMeta{
			FileName:         item.Metadata.FileName,
			Extension:        item.Metadata.Extension,
			MIMEType:         item.Metadata.MIMEType,
			SizeBytes:        item.Metadata.SizeBytes,
			ThumbnailDataURL: item.Metadata.ThumbnailDataURL,
			TransferState:    store.TransferStateReceived,
			ProgressPercent:  100,
			LocalPath:        &item.Path,
			DownloadedAt:     optionalNowMs(now),
		},
		SourceKind:     "desktop_local",
		SourceDeviceID: c.sourceDeviceID,
		Pinned:         false,
		MarkCurrent:    true,
	}, c.dedupWindowMs, c.maxTextBytes)
	if err != nil {
		c.mu.Lock()
		c.lastFailedKey = key
		c.lastFailedAtMs = now
		c.mu.Unlock()
		return
	}

	c.mu.Lock()
	c.lastProcessedKey = key
	c.lastFailedKey = ""
	c.lastFailedAtMs = 0
	c.mu.Unlock()

	if c.onRefresh != nil {
		c.onRefresh(RefreshEvent{
			ItemID:         result.Item.ID,
			Created:        result.Created,
			ReusedExisting: result.ReusedExisting,
			IsCurrent:      result.Item.IsCurrent,
			SourceKind:     result.Item.SourceKind,
			ObservedAtMs:   nowMs(),
		})
	}
}

func optionalNowMs(value int64) *int64 {
	return &value
}

func hashText(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func textClipboardKey(value string) string {
	return "text:" + hashText(value)
}

func fileClipboardKey(path string, meta fileclip.Metadata) string {
	return "file:" + hashText(strings.Join([]string{
		path,
		meta.FileName,
		meta.Extension,
		meta.MIMEType,
		fmt.Sprintf("%d", meta.SizeBytes),
	}, "\n"))
}

func buildTextPreview(content string) string {
	normalized := strings.Join(strings.Fields(content), " ")
	runes := []rune(normalized)
	if len(runes) <= 120 {
		return normalized
	}
	return string(runes[:120]) + "..."
}

func nowMs() int64 {
	return time.Now().UnixMilli()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
