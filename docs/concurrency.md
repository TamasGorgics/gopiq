# Concurrency

**gopiq** is designed to be thread-safe and supports concurrent usage out of the box.

### Safe Concurrent Reading
Multiple goroutines can safely read from the same processor instance. The library uses `sync.RWMutex` to ensure that read operations do not conflict with write operations.

```go
proc := gopiq.New(image).Resize(800, 600).Grayscale()

// Multiple goroutines can safely call these concurrently:
go func() {
    img, err := proc.Image()
    // Use img...
}()

go func() {
    bytes, err := proc.ToBytes(gopiq.FormatPNG)
    // Use bytes...
}()
```

### Clone for Independent Processing
For parallel *write* operations, you should create independent copies of the processor. The `Clone()` method creates a shallow copy of the processor that is safe to use in a separate goroutine.

```go
original := gopiq.New(image)

// Each goroutine gets its own independent processor
for i := 0; i < 10; i++ {
    go func() {
        processor := original.Clone().
            Resize(100, 100).
            AddTextWatermark("Processed")
        // Process independently...
    }()
}
```

### Concurrent Processing Pattern
Here is a common pattern for processing multiple images in parallel.

```go
func processImagesInParallel(images []image.Image) []*gopiq.ImageProcessor {
    results := make([]*gopiq.ImageProcessor, len(images))
    var wg sync.WaitGroup
    
    for i, img := range images {
        wg.Add(1)
        go func(index int, image image.Image) {
            defer wg.Done()
            results[index] = gopiq.New(image).
                Resize(200, 200).
                Grayscale()
        }(i, img)
    }
    
    wg.Wait()
    return results
}
``` 