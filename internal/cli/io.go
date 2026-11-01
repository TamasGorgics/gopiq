package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/TamasGorgics/gopiq"
)

// G304 is intentional: the CLI reads caller-supplied input paths.

var errGIFOutput = errors.New("GIF output is not supported because palette quantization is not implemented")

func process(e *env, input, output, formatValue string, apply func(*gopiq.ImageProcessor), usage string) error {
	format, err := resolveOutputFormat(output, formatValue)
	if err != nil {
		if errors.Is(err, errGIFOutput) {
			return err
		}
		return &usageError{err: err, usage: usage}
	}

	data, err := readInput(e.stdin, input)
	if err != nil {
		return fmt.Errorf("read %s: %w", inputName(input), err)
	}

	processor := gopiq.FromBytes(data)
	if err := processor.Err(); err != nil {
		return fmt.Errorf("decode %s: %w", inputName(input), err)
	}

	if apply != nil {
		apply(processor)
		if err := processor.Err(); err != nil {
			return fmt.Errorf("process %s: %w", inputName(input), err)
		}
	}

	encoded, err := processor.ToBytes(format)
	if err != nil {
		return fmt.Errorf("encode %s: %w", outputName(output), err)
	}
	if err := writeOutput(e.stdout, output, encoded); err != nil {
		return fmt.Errorf("write %s: %w", outputName(output), err)
	}
	return nil
}

func readInput(stdin io.Reader, path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(stdin)
	}
	return os.ReadFile(path) //nolint:gosec // The CLI intentionally reads caller-supplied paths.
}

func writeOutput(stdout io.Writer, path string, data []byte) error {
	if path == "-" {
		_, err := stdout.Write(data)
		return err
	}
	return writeFileAtomic(path, data)
}

func writeFileAtomic(path string, data []byte) (err error) {
	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			if removeErr := os.Remove(tmpName); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
				err = errors.Join(err, removeErr)
			}
		}
	}()

	if err = tmp.Chmod(0o644); err != nil {
		return errors.Join(err, tmp.Close())
	}
	if _, err = tmp.Write(data); err != nil {
		return errors.Join(err, tmp.Close())
	}
	if closeErr := tmp.Close(); closeErr != nil {
		return closeErr
	}
	err = os.Rename(tmpName, path)
	return err
}

func resolveOutputFormat(output, formatValue string) (gopiq.ImageFormat, error) {
	if formatValue != "" {
		format := gopiq.FormatFromString(formatValue)
		if format == gopiq.FormatUnknown {
			return gopiq.FormatUnknown, fmt.Errorf("unknown output format %q", formatValue)
		}
		if format == gopiq.FormatGIF {
			return gopiq.FormatUnknown, errGIFOutput
		}
		return format, nil
	}

	if output == "-" {
		return gopiq.FormatUnknown, fmt.Errorf("--format is required when writing to stdout")
	}

	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(output)), ".")
	format := gopiq.FormatFromString(extension)
	if format == gopiq.FormatUnknown {
		return gopiq.FormatUnknown, fmt.Errorf("cannot infer output format from %q; use --format", output)
	}
	if format == gopiq.FormatGIF {
		return gopiq.FormatUnknown, errGIFOutput
	}
	return format, nil
}

func inputName(path string) string {
	if path == "-" {
		return "stdin"
	}
	return path
}

func outputName(path string) string {
	if path == "-" {
		return "stdout"
	}
	return path
}
