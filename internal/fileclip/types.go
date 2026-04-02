package fileclip

type Metadata struct {
	FileName         string  `json:"fileName"`
	Extension        string  `json:"extension"`
	MIMEType         string  `json:"mimeType"`
	SizeBytes        int64   `json:"sizeBytes"`
	ThumbnailDataURL *string `json:"thumbnailDataUrl"`
}

type ClipboardFile struct {
	Path     string
	Metadata Metadata
}
