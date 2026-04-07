package auth

const EventName = "localshare://session/refresh"
const PairRequestEventName = "localshare://pair-request/pending"

type AuthStatus struct {
	TokenTTLMinutes  int    `json:"tokenTtlMinutes"`
	RotationEnabled  bool   `json:"rotationEnabled"`
	BearerHeaderName string `json:"bearerHeaderName"`
}

type SessionSnapshot struct {
	SessionID        string `json:"sessionId"`
	SessionKind      string `json:"sessionKind"`
	DeviceID         string `json:"deviceId"`
	DeviceName       string `json:"deviceName"`
	ExpiresAt        int64  `json:"expiresAt"`
	Status           string `json:"status"`
	AccessURL        string `json:"accessUrl"`
	PublicHost       string `json:"publicHost"`
	PublicPort       int    `json:"publicPort"`
	WebBasePath      string `json:"webBasePath"`
	TokenTTLMinutes  int    `json:"tokenTtlMinutes"`
	BearerHeaderName string `json:"bearerHeaderName"`
	TokenQueryKey    string `json:"tokenQueryKey"`
}

type PairRequestSummary struct {
	ID         string `json:"id"`
	DeviceID   string `json:"deviceId"`
	DeviceName string `json:"deviceName"`
	Status     string `json:"status"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
	ExpiresAt  int64  `json:"expiresAt"`
}

type PairRequestStatus struct {
	ID         string `json:"id"`
	DeviceID   string `json:"deviceId"`
	DeviceName string `json:"deviceName"`
	Status     string `json:"status"`
	AccessURL  string `json:"accessUrl"`
	Credential string `json:"credential"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
	ExpiresAt  int64  `json:"expiresAt"`
}

type LinkedDeviceSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	LastKnownIP string `json:"lastKnownIp"`
	LastSeenAt  int64  `json:"lastSeenAt"`
	Online      bool   `json:"online"`
	ExpiresAt   int64  `json:"expiresAt"`
}
