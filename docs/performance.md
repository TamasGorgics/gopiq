# Performance

Performance depends on image size, CPU, workload, and the selected processing options. No universal speedup is guaranteed; measure your own workload.

## Benchmark environment

Results below were measured on:

- **Date:** 2026-11-01
- **Machine:** Apple M1 Pro, 10 logical CPUs
- **OS/architecture:** darwin/arm64
- **Go:** 1.27.1

Command:

```bash
go test -run '^$' \
  -bench 'BenchmarkGrayscale(Small|Medium|Large)_(Optimized|Fast)$|BenchmarkGrayscaleThreshold$|BenchmarkGrayscaleParallelGoroutines$|BenchmarkPipelineAllocations$|BenchmarkConcurrentProcessing$|BenchmarkProcessBatch$' \
  -benchmem -count=5 .
```

The tables report the median of five runs. Source images are deterministic RGB patterns generated in `performance_test.go`; each benchmark creates the processor inside the measured loop while source-image construction is outside the timer. `GrayscaleFast()` used the default performance options: parallel processing enabled, `MaxGoroutines = runtime.NumCPU()` (10), and a 10,000-pixel threshold.

## Grayscale

| Image | Method | Time/op | Memory/op | Allocs/op | Relative to `Grayscale` |
|---|---:|---:|---:|---:|---:|
| 100×100 | `Grayscale` | 28.8 µs | 41,128 B | 4 | 1.00× |
| 100×100 | `GrayscaleFast` | 26.3 µs | 41,864 B | 25 | 1.10× |
| 500×500 | `Grayscale` | 696.7 µs | 1,007,790 B | 4 | 1.00× |
| 500×500 | `GrayscaleFast` | 308.6 µs | 1,008,530 B | 25 | 2.26× |
| 1920×1080 | `Grayscale` | 5.44 ms | 8,298,669 B | 4 | 1.00× |
| 1920×1080 | `GrayscaleFast` | 1.27 ms | 8,299,404 B | 25 | 4.29× |

`GrayscaleFast` has additional goroutine and scheduling overhead for small images. The parallel path becomes more useful as image size increases, but the crossover point depends on the machine and workload.

## Parallel worker count

1920×1080 image, using `GrayscaleFast`:

| `MaxGoroutines` | Time/op | Memory/op | Allocs/op |
|---:|---:|---:|---:|
| 1 | 5.27 ms | 8,298,752 B | 7 |
| 2 | 2.91 ms | 8,298,827 B | 9 |
| 4 | 1.66 ms | 8,298,971 B | 13 |
| 8 | 1.23 ms | 8,299,262 B | 21 |
| 10 (default) | 1.34 ms | 8,299,401 B | 25 |
| 16 | 1.18 ms | 8,299,835 B | 37 |

Using more workers than available CPUs can still change timing, but it also increases allocation and scheduling overhead. Choose a worker count from measurements for the target environment.

## Threshold and batch processing

Grayscale crossover measurements:

| Image | `Grayscale` | `GrayscaleFast` | Relative to `Grayscale` |
|---|---:|---:|---:|
| 32×32 | 3.14 µs | 3.16 µs | 0.99× |
| 64×64 | 11.7 µs | 11.8 µs | 0.99× |
| 96×96 | 26.2 µs | 25.9 µs | 1.01× |
| 100×100 (10,000 pixels) | 29.6 µs | 26.9 µs | 1.10× |
| 128×128 | 47.1 µs | 36.4 µs | 1.30× |
| 256×256 | 185.8 µs | 102.4 µs | 1.81× |

The default 10,000-pixel threshold was retained after this measurement. Below the crossover, parallel scheduling overhead can outweigh the benefit.

`ProcessBatch` with eight 500×500 grayscale operations:

| Workers | Time/op | Memory/op | Allocs/op |
|---:|---:|---:|---:|
| 1 | 5.25 ms | 8,062,498 B | 38 |
| 2 | 2.80 ms | 8,062,587 B | 40 |
| 4 | 1.68 ms | 8,062,739 B | 44 |
| 8 | 1.48 ms | 8,063,024 B | 52 |

The batch worker count is bounded by the caller and by the number of inputs. These results are specific to the recorded machine and workload.

## Pipelines and concurrent processing

| Benchmark | Time/op | Memory/op | Allocs/op |
|---|---:|---:|---:|
| `BenchmarkPipelineAllocations` | 9.13 ms | 12,151,675 B | 39 |
| `BenchmarkConcurrentProcessing/sequential` | 16.69 ms | 18,556,075 B | 35 |
| `BenchmarkConcurrentProcessing/parallel` | 2.43 ms | 18,556,075 B | 35 |

`BenchmarkPipelineAllocations` measures creating two clones and running grayscale, fast grayscale, and resize on an 800×600 source. `BenchmarkConcurrentProcessing` measures resize, fast grayscale, and crop on a 1000×1000 source, both sequentially and through parallel `b.RunParallel` execution.

## Configuration

```go
opts := gopiq.PerformanceOptions{
    MaxGoroutines:            8,
    EnableParallelProcessing: true,
    MinSizeForParallel:       10000,
}

processor := gopiq.NewWithPerformanceOptions(image, opts)
```

The default options enable parallel processing at 10,000 pixels and use `runtime.NumCPU()` goroutines. A `MaxGoroutines` value less than or equal to zero uses that runtime default; a `MinSizeForParallel` value less than or equal to zero allows parallel processing for any non-empty image. `GrayscaleFast()` leaves empty images unchanged, and disabling parallel processing always selects the sequential path.

## Implementation facts

- Grayscale conversion uses direct RGBA buffer access and the BT.709 luma calculation.
- `GrayscaleFast()` uses parallel row workers when enabled and above the configured threshold.
- The default 10,000-pixel threshold was evaluated and retained.
- Specialized pooling, fixed-point, and SIMD paths were evaluated; none were selected for this release, and this page makes no claims for them.
