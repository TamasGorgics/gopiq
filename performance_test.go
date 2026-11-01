package gopiq

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"runtime"
	"testing"
)

// createLargeTestImage creates a large test image for performance testing
func createLargeTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a realistic pattern that exercises all RGB channels
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x*256/width + y*128/height) % 256)
			g := uint8((y*256/height + x*64/width) % 256)
			b := uint8(((x+y)*256/(width+height) + 128) % 256)
			img.Set(x, y, color.RGBA{r, g, b, 255})
		}
	}
	return img
}

// Benchmark the optimized grayscale implementation
func benchmarkGrayscaleOptimized(b *testing.B, width, height int) {
	b.Helper()
	img := createLargeTestImage(width, height)

	b.ReportAllocs()

	for b.Loop() {
		proc := New(img)
		proc.Grayscale()
		if proc.Err() != nil {
			b.Fatal(proc.Err())
		}
	}
}

// Benchmark the parallel grayscale implementation
func benchmarkGrayscaleFast(b *testing.B, width, height int) {
	b.Helper()
	img := createLargeTestImage(width, height)

	b.ReportAllocs()

	for b.Loop() {
		proc := New(img)
		proc.GrayscaleFast()
		if proc.Err() != nil {
			b.Fatal(proc.Err())
		}
	}
}

// Small image benchmarks (100x100)
func BenchmarkGrayscaleSmall_Optimized(b *testing.B) {
	benchmarkGrayscaleOptimized(b, 100, 100)
}

func BenchmarkGrayscaleSmall_Fast(b *testing.B) {
	benchmarkGrayscaleFast(b, 100, 100)
}

// Medium image benchmarks (500x500)
func BenchmarkGrayscaleMedium_Optimized(b *testing.B) {
	benchmarkGrayscaleOptimized(b, 500, 500)
}

func BenchmarkGrayscaleMedium_Fast(b *testing.B) {
	benchmarkGrayscaleFast(b, 500, 500)
}

// Large image benchmarks (1920x1080)
func BenchmarkGrayscaleLarge_Optimized(b *testing.B) {
	benchmarkGrayscaleOptimized(b, 1920, 1080)
}

func BenchmarkGrayscaleLarge_Fast(b *testing.B) {
	benchmarkGrayscaleFast(b, 1920, 1080)
}

// Very large image benchmarks (4K: 3840x2160)
func BenchmarkGrayscaleXL_Optimized(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark in short mode")
	}
	benchmarkGrayscaleOptimized(b, 3840, 2160)
}

func BenchmarkGrayscaleXL_Fast(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping large benchmark in short mode")
	}
	benchmarkGrayscaleFast(b, 3840, 2160)
}

// Benchmark grayscale paths around the default parallel threshold.
func BenchmarkGrayscaleThreshold(b *testing.B) {
	for _, size := range []int{32, 64, 96, 100, 128, 256} {
		b.Run(fmt.Sprintf("%dx%d/optimized", size, size), func(b *testing.B) {
			benchmarkGrayscaleOptimized(b, size, size)
		})
		b.Run(fmt.Sprintf("%dx%d/fast", size, size), func(b *testing.B) {
			benchmarkGrayscaleFast(b, size, size)
		})
	}
}

// Benchmark bounded batch processing at several worker counts.
func BenchmarkProcessBatch(b *testing.B) {
	images := make([]image.Image, 8)
	for i := range images {
		images[i] = createLargeTestImage(500, 500)
	}

	for _, workers := range []int{1, 2, 4, 8} {
		b.Run(fmt.Sprintf("workers_%d", workers), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				_, err := ProcessBatch(context.Background(), images, func(processor *ImageProcessor) *ImageProcessor {
					processor.Grayscale()
					return processor
				}, BatchOptions{MaxWorkers: workers})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark different numbers of goroutines for parallel processing
func BenchmarkGrayscaleParallelGoroutines(b *testing.B) {
	img := createLargeTestImage(1920, 1080)

	for _, numGoroutines := range []int{1, 2, 4, 8, 16, runtime.NumCPU()} {
		b.Run(fmt.Sprintf("goroutines_%d", numGoroutines), func(b *testing.B) {
			opts := DefaultPerformanceOptions()
			opts.MaxGoroutines = numGoroutines

			b.ReportAllocs()

			for b.Loop() {
				proc := NewWithPerformanceOptions(img, opts)
				proc.GrayscaleFast()
				if proc.Err() != nil {
					b.Fatal(proc.Err())
				}
			}
		})
	}
}

// Benchmark allocation patterns for common processing pipelines.
func BenchmarkPipelineAllocations(b *testing.B) {
	img := createLargeTestImage(800, 600)

	b.ReportAllocs()
	for b.Loop() {
		proc1 := New(img)
		proc2 := proc1.Clone()
		proc3 := proc2.Clone()

		proc1.Grayscale()
		proc2.GrayscaleFast()
		proc3.Resize(400, 300)
	}
}

// Benchmark concurrent processing under realistic load.
func BenchmarkConcurrentProcessing(b *testing.B) {
	img := createLargeTestImage(1000, 1000)

	b.Run("sequential", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			proc := New(img)
			proc.Resize(500, 500).GrayscaleFast().Crop(100, 100, 300, 300)
		}
	})

	b.Run("parallel", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				proc := New(img)
				proc.Resize(500, 500).GrayscaleFast().Crop(100, 100, 300, 300)
			}
		})
	})
}
