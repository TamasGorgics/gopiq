package main

import (
	"fmt"
	"log"
	"os"

	"github.com/TamasGorgics/gopiq"
)

func main() {
	fmt.Println("=== Resize Example ===")

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

	// Different resize scenarios
	resizeScenarios := []struct {
		name     string
		width    int
		height   int
		filename string
	}{
		{"Small Thumbnail", 100, 75, "resize_thumbnail.png"},
		{"Medium Size", 400, 300, "resize_medium.png"},
		{"Large Size", 1200, 900, "resize_large.png"},
		{"Square Format", 500, 500, "resize_square.png"},
		{"Wide Format", 800, 200, "resize_wide.png"},
		{"Tall Format", 200, 800, "resize_tall.png"},
	}

	for _, scenario := range resizeScenarios {
		// Clone the original processor for each resize operation
		resizeProcessor := originalProcessor.Clone()

		// Resize the image
		resizeProcessor.Resize(scenario.width, scenario.height)
		if err := resizeProcessor.Err(); err != nil {
			log.Fatalf("Failed to resize (%s): %v", scenario.name, err)
		}

		// Get the resized image to verify dimensions
		resizedImg, err := resizeProcessor.Image()
		if err != nil {
			log.Fatalf("Failed to get resized image: %v", err)
		}
		resizedBounds := resizedImg.Bounds()
		fmt.Printf("Resized to %dx%d (requested: %dx%d)\n",
			resizedBounds.Dx(), resizedBounds.Dy(), scenario.width, scenario.height)

		// Save resized image
		resizedData, err := resizeProcessor.ToBytes(gopiq.FormatPNG)
		if err != nil {
			log.Fatalf("Failed to convert resized image to bytes: %v", err)
		}

		err = os.WriteFile(scenario.filename, resizedData, 0644)
		if err != nil {
			log.Fatalf("Failed to save resized image: %v", err)
		}
		fmt.Printf("✓ %s saved as: %s\n", scenario.name, scenario.filename)
	}

	// Demonstrate chaining operations: resize then grayscale
	fmt.Println("\nDemonstrating chaining operations...")
	chainProcessor := originalProcessor.Clone()

	// Chain: resize to medium size, then convert to grayscale
	chainProcessor.Resize(400, 300).Grayscale()
	if err := chainProcessor.Err(); err != nil {
		log.Fatalf("Failed to chain operations: %v", err)
	}

	// Save chained result
	chainedData, err := chainProcessor.ToBytes(gopiq.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to convert chained result to bytes: %v", err)
	}

	err = os.WriteFile("resize_chain_grayscale.png", chainedData, 0644)
	if err != nil {
		log.Fatalf("Failed to save chained result: %v", err)
	}
	fmt.Println("✓ Chained operations (resize + grayscale) saved as: resize_chain_grayscale.png")

	// Demonstrate aspect ratio preservation
	fmt.Println("\nDemonstrating aspect ratio calculations...")
	originalWidth := float64(originalBounds.Dx())
	originalHeight := float64(originalBounds.Dy())
	aspectRatio := originalWidth / originalHeight
	fmt.Printf("Original aspect ratio: %.2f\n", aspectRatio)

	// Resize while maintaining aspect ratio
	targetWidth := 600
	targetHeight := int(float64(targetWidth) / aspectRatio)

	aspectProcessor := originalProcessor.Clone()
	aspectProcessor.Resize(targetWidth, targetHeight)
	if err := aspectProcessor.Err(); err != nil {
		log.Fatalf("Failed to resize with aspect ratio: %v", err)
	}

	aspectData, err := aspectProcessor.ToBytes(gopiq.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to convert aspect ratio image to bytes: %v", err)
	}

	err = os.WriteFile("resize_aspect_ratio.png", aspectData, 0644)
	if err != nil {
		log.Fatalf("Failed to save aspect ratio image: %v", err)
	}
	fmt.Printf("✓ Aspect ratio preserved resize (%dx%d) saved as: resize_aspect_ratio.png\n",
		targetWidth, targetHeight)

	fmt.Println("\n=== Resize Example Complete ===")
	fmt.Println("Generated files:")
	fmt.Println("- resize_thumbnail.png (100x75)")
	fmt.Println("- resize_medium.png (400x300)")
	fmt.Println("- resize_large.png (1200x900)")
	fmt.Println("- resize_square.png (500x500)")
	fmt.Println("- resize_wide.png (800x200)")
	fmt.Println("- resize_tall.png (200x800)")
	fmt.Println("- resize_chain_grayscale.png (resize + grayscale)")
	fmt.Println("- resize_aspect_ratio.png (600x450, maintaining aspect ratio)")
}
