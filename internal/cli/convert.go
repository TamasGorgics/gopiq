package cli

const convertUsage = `Usage: gopiq convert [flags] <input> <output>

Decode an image and encode it in the output format.

Flags:
  --format string   output format: jpeg, png, or gif

Format is inferred from the output extension when omitted.
Use "-" for standard input or standard output.
`

func runConvert(e *env, args []string) error {
	fs := newFlagSet("convert")
	var format string
	addFormatFlag(fs, &format)
	if err := parseFlags(e, fs, args, convertUsage); err != nil {
		return err
	}
	input, output, err := positionalArgs(fs, convertUsage)
	if err != nil {
		return err
	}
	return process(e, input, output, format, nil, convertUsage)
}
