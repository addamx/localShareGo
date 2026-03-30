package network

type NetworkStatus struct {
	DeviceName          string   `json:"deviceName"`
	AccessHost          string   `json:"accessHost"`
	AccessHosts         []string `json:"accessHosts"`
	LanDiscoveryEnabled bool     `json:"lanDiscoveryEnabled"`
}
