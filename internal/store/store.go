package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"localShareGo/internal/apierr"
)

const defaultPreviewCharacterLimit = 120

type Store struct {
	path string
	mu   sync.RWMutex
	data persistentState
}

func New(path string) (*Store, error) {
	store := &Store{
		path: path,
		data: persistentState{
			Devices:        []DeviceRecord{},
			Sessions:       []SessionRecord{},
			ClipboardItems: []ClipboardItemRecord{},
		},
	}

	if err := store.load(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) Status() PersistenceStatus {
	return PersistenceStatus{
		DatabasePath:      s.path,
		MigrationsEnabled: false,
		SchemaVersion:     1,
		Ready:             true,
	}
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	content, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if len(content) == 0 {
		return nil
	}

	var next persistentState
	if err := json.Unmarshal(content, &next); err != nil {
		return err
	}

	s.data = next
	return nil
}

func (s *Store) saveLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}

	content, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}

	tempPath := s.path + ".tmp"
	if err := os.WriteFile(tempPath, content, 0o644); err != nil {
		return err
	}

	return os.Rename(tempPath, s.path)
}

func (s *Store) UpsertDevice(name string) (DeviceRecord, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return DeviceRecord{}, apierr.InvalidArgument("device name cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := nowMs()
	for index, item := range s.data.Devices {
		if item.Name == trimmed {
			item.UpdatedAt = now
			s.data.Devices[index] = item
			return item, s.saveLocked()
		}
	}

	device := DeviceRecord{
		ID:        uuid.NewString(),
		Name:      trimmed,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.data.Devices = append(s.data.Devices, device)
	return device, s.saveLocked()
}

func (s *Store) GetDevice(id string) *DeviceRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.data.Devices {
		if item.ID == id {
			copy := item
			return &copy
		}
	}
	return nil
}

func (s *Store) CreateSession(tokenHash string, expiresAt int64) (SessionRecord, error) {
	if strings.TrimSpace(tokenHash) == "" {
		return SessionRecord{}, apierr.InvalidArgument("token hash cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := nowMs()
	expireSessionsLocked(&s.data.Sessions, now)
	for index, item := range s.data.Sessions {
		if item.Status == "active" {
			rotatedAt := now
			item.Status = "expired"
			item.RotatedAt = &rotatedAt
			s.data.Sessions[index] = item
		}
	}

	session := SessionRecord{
		ID:        uuid.NewString(),
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		Status:    "active",
		CreatedAt: now,
	}
	s.data.Sessions = append(s.data.Sessions, session)
	return session, s.saveLocked()
}

func (s *Store) ReplacePendingSession(tokenHash string, createdAt int64) (SessionRecord, error) {
	if strings.TrimSpace(tokenHash) == "" {
		return SessionRecord{}, apierr.InvalidArgument("token hash cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, createdAt)
	for index, item := range s.data.Sessions {
		if item.Status != "pending" {
			continue
		}
		item.Status = "rotated"
		item.RotatedAt = &createdAt
		s.data.Sessions[index] = item
	}

	session := SessionRecord{
		ID:        uuid.NewString(),
		TokenHash: tokenHash,
		ExpiresAt: 0,
		Status:    "pending",
		CreatedAt: createdAt,
	}
	s.data.Sessions = append(s.data.Sessions, session)
	return session, s.saveLocked()
}

func (s *Store) GetPendingSession(sessionID string, now int64) (*SessionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, now)
	for _, item := range s.data.Sessions {
		if item.ID != sessionID || item.Status != "pending" {
			continue
		}
		copy := item
		if err := s.saveLocked(); err != nil {
			return nil, err
		}
		return &copy, nil
	}

	if err := s.saveLocked(); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Store) GetCurrentSession(now int64) (*SessionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, now)

	var current *SessionRecord
	for _, item := range s.data.Sessions {
		if item.Status != "active" || item.ExpiresAt <= now {
			continue
		}
		copy := item
		if current == nil || copy.CreatedAt > current.CreatedAt {
			current = &copy
		}
	}

	if err := s.saveLocked(); err != nil {
		return nil, err
	}

	return current, nil
}

func (s *Store) GetSessionByHash(tokenHash string, now int64) (*SessionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, now)
	var current *SessionRecord
	for _, item := range s.data.Sessions {
		if item.TokenHash != tokenHash {
			continue
		}
		if item.Status == "pending" {
			copy := item
			current = &copy
			break
		}
		if item.Status != "active" || item.ExpiresAt <= now {
			continue
		}
		copy := item
		if current == nil || copy.CreatedAt > current.CreatedAt {
			current = &copy
		}
	}

	if err := s.saveLocked(); err != nil {
		return nil, err
	}

	return current, nil
}

func (s *Store) ActivateSession(sessionID string, expiresAt, activatedAt int64) (SessionRecord, error) {
	if strings.TrimSpace(sessionID) == "" {
		return SessionRecord{}, apierr.InvalidArgument("session id cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, activatedAt)
	for index, item := range s.data.Sessions {
		if item.ID != sessionID {
			continue
		}
		if item.Status != "pending" {
			return SessionRecord{}, apierr.InvalidArgument("session is not pending")
		}
		item.Status = "active"
		item.ExpiresAt = expiresAt
		item.ActivatedAt = &activatedAt
		s.data.Sessions[index] = item
		return item, s.saveLocked()
	}

	return SessionRecord{}, apierr.NotFound(fmt.Sprintf("pending session `%s` not found", sessionID))
}

func (s *Store) RotateSession(currentSessionID, nextTokenHash string, expiresAt, rotatedAt int64) (SessionRecord, error) {
	if strings.TrimSpace(currentSessionID) == "" {
		return SessionRecord{}, apierr.InvalidArgument("session id cannot be empty")
	}
	if strings.TrimSpace(nextTokenHash) == "" {
		return SessionRecord{}, apierr.InvalidArgument("token hash cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, rotatedAt)
	found := false
	for index, item := range s.data.Sessions {
		if item.Status == "active" {
			item.Status = "rotated"
			item.RotatedAt = &rotatedAt
			if item.ID == currentSessionID {
				found = true
			}
			s.data.Sessions[index] = item
		}
	}
	if !found {
		return SessionRecord{}, apierr.NotFound(fmt.Sprintf("active session `%s` not found", currentSessionID))
	}

	session := SessionRecord{
		ID:        uuid.NewString(),
		TokenHash: nextTokenHash,
		ExpiresAt: expiresAt,
		Status:    "active",
		CreatedAt: rotatedAt,
	}
	s.data.Sessions = append(s.data.Sessions, session)
	return session, s.saveLocked()
}

func (s *Store) SaveClipboardItem(input SaveClipboardInput, dedupWindowMs, maxTextBytes int) (SaveClipboardResult, error) {
	content := strings.TrimSpace(input.Content)
	if content == "" {
		return SaveClipboardResult{}, apierr.InvalidArgument("clipboard content cannot be empty")
	}
	if len([]byte(content)) > maxTextBytes {
		return SaveClipboardResult{}, apierr.InvalidArgument("clipboard content exceeds max size")
	}
	if input.SourceDeviceID != nil && s.GetDevice(*input.SourceDeviceID) == nil {
		return SaveClipboardResult{}, apierr.InvalidArgument("source device does not exist")
	}

	now := nowMs()
	hash := hashText(content)
	preview := buildPreview(content)

	s.mu.Lock()
	defer s.mu.Unlock()

	if input.MarkCurrent {
		for index, item := range s.data.ClipboardItems {
			if item.DeletedAt == nil && item.IsCurrent {
				item.IsCurrent = false
				item.UpdatedAt = now
				s.data.ClipboardItems[index] = item
			}
		}
	}

	dedupThreshold := now - int64(maxInt(dedupWindowMs, 0))
	for index, item := range s.data.ClipboardItems {
		if item.DeletedAt != nil || item.Hash != hash || item.CreatedAt < dedupThreshold {
			continue
		}
		item.Pinned = item.Pinned || input.Pinned
		if input.MarkCurrent {
			item.IsCurrent = true
		}
		item.UpdatedAt = now
		s.data.ClipboardItems[index] = item
		if err := s.saveLocked(); err != nil {
			return SaveClipboardResult{}, err
		}
		return SaveClipboardResult{
			Item:           item,
			Created:        false,
			ReusedExisting: true,
		}, nil
	}

	record := ClipboardItemRecord{
		ID:             uuid.NewString(),
		Content:        content,
		ContentType:    "text/plain",
		Hash:           hash,
		Preview:        preview,
		CharCount:      len([]rune(content)),
		SourceKind:     input.SourceKind,
		SourceDeviceID: input.SourceDeviceID,
		Pinned:         input.Pinned,
		IsCurrent:      input.MarkCurrent,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.data.ClipboardItems = append(s.data.ClipboardItems, record)
	if err := s.saveLocked(); err != nil {
		return SaveClipboardResult{}, err
	}
	return SaveClipboardResult{
		Item:           record,
		Created:        true,
		ReusedExisting: false,
	}, nil
}

func (s *Store) ListClipboardItems(query ClipboardListQuery) ([]ClipboardItemRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	search := ""
	if query.Search != nil {
		search = strings.ToLower(strings.TrimSpace(*query.Search))
	}

	items := make([]ClipboardItemRecord, 0, len(s.data.ClipboardItems))
	for _, item := range s.data.ClipboardItems {
		if !query.IncludeDeleted && item.DeletedAt != nil {
			continue
		}
		if query.PinnedOnly && !item.Pinned {
			continue
		}
		if query.CreatedBefore != nil && query.BeforeID != nil {
			if !(item.CreatedAt < *query.CreatedBefore || (item.CreatedAt == *query.CreatedBefore && item.ID < *query.BeforeID)) {
				continue
			}
		}
		if search != "" {
			content := strings.ToLower(item.Content)
			preview := strings.ToLower(item.Preview)
			if !strings.Contains(content, search) && !strings.Contains(preview, search) {
				continue
			}
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Pinned != items[j].Pinned {
			return items[i].Pinned
		}
		if items[i].CreatedAt != items[j].CreatedAt {
			return items[i].CreatedAt > items[j].CreatedAt
		}
		return items[i].ID > items[j].ID
	})

	if len(items) > limit {
		items = items[:limit]
	}

	return append([]ClipboardItemRecord(nil), items...), nil
}

func (s *Store) GetClipboardItem(itemID string) (*ClipboardItemRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.data.ClipboardItems {
		if item.ID == itemID && item.DeletedAt == nil {
			copy := item
			return &copy, nil
		}
	}
	return nil, nil
}

func (s *Store) ActivateClipboardItem(itemID string) (ClipboardItemRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := nowMs()
	found := -1
	for index, item := range s.data.ClipboardItems {
		if item.DeletedAt == nil && item.IsCurrent {
			item.IsCurrent = false
			item.UpdatedAt = now
			s.data.ClipboardItems[index] = item
		}
		if item.ID == itemID && item.DeletedAt == nil {
			found = index
		}
	}
	if found < 0 {
		return ClipboardItemRecord{}, apierr.NotFound(fmt.Sprintf("clipboard item `%s` not found", itemID))
	}

	item := s.data.ClipboardItems[found]
	item.IsCurrent = true
	item.UpdatedAt = now
	s.data.ClipboardItems[found] = item
	return item, s.saveLocked()
}

func (s *Store) ReplaceClipboardItemWithCurrent(itemID, sourceKind string, sourceDeviceID *string) (ClipboardItemRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := nowMs()
	found := -1
	for index, item := range s.data.ClipboardItems {
		if item.DeletedAt == nil && item.IsCurrent {
			item.IsCurrent = false
			item.UpdatedAt = now
			s.data.ClipboardItems[index] = item
		}
		if item.ID == itemID && item.DeletedAt == nil {
			found = index
		}
	}
	if found < 0 {
		return ClipboardItemRecord{}, apierr.NotFound(fmt.Sprintf("clipboard item `%s` not found", itemID))
	}

	replaced := s.data.ClipboardItems[found]
	replaced.DeletedAt = &now
	replaced.IsCurrent = false
	replaced.UpdatedAt = now
	s.data.ClipboardItems[found] = replaced

	record := ClipboardItemRecord{
		ID:             uuid.NewString(),
		Content:        replaced.Content,
		ContentType:    replaced.ContentType,
		Hash:           replaced.Hash,
		Preview:        replaced.Preview,
		CharCount:      replaced.CharCount,
		SourceKind:     sourceKind,
		SourceDeviceID: sourceDeviceID,
		Pinned:         replaced.Pinned,
		IsCurrent:      true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.data.ClipboardItems = append(s.data.ClipboardItems, record)
	return record, s.saveLocked()
}

func (s *Store) UpdateClipboardItemPin(itemID string, pinned bool) (ClipboardItemRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := nowMs()
	for index, item := range s.data.ClipboardItems {
		if item.ID != itemID || item.DeletedAt != nil {
			continue
		}
		item.Pinned = pinned
		item.UpdatedAt = now
		s.data.ClipboardItems[index] = item
		return item, s.saveLocked()
	}

	return ClipboardItemRecord{}, apierr.NotFound(fmt.Sprintf("clipboard item `%s` not found", itemID))
}

func (s *Store) SoftDeleteClipboardItem(itemID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := nowMs()
	for index, item := range s.data.ClipboardItems {
		if item.ID != itemID || item.DeletedAt != nil {
			continue
		}
		item.DeletedAt = &now
		item.IsCurrent = false
		item.UpdatedAt = now
		s.data.ClipboardItems[index] = item
		return s.saveLocked()
	}

	return apierr.NotFound(fmt.Sprintf("clipboard item `%s` not found", itemID))
}

func (s *Store) ClearClipboardHistory() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := nowMs()
	count := 0
	for index, item := range s.data.ClipboardItems {
		if item.DeletedAt != nil {
			continue
		}
		item.DeletedAt = &now
		item.IsCurrent = false
		item.UpdatedAt = now
		s.data.ClipboardItems[index] = item
		count++
	}

	return count, s.saveLocked()
}

func CloneSummary(item ClipboardItemRecord) ClipboardItemSummary {
	return ClipboardItemSummary{
		ID:             item.ID,
		Preview:        item.Preview,
		CharCount:      item.CharCount,
		SourceKind:     item.SourceKind,
		SourceDeviceID: item.SourceDeviceID,
		Pinned:         item.Pinned,
		IsCurrent:      item.IsCurrent,
		DeletedAt:      item.DeletedAt,
		CreatedAt:      item.CreatedAt,
		UpdatedAt:      item.UpdatedAt,
	}
}

func expireSessionsLocked(items *[]SessionRecord, now int64) {
	for index, item := range *items {
		if item.Status == "active" && item.ExpiresAt <= now {
			item.Status = "expired"
			item.RotatedAt = &now
			(*items)[index] = item
		}
	}
}

func hashText(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func buildPreview(content string) string {
	normalized := strings.Join(strings.Fields(content), " ")
	runes := []rune(normalized)
	if len(runes) <= defaultPreviewCharacterLimit {
		return normalized
	}
	return string(runes[:defaultPreviewCharacterLimit]) + "..."
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
