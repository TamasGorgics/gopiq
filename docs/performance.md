# Performance

**gopiq** includes advanced performance optimizations that can provide dramatic speed improvements.

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