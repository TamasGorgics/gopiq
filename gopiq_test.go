package gopiq

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"sync"
	"testing"
	"time"
)

// Helper function to create a simple test image (RGBA)
func createTestImage(width, height int) image.Image {
	img := newRGBA(image.Rect(0, 0, width, height))
	// Fill with a simple pattern for visual distinction
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if (x/10)%2 == (y/10)%2 {
				img.Set(x, y, color.RGBA{0, 0, 0, 255}) // Black
			} else {
				img.Set(x, y, color.RGBA{255, 255, 255, 255}) // White
			}
		}
	}
	return img
}

// Helper to convert image.Image to JPEG bytes
func imageToJPEGBytes(img image.Image) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	return buf.Bytes(), err
}

// Helper to convert image.Image to PNG bytes
func imageToPNGBytes(img image.Image) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	return buf.Bytes(), err
}

func TestNew(t *testing.T) {
	// Test case: Valid image
	img := createTestImage(10, 10)
	proc := New(img)
	if proc.Err() != nil {
		t.Fatalf("New() with valid image should not return an error, got: %v", proc.Err())
	}
	if proc.currentImage == nil {
		t.Fatal("New() should set currentImage")
	}

	// Test case: Nil image
	proc = New(nil)
	if proc.Err() == nil {
		t.Fatal("New() with nil image should return an error")
	}
	if proc.currentImage != nil {
		t.Fatal("New() with nil image should not set currentImage")
	}
}

func TestFromBytes(t *testing.T) {
	// Test case: Valid JPEG bytes
	testImg := createTestImage(100, 100)
	jpegBytes, _ := imageToJPEGBytes(testImg)
	proc := FromBytes(jpegBytes)
	if proc.Err() != nil {
		t.Fatalf("FromBytes() with valid JPEG should not error, got: %v", proc.Err())
	}
	if proc.currentImage == nil {
		t.Fatal("FromBytes() should decode image")
	}
	if proc.currentImage.Bounds().Dx() != 100 || proc.currentImage.Bounds().Dy() != 100 {
		t.Errorf("Decoded image dimensions mismatch, got %v", proc.currentImage.Bounds())
	}

	// Test case: Valid PNG bytes
	pngBytes, _ := imageToPNGBytes(testImg)
	proc = FromBytes(pngBytes)
	if proc.Err() != nil {
		t.Fatalf("FromBytes() with valid PNG should not error, got: %v", proc.Err())
	}

	// Test case: Empty bytes
	proc = FromBytes([]byte{})
	if proc.Err() == nil {
		t.Fatal("FromBytes() with empty bytes should return an error")
	}

	// Test case: Invalid format bytes (not a known image header)
	proc = FromBytes([]byte{1, 2, 3, 4, 5})
	if proc.Err() == nil {
		t.Fatal("FromBytes() with invalid format bytes should return an error")
	}
}

func TestToBytes(t *testing.T) {
	testImg := createTestImage(50, 50)
	proc := New(testImg)

	// Test case: To JPEG bytes
	jpegData, err := proc.ToBytes(FormatJPEG)
	if err != nil {
		t.Fatalf("ToBytes(FormatJPEG) should not error, got: %v", err)
	}
	if len(jpegData) == 0 {
		t.Fatal("ToBytes(FormatJPEG) returned empty bytes")
	}
	// Try decoding back to verify
	_, err = decodeImage(bytes.NewReader(jpegData))
	if err != nil {
		t.Errorf("Failed to decode JPEG bytes produced by ToBytes: %v", err)
	}

	// Test case: To PNG bytes
	pngData, err := proc.ToBytes(FormatPNG)
	if err != nil {
		t.Fatalf("ToBytes(FormatPNG) should not error, got: %v", err)
	}
	if len(pngData) == 0 {
		t.Fatal("ToBytes(FormatPNG) returned empty bytes")
	}
	// Try decoding back to verify
	_, err = decodeImage(bytes.NewReader(pngData))
	if err != nil {
		t.Errorf("Failed to decode PNG bytes produced by ToBytes: %v", err)
	}

	// Test case: Unsupported format (e.g., GIF)
	_, err = proc.ToBytes(FormatGIF) // GIF encoding is not supported in stdlib without color quantization
	if err == nil {
		t.Fatal("ToBytes() with unsupported format (GIF) should return an error")
	}

	// Test case: Processor with a prior error
	procWithErr := New(nil) // Create a processor with an initial error
	_, err = procWithErr.ToBytes(FormatPNG)
	if err == nil {
		t.Fatal("ToBytes() on a processor with prior error should return that error")
	}
}

func TestCrop(t *testing.T) {
	originalImg := createTestImage(200, 150)
	proc := New(originalImg)

	// Test case: Valid crop
	croppedProc := proc.Crop(50, 50, 100, 75)
	if croppedProc.Err() != nil {
		t.Fatalf("Crop() with valid dimensions should not error, got: %v", croppedProc.Err())
	}
	if croppedProc.currentImage.Bounds().Dx() != 100 || croppedProc.currentImage.Bounds().Dy() != 75 {
		t.Errorf("Cropped image dimensions mismatch, expected 100x75, got %v", croppedProc.currentImage.Bounds().Size())
	}
	// Verify a pixel was moved correctly (e.g., original (50,50) is now (0,0) in cropped)
	originalPixel := originalImg.At(50, 50)
	croppedPixel := croppedProc.currentImage.At(0, 0)

	origR, origG, origB, origA := originalPixel.RGBA()
	cropR, cropG, cropB, cropA := croppedPixel.RGBA()

	if origR != cropR || origG != cropG || origB != cropB || origA != cropA {
		t.Errorf("Pixel at (0,0) of cropped image does not match original at (50,50): orig RGBA(%d,%d,%d,%d) vs crop RGBA(%d,%d,%d,%d)",
			origR>>8, origG>>8, origB>>8, origA>>8, cropR>>8, cropG>>8, cropB>>8, cropA>>8)
	}

	// Test case: Crop with zero width
	proc = New(originalImg)
	croppedProc = proc.Crop(0, 0, 0, 50)
	if croppedProc.Err() == nil {
		t.Fatal("Crop() with zero width should return an error")
	}

	// Test case: Crop with negative height
	proc = New(originalImg)
	croppedProc = proc.Crop(0, 0, 50, -10)
	if croppedProc.Err() == nil {
		t.Fatal("Crop() with negative height should return an error")
	}

	// Test case: Crop out of bounds (x-axis)
	proc = New(originalImg)
	croppedProc = proc.Crop(150, 0, 100, 50) // x+width = 250, original width = 200
	if croppedProc.Err() == nil {
		t.Fatal("Crop() out of bounds should return an error")
	}

	// Test case: Crop out of bounds (y-axis)
	proc = New(originalImg)
	croppedProc = proc.Crop(0, 100, 50, 100) // y+height = 200, original height = 150
	if croppedProc.Err() == nil {
		t.Fatal("Crop() out of bounds should return an error")
	}

	// Test case: Chaining with a prior error
	procWithErr := New(nil) // Create a processor with an initial error
	croppedProc = procWithErr.Crop(10, 10, 10, 10)
	if croppedProc.Err() == nil {
		t.Fatal("Crop() on a processor with prior error should propagate that error")
	}
}

func TestResize(t *testing.T) {
	originalImg := createTestImage(200, 150)
	proc := New(originalImg)

	// Test case: Valid resize (downscale)
	resizedProc := proc.Resize(100, 75)
	if resizedProc.Err() != nil {
		t.Fatalf("Resize() with valid dimensions should not error, got: %v", resizedProc.Err())
	}
	if resizedProc.currentImage.Bounds().Dx() != 100 || resizedProc.currentImage.Bounds().Dy() != 75 {
		t.Errorf("Resized image dimensions mismatch, expected 100x75, got %v", resizedProc.currentImage.Bounds().Size())
	}

	// Test case: Valid resize (upscale)
	proc = New(originalImg)
	resizedProc = proc.Resize(400, 300)
	if resizedProc.Err() != nil {
		t.Fatalf("Resize() with valid dimensions should not error, got: %v", resizedProc.Err())
	}
	if resizedProc.currentImage.Bounds().Dx() != 400 || resizedProc.currentImage.Bounds().Dy() != 300 {
		t.Errorf("Resized image dimensions mismatch, expected 400x300, got %v", resizedProc.currentImage.Bounds().Size())
	}

	// Test case: Resize with zero width
	proc = New(originalImg)
	resizedProc = proc.Resize(0, 50)
	if resizedProc.Err() == nil {
		t.Fatal("Resize() with zero width should return an error")
	}

	// Test case: Resize with negative height
	proc = New(originalImg)
	resizedProc = proc.Resize(50, -10)
	if resizedProc.Err() == nil {
		t.Fatal("Resize() with negative height should return an error")
	}

	// Test case: Chaining with a prior error
	procWithErr := New(nil)
	resizedProc = procWithErr.Resize(100, 100)
	if resizedProc.Err() == nil {
		t.Fatal("Resize() on a processor with prior error should propagate that error")
	}
}

func TestGrayscale(t *testing.T) {
	originalImg := image.NewRGBA(image.Rect(0, 0, 50, 50))
	// Fill with a colorful pixel
	originalImg.Set(25, 25, color.RGBA{R: 100, G: 150, B: 200, A: 255}) // A blue-ish color

	proc := New(originalImg)
	grayProc := proc.Grayscale()

	if grayProc.Err() != nil {
		t.Fatalf("Grayscale() should not return an error, got: %v", grayProc.Err())
	}

	// Verify dimensions are the same
	if grayProc.currentImage.Bounds() != originalImg.Bounds() {
		t.Errorf("Grayscale image dimensions mismatch, expected %v, got %v", originalImg.Bounds(), grayProc.currentImage.Bounds())
	}

	// Verify a pixel is truly grayscale by checking that R=G=B
	// The grayscale image is stored as RGBA, so we need to extract the RGBA values
	grayPixelColor := grayProc.currentImage.At(25, 25)
	r, g, b, _ := grayPixelColor.RGBA()

	// In a grayscale image, R, G, and B should be equal
	if r != g || g != b {
		t.Errorf("Pixel at (25,25) is not grayscale: R=%d, G=%d, B=%d", r>>8, g>>8, b>>8)
	}

	// Check another pixel with a different color
	originalImg.Set(10, 10, color.RGBA{R: 255, G: 0, B: 0, A: 255}) // Red
	proc = New(originalImg)
	grayProc = proc.Grayscale()

	grayPixelColor = grayProc.currentImage.At(10, 10)
	r, g, b, _ = grayPixelColor.RGBA()

	// In a grayscale image, R, G, and B should be equal
	if r != g || g != b {
		t.Errorf("Pixel at (10,10) is not grayscale: R=%d, G=%d, B=%d", r>>8, g>>8, b>>8)
	}

	// Test case: Chaining with a prior error
	procWithErr := New(nil)
	grayProc = procWithErr.Grayscale()
	if grayProc.Err() == nil {
		t.Fatal("Grayscale() on a processor with prior error should propagate that error")
	}
}

func TestAddTextWatermark(t *testing.T) {
	originalImg := createTestImage(300, 200)
	proc := New(originalImg)

	// Test case 1: Basic watermark using default Go font
	watermarkedProc := proc.AddTextWatermark("TEST",
		WithFontSize(20),                      // Use float64 now
		WithColor(color.RGBA{255, 0, 0, 255}), // Red
		WithPosition(PositionBottomRight),
		WithOffset(5, 5),
	)
	if watermarkedProc.Err() != nil {
		t.Fatalf("AddTextWatermark basic should not error: %v", watermarkedProc.Err())
	}
	// For visual inspection of proper text rendering:
	// outFile, _ := os.Create("watermarked_std_lib_font.png")
	// defer outFile.Close()
	// watermarkedProc.ToBytes(FormatPNG)

	// Test case 2: Watermark with empty text
	proc = New(originalImg)
	watermarkedProc = proc.AddTextWatermark("", WithFontSize(20))
	if watermarkedProc.Err() == nil {
		t.Fatal("AddTextWatermark with empty text should error")
	}

	// Test case 3: Chaining with a prior error
	procWithErr := New(nil)
	watermarkedProc = procWithErr.AddTextWatermark("Error Propagated", WithFontSize(20))
	if watermarkedProc.Err() == nil {
		t.Fatal("AddTextWatermark on processor with prior error should propagate")
	}

	// Test case 4: Different position (Center)
	proc = New(originalImg)
	watermarkedProc = proc.AddTextWatermark("CENTER",
		WithFontSize(30),
		WithColor(color.RGBA{0, 255, 0, 255}), // Green
		WithPosition(PositionCenter),
	)
	if watermarkedProc.Err() != nil {
		t.Fatalf("AddTextWatermark center should not error: %v", watermarkedProc.Err())
	}

	// Test case 5: Invalid font bytes
	proc = New(originalImg)
	watermarkedProc = proc.AddTextWatermark("Invalid Bytes", WithFontBytes([]byte{1, 2, 3, 4}))
	if watermarkedProc.Err() == nil {
		t.Fatal("AddTextWatermark with invalid font bytes should error")
	}

	// Test case 6: No font specified (clearing default)
	proc = New(originalImg)
	watermarkedProc = proc.AddTextWatermark("No Font", func(wc *watermarkConfig) { wc.FontBytes = nil })
	if watermarkedProc.Err() == nil {
		t.Fatal("AddTextWatermark with no font bytes should error")
	}
}

func TestChainingOperations(t *testing.T) {
	originalImg := createTestImage(400, 300)
	proc := New(originalImg)

	// Chain: Resize (Catmull-Rom) -> Grayscale -> Crop -> AddTextWatermark (proper font)
	finalProc := proc.
		Resize(200, 150). // Catmull-Rom resize
		Grayscale().
		Crop(10, 10, 100, 50).
		AddTextWatermark("GOPIQ",
			WithFontSize(10), // Font size is float64 now
			WithColor(color.RGBA{255, 255, 0, 200}),
			WithPosition(PositionCenter))

	if finalProc.Err() != nil {
		t.Fatalf("Chained operations should not error: %v", finalProc.Err())
	}

	// Verify final dimensions
	expectedWidth, expectedHeight := 100, 50
	if finalProc.currentImage.Bounds().Dx() != expectedWidth || finalProc.currentImage.Bounds().Dy() != expectedHeight {
		t.Errorf("Final image dimensions mismatch, expected %dx%d, got %dx%d",
			expectedWidth, expectedHeight, finalProc.currentImage.Bounds().Dx(), finalProc.currentImage.Bounds().Dy())
	}

	// Check if the output can be encoded to bytes (implies valid image)
	_, err := finalProc.ToBytes(FormatPNG)
	if err != nil {
		t.Fatalf("Final processed image should be encodable: %v", err)
	}

	// Test case: Chaining with an error early in the chain
	procWithEarlyError := New(originalImg).
		Resize(0, 0). // This will cause an error
		Grayscale().
		Crop(10, 10, 50, 50).
		AddTextWatermark("Should not reach", WithFontSize(10))

	if procWithEarlyError.Err() == nil {
		t.Fatal("Chaining with an early error should propagate the error")
	}
	// Check that the error is indeed from the Resize operation
	if !bytes.Contains([]byte(procWithEarlyError.Err().Error()), []byte("resize dimensions must be positive")) {
		t.Errorf("Expected resize error, got: %v", procWithEarlyError.Err())
	}
}

func TestImageAndErrMethods(t *testing.T) {
	// Test on a successful processor
	img := createTestImage(10, 10)
	proc := New(img)
	retImg, err := proc.Image()
	if err != nil {
		t.Fatalf("Image() on successful processor should not error, got %v", err)
	}
	if retImg == nil {
		t.Fatal("Image() on successful processor returned nil image")
	}
	if proc.Err() != nil {
		t.Fatalf("Err() on successful processor should be nil, got %v", proc.Err())
	}

	// Test on a processor with an error
	procErr := New(nil) // Initialize with an error
	retImg, err = procErr.Image()
	if err == nil {
		t.Fatal("Image() on erroneous processor should return error")
	}
	if retImg != nil {
		t.Fatal("Image() on erroneous processor should return nil image")
	}
	if procErr.Err() == nil {
		t.Fatal("Err() on erroneous processor should return error")
	}
}

// Test case for DecodeImage in formats.go
func TestDecodeImage(t *testing.T) {
	testImg := createTestImage(50, 50)
	jpegBytes, _ := imageToJPEGBytes(testImg)

	// Valid JPEG decode
	img, err := decodeImage(bytes.NewReader(jpegBytes))
	if err != nil {
		t.Fatalf("decodeImage for JPEG failed: %v", err)
	}
	if img == nil {
		t.Fatal("decodeImage for JPEG returned nil")
	}
	if img.Bounds().Dx() != 50 {
		t.Errorf("Decoded JPEG image has wrong width: %d", img.Bounds().Dx())
	}

	pngBytes, _ := imageToPNGBytes(testImg)
	// Valid PNG decode
	img, err = decodeImage(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("decodeImage for PNG failed: %v", err)
	}
	if img == nil {
		t.Fatal("decodeImage for PNG returned nil")
	}
	if img.Bounds().Dy() != 50 {
		t.Errorf("Decoded PNG image has wrong height: %d", img.Bounds().Dy())
	}

	// Invalid image data
	_, err = decodeImage(bytes.NewReader([]byte("not an image")))
	if err == nil {
		t.Fatal("decodeImage with invalid data should return error")
	}
}

// Test case for encodeImage in formats.go
func TestEncodeImage(t *testing.T) {
	testImg := createTestImage(20, 20)
	var buf bytes.Buffer

	// Valid JPEG encode
	err := encodeImage(&buf, testImg, FormatJPEG)
	if err != nil {
		t.Fatalf("encodeImage for JPEG failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("encodeImage for JPEG returned empty bytes")
	}

	buf.Reset() // Clear buffer for next test
	// Valid PNG encode
	err = encodeImage(&buf, testImg, FormatPNG)
	if err != nil {
		t.Fatalf("encodeImage for PNG failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatal("encodeImage for PNG returned empty bytes")
	}

	buf.Reset()
	// Unsupported format (GIF encoding is not supported in stdlib without color quantization)
	err = encodeImage(&buf, testImg, FormatGIF)
	if err == nil {
		t.Fatal("encodeImage with unsupported format (GIF) should return error")
	}
}

func TestNewRGBA(t *testing.T) {
	bounds := image.Rect(0, 0, 10, 10)
	img := newRGBA(bounds)
	if img == nil {
		t.Fatal("newRGBA returned nil")
	}
	if img.Bounds() != bounds {
		t.Errorf("newRGBA bounds mismatch, expected %v, got %v", bounds, img.Bounds())
	}
}

// --- Thread Safety Tests ---

func TestClone(t *testing.T) {
	originalImg := createTestImage(100, 100)
	proc := New(originalImg)

	// Apply some transformations to the original
	proc.Resize(50, 50).Grayscale()

	// Clone the processor
	cloned := proc.Clone()

	// Verify the clone has the same state
	originalImage, originalErr := proc.Image()
	clonedImage, clonedErr := cloned.Image()

	if originalErr != clonedErr {
		t.Errorf("Clone error mismatch: original=%v, cloned=%v", originalErr, clonedErr)
	}

	if originalImage != clonedImage {
		t.Error("Clone should have the same image reference")
	}

	// Verify they are independent - modify original
	proc.Crop(10, 10, 20, 20)

	// Clone should be unchanged
	clonedImageAfter, _ := cloned.Image()
	if clonedImageAfter != clonedImage {
		t.Error("Clone was affected by changes to original")
	}
}

func TestConcurrentRead(t *testing.T) {
	originalImg := createTestImage(100, 100)
	proc := New(originalImg).Resize(50, 50).Grayscale()

	const numGoroutines = 10
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Multiple goroutines reading concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Test concurrent reads
				img, err := proc.Image()
				if err != nil {
					t.Errorf("Concurrent read failed: %v", err)
					return
				}
				if img == nil {
					t.Error("Concurrent read returned nil image")
					return
				}

				// Test other read methods
				if proc.Err() != nil {
					t.Errorf("Unexpected error in concurrent read: %v", proc.Err())
					return
				}

				_, err = proc.ToBytes(FormatPNG)
				if err != nil {
					t.Errorf("ToBytes failed in concurrent read: %v", err)
					return
				}
			}
		}()
	}

	wg.Wait()
}

func TestConcurrentProcessing(t *testing.T) {
	originalImg := createTestImage(100, 100)

	const numGoroutines = 5
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	results := make([]*ImageProcessor, numGoroutines)

	// Multiple goroutines processing different clones
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()

			// Each goroutine works with its own clone
			proc := New(originalImg).Clone()

			// Apply different transformations based on index
			switch index % 3 {
			case 0:
				proc.Resize(80, 80).Grayscale()
			case 1:
				proc.Crop(10, 10, 50, 50).AddTextWatermark("TEST", WithFontSize(12))
			case 2:
				proc.Resize(60, 60).Crop(5, 5, 40, 40)
			}

			results[index] = proc
		}(i)
	}

	wg.Wait()

	// Verify all operations completed successfully
	for i, result := range results {
		if result.Err() != nil {
			t.Errorf("Goroutine %d failed with error: %v", i, result.Err())
		}

		img, err := result.Image()
		if err != nil {
			t.Errorf("Failed to get image from goroutine %d: %v", i, err)
		}
		if img == nil {
			t.Errorf("Goroutine %d produced nil image", i)
		}
	}
}

func TestConcurrentChaining(t *testing.T) {
	originalImg := createTestImage(200, 200)

	const numGoroutines = 10
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Test concurrent chaining operations
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()

			// Each goroutine creates its own processor
			proc := New(originalImg)

			// Perform a chain of operations
			finalProc := proc.
				Resize(100, 100).
				Grayscale().
				Crop(20, 20, 60, 60).
				AddTextWatermark("Concurrent",
					WithFontSize(10),
					WithPosition(PositionCenter))

			if finalProc.Err() != nil {
				t.Errorf("Concurrent chaining failed in goroutine %d: %v", index, finalProc.Err())
				return
			}

			// Verify final result
			img, err := finalProc.Image()
			if err != nil {
				t.Errorf("Failed to get final image in goroutine %d: %v", index, err)
				return
			}

			if img.Bounds().Dx() != 60 || img.Bounds().Dy() != 60 {
				t.Errorf("Unexpected final dimensions in goroutine %d: %v", index, img.Bounds())
			}
		}(i)
	}

	wg.Wait()
}

func TestRaceConditionDetection(t *testing.T) {
	// This test is designed to catch race conditions when run with -race flag
	originalImg := createTestImage(50, 50)
	proc := New(originalImg)

	const numGoroutines = 20
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Mix of read and write operations to stress test the locking
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()

			if index%2 == 0 {
				// Read operations
				for j := 0; j < 50; j++ {
					proc.Image()
					proc.Err()
					proc.ToBytes(FormatPNG)
					time.Sleep(time.Microsecond)
				}
			} else {
				// Write operations (on clones to avoid interfering with reads)
				clone := proc.Clone()
				for j := 0; j < 20; j++ {
					clone.Resize(30+j, 30+j)
					if j%5 == 0 {
						clone.Grayscale()
					}
					time.Sleep(time.Microsecond)
				}
			}
		}(i)
	}

	wg.Wait()
}

// --- Performance Optimization Tests ---

func TestGrayscaleFast(t *testing.T) {
	originalImg := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Fill with a known pattern
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			originalImg.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}

	proc := New(originalImg)
	grayProc := proc.GrayscaleFast()

	if grayProc.Err() != nil {
		t.Fatalf("GrayscaleFast() should not return an error, got: %v", grayProc.Err())
	}

	// Verify dimensions are the same
	if grayProc.currentImage.Bounds() != originalImg.Bounds() {
		t.Errorf("GrayscaleFast image dimensions mismatch, expected %v, got %v",
			originalImg.Bounds(), grayProc.currentImage.Bounds())
	}

	// Verify a pixel is truly grayscale by checking that R=G=B
	grayPixelColor := grayProc.currentImage.At(50, 50)
	r, g, b, _ := grayPixelColor.RGBA()

	// In a grayscale image, R, G, and B should be equal
	if r != g || g != b {
		t.Errorf("Pixel at (50,50) is not grayscale: R=%d, G=%d, B=%d", r>>8, g>>8, b>>8)
	}
}

func TestGrayscaleConsistency(t *testing.T) {
	// Test that Grayscale() and GrayscaleFast() produce similar results
	originalImg := createTestImage(200, 150)

	proc1 := New(originalImg)
	proc2 := New(originalImg)

	grayStandard := proc1.Grayscale()
	grayFast := proc2.GrayscaleFast()

	if grayStandard.Err() != nil {
		t.Fatalf("Standard grayscale failed: %v", grayStandard.Err())
	}
	if grayFast.Err() != nil {
		t.Fatalf("Fast grayscale failed: %v", grayFast.Err())
	}

	// Compare a few pixels to ensure similar results
	standardImg, _ := grayStandard.Image()
	fastImg, _ := grayFast.Image()

	for _, point := range []image.Point{{50, 50}, {100, 75}, {150, 100}} {
		standardColor := standardImg.At(point.X, point.Y)
		fastColor := fastImg.At(point.X, point.Y)

		sr, sg, sb, sa := standardColor.RGBA()
		fr, fg, fb, fa := fastColor.RGBA()

		// Colors should be very close (allow small differences due to rounding)
		if abs(int(sr>>8)-int(fr>>8)) > 1 ||
			abs(int(sg>>8)-int(fg>>8)) > 1 ||
			abs(int(sb>>8)-int(fb>>8)) > 1 ||
			abs(int(sa>>8)-int(fa>>8)) > 1 {
			t.Errorf("Grayscale methods differ significantly at (%d,%d): standard RGBA(%d,%d,%d,%d) vs fast RGBA(%d,%d,%d,%d)",
				point.X, point.Y, sr>>8, sg>>8, sb>>8, sa>>8, fr>>8, fg>>8, fb>>8, fa>>8)
		}
	}
}

func TestPerformanceOptions(t *testing.T) {
	originalImg := createTestImage(100, 100)

	// Test custom performance options
	opts := PerformanceOptions{
		MaxGoroutines:            2,
		EnableParallelProcessing: true,
		MinSizeForParallel:       5000, // 50x100 = 5000
	}

	proc := NewWithPerformanceOptions(originalImg, opts)
	result := proc.GrayscaleFast()

	if result.Err() != nil {
		t.Fatalf("Performance options test failed: %v", result.Err())
	}

	// Test SetPerformanceOptions
	proc2 := New(originalImg)
	proc2.SetPerformanceOptions(opts)
	result2 := proc2.GrayscaleFast()

	if result2.Err() != nil {
		t.Fatalf("SetPerformanceOptions test failed: %v", result2.Err())
	}
}

func TestParallelProcessingThreshold(t *testing.T) {
	// Test that small images use direct processing, large images use parallel
	smallImg := createTestImage(50, 50)   // 2500 pixels < default threshold
	largeImg := createTestImage(200, 200) // 40000 pixels > default threshold

	opts := DefaultPerformanceOptions()

	proc1 := NewWithPerformanceOptions(smallImg, opts)
	proc2 := NewWithPerformanceOptions(largeImg, opts)

	// Both should work regardless of size
	result1 := proc1.GrayscaleFast()
	result2 := proc2.GrayscaleFast()

	if result1.Err() != nil {
		t.Errorf("Small image processing failed: %v", result1.Err())
	}
	if result2.Err() != nil {
		t.Errorf("Large image processing failed: %v", result2.Err())
	}
}

// Helper function for absolute difference
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
