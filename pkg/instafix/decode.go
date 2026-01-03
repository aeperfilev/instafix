package instafix

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// DecodeImage reads image data and handles common formats plus DNG/RAW previews.
func DecodeImage(r io.Reader, filename string) (image.Image, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".dng", ".raw":
		return decodeDNGPreview(data)
	default:
		img, err := imaging.Decode(bytes.NewReader(data), imaging.AutoOrientation(true))
		if err == nil {
			return img, nil
		}
		// Fallback: try to parse embedded JPEG if body contains one (Tasker/raw uploads).
		if preview, perr := decodeDNGPreview(data); perr == nil {
			return preview, nil
		}
		return nil, err
	}
}

func decodeDNGPreview(data []byte) (image.Image, error) {
	candidates := extractJPEGs(data)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no embedded JPEG preview found in DNG")
	}

	var best image.Image
	bestArea := 0
	for _, candidate := range candidates {
		img, err := imaging.Decode(bytes.NewReader(candidate), imaging.AutoOrientation(true))
		if err != nil {
			continue
		}
		area := img.Bounds().Dx() * img.Bounds().Dy()
		if area > bestArea {
			bestArea = area
			best = img
		}
	}
	if best == nil {
		return nil, fmt.Errorf("failed to decode embedded JPEG preview")
	}
	return best, nil
}

func extractJPEGs(data []byte) [][]byte {
	var results [][]byte
	for i := 0; i < len(data)-1; i++ {
		if data[i] != 0xFF || data[i+1] != 0xD8 {
			continue
		}
		j := bytes.Index(data[i+2:], []byte{0xFF, 0xD9})
		if j == -1 {
			continue
		}
		end := i + 2 + j + 2
		if end > len(data) {
			continue
		}
		results = append(results, data[i:end])
		i = end - 1
	}
	return results
}
