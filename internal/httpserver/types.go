package httpserver

import (
	"localShareGo/internal/apierr"
	"localShareGo/internal/store"
)

type HttpServerStatus struct {
	BindHost       string  `json:"bindHost"`
	PreferredPort  int     `json:"preferredPort"`
	EffectivePort  *int    `json:"effectivePort"`
	State          string  `json:"state"`
	LastError      *string `json:"lastError"`
	HealthEndpoint string  `json:"healthEndpoint"`
	WebBasePath    string  `json:"webBasePath"`
	SSEEndpoint    string  `json:"sseEndpoint"`
}

type HealthResponse struct {
	Service        string `json:"service"`
	Status         string `json:"status"`
	BindHost       string `json:"bindHost"`
	PreferredPort  int    `json:"preferredPort"`
	EffectivePort  *int   `json:"effectivePort"`
	DatabaseReady  bool   `json:"databaseReady"`
	SessionReady   bool   `json:"sessionReady"`
	WebBasePath    string `json:"webBasePath"`
	HealthEndpoint string `json:"healthEndpoint"`
	SSEEndpoint    string `json:"sseEndpoint"`
}

type SessionResponse struct {
	DeviceName       string `json:"deviceName"`
	PublicHost       string `json:"publicHost"`
	PublicPort       int    `json:"publicPort"`
	AccessURL        string `json:"accessUrl"`
	HealthEndpoint   string `json:"healthEndpoint"`
	SSEEndpoint      string `json:"sseEndpoint"`
	WebBasePath      string `json:"webBasePath"`
	SessionID        string `json:"sessionId"`
	SessionStatus    string `json:"sessionStatus"`
	ExpiresAt        int64  `json:"expiresAt"`
	TokenTTLMinutes  int    `json:"tokenTtlMinutes"`
	BearerHeaderName string `json:"bearerHeaderName"`
	TokenQueryKey    string `json:"tokenQueryKey"`
	RotationEnabled  bool   `json:"rotationEnabled"`
	MaxTextBytes     int    `json:"maxTextBytes"`
	ReadOnly         bool   `json:"readOnly"`
}

type OnlineDevice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type DeviceRegisterRequest struct {
	DeviceID string `json:"deviceId"`
	Name     string `json:"name"`
}

type DeviceHeartbeatRequest struct {
	DeviceID string `json:"deviceId"`
}

type DevicePresenceResponse struct {
	Self    OnlineDevice   `json:"self"`
	Devices []OnlineDevice `json:"devices"`
}

type SyncClipboardRequest struct {
	ItemID          *string  `json:"itemId"`
	Content         string   `json:"content"`
	SourceDeviceID  string   `json:"sourceDeviceId"`
	TargetDeviceIDs []string `json:"targetDeviceIds"`
	SyncAll         bool     `json:"syncAll"`
}

type SyncClipboardResponse struct {
	DeliveredDevices []OnlineDevice             `json:"deliveredDevices"`
	DesktopItem      *store.ClipboardItemRecord `json:"desktopItem"`
}

type SyncClipboardEvent struct {
	TargetDeviceIDs []string                  `json:"targetDeviceIds"`
	Item            store.ClipboardItemRecord `json:"item"`
	CreatedAt       int64                     `json:"createdAt"`
}

type FileTransferEvent struct {
	ItemID           string  `json:"itemId"`
	Status           string  `json:"status"`
	ProgressPercent  int     `json:"progressPercent"`
	BytesTransferred int64   `json:"bytesTransferred"`
	BytesTotal       int64   `json:"bytesTotal"`
	ErrorMessage     *string `json:"errorMessage"`
}

type ServerEvent struct {
	Kind         string              `json:"kind"`
	Scope        string              `json:"scope"`
	ItemID       *string             `json:"itemId"`
	Sync         *SyncClipboardEvent `json:"sync"`
	FileTransfer *FileTransferEvent  `json:"fileTransfer"`
	TS           int64               `json:"ts"`
}

type APIEnvelope[T any] struct {
	OK    bool                      `json:"ok"`
	Data  *T                        `json:"data"`
	Error *apierr.WorkbenchAPIError `json:"error"`
	TS    int64                     `json:"ts"`
}
