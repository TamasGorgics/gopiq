# Architecture

The module has one public package, an implementation engine, and an internal CLI layer:

```text
github.com/TamasGorgics/gopiq                 # public API
github.com/TamasGorgics/gopiq/internal/engine # image implementation
github.com/TamasGorgics/gopiq/internal/cli    # CLI command logic
github.com/TamasGorgics/gopiq/cmd/gopiq       # CLI entry point
```

The root package owns the exported API: `ImageProcessor`, formats, performance options, and watermark options. The engine owns image state, synchronization, decoding/encoding, and image processing. Engine types are intentionally not used in public function signatures. The CLI layer translates command-line input into calls to the public root API.

## Dependency direction

- The root package may import `internal/engine`.
- `internal/engine` must not import the root package.
- Examples import only the root package.
- `cmd/gopiq` imports `internal/cli`; `internal/cli` imports only the root package, not `internal/engine`.

The `internal` directory prevents external modules from importing the engine. Packages inside this module can still import it, so the dependency rule is enforced by review and can later be checked with a linter.

## Tests

- Root-package tests exercise behavior through the facade, and `api_test.go` compiles against the public package as an external consumer.
- `internal/engine` tests cover implementation details such as checked arithmetic, worker counts, fixed-point conversion, and decoding/encoding helpers.
- `internal/cli` tests cover commands, flags, format inference, standard streams, failures, and output writes.
- The coverage command includes the root package, the engine, and the CLI layer.

## CLI

The standalone command is split into a thin entry point and testable command logic:

```text
cmd/gopiq/main.go       # os.Exit(cli.Main(...))
internal/cli/            # command parsing, I/O, and execution
```

`internal/cli` uses `FromBytes`, the chainable operations, and `ToBytes` from the public root package. It must not import `internal/engine` directly. This keeps the CLI a consumer of the same API that external applications use.
