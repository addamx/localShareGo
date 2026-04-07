package auth

import (
	"strings"

	"github.com/google/uuid"

	"localShareGo/internal/apierr"
)

const (
	pairRequestStatusPending  = "pending"
	pairRequestStatusApproved = "approved"
	pairRequestStatusRejected = "rejected"
	pairRequestStatusExpired  = "expired"
)

type pairRequest struct {
	ID         string
	DeviceID   string
	DeviceName string
	Status     string
	AccessURL  string
	Credential string
	CreatedAt  int64
	UpdatedAt  int64
	ExpiresAt  int64
}

func (a *Service) CreatePairRequest(deviceID, deviceName string, now int64) (PairRequestSummary, error) {
	trimmedDeviceID := strings.TrimSpace(deviceID)
	trimmedDeviceName := strings.TrimSpace(deviceName)
	if trimmedDeviceID == "" {
		return PairRequestSummary{}, apierr.InvalidArgument("device id cannot be empty")
	}
	if trimmedDeviceName == "" {
		return PairRequestSummary{}, apierr.InvalidArgument("device name cannot be empty")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	a.prunePairRequestsLocked(now)
	for id, request := range a.pairRequests {
		if request.DeviceID != trimmedDeviceID || request.Status != pairRequestStatusPending {
			continue
		}
		request.DeviceName = trimmedDeviceName
		request.UpdatedAt = now
		a.pairRequests[id] = request
		return request.summary(), nil
	}

	request := pairRequest{
		ID:         createPairRequestID(),
		DeviceID:   trimmedDeviceID,
		DeviceName: trimmedDeviceName,
		Status:     pairRequestStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
		ExpiresAt:  now + a.pairRequestTTLms,
	}
	a.pairRequests[request.ID] = request
	return request.summary(), nil
}

func (a *Service) ListPairRequests(now int64) []PairRequestSummary {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.prunePairRequestsLocked(now)
	items := make([]PairRequestSummary, 0, len(a.pairRequests))
	for _, request := range a.pairRequests {
		if request.Status != pairRequestStatusPending {
			continue
		}
		items = append(items, request.summary())
	}
	return items
}

func (a *Service) GetPairRequestStatus(requestID string, now int64) (PairRequestStatus, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.prunePairRequestsLocked(now)
	request, ok := a.pairRequests[strings.TrimSpace(requestID)]
	if !ok {
		return PairRequestStatus{}, apierr.NotFound("pair request not found")
	}
	return request.status(), nil
}

func (a *Service) ApprovePairRequest(requestID, publicHost string, publicPort int, webBasePath string, now int64) (PairRequestSummary, error) {
	a.mu.Lock()
	request, ok := a.pairRequests[strings.TrimSpace(requestID)]
	if !ok {
		a.mu.Unlock()
		return PairRequestSummary{}, apierr.NotFound("pair request not found")
	}
	if request.Status != pairRequestStatusPending || request.ExpiresAt <= now {
		a.mu.Unlock()
		return PairRequestSummary{}, apierr.InvalidArgument("pair request is not pending")
	}
	a.mu.Unlock()

	session, token, err := a.IssueDeviceSession(request.DeviceID, request.DeviceName, "", now)
	if err != nil {
		return PairRequestSummary{}, err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	request, ok = a.pairRequests[strings.TrimSpace(requestID)]
	if !ok {
		return PairRequestSummary{}, apierr.NotFound("pair request not found")
	}
	request.Status = pairRequestStatusApproved
	request.UpdatedAt = now
	request.Credential = token
	request.AccessURL = BuildAccessURL(publicHost, publicPort, webBasePath, token)
	if session.DeviceName != nil {
		request.DeviceName = strings.TrimSpace(*session.DeviceName)
	}
	a.pairRequests[request.ID] = request
	return request.summary(), nil
}

func (a *Service) RejectPairRequest(requestID string, now int64) (PairRequestSummary, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	request, ok := a.pairRequests[strings.TrimSpace(requestID)]
	if !ok {
		return PairRequestSummary{}, apierr.NotFound("pair request not found")
	}
	if request.Status != pairRequestStatusPending {
		return PairRequestSummary{}, apierr.InvalidArgument("pair request is not pending")
	}
	request.Status = pairRequestStatusRejected
	request.UpdatedAt = now
	a.pairRequests[request.ID] = request
	return request.summary(), nil
}

func (a *Service) prunePairRequestsLocked(now int64) {
	for id, request := range a.pairRequests {
		if request.Status == pairRequestStatusPending && request.ExpiresAt <= now {
			request.Status = pairRequestStatusExpired
			request.UpdatedAt = now
			a.pairRequests[id] = request
		}
	}
}

func (request pairRequest) summary() PairRequestSummary {
	return PairRequestSummary{
		ID:         request.ID,
		DeviceID:   request.DeviceID,
		DeviceName: request.DeviceName,
		Status:     request.Status,
		CreatedAt:  request.CreatedAt,
		UpdatedAt:  request.UpdatedAt,
		ExpiresAt:  request.ExpiresAt,
	}
}

func (request pairRequest) status() PairRequestStatus {
	return PairRequestStatus{
		ID:         request.ID,
		DeviceID:   request.DeviceID,
		DeviceName: request.DeviceName,
		Status:     request.Status,
		AccessURL:  request.AccessURL,
		Credential: request.Credential,
		CreatedAt:  request.CreatedAt,
		UpdatedAt:  request.UpdatedAt,
		ExpiresAt:  request.ExpiresAt,
	}
}

func createPairRequestID() string {
	return uuid.NewString()
}
