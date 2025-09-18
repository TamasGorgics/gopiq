package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"time"

	"github.com/TamasGorgics/gopiq"
)

func main() {
	fmt.Println("=== Comprehensive Gopiq Example ===")
	fmt.Println("This example demonstrates all major features of gopiq")

	// Read the sample image
	imageData, err := os.ReadFile("sample_image.png")
	if err != nil {
		log.Fatalf("Failed to read sample image: %v", err)
	}

	// Create processor with custom performance options
	perfOpts := gopiq.PerformanceOptions{
		MaxGoroutines:            4,
		EnableParallelProcessing: true,
		MinSizeForParallel:       5000,
	}

	processor := gopiq.NewWithPerformanceOptions(nil, perfOpts)
	processor = gopiq.FromBytes(imageData) // Override with default options for this example
	if err := processor.Err(); err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}

	// Get original image info
	originalImg, err := processor.Image()
	if err != nil {
		log.Fatalf("Failed to get original image: %v", err)
	}
	originalBounds := originalImg.Bounds()
	fmt.Printf("Original image size: %dx%d\n", originalBounds.Dx(), originalBounds.Dy())

	// 1. Basic operations chain
	fmt.Println("\n1. Basic Operations Chain:")
	start := time.Now()

	basicProcessor := processor.Clone()
	basicProcessor.Grayscale().Resize(400, 300)
	if err := basicProcessor.Err(); err != nil {
		log.Fatalf("Failed basic operations: %v", err)
	}

	basicTime := time.Since(start)
	fmt.Printf("Basic operations (grayscale + resize) took: %v\n", basicTime)

	// Save basic result
	basicData, err := basicProcessor.ToBytes(gopiq.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to convert basic result: %v", err)
	}
	os.WriteFile("comprehensive_basic.png", basicData, 0644)
	fmt.Println("✓ Basic operations saved as: comprehensive_basic.png")

	// 2. Watermark with different styles
	fmt.Println("\n2. Watermark Variations:")

	watermarkStyles := []struct {
		name     string
		text     string
		position gopiq.WatermarkPosition
		color    color.Color
		size     float64
		filename string
	}{
		{"Elegant", "ELEGANT", gopiq.PositionBottomRight, color.RGBA{255, 255, 255, 150}, 32, "comprehensive_watermark_elegant.png"},
		{"Bold", "BOLD", gopiq.PositionCenter, color.RGBA{255, 0, 0, 200}, 48, "comprehensive_watermark_bold.png"},
		{"Subtle", "SUBTLE", gopiq.PositionTopLeft, color.RGBA{0, 0, 0, 100}, 24, "comprehensive_watermark_subtle.png"},
	}

	for _, style := range watermarkStyles {
		watermarkProcessor := processor.Clone()
		watermarkProcessor.AddTextWatermark(style.text,
			gopiq.WithPosition(style.position),
			gopiq.WithColor(style.color),
			gopiq.WithFontSize(style.size),
			gopiq.WithOffset(20, 20),
		)
		if err := watermarkProcessor.Err(); err != nil {
			log.Fatalf("Failed watermark (%s): %v", style.name, err)
		}

		watermarkData, err := watermarkProcessor.ToBytes(gopiq.FormatPNG)
		if err != nil {
			log.Fatalf("Failed to convert watermark: %v", err)
		}
		os.WriteFile(style.filename, watermarkData, 0644)
		fmt.Printf("✓ %s watermark saved as: %s\n", style.name, style.filename)
	}

	// 3. Performance comparison
	fmt.Println("\n3. Performance Comparison:")

	// Regular grayscale
	start = time.Now()
	regularProcessor := processor.Clone()
	regularProcessor.Grayscale()
	regularTime := time.Since(start)

	regularData, _ := regularProcessor.ToBytes(gopiq.FormatPNG)
	os.WriteFile("comprehensive_grayscale_regular.png", regularData, 0644)
	fmt.Printf("Regular grayscale: %v\n", regularTime)

	// Fast grayscale
	start = time.Now()
	fastProcessor := processor.Clone()
	fastProcessor.GrayscaleFast()
	fastTime := time.Since(start)

	fastData, _ := fastProcessor.ToBytes(gopiq.FormatPNG)
	os.WriteFile("comprehensive_grayscale_fast.png", fastData, 0644)
	fmt.Printf("Fast grayscale: %v\n", fastTime)

	if fastTime < regularTime {
		speedup := float64(regularTime) / float64(fastTime)
		fmt.Printf("Fast method is %.2fx faster\n", speedup)
	}

	// 4. Complex processing pipeline
	fmt.Println("\n4. Complex Processing Pipeline:")

	pipelineProcessor := processor.Clone()

	// Pipeline: crop -> resize -> grayscale -> watermark
	pipelineProcessor.Crop(100, 100, 600, 400).
		Resize(800, 600).
		GrayscaleFast().
		AddTextWatermark("PROCESSED",
			gopiq.WithPosition(gopiq.PositionCenter),
			gopiq.WithColor(color.RGBA{255, 255, 255, 180}),
			gopiq.WithFontSize(36),
		)

	if err := pipelineProcessor.Err(); err != nil {
		log.Fatalf("Failed pipeline: %v", err)
	}

	pipelineData, err := pipelineProcessor.ToBytes(gopiq.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to convert pipeline result: %v", err)
	}
	os.WriteFile("comprehensive_pipeline.png", pipelineData, 0644)
	fmt.Println("✓ Complex pipeline saved as: comprehensive_pipeline.png")

	// 5. Different output formats
	fmt.Println("\n5. Different Output Formats:")

	formatProcessor := processor.Clone()
	formatProcessor.Resize(300, 200)
	if err := formatProcessor.Err(); err != nil {
		log.Fatalf("Failed to resize for format test: %v", err)
	}

	// PNG format
	pngData, err := formatProcessor.ToBytes(gopiq.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to convert to PNG: %v", err)
	}
	os.WriteFile("comprehensive_format.png", pngData, 0644)
	fmt.Println("✓ PNG format saved as: comprehensive_format.png")

	// JPEG format
	jpegData, err := formatProcessor.ToBytes(gopiq.FormatJPEG)
	if err != nil {
		log.Fatalf("Failed to convert to JPEG: %v", err)
	}
	os.WriteFile("comprehensive_format.jpg", jpegData, 0644)
	fmt.Println("✓ JPEG format saved as: comprehensive_format.jpg")

	// 6. Error handling demonstration
	fmt.Println("\n6. Error Handling Demonstration:")

	errorProcessor := processor.Clone()

	// Try invalid crop
	errorProcessor.Crop(1000, 1000, 200, 200)
	if err := errorProcessor.Err(); err != nil {
		fmt.Printf("✓ Caught crop error: %v\n", err)
	}

	// Try invalid resize
	errorProcessor2 := processor.Clone()
	errorProcessor2.Resize(-100, 200)
	if err := errorProcessor2.Err(); err != nil {
		fmt.Printf("✓ Caught resize error: %v\n", err)
	}

	// Try empty watermark
	errorProcessor3 := processor.Clone()
	errorProcessor3.AddTextWatermark("")
	if err := errorProcessor3.Err(); err != nil {
		fmt.Printf("✓ Caught watermark error: %v\n", err)
	}

	// 7. Memory efficiency demonstration
	fmt.Println("\n7. Memory Efficiency:")

	// Process multiple images using the same processor
	for i := 0; i < 3; i++ {
		efficiencyProcessor := processor.Clone()
		efficiencyProcessor.Resize(200, 150).Grayscale()
		if err := efficiencyProcessor.Err(); err != nil {
			log.Fatalf("Failed efficiency test %d: %v", i, err)
		}

		efficiencyData, err := efficiencyProcessor.ToBytes(gopiq.FormatPNG)
		if err != nil {
			log.Fatalf("Failed to convert efficiency result %d: %v", i, err)
		}

		filename := fmt.Sprintf("comprehensive_efficiency_%d.png", i)
		os.WriteFile(filename, efficiencyData, 0644)
		fmt.Printf("✓ Efficiency test %d saved as: %s\n", i, filename)
	}

	fmt.Println("\n=== Comprehensive Example Complete ===")
	fmt.Println("This example demonstrated:")
	fmt.Println("- Basic operations chaining")
	fmt.Println("- Multiple watermark styles")
	fmt.Println("- Performance comparison")
	fmt.Println("- Complex processing pipeline")
	fmt.Println("- Different output formats")
	fmt.Println("- Error handling")
	fmt.Println("- Memory efficiency")
	fmt.Println("\nAll output files have been saved in the examples directory.")
}
