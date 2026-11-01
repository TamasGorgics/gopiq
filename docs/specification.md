# gopiq Specification

**Status:** Draft implementation target · **Scope:** Correctness, safety, supported formats, concurrency, verification, and planned extensions

This document defines the intended behavior of gopiq and the acceptance criteria for fixing the gaps found in the repository audit. “Must” and “must not” are requirements; “should” indicates a recommended implementation choice.

## 1. Goals

gopiq must provide a predictable, chainable image-processing API that:

- builds and tests as a complete Go module, including its examples;
- handles ordinary and non-zero-origin image bounds without panicking;
- reports invalid operations through the processor error chain;
- has clear, testable format and concurrency guarantees; and
- makes performance claims only when supported by reproducible measurements.

The Go 1.27.0 minimum is part of this specification. The CI matrix checks that minimum and the latest supported Go 1.27 patch release.

## 2. Audit baseline and resolution status

The initial audit found the following gaps. M0–M3 have resolved the first five listed items; the remaining two are assigned to M4.

- **Resolved in M0:** Standalone examples previously redeclared `main` in one package, breaking `go build ./...` and `go test ./...`; they now build independently.
- **Resolved in M1:** Parallel grayscale previously mishandled non-zero-origin bounds, and empty images or unsafe dimensions could reach invalid operations.
- **Resolved in M2:** `WithFontPath` previously had no effect; font-source ordering and option validation are now implemented.
- **Resolved in M2:** GIF was described as decodable but its decoder was not registered by the library.
- **Resolved in M3:** `Clone` is now documented and tested as a shallow copy, and the concurrency contract covers per-call synchronization, non-atomic chains, and read-only images.
- **Open for M4:** The RGBA pool is not used by processing operations, and current performance documentation makes pooling/SIMD claims the implementation does not substantiate.
- **Open for M4:** `TestPerformanceComparison` performs timing work as a regular test despite having no performance assertion.

## 3. Requirements

### 3.1 Build, examples, and documentation

- **BLD-1:** Every Go package in the module must compile under `go build ./...` and `go test ./...`.
- **BLD-2:** Each standalone example must be in its own package/directory (or an equivalent structure that avoids duplicate `main` declarations). Example instructions must work from the documented working directory.
- **BLD-3:** Building examples must not require running them or writing generated images into tracked fixture locations.
- **BLD-4:** The MkDocs site must build with the documented requirements, and the specification and roadmap must be discoverable in its navigation.

### 3.2 Image operations and bounds

- **IMG-1:** `Crop(x, y, width, height)` interprets `x` and `y` in the source image's coordinate space. The crop rectangle must have positive dimensions, fit within the source bounds, and be computed without integer overflow. The resulting image has bounds starting at `(0, 0)`.
- **IMG-2:** `Resize(width, height)` requires positive target dimensions and produces a zero-origin image of the requested size. Invalid or unrepresentable buffer sizes must become chain errors rather than arithmetic-driven panics.
- **IMG-3:** Operations must support valid images whose bounds do not start at `(0, 0)`, including subimages. Pixel loops must use coordinates or offsets appropriate to the image's actual bounds and stride.
- **IMG-4:** Operations on empty bounds must have documented behavior and must not panic. Grayscale operations on an empty image are a no-op that preserves its bounds. Crop, resize, and text-watermark operations on an empty image report a chain error because they cannot produce a valid result.
- **IMG-5:** Processing operations must not mutate the source image in place.

### 3.3 Grayscale and performance options

- **GRAY-1:** `Grayscale` and `GrayscaleFast` must support arbitrary valid bounds and subimages without out-of-range buffer access.
- **GRAY-2:** Both methods must use the same documented BT.709 luma calculation and produce identical pixel values for the same input. For the current 8-bit RGBA implementation, the calculation is `uint8(0.2126*R + 0.7152*G + 0.0722*B)`; output RGB channels are equal and alpha is preserved.
- **GRAY-3:** Parallel work must use at least one worker only when there are rows to process, and no more workers than rows or the configured maximum. Empty images must never reach a division by zero.
- **GRAY-4:** `MaxGoroutines <= 0` uses the documented runtime default. `MinSizeForParallel <= 0` means no minimum threshold for a non-empty image; it must not make empty-image processing unsafe. `EnableParallelProcessing == false` always selects the sequential path.
- **GRAY-5:** Default options and threshold behavior must be documented and tested. `SetPerformanceOptions` must remain safe to call concurrently with processor methods.

### 3.4 Watermarks

- **WM-1:** Font-source options are applied in order: the last supplied `WithFontPath` or `WithFontBytes` option selects the source. With neither option, the embedded Go font is used.
- **WM-2:** `WithFontPath` must load and parse the requested file. File-read, parse, and face-construction failures must be returned through the processor error chain.
- **WM-3:** Empty text, a nil option, a nil color, a non-finite, non-positive, or renderer-unrepresentable font size, non-finite offsets, and an unknown position must produce a processor error rather than a panic or silent fallback. Offsets are anchor-relative: positive offsets inset corner anchors and move center-anchored text right/down; negative finite offsets remain valid and may move text partly outside the image.
- **WM-4:** Named positions are calculated relative to the image bounds, including non-zero `Min` coordinates. Text outside the destination is clipped by drawing; it is not silently repositioned.
- **WM-5:** Watermarking must preserve the source image and must be covered by tests for each anchor, custom fonts, errors, and non-zero-origin bounds.

### 3.5 Formats

- **FMT-1:** JPEG and PNG decoding and encoding remain supported and covered by round-trip tests.
- **FMT-2:** GIF decoding is supported independently of whether the importing application imports `image/gif`; the library registers the decoder itself.
- **FMT-3:** GIF encoding remains unsupported until a palette-quantization policy is selected. Calls to encode GIF must return a clear error, not imply successful support.
- **FMT-4:** `FormatFromString`, `ImageFormat.String`, and format comments must agree with the formats actually accepted for decoding and encoding.

### 3.6 Error handling and size safety

- **SAFE-1:** Chainable operations preserve the first error and do not continue processing after it. Export methods return that error.
- **SAFE-2:** Crop arithmetic and image-buffer size calculations must be checked for integer overflow before use or allocation.
- **SAFE-3:** Invalid dimensions and configuration must return errors through the existing API. The library is not required to recover from operating-system out-of-memory termination.
- **SAFE-4:** Decoding limits for encoded bytes, dimensions, and total pixels are a planned extension; callers must not be told that untrusted input is resource-bounded until those limits exist.

### 3.7 Concurrency and cloning contract

To preserve current behavior and avoid an implicit full-image copy, the v1 `Clone` contract is **shallow**:

- **CON-1:** `Clone` copies processor state and performance options but shares the current image reference. Go documentation, README, tests, and concurrency docs must say this; they must not call it a deep copy or promise independent backing pixels.
- **CON-2:** Processing operations treat their input image as read-only and create a new destination. Clones can therefore process the shared source independently as long as the source is not externally mutated.
- **CON-3:** Methods on one processor must be internally race-safe. Individual calls are synchronized, but a multi-call chain is not an atomic transaction and may interleave with calls from another goroutine.
- **CON-4:** `Image()` returns the current image for reading. The caller must treat it, and images passed to constructors, as read-only while the processor or its clones may be used concurrently. Mutating an image through a concrete-type assertion is outside the thread-safety guarantee.
- **CON-5:** Race tests must exercise concurrent reads and state-changing calls on the same processor, not only writes to separate clones.

If independent pixel storage is later needed, it should be provided by an explicitly named deep-copy API rather than changing the meaning of `Clone` silently.

### 3.8 Tests, CI, and performance claims

- **Q-1:** CI must build and test all packages, run the race detector, and include `go vet` and the configured linter. It must build the documentation site.
- **Q-2:** Tests must cover non-zero-origin images, subimages, empty bounds, alpha preservation, grayscale equivalence, invalid sizes/options, font-path loading, and the declared format behavior.
- **Q-3:** Timing comparisons belong in Go benchmarks or a separately invoked performance tool, not ordinary unit tests. Unit tests must not depend on machine speed.
- **Q-4:** Benchmark results must include enough environment and workload detail to be reproduced. Documentation must not present old measurements as universal guarantees.
- **Q-5:** Pooling or SIMD optimizations may be claimed only when the implementation uses them and reproducible benchmarks demonstrate a benefit. Otherwise, unused pool code and unsupported claims should be removed.
- **Q-6:** The supported Go-version policy must be explicit. CI must test the module's declared minimum (currently Go 1.27.0) and the latest supported stable patch release in that Go line.

### 3.9 Command-line interface

- **CLI-1:** The module must provide a `cmd/gopiq` main package that builds with `go build ./cmd/gopiq` and installs with `go install github.com/TamasGorgics/gopiq/cmd/gopiq@latest`. Command parsing and execution logic must live in a testable internal package rather than in `main`.
- **CLI-2:** The CLI must provide documented `convert`, `resize`, `crop`, `grayscale`, and `watermark` commands, plus help and version output. Required arguments, optional flags, and operation semantics must match the public library API.
- **CLI-3:** The CLI must read input from a file or standard input and write output to a file or standard output. Output format is inferred from the destination extension when possible and may be overridden with `--format`. JPEG, PNG, and GIF input are supported; JPEG and PNG output are supported; GIF output must return a clear unsupported error.
- **CLI-4:** Usage and flag errors must return exit code 2; input, processing, encoding, and write errors must return exit code 1; success, help, and version output must return exit code 0. Errors must be written to standard error. Failed output writes must not leave a truncated destination when an atomic write can be used.
- **CLI-5:** The CLI must use only the supported root package and the standard library, not the internal engine directly. Tests must cover every command, format inference and overrides, standard input/output, invalid arguments, unsupported GIF output, exit codes, and output-write failure behavior; CLI usage must be documented in the MkDocs site and linked from the README.

## 4. Planned extensions

These are not prerequisites for the correctness milestones above; they are future capabilities to consider and scope separately.

- **EXT-1 — Streaming I/O:** Add reader-based construction and writer-based encoding so callers need not hold the entire encoded image in a byte slice. Existing byte APIs remain supported wrappers.
- **EXT-2 — Decode limits:** Add configurable limits for encoded input size, width, height, and total pixels, with validation before full decode where possible.
- **EXT-3 — Aspect-ratio operations:** Add explicit fit/fill or proportional-resize APIs with documented rounding, cropping, and upscaling behavior.
- **EXT-4 — Additional operations:** Consider rotation, flips, and image watermarks with clear bounds, anchor, and alpha-compositing behavior.
- **EXT-5 — Additional formats:** Consider GIF encoding only with an explicit quantization policy; assess additional formats and dependencies independently.
- **EXT-6 — Bounded batch processing:** Consider a worker-pool API with bounded concurrency, ordering guarantees, and cancellation behavior for batch workloads.
- **EXT-7 — Benchmark-led optimization:** Profile representative workloads before adding pooling, fixed-point arithmetic, SIMD, or other specialized paths. Keep a simple correct fallback.
- **EXT-8 — Expanded assurance:** Add fuzzing for malformed image data and options, broader platform/version CI, and automated checks that documentation examples stay compilable.

## 5. Release acceptance gate

A release implementing the corrective requirements is ready when all of the following hold:

1. `go build ./...` and `go test ./...` pass.
2. `go test -race ./...`, `go vet ./...`, and the configured linter pass.
3. The MkDocs site builds and all example instructions work as written.
4. Requirements in sections 3.2–3.8 have regression tests or an explicit documented exception. The CLI requirements in section 3.9 apply to releases that ship the CLI.
5. Performance documentation describes measured behavior without unsupported guarantees.
