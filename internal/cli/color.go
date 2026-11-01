package cli

import (
	"encoding/hex"
	"fmt"
	"image/color"
	"strconv"
	"strings"
)

func parseColor(value string) (color.Color, error) {
	digits := strings.TrimPrefix(strings.TrimSpace(value), "#")

	switch len(digits) {
	case 3, 4:
		var channels [4]uint8
		channels[3] = 255
		for i := 0; i < len(digits); i++ {
			nibble, err := strconv.ParseUint(string(digits[i]), 16, 8)
			if err != nil {
				return nil, invalidColorError(value)
			}
			channels[i] = uint8(nibble * 17)
		}
		return color.NRGBA{R: channels[0], G: channels[1], B: channels[2], A: channels[3]}, nil
	case 6, 8:
		channels, err := hex.DecodeString(digits)
		if err != nil {
			return nil, invalidColorError(value)
		}
		alpha := uint8(255)
		if len(channels) == 4 {
			alpha = channels[3]
		}
		return color.NRGBA{R: channels[0], G: channels[1], B: channels[2], A: alpha}, nil
	default:
		return nil, invalidColorError(value)
	}
}

func invalidColorError(value string) error {
	return fmt.Errorf("invalid color %q: expected #RGB, #RGBA, #RRGGBB, or #RRGGBBAA", value)
}
