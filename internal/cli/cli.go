// Package cli implements the gopiq command-line interface.
package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"runtime/debug"
	"strings"
)

const (
	exitSuccess = 0
	exitFailure = 1
	exitUsage   = 2
)

const generalUsage = `Usage: gopiq <command> [flags] <input> <output>

Commands:
  convert    Decode and re-encode an image
  resize     Resize an image to exact dimensions
  crop       Crop an image using source coordinates
  grayscale  Convert an image to grayscale
  watermark  Add a text watermark
  help       Show help for a command
  version    Show the gopiq version

Global flags:
  --help     Show help
  --version  Show the gopiq version

Input and output may be "-" for standard input or standard output.
Run "gopiq help <command>" for command-specific flags.
`

type env struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

type command struct {
	name  string
	usage string
	run   func(*env, []string) error
}

var commands = map[string]command{
	"convert":   {name: "convert", usage: convertUsage, run: runConvert},
	"resize":    {name: "resize", usage: resizeUsage, run: runResize},
	"crop":      {name: "crop", usage: cropUsage, run: runCrop},
	"grayscale": {name: "grayscale", usage: grayscaleUsage, run: runGrayscale},
	"watermark": {name: "watermark", usage: watermarkUsage, run: runWatermark},
}

// Main runs the gopiq CLI and returns the process exit code.
func Main(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if stdin == nil {
		stdin = strings.NewReader("")
	}
	if stdout == nil {
		stdout = io.Discard
	}
	if stderr == nil {
		stderr = io.Discard
	}

	e := &env{stdin: stdin, stdout: stdout, stderr: stderr}
	if len(args) == 0 {
		fmt.Fprint(stderr, generalUsage)
		return exitUsage
	}

	switch args[0] {
	case "help", "-h", "--help":
		return e.finish(runHelp(e, args[1:]))
	case "version", "--version":
		fmt.Fprintf(stdout, "gopiq %s\n", version())
		return exitSuccess
	}

	cmd, ok := commands[args[0]]
	if !ok {
		return e.finish(&usageError{
			err:   fmt.Errorf("unknown command %q", args[0]),
			usage: generalUsage,
		})
	}
	return e.finish(cmd.run(e, args[1:]))
}

func (e *env) finish(err error) int {
	if err == nil || errors.Is(err, flag.ErrHelp) {
		return exitSuccess
	}

	var usageErr *usageError
	if errors.As(err, &usageErr) {
		fmt.Fprintf(e.stderr, "gopiq: %v\n\n", usageErr.err)
		fmt.Fprint(e.stderr, usageErr.usage)
		return exitUsage
	}

	fmt.Fprintf(e.stderr, "gopiq: %v\n", err)
	return exitFailure
}

func runHelp(e *env, args []string) error {
	if len(args) == 0 {
		fmt.Fprint(e.stdout, generalUsage)
		return nil
	}
	cmd, ok := commands[args[0]]
	if !ok {
		return &usageError{
			err:   fmt.Errorf("unknown command %q", args[0]),
			usage: generalUsage,
		}
	}
	fmt.Fprint(e.stdout, cmd.usage)
	return nil
}

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return "dev"
	}
	return info.Main.Version
}

type usageError struct {
	err   error
	usage string
}

func (e *usageError) Error() string { return e.err.Error() }
func (e *usageError) Unwrap() error { return e.err }

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func parseFlags(e *env, fs *flag.FlagSet, args []string, usage string) error {
	if hasHelpFlag(args) {
		fmt.Fprint(e.stdout, usage)
		return flag.ErrHelp
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(e.stdout, usage)
			return flag.ErrHelp
		}
		return &usageError{err: err, usage: usage}
	}
	return nil
}

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func positionalArgs(fs *flag.FlagSet, usage string) (string, string, error) {
	if fs.NArg() != 2 {
		return "", "", &usageError{
			err:   fmt.Errorf("expected exactly two positional arguments: input output"),
			usage: usage,
		}
	}
	return fs.Arg(0), fs.Arg(1), nil
}

func addFormatFlag(fs *flag.FlagSet, value *string) {
	fs.StringVar(value, "format", "", "output format: jpeg or png")
}
