package instafix

import (
	"fmt"
	"image"
	"image/color"
	"strings"

	"github.com/disintegration/imaging"
)

func setHexColor(dc interface {
	SetRGBA(float64, float64, float64, float64)
}, hex string, opacity float64) {
	c, err := parseHexColor(hex)
	if err != nil {
		dc.SetRGBA(1, 1, 1, opacity)
		return
	}
	dc.SetRGBA(float64(c.R)/255.0, float64(c.G)/255.0, float64(c.B)/255.0, opacity)
}

func parseHexColor(hex string) (color.NRGBA, error) {
	hex = strings.TrimSpace(hex)
	if hex == "" {
		return color.NRGBA{}, fmt.Errorf("empty color")
	}
	if hex[0] != '#' {
		return color.NRGBA{}, fmt.Errorf("invalid color: %s", hex)
	}
	switch len(hex) {
	case 4:
		var r, g, b uint8
		_, err := fmt.Sscanf(hex, "#%1x%1x%1x", &r, &g, &b)
		if err != nil {
			return color.NRGBA{}, err
		}
		return color.NRGBA{R: r * 17, G: g * 17, B: b * 17, A: 255}, nil
	case 7:
		var r, g, b uint8
		_, err := fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
		if err != nil {
			return color.NRGBA{}, err
		}
		return color.NRGBA{R: r, G: g, B: b, A: 255}, nil
	default:
		return color.NRGBA{}, fmt.Errorf("invalid color length: %s", hex)
	}
}

func averageColor(src image.Image) color.NRGBA {
	thumb := imaging.Resize(src, 32, 32, imaging.Lanczos)
	b := thumb.Bounds()
	if b.Dx() == 0 || b.Dy() == 0 {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	}

	var r, g, bl uint64
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			cr, cg, cb, _ := thumb.At(x, y).RGBA()
			r += uint64(cr)
			g += uint64(cg)
			bl += uint64(cb)
		}
	}

	total := uint64(b.Dx() * b.Dy())
	return color.NRGBA{
		R: uint8((r / total) >> 8),
		G: uint8((g / total) >> 8),
		B: uint8((bl / total) >> 8),
		A: 255,
	}
}
