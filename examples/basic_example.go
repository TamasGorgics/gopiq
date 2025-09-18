package main

import (
	"fmt"
	"log"
	"os"

	"github.com/TamasGorgics/gopiq"
)

func main() {
	fmt.Println("=== Basic Gopiq Example ===")

	// Read the sample image
	imageData, err := os.ReadFile("sample_image.png")
	if err != nil {
		log.Fatalf("Failed to read sample image: %v", err)
	}

	// Create ImageProcessor from bytes
	processor := gopiq.FromBytes(imageData)
	if err := processor.Err(); err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}

	// Get original image info
	img, err := processor.Image()
	if err != nil {
		log.Fatalf("Failed to get image: %v", err)
	}
	bounds := img.Bounds()
	fmt.Printf("Original image size: %dx%d\n", bounds.Dx(), bounds.Dy())

	// Convert to grayscale
	processor.Grayscale()
	if err := processor.Err(); err != nil {
		log.Fatalf("Failed to convert to grayscale: %v", err)
	}

	// Save grayscale version
	grayscaleData, err := processor.ToBytes(gopiq.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to convert to bytes: %v", err)
	}

	err = os.WriteFile("output_grayscale.png", grayscaleData, 0644)
	if err != nil {
		log.Fatalf("Failed to save grayscale image: %v", err)
	}
	fmt.Println("✓ Grayscale image saved as: output_grayscale.png")

	// Resize the image
	processor.Resize(400, 300)
	if err := processor.Err(); err != nil {
		log.Fatalf("Failed to resize: %v", err)
	}

	// Save resized version
	resizedData, err := processor.ToBytes(gopiq.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to convert resized image to bytes: %v", err)
	}

	err = os.WriteFile("output_resized.png", resizedData, 0644)
	if err != nil {
		log.Fatalf("Failed to save resized image: %v", err)
	}
	fmt.Println("✓ Resized image saved as: output_resized.png")

	// Chain operations: crop a portion of the image
	processor.Crop(50, 50, 200, 150)
	if err := processor.Err(); err != nil {
		log.Fatalf("Failed to crop: %v", err)
	}

	// Save cropped version
	croppedData, err := processor.ToBytes(gopiq.FormatPNG)
	if err != nil {
		log.Fatalf("Failed to convert cropped image to bytes: %v", err)
	}

	err = os.WriteFile("output_cropped.png", croppedData, 0644)
	if err != nil {
		log.Fatalf("Failed to save cropped image: %v", err)
	}
	fmt.Println("✓ Cropped image saved as: output_cropped.png")

	fmt.Println("\n=== Basic Example Complete ===")
	fmt.Println("Generated files:")
	fmt.Println("- output_grayscale.png (grayscale version)")
	fmt.Println("- output_resized.png (400x300 resized version)")
	fmt.Println("- output_cropped.png (cropped 200x150 portion)")
}
