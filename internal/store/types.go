package store

type PersistenceStatus struct {
	DatabasePath      string `json:"databasePath"`
	MigrationsEnabled bool   `json:"migrationsEnabled"`
	SchemaVersion     int    `json:"schemaVersion"`
	Ready             bool   `json:"ready"`
}

type ClipboardListQuery struct {
	Search         *string `json:"search"`
	PinnedOnly     bool    `json:"pinnedOnly"`
	IncludeDeleted bool    `json:"includeDeleted"`
	CreatedBefore  *int64  `json:"createdBefore"`
	BeforeID       *string `json:"beforeId"`
	Limit          int     `json:"limit"`
}

const (
	ClipboardItemKindText = "text"
	ClipboardItemKindFile = "file"

	TransferStateMetadataOnly = "metadata_only"
	TransferStateReceiving    = "receiving"
	TransferStateReceived     = "received"
	TransferStateFailed       = "failed"
)

type ClipboardFileMeta struct {
	FileName         string  `json:"fileName"`
	Extension        string  `json:"extension"`
	MIMEType         string  `json:"mimeType"`
	SizeBytes        int64   `json:"sizeBytes"`
	ThumbnailDataURL *string `json:"thumbnailDataUrl"`
	TransferState    string  `json:"transferState"`
	ProgressPercent  int     `json:"progressPercent"`
	LocalPath        *string `json:"localPath"`
	DownloadedAt     *int64  `json:"downloadedAt"`
}

type ClipboardItemRecord struct {
	ItemKind       string             `json:"itemKind"`
	ID             string             `json:"id"`
	Content        string             `json:"content"`
	ContentType    string             `json:"contentType"`
	Hash           string             `json:"hash"`
	Preview        string             `json:"preview"`
	CharCount      int                `json:"charCount"`
	FileMeta       *ClipboardFileMeta `json:"fileMeta"`
	SourceKind     string             `json:"sourceKind"`
	SourceDeviceID *string            `json:"sourceDeviceId"`
	Pinned         bool               `json:"pinned"`
	IsCurrent      bool               `json:"isCurrent"`
	DeletedAt      *int64             `json:"deletedAt"`
	CreatedAt      int64              `json:"createdAt"`
	UpdatedAt      int64              `json:"updatedAt"`
}

type ClipboardItemSummary struct {
	ItemKind       string             `json:"itemKind"`
	ID             string             `json:"id"`
	Preview        string             `json:"preview"`
	CharCount      int                `json:"charCount"`
	ContentType    string             `json:"contentType"`
	FileMeta       *ClipboardFileMeta `json:"fileMeta"`
	SourceKind     string             `json:"sourceKind"`
	SourceDeviceID *string            `json:"sourceDeviceId"`
	Pinned         bool               `json:"pinned"`
	IsCurrent      bool               `json:"isCurrent"`
	DeletedAt      *int64             `json:"deletedAt"`
	CreatedAt      int64              `json:"createdAt"`
	UpdatedAt      int64              `json:"updatedAt"`
}

type SaveClipboardInput struct {
	ItemKind       string
	Content        string
	ContentType    string
	Hash           string
	Preview        string
	CharCount      int
	FileMeta       *ClipboardFileMeta
	SourceKind     string
	SourceDeviceID *string
	Pinned         bool
	MarkCurrent    bool
}

type SaveClipboardResult struct {
	Item           ClipboardItemRecord
	Created        bool
	ReusedExisting bool
}

type UpdateClipboardTransferInput struct {
	TransferState   string
	ProgressPercent int
	LocalPath       *string
	DownloadedAt    *int64
}

type SessionRecord struct {
	ID          string `json:"id"`
	TokenHash   string `json:"tokenHash"`
	ExpiresAt   int64  `json:"expiresAt"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"createdAt"`
	ActivatedAt *int64 `json:"activatedAt"`
	RotatedAt   *int64 `json:"rotatedAt"`
}

type DeviceRecord struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

type persistentState struct {
	Devices        []DeviceRecord        `json:"devices"`
	Sessions       []SessionRecord       `json:"sessions"`
	ClipboardItems []ClipboardItemRecord `json:"clipboardItems"`
}
