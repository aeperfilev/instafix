package config

import (
	"errors"
	"testing"
)

func TestValidateRejectsAutoWithSize(t *testing.T) {
	cfg := Config{
		Settings: Settings{
			JpegQuality: 90,
			AssetsPath:  "assets",
		},
		Backgrounds: map[string]Background{
			"black": {Type: "solid", Color: "#000000"},
		},
		Watermarks: map[string]Watermark{
			"standard": {
				Font:    "roboto.ttf",
				Size:    12,
				Color:   "#ffffff",
				Opacity: 1,
			},
		},
		Formats: map[string]Format{
			"auto": {Type: "auto", Width: 100, FromList: []string{"square"}},
			"square": {
				Type:   "fixed",
				Width:  100,
				Height: 100,
			},
		},
		Profiles: map[string]Profile{
			"default": {
				BackgroundRef: "black",
				WatermarkRef:  "standard",
				FormatRef:     "auto",
			},
		},
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("expected validation error for auto format with width/height")
	}
}

func TestResolveProfileMissing(t *testing.T) {
	cfg := Config{
		Settings: Settings{
			JpegQuality: 90,
			AssetsPath:  "assets",
		},
	}
	_, err := cfg.ResolveProfile("missing")
	if err == nil {
		t.Fatal("expected error for missing profile")
	}
	if !errors.Is(err, ErrProfileNotFound) {
		t.Fatalf("expected ErrProfileNotFound, got %v", err)
	}
}
