package auth

type AuthStatus struct {
	TokenTTLMinutes  int    `json:"tokenTtlMinutes"`
	RotationEnabled  bool   `json:"rotationEnabled"`
	BearerHeaderName string `json:"bearerHeaderName"`
}

type SessionSnapshot struct {
	SessionID        string `json:"sessionId"`
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
