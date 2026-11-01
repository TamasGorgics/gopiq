package cli

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/TamasGorgics/gopiq"
)

const watermarkUsage = `Usage: gopiq watermark --text TEXT [flags] <input> <output>

Add a text watermark. The embedded Go font is used unless --font is set.

Flags:
  --text string       watermark text (required)
  --font string       path to a .ttf or .otf font file
  --font-size float   font size in points
  --color string      text color: #RGB, #RGBA, #RRGGBB, or #RRGGBBAA
  --position string   top-left, top-right, bottom-left, bottom-right, or center
  --offset string     offset from the anchor as X,Y pixels
  --format string     output format: jpeg or png

Format is inferred from the output extension when omitted.
Use "-" for standard input or standard output.
`

func runWatermark(e *env, args []string) error {
	fs := newFlagSet("watermark")
	text := fs.String("text", "", "watermark text")
	fontPath := fs.String("font", "", "path to a font file")
	fontSizeValue := fs.String("font-size", "", "font size in points")
	colorValue := fs.String("color", "", "text color")
	positionValue := fs.String("position", "", "watermark position")
	offsetValue := fs.String("offset", "", "offset as X,Y pixels")
	var format string
	addFormatFlag(fs, &format)

	if err := parseFlags(e, fs, args, watermarkUsage); err != nil {
		return err
	}
	if strings.TrimSpace(*text) == "" {
		return &usageError{
			err:   fmt.Errorf("--text is required"),
			usage: watermarkUsage,
		}
	}
	input, output, err := positionalArgs(fs, watermarkUsage)
	if err != nil {
		return err
	}

	options, err := watermarkOptions(*fontPath, *fontSizeValue, *colorValue, *positionValue, *offsetValue)
	if err != nil {
		return &usageError{err: err, usage: watermarkUsage}
	}

	return process(e, input, output, format, func(processor *gopiq.ImageProcessor) {
		processor.AddTextWatermark(*text, options...)
	}, watermarkUsage)
}

func watermarkOptions(fontPath, fontSizeValue, colorValue, positionValue, offsetValue string) ([]gopiq.WatermarkOption, error) {
	options := make([]gopiq.WatermarkOption, 0, 5)
	if fontPath != "" {
		options = append(options, gopiq.WithFontPath(fontPath))
	}
	if fontSizeValue != "" {
		size, err := parseFontSize(fontSizeValue)
		if err != nil {
			return nil, err
		}
		options = append(options, gopiq.WithFontSize(size))
	}
	if colorValue != "" {
		value, err := parseColor(colorValue)
		if err != nil {
			return nil, err
		}
		options = append(options, gopiq.WithColor(value))
	}
	if positionValue != "" {
		position, err := parsePosition(positionValue)
		if err != nil {
			return nil, err
		}
		options = append(options, gopiq.WithPosition(position))
	}
	if offsetValue != "" {
		x, y, err := parseOffset(offsetValue)
		if err != nil {
			return nil, err
		}
		options = append(options, gopiq.WithOffset(x, y))
	}
	return options, nil
}

func parseFontSize(value string) (float64, error) {
	size, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || math.IsNaN(size) || math.IsInf(size, 0) || size <= 0 {
		return 0, fmt.Errorf("invalid font size %q: expected a finite positive number", value)
	}
	return size, nil
}

func parseOffset(value string) (float64, float64, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid offset %q: expected X,Y", value)
	}
	x, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil || math.IsNaN(x) || math.IsInf(x, 0) {
		return 0, 0, fmt.Errorf("invalid offset %q: expected finite X,Y values", value)
	}
	y, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil || math.IsNaN(y) || math.IsInf(y, 0) {
		return 0, 0, fmt.Errorf("invalid offset %q: expected finite X,Y values", value)
	}
	return x, y, nil
}
