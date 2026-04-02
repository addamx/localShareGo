package fileclip

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	png "image/png"
)

const thumbnailMaxEdge = 256

func InspectPath(path string) (Metadata, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Metadata{}, err
	}
	if info.IsDir() {
		return Metadata{}, fmt.Errorf("path is not a file")
	}

	name := filepath.Base(path)
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(name)), ".")
	mimeType := detectMIMEType(path, ext)
	thumbnail, _ := buildThumbnailDataURL(path, mimeType)

	return Metadata{
		FileName:         name,
		Extension:        ext,
		MIMEType:         mimeType,
		SizeBytes:        info.Size(),
		ThumbnailDataURL: thumbnail,
	}, nil
}

func detectMIMEType(path, extension string) string {
	if extension != "" {
		if value := mime.TypeByExtension("." + extension); value != "" {
			return value
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return "application/octet-stream"
	}
	defer file.Close()

	header := make([]byte, 512)
	size, err := file.Read(header)
	if err != nil && err != io.EOF {
		return "application/octet-stream"
	}
	return http.DetectContentType(header[:size])
}

func buildThumbnailDataURL(path, mimeType string) (*string, error) {
	if !strings.HasPrefix(mimeType, "image/") {
		return nil, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, nil
	}

	thumb := resizeImage(img, thumbnailMaxEdge)
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, thumb); err != nil {
		return nil, nil
	}

	encoded := base64.StdEncoding.EncodeToString(buffer.Bytes())
	value := "data:image/png;base64," + encoded
	return &value, nil
}

func resizeImage(src image.Image, maxEdge int) image.Image {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= maxEdge && height <= maxEdge {
		return src
	}

	if width >= height {
		height = (height * maxEdge) / width
		width = maxEdge
	} else {
		width = (width * maxEdge) / height
		height = maxEdge
	}
	if width <= 0 {
		width = 1
	}
	if height <= 0 {
		height = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		srcY := bounds.Min.Y + (y*bounds.Dy())/height
		for x := 0; x < width; x++ {
			srcX := bounds.Min.X + (x*bounds.Dx())/width
			dst.Set(x, y, src.At(srcX, srcY))
		}
	}
	return dst
}
