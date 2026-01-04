package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	FormatTypeFixed = "fixed"
	FormatTypeAuto  = "auto"
)

var (
	ErrProfileNotFound    = errors.New("profile not found")
	ErrFormatNotFound     = errors.New("format not found")
	ErrBackgroundNotFound = errors.New("background not found")
	ErrWatermarkNotFound  = errors.New("watermark not found")
)

type Config struct {
	Settings    Settings              `toml:"settings"`
	Backgrounds map[string]Background `toml:"backgrounds"`
	Watermarks  map[string]Watermark  `toml:"watermarks"`
	Formats     map[string]Format     `toml:"formats"`
	Profiles    map[string]Profile    `toml:"profiles"`
}

type Settings struct {
	JpegQuality int    `toml:"jpeg_quality"`
	AssetsPath  string `toml:"assets_path"`
}

type Profile struct {
	BackgroundRef  string   `toml:"background_ref"`
	WatermarkRef   string   `toml:"watermark_ref"`
	FormatRef      string   `toml:"format_ref"`
	PaddingPercent *float64 `toml:"padding_percent"`
	BorderWidth    int      `toml:"border_width"`
	BorderColor    string   `toml:"border_color"`
	NoUpscale      bool     `toml:"no_upscale"`
	JpegQuality    int      `toml:"jpeg_quality"`
}

type Format struct {
	Type           string   `toml:"type"`
	Width          int      `toml:"width"`
	Height         int      `toml:"height"`
	FromList       []string `toml:"from_list"`
	PaddingPercent float64  `toml:"padding_percent"`
}

type Background struct {
	Type       string  `toml:"type"`
	Color      string  `toml:"color"`
	BlurRadius float64 `toml:"blur_radius"`
	Darken     float64 `toml:"darken"`
}

type Watermark struct {
	Font         string  `toml:"font"`
	Size         float64 `toml:"size"`
	Color        string  `toml:"color"`
	Opacity      float64 `toml:"opacity"`
	Align        string  `toml:"align"`
	OffsetX      float64 `toml:"offset_x"`
	OffsetY      float64 `toml:"offset_y"`
	Outline      bool    `toml:"outline"`
	OutlineColor string  `toml:"outline_color"`
	OutlineWidth float64 `toml:"outline_width"`
}

type ResolvedProfile struct {
	Name           string
	Background     Background
	Watermark      *Watermark
	Format         Format
	FormatName     string
	PaddingPercent float64
	BorderWidth    int
	BorderColor    string
	NoUpscale      bool
	JpegQuality    int
	AssetsPath     string
}

// Load reads a TOML config file and validates it.
func Load(path string) (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// LoadDefault finds and loads the default config path.
func LoadDefault() (Config, string, error) {
	path, err := FindDefaultPath()
	if err != nil {
		return Config{}, "", err
	}
	cfg, err := Load(path)
	if err != nil {
		return Config{}, "", err
	}
	return cfg, path, nil
}

// FindDefaultPath tries to locate a config without explicit path.
func FindDefaultPath() (string, error) {
	if envPath := strings.TrimSpace(os.Getenv("INSTAFIX_CONFIG")); envPath != "" {
		if fileExists(envPath) {
			return envPath, nil
		}
		return "", fmt.Errorf("config path from INSTAFIX_CONFIG not found: %s", envPath)
	}

	cwd, err := os.Getwd()
	if err == nil {
		candidates := []string{
			filepath.Join(cwd, "profiles.toml"),
			filepath.Join(cwd, "config", "profiles.toml"),
		}
		for _, p := range candidates {
			if fileExists(p) {
				return p, nil
			}
		}
	}

	exe, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exe)
		p := filepath.Join(exeDir, "profiles.toml")
		if fileExists(p) {
			return p, nil
		}
	}

	return "", errors.New("default config not found (looked for profiles.toml)")
}

func (c *Config) Validate() error {
	if c.Settings.JpegQuality == 0 {
		c.Settings.JpegQuality = 90
	}
	if c.Settings.JpegQuality < 1 || c.Settings.JpegQuality > 100 {
		return fmt.Errorf("settings.jpeg_quality out of range: %d", c.Settings.JpegQuality)
	}
	if strings.TrimSpace(c.Settings.AssetsPath) == "" {
		c.Settings.AssetsPath = "assets"
	}

	for name, format := range c.Formats {
		if err := validateFormat(name, format); err != nil {
			return err
		}
	}
	for name, format := range c.Formats {
		if strings.ToLower(strings.TrimSpace(format.Type)) != FormatTypeAuto {
			continue
		}
		for _, ref := range format.FromList {
			candidate, ok := c.Formats[ref]
			if !ok {
				return fmt.Errorf("formats.%s references unknown format: %s", name, ref)
			}
			if strings.ToLower(strings.TrimSpace(candidate.Type)) != FormatTypeFixed {
				return fmt.Errorf("formats.%s references non-fixed format: %s", name, ref)
			}
		}
	}
	for name, bg := range c.Backgrounds {
		if err := validateBackground(name, bg); err != nil {
			return err
		}
	}
	for name, wm := range c.Watermarks {
		if err := validateWatermark(name, wm); err != nil {
			return err
		}
	}
	for name, profile := range c.Profiles {
		if err := c.validateProfile(name, profile); err != nil {
			return err
		}
	}

	return nil
}

func (c Config) validateProfile(name string, profile Profile) error {
	if strings.TrimSpace(profile.BackgroundRef) == "" {
		return fmt.Errorf("profiles.%s.background_ref is required", name)
	}
	if strings.TrimSpace(profile.FormatRef) == "" {
		return fmt.Errorf("profiles.%s.format_ref is required", name)
	}
	if _, ok := c.Backgrounds[profile.BackgroundRef]; !ok {
		return fmt.Errorf("profiles.%s.background_ref not found: %s", name, profile.BackgroundRef)
	}
	if _, ok := c.Formats[profile.FormatRef]; !ok {
		return fmt.Errorf("profiles.%s.format_ref not found: %s", name, profile.FormatRef)
	}
	if profile.WatermarkRef != "" {
		if _, ok := c.Watermarks[profile.WatermarkRef]; !ok {
			return fmt.Errorf("profiles.%s.watermark_ref not found: %s", name, profile.WatermarkRef)
		}
	}
	if profile.JpegQuality != 0 && (profile.JpegQuality < 1 || profile.JpegQuality > 100) {
		return fmt.Errorf("profiles.%s.jpeg_quality out of range: %d", name, profile.JpegQuality)
	}
	if profile.PaddingPercent != nil && (*profile.PaddingPercent < 0 || *profile.PaddingPercent > 50) {
		return fmt.Errorf("profiles.%s.padding_percent must be 0..50", name)
	}
	return nil
}

// ResolveProfile merges defaults and returns fully resolved references.
func (c Config) ResolveProfile(name string) (ResolvedProfile, error) {
	profile, ok := c.Profiles[name]
	if !ok {
		return ResolvedProfile{}, fmt.Errorf("%w: %s", ErrProfileNotFound, name)
	}

	format, ok := c.Formats[profile.FormatRef]
	if !ok {
		return ResolvedProfile{}, fmt.Errorf("%w: %s", ErrFormatNotFound, profile.FormatRef)
	}
	if err := validateFormat(profile.FormatRef, format); err != nil {
		return ResolvedProfile{}, err
	}

	background, ok := c.Backgrounds[profile.BackgroundRef]
	if !ok {
		return ResolvedProfile{}, fmt.Errorf("%w: %s", ErrBackgroundNotFound, profile.BackgroundRef)
	}
	if err := validateBackground(profile.BackgroundRef, background); err != nil {
		return ResolvedProfile{}, err
	}

	var watermark *Watermark
	if profile.WatermarkRef != "" {
		wm, ok := c.Watermarks[profile.WatermarkRef]
		if !ok {
			return ResolvedProfile{}, fmt.Errorf("%w: %s", ErrWatermarkNotFound, profile.WatermarkRef)
		}
		if err := validateWatermark(profile.WatermarkRef, wm); err != nil {
			return ResolvedProfile{}, err
		}
		watermark = &wm
	}

	jpegQuality := c.Settings.JpegQuality
	if profile.JpegQuality != 0 {
		jpegQuality = profile.JpegQuality
	}

	assetsPath := strings.TrimSpace(c.Settings.AssetsPath)
	if assetsPath == "" {
		assetsPath = "assets"
	}

	paddingPercent := format.PaddingPercent
	if profile.PaddingPercent != nil {
		paddingPercent = *profile.PaddingPercent
	}

	return ResolvedProfile{
		Name:           name,
		Background:     background,
		Watermark:      watermark,
		Format:         format,
		FormatName:     profile.FormatRef,
		PaddingPercent: paddingPercent,
		BorderWidth:    profile.BorderWidth,
		BorderColor:    profile.BorderColor,
		NoUpscale:      profile.NoUpscale,
		JpegQuality:    jpegQuality,
		AssetsPath:     assetsPath,
	}, nil
}

func validateFormat(name string, format Format) error {
	switch strings.ToLower(strings.TrimSpace(format.Type)) {
	case FormatTypeFixed:
		if format.Width <= 0 || format.Height <= 0 {
			return fmt.Errorf("formats.%s fixed format requires width and height", name)
		}
		if len(format.FromList) > 0 {
			return fmt.Errorf("formats.%s fixed format must not have from_list", name)
		}
	case FormatTypeAuto:
		if len(format.FromList) == 0 {
			return fmt.Errorf("formats.%s auto format requires from_list", name)
		}
		if format.Width != 0 || format.Height != 0 {
			return fmt.Errorf("formats.%s auto format must not have width/height", name)
		}
	default:
		return fmt.Errorf("formats.%s has unknown type: %s", name, format.Type)
	}
	if format.PaddingPercent < 0 || format.PaddingPercent > 50 {
		return fmt.Errorf("formats.%s.padding_percent must be 0..50", name)
	}
	return nil
}

func validateBackground(name string, bg Background) error {
	switch strings.ToLower(strings.TrimSpace(bg.Type)) {
	case "solid", "blur", "stretch", "average":
	default:
		return fmt.Errorf("backgrounds.%s has unknown type: %s", name, bg.Type)
	}
	if bg.Darken < 0 || bg.Darken > 1 {
		return fmt.Errorf("backgrounds.%s.darken must be 0..1", name)
	}
	return nil
}

func validateWatermark(name string, wm Watermark) error {
	if strings.TrimSpace(wm.Font) == "" {
		return fmt.Errorf("watermarks.%s.font is required", name)
	}
	if wm.Size <= 0 {
		return fmt.Errorf("watermarks.%s.size must be > 0", name)
	}
	if wm.Opacity < 0 || wm.Opacity > 1 {
		return fmt.Errorf("watermarks.%s.opacity must be 0..1", name)
	}
	if wm.OutlineWidth < 0 {
		return fmt.Errorf("watermarks.%s.outline_width must be >= 0", name)
	}
	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
