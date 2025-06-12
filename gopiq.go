package gopiq

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"runtime"
	"sync"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular" // A basic font for demonstration
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// ImageProcessor holds the current state of the image being processed
// and any error that occurred during a chainable operation.
// It is safe for concurrent use by multiple goroutines.
type ImageProcessor struct {
	mu           sync.RWMutex // Protects currentImage and err from concurrent access
	currentImage image.Image
	err          error // Stores the first error in a chain
	perfOpts     PerformanceOptions
}

// WatermarkPosition defines common positions for the watermark.
type WatermarkPosition int

const (
	PositionTopLeft WatermarkPosition = iota
	PositionTopRight
	PositionBottomLeft
	PositionBottomRight
	PositionCenter
)

// watermarkConfig holds configuration for adding text watermark.
type watermarkConfig struct {
	Text      string
	FontPath  string  // Optional: path to .ttf or .otf font file
	FontBytes []byte  // Optional: raw font bytes (preferred for embedding)
	FontSize  float64 // Font size in points
	Color     color.Color
	Position  WatermarkPosition
	OffsetX   float64 // Offset from chosen position
	OffsetY   float64
}

// defaultWatermarkConfig provides sane defaults.
func defaultWatermarkConfig() *watermarkConfig {
	return &watermarkConfig{
		FontSize:  24,
		Color:     color.RGBA{255, 255, 255, 128}, // White with 50% opacity
		Position:  PositionBottomRight,
		OffsetX:   10,
		OffsetY:   10,
		FontBytes: goregular.TTF, // Use default Go font if no other font is specified
	}
}

// WatermarkOption is a functional option for configuring the watermark.
type WatermarkOption func(*watermarkConfig)

// WithFontPath specifies the font path for the watermark.
// Use this if the font file is external.
func WithFontPath(path string) WatermarkOption {
	return func(wc *watermarkConfig) { wc.FontPath = path }
}

// WithFontBytes specifies font data directly (e.g., from an embedded font).
// This is generally preferred for self-contained libraries.
func WithFontBytes(data []byte) WatermarkOption {
	return func(wc *watermarkConfig) { wc.FontBytes = data }
}

// WithFontSize sets the font size for the watermark.
func WithFontSize(size float64) WatermarkOption {
	return func(wc *watermarkConfig) { wc.FontSize = size }
}

// WithColor sets the color for the watermark.
func WithColor(c color.Color) WatermarkOption {
	return func(wc *watermarkConfig) { wc.Color = c }
}

// WithPosition sets the position of the watermark.
func WithPosition(pos WatermarkPosition) WatermarkOption {
	return func(wc *watermarkConfig) { wc.Position = pos }
}

// WithOffset sets an additional offset (in pixels) from the chosen position.
// Positive X moves right, positive Y moves down.
func WithOffset(x, y float64) WatermarkOption {
	return func(wc *watermarkConfig) { wc.OffsetX = x; wc.OffsetY = y }
}

// rgbaPool is a sync.Pool for reusing RGBA image buffers to reduce allocations
var rgbaPool = sync.Pool{
	New: func() interface{} {
		// Create a modest-sized RGBA image that can be resized as needed
		return image.NewRGBA(image.Rect(0, 0, 100, 100))
	},
}

// getPooledRGBA returns an RGBA image from the pool, resized to the given bounds
func getPooledRGBA(bounds image.Rectangle) *image.RGBA {
	img := rgbaPool.Get().(*image.RGBA)
	width, height := bounds.Dx(), bounds.Dy()

	// Resize the pooled image if needed
	if img.Bounds().Dx() < width || img.Bounds().Dy() < height {
		img = image.NewRGBA(bounds)
	} else {
		// Adjust the bounds to match what we need
		img.Rect = bounds
		// Clear the pixel data for the used area
		pixelsNeeded := width * height * 4
		if len(img.Pix) < pixelsNeeded {
			img.Pix = make([]uint8, pixelsNeeded)
		} else {
			// Clear only the pixels we'll use
			for i := 0; i < pixelsNeeded; i++ {
				img.Pix[i] = 0
			}
		}
		img.Stride = 4 * width
	}

	return img
}

// returnPooledRGBA returns an RGBA image to the pool for reuse
func returnPooledRGBA(img *image.RGBA) {
	// Don't pool very large images to avoid memory waste
	if img.Bounds().Dx()*img.Bounds().Dy() <= 2000*2000 {
		rgbaPool.Put(img)
	}
}

// New creates a new ImageProcessor from an existing image.Image.
// Returns an error if the provided image is nil.
func New(img image.Image) *ImageProcessor {
	if img == nil {
		return &ImageProcessor{err: fmt.Errorf("initial image cannot be nil")}
	}
	return &ImageProcessor{
		currentImage: img,
		perfOpts:     DefaultPerformanceOptions(),
	}
}

// NewWithPerformanceOptions creates a new ImageProcessor with custom performance settings.
func NewWithPerformanceOptions(img image.Image, opts PerformanceOptions) *ImageProcessor {
	if img == nil {
		return &ImageProcessor{err: fmt.Errorf("initial image cannot be nil")}
	}
	return &ImageProcessor{
		currentImage: img,
		perfOpts:     opts,
	}
}

// SetPerformanceOptions updates the performance settings for this processor.
func (ip *ImageProcessor) SetPerformanceOptions(opts PerformanceOptions) *ImageProcessor {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	ip.perfOpts = opts
	return ip
}

// FromBytes creates a new ImageProcessor by decoding an image from a byte slice.
// It supports JPEG and PNG formats. Returns an error if decoding fails.
func FromBytes(data []byte) *ImageProcessor {
	if len(data) == 0 {
		return &ImageProcessor{err: fmt.Errorf("input byte slice is empty")}
	}
	reader := bytes.NewReader(data)
	img, err := decodeImage(reader)
	if err != nil {
		return &ImageProcessor{err: err}
	}
	return &ImageProcessor{
		currentImage: img,
		perfOpts:     DefaultPerformanceOptions(),
	}
}

// ToBytes converts the current processed image to a byte slice in the specified format.
// Supports FormatJPEG and FormatPNG. Returns an error if encoding fails or if
// a previous error in the chain exists.
// This method is safe for concurrent use.
func (ip *ImageProcessor) ToBytes(format ImageFormat) ([]byte, error) {
	ip.mu.RLock()
	defer ip.mu.RUnlock()

	if ip.err != nil {
		return nil, ip.err
	}
	if ip.currentImage == nil {
		return nil, fmt.Errorf("no image available to convert to bytes")
	}

	var buf bytes.Buffer
	err := encodeImage(&buf, ip.currentImage, format)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image to bytes: %w", err)
	}
	return buf.Bytes(), nil
}

// Image returns the current image.Image and any error encountered in the processing chain.
// This method is safe for concurrent use.
func (ip *ImageProcessor) Image() (image.Image, error) {
	ip.mu.RLock()
	defer ip.mu.RUnlock()
	return ip.currentImage, ip.err
}

// Err returns the first error encountered in the processing chain.
// This method is safe for concurrent use.
func (ip *ImageProcessor) Err() error {
	ip.mu.RLock()
	defer ip.mu.RUnlock()
	return ip.err
}

// Clone creates a deep copy of the ImageProcessor that can be safely used
// in a different goroutine. The returned processor shares no mutable state
// with the original.
func (ip *ImageProcessor) Clone() *ImageProcessor {
	ip.mu.RLock()
	defer ip.mu.RUnlock()

	return &ImageProcessor{
		currentImage: ip.currentImage,
		err:          ip.err,
		perfOpts:     ip.perfOpts, // Copy performance options
	}
}

// --- Image Processing Chainable Methods ---

// Crop crops the image to the specified rectangle defined by x, y, width, and height.
// Returns the ImageProcessor for chaining. An error is set if the crop area is out of bounds
// or dimensions are invalid.
// This method is safe for concurrent use.
func (ip *ImageProcessor) Crop(x, y, width, height int) *ImageProcessor {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	if ip.err != nil {
		return ip
	}
	if width <= 0 || height <= 0 {
		ip.err = fmt.Errorf("crop dimensions must be positive (width: %d, height: %d)", width, height)
		return ip
	}

	bounds := ip.currentImage.Bounds()
	cropRect := image.Rect(x, y, x+width, y+height)

	if !cropRect.In(bounds) {
		ip.err = fmt.Errorf("crop rectangle %v is out of image bounds %v", cropRect, bounds)
		return ip
	}

	// Create a new RGBA image and draw the cropped portion onto it.
	croppedImg := newRGBA(image.Rect(0, 0, width, height))
	draw.Draw(croppedImg, croppedImg.Bounds(), ip.currentImage, cropRect.Min, draw.Src)

	ip.currentImage = croppedImg
	return ip
}

// Resize resizes the image to the specified width and height using Catmull-Rom interpolation.
// Catmull-Rom provides a good balance of quality and performance among standard library options
// (available in image/draw since Go 1.18).
// Returns the ImageProcessor for chaining. An error is set if dimensions are invalid.
// This method is safe for concurrent use.
func (ip *ImageProcessor) Resize(width, height int) *ImageProcessor {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	if ip.err != nil {
		return ip
	}
	if width <= 0 || height <= 0 {
		ip.err = fmt.Errorf("resize dimensions must be positive (width: %d, height: %d)", width, height)
		return ip
	}

	originalBounds := ip.currentImage.Bounds()
	dstRect := image.Rect(0, 0, width, height)
	newImg := newRGBA(dstRect)

	// Use Catmull-Rom interpolator from image/draw package (standard library)
	draw.CatmullRom.Scale(newImg, dstRect, ip.currentImage, originalBounds, draw.Src, nil)

	ip.currentImage = newImg
	return ip
}

// Grayscale converts the image to grayscale using optimized direct buffer access.
// For maximum performance on large images, consider using GrayscaleFast() instead.
// Returns the ImageProcessor for chaining.
// This method is safe for concurrent use.
func (ip *ImageProcessor) Grayscale() *ImageProcessor {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	if ip.err != nil {
		return ip
	}

	bounds := ip.currentImage.Bounds()

	// Convert source to RGBA for direct pixel access
	srcRGBA, ok := ip.currentImage.(*image.RGBA)
	if !ok {
		srcRGBA = image.NewRGBA(bounds)
		draw.Draw(srcRGBA, bounds, ip.currentImage, bounds.Min, draw.Src)
	}

	// Create destination image
	dstRGBA := image.NewRGBA(bounds)
	width, height := bounds.Dx(), bounds.Dy()

	// Process all pixels using direct buffer access (much faster than At/Set)
	for y := 0; y < height; y++ {
		srcRowStart := y * srcRGBA.Stride
		dstRowStart := y * dstRGBA.Stride

		for x := 0; x < width; x++ {
			srcIdx := srcRowStart + x*4
			dstIdx := dstRowStart + x*4

			// Get RGB values directly from buffer
			r := srcRGBA.Pix[srcIdx]
			g := srcRGBA.Pix[srcIdx+1]
			b := srcRGBA.Pix[srcIdx+2]
			a := srcRGBA.Pix[srcIdx+3]

			// Calculate grayscale using luminosity formula (ITU-R BT.709)
			gray := uint8(0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b))

			// Set grayscale value to all RGB channels
			dstRGBA.Pix[dstIdx] = gray   // R
			dstRGBA.Pix[dstIdx+1] = gray // G
			dstRGBA.Pix[dstIdx+2] = gray // B
			dstRGBA.Pix[dstIdx+3] = a    // A (preserve alpha)
		}
	}

	ip.currentImage = dstRGBA
	return ip
}

// GrayscaleFast converts the image to grayscale using optimized parallel processing.
// This method is significantly faster than Grayscale() for large images.
// Returns the ImageProcessor for chaining.
// This method is safe for concurrent use.
func (ip *ImageProcessor) GrayscaleFast() *ImageProcessor {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	if ip.err != nil {
		return ip
	}

	bounds := ip.currentImage.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Use parallel processing for large images
	if ip.perfOpts.EnableParallelProcessing && width*height >= ip.perfOpts.MinSizeForParallel {
		return ip.grayscaleParallel()
	}

	// For small images, use direct buffer access but single-threaded
	return ip.grayscaleDirect()
}

// grayscaleParallel processes the image using multiple goroutines for better performance.
func (ip *ImageProcessor) grayscaleParallel() *ImageProcessor {
	bounds := ip.currentImage.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Convert source to RGBA for direct pixel access
	srcRGBA, ok := ip.currentImage.(*image.RGBA)
	if !ok {
		// Convert to RGBA first
		srcRGBA = image.NewRGBA(bounds)
		draw.Draw(srcRGBA, bounds, ip.currentImage, bounds.Min, draw.Src)
	}

	// Create destination image
	dstRGBA := image.NewRGBA(bounds)

	// Calculate optimal number of goroutines
	numGoroutines := ip.perfOpts.MaxGoroutines
	if numGoroutines <= 0 {
		numGoroutines = runtime.NumCPU()
	}

	// Don't use more goroutines than we have rows
	if numGoroutines > height {
		numGoroutines = height
	}

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Process image in horizontal strips
	rowsPerGoroutine := height / numGoroutines

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			startRow := goroutineID * rowsPerGoroutine
			endRow := startRow + rowsPerGoroutine

			// Last goroutine handles remaining rows
			if goroutineID == numGoroutines-1 {
				endRow = height
			}

			// Process rows assigned to this goroutine
			for y := startRow; y < endRow; y++ {
				rowStart := (y-bounds.Min.Y)*srcRGBA.Stride + (0-bounds.Min.X)*4

				for x := 0; x < width; x++ {
					pixelIdx := rowStart + x*4

					// Get RGB values directly from buffer
					r := srcRGBA.Pix[pixelIdx]
					g := srcRGBA.Pix[pixelIdx+1]
					b := srcRGBA.Pix[pixelIdx+2]
					a := srcRGBA.Pix[pixelIdx+3]

					// Calculate grayscale using luminosity formula (ITU-R BT.709)
					// This is more accurate than simple averaging
					gray := uint8(0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b))

					// Set grayscale value to all RGB channels
					dstRowStart := (y-bounds.Min.Y)*dstRGBA.Stride + (0-bounds.Min.X)*4
					dstPixelIdx := dstRowStart + x*4
					dstRGBA.Pix[dstPixelIdx] = gray   // R
					dstRGBA.Pix[dstPixelIdx+1] = gray // G
					dstRGBA.Pix[dstPixelIdx+2] = gray // B
					dstRGBA.Pix[dstPixelIdx+3] = a    // A (preserve alpha)
				}
			}
		}(i)
	}

	wg.Wait()
	ip.currentImage = dstRGBA
	return ip
}

// grayscaleDirect processes the image using direct buffer access in a single thread.
func (ip *ImageProcessor) grayscaleDirect() *ImageProcessor {
	bounds := ip.currentImage.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Convert source to RGBA for direct pixel access
	srcRGBA, ok := ip.currentImage.(*image.RGBA)
	if !ok {
		srcRGBA = image.NewRGBA(bounds)
		draw.Draw(srcRGBA, bounds, ip.currentImage, bounds.Min, draw.Src)
	}

	// Create destination image
	dstRGBA := image.NewRGBA(bounds)

	// Process all pixels using direct buffer access
	for y := 0; y < height; y++ {
		rowStart := y * srcRGBA.Stride
		dstRowStart := y * dstRGBA.Stride

		for x := 0; x < width; x++ {
			pixelIdx := rowStart + x*4
			dstPixelIdx := dstRowStart + x*4

			// Get RGB values directly from buffer
			r := srcRGBA.Pix[pixelIdx]
			g := srcRGBA.Pix[pixelIdx+1]
			b := srcRGBA.Pix[pixelIdx+2]
			a := srcRGBA.Pix[pixelIdx+3]

			// Calculate grayscale using luminosity formula
			gray := uint8(0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b))

			// Set grayscale value to all RGB channels
			dstRGBA.Pix[dstPixelIdx] = gray   // R
			dstRGBA.Pix[dstPixelIdx+1] = gray // G
			dstRGBA.Pix[dstPixelIdx+2] = gray // B
			dstRGBA.Pix[dstPixelIdx+3] = a    // A (preserve alpha)
		}
	}

	ip.currentImage = dstRGBA
	return ip
}

// AddTextWatermark adds a text watermark to the image with anti-aliasing.
// This uses golang.org/x/image/font package for proper font rendering.
// Returns the ImageProcessor for chaining. An error is set if text is empty,
// font fails to load, or drawing fails.
// This method is safe for concurrent use.
func (ip *ImageProcessor) AddTextWatermark(text string, options ...WatermarkOption) *ImageProcessor {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	if ip.err != nil {
		return ip
	}
	if text == "" {
		ip.err = fmt.Errorf("watermark text cannot be empty")
		return ip
	}

	cfg := defaultWatermarkConfig()
	cfg.Text = text

	for _, opt := range options {
		opt(cfg)
	}

	// Load font
	fnt, err := opentype.Parse(cfg.FontBytes)
	if err != nil {
		ip.err = fmt.Errorf("failed to parse font bytes for watermark: %w", err)
		return ip
	}

	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    cfg.FontSize,
		DPI:     72, // Standard DPI
		Hinting: font.HintingNone,
	})
	if err != nil {
		ip.err = fmt.Errorf("failed to create font face for watermark: %w", err)
		return ip
	}
	defer face.Close()

	// Create a new RGBA image to draw on to avoid modifying the original directly
	bounds := ip.currentImage.Bounds()
	imgWithWatermark := newRGBA(bounds)
	draw.Draw(imgWithWatermark, bounds, ip.currentImage, bounds.Min, draw.Src) // Copy original image

	dr := &font.Drawer{
		Dst:  imgWithWatermark,
		Src:  image.NewUniform(cfg.Color),
		Face: face,
	}

	// Measure text bounds and position
	textBounds, _ := dr.BoundString(cfg.Text)                    // Bounds of the text if drawn at (0,0)
	textWidth := float64(textBounds.Max.X-textBounds.Min.X) / 64 // Convert fixed.Int26_6 to float64 pixels
	textHeight := float64(face.Metrics().Height) / 64            // Ascent + descent in pixels

	var x, y float64

	switch cfg.Position {
	case PositionTopLeft:
		x = cfg.OffsetX
		y = cfg.OffsetY + (float64(face.Metrics().Ascent) / 64) // Adjust for baseline
	case PositionTopRight:
		x = float64(bounds.Dx()) - textWidth - cfg.OffsetX
		y = cfg.OffsetY + (float64(face.Metrics().Ascent) / 64)
	case PositionBottomLeft:
		x = cfg.OffsetX
		y = float64(bounds.Dy()) - cfg.OffsetY - (float64(face.Metrics().Descent) / 64) // Adjust for baseline
	case PositionBottomRight:
		x = float64(bounds.Dx()) - textWidth - cfg.OffsetX
		y = float64(bounds.Dy()) - cfg.OffsetY - (float64(face.Metrics().Descent) / 64)
	case PositionCenter:
		x = (float64(bounds.Dx()) - textWidth) / 2
		y = (float64(bounds.Dy())-textHeight)/2 + (float64(face.Metrics().Ascent) / 64) // Center of block + ascent
	}

	dr.Dot = fixed.Point26_6{
		X: fixed.I(int(x)),
		Y: fixed.I(int(y)),
	}

	dr.DrawString(cfg.Text)

	ip.currentImage = imgWithWatermark
	return ip
}

// PerformanceOptions controls optimization settings for image processing.
type PerformanceOptions struct {
	// MaxGoroutines limits the number of parallel goroutines for heavy operations.
	// If 0, defaults to runtime.NumCPU().
	MaxGoroutines int
	// EnableParallelProcessing enables parallel processing for suitable operations.
	EnableParallelProcessing bool
	// MinSizeForParallel sets the minimum image size (width * height) before
	// parallel processing is used. Smaller images process faster sequentially.
	MinSizeForParallel int
}

// DefaultPerformanceOptions returns optimized defaults for most use cases.
func DefaultPerformanceOptions() PerformanceOptions {
	return PerformanceOptions{
		MaxGoroutines:            runtime.NumCPU(),
		EnableParallelProcessing: true,
		MinSizeForParallel:       10000, // 100x100 pixels
	}
}
