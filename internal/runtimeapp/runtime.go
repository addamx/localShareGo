package runtimeapp

import (
	"bufio"
	"context"
	"embed"
	"net"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"localShareGo/internal/auth"
	"localShareGo/internal/clipboard"
	"localShareGo/internal/config"
	"localShareGo/internal/httpserver"
	"localShareGo/internal/network"
	"localShareGo/internal/store"
)

type AppRuntime struct {
	config    config.RuntimeConfig
	paths     config.AppPaths
	store     *store.Store
	auth      *auth.Service
	network   *network.Service
	clipboard *clipboard.Service
	http      *httpserver.HTTPServer
	deviceID  string
}

func New(ctx context.Context, assets embed.FS) (*AppRuntime, error) {
	appConfig := config.DefaultRuntimeConfig()
	paths := config.ResolveAppPaths(appConfig)
	if err := config.EnsureAppDirs(paths); err != nil {
		return nil, err
	}

	dataStore, err := store.New(paths.DatabasePath)
	if err != nil {
		return nil, err
	}

	networkService := network.New()
	device, err := dataStore.UpsertDevice(networkService.DeviceName())
	if err != nil {
		return nil, err
	}

	authService := auth.New(dataStore, appConfig.TokenTTLMinutes)
	if _, _, err := authService.EnsureSession(nowMs()); err != nil {
		return nil, err
	}

	deviceID := device.ID
	clipboardService := clipboard.New(dataStore, appConfig.ClipboardPollInterval, appConfig.MaxTextBytes, &deviceID, func(event clipboard.RefreshEvent) {
		wruntime.EventsEmit(ctx, clipboard.EventName, event)
	})

	server, err := httpserver.New(appConfig, dataStore, authService, clipboardService, networkService, assets, func(event clipboard.RefreshEvent) {
		wruntime.EventsEmit(ctx, clipboard.EventName, event)
	})
	if err != nil {
		return nil, err
	}

	return &AppRuntime{
		config:    appConfig,
		paths:     paths,
		store:     dataStore,
		auth:      authService,
		network:   networkService,
		clipboard: clipboardService,
		http:      server,
		deviceID:  deviceID,
	}, nil
}

func (r *AppRuntime) Start() error {
	if err := r.clipboard.Start(); err != nil {
		return err
	}
	return r.http.Start()
}

func (r *AppRuntime) Stop() {
	r.clipboard.StopLoop()
	r.http.Stop(context.Background())
}

func (r *AppRuntime) Bootstrap() (AppBootstrap, error) {
	session, _, err := r.auth.EnsureSession(nowMs())
	if err != nil {
		return AppBootstrap{}, err
	}

	httpStatus := r.http.Status()
	effectivePort := httpStatus.PreferredPort
	if httpStatus.EffectivePort != nil {
		effectivePort = *httpStatus.EffectivePort
	}

	return AppBootstrap{
		AppName: "LocalShareGo",
		Routes: RouteOverview{
			NaiveDesktop: "/desktop",
			Web:          auth.NormalizeWebBasePath(r.config.WebRoute),
		},
		RuntimeConfig: r.config,
		Paths:         r.paths,
		Services: ServiceOverview{
			Clipboard:   r.clipboard.Status(),
			HTTPServer:  httpStatus,
			Auth:        r.auth.Status(),
			Session:     r.auth.CurrentSessionSnapshot(session, r.network.AccessHost(), effectivePort, r.config.WebRoute),
			Persistence: r.store.Status(),
			Network:     r.network.Status(),
		},
	}, nil
}

func (r *AppRuntime) RotateSession() (auth.SessionSnapshot, error) {
	session, _, err := r.auth.RotateSession(nowMs())
	if err != nil {
		return auth.SessionSnapshot{}, err
	}
	httpStatus := r.http.Status()
	effectivePort := httpStatus.PreferredPort
	if httpStatus.EffectivePort != nil {
		effectivePort = *httpStatus.EffectivePort
	}
	r.http.PublishRefresh("session", nil)
	return r.auth.CurrentSessionSnapshot(session, r.network.AccessHost(), effectivePort, r.config.WebRoute), nil
}

func (r *AppRuntime) GetConnectivityReport() (ConnectivityReport, error) {
	status := r.http.Status()
	port := status.PreferredPort
	if status.EffectivePort != nil {
		port = *status.EffectivePort
	}

	session, _, err := r.auth.EnsureSession(nowMs())
	if err != nil {
		return ConnectivityReport{}, err
	}
	accessURL := r.auth.CurrentSessionSnapshot(session, r.network.AccessHost(), port, r.config.WebRoute).AccessURL

	hosts := []string{"127.0.0.1", "localhost", r.network.AccessHost()}
	hosts = append(hosts, r.network.AccessHosts()...)
	hosts = sortUniqueStrings(hosts)

	checks := make([]ConnectivityCheck, 0, len(hosts))
	for _, host := range hosts {
		checks = append(checks, probeHost(host, port))
	}

	return ConnectivityReport{
		BindHost:      status.BindHost,
		PreferredPort: status.PreferredPort,
		EffectivePort: port,
		ServerState:   status.State,
		ServerError:   status.LastError,
		AccessURL:     accessURL,
		Checks:        checks,
	}, nil
}

func (r *AppRuntime) Store() *store.Store {
	return r.store
}

func (r *AppRuntime) Clipboard() *clipboard.Service {
	return r.clipboard
}

func (r *AppRuntime) HTTP() *httpserver.HTTPServer {
	return r.http
}

func probeHost(host string, port int) ConnectivityCheck {
	url := "http://" + host + ":" + intToString(port) + "/api/v1/health"
	timeout := 900 * time.Millisecond
	result := ConnectivityCheck{
		Host: host,
		URL:  url,
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, intToString(port)), timeout)
	if err != nil {
		message := "tcp connect failed: " + err.Error()
		result.Error = &message
		return result
	}
	result.TCPOk = true
	_ = conn.SetDeadline(time.Now().Add(timeout))

	request, _ := http.NewRequest(http.MethodGet, url, nil)
	if err := request.Write(conn); err != nil {
		message := "http write failed: " + err.Error()
		result.Error = &message
		_ = conn.Close()
		return result
	}

	response, err := http.ReadResponse(bufio.NewReader(conn), request)
	if err != nil {
		message := "http read failed: " + err.Error()
		result.Error = &message
		_ = conn.Close()
		return result
	}
	defer response.Body.Close()
	_ = conn.Close()

	statusLine := response.Proto + " " + response.Status
	result.HTTPStatusLine = &statusLine
	result.HTTPOk = response.StatusCode == http.StatusOK
	if !result.HTTPOk {
		message := "health endpoint did not return 200"
		result.Error = &message
	}
	return result
}

func intToString(value int) string {
	return strconv.Itoa(value)
}

func nowMs() int64 {
	return time.Now().UnixMilli()
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
