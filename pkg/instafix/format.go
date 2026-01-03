package instafix

import (
	"fmt"
	"math"
	"strings"

	"github.com/aeperfilev/instafix/config"
)

func resolveFormat(formats map[string]config.Format, format config.Format, srcW, srcH int) (config.Format, error) {
	if strings.ToLower(format.Type) == config.FormatTypeFixed {
		return format, nil
	}
	if strings.ToLower(format.Type) != config.FormatTypeAuto {
		return config.Format{}, fmt.Errorf("unknown format type: %s", format.Type)
	}
	if len(format.FromList) == 0 {
		return config.Format{}, fmt.Errorf("auto format requires from_list")
	}
	if srcW <= 0 || srcH <= 0 {
		return config.Format{}, fmt.Errorf("invalid source size")
	}

	best := ""
	bestDiff := math.MaxFloat64
	srcRatio := float64(srcW) / float64(srcH)
	for _, name := range format.FromList {
		candidate, ok := formats[name]
		if !ok {
			return config.Format{}, fmt.Errorf("auto format references unknown format: %s", name)
		}
		if strings.ToLower(candidate.Type) != config.FormatTypeFixed {
			return config.Format{}, fmt.Errorf("auto format references non-fixed format: %s", name)
		}
		ratio := float64(candidate.Width) / float64(candidate.Height)
		diff := math.Abs(srcRatio - ratio)
		if diff < bestDiff {
			bestDiff = diff
			best = name
		}
	}
	if best == "" {
		return config.Format{}, fmt.Errorf("auto format has no valid candidates")
	}
	return formats[best], nil
}
