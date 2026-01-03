package instafix

import (
	"image"
	"image/color"
	"testing"

	"github.com/aeperfilev/instafix/config"
)

func TestProcess_NoUpscaleKeepsSmallImage(t *testing.T) {
	cfg := config.Config{
		Settings: config.Settings{
			JpegQuality: 90,
			AssetsPath:  "assets",
		},
		Backgrounds: map[string]config.Background{
			"black": {Type: "solid", Color: "#000000"},
		},
		Formats: map[string]config.Format{
			"square": {Type: "fixed", Width: 500, Height: 500},
		},
		Profiles: map[string]config.Profile{
			"default": {
				BackgroundRef:  "black",
				FormatRef:      "square",
				PaddingPercent: 0,
				NoUpscale:      true,
			},
		},
	}

	processor, err := NewProcessor(cfg)
	if err != nil {
		t.Fatalf("NewProcessor: %v", err)
	}

	src := solidImage(100, 100, color.NRGBA{R: 255, A: 255})
	out, _, err := processor.Process(src, "default", "")
	if err != nil {
		t.Fatalf("Process: %v", err)
	}

	if out.Bounds().Dx() != 500 || out.Bounds().Dy() != 500 {
		t.Fatalf("unexpected output size: %dx%d", out.Bounds().Dx(), out.Bounds().Dy())
	}

	// Corners should remain background when no upscaling is applied.
	corner := out.At(0, 0)
	if !sameColor(corner, color.NRGBA{A: 255}) {
		t.Fatalf("expected black background at corner, got %v", corner)
	}
}

func TestProcess_WatermarkRequiresStyle(t *testing.T) {
	cfg := config.Config{
		Settings: config.Settings{
			JpegQuality: 90,
			AssetsPath:  "assets",
		},
		Backgrounds: map[string]config.Background{
			"black": {Type: "solid", Color: "#000000"},
		},
		Formats: map[string]config.Format{
			"square": {Type: "fixed", Width: 100, Height: 100},
		},
		Profiles: map[string]config.Profile{
			"default": {
				BackgroundRef: "black",
				FormatRef:     "square",
			},
		},
	}

	processor, err := NewProcessor(cfg)
	if err != nil {
		t.Fatalf("NewProcessor: %v", err)
	}

	src := solidImage(10, 10, color.NRGBA{R: 255, A: 255})
	_, _, err = processor.Process(src, "default", "text")
	if err == nil {
		t.Fatal("expected error when watermark style is missing")
	}
}

func TestResolveFormatAutoChoosesClosestRatio(t *testing.T) {
	formats := map[string]config.Format{
		"square":   {Type: "fixed", Width: 1080, Height: 1080},
		"portrait": {Type: "fixed", Width: 1080, Height: 1350},
		"land":     {Type: "fixed", Width: 1080, Height: 566},
		"auto":     {Type: "auto", FromList: []string{"square", "portrait", "land"}},
	}

	format, err := resolveFormat(formats, formats["auto"], 1000, 1250)
	if err != nil {
		t.Fatalf("resolveFormat: %v", err)
	}
	if format.Width != 1080 || format.Height != 1350 {
		t.Fatalf("expected portrait format, got %dx%d", format.Width, format.Height)
	}
}

func solidImage(w, h int, c color.NRGBA) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}

func sameColor(got color.Color, want color.NRGBA) bool {
	r, g, b, a := got.RGBA()
	return uint8(r>>8) == want.R && uint8(g>>8) == want.G && uint8(b>>8) == want.B && uint8(a>>8) == want.A
}
