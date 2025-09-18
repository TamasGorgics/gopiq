package main

import (
	"fmt"
	"log"
	"os"

	"github.com/TamasGorgics/gopiq"
)

func main() {
	fmt.Println("=== Crop Example ===")

	// Read the sample image
	imageData, err := os.ReadFile("sample_image.png")
	if err != nil {
		log.Fatalf("Failed to read sample image: %v", err)
	}

	// Get original image dimensions
	originalProcessor := gopiq.FromBytes(imageData)
	if err := originalProcessor.Err(); err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}

	originalImg, err := originalProcessor.Image()
	if err != nil {
		log.Fatalf("Failed to get original image: %v", err)
	}
	originalBounds := originalImg.Bounds()
	fmt.Printf("Original image size: %dx%d\n", originalBounds.Dx(), originalBounds.Dy())

	// Different crop scenarios
	cropScenarios := []struct {
		name     string
		x        int
		y        int
		width    int
		height   int
		filename string
	}{
		{"Top Left Corner", 0, 0, 200, 150, "crop_top_left.png"},
		{"Top Right Corner", 600, 0, 200, 150, "crop_top_right.png"},
		{"Bottom Left Corner", 0, 450, 200, 150, "crop_bottom_left.png"},
		{"Bottom Right Corner", 600, 450, 200, 150, "crop_bottom_right.png"},
		{"Center Square", 300, 200, 200, 200, "crop_center_square.png"},
		{"Center Rectangle", 250, 150, 300, 200, "crop_center_rect.png"},
		{"Small Detail", 100, 100, 100, 100, "crop_small_detail.png"},
		{"Wide Strip", 0, 250, 800, 100, "crop_wide_strip.png"},
		{"Tall Strip", 350, 0, 100, 600, "crop_tall_strip.png"},
	}

	for _, scenario := range cropScenarios {
		// Clone the original processor for each crop operation
		cropProcessor := originalProcessor.Clone()

		// Crop the image
		cropProcessor.Crop(scenario.x, scenario.y, scenario.width, scenario.height)
		if err := cropProcessor.Err(); err != nil {
			log.Fatalf("Failed to crop (%s): %v", scenario.name, err)
		}

		// Get the cropped image to verify dimensions
		croppedImg, err := cropProcessor.Image()
		if err != nil {
			log.Fatalf("Failed to get cropped image: %v", err)
		}
		croppedBounds := croppedImg.Bounds()
		fmt.Printf("Cropped to %dx%d (requested: %dx%d from %d,%d)\n",
			croppedBounds.Dx(), croppedBounds.Dy(),
			scenario.width, scenario.height, scenario.x, scenario.y)

		// Save cropped image
		croppedData, err := cropProcessor.ToBytes(gopiq.FormatPNG)
		if err != nil {
			log.Fatalf("Failed to convert cropped image to bytes: %v", err)
		}

		err = os.WriteFile(scenario.filename, croppedData, 0644)
		if err != nil {
			log.Fatalf("Failed to save cropped image: %v", err)
		}
		fmt.Printf("✓ %s saved as: %s\n", scenario.name, scenario.filename)
	}

	// Demonstrate chaining operations: crop then resize
	fmt.Println("\nDemonstrating chaining operations...")
	chainProcessor := originalProcessor.Clone()

	// Chain: crop center square, then resize to thumbnail
	chainProcessor.Crop(300, 200, 200, 200).Resize(100, 100)
	if err := chainProcessor.Err(); err != nil {
		log.Fatalf("Failed to chain crop and resize: %v", err)
	}

	// Save chained result
	chainedData, err := chainProcessor.ToBytes(gopiq.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to convert chained result to bytes: %v", err)
	}

	err = os.WriteFile("crop_chain_resize.png", chainedData, 0644)
	if err != nil {
		log.Fatalf("Failed to save chained result: %v", err)
	}
	fmt.Println("✓ Chained operations (crop + resize) saved as: crop_chain_resize.png")

	// Demonstrate multiple crops from different areas
	fmt.Println("\nDemonstrating multiple crops...")

	// Create a grid of crops
	gridSize := 4
	cropWidth := originalBounds.Dx() / gridSize
	cropHeight := originalBounds.Dy() / gridSize

	for row := 0; row < gridSize; row++ {
		for col := 0; col < gridSize; col++ {
			x := col * cropWidth
			y := row * cropHeight

			gridProcessor := originalProcessor.Clone()
			gridProcessor.Crop(x, y, cropWidth, cropHeight)
			if err := gridProcessor.Err(); err != nil {
				log.Fatalf("Failed to crop grid cell (%d,%d): %v", row, col, err)
			}

			gridData, err := gridProcessor.ToBytes(gopiq.FormatPNG)
			if err != nil {
				log.Fatalf("Failed to convert grid cell to bytes: %v", err)
			}

			filename := fmt.Sprintf("crop_grid_%d_%d.png", row, col)
			err = os.WriteFile(filename, gridData, 0644)
			if err != nil {
				log.Fatalf("Failed to save grid cell: %v", err)
			}
		}
	}
	fmt.Printf("✓ Grid of %dx%d crops saved as crop_grid_*.png\n", gridSize, gridSize)

	// Test error handling with invalid crop parameters
	fmt.Println("\nTesting error handling...")
	errorProcessor := originalProcessor.Clone()

	// Try to crop outside image bounds
	errorProcessor.Crop(900, 700, 200, 200) // This should fail
	if err := errorProcessor.Err(); err != nil {
		fmt.Printf("✓ Correctly caught error: %v\n", err)
	} else {
		fmt.Println("✗ Expected error for out-of-bounds crop, but none occurred")
	}

	// Try with negative dimensions
	errorProcessor2 := originalProcessor.Clone()
	errorProcessor2.Crop(100, 100, -50, 100) // This should fail
	if err := errorProcessor2.Err(); err != nil {
		fmt.Printf("✓ Correctly caught error: %v\n", err)
	} else {
		fmt.Println("✗ Expected error for negative dimensions, but none occurred")
	}

	fmt.Println("\n=== Crop Example Complete ===")
	fmt.Println("Generated files:")
	fmt.Println("- crop_top_left.png")
	fmt.Println("- crop_top_right.png")
	fmt.Println("- crop_bottom_left.png")
	fmt.Println("- crop_bottom_right.png")
	fmt.Println("- crop_center_square.png")
	fmt.Println("- crop_center_rect.png")
	fmt.Println("- crop_small_detail.png")
	fmt.Println("- crop_wide_strip.png")
	fmt.Println("- crop_tall_strip.png")
	fmt.Println("- crop_chain_resize.png (crop + resize)")
	fmt.Println("- crop_grid_*.png (4x4 grid of crops)")
}
