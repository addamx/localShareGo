package clipboard

import "localShareGo/internal/store"

const EventName = "localshare://clipboard/refresh"

type ClipboardStatus struct {
	Mode                string `json:"mode"`
	PollIntervalMs      int    `json:"pollIntervalMs"`
	DedupWindowMs       int    `json:"dedupWindowMs"`
	MaxTextBytes        int    `json:"maxTextBytes"`
	CurrentItemTracking bool   `json:"currentItemTracking"`
	Running             bool   `json:"running"`
	SubscriberCount     int    `json:"subscriberCount"`
	RefreshEventTopic   string `json:"refreshEventTopic"`
}

type RefreshEvent struct {
	ItemID         string `json:"itemId"`
	Created        bool   `json:"created"`
	ReusedExisting bool   `json:"reusedExisting"`
	IsCurrent      bool   `json:"isCurrent"`
	SourceKind     string `json:"sourceKind"`
	ObservedAtMs   int64  `json:"observedAtMs"`
}

type WriteRequest struct {
	Content  string `json:"content"`
	Pinned   bool   `json:"pinned"`
	Activate bool   `json:"activate"`
}

type WriteResponse struct {
	Item           store.ClipboardItemRecord `json:"item"`
	Created        bool                      `json:"created"`
	ReusedExisting bool                      `json:"reusedExisting"`
}

type ListResponse struct {
	Items []store.ClipboardItemSummary `json:"items"`
}

type PinRequest struct {
	Pinned bool `json:"pinned"`
}

type DeleteResponse struct {
	ItemID string `json:"itemId"`
}

type ClearResponse struct {
	ClearedCount int `json:"clearedCount"`
}
