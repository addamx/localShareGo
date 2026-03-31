package presence

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	KindDesktop = "desktop"
	KindWeb     = "web"
)

type Device struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Kind         string `json:"kind"`
	RegisteredAt int64  `json:"registeredAt"`
	LastSeenAt   int64  `json:"lastSeenAt"`
}

type Registry struct {
	ttlMs   int64
	mu      sync.RWMutex
	devices map[string]Device
}

func New(ttl time.Duration) *Registry {
	return &Registry{
		ttlMs:   ttl.Milliseconds(),
		devices: make(map[string]Device),
	}
}

func (r *Registry) Register(name, kind string) Device {
	return r.RegisterWithID("", name, kind)
}

func (r *Registry) RegisterWithID(deviceID, name, kind string) Device {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pruneLocked(nowMs())

	baseName := strings.TrimSpace(name)
	if baseName == "" {
		baseName = "Unknown Device"
	}

	now := nowMs()
	trimmedID := strings.TrimSpace(deviceID)
	if trimmedID != "" {
		if device, ok := r.devices[trimmedID]; ok {
			device.Name = r.resolveUniqueNameLocked(baseName, trimmedID)
			device.Kind = strings.TrimSpace(kind)
			device.LastSeenAt = now
			r.devices[trimmedID] = device
			return device
		}
	}

	if trimmedID == "" {
		trimmedID = uuid.NewString()
	}

	device := Device{
		ID:           trimmedID,
		Name:         r.resolveUniqueNameLocked(baseName, trimmedID),
		Kind:         strings.TrimSpace(kind),
		RegisteredAt: now,
		LastSeenAt:   now,
	}
	r.devices[device.ID] = device
	return device
}

func (r *Registry) Touch(deviceID string) (Device, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pruneLocked(nowMs())

	device, ok := r.devices[deviceID]
	if !ok {
		return Device{}, false
	}

	device.LastSeenAt = nowMs()
	r.devices[deviceID] = device
	return device, true
}

func (r *Registry) Get(deviceID string) (Device, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pruneLocked(nowMs())

	device, ok := r.devices[deviceID]
	return device, ok
}

func (r *Registry) List(excludeIDs ...string) []Device {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pruneLocked(nowMs())

	excluded := make(map[string]struct{}, len(excludeIDs))
	for _, item := range excludeIDs {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		excluded[trimmed] = struct{}{}
	}

	items := make([]Device, 0, len(r.devices))
	for _, device := range r.devices {
		if _, ok := excluded[device.ID]; ok {
			continue
		}
		items = append(items, device)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Kind != items[j].Kind {
			return kindOrder(items[i].Kind) < kindOrder(items[j].Kind)
		}
		if items[i].Name != items[j].Name {
			return items[i].Name < items[j].Name
		}
		return items[i].ID < items[j].ID
	})

	return items
}

func (r *Registry) resolveUniqueNameLocked(baseName, exceptID string) string {
	uniqueName := baseName
	for suffix := 2; r.nameExistsLocked(uniqueName, exceptID); suffix++ {
		uniqueName = fmt.Sprintf("%s (%d)", baseName, suffix)
	}
	return uniqueName
}

func (r *Registry) nameExistsLocked(name, exceptID string) bool {
	for _, device := range r.devices {
		if device.ID == exceptID {
			continue
		}
		if device.Name == name {
			return true
		}
	}
	return false
}

func (r *Registry) pruneLocked(now int64) {
	if r.ttlMs <= 0 {
		return
	}

	for id, device := range r.devices {
		if device.Kind == KindDesktop {
			continue
		}
		if now-device.LastSeenAt > r.ttlMs {
			delete(r.devices, id)
		}
	}
}

func nowMs() int64 {
	return time.Now().UnixMilli()
}

func kindOrder(kind string) int {
	switch kind {
	case KindDesktop:
		return 0
	case KindWeb:
		return 1
	default:
		return 2
	}
}
