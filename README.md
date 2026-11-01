# gopiq

[![Go](https://github.com/TamasGorgics/gopiq/actions/workflows/go.yml/badge.svg)](https://github.com/TamasGorgics/gopiq/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/TamasGorgics/gopiq/branch/main/graph/badge.svg)](https://codecov.io/gh/TamasGorgics/gopiq)
[![Go Version](https://img.shields.io/badge/Go-1.27-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/TamasGorgics/gopiq?status.svg)](https://godoc.org/github.com/TamasGorgics/gopiq)

A fluent, thread-safe Go image processing library with chainable operations.

## Features

- **Fluent Interface**: Chain multiple operations together naturally
- **Thread-Safe**: Individually synchronized methods; multi-call chains are not atomic
- **High-Quality Processing**: Uses Catmull-Rom interpolation for resizing
- **Comprehensive Error Handling**: Errors propagate through the chain
- **Multiple Format Support**: JPEG and PNG input/output; GIF input (decode-only)
- **Text Watermarks**: Add customizable text overlays with font control
- **Streaming I/O and Decode Limits**: Decode from readers, encode to writers, and bound untrusted input
- **Standalone CLI**: Use the same operations from the command line

The root `gopiq` package is the supported public API; implementation details live in `internal/engine`. See [Architecture](docs/architecture.md) for the package boundaries and planned CLI layout.

## Installation

```bash
go get github.com/TamasGorgics/gopiq
```

## CLI

Install the standalone command:

```bash
go install github.com/TamasGorgics/gopiq/cmd/gopiq@latest
```

```bash
gopiq resize --width 800 --height 600 input.jpg output.png
gopiq grayscale --fast input.jpg output.png
gopiq watermark --text "© 2026" --position bottom-right input.jpg output.png
```

See the [CLI documentation](docs/cli.md) for commands, flags, stdin/stdout use, and exit codes.

## Quick Start

```go
package main

import (
    "fmt"
    "os"

    "github.com/TamasGorgics/gopiq"
)

func main() {
    imageData, err := os.ReadFile("img.jpg")
    if err != nil {
        fmt.Printf("Error reading image: %v\n", err)
        return
    }

    // Load and process an image with method chaining
    processor := gopiq.FromBytes(imageData).
        Resize(800, 600).
        Grayscale().
        Crop(100, 100, 400, 300).
        AddTextWatermark("© 2025",
            gopiq.WithFontSize(24),
            gopiq.WithPosition(gopiq.PositionBottomRight))
    
    if processor.Err() != nil {
        fmt.Printf("Error: %v\n", processor.Err())
        return
    }
    
    // Export the result
    resultBytes, err := processor.ToBytes(gopiq.FormatJPEG)
    if err != nil {
        fmt.Printf("Export error: %v\n", err)
        return
    }
    
    err = os.WriteFile("output.jpg", resultBytes, 0644)
    if err != nil {
        fmt.Printf("Error writing output file: %v\n", err)
        return
    }

    fmt.Println("Image processed and saved successfully")
}
```

## Thread Safety

**gopiq** synchronizes individual method calls. `Image`, `Err`, `ToBytes`, and `Clone` take read locks; processing methods and `SetPerformanceOptions` take write locks.

A multi-call chain is not atomic. Another goroutine can interleave between calls in a chain, so use separate processors or external synchronization when an entire sequence must be atomic.

### Read-only images
`New` retains the image you pass in, and `Image()` returns the current image. Treat both as read-only while the processor or any clone may be used concurrently. Processing operations do not modify their source; for non-empty images they create a new destination. Mutating an image through a concrete-type assertion is outside the thread-safety guarantee.

### Clone shares image storage
`Clone` creates a shallow copy: error and performance-option state are copied, while the current image reference and pixel storage are shared. Clones have independent processor state and can process the shared source concurrently as long as the source image itself is not mutated. `Clone` does not deep-copy pixels; copy the image before constructing processors if independent pixel storage is required.

```go
original := gopiq.New(image)
var wg sync.WaitGroup

for range 10 {
    wg.Go(func() {
        processor := original.Clone().
            Resize(100, 100).
            AddTextWatermark("Processed")
        // processor starts from original's shared read-only image.
        if err := processor.Err(); err != nil {
            // Handle the processing error.
        }
    })
}
wg.Wait()
```

### Concurrent Processing Pattern
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

## API Reference

### Core Methods

- `New(img image.Image) *ImageProcessor` - Create a processor that retains `img` as read-only
- `FromBytes(data []byte) *ImageProcessor` - Decode image bytes with legacy unlimited behavior
- `FromBytesWithLimits(data []byte, limits DecodeLimits) *ImageProcessor` - Decode bytes with explicit limits
- `FromReader(r io.Reader) *ImageProcessor` - Decode from a reader using `DefaultDecodeLimits`
- `FromReaderWithLimits(r io.Reader, limits DecodeLimits) *ImageProcessor` - Decode from a reader with explicit limits
- `Clone() *ImageProcessor` - Create a shallow processor copy; image storage is shared
- `Image() (image.Image, error)` - Get the current image for reading
- `ToBytes(format ImageFormat) ([]byte, error)` - Export to bytes
- `ToWriter(w io.Writer, format ImageFormat) error` - Encode directly to a writer
- `Err() error` - Get any error from the processing chain

`FromBytes` decodes JPEG, PNG, and GIF. `FromReader` and `FromBytesWithLimits` support `DecodeLimits` and `ErrDecodeLimit`; see [Streaming and Decode Limits](docs/streaming.md). `ToBytes` and `ToWriter` encode JPEG and PNG; GIF encoding is not supported. `Clone` does not deep-copy pixels, and images passed to `New` or returned by `Image` must not be mutated while processors or clones are in use.

### Processing Operations

- `Resize(width, height int)` - Resize using Catmull-Rom interpolation
- `Crop(x, y, width, height int)` - Crop to specified rectangle
- `Grayscale()` - Convert to grayscale
- `AddTextWatermark(text, ...options)` - Add text watermark

### Watermark Options

- `WithFontSize(size float64)` - Set font size
- `WithColor(color color.Color)` - Set text color
- `WithPosition(pos WatermarkPosition)` - Set position
- `WithOffset(x, y float64)` - Set offset from position
- `WithFontPath(path string)` - Load a font file
- `WithFontBytes(data []byte)` - Use custom font

If both font-source options are provided, the last one wins. With neither option, the embedded Go font is used. Watermark positions are relative to the image bounds; text outside the image is clipped. Positive offsets inset corner anchors and move centered text right/down.

## Processing Notes

- For non-empty images, each operation creates a new destination image and leaves its source unchanged (memory-intensive)
- Use `Clone()` to branch processor state; clones share the read-only source image
- Consider image size limits for memory-constrained environments
- Catmull-Rom resizing provides high quality but is computationally intensive

## Performance

`GrayscaleFast()` uses parallel workers for non-empty images when parallel processing is enabled and the image meets the configured threshold. Performance depends on image size, CPU, workload, and options; no universal speedup is guaranteed.

See [Performance](docs/performance.md) for the benchmark environment, commands, and current measured results.

### Performance Configuration

```go
opts := gopiq.PerformanceOptions{
    MaxGoroutines:            8,
    EnableParallelProcessing: true,
    MinSizeForParallel:       10000,
}

processor := gopiq.NewWithPerformanceOptions(image, opts)
```

Use `GrayscaleFast()` for large images or workloads where the measured parallel path is beneficial. Use `Clone()` when processing multiple independent images in parallel.

## License

MIT License - see [LICENSE](LICENSE) file for details.
