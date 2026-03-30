package httpserver

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"path"
	"strings"
	"sync"
	"time"

	"localShareGo/internal/apierr"
	"localShareGo/internal/auth"
	"localShareGo/internal/clipboard"
	"localShareGo/internal/config"
	"localShareGo/internal/network"
	"localShareGo/internal/store"
)

type EventBroker struct {
	mu     sync.RWMutex
	nextID int
	subs   map[int]chan ServerEvent
}

func newEventBroker() *EventBroker {
	return &EventBroker{
		subs: make(map[int]chan ServerEvent),
	}
}

func (b *EventBroker) subscribe() (int, <-chan ServerEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()

	id := b.nextID
	b.nextID++
	ch := make(chan ServerEvent, 16)
	b.subs[id] = ch
	return id, ch
}

func (b *EventBroker) unsubscribe(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if ch, ok := b.subs[id]; ok {
		delete(b.subs, id)
		close(ch)
	}
}

func (b *EventBroker) publish(event ServerEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subs {
		select {
		case ch <- event:
		default:
		}
	}
}

type HTTPServer struct {
	config             config.RuntimeConfig
	store              *store.Store
	auth               *auth.Service
	clipboard          *clipboard.Service
	network            *network.Service
	assets             fs.FS
	broker             *EventBroker
	onClipboardRefresh func(clipboard.RefreshEvent)
	mu                 sync.RWMutex
	effectivePort      *int
	state              string
	lastError          *string
	server             *http.Server
}

func New(appConfig config.RuntimeConfig, dataStore *store.Store, authService *auth.Service, clipboardService *clipboard.Service, networkService *network.Service, assets embed.FS, onClipboardRefresh func(clipboard.RefreshEvent)) (*HTTPServer, error) {
	assetFS, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		return nil, err
	}

	return &HTTPServer{
		config:             appConfig,
		store:              dataStore,
		auth:               authService,
		clipboard:          clipboardService,
		network:            networkService,
		assets:             assetFS,
		broker:             newEventBroker(),
		onClipboardRefresh: onClipboardRefresh,
		state:              "stopped",
	}, nil
}

func (s *HTTPServer) Status() HttpServerStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return HttpServerStatus{
		BindHost:       s.config.LanHost,
		PreferredPort:  s.config.PreferredPort,
		EffectivePort:  cloneIntPointer(s.effectivePort),
		State:          s.state,
		LastError:      cloneStringPointer(s.lastError),
		HealthEndpoint: "/api/v1/health",
		WebBasePath:    auth.NormalizeWebBasePath(s.config.WebRoute),
		SSEEndpoint:    "/api/v1/events",
	}
}

func (s *HTTPServer) Start() error {
	s.mu.Lock()
	if s.state == "running" || s.state == "starting" {
		s.mu.Unlock()
		return nil
	}
	s.state = "starting"
	s.lastError = nil
	s.mu.Unlock()

	var listener net.Listener
	var lastErr error
	for offset := 0; offset <= 8; offset++ {
		port := s.config.PreferredPort + offset
		candidate, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.config.LanHost, port))
		if err != nil {
			lastErr = err
			continue
		}
		listener = candidate
		s.mu.Lock()
		s.effectivePort = &port
		s.state = "running"
		s.mu.Unlock()
		break
	}

	if listener == nil {
		message := "failed to bind HTTP server"
		if lastErr != nil {
			message = lastErr.Error()
		}
		s.mu.Lock()
		s.state = "failed"
		s.lastError = &message
		s.mu.Unlock()
		return apierr.State(message)
	}

	server := &http.Server{Handler: s.routes()}

	s.mu.Lock()
	s.server = server
	s.mu.Unlock()

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			message := err.Error()
			s.mu.Lock()
			s.state = "failed"
			s.lastError = &message
			s.mu.Unlock()
		}
	}()

	return nil
}

func (s *HTTPServer) Stop(ctx context.Context) {
	s.mu.RLock()
	server := s.server
	s.mu.RUnlock()
	if server != nil {
		_ = server.Shutdown(ctx)
	}
}

func (s *HTTPServer) PublishRefresh(scope string, itemID *string) {
	s.broker.publish(ServerEvent{
		Kind:   "refresh",
		Scope:  scope,
		ItemID: itemID,
		TS:     nowMs(),
	})
}

func (s *HTTPServer) routes() http.Handler {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.FS(s.assets))

	mux.HandleFunc("/", s.handleRoot)
	mux.Handle("/assets/", http.StripPrefix("/", fileServer))
	mux.HandleFunc(auth.NormalizeWebBasePath(s.config.WebRoute), s.handleWebIndex)
	mux.HandleFunc(auth.NormalizeWebBasePath(s.config.WebRoute)+"/", s.handleWebIndex)

	mux.HandleFunc("/api/v1/health", s.handleHealth)
	mux.HandleFunc("/api/v1/session", s.handleSession)
	mux.HandleFunc("/api/v1/session/rotate-token", s.handleRotateToken)
	mux.HandleFunc("/api/v1/clipboard-items", s.handleClipboardCollection)
	mux.HandleFunc("/api/v1/clipboard-items/clear", s.handleClearClipboardHistory)
	mux.HandleFunc("/api/v1/clipboard-items/", s.handleClipboardItem)
	mux.HandleFunc("/api/v1/events", s.handleEvents)
	return mux
}

func (s *HTTPServer) handleRoot(writer http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.NotFound(writer, request)
		return
	}

	status := s.Status()
	_, _ = io.WriteString(writer, fmt.Sprintf(`<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>LocalShareGo</title></head>
<body style="font-family: system-ui, sans-serif; padding: 24px;">
<h1>LocalShareGo</h1>
<p>Service state: <code>%s</code></p>
<p>Health: <code>/api/v1/health</code></p>
<p>Web route: <code>%s</code></p>
</body></html>`, status.State, status.WebBasePath))
}

func (s *HTTPServer) handleWebIndex(writer http.ResponseWriter, request *http.Request) {
	if !strings.HasPrefix(request.URL.Path, auth.NormalizeWebBasePath(s.config.WebRoute)) {
		http.NotFound(writer, request)
		return
	}

	content, err := fs.ReadFile(s.assets, "index.html")
	if err != nil {
		writeAPIError(writer, apierr.WrapInternal("failed to read web assets", err))
		return
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	writer.Header().Set("Cache-Control", "no-cache")
	_, _ = writer.Write(content)
}

func (s *HTTPServer) handleHealth(writer http.ResponseWriter, _ *http.Request) {
	status := s.Status()
	writeAPIResponse(writer, http.StatusOK, HealthResponse{
		Service:        "LocalShareGo",
		Status:         status.State,
		BindHost:       status.BindHost,
		PreferredPort:  status.PreferredPort,
		EffectivePort:  status.EffectivePort,
		DatabaseReady:  true,
		SessionReady:   s.auth.CurrentToken() != "",
		WebBasePath:    status.WebBasePath,
		HealthEndpoint: status.HealthEndpoint,
		SSEEndpoint:    status.SSEEndpoint,
	})
}

func (s *HTTPServer) handleSession(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	token := extractToken(request)
	session, err := s.auth.ValidateToken(token, nowMs())
	if err != nil {
		writeAPIError(writer, err)
		return
	}

	writeAPIResponse(writer, http.StatusOK, s.buildSessionResponse(session))
}

func (s *HTTPServer) handleRotateToken(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if _, err := s.authorizeRequest(request); err != nil {
		writeAPIError(writer, err)
		return
	}

	session, _, err := s.auth.RotateSession(nowMs())
	if err != nil {
		writeAPIError(writer, err)
		return
	}

	s.PublishRefresh("session", nil)
	writeAPIResponse(writer, http.StatusOK, s.buildSessionResponse(session))
}

func (s *HTTPServer) handleClipboardCollection(writer http.ResponseWriter, request *http.Request) {
	if _, err := s.authorizeRequest(request); err != nil {
		writeAPIError(writer, err)
		return
	}

	switch request.Method {
	case http.MethodGet:
		query := store.ClipboardListQuery{Limit: 80}
		if value := strings.TrimSpace(request.URL.Query().Get("search")); value != "" {
			query.Search = &value
		}
		if request.URL.Query().Get("pinnedOnly") == "true" {
			query.PinnedOnly = true
		}
		if value := request.URL.Query().Get("limit"); value != "" {
			var parsed int
			_, _ = fmt.Sscanf(value, "%d", &parsed)
			if parsed > 0 {
				query.Limit = parsed
			}
		}

		items, err := s.store.ListClipboardItems(query)
		if err != nil {
			writeAPIError(writer, err)
			return
		}

		summaries := make([]store.ClipboardItemSummary, 0, len(items))
		for _, item := range items {
			summaries = append(summaries, store.CloneSummary(item))
		}
		writeAPIResponse(writer, http.StatusOK, clipboard.ListResponse{Items: summaries})
	case http.MethodPost:
		var payload clipboard.WriteRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			writeAPIError(writer, apierr.InvalidArgument("invalid request body"))
			return
		}

		result, err := s.store.SaveClipboardItem(store.SaveClipboardInput{
			Content:     payload.Content,
			SourceKind:  "mobile_web",
			Pinned:      payload.Pinned,
			MarkCurrent: false,
		}, s.config.ClipboardPollInterval, s.config.MaxTextBytes)
		if err != nil {
			writeAPIError(writer, err)
			return
		}

		if payload.Activate {
			if err := s.clipboard.WriteText(result.Item.Content); err != nil {
				writeAPIError(writer, err)
				return
			}
			result.Item, err = s.store.ActivateClipboardItem(result.Item.ID)
			if err != nil {
				writeAPIError(writer, err)
				return
			}
		}

		event := clipboard.RefreshEvent{
			ItemID:         result.Item.ID,
			Created:        result.Created,
			ReusedExisting: result.ReusedExisting,
			IsCurrent:      result.Item.IsCurrent,
			SourceKind:     result.Item.SourceKind,
			ObservedAtMs:   nowMs(),
		}
		s.PublishRefresh("clipboard", &result.Item.ID)
		if s.onClipboardRefresh != nil {
			s.onClipboardRefresh(event)
		}

		writeAPIResponse(writer, http.StatusOK, clipboard.WriteResponse{
			Item:           result.Item,
			Created:        result.Created,
			ReusedExisting: result.ReusedExisting,
		})
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *HTTPServer) handleClipboardItem(writer http.ResponseWriter, request *http.Request) {
	if _, err := s.authorizeRequest(request); err != nil {
		writeAPIError(writer, err)
		return
	}

	trimmed := strings.TrimPrefix(request.URL.Path, "/api/v1/clipboard-items/")
	if trimmed == "" {
		http.NotFound(writer, request)
		return
	}

	if strings.HasSuffix(trimmed, "/activate") {
		itemID := strings.TrimSuffix(trimmed, "/activate")
		s.handleActivateClipboardItem(writer, request, itemID)
		return
	}

	itemID := path.Clean("/" + trimmed)
	itemID = strings.TrimPrefix(itemID, "/")
	if itemID == "" || itemID == "." {
		http.NotFound(writer, request)
		return
	}

	switch request.Method {
	case http.MethodGet:
		item, err := s.store.GetClipboardItem(itemID)
		if err != nil {
			writeAPIError(writer, err)
			return
		}
		if item == nil {
			writeAPIError(writer, apierr.NotFound(fmt.Sprintf("clipboard item `%s` not found", itemID)))
			return
		}
		writeAPIResponse(writer, http.StatusOK, *item)
	case http.MethodPatch:
		var payload clipboard.PinRequest
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			writeAPIError(writer, apierr.InvalidArgument("invalid request body"))
			return
		}
		item, err := s.store.UpdateClipboardItemPin(itemID, payload.Pinned)
		if err != nil {
			writeAPIError(writer, err)
			return
		}
		s.PublishRefresh("clipboard", &itemID)
		if s.onClipboardRefresh != nil {
			s.onClipboardRefresh(clipboard.RefreshEvent{
				ItemID:       item.ID,
				IsCurrent:    item.IsCurrent,
				SourceKind:   item.SourceKind,
				ObservedAtMs: nowMs(),
			})
		}
		writeAPIResponse(writer, http.StatusOK, item)
	case http.MethodDelete:
		if err := s.store.SoftDeleteClipboardItem(itemID); err != nil {
			writeAPIError(writer, err)
			return
		}
		s.PublishRefresh("clipboard", &itemID)
		if s.onClipboardRefresh != nil {
			s.onClipboardRefresh(clipboard.RefreshEvent{
				ItemID:       itemID,
				IsCurrent:    false,
				ObservedAtMs: nowMs(),
			})
		}
		writeAPIResponse(writer, http.StatusOK, clipboard.DeleteResponse{ItemID: itemID})
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *HTTPServer) handleActivateClipboardItem(writer http.ResponseWriter, request *http.Request, itemID string) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	item, err := s.store.GetClipboardItem(itemID)
	if err != nil {
		writeAPIError(writer, err)
		return
	}
	if item == nil {
		writeAPIError(writer, apierr.NotFound(fmt.Sprintf("clipboard item `%s` not found", itemID)))
		return
	}
	if err := s.clipboard.WriteText(item.Content); err != nil {
		writeAPIError(writer, err)
		return
	}
	activated, err := s.store.ActivateClipboardItem(itemID)
	if err != nil {
		writeAPIError(writer, err)
		return
	}

	s.PublishRefresh("clipboard", &itemID)
	if s.onClipboardRefresh != nil {
		s.onClipboardRefresh(clipboard.RefreshEvent{
			ItemID:       itemID,
			IsCurrent:    true,
			SourceKind:   activated.SourceKind,
			ObservedAtMs: nowMs(),
		})
	}
	writeAPIResponse(writer, http.StatusOK, activated)
}

func (s *HTTPServer) handleClearClipboardHistory(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if _, err := s.authorizeRequest(request); err != nil {
		writeAPIError(writer, err)
		return
	}
	count, err := s.store.ClearClipboardHistory()
	if err != nil {
		writeAPIError(writer, err)
		return
	}
	s.PublishRefresh("clipboard", nil)
	if s.onClipboardRefresh != nil {
		s.onClipboardRefresh(clipboard.RefreshEvent{
			ItemID:       "",
			ObservedAtMs: nowMs(),
		})
	}
	writeAPIResponse(writer, http.StatusOK, clipboard.ClearResponse{ClearedCount: count})
}

func (s *HTTPServer) handleEvents(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if _, err := s.authorizeRequest(request); err != nil {
		writeAPIError(writer, err)
		return
	}

	flusher, ok := writer.(http.Flusher)
	if !ok {
		writeAPIError(writer, apierr.State("streaming is not supported"))
		return
	}

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	id, ch := s.broker.subscribe()
	defer s.broker.unsubscribe(id)

	_, _ = io.WriteString(writer, "event: ready\ndata: connected\n\n")
	flusher.Flush()

	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-request.Context().Done():
			return
		case <-keepAlive.C:
			_, _ = io.WriteString(writer, ": keepalive\n\n")
			flusher.Flush()
		case event, ok := <-ch:
			if !ok {
				return
			}
			body, _ := json.Marshal(event)
			_, _ = fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", event.Kind, body)
			flusher.Flush()
		}
	}
}

func (s *HTTPServer) authorizeRequest(request *http.Request) (store.SessionRecord, error) {
	return s.auth.ValidateToken(extractToken(request), nowMs())
}

func (s *HTTPServer) buildSessionResponse(session store.SessionRecord) SessionResponse {
	status := s.Status()
	port := status.PreferredPort
	if status.EffectivePort != nil {
		port = *status.EffectivePort
	}
	return SessionResponse{
		DeviceName:       s.network.DeviceName(),
		PublicHost:       s.network.AccessHost(),
		PublicPort:       port,
		AccessURL:        auth.BuildAccessURL(s.network.AccessHost(), port, s.config.WebRoute, s.auth.CurrentToken()),
		HealthEndpoint:   "/api/v1/health",
		SSEEndpoint:      "/api/v1/events",
		WebBasePath:      auth.NormalizeWebBasePath(s.config.WebRoute),
		SessionID:        session.ID,
		SessionStatus:    session.Status,
		ExpiresAt:        session.ExpiresAt,
		TokenTTLMinutes:  s.config.TokenTTLMinutes,
		BearerHeaderName: "Authorization",
		TokenQueryKey:    "token",
		RotationEnabled:  true,
		MaxTextBytes:     s.config.MaxTextBytes,
		ReadOnly:         false,
	}
}

func extractToken(request *http.Request) string {
	header := strings.TrimSpace(request.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return strings.TrimSpace(header[7:])
	}
	return strings.TrimSpace(request.URL.Query().Get("token"))
}

func writeAPIResponse[T any](writer http.ResponseWriter, status int, data T) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	payload := APIEnvelope[T]{
		OK:   true,
		Data: &data,
		TS:   nowMs(),
	}
	_ = json.NewEncoder(writer).Encode(payload)
}

func writeAPIError(writer http.ResponseWriter, err error) {
	apiError := apierr.AsAPIError(err)
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(apiError.HTTPStatus)
	payload := APIEnvelope[map[string]any]{
		OK:    false,
		Error: apiError,
		TS:    nowMs(),
	}
	_ = json.NewEncoder(writer).Encode(payload)
}

func cloneIntPointer(value *int) *int {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	copy := *value
	return &copy
}

func nowMs() int64 {
	return time.Now().UnixMilli()
}
