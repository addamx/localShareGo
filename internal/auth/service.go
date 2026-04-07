package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"

	"github.com/google/uuid"

	"localShareGo/internal/apierr"
	"localShareGo/internal/config"
	"localShareGo/internal/store"
)

type Service struct {
	store             *store.Store
	tokenTTLMinutes   int
	mu                sync.Mutex
	currentEntryID    string
	currentEntryToken string
	pairRequests      map[string]pairRequest
	pairRequestTTLms  int64
}

func New(store *store.Store, tokenTTLMinutes int) *Service {
	return &Service{
		store:            store,
		tokenTTLMinutes:  tokenTTLMinutes,
		pairRequests:     make(map[string]pairRequest),
		pairRequestTTLms: 5 * 60_000,
	}
}

func (a *Service) Status() AuthStatus {
	return AuthStatus{
		TokenTTLMinutes:  a.tokenTTLMinutes,
		RotationEnabled:  true,
		BearerHeaderName: "Authorization",
	}
}

func (a *Service) EnsureSession(now int64) (store.SessionRecord, string, error) {
	return a.EnsureEntrySession(now)
}

func (a *Service) EnsureEntrySession(now int64) (store.SessionRecord, string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentEntryToken != "" && a.currentEntryID != "" {
		current, err := a.store.GetPendingEntrySession(a.currentEntryID, now)
		if err != nil {
			return store.SessionRecord{}, "", err
		}
		if current != nil {
			return *current, a.currentEntryToken, nil
		}
	}

	return a.issuePendingEntryLocked(now)
}

func (a *Service) RotateSession(now int64) (store.SessionRecord, string, error) {
	return a.RotateEntrySession(now)
}

func (a *Service) RotateEntrySession(now int64) (store.SessionRecord, string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.issuePendingEntryLocked(now)
}

func (a *Service) ActivateEntry(entryToken, deviceID, deviceName, lastKnownIP string, now int64) (store.SessionRecord, string, error) {
	trimmedToken := strings.TrimSpace(entryToken)
	trimmedDeviceID := strings.TrimSpace(deviceID)
	trimmedDeviceName := strings.TrimSpace(deviceName)
	if trimmedToken == "" {
		return store.SessionRecord{}, "", apierr.Unauthorized("missing entry token")
	}
	if trimmedDeviceID == "" {
		return store.SessionRecord{}, "", apierr.InvalidArgument("device id cannot be empty")
	}
	if trimmedDeviceName == "" {
		return store.SessionRecord{}, "", apierr.InvalidArgument("device name cannot be empty")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	entrySession, err := a.store.GetPendingEntrySessionByHash(hashText(trimmedToken), now)
	if err != nil {
		return store.SessionRecord{}, "", err
	}
	if entrySession == nil {
		return store.SessionRecord{}, "", apierr.Unauthorized("invalid or expired entry token")
	}

	if _, err := a.store.UpsertLinkedWebDevice(trimmedDeviceID, trimmedDeviceName, lastKnownIP, now); err != nil {
		return store.SessionRecord{}, "", err
	}

	deviceToken := uuid.NewString()
	_, deviceSession, err := a.store.ConsumeEntrySession(
		entrySession.ID,
		hashText(deviceToken),
		trimmedDeviceID,
		trimmedDeviceName,
		lastKnownIP,
		now+int64(a.tokenTTLMinutes)*60_000,
		now,
	)
	if err != nil {
		return store.SessionRecord{}, "", err
	}

	if a.currentEntryID == entrySession.ID {
		if _, _, err := a.issuePendingEntryLocked(now); err != nil {
			return store.SessionRecord{}, "", err
		}
	}

	return deviceSession, deviceToken, nil
}

func (a *Service) IssueDeviceSession(deviceID, deviceName, lastKnownIP string, now int64) (store.SessionRecord, string, error) {
	trimmedDeviceID := strings.TrimSpace(deviceID)
	trimmedDeviceName := strings.TrimSpace(deviceName)
	if trimmedDeviceID == "" {
		return store.SessionRecord{}, "", apierr.InvalidArgument("device id cannot be empty")
	}
	if trimmedDeviceName == "" {
		return store.SessionRecord{}, "", apierr.InvalidArgument("device name cannot be empty")
	}

	if _, err := a.store.UpsertLinkedWebDevice(trimmedDeviceID, trimmedDeviceName, lastKnownIP, now); err != nil {
		return store.SessionRecord{}, "", err
	}

	token := uuid.NewString()
	session, err := a.store.CreateOrReplaceDeviceSession(
		hashText(token),
		trimmedDeviceID,
		trimmedDeviceName,
		lastKnownIP,
		now+int64(a.tokenTTLMinutes)*60_000,
		now,
	)
	if err != nil {
		return store.SessionRecord{}, "", err
	}

	return session, token, nil
}

func (a *Service) ValidateToken(token string, now int64) (store.SessionRecord, bool, error) {
	session, err := a.ValidateDeviceToken(token, now)
	return session, false, err
}

func (a *Service) ValidateDeviceToken(token string, now int64) (store.SessionRecord, error) {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return store.SessionRecord{}, apierr.Unauthorized("missing bearer token")
	}

	session, err := a.store.GetActiveDeviceSessionByHash(hashText(trimmed), now)
	if err != nil {
		return store.SessionRecord{}, err
	}
	if session == nil {
		return store.SessionRecord{}, apierr.Unauthorized("invalid or expired device credential")
	}
	if session.DeviceID == nil || strings.TrimSpace(*session.DeviceID) == "" {
		return store.SessionRecord{}, apierr.Unauthorized("device session is missing device binding")
	}

	return *session, nil
}

func (a *Service) RenewDeviceSession(token string, now int64) (store.SessionRecord, error) {
	session, err := a.ValidateDeviceToken(token, now)
	if err != nil {
		return store.SessionRecord{}, err
	}

	remaining := session.ExpiresAt - now
	maxRemaining := int64(a.tokenTTLMinutes) * 60_000 / 2
	if remaining > maxRemaining {
		return store.SessionRecord{}, apierr.InvalidArgument("device association cannot be renewed yet")
	}

	return a.store.RenewDeviceSession(session.ID, now+int64(a.tokenTTLMinutes)*60_000, now)
}

func (a *Service) UpdateDeviceSessionIP(sessionID, lastKnownIP string, now int64) (store.SessionRecord, error) {
	return a.store.UpdateDeviceSessionIP(sessionID, lastKnownIP, now)
}

func (a *Service) RevokeDevice(deviceID string, now int64) ([]store.SessionRecord, error) {
	activeSessions, err := a.store.ListActiveDeviceSessions(now)
	if err != nil {
		return nil, err
	}

	if err := a.store.RevokeLinkedWebDevice(deviceID, now); err != nil {
		return nil, err
	}

	revoked := make([]store.SessionRecord, 0, 1)
	for _, session := range activeSessions {
		if session.DeviceID == nil || strings.TrimSpace(*session.DeviceID) != strings.TrimSpace(deviceID) {
			continue
		}
		revoked = append(revoked, session)
	}
	return revoked, nil
}

func (a *Service) ListLinkedDevices(now int64) ([]LinkedDeviceSummary, error) {
	devices, err := a.store.ListLinkedWebDevices()
	if err != nil {
		return nil, err
	}

	items := make([]LinkedDeviceSummary, 0, len(devices))
	for _, device := range devices {
		expiresAt := int64(0)
		session, sessionErr := a.store.GetActiveDeviceSessionByDeviceID(device.ID, now)
		if sessionErr != nil {
			return nil, sessionErr
		}
		if session != nil {
			expiresAt = session.ExpiresAt
		}
		items = append(items, LinkedDeviceSummary{
			ID:          device.ID,
			Name:        device.Name,
			LastKnownIP: optionalString(device.LastKnownIP),
			LastSeenAt:  deviceLastSeenAt(device.LastSeenAt),
			ExpiresAt:   expiresAt,
		})
	}
	return items, nil
}

func (a *Service) CurrentToken() string {
	return a.CurrentEntryToken()
}

func (a *Service) CurrentEntryToken() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentEntryToken
}

func (a *Service) CurrentSessionSnapshot(session store.SessionRecord, publicHost string, publicPort int, webBasePath string) SessionSnapshot {
	return a.EntrySessionSnapshot(session, publicHost, publicPort, webBasePath)
}

func (a *Service) EntrySessionSnapshot(session store.SessionRecord, publicHost string, publicPort int, webBasePath string) SessionSnapshot {
	return SessionSnapshot{
		SessionID:        session.ID,
		SessionKind:      session.Kind,
		DeviceID:         optionalString(session.DeviceID),
		DeviceName:       optionalString(session.DeviceName),
		ExpiresAt:        session.ExpiresAt,
		Status:           session.Status,
		AccessURL:        BuildAccessURL(publicHost, publicPort, webBasePath, a.CurrentEntryToken()),
		PublicHost:       publicHost,
		PublicPort:       publicPort,
		WebBasePath:      NormalizeWebBasePath(webBasePath),
		TokenTTLMinutes:  a.tokenTTLMinutes,
		BearerHeaderName: "Authorization",
		TokenQueryKey:    "token",
	}
}

func (a *Service) DeviceSessionSnapshot(session store.SessionRecord, publicHost string, publicPort int, webBasePath, token string) SessionSnapshot {
	return SessionSnapshot{
		SessionID:        session.ID,
		SessionKind:      session.Kind,
		DeviceID:         optionalString(session.DeviceID),
		DeviceName:       optionalString(session.DeviceName),
		ExpiresAt:        session.ExpiresAt,
		Status:           session.Status,
		AccessURL:        BuildAccessURL(publicHost, publicPort, webBasePath, token),
		PublicHost:       publicHost,
		PublicPort:       publicPort,
		WebBasePath:      NormalizeWebBasePath(webBasePath),
		TokenTTLMinutes:  a.tokenTTLMinutes,
		BearerHeaderName: "Authorization",
		TokenQueryKey:    "token",
	}
}

func (a *Service) issuePendingEntryLocked(now int64) (store.SessionRecord, string, error) {
	token := uuid.NewString()
	session, err := a.store.ReplacePendingEntrySession(hashText(token), now)
	if err != nil {
		return store.SessionRecord{}, "", err
	}

	a.currentEntryID = session.ID
	a.currentEntryToken = token
	return session, token, nil
}

func BuildAccessURL(publicHost string, publicPort int, webBasePath, token string) string {
	path := NormalizeWebBasePath(webBasePath)
	return fmt.Sprintf("http://%s:%d%s?token=%s", strings.TrimSpace(publicHost), publicPort, path, token)
}

func NormalizeWebBasePath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return config.DefaultWebRoute
	}
	trimmed = strings.TrimRight(trimmed, "/")
	if !strings.HasPrefix(trimmed, "/") {
		return "/" + trimmed
	}
	return trimmed
}

func hashText(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}

func deviceLastSeenAt(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}
