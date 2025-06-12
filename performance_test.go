package gopiq

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"testing"
	"time"
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

// Benchmark the original slow grayscale implementation for comparison
func benchmarkGrayscaleSlow(b *testing.B, width, height int) {
	img := createLargeTestImage(width, height)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		proc := New(img)
		// Simulate the old slow method using At/Set
		proc.grayscaleSlowMethod()
		if proc.Err() != nil {
			b.Fatal(proc.Err())
		}
	}
}

// grayscaleSlowMethod simulates the original slow implementation
func (ip *ImageProcessor) grayscaleSlowMethod() *ImageProcessor {
	if ip.err != nil {
		return ip
	}

	bounds := ip.currentImage.Bounds()
	grayImg := image.NewRGBA(bounds)

	// Original slow method using At/Set interface calls
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalColor := ip.currentImage.At(x, y)
			grayColor := color.GrayModel.Convert(originalColor).(color.Gray)
			grayImg.Set(x, y, grayColor)
		}
	}

	ip.currentImage = grayImg
	return ip
}

// Benchmark the optimized grayscale implementation
func benchmarkGrayscaleOptimized(b *testing.B, width, height int) {
	img := createLargeTestImage(width, height)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		proc := New(img)
		proc.Grayscale()
		if proc.Err() != nil {
			b.Fatal(proc.Err())
		}
	}
}

// Benchmark the parallel grayscale implementation
func benchmarkGrayscaleFast(b *testing.B, width, height int) {
	img := createLargeTestImage(width, height)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		proc := New(img)
		proc.GrayscaleFast()
		if proc.Err() != nil {
			b.Fatal(proc.Err())
		}
	}
}

// Small image benchmarks (100x100)
func BenchmarkGrayscaleSmall_Slow(b *testing.B) {
	benchmarkGrayscaleSlow(b, 100, 100)
}

func BenchmarkGrayscaleSmall_Optimized(b *testing.B) {
	benchmarkGrayscaleOptimized(b, 100, 100)
}

func BenchmarkGrayscaleSmall_Fast(b *testing.B) {
	benchmarkGrayscaleFast(b, 100, 100)
}

// Medium image benchmarks (500x500)
func BenchmarkGrayscaleMedium_Slow(b *testing.B) {
	benchmarkGrayscaleSlow(b, 500, 500)
}

func BenchmarkGrayscaleMedium_Optimized(b *testing.B) {
	benchmarkGrayscaleOptimized(b, 500, 500)
}

func BenchmarkGrayscaleMedium_Fast(b *testing.B) {
	benchmarkGrayscaleFast(b, 500, 500)
}

// Large image benchmarks (1920x1080)
func BenchmarkGrayscaleLarge_Slow(b *testing.B) {
	benchmarkGrayscaleSlow(b, 1920, 1080)
}

func BenchmarkGrayscaleLarge_Optimized(b *testing.B) {
	benchmarkGrayscaleOptimized(b, 1920, 1080)
}

func BenchmarkGrayscaleLarge_Fast(b *testing.B) {
	benchmarkGrayscaleFast(b, 1920, 1080)
}

// Very large image benchmarks (4K: 3840x2160)
func BenchmarkGrayscaleXL_Slow(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping slow benchmark in short mode")
	}
	benchmarkGrayscaleSlow(b, 3840, 2160)
}

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

// Benchmark different numbers of goroutines for parallel processing
func BenchmarkGrayscaleParallelGoroutines(b *testing.B) {
	img := createLargeTestImage(1920, 1080)

	for _, numGoroutines := range []int{1, 2, 4, 8, 16, runtime.NumCPU()} {
		b.Run(fmt.Sprintf("goroutines_%d", numGoroutines), func(b *testing.B) {
			opts := DefaultPerformanceOptions()
			opts.MaxGoroutines = numGoroutines

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				proc := NewWithPerformanceOptions(img, opts)
				proc.GrayscaleFast()
				if proc.Err() != nil {
					b.Fatal(proc.Err())
				}
			}
		})
	}
}

// Benchmark memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	img := createLargeTestImage(800, 600)

	b.Run("without_pool", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			// Create multiple processors to test allocation patterns
			proc1 := New(img)
			proc2 := proc1.Clone()
			proc3 := proc2.Clone()

			proc1.Grayscale()
			proc2.GrayscaleFast()
			proc3.Resize(400, 300)
		}
	})
}

// Test concurrent performance under realistic load
func BenchmarkConcurrentProcessing(b *testing.B) {
	img := createLargeTestImage(1000, 1000)

	b.Run("sequential", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			proc := New(img)
			proc.Resize(500, 500).GrayscaleFast().Crop(100, 100, 300, 300)
		}
	})

	b.Run("parallel_4", func(b *testing.B) {
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				proc := New(img)
				proc.Resize(500, 500).GrayscaleFast().Crop(100, 100, 300, 300)
			}
		})
	})
}

// Performance test that prints detailed timing information
func TestPerformanceComparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance comparison in short mode")
	}

	t.Log("Performance Comparison Test")
	t.Log("============================")

	sizes := []struct {
		name   string
		width  int
		height int
	}{
		{"Small (400x300)", 400, 300},
		{"Medium (800x600)", 800, 600},
		{"Large (1920x1080)", 1920, 1080},
	}

	for _, size := range sizes {
		t.Logf("\n%s Images:", size.name)
		img := createLargeTestImage(size.width, size.height)

		// Test slow method
		start := time.Now()
		proc1 := New(img)
		proc1.grayscaleSlowMethod()
		slowTime := time.Since(start)

		// Test optimized method
		start = time.Now()
		proc2 := New(img)
		proc2.Grayscale()
		optimizedTime := time.Since(start)

		// Test parallel method
		start = time.Now()
		proc3 := New(img)
		proc3.GrayscaleFast()
		fastTime := time.Since(start)

		speedupOptimized := float64(slowTime) / float64(optimizedTime)
		speedupFast := float64(slowTime) / float64(fastTime)

		t.Logf("  Slow method:      %v", slowTime)
		t.Logf("  Optimized method: %v (%.1fx faster)", optimizedTime, speedupOptimized)
		t.Logf("  Parallel method:  %v (%.1fx faster)", fastTime, speedupFast)
	}
}
