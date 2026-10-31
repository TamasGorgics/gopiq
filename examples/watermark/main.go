package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"

	"github.com/TamasGorgics/gopiq"
)

func main() {
	fmt.Println("=== Watermark Example ===")

	// Read the sample image
	imageData, err := os.ReadFile("examples/sample_image.png")
	if err != nil {
		log.Fatalf("Failed to read sample image: %v", err)
	}
	outputDir := filepath.Join("examples", "watermark", "output")
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}
	outputPath := func(name string) string { return filepath.Join(outputDir, name) }

	// Create ImageProcessor from bytes
	processor := gopiq.FromBytes(imageData)
	if err := processor.Err(); err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}

	// Add watermark in different positions
	positions := []struct {
		pos      gopiq.WatermarkPosition
		name     string
		filename string
	}{
		{gopiq.PositionTopLeft, "Top Left", "watermark_top_left.png"},
		{gopiq.PositionTopRight, "Top Right", "watermark_top_right.png"},
		{gopiq.PositionBottomLeft, "Bottom Left", "watermark_bottom_left.png"},
		{gopiq.PositionBottomRight, "Bottom Right", "watermark_bottom_right.png"},
		{gopiq.PositionCenter, "Center", "watermark_center.png"},
	}

	for _, p := range positions {
		// Clone the processor for each watermark position
		watermarkProcessor := processor.Clone()

		// Add watermark with different colors and sizes
		watermarkProcessor.AddTextWatermark("SAMPLE",
			gopiq.WithPosition(p.pos),
			gopiq.WithFontSize(48),
			gopiq.WithColor(color.RGBA{255, 255, 255, 200}), // White with transparency
			gopiq.WithOffset(20, 20),
		)

		if err := watermarkProcessor.Err(); err != nil {
			log.Fatalf("Failed to add watermark (%s): %v", p.name, err)
		}

		// Save watermarked image
		watermarkedData, err := watermarkProcessor.ToBytes(gopiq.FormatPNG)
		if err != nil {
			log.Fatalf("Failed to convert watermarked image to bytes: %v", err)
		}

		err = os.WriteFile(outputPath(p.filename), watermarkedData, 0644)
		if err != nil {
			log.Fatalf("Failed to save watermarked image: %v", err)
		}
		fmt.Printf("✓ %s watermark saved as: %s\n", p.name, outputPath(p.filename))
	}

	// Example with custom styling
	customProcessor := processor.Clone()
	customProcessor.AddTextWatermark("CUSTOM STYLE",
		gopiq.WithPosition(gopiq.PositionCenter),
		gopiq.WithFontSize(36),
		gopiq.WithColor(color.RGBA{255, 0, 0, 180}), // Red with transparency
		gopiq.WithOffset(0, 0),
	)

	if err := customProcessor.Err(); err != nil {
		log.Fatalf("Failed to add custom watermark: %v", err)
	}

	customData, err := customProcessor.ToBytes(gopiq.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to convert custom watermarked image to bytes: %v", err)
	}

	err = os.WriteFile(outputPath("watermark_custom_style.png"), customData, 0644)
	if err != nil {
		log.Fatalf("Failed to save custom watermarked image: %v", err)
	}
	fmt.Printf("✓ Custom style watermark saved as: %s\n", outputPath("watermark_custom_style.png"))

	fmt.Println("\n=== Watermark Example Complete ===")
	fmt.Println("Generated files:")
	fmt.Printf("- %s\n", outputPath("watermark_top_left.png"))
	fmt.Printf("- %s\n", outputPath("watermark_top_right.png"))
	fmt.Printf("- %s\n", outputPath("watermark_bottom_left.png"))
	fmt.Printf("- %s\n", outputPath("watermark_bottom_right.png"))
	fmt.Printf("- %s\n", outputPath("watermark_center.png"))
	fmt.Printf("- %s\n", outputPath("watermark_custom_style.png"))
}
