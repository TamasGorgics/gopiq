# Gopiq Examples

Each example is a standalone Go program in its own package, so the examples can be compiled together with `go build ./...`.

## Shared input and generated output

- `sample_image.png` is the shared 800x600 input image.
- Each program reads `examples/sample_image.png` relative to the repository root.
- Generated files are written to `examples/<name>/output/`. These output directories are ignored by Git; the existing image files in `examples/` are left untouched.

## Programs

| Directory | Demonstrates |
| --- | --- |
| `basic` | Loading, grayscale, resize, crop, and chaining |
| `watermark` | Watermark positions and custom styling |
| `grayscale` | Regular and parallel grayscale, performance options |
| `resize` | Resize scenarios and aspect-ratio calculations |
| `crop` | Corner, center, grid, and invalid crops |
| `comprehensive` | Processing pipelines, formats, errors, and performance examples |

## Run an example

Run commands from the repository root:

```bash
go run ./examples/basic
go run ./examples/watermark
go run ./examples/grayscale
go run ./examples/resize
go run ./examples/crop
go run ./examples/comprehensive
```

Build all packages and examples without running them:

```bash
go build ./...
```

## Generated files

- `basic/output/`: grayscale, resized, and cropped images.
- `watermark/output/`: positional and custom-style watermarks.
- `grayscale/output/`: regular, fast, and custom-performance grayscale images.
- `resize/output/`: thumbnail, medium, large, square, wide, tall, chained, and aspect-ratio results.
- `crop/output/`: corner, center, strip, chained, and 4x4 grid crops.
- `comprehensive/output/`: pipeline, watermark, grayscale, format, and efficiency results.
