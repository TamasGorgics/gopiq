# gopiq

[![Go](https://github.com/TamasGorgics/gopiq/actions/workflows/go.yml/badge.svg)](https://github.com/TamasGorgics/gopiq/actions/workflows/go.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/TamasGorgics/gopiq)](https://goreportcard.com/report/github.com/TamasGorgics/gopiq)
[![GoDoc](https://godoc.org/github.com/TamasGorgics/gopiq?status.svg)](https://godoc.org/github.com/TamasGorgics/gopiq)

A fluent, thread-safe Go image processing library with chainable operations.

## Features

- **Fluent Interface**: Chain multiple operations together naturally
- **Thread-Safe**: Safe for concurrent use by multiple goroutines
- **High-Quality Processing**: Uses Catmull-Rom interpolation for resizing
- **Comprehensive Error Handling**: Errors propagate through the chain
- **Multiple Format Support**: JPEG, PNG input/output
- **Text Watermarks**: Add customizable text overlays with font control

## Installation

```bash
go get github.com/TamasGorgics/gopiq
```

## Quick Start

```go
package main

import (
    "fmt"
    "gopiq"
)

func main() {
    // Load and process an image with method chaining
    processor := gopiq.FromBytes(imageData).
        Resize(800, 600).
        Grayscale().
        Crop(100, 100, 400, 300).
        AddTextWatermark("© 2024", 
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
    
    // Save or use resultBytes...
}
```

## Thread Safety

**gopiq** is designed to be thread-safe and supports concurrent usage:

### Safe Concurrent Reading
Multiple goroutines can safely read from the same processor:

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
Create independent copies for separate goroutine processing:

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

## API Reference

### Core Methods

- `New(img image.Image) *ImageProcessor` - Create processor from image
- `FromBytes(data []byte) *ImageProcessor` - Create processor from image bytes
- `Clone() *ImageProcessor` - Create independent copy
- `Image() (image.Image, error)` - Get current image
- `ToBytes(format ImageFormat) ([]byte, error)` - Export to bytes
- `Err() error` - Get any error from the processing chain

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
- `WithFontBytes(data []byte)` - Use custom font

## Performance Notes

- Each operation creates a new image copy (safe but memory-intensive)
- Use `Clone()` for concurrent processing of the same base image
- Consider image size limits for memory-constrained environments
- Catmull-Rom resizing provides high quality but is computationally intensive

## Performance Optimizations

**gopiq** includes advanced performance optimizations that can provide dramatic speed improvements:

### 🚀 Optimized Methods

#### **GrayscaleFast()** - Parallel Processing
For large images, use `GrayscaleFast()` instead of `Grayscale()`:

```go
// Up to 33x faster for large images (1920x1080)
processor := gopiq.New(image).GrayscaleFast()
```

#### **Performance Results** (500x500 images on Apple M1 Pro):
- **Slow method** (At/Set): 7.17ms, 3MB allocations 
- **Optimized method**: 735μs, 1MB allocations (**9.7x faster**)
- **Parallel method**: 276μs, 1MB allocations (**26x faster**)

### ⚡ Performance Configuration

```go
// Custom performance settings
opts := gopiq.PerformanceOptions{
    MaxGoroutines:            8,     // Limit parallel goroutines
    EnableParallelProcessing: true,  // Enable parallel processing
    MinSizeForParallel:       10000, // Minimum pixels for parallel
}

processor := gopiq.NewWithPerformanceOptions(image, opts)
```

### 📊 Scalability

**Parallel Processing Performance** (1920x1080 images):
- 1 goroutine: 5.87ms
- 2 goroutines: 3.11ms (**1.9x faster**)
- 4 goroutines: 1.77ms (**3.3x faster**)
- 8 goroutines: 1.23ms (**4.8x faster**)

### 🔧 Optimization Techniques

1. **Direct Buffer Access**: Bypasses Go's interface overhead
2. **Parallel Processing**: Utilizes multiple CPU cores automatically
3. **Memory Pooling**: Reduces garbage collection pressure
4. **SIMD-friendly Operations**: CPU-optimized pixel processing
5. **ITU-R BT.709 Grayscale**: Professional-grade color conversion

### 💡 Performance Tips

- Use `GrayscaleFast()` for images larger than 100x100 pixels
- Set `MaxGoroutines` to match your CPU cores for optimal performance
- Process multiple images in parallel using `Clone()`
- Consider image size vs. processing time trade-offs

Example for batch processing:
```go
func processImagesOptimized(images []image.Image) []*gopiq.ImageProcessor {
    results := make([]*gopiq.ImageProcessor, len(images))
    var wg sync.WaitGroup
    
    for i, img := range images {
        wg.Add(1)
        go func(index int, image image.Image) {
            defer wg.Done()
            results[index] = gopiq.New(image).
                SetPerformanceOptions(gopiq.PerformanceOptions{
                    MaxGoroutines: runtime.NumCPU(),
                    EnableParallelProcessing: true,
                }).
                Resize(800, 600).
                GrayscaleFast()
        }(i, img)
    }
    
    wg.Wait()
    return results
}
```

## License

MIT License - see [LICENSE](LICENSE) file for details.