package instafix

import (
	"errors"
	"fmt"
	"image"

	"github.com/aeperfilev/instafix/config"
)

// UserError means the request parameters are invalid.
type UserError struct {
	Err error
}

func (e UserError) Error() string {
	return e.Err.Error()
}

func (e UserError) Unwrap() error {
	return e.Err
}

// Processor applies profiles from the config to images.
type Processor struct {
	cfg config.Config
}

// NewProcessor validates the config and returns a processor.
func NewProcessor(cfg config.Config) (*Processor, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Processor{cfg: cfg}, nil
}

// Process applies a profile to the source image.
// It returns the resulting image and JPEG quality to use for encoding.
func (p *Processor) Process(src image.Image, profileName, watermarkText string) (image.Image, int, error) {
	resolved, err := p.cfg.ResolveProfile(profileName)
	if err != nil {
		if errors.Is(err, config.ErrProfileNotFound) {
			return nil, 0, UserError{Err: err}
		}
		return nil, 0, err
	}

	format, err := resolveFormat(p.cfg.Formats, resolved.Format, src.Bounds().Dx(), src.Bounds().Dy())
	if err != nil {
		return nil, 0, err
	}

	if watermarkText != "" && resolved.Watermark == nil {
		return nil, 0, UserError{Err: fmt.Errorf("watermark text provided, but profile has no watermark_ref")}
	}

	result, err := renderImage(src, resolved, format, watermarkText)
	if err != nil {
		return nil, 0, err
	}

	return result, resolved.JpegQuality, nil
}
