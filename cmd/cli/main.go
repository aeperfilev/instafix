package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aeperfilev/instafix/config"
	"github.com/aeperfilev/instafix/pkg/instafix"

	"github.com/disintegration/imaging"
)

func main() {
	var (
		configPath  string
		profileName string
		watermark   string
		outputPath  string
	)

	flag.StringVar(&configPath, "config", "", "Path to profiles.toml (optional)")
	flag.StringVar(&profileName, "profile", "default", "Profile name to apply")
	flag.StringVar(&watermark, "watermark", "", "Watermark text (optional)")
	flag.StringVar(&outputPath, "out", "", "Output image path (optional)")
	flag.Parse()

	if flag.NArg() < 1 {
		exitWithError("input image path is required")
	}
	inputPath := flag.Arg(0)
	if outputPath == "" {
		outputPath = defaultOutputPath(inputPath)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		exitWithError(err.Error())
	}

	processor, err := instafix.NewProcessor(cfg)
	if err != nil {
		exitWithError(err.Error())
	}

	srcFile, err := os.Open(inputPath)
	if err != nil {
		exitWithError(fmt.Sprintf("open input: %v", err))
	}
	defer srcFile.Close()

	srcImg, err := instafix.DecodeImage(srcFile, inputPath)
	if err != nil {
		exitWithError(fmt.Sprintf("decode input image: %v", err))
	}

	result, quality, err := processor.Process(srcImg, profileName, watermark)
	if err != nil {
		exitWithError(err.Error())
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		exitWithError(fmt.Sprintf("create output dir: %v", err))
	}
	outFile, err := os.Create(outputPath)
	if err != nil {
		exitWithError(fmt.Sprintf("create output: %v", err))
	}
	defer outFile.Close()

	if err := imaging.Encode(outFile, result, imaging.JPEG, imaging.JPEGQuality(quality)); err != nil {
		exitWithError(fmt.Sprintf("encode output: %v", err))
	}
}

func loadConfig(path string) (config.Config, error) {
	if strings.TrimSpace(path) == "" {
		cfg, _, err := config.LoadDefault()
		return cfg, err
	}
	return config.Load(path)
}

func defaultOutputPath(inputPath string) string {
	base := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	dir := filepath.Dir(inputPath)
	return filepath.Join(dir, base+"_instafix.jpg")
}

func exitWithError(msg string) {
	fmt.Fprintln(os.Stderr, "instafix:", msg)
	os.Exit(1)
}
