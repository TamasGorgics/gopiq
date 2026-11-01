# Bounded Batch Processing

`ProcessBatch` processes a slice of images with a bounded number of workers while preserving input order.

```go
type BatchOptions struct {
    MaxWorkers int
}

func ProcessBatch(
    ctx context.Context,
    images []image.Image,
    operation func(*ImageProcessor) *ImageProcessor,
    options BatchOptions,
) ([]*ImageProcessor, error)
```

## Example

```go
images := []image.Image{first, second, third}

results, err := gopiq.ProcessBatch(
    context.Background(),
    images,
    func(processor *gopiq.ImageProcessor) *gopiq.ImageProcessor {
        return processor.Resize(800, 600).Grayscale()
    },
    gopiq.BatchOptions{MaxWorkers: 4},
)
if err != nil {
    return err
}

for index, processor := range results {
    if processor == nil {
        continue
    }
    if err := processor.Err(); err != nil {
        log.Printf("image %d failed: %v", index, err)
    }
}
```

## Semantics

- Results are indexed by input position and do not depend on completion order.
- Each input gets a separate `ImageProcessor` created with `New`.
- Input images are read-only; operations should create their own destination images using the normal API.
- The `operation` callback can run concurrently. If it accesses shared caller state, that state must be synchronized. It should return the processor it received; a nil return leaves a nil result slot. Callback panics are the caller’s responsibility.
- Per-image processing errors stay on the corresponding result’s `Err()` and do not stop unrelated items.
- `MaxWorkers <= 0` uses `runtime.NumCPU()`.
- The worker count is capped by `len(images)`.
- An empty input returns an empty result slice and no error.

## Cancellation

`ProcessBatch` checks the context before starting each new item. Cancellation does not interrupt an operation that has already started. When the context is canceled:

- no new items are scheduled;
- in-flight operations are allowed to finish;
- the returned results contain completed items;
- unstarted result slots remain `nil`;
- the returned error is `ctx.Err()`.

Use a per-operation timeout or an operation that observes cancellation itself when long-running work must be interrupted.

## Errors

A nil context or nil operation returns an error immediately. Processing errors are per result:

```go
if results[0] != nil && results[0].Err() != nil {
    // Handle the first image's error.
}
```

A nil result means that item was not processed, typically because the context was canceled before its work started.
