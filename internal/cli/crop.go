package cli

import (
	"fmt"

	"github.com/TamasGorgics/gopiq"
)

const cropUsage = `Usage: gopiq crop --x N --y N --width N --height N [flags] <input> <output>

Crop an image using source-image coordinates. The output starts at (0, 0).

Flags:
  --x int           source x coordinate
  --y int           source y coordinate
  --width int       crop width (required, positive)
  --height int      crop height (required, positive)
  --format string   output format: jpeg, png, or gif

Format is inferred from the output extension when omitted.
Use "-" for standard input or standard output.
`

func runCrop(e *env, args []string) error {
	fs := newFlagSet("crop")
	x := fs.Int("x", 0, "source x coordinate")
	y := fs.Int("y", 0, "source y coordinate")
	width := fs.Int("width", 0, "crop width")
	height := fs.Int("height", 0, "crop height")
	var format string
	addFormatFlag(fs, &format)
	if err := parseFlags(e, fs, args, cropUsage); err != nil {
		return err
	}
	if *width <= 0 || *height <= 0 {
		return &usageError{
			err:   fmt.Errorf("width and height must be positive"),
			usage: cropUsage,
		}
	}
	input, output, err := positionalArgs(fs, cropUsage)
	if err != nil {
		return err
	}
	return process(e, input, output, format, func(processor *gopiq.ImageProcessor) {
		processor.Crop(*x, *y, *width, *height)
	}, cropUsage)
}
