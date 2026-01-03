package instafix

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

func TestDecodeImage_DNGPreview(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 32, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.NRGBA{R: 200, G: 10, B: 20, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("jpeg encode: %v", err)
	}

	payload := append([]byte("DNGFAKE"), buf.Bytes()...)
	payload = append(payload, []byte("TRAILER")...)

	decoded, err := DecodeImage(bytes.NewReader(payload), "test.dng")
	if err != nil {
		t.Fatalf("DecodeImage: %v", err)
	}
	if decoded.Bounds().Dx() != 32 || decoded.Bounds().Dy() != 24 {
		t.Fatalf("unexpected decoded size: %dx%d", decoded.Bounds().Dx(), decoded.Bounds().Dy())
	}
}
