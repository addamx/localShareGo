package network

import (
	"net"
	"os"
	"sort"
	"strings"
)

type Service struct {
	deviceName  string
	accessHost  string
	accessHosts []string
}

func New() *Service {
	deviceName := strings.TrimSpace(os.Getenv("COMPUTERNAME"))
	if deviceName == "" {
		deviceName, _ = os.Hostname()
	}
	if strings.TrimSpace(deviceName) == "" {
		deviceName = "naivedesktop-host"
	}

	accessHosts := resolveAccessHosts()
	return &Service{
		deviceName:  deviceName,
		accessHost:  selectPrimaryAccessHost(accessHosts),
		accessHosts: accessHosts,
	}
}

func (n *Service) Status() NetworkStatus {
	return NetworkStatus{
		DeviceName:          n.deviceName,
		AccessHost:          n.accessHost,
		AccessHosts:         append([]string(nil), n.accessHosts...),
		LanDiscoveryEnabled: true,
	}
}

func (n *Service) DeviceName() string {
	return n.deviceName
}

func (n *Service) AccessHost() string {
	return n.accessHost
}

func (n *Service) AccessHosts() []string {
	return append([]string(nil), n.accessHosts...)
}

func resolveAccessHosts() []string {
	results := make([]string, 0, 4)
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			if iface.Flags&net.FlagUp == 0 {
				continue
			}
			addresses, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, address := range addresses {
				var ip net.IP
				switch value := address.(type) {
				case *net.IPNet:
					ip = value.IP
				case *net.IPAddr:
					ip = value.IP
				default:
					continue
				}
				ip = ip.To4()
				if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
					continue
				}
				results = append(results, ip.String())
			}
		}
	}

	results = sortUniqueStrings(results)
	if len(results) == 0 {
		results = []string{"127.0.0.1"}
	}
	return results
}

func selectPrimaryAccessHost(candidates []string) string {
	conn, err := net.Dial("udp4", "8.8.8.8:80")
	if err == nil {
		if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			_ = conn.Close()
			host := addr.IP.To4()
			if host != nil {
				hostText := host.String()
				for _, item := range candidates {
					if item == hostText {
						return hostText
					}
				}
			}
		}
	}

	sorted := append([]string(nil), candidates...)
	sort.Strings(sorted)
	if len(sorted) == 0 {
		return "127.0.0.1"
	}
	return sorted[0]
}

func sortUniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	sort.Strings(result)
	return result
}
