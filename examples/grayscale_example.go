package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/TamasGorgics/gopiq"
)

func main() {
	fmt.Println("=== Grayscale Conversion Example ===")

	// Read the sample image
	imageData, err := os.ReadFile("sample_image.png")
	if err != nil {
		log.Fatalf("Failed to read sample image: %v", err)
	}

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

	err = os.WriteFile("grayscale_regular.png", grayscaleData, 0644)
	if err != nil {
		log.Fatalf("Failed to save grayscale image: %v", err)
	}
	fmt.Println("✓ Regular grayscale image saved as: grayscale_regular.png")

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

	err = os.WriteFile("grayscale_fast.png", fastGrayscaleData, 0644)
	if err != nil {
		log.Fatalf("Failed to save fast grayscale image: %v", err)
	}
	fmt.Println("✓ Fast grayscale image saved as: grayscale_fast.png")

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

	err = os.WriteFile("grayscale_custom_perf.png", customGrayscaleData, 0644)
	if err != nil {
		log.Fatalf("Failed to save custom performance grayscale image: %v", err)
	}
	fmt.Println("✓ Custom performance grayscale image saved as: grayscale_custom_perf.png")

	fmt.Println("\n=== Grayscale Example Complete ===")
	fmt.Println("Generated files:")
	fmt.Println("- grayscale_regular.png (standard grayscale conversion)")
	fmt.Println("- grayscale_fast.png (optimized parallel grayscale conversion)")
	fmt.Println("- grayscale_custom_perf.png (custom performance options)")
}
