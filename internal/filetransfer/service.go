package filetransfer

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"localShareGo/internal/apierr"
	"localShareGo/internal/store"
)

type ProgressEvent struct {
	ItemID           string  `json:"itemId"`
	Status           string  `json:"status"`
	ProgressPercent  int     `json:"progressPercent"`
	BytesTransferred int64   `json:"bytesTransferred"`
	BytesTotal       int64   `json:"bytesTotal"`
	ErrorMessage     *string `json:"errorMessage"`
}

type Publisher func(ProgressEvent)

type Service struct {
	stagingDir string
	publish    Publisher
}

func New(stagingDir string, publish Publisher) (*Service, error) {
	if err := os.MkdirAll(stagingDir, 0o755); err != nil {
		return nil, err
	}

	return &Service{
		stagingDir: stagingDir,
		publish:    publish,
	}, nil
}

func (s *Service) CreateFileItem(dataStore *store.Store, sourceKind string, sourceDeviceID *string, filename, declaredContentType string, source io.Reader) (store.SaveClipboardResult, error) {
	displayName := normalizeFilename(filename)
	stagedPath, err := nextAvailablePath(s.stagingDir, displayName)
	if err != nil {
		return store.SaveClipboardResult{}, err
	}

	stagedFile, err := os.Create(stagedPath)
	if err != nil {
		return store.SaveClipboardResult{}, err
	}

	cleanup := true
	defer func() {
		_ = stagedFile.Close()
		if cleanup {
			_ = os.Remove(stagedPath)
		}
	}()

	hash := sha256.New()
	var size int64
	prefix := make([]byte, 0, 512)
	buffer := make([]byte, 32*1024)

	for {
		n, readErr := source.Read(buffer)
		if n > 0 {
			chunk := buffer[:n]
			if len(prefix) < 512 {
				need := 512 - len(prefix)
				if need > n {
					need = n
				}
				prefix = append(prefix, chunk[:need]...)
			}
			if _, err := stagedFile.Write(chunk); err != nil {
				return store.SaveClipboardResult{}, err
			}
			_, _ = hash.Write(chunk)
			size += int64(n)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return store.SaveClipboardResult{}, readErr
		}
	}

	if err := stagedFile.Close(); err != nil {
		return store.SaveClipboardResult{}, err
	}

	contentType := resolveContentType(filename, declaredContentType, prefix)
	thumbnail := makeThumbnailDataURL(stagedPath, contentType)
	pathValue := stagedPath
	downloadedAt := nowMs()
	meta := &store.ClipboardFileMeta{
		FileName:         displayName,
		Extension:        normalizeExtension(displayName),
		MIMEType:         contentType,
		SizeBytes:        size,
		ThumbnailDataURL: thumbnail,
		TransferState:    store.TransferStateReceived,
		ProgressPercent:  100,
		LocalPath:        &pathValue,
		DownloadedAt:     &downloadedAt,
	}

	result, err := dataStore.SaveClipboardItem(store.SaveClipboardInput{
		ItemKind:       store.ClipboardItemKindFile,
		ContentType:    contentType,
		Hash:           hex.EncodeToString(hash.Sum(nil)),
		Preview:        displayName,
		CharCount:      0,
		FileMeta:       meta,
		SourceKind:     sourceKind,
		SourceDeviceID: sourceDeviceID,
		Pinned:         false,
		MarkCurrent:    false,
	}, 0, 0)
	if err != nil {
		return store.SaveClipboardResult{}, err
	}

	cleanup = result.ReusedExisting

	s.publishProgress(
		result.Item.ID,
		result.Item.FileMeta.TransferState,
		result.Item.FileMeta.ProgressPercent,
		result.Item.FileMeta.SizeBytes,
		result.Item.FileMeta.SizeBytes,
		nil,
	)
	return result, nil
}

func (s *Service) PrepareReceive(dataStore *store.Store, itemID string) (store.ClipboardItemRecord, error) {
	item, err := s.loadFileItem(dataStore, itemID)
	if err != nil {
		return store.ClipboardItemRecord{}, err
	}

	updated, err := dataStore.UpdateClipboardFileTransfer(itemID, store.UpdateClipboardTransferInput{
		TransferState:   store.TransferStateReceiving,
		ProgressPercent: 0,
		LocalPath:       item.FileMeta.LocalPath,
		DownloadedAt:    nil,
	})
	if err != nil {
		return store.ClipboardItemRecord{}, err
	}

	s.publishProgress(itemID, store.TransferStateReceiving, 0, 0, updated.FileMeta.SizeBytes, nil)
	return updated, nil
}

func (s *Service) ServeContent(dataStore *store.Store, writer http.ResponseWriter, request *http.Request, itemID string) error {
	item, err := s.loadFileItem(dataStore, itemID)
	if err != nil {
		return err
	}

	localPath := ""
	if item.FileMeta != nil && item.FileMeta.LocalPath != nil {
		localPath = strings.TrimSpace(*item.FileMeta.LocalPath)
	}
	if localPath == "" {
		return apierr.NotFound(fmt.Sprintf("file item `%s` content not staged", itemID))
	}

	file, err := os.Open(localPath)
	if err != nil {
		return apierr.NotFound(fmt.Sprintf("file item `%s` content not found", itemID))
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	total := stat.Size()
	if total <= 0 && item.FileMeta != nil {
		total = item.FileMeta.SizeBytes
	}
	if total < 0 {
		total = 0
	}

	writer.Header().Set("Content-Type", item.ContentType)
	writer.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, item.FileMeta.FileName))
	if total > 0 {
		writer.Header().Set("Content-Length", fmt.Sprintf("%d", total))
	}
	writer.WriteHeader(http.StatusOK)

	buffer := make([]byte, 32*1024)
	flusher, _ := writer.(http.Flusher)

	for {
		if request.Context().Err() != nil {
			return nil
		}

		n, readErr := file.Read(buffer)
		if n > 0 {
			if _, writeErr := writer.Write(buffer[:n]); writeErr != nil {
				return nil
			}
			if flusher != nil {
				flusher.Flush()
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil
		}
	}
	return nil
}

func (s *Service) ReceiveToDirectory(dataStore *store.Store, itemID, targetDir string) (store.ClipboardItemRecord, error) {
	item, err := s.PrepareReceive(dataStore, itemID)
	if err != nil {
		return store.ClipboardItemRecord{}, err
	}

	sourcePath := ""
	if item.FileMeta != nil && item.FileMeta.LocalPath != nil {
		sourcePath = strings.TrimSpace(*item.FileMeta.LocalPath)
	}
	if sourcePath == "" {
		return store.ClipboardItemRecord{}, apierr.NotFound(fmt.Sprintf("file item `%s` content not staged", itemID))
	}

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		s.markFailed(dataStore, itemID, 0, item.FileMeta.SizeBytes, &sourcePath, err.Error())
		return store.ClipboardItemRecord{}, err
	}

	destinationPath, err := nextAvailablePath(targetDir, item.FileMeta.FileName)
	if err != nil {
		s.markFailed(dataStore, itemID, 0, item.FileMeta.SizeBytes, &sourcePath, err.Error())
		return store.ClipboardItemRecord{}, err
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		s.markFailed(dataStore, itemID, 0, item.FileMeta.SizeBytes, &sourcePath, err.Error())
		return store.ClipboardItemRecord{}, apierr.NotFound(fmt.Sprintf("file item `%s` content not found", itemID))
	}
	defer source.Close()

	destination, err := os.Create(destinationPath)
	if err != nil {
		s.markFailed(dataStore, itemID, 0, item.FileMeta.SizeBytes, &sourcePath, err.Error())
		return store.ClipboardItemRecord{}, err
	}

	total := item.FileMeta.SizeBytes
	if stat, statErr := source.Stat(); statErr == nil && stat.Size() > 0 {
		total = stat.Size()
	}

	buffer := make([]byte, 32*1024)
	var transferred int64
	lastPercent := -1

	for {
		n, readErr := source.Read(buffer)
		if n > 0 {
			if _, writeErr := destination.Write(buffer[:n]); writeErr != nil {
				_ = destination.Close()
				_ = os.Remove(destinationPath)
				s.markFailed(dataStore, itemID, transferred, total, &sourcePath, writeErr.Error())
				return store.ClipboardItemRecord{}, writeErr
			}

			transferred += int64(n)
			percent := progressPercent(transferred, total)
			if percent != lastPercent {
				lastPercent = percent
				if _, updateErr := dataStore.UpdateClipboardFileTransfer(itemID, store.UpdateClipboardTransferInput{
					TransferState:   store.TransferStateReceiving,
					ProgressPercent: percent,
					LocalPath:       &sourcePath,
					DownloadedAt:    nil,
				}); updateErr == nil {
					s.publishProgress(itemID, store.TransferStateReceiving, percent, transferred, total, nil)
				}
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			_ = destination.Close()
			_ = os.Remove(destinationPath)
			s.markFailed(dataStore, itemID, transferred, total, &sourcePath, readErr.Error())
			return store.ClipboardItemRecord{}, readErr
		}
	}

	if err := destination.Close(); err != nil {
		_ = os.Remove(destinationPath)
		s.markFailed(dataStore, itemID, transferred, total, &sourcePath, err.Error())
		return store.ClipboardItemRecord{}, err
	}

	now := nowMs()
	updated, err := dataStore.UpdateClipboardFileTransfer(itemID, store.UpdateClipboardTransferInput{
		TransferState:   store.TransferStateReceived,
		ProgressPercent: 100,
		LocalPath:       &destinationPath,
		DownloadedAt:    &now,
	})
	if err != nil {
		_ = os.Remove(destinationPath)
		s.markFailed(dataStore, itemID, transferred, total, &sourcePath, err.Error())
		return store.ClipboardItemRecord{}, err
	}

	s.publishProgress(itemID, store.TransferStateReceived, 100, total, total, nil)
	return updated, nil
}

func (s *Service) loadFileItem(dataStore *store.Store, itemID string) (store.ClipboardItemRecord, error) {
	item, err := dataStore.GetClipboardItem(itemID)
	if err != nil {
		return store.ClipboardItemRecord{}, err
	}
	if item == nil {
		return store.ClipboardItemRecord{}, apierr.NotFound(fmt.Sprintf("clipboard item `%s` not found", itemID))
	}
	if item.ItemKind != store.ClipboardItemKindFile || item.FileMeta == nil {
		return store.ClipboardItemRecord{}, apierr.InvalidArgument("clipboard item is not a file")
	}
	if item.FileMeta.LocalPath == nil || strings.TrimSpace(*item.FileMeta.LocalPath) == "" {
		return store.ClipboardItemRecord{}, apierr.NotFound(fmt.Sprintf("file item `%s` content not staged", itemID))
	}
	return *item, nil
}

func (s *Service) markFailed(dataStore *store.Store, itemID string, transferred, total int64, localPath *string, message string) {
	text := strings.TrimSpace(message)
	if text == "" {
		text = "file transfer failed"
	}
	percent := progressPercent(transferred, total)
	_, err := dataStore.UpdateClipboardFileTransfer(itemID, store.UpdateClipboardTransferInput{
		TransferState:   store.TransferStateFailed,
		ProgressPercent: percent,
		LocalPath:       localPath,
		DownloadedAt:    nil,
	})
	if err == nil {
		s.publishProgress(itemID, store.TransferStateFailed, percent, transferred, total, &text)
		return
	}
	s.publishProgress(itemID, store.TransferStateFailed, percent, transferred, total, &text)
}

func (s *Service) publishProgress(itemID, status string, percent int, transferred, total int64, message *string) {
	if s.publish == nil {
		return
	}
	s.publish(ProgressEvent{
		ItemID:           itemID,
		Status:           status,
		ProgressPercent:  percent,
		BytesTransferred: transferred,
		BytesTotal:       total,
		ErrorMessage:     message,
	})
}

func resolveContentType(filename, declared string, sample []byte) string {
	value := strings.TrimSpace(declared)
	if value != "" && !strings.EqualFold(value, "application/octet-stream") {
		return value
	}

	detected := http.DetectContentType(sample)
	if !strings.EqualFold(detected, "application/octet-stream") {
		return detected
	}

	ext := strings.TrimSpace(filepath.Ext(filename))
	if ext != "" {
		if fromExt := mime.TypeByExtension(ext); fromExt != "" {
			return fromExt
		}
	}

	return detected
}

func normalizeFilename(value string) string {
	trimmed := strings.TrimSpace(filepath.Base(value))
	if trimmed == "" || trimmed == "." || trimmed == string(filepath.Separator) {
		return "file"
	}
	return trimmed
}

func normalizeExtension(filename string) string {
	ext := strings.TrimSpace(filepath.Ext(filename))
	return strings.TrimPrefix(ext, ".")
}

func progressPercent(transferred, total int64) int {
	if total <= 0 {
		if transferred > 0 {
			return 100
		}
		return 0
	}
	percent := int((transferred * 100) / total)
	if percent < 0 {
		return 0
	}
	if percent > 100 {
		return 100
	}
	return percent
}

func dataURLPrefix(mimeType string) string {
	return fmt.Sprintf("data:%s;base64,", mimeType)
}

func encodeDataURL(mimeType string, raw []byte) *string {
	value := dataURLPrefix(mimeType) + base64.StdEncoding.EncodeToString(raw)
	return &value
}

func nextAvailablePath(dir, fileName string) (string, error) {
	base := normalizeFilename(fileName)
	extension := filepath.Ext(base)
	name := strings.TrimSuffix(base, extension)
	candidate := filepath.Join(dir, base)
	for index := 1; ; index++ {
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate, nil
		} else if err != nil {
			return "", err
		}
		candidate = filepath.Join(dir, fmt.Sprintf("%s (%d)%s", name, index, extension))
	}
}

func nowMs() int64 {
	return time.Now().UnixMilli()
}
