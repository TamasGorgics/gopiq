package cli

import (
	"fmt"

	"github.com/TamasGorgics/gopiq"
)

const resizeUsage = `Usage: gopiq resize --width N --height N [flags] <input> <output>

Resize an image to exact dimensions using Catmull-Rom interpolation.

Flags:
  --width int       target width (required, positive)
  --height int      target height (required, positive)
  --format string   output format: jpeg or png

Format is inferred from the output extension when omitted.
Use "-" for standard input or standard output.
`

func runResize(e *env, args []string) error {
	fs := newFlagSet("resize")
	width := fs.Int("width", 0, "target width")
	height := fs.Int("height", 0, "target height")
	var format string
	addFormatFlag(fs, &format)
	if err := parseFlags(e, fs, args, resizeUsage); err != nil {
		return err
	}
	if *width <= 0 || *height <= 0 {
		return &usageError{
			err:   fmt.Errorf("width and height must be positive"),
			usage: resizeUsage,
		}
	}
	input, output, err := positionalArgs(fs, resizeUsage)
	if err != nil {
		return err
	}
	return process(e, input, output, format, func(processor *gopiq.ImageProcessor) {
		processor.Resize(*width, *height)
	}, resizeUsage)
}
