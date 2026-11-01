# Home

A fluent, thread-safe Go image processing library with chainable operations.

[![Go](https://github.com/TamasGorgics/gopiq/actions/workflows/go.yml/badge.svg)](https://github.com/TamasGorgics/gopiq/actions/workflows/go.yml)
[![codecov](https://codecov.io/gh/TamasGorgics/gopiq/branch/main/graph/badge.svg)](https://codecov.io/gh/TamasGorgics/gopiq)
[![Go Version](https://img.shields.io/badge/Go-1.27-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![GoDoc](https://godoc.org/github.com/TamasGorgics/gopiq?status.svg)](https://godoc.org/github.com/TamasGorgics/gopiq)

## Features

- **Fluent Interface**: Chain multiple operations together naturally
- **Thread-Safe**: Individually synchronized methods; multi-call chains are not atomic
- **High-Quality Processing**: Catmull-Rom resizing plus exact, fit, and fill modes
- **Transformations**: Cropping, horizontal/vertical flips, and 90/180/270-degree rotation
- **Comprehensive Error Handling**: Errors propagate through the chain
- **Multiple Format Support**: JPEG, PNG, and GIF input/output; GIF output uses a documented lossy palette policy
- **Text Watermarks**: Add customizable text overlays with font control
- **Streaming I/O and Decode Limits**: Decode from readers, encode to writers, and bound untrusted input
- **Bounded Batch Processing**: Process multiple images with ordered results and cancellation
- **Standalone CLI**: Run common operations without writing Go code

## Installation

```bash
go get github.com/TamasGorgics/gopiq
```

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
