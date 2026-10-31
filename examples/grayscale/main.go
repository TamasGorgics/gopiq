package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/TamasGorgics/gopiq"
)

func main() {
	fmt.Println("=== Grayscale Conversion Example ===")

	// Read the sample image
	imageData, err := os.ReadFile("examples/sample_image.png")
	if err != nil {
		log.Fatalf("Failed to read sample image: %v", err)
	}
	outputDir := filepath.Join("examples", "grayscale", "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
	outputPath := func(name string) string { return filepath.Join(outputDir, name) }

	// Test regular grayscale conversion
	fmt.Println("Testing regular grayscale conversion...")
	start := time.Now()

	processor1 := gopiq.FromBytes(imageData)
	if err := processor1.Err(); err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}

	processor1.Grayscale()
	if err := processor1.Err(); err != nil {
		log.Fatalf("Failed to convert to grayscale: %v", err)
	}

	regularTime := time.Since(start)
	fmt.Printf("Regular grayscale conversion took: %v\n", regularTime)

	// Save regular grayscale version
	grayscaleData, err := processor1.ToBytes(gopiq.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to convert to bytes: %v", err)
	}

	err = os.WriteFile(outputPath("grayscale_regular.png"), grayscaleData, 0644)
	if err != nil {
		log.Fatalf("Failed to save grayscale image: %v", err)
	}
	fmt.Printf("✓ Regular grayscale image saved as: %s\n", outputPath("grayscale_regular.png"))

	// Test fast grayscale conversion
	fmt.Println("\nTesting fast grayscale conversion...")
	start = time.Now()

	processor2 := gopiq.FromBytes(imageData)
	if err := processor2.Err(); err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}

	processor2.GrayscaleFast()
	if err := processor2.Err(); err != nil {
		log.Fatalf("Failed to convert to grayscale (fast): %v", err)
	}

	fastTime := time.Since(start)
	fmt.Printf("Fast grayscale conversion took: %v\n", fastTime)

	// Save fast grayscale version
	fastGrayscaleData, err := processor2.ToBytes(gopiq.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to convert to bytes: %v", err)
	}

	err = os.WriteFile(outputPath("grayscale_fast.png"), fastGrayscaleData, 0644)
	if err != nil {
		log.Fatalf("Failed to save fast grayscale image: %v", err)
	}
	fmt.Printf("✓ Fast grayscale image saved as: %s\n", outputPath("grayscale_fast.png"))

	// Performance comparison
	fmt.Printf("\nPerformance comparison:\n")
	fmt.Printf("Regular: %v\n", regularTime)
	fmt.Printf("Fast:    %v\n", fastTime)
	if fastTime < regularTime {
		speedup := float64(regularTime) / float64(fastTime)
		fmt.Printf("Fast method is %.2fx faster\n", speedup)
	} else {
		fmt.Println("Regular method was faster (image might be too small for parallel processing)")
	}

	// Test with performance options
	fmt.Println("\nTesting with custom performance options...")

	// Create processor with custom performance options
	perfOpts := gopiq.PerformanceOptions{
		MaxGoroutines:            4,
		EnableParallelProcessing: true,
		MinSizeForParallel:       5000, // Lower threshold for this example
	}

	// Get the image from processor1 and create a new processor with custom options
	img, err := processor1.Image()
	if err != nil {
		log.Fatalf("Failed to get image: %v", err)
	}
	processor3 := gopiq.NewWithPerformanceOptions(img, perfOpts)
	start = time.Now()

	processor3.GrayscaleFast()
	if err := processor3.Err(); err != nil {
		log.Fatalf("Failed to convert to grayscale with custom options: %v", err)
	}

	customTime := time.Since(start)
	fmt.Printf("Custom performance options grayscale took: %v\n", customTime)

	// Save custom performance version
	customGrayscaleData, err := processor3.ToBytes(gopiq.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to convert to bytes: %v", err)
	}

	err = os.WriteFile(outputPath("grayscale_custom_perf.png"), customGrayscaleData, 0644)
	if err != nil {
		log.Fatalf("Failed to save custom performance grayscale image: %v", err)
	}
	fmt.Printf("✓ Custom performance grayscale image saved as: %s\n", outputPath("grayscale_custom_perf.png"))

	fmt.Println("\n=== Grayscale Example Complete ===")
	fmt.Println("Generated files:")
	fmt.Printf("- %s (standard grayscale conversion)\n", outputPath("grayscale_regular.png"))
	fmt.Printf("- %s (optimized parallel grayscale conversion)\n", outputPath("grayscale_fast.png"))
	fmt.Printf("- %s (custom performance options)\n", outputPath("grayscale_custom_perf.png"))
}
