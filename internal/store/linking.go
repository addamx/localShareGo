package store

import (
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"

	"localShareGo/internal/apierr"
)

func normalizeSessionRecord(session *SessionRecord) {
	if session == nil {
		return
	}

	session.Kind = strings.TrimSpace(session.Kind)
	if session.Kind == "" {
		if session.Status == SessionStatusPending {
			session.Kind = SessionKindEntry
		} else {
			session.Kind = SessionKindDevice
		}
	}

	session.Status = strings.TrimSpace(session.Status)
	if session.Status == "" {
		if session.Kind == SessionKindEntry {
			session.Status = SessionStatusPending
		} else {
			session.Status = SessionStatusActive
		}
	}

	if session.Kind == SessionKindEntry {
		session.DeviceID = nil
		session.DeviceName = nil
		session.LastKnownIP = nil
	}
}

func normalizeDeviceRecord(device *DeviceRecord) {
	if device == nil {
		return
	}

	device.Kind = strings.TrimSpace(device.Kind)
	if device.Kind == "" {
		device.Kind = DeviceKindDesktop
	}
}

func (s *Store) UpsertLinkedWebDevice(deviceID, name, lastKnownIP string, now int64) (DeviceRecord, error) {
	trimmedID := strings.TrimSpace(deviceID)
	trimmedName := strings.TrimSpace(name)
	trimmedIP := strings.TrimSpace(lastKnownIP)
	if trimmedID == "" {
		return DeviceRecord{}, apierr.InvalidArgument("device id cannot be empty")
	}
	if trimmedName == "" {
		return DeviceRecord{}, apierr.InvalidArgument("device name cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.data.Devices {
		if item.ID != trimmedID {
			continue
		}
		item.Kind = DeviceKindWeb
		item.Name = trimmedName
		item.UpdatedAt = now
		item.LastSeenAt = &now
		if item.LinkedAt == nil {
			item.LinkedAt = &now
		}
		item.RevokedAt = nil
		if trimmedIP != "" {
			item.LastKnownIP = &trimmedIP
		}
		s.data.Devices[index] = item
		return item, s.saveLocked()
	}

	device := DeviceRecord{
		ID:         trimmedID,
		Kind:       DeviceKindWeb,
		Name:       trimmedName,
		CreatedAt:  now,
		UpdatedAt:  now,
		LastSeenAt: &now,
		LinkedAt:   &now,
	}
	if trimmedIP != "" {
		device.LastKnownIP = &trimmedIP
	}
	s.data.Devices = append(s.data.Devices, device)
	return device, s.saveLocked()
}

func (s *Store) TouchLinkedWebDevice(deviceID, lastKnownIP string, now int64) (DeviceRecord, error) {
	trimmedID := strings.TrimSpace(deviceID)
	if trimmedID == "" {
		return DeviceRecord{}, apierr.InvalidArgument("device id cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.data.Devices {
		if item.ID != trimmedID || item.Kind != DeviceKindWeb || item.RevokedAt != nil {
			continue
		}
		item.UpdatedAt = now
		item.LastSeenAt = &now
		if trimmedIP := strings.TrimSpace(lastKnownIP); trimmedIP != "" {
			item.LastKnownIP = &trimmedIP
		}
		s.data.Devices[index] = item
		return item, s.saveLocked()
	}

	return DeviceRecord{}, apierr.NotFound(fmt.Sprintf("linked device `%s` not found", trimmedID))
}

func (s *Store) RevokeLinkedWebDevice(deviceID string, now int64) error {
	trimmedID := strings.TrimSpace(deviceID)
	if trimmedID == "" {
		return apierr.InvalidArgument("device id cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	found := false
	for index, item := range s.data.Devices {
		if item.ID != trimmedID || item.Kind != DeviceKindWeb || item.RevokedAt != nil {
			continue
		}
		item.UpdatedAt = now
		item.RevokedAt = &now
		s.data.Devices[index] = item
		found = true
	}
	if !found {
		return apierr.NotFound(fmt.Sprintf("linked device `%s` not found", trimmedID))
	}

	for index, item := range s.data.Sessions {
		if item.Kind != SessionKindDevice || item.DeviceID == nil || *item.DeviceID != trimmedID {
			continue
		}
		if item.Status != SessionStatusActive {
			continue
		}
		item.Status = SessionStatusRevoked
		item.RevokedAt = &now
		s.data.Sessions[index] = item
	}

	return s.saveLocked()
}

func (s *Store) ListLinkedWebDevices() ([]DeviceRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := make([]DeviceRecord, 0, len(s.data.Devices))
	for _, item := range s.data.Devices {
		if item.Kind != DeviceKindWeb || item.RevokedAt != nil {
			continue
		}
		copy := item
		items = append(items, copy)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].UpdatedAt != items[j].UpdatedAt {
			return items[i].UpdatedAt > items[j].UpdatedAt
		}
		return items[i].ID < items[j].ID
	})

	if err := s.saveLocked(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) ReplacePendingEntrySession(tokenHash string, createdAt int64) (SessionRecord, error) {
	if strings.TrimSpace(tokenHash) == "" {
		return SessionRecord{}, apierr.InvalidArgument("token hash cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, createdAt)
	for index, item := range s.data.Sessions {
		if item.Kind != SessionKindEntry || item.Status != SessionStatusPending {
			continue
		}
		item.Status = SessionStatusRotated
		item.RotatedAt = &createdAt
		s.data.Sessions[index] = item
	}

	session := SessionRecord{
		ID:        uuid.NewString(),
		Kind:      SessionKindEntry,
		TokenHash: tokenHash,
		Status:    SessionStatusPending,
		CreatedAt: createdAt,
	}
	s.data.Sessions = append(s.data.Sessions, session)
	return session, s.saveLocked()
}

func (s *Store) GetPendingEntrySession(sessionID string, now int64) (*SessionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, now)
	for _, item := range s.data.Sessions {
		if item.Kind != SessionKindEntry || item.ID != sessionID || item.Status != SessionStatusPending {
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

func (s *Store) GetPendingEntrySessionByHash(tokenHash string, now int64) (*SessionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, now)
	for _, item := range s.data.Sessions {
		if item.Kind != SessionKindEntry || item.Status != SessionStatusPending || item.TokenHash != tokenHash {
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

func (s *Store) CreateOrReplaceDeviceSession(tokenHash, deviceID, deviceName, lastKnownIP string, expiresAt, now int64) (SessionRecord, error) {
	trimmedTokenHash := strings.TrimSpace(tokenHash)
	trimmedDeviceID := strings.TrimSpace(deviceID)
	trimmedDeviceName := strings.TrimSpace(deviceName)
	trimmedIP := strings.TrimSpace(lastKnownIP)
	if trimmedTokenHash == "" {
		return SessionRecord{}, apierr.InvalidArgument("token hash cannot be empty")
	}
	if trimmedDeviceID == "" {
		return SessionRecord{}, apierr.InvalidArgument("device id cannot be empty")
	}
	if trimmedDeviceName == "" {
		return SessionRecord{}, apierr.InvalidArgument("device name cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, now)
	for index, item := range s.data.Sessions {
		if item.Kind != SessionKindDevice || item.DeviceID == nil || *item.DeviceID != trimmedDeviceID {
			continue
		}
		if item.Status != SessionStatusActive {
			continue
		}
		item.Status = SessionStatusRevoked
		item.RevokedAt = &now
		s.data.Sessions[index] = item
	}

	session := SessionRecord{
		ID:         uuid.NewString(),
		Kind:       SessionKindDevice,
		TokenHash:  trimmedTokenHash,
		DeviceID:   &trimmedDeviceID,
		DeviceName: &trimmedDeviceName,
		ExpiresAt:  expiresAt,
		Status:     SessionStatusActive,
		CreatedAt:  now,
	}
	if trimmedIP != "" {
		session.LastKnownIP = &trimmedIP
	}
	s.data.Sessions = append(s.data.Sessions, session)
	return session, s.saveLocked()
}

func (s *Store) ConsumeEntrySession(entrySessionID, tokenHash, deviceID, deviceName, lastKnownIP string, expiresAt, now int64) (SessionRecord, SessionRecord, error) {
	trimmedEntryID := strings.TrimSpace(entrySessionID)
	if trimmedEntryID == "" {
		return SessionRecord{}, SessionRecord{}, apierr.InvalidArgument("entry session id cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, now)

	entryIndex := -1
	for index, item := range s.data.Sessions {
		if item.Kind != SessionKindEntry || item.ID != trimmedEntryID {
			continue
		}
		if item.Status != SessionStatusPending {
			return SessionRecord{}, SessionRecord{}, apierr.InvalidArgument("entry session is not pending")
		}
		entryIndex = index
		break
	}
	if entryIndex < 0 {
		return SessionRecord{}, SessionRecord{}, apierr.NotFound(fmt.Sprintf("entry session `%s` not found", trimmedEntryID))
	}

	entry := s.data.Sessions[entryIndex]
	entry.Status = SessionStatusConsumed
	entry.ActivatedAt = &now
	s.data.Sessions[entryIndex] = entry

	for index, item := range s.data.Sessions {
		if item.Kind != SessionKindDevice || item.DeviceID == nil || *item.DeviceID != strings.TrimSpace(deviceID) {
			continue
		}
		if item.Status != SessionStatusActive {
			continue
		}
		item.Status = SessionStatusRevoked
		item.RevokedAt = &now
		s.data.Sessions[index] = item
	}

	deviceIDCopy := strings.TrimSpace(deviceID)
	deviceNameCopy := strings.TrimSpace(deviceName)
	deviceSession := SessionRecord{
		ID:         uuid.NewString(),
		Kind:       SessionKindDevice,
		TokenHash:  strings.TrimSpace(tokenHash),
		DeviceID:   &deviceIDCopy,
		DeviceName: &deviceNameCopy,
		ExpiresAt:  expiresAt,
		Status:     SessionStatusActive,
		CreatedAt:  now,
	}
	if trimmedIP := strings.TrimSpace(lastKnownIP); trimmedIP != "" {
		deviceSession.LastKnownIP = &trimmedIP
	}
	s.data.Sessions = append(s.data.Sessions, deviceSession)

	if err := s.saveLocked(); err != nil {
		return SessionRecord{}, SessionRecord{}, err
	}
	return entry, deviceSession, nil
}

func (s *Store) GetActiveDeviceSessionByHash(tokenHash string, now int64) (*SessionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, now)
	for _, item := range s.data.Sessions {
		if item.Kind != SessionKindDevice || item.TokenHash != tokenHash || item.Status != SessionStatusActive || item.ExpiresAt <= now {
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

func (s *Store) GetActiveDeviceSessionByDeviceID(deviceID string, now int64) (*SessionRecord, error) {
	trimmedID := strings.TrimSpace(deviceID)
	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, now)
	var current *SessionRecord
	for _, item := range s.data.Sessions {
		if item.Kind != SessionKindDevice || item.Status != SessionStatusActive || item.DeviceID == nil || *item.DeviceID != trimmedID || item.ExpiresAt <= now {
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

func (s *Store) RenewDeviceSession(sessionID string, expiresAt, now int64) (SessionRecord, error) {
	trimmedID := strings.TrimSpace(sessionID)
	if trimmedID == "" {
		return SessionRecord{}, apierr.InvalidArgument("session id cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, now)
	for index, item := range s.data.Sessions {
		if item.Kind != SessionKindDevice || item.ID != trimmedID {
			continue
		}
		if item.Status != SessionStatusActive {
			return SessionRecord{}, apierr.InvalidArgument("device session is not active")
		}
		item.ExpiresAt = expiresAt
		item.UpdatedLastKnownIP(now, item.LastKnownIP)
		s.data.Sessions[index] = item
		return item, s.saveLocked()
	}

	return SessionRecord{}, apierr.NotFound(fmt.Sprintf("device session `%s` not found", trimmedID))
}

func (s *Store) UpdateDeviceSessionIP(sessionID, lastKnownIP string, now int64) (SessionRecord, error) {
	trimmedID := strings.TrimSpace(sessionID)
	trimmedIP := strings.TrimSpace(lastKnownIP)
	if trimmedID == "" {
		return SessionRecord{}, apierr.InvalidArgument("session id cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for index, item := range s.data.Sessions {
		if item.ID != trimmedID || item.Kind != SessionKindDevice {
			continue
		}
		if trimmedIP != "" {
			item.LastKnownIP = &trimmedIP
		}
		s.data.Sessions[index] = item
		return item, s.saveLocked()
	}

	return SessionRecord{}, apierr.NotFound(fmt.Sprintf("device session `%s` not found", trimmedID))
}

func (s *Store) ListActiveDeviceSessions(now int64) ([]SessionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, now)
	items := make([]SessionRecord, 0, len(s.data.Sessions))
	for _, item := range s.data.Sessions {
		if item.Kind != SessionKindDevice || item.Status != SessionStatusActive || item.ExpiresAt <= now {
			continue
		}
		copy := item
		items = append(items, copy)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt != items[j].CreatedAt {
			return items[i].CreatedAt > items[j].CreatedAt
		}
		return items[i].ID < items[j].ID
	})

	if err := s.saveLocked(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) RevokeDeviceSessions(deviceID string, now int64) ([]SessionRecord, error) {
	trimmedID := strings.TrimSpace(deviceID)
	if trimmedID == "" {
		return nil, apierr.InvalidArgument("device id cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	expireSessionsLocked(&s.data.Sessions, now)
	revoked := make([]SessionRecord, 0, 1)
	for index, item := range s.data.Sessions {
		if item.Kind != SessionKindDevice || item.DeviceID == nil || *item.DeviceID != trimmedID {
			continue
		}
		if item.Status != SessionStatusActive {
			continue
		}
		item.Status = SessionStatusRevoked
		item.RevokedAt = &now
		s.data.Sessions[index] = item
		revoked = append(revoked, item)
	}

	if err := s.saveLocked(); err != nil {
		return nil, err
	}
	return revoked, nil
}

func (session *SessionRecord) UpdatedLastKnownIP(now int64, ip *string) {
	if ip != nil && strings.TrimSpace(*ip) != "" {
		value := strings.TrimSpace(*ip)
		session.LastKnownIP = &value
	}
	if session.ActivatedAt == nil && session.Kind == SessionKindDevice {
		session.ActivatedAt = &now
	}
}
