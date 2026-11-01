# Changelog

All notable changes to gopiq are documented here. The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-11-01

First stable release. The root `gopiq` package is the supported public API; implementation details remain under `internal/`.

### Added

- Fluent `ImageProcessor` API with chainable operations and error propagation.
- JPEG, PNG, and GIF decoding, plus JPEG, PNG, and GIF encoding.
- Catmull-Rom `Resize`, aspect-preserving `ResizeFit` and `ResizeFill`, and bounds-checked `Crop`.
- `Grayscale` and parallel `GrayscaleFast`, including non-zero-origin and empty-image handling.
- Text watermarks with configurable font, size, color, opacity, and position.
- Horizontal and vertical flips and clockwise 90/180/270-degree rotations.
- `New`, `NewWithPerformanceOptions`, `Clone`, `Image`, `Err`, `ToBytes`, and `ToWriter`.
- Streaming input through `FromReader` and `FromReaderWithLimits`; streaming output through `ToWriter`.
- `DecodeLimits`, `DefaultDecodeLimits`, and `ErrDecodeLimit` for bounded untrusted input.
- Bounded batch processing through `ProcessBatch` and `BatchOptions`, with ordered results, per-item errors, and context cancellation.
- Standalone `gopiq` CLI with `convert`, `resize`, `crop`, `grayscale`, and `watermark` commands, help and version output, stdin/stdout support, and atomic file writes.
- API reference, CLI, streaming, batch, performance, concurrency, architecture, specification, and roadmap documentation.
- Fuzz targets for bounded byte and reader decoding.
- Multi-platform CI for Ubuntu, macOS, and Windows, testing Go 1.27.0 and Go 1.27.3.
- Automated checks that documentation examples compile or parse.

### Changed

- The public facade was separated from the implementation in `internal/engine`; the CLI lives in `internal/cli` with a thin `cmd/gopiq` entry point.
- The minimum supported Go version is 1.27.0.
- GIF output uses a documented lossy policy: a Plan 9 256-color palette, Floyd–Steinberg dithering, white alpha flattening, and a single frame.
- The parallel grayscale threshold was measured and retained at 10,000 pixels.
- Performance documentation now reports reproducible benchmark results without unsupported pooling or SIMD claims.

### Fixed

- Parallel grayscale indexing for non-zero-origin bounds and subimages.
- Empty-image, invalid-dimension, and checked-arithmetic edge cases across processing operations.
- Font-source option ordering and validation for text watermarks.
- GIF decoder registration.
- Misleading `Clone`, concurrency, and performance guarantees in the documentation.

### Known limitations

- Image watermarks (EXT-4) and output formats beyond JPEG, PNG, and GIF (EXT-5) are not part of v1.0.0.
- GIF encoding is lossy and produces a single frame.
- `Clone` is shallow and shares image storage with the original.
- `ToWriter` may leave partial output after a write failure; use `ToBytes` when an in-memory result is required.
- `ProcessBatch` waits for in-flight operations after cancellation; it does not interrupt an operation that has already started.
- A multi-call processing chain is not atomic even though individual methods are synchronized.

### Verification

Release verification covers `go build ./...`, `go test ./...`, `go test -race ./...`, `go vet ./...`, `golangci-lint`, the strict MkDocs build, all documented example commands, short fuzz runs, and Windows/macOS cross-builds.

[1.0.0]: https://github.com/TamasGorgics/gopiq/releases/tag/v1.0.0
