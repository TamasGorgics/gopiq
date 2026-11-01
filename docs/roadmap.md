# gopiq Roadmap

This roadmap turns the [specification](specification.md) into ordered, verifiable milestones. It is dependency-ordered, not calendar-based; no dates or delivery estimates are implied. Milestones are **Planned** unless marked **Complete**.

## Milestones

### M0 — Make the repository buildable

**Status:** Complete · **Priority:** P0 · **Requirements:** BLD-1–BLD-3

- Move each standalone example into an independently buildable package/directory, or use an equivalent layout that avoids duplicate `main` functions.
- Correct example asset paths and run instructions after reorganizing.
- Keep generated example output separate from tracked input fixtures.
- Ensure the CI build/test commands cover the resulting package layout.

**Exit criteria:** `go build ./...` and `go test ./...` both complete successfully, and each documented example can be built independently.

### M1 — Correct core image processing and guard edge cases

**Status:** Complete · **Priority:** P1 · **Requirements:** IMG-1–IMG-5, GRAY-1–GRAY-5, SAFE-1–SAFE-3

- Fix parallel grayscale indexing for non-zero-origin bounds and subimages.
- Handle empty images and performance-option edge cases without zero-worker division or invalid pixel access.
- Verify exact grayscale agreement between sequential and parallel paths, including alpha.
- Check crop coordinate arithmetic and image-buffer size calculations before use or allocation.
- Define and test empty-source behavior for crop, resize, grayscale, and watermarking.

**Exit criteria:** Regression tests cover zero-origin and non-zero-origin images, subimages, empty bounds, alpha, and invalid/overflowing dimensions; invalid inputs report chain errors rather than arithmetic-driven panics.

### M2 — Make watermark and format behavior match the API

**Status:** Complete · **Priority:** P1 · **Requirements:** WM-1–WM-5, FMT-1–FMT-4

- Implement `WithFontPath` and deterministic font-source option ordering.
- Validate nil/invalid watermark options and calculate all positions relative to image bounds.
- Register GIF decoding in the library; GIF encoding was added later in M7 with an explicit policy.
- Add tests for font files, option errors, anchor positions, non-zero bounds, and JPEG/PNG/GIF behavior.

**Exit criteria:** Every advertised option has observable behavior or is removed; format conversion and errors match the specification.

### M3 — Resolve clone and concurrency semantics

**Status:** Complete · **Priority:** P1 · **Requirements:** CON-1–CON-5

- Make Go docs, README, tests, and concurrency docs consistently describe `Clone` as shallow.
- Document that input and returned images are read-only during concurrent processing.
- Clarify that individual method calls are synchronized but a multi-call chain is not atomic.
- Replace or extend the current race test so readers and state-changing operations act on the same processor.

**Exit criteria:** The public contract is consistent across code and docs, and `go test -race ./...` covers the stated concurrency guarantee.

### M4 — Establish reliable quality gates and truthful performance docs

**Status:** Complete · **Priority:** P1 · **Requirements:** BLD-4, Q-1–Q-6

- Move timing comparisons out of regular tests and into benchmarks or a separately invoked tool.
- Decide whether the unused RGBA pool is worth implementing; remove it and its claims if measurements do not support it.
- Remove or substantiate SIMD and fixed speedup claims with reproducible benchmark data.
- Add regression coverage for the requirements in M1–M3.
- Run build, tests, race detection, vet, lint, and documentation build in CI; retain Go 1.27.0 as the declared minimum unless the policy is deliberately changed.

**Exit criteria:** Ordinary tests are correctness-focused and stable across machines; all release-gate checks in the specification pass.

### M5 — Add a standalone CLI

**Status:** Complete · **Priority:** P2 · **Requirements:** CLI-1–CLI-5

- Add a thin `cmd/gopiq` entry point and testable command logic under `internal/cli`.
- Implement `convert`, `resize`, `crop`, `grayscale`, and `watermark` commands with help and version output.
- Support file and standard input/output, format inference and `--format`, clear GIF-output errors, documented exit codes, and safe output writes.
- Add command integration tests, CLI documentation, Makefile build/install targets, and coverage reporting for the CLI package.
- Keep the CLI dependent only on the public root package; do not import `internal/engine` directly.

**Exit criteria:** The CLI builds and installs, every command and documented flag is covered by tests, format and exit-code behavior matches the specification, CLI documentation is published, and CI passes.

### M6 — Add streaming I/O and decode limits

**Status:** Complete · **Priority:** P2 · **Requirements:** EXT-1, EXT-2

- Design reader-based construction and writer-based encoding while keeping `FromBytes` and `ToBytes` compatible.
- Add configurable encoded-size, dimension, and total-pixel limits, validating before full decode where possible.
- Document defaults, limit errors, ownership, and resource behavior.

**Exit criteria:** Streaming callers avoid mandatory full encoded-buffer copies, and configured decode limits are enforced by tests including oversized and malformed inputs.

### M7 — Expand transformation and format capabilities

**Status:** Complete · **Priority:** P2 · **Requirements:** EXT-3–EXT-5

- Implement proportional resize fit/fill behavior, including rounding, crop, and upscaling rules.
- Implement flips and 90-degree rotations; evaluate image watermarks for a later milestone with a separate alpha/ownership contract.
- Implement single-frame GIF output with a documented standard-library palette policy; assess other formats independently.

**Exit criteria:** Each selected feature has a written API contract, pixel-level tests, and documentation before implementation is considered complete.

### M8 — Optimize measured workloads and support bounded batches

**Status:** Complete · **Priority:** P3 · **Requirements:** EXT-6–EXT-8

- Add `ProcessBatch` with bounded workers, ordered results, per-item errors, and context cancellation.
- Benchmark representative image sizes and processor counts; document the selected defaults and evaluate specialized paths.
- Keep pooling, fixed-point arithmetic, and SIMD out unless measurements justify them.
- Expand fuzz, platform, and Go-version coverage and add documentation-example checks.

**Exit criteria:** Optimizations show a repeatable benefit without changing specified output semantics, and batch/resource behavior is bounded and tested.

## Cross-milestone rules

- Fixes must include regression tests and keep error propagation consistent with the chain API.
- Do not claim an optimization or format is supported until implementation and tests demonstrate it.
- Avoid calendar commitments until scope and maintainer capacity are known.
- A release is gated by the acceptance criteria in the [specification](specification.md), not by completion of every future extension.
