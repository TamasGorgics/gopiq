package cli

import (
	"fmt"
	"strings"

	"github.com/TamasGorgics/gopiq"
)

func parsePosition(value string) (gopiq.WatermarkPosition, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "top-left":
		return gopiq.PositionTopLeft, nil
	case "top-right":
		return gopiq.PositionTopRight, nil
	case "bottom-left":
		return gopiq.PositionBottomLeft, nil
	case "bottom-right":
		return gopiq.PositionBottomRight, nil
	case "center":
		return gopiq.PositionCenter, nil
	default:
		return 0, fmt.Errorf("unknown watermark position %q", value)
	}
}
