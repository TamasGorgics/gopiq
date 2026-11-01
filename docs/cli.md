# Command-Line Interface

The `gopiq` CLI exposes the library's common image operations as a standalone command.

## Installation

Install the latest version with Go:

```bash
go install github.com/TamasGorgics/gopiq/cmd/gopiq@latest
```

From a checkout, build or install it with:

```bash
make build
make install
```

`make build` writes `bin/gopiq`. `go build ./cmd/gopiq` is also supported.

## Usage

```text
gopiq <command> [flags] <input> <output>
```

Input and output may be `-` for standard input or standard output. Flags must come before the positional paths because the standard `flag` parser stops at the first non-flag argument.

### Convert

Decode an image and re-encode it in the output format.

```bash
gopiq convert photo.jpg photo.png
gopiq convert --format png photo.jpg output
```

### Resize

Resize to exact dimensions using Catmull-Rom interpolation.

```bash
gopiq resize --width 800 --height 600 photo.jpg resized.png
```

Both dimensions are required and must be positive. Aspect-ratio fit/fill is not part of this release.

### Crop

Crop using source-image coordinates. The result starts at `(0, 0)`.

```bash
gopiq crop --x 100 --y 50 --width 400 --height 300 photo.jpg cropped.png
```

### Grayscale

Convert to grayscale using the library's BT.709 calculation.

```bash
gopiq grayscale photo.jpg gray.png
gopiq grayscale --fast photo.jpg gray-fast.png
```

`--fast` selects `GrayscaleFast`, which uses the configured parallel path.

### Watermark

Add a text watermark.

```bash
gopiq watermark \
  --text "© 2026" \
  --position bottom-right \
  --font-size 24 \
  --color "#ffffffaa" \
  photo.jpg marked.png
```

Watermark flags:

| Flag | Description |
|---|---|
| `--text` | Watermark text; required |
| `--font` | Path to a `.ttf` or `.otf` font file |
| `--font-size` | Font size in points |
| `--color` | `#RGB`, `#RGBA`, `#RRGGBB`, or `#RRGGBBAA` |
| `--position` | `top-left`, `top-right`, `bottom-left`, `bottom-right`, or `center` |
| `--offset` | Offset from the anchor as `X,Y` pixels |

Without watermark flags, the library defaults are used: the embedded Go font, font size 24, bottom-right position, and offset `10,10`. Negative offsets must use the equals form:

```bash
gopiq watermark --text "© 2026" --offset=-5,-5 photo.jpg marked.png
```

### Help and version

```bash
gopiq help
gopiq help resize
gopiq resize --help
gopiq version
gopiq --version
```

## Input and output formats

- Input format is detected from the image content. JPEG, PNG, and GIF are supported.
- Output format is inferred from `.jpg`, `.jpeg`, `.png`, or `.gif`.
- `--format jpeg|png|gif` overrides the output extension.
- When output is `-`, `--format` is required.
- GIF output uses the Go standard-library encoder: Plan 9 256-color palette, Floyd–Steinberg dithering, white alpha flattening, and a single frame.

Examples:

```bash
cat photo.jpg | gopiq grayscale --format png - - > gray.png
gopiq convert --format png photo.jpg - > photo.png
```

Output files are written through a temporary file and renamed into place where supported, so a failed operation does not leave a truncated destination. Existing output files are replaced.

## Exit codes

| Code | Meaning |
|---:|---|
| `0` | Success, help, or version |
| `1` | Input, decode, processing, encoding, or write failure |
| `2` | Usage, flag, argument, color, position, or output-format error |

Errors are written to standard error. Successful image output written to stdout contains only the encoded image bytes.

## Current limitations

- One operation per invocation; use standard pipelines for multiple steps.
- No aspect-ratio resize fit/fill.
- No batch or directory processing.
- No streaming decode limits; the CLI reads the complete input before processing.
- GIF output is lossy and single-frame.
