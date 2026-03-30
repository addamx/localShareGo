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
	store            *store.Store
	tokenTTLMinutes  int
	mu               sync.Mutex
	currentSessionID string
	currentTokenRaw  string
	currentExpiresAt int64
}

func New(store *store.Store, tokenTTLMinutes int) *Service {
	return &Service{
		store:           store,
		tokenTTLMinutes: tokenTTLMinutes,
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
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentTokenRaw != "" && a.currentExpiresAt > now {
		current, err := a.store.GetSessionByHash(hashText(a.currentTokenRaw), now)
		if err != nil {
			return store.SessionRecord{}, "", err
		}
		if current != nil && current.ID == a.currentSessionID {
			return *current, a.currentTokenRaw, nil
		}
	}

	active, err := a.store.GetCurrentSession(now)
	if err != nil {
		return store.SessionRecord{}, "", err
	}

	token := uuid.NewString()
	tokenHash := hashText(token)
	expiresAt := now + int64(a.tokenTTLMinutes)*60_000

	var session store.SessionRecord
	if active == nil {
		session, err = a.store.CreateSession(tokenHash, expiresAt)
	} else {
		session, err = a.store.RotateSession(active.ID, tokenHash, expiresAt, now)
	}
	if err != nil {
		return store.SessionRecord{}, "", err
	}

	a.currentSessionID = session.ID
	a.currentTokenRaw = token
	a.currentExpiresAt = session.ExpiresAt
	return session, token, nil
}

func (a *Service) RotateSession(now int64) (store.SessionRecord, string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	current, err := a.store.GetCurrentSession(now)
	if err != nil {
		return store.SessionRecord{}, "", err
	}
	if current == nil {
		return store.SessionRecord{}, "", apierr.NotFound("no active session available to rotate")
	}

	token := uuid.NewString()
	tokenHash := hashText(token)
	expiresAt := now + int64(a.tokenTTLMinutes)*60_000
	session, err := a.store.RotateSession(current.ID, tokenHash, expiresAt, now)
	if err != nil {
		return store.SessionRecord{}, "", err
	}

	a.currentSessionID = session.ID
	a.currentTokenRaw = token
	a.currentExpiresAt = session.ExpiresAt
	return session, token, nil
}

func (a *Service) ValidateToken(token string, now int64) (store.SessionRecord, error) {
	trimmed := strings.TrimSpace(token)
	if trimmed == "" {
		return store.SessionRecord{}, apierr.Unauthorized("missing bearer token")
	}
	session, err := a.store.GetSessionByHash(hashText(trimmed), now)
	if err != nil {
		return store.SessionRecord{}, err
	}
	if session == nil {
		return store.SessionRecord{}, apierr.Unauthorized("invalid or expired token")
	}
	return *session, nil
}

func (a *Service) CurrentToken() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.currentTokenRaw
}

func (a *Service) CurrentSessionSnapshot(session store.SessionRecord, publicHost string, publicPort int, webBasePath string) SessionSnapshot {
	return SessionSnapshot{
		SessionID:        session.ID,
		ExpiresAt:        session.ExpiresAt,
		Status:           session.Status,
		AccessURL:        BuildAccessURL(publicHost, publicPort, webBasePath, a.CurrentToken()),
		PublicHost:       publicHost,
		PublicPort:       publicPort,
		WebBasePath:      NormalizeWebBasePath(webBasePath),
		TokenTTLMinutes:  a.tokenTTLMinutes,
		BearerHeaderName: "Authorization",
		TokenQueryKey:    "token",
	}
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
