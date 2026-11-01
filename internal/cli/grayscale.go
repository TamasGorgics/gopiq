package cli

import "github.com/TamasGorgics/gopiq"

const grayscaleUsage = `Usage: gopiq grayscale [flags] <input> <output>

Convert an image to grayscale using the BT.709 luma calculation.

Flags:
  --fast            use the parallel GrayscaleFast path
  --format string   output format: jpeg or png

Format is inferred from the output extension when omitted.
Use "-" for standard input or standard output.
`

func runGrayscale(e *env, args []string) error {
	fs := newFlagSet("grayscale")
	fast := fs.Bool("fast", false, "use the parallel grayscale path")
	var format string
	addFormatFlag(fs, &format)
	if err := parseFlags(e, fs, args, grayscaleUsage); err != nil {
		return err
	}
	input, output, err := positionalArgs(fs, grayscaleUsage)
	if err != nil {
		return err
	}
	return process(e, input, output, format, func(processor *gopiq.ImageProcessor) {
		if *fast {
			processor.GrayscaleFast()
			return
		}
		processor.Grayscale()
	}, grayscaleUsage)
}
