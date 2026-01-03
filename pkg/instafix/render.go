package instafix

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"path/filepath"
	"strings"

	"github.com/aeperfilev/instafix/config"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

func renderImage(src image.Image, resolved config.ResolvedProfile, format config.Format, watermarkText string) (image.Image, error) {
	targetW := format.Width
	targetH := format.Height
	if targetW <= 0 || targetH <= 0 {
		return nil, fmt.Errorf("invalid target size: %dx%d", targetW, targetH)
	}

	dc := gg.NewContext(targetW, targetH)

	img, x, y := fitImage(src, targetW, targetH, resolved.PaddingPercent, resolved.NoUpscale)
	drawBackground(dc, src, resolved.Background, targetW, targetH, img, int(x), int(y))
	imgW := float64(img.Bounds().Dx())
	imgH := float64(img.Bounds().Dy())

	if resolved.BorderWidth > 0 {
		drawBorder(dc, x, y, imgW, imgH, resolved.BorderWidth, resolved.BorderColor)
	}
	dc.DrawImage(img, int(x), int(y))

	if watermarkText != "" && resolved.Watermark != nil {
		if err := drawWatermark(dc, watermarkText, *resolved.Watermark, resolved.AssetsPath); err != nil {
			return nil, err
		}
	}

	return dc.Image(), nil
}

func drawBackground(dc *gg.Context, src image.Image, bg config.Background, width, height int, fitted image.Image, fitX, fitY int) {
	switch strings.ToLower(bg.Type) {
	case "solid":
		setHexColor(dc, bg.Color, 1.0)
		dc.Clear()
	case "average":
		avg := averageColor(src)
		dc.SetRGBA(float64(avg.R)/255.0, float64(avg.G)/255.0, float64(avg.B)/255.0, 1.0)
		dc.Clear()
	case "stretch":
		stretched, err := stretchBackground(src, fitted, width, height, fitX, fitY)
		if err != nil {
			return
		}
		dc.DrawImage(stretched, 0, 0)
	case "blur":
		bgFill := imaging.Fill(src, width, height, imaging.Center, imaging.Lanczos)
		blurred := imaging.Blur(bgFill, bg.BlurRadius)
		dc.DrawImage(blurred, 0, 0)
		if bg.Darken > 0 {
			darken := bg.Darken
			if darken > 1 {
				darken = 1
			}
			dc.SetRGBA(0, 0, 0, darken)
			dc.DrawRectangle(0, 0, float64(width), float64(height))
			dc.Fill()
		}
	default:
		dc.SetRGB(0, 0, 0)
		dc.Clear()
	}
}

func fitImage(src image.Image, targetW, targetH int, paddingPercent float64, noUpscale bool) (image.Image, float64, float64) {
	canvasW := float64(targetW)
	canvasH := float64(targetH)

	padding := math.Max(paddingPercent, 0)
	availW := canvasW * (1.0 - (padding * 2 / 100.0))
	availH := canvasH * (1.0 - (padding * 2 / 100.0))
	if availW < 1 {
		availW = 1
	}
	if availH < 1 {
		availH = 1
	}

	srcW := float64(src.Bounds().Dx())
	srcH := float64(src.Bounds().Dy())
	scale := math.Min(availW/srcW, availH/srcH)

	var fitted image.Image
	if noUpscale && scale > 1.0 {
		fitted = src
	} else {
		fitted = imaging.Fit(src, int(availW), int(availH), imaging.Lanczos)
	}

	x := (canvasW - float64(fitted.Bounds().Dx())) / 2
	y := (canvasH - float64(fitted.Bounds().Dy())) / 2
	return fitted, x, y
}

func drawBorder(dc *gg.Context, x, y, w, h float64, borderWidth int, borderColor string) {
	if borderWidth <= 0 {
		return
	}
	setHexColor(dc, borderColor, 1.0)
	bw := float64(borderWidth)
	dc.DrawRectangle(x-bw, y-bw, w+bw*2, h+bw*2)
	dc.Fill()
}

func drawWatermark(dc *gg.Context, text string, wm config.Watermark, assetsPath string) error {
	fontPath := wm.Font
	if !filepath.IsAbs(fontPath) {
		fontPath = filepath.Join(assetsPath, wm.Font)
	}
	if err := dc.LoadFontFace(fontPath, wm.Size); err != nil {
		return fmt.Errorf("load watermark font: %w", err)
	}

	x, y, ax, ay := anchorForAlign(dc.Width(), dc.Height(), wm.Align, wm.OffsetX, wm.OffsetY)

	if wm.Outline {
		outlineWidth := wm.OutlineWidth
		if outlineWidth == 0 {
			outlineWidth = 2
		}
		setHexColor(dc, wm.OutlineColor, wm.Opacity)
		for dy := -outlineWidth; dy <= outlineWidth; dy++ {
			for dx := -outlineWidth; dx <= outlineWidth; dx++ {
				if dx*dx+dy*dy >= outlineWidth*outlineWidth {
					continue
				}
				dc.DrawStringAnchored(text, x+float64(dx), y+float64(dy), ax, ay)
			}
		}
	}

	setHexColor(dc, wm.Color, wm.Opacity)
	dc.DrawStringAnchored(text, x, y, ax, ay)
	return nil
}

func stretchBackground(src image.Image, fitted image.Image, width, height, x0, y0 int) (image.Image, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid stretch size: %dx%d", width, height)
	}
	fitW := fitted.Bounds().Dx()
	fitH := fitted.Bounds().Dy()
	if fitW == 0 || fitH == 0 {
		return nil, fmt.Errorf("invalid fitted size")
	}

	edgeSample := imaging.Fill(src, fitW, fitH, imaging.Center, imaging.NearestNeighbor)
	topRow := make([]color.NRGBA, fitW)
	bottomRow := make([]color.NRGBA, fitW)
	for x := 0; x < fitW; x++ {
		topRow[x] = colorToNRGBA(edgeSample.At(x, 0))
		bottomRow[x] = colorToNRGBA(edgeSample.At(x, fitH-1))
	}

	bg := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.Draw(bg, bg.Bounds(), image.NewUniform(color.NRGBA{A: 255}), image.Point{}, draw.Src)

	draw.Draw(bg, image.Rect(x0, y0, x0+fitW, y0+fitH), fitted, image.Point{}, draw.Over)

	if x0 > 0 {
		leftStrip := imaging.Crop(fitted, image.Rect(0, 0, 1, fitH))
		leftFill := imaging.Resize(leftStrip, x0, fitH, imaging.NearestNeighbor)
		draw.Draw(bg, image.Rect(0, y0, x0, y0+fitH), leftFill, image.Point{}, draw.Src)
	}
	rightPad := width - (x0 + fitW)
	if rightPad > 0 {
		rightStrip := imaging.Crop(fitted, image.Rect(fitW-1, 0, fitW, fitH))
		rightFill := imaging.Resize(rightStrip, rightPad, fitH, imaging.NearestNeighbor)
		draw.Draw(bg, image.Rect(x0+fitW, y0, width, y0+fitH), rightFill, image.Point{}, draw.Src)
	}
	if y0 > 0 {
		topStrip := imaging.Crop(fitted, image.Rect(0, 0, fitW, 1))
		topFill := imaging.Resize(topStrip, fitW, y0, imaging.NearestNeighbor)
		draw.Draw(bg, image.Rect(x0, 0, x0+fitW, y0), topFill, image.Point{}, draw.Src)
	}
	bottomPad := height - (y0 + fitH)
	if bottomPad > 0 {
		bottomStrip := imaging.Crop(fitted, image.Rect(0, fitH-1, fitW, fitH))
		bottomFill := imaging.Resize(bottomStrip, fitW, bottomPad, imaging.NearestNeighbor)
		draw.Draw(bg, image.Rect(x0, y0+fitH, x0+fitW, height), bottomFill, image.Point{}, draw.Src)
	}

	fillRect(bg, image.Rect(0, 0, x0, y0), topRow[0])
	fillRect(bg, image.Rect(x0+fitW, 0, width, y0), topRow[fitW-1])
	fillRect(bg, image.Rect(0, y0+fitH, x0, height), bottomRow[0])
	fillRect(bg, image.Rect(x0+fitW, y0+fitH, width, height), bottomRow[fitW-1])

	return bg, nil
}

func cornerColorsFlat(x, y, x0, y0, fitW, fitH int,
	topRow, bottomRow, leftCol, rightCol []color.NRGBA) (color.NRGBA, color.NRGBA) {
	if x < x0 && y < y0 {
		return topRow[0], leftCol[0]
	}
	if x >= x0+fitW && y < y0 {
		return topRow[fitW-1], rightCol[0]
	}
	if x < x0 && y >= y0+fitH {
		return bottomRow[0], leftCol[fitH-1]
	}
	return bottomRow[fitW-1], rightCol[fitH-1]
}

func colorToNRGBA(c color.Color) color.NRGBA {
	r, g, b, a := c.RGBA()
	return color.NRGBA{
		R: uint8(r >> 8),
		G: uint8(g >> 8),
		B: uint8(b >> 8),
		A: uint8(a >> 8),
	}
}

func fillRect(dst draw.Image, rect image.Rectangle, c color.NRGBA) {
	if rect.Empty() {
		return
	}
	draw.Draw(dst, rect, image.NewUniform(c), image.Point{}, draw.Src)
}
