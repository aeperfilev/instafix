package instafix

import "strings"

func anchorForAlign(width, height int, align string, offsetX, offsetY float64) (float64, float64, float64, float64) {
	w := float64(width)
	h := float64(height)
	align = strings.ToLower(strings.TrimSpace(align))
	if align == "" {
		align = "bottom-center"
	}

	hAnchor := "center"
	vAnchor := "bottom"
	parts := strings.Split(align, "-")
	if len(parts) == 2 {
		vAnchor = parts[0]
		hAnchor = parts[1]
	} else if len(parts) == 1 {
		switch parts[0] {
		case "top", "bottom":
			vAnchor = parts[0]
			hAnchor = "center"
		case "left", "right":
			vAnchor = "center"
			hAnchor = parts[0]
		case "center", "middle":
			vAnchor = "center"
			hAnchor = "center"
		default:
			hAnchor = parts[0]
		}
	}

	var x, y float64
	var ax, ay float64

	switch hAnchor {
	case "left":
		x = offsetX
		ax = 0
	case "right":
		x = w - offsetX
		ax = 1
	default:
		x = w/2 + offsetX
		ax = 0.5
	}

	switch vAnchor {
	case "top":
		y = offsetY
		ay = 0
	case "center", "middle":
		y = h/2 + offsetY
		ay = 0.5
	default:
		y = h - offsetY
		ay = 1
	}

	return x, y, ax, ay
}
