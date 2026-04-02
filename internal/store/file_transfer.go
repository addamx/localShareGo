package store

import (
	"fmt"

	"localShareGo/internal/apierr"
)

func (s *Store) UpdateClipboardFileMeta(itemID string, fileMeta ClipboardFileMeta) (ClipboardItemRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := nowMs()
	for index, item := range s.data.ClipboardItems {
		if item.ID != itemID || item.DeletedAt != nil {
			continue
		}
		if item.ItemKind != ClipboardItemKindFile {
			return ClipboardItemRecord{}, apierr.InvalidArgument("clipboard item is not a file")
		}

		item.FileMeta = cloneFileMeta(&fileMeta)
		normalizeClipboardItemRecord(&item)
		item.UpdatedAt = now
		s.data.ClipboardItems[index] = item
		return item, s.saveLocked()
	}

	return ClipboardItemRecord{}, apierr.NotFound(fmt.Sprintf("clipboard item `%s` not found", itemID))
}

func (s *Store) UpdateClipboardFileTransfer(itemID string, input UpdateClipboardTransferInput) (ClipboardItemRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := nowMs()
	for index, item := range s.data.ClipboardItems {
		if item.ID != itemID || item.DeletedAt != nil {
			continue
		}
		if item.ItemKind != ClipboardItemKindFile || item.FileMeta == nil {
			return ClipboardItemRecord{}, apierr.InvalidArgument("clipboard item is not a file")
		}

		item.FileMeta.TransferState = input.TransferState
		item.FileMeta.ProgressPercent = input.ProgressPercent
		item.FileMeta.LocalPath = input.LocalPath
		item.FileMeta.DownloadedAt = input.DownloadedAt
		normalizeClipboardItemRecord(&item)
		item.UpdatedAt = now
		s.data.ClipboardItems[index] = item
		return item, s.saveLocked()
	}

	return ClipboardItemRecord{}, apierr.NotFound(fmt.Sprintf("clipboard item `%s` not found", itemID))
}
