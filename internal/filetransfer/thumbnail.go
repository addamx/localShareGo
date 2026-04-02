package filetransfer

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"os"
	"strings"
)

func makeThumbnailDataURL(path, mimeType string) *string {
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(mimeType)), "image/") {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer file.Close()

	source, _, err := image.Decode(file)
	if err != nil {
		return nil
	}

	thumb := resizeImage(source, 256)
	var buffer bytes.Buffer
	if err := png.Encode(&buffer, thumb); err != nil {
		return nil
	}

	return encodeDataURL("image/png", buffer.Bytes())
}

func resizeImage(source image.Image, maxSide int) image.Image {
	bounds := source.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= maxSide && height <= maxSide {
		return source
	}

	nextWidth := width
	nextHeight := height
	if width >= height {
		nextWidth = maxSide
		nextHeight = widthToHeight(width, height, maxSide)
	} else {
		nextHeight = maxSide
		nextWidth = widthToHeight(height, width, maxSide)
	}
	if nextWidth < 1 {
		nextWidth = 1
	}
	if nextHeight < 1 {
		nextHeight = 1
	}

	target := image.NewRGBA(image.Rect(0, 0, nextWidth, nextHeight))
	for y := 0; y < nextHeight; y++ {
		sourceY := bounds.Min.Y + (y*height)/nextHeight
		for x := 0; x < nextWidth; x++ {
			sourceX := bounds.Min.X + (x*width)/nextWidth
			target.Set(x, y, source.At(sourceX, sourceY))
		}
	}

	return target
}

func widthToHeight(sourceLength, sourceOther, targetLength int) int {
	if sourceLength <= 0 {
		return targetLength
	}
	return (sourceOther * targetLength) / sourceLength
}
