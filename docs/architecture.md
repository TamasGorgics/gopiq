# Architecture

The module has one public package and one implementation package:

```text
github.com/TamasGorgics/gopiq          # public API
github.com/TamasGorgics/gopiq/internal/engine
```

The root package owns the exported API: `ImageProcessor`, formats, performance options, and watermark options. The engine owns image state, synchronization, decoding/encoding, and image processing. Engine types are intentionally not used in public function signatures.

## Dependency direction

- The root package may import `internal/engine`.
- `internal/engine` must not import the root package.
- Examples import only the root package.
- A future `cmd/gopiq` CLI must import only the root package. It may use `internal/cli` for command logic, but not `internal/engine`.

The `internal` directory prevents external modules from importing the engine. Packages inside this module can still import it, so the dependency rule is enforced by review and can later be checked with a linter.

## Tests

- Root-package tests exercise behavior through the facade, and `api_test.go` compiles against the public package as an external consumer.
- `internal/engine` tests cover implementation details such as checked arithmetic, worker counts, fixed-point conversion, and decoding/encoding helpers.
- The coverage command includes both the root package and the engine.

## Planned CLI

The CLI will live under `cmd/gopiq`, with testable command logic under `internal/cli`. It will use `FromBytes`, the chainable operations, and `ToBytes` from the public package. This keeps the CLI a consumer of the same API that external applications use.
