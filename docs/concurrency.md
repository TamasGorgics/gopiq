# Concurrency

**gopiq** synchronizes individual method calls. `Image`, `Err`, `ToBytes`, and `Clone` take read locks; processing methods and `SetPerformanceOptions` take write locks. The processor state and its image reference can therefore be accessed safely from multiple goroutines.

## Chains are not atomic

A multi-call chain is not an atomic transaction. Another goroutine may interleave between calls in a chain, so the image seen by the next call is not guaranteed to be the result of the previous call in your goroutine. Use separate processors or an external mutex when an entire sequence must be atomic.

```go
proc := gopiq.New(image)

// Another goroutine may run between these calls; this is not one atomic step.
proc.Resize(800, 600).Crop(10, 10, 200, 200).Grayscale()
```

## Read-only images

`New` retains the image passed to it. `Image()` returns the processor's current image. Treat both as read-only while the processor or any clone may be used concurrently.

Processing operations do not modify their source in place; for non-empty images they create a new destination image. This is what allows clones to process a shared source independently. Mutating an image through a concrete-type assertion is outside the thread-safety guarantee.

```go
source := image.NewRGBA(image.Rect(0, 0, 800, 600))
proc := gopiq.New(source)

// Do not call source.Set while proc or its clones are in use.
_, err := proc.Resize(400, 300).Image()
if err != nil {
    // Handle the error.
}
// Any image obtained from Image is read-only too.
```

## Clone is shallow

`Clone` creates a shallow copy of the processor. It copies the current error and performance options, but shares the current image reference and its pixel storage with the original. Clones have independent processor state and can be used from separate goroutines as long as the shared source image is not mutated.

`Clone` does not deep-copy pixels. If you need independent pixel storage, copy the image before constructing processors.

```go
original := gopiq.New(image)
var wg sync.WaitGroup

for range 10 {
    wg.Go(func() {
        processor := original.Clone().
            Resize(100, 100).
            AddTextWatermark("Processed")
        // processor has independent state; its starting image is the shared source.
        if processor.Err() != nil {
            // Handle the processing error.
        }
    })
}
wg.Wait()
```

## Concurrent processing pattern

For independent images, create a processor per image. If several processors branch from one base image, use shallow clones and keep the shared source read-only.

```go
func processImagesInParallel(images []image.Image) []*gopiq.ImageProcessor {
    results := make([]*gopiq.ImageProcessor, len(images))
    var wg sync.WaitGroup

    for index, img := range images {
        wg.Go(func() {
            results[index] = gopiq.New(img).
                Resize(200, 200).
                Grayscale()
        })
    }

    wg.Wait()
    return results
}
```
