package gopiq

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"reflect"
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
//
// Methods are individually safe for concurrent use, but a multi-call chain is
// not atomic: calls from other goroutines may interleave between chained
// methods. Images passed to constructors and returned by Image are read-only
// while the processor or its clones may be used concurrently.
type ImageProcessor struct {
	mu           sync.RWMutex // Protects currentImage, err, and perfOpts
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
	Text       string
	FontPath   string // Optional: path to .ttf or .otf font file
	FontBytes  []byte // Optional: raw font bytes (preferred for embedding)
	fontSource watermarkFontSource
	FontSize   float64 // Font size in points
	Color      color.Color
	Position   WatermarkPosition
	OffsetX    float64 // Offset from chosen position
	OffsetY    float64
}

type watermarkFontSource uint8

const (
	watermarkFontSourceUnset watermarkFontSource = iota
	watermarkFontSourceBytes
	watermarkFontSourcePath
)

// defaultWatermarkConfig provides sane defaults.
func defaultWatermarkConfig() *watermarkConfig {
	return &watermarkConfig{
		FontSize:   24,
		Color:      color.RGBA{255, 255, 255, 128}, // White with 50% opacity
		Position:   PositionBottomRight,
		OffsetX:    10,
		OffsetY:    10,
		FontBytes:  goregular.TTF, // Use default Go font if no other font is specified
		fontSource: watermarkFontSourceBytes,
	}
}

// WatermarkOption is a functional option for configuring the watermark.
type WatermarkOption func(*watermarkConfig)

// WithFontPath specifies a font file for the watermark. If both font-source
// options are supplied, the last one takes precedence.
func WithFontPath(path string) WatermarkOption {
	return func(wc *watermarkConfig) {
		wc.FontPath = path
		wc.fontSource = watermarkFontSourcePath
	}
}

// WithFontBytes specifies font data directly (e.g., from an embedded font).
// This is generally preferred for self-contained libraries. If both font-source
// options are supplied, the last one takes precedence.
func WithFontBytes(data []byte) WatermarkOption {
	return func(wc *watermarkConfig) {
		wc.FontBytes = data
		wc.fontSource = watermarkFontSourceBytes
	}
}

// WithFontSize sets the font size for the watermark. The size must be finite,
// positive, and within the renderer's fixed-point range.
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
// Positive offsets move inward from corner anchors; for the center position,
// positive X moves right and positive Y moves down. Negative offsets can move
// text partly outside the image.
func WithOffset(x, y float64) WatermarkOption {
	return func(wc *watermarkConfig) { wc.OffsetX = x; wc.OffsetY = y }
}

// rgbaPool is a sync.Pool for reusing RGBA image buffers to reduce allocations
var rgbaPool = sync.Pool{
	New: func() any {
		// Create a modest-sized RGBA image that can be resized as needed
		return image.NewRGBA(image.Rect(0, 0, 100, 100))
	},
}

const rgbaBytesPerPixel = 4

// checkedImageDimensions returns the extent of bounds without allowing an
// invalid or overflowing rectangle to wrap around.
func checkedImageDimensions(bounds image.Rectangle) (width, height int, err error) {
	width, err = checkedExtent(bounds.Min.X, bounds.Max.X)
	if err != nil {
		return 0, 0, err
	}
	height, err = checkedExtent(bounds.Min.Y, bounds.Max.Y)
	if err != nil {
		return 0, 0, err
	}
	return width, height, nil
}

func checkedExtent(min, max int) (int, error) {
	if max < min {
		return 0, fmt.Errorf("invalid image bounds: maximum %d is less than minimum %d", max, min)
	}

	maxInt := int(^uint(0) >> 1)
	if min < 0 && max > maxInt+min {
		return 0, fmt.Errorf("image bounds extent overflows int")
	}
	return max - min, nil
}

// checkedRGBABytes verifies that an RGBA image with these dimensions can be
// represented by a Go slice and returns its pixel-buffer length.
func checkedRGBABytes(width, height int) (int, error) {
	if width < 0 || height < 0 {
		return 0, fmt.Errorf("image dimensions must not be negative")
	}
	if width == 0 || height == 0 {
		return 0, nil
	}

	maxInt := int(^uint(0) >> 1)
	if width > maxInt/rgbaBytesPerPixel {
		return 0, fmt.Errorf("RGBA image buffer size overflows int")
	}
	rowBytes := width * rgbaBytesPerPixel
	if height > maxInt/rowBytes {
		return 0, fmt.Errorf("RGBA image buffer size overflows int")
	}
	return rowBytes * height, nil
}

func newRGBAChecked(bounds image.Rectangle) (*image.RGBA, error) {
	width, height, err := checkedImageDimensions(bounds)
	if err != nil {
		return nil, err
	}
	if _, err := checkedRGBABytes(width, height); err != nil {
		return nil, err
	}
	return image.NewRGBA(bounds), nil
}

// getPooledRGBA returns an RGBA image from the pool, resized to the given bounds
func getPooledRGBA(bounds image.Rectangle) *image.RGBA {
	width, height, err := checkedImageDimensions(bounds)
	if err != nil {
		return nil
	}
	pixelsNeeded, err := checkedRGBABytes(width, height)
	if err != nil {
		return nil
	}

	img := rgbaPool.Get().(*image.RGBA)

	// Resize the pooled image if needed
	if img.Bounds().Dx() < width || img.Bounds().Dy() < height {
		fresh, allocErr := newRGBAChecked(bounds)
		if allocErr != nil {
			rgbaPool.Put(img)
			return nil
		}
		img = fresh
	} else {
		// Adjust the bounds to match what we need
		img.Rect = bounds
		// Clear the pixel data for the used area
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
	if img == nil {
		return
	}
	width, height, err := checkedImageDimensions(img.Bounds())
	if err != nil {
		return
	}
	bufferBytes, err := checkedRGBABytes(width, height)
	if err != nil {
		return
	}

	// Don't pool very large images to avoid memory waste
	if bufferBytes/rgbaBytesPerPixel <= 2000*2000 {
		rgbaPool.Put(img)
	}
}

// New creates a new ImageProcessor from an existing image.Image. The processor
// retains the supplied image rather than copying it; callers must treat it as
// read-only while the processor or its clones are in use.
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
// The processor retains the supplied image rather than copying it; callers must treat it as
// read-only while the processor or its clones are in use.
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
// This method is safe for concurrent use.
func (ip *ImageProcessor) SetPerformanceOptions(opts PerformanceOptions) *ImageProcessor {
	ip.mu.Lock()
	defer ip.mu.Unlock()
	ip.perfOpts = opts
	return ip
}

// FromBytes creates a new ImageProcessor by decoding an image from a byte slice.
// It supports JPEG, PNG, and GIF formats. Returns an error if decoding fails.
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
// It supports FormatJPEG and FormatPNG; FormatGIF is decode-only. Returns an error if
// encoding fails or if a previous error in the chain exists.
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
// The returned image must be treated as read-only. It may share storage with the image
// passed to New and with images held by clones; callers must not mutate it while the
// processor or its clones are in use. Mutating an image through a concrete-type
// assertion is outside the thread-safety guarantee.
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

// Clone creates a shallow copy of the ImageProcessor. It copies the current
// error and performance options but shares the current image reference, and
// therefore its pixel storage, with the original.
//
// The returned processor has an independent mutex and can be used from another
// goroutine when the shared image is treated as read-only. Processing methods
// do not modify their input; for non-empty images they create a new destination.
// To obtain independent pixels, copy the image before constructing a processor;
// Clone does not deep-copy image data.
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
// Coordinates are in the source image's coordinate space; the result starts at (0, 0).
// Returns the ImageProcessor for chaining. An error is set if the crop area is out of bounds
// or dimensions are invalid or cannot be represented safely.
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
	if ip.currentImage == nil {
		ip.err = fmt.Errorf("no image available to crop")
		return ip
	}

	bounds := ip.currentImage.Bounds()
	if _, _, err := checkedImageDimensions(bounds); err != nil {
		ip.err = fmt.Errorf("invalid source image bounds: %w", err)
		return ip
	}

	maxInt := int(^uint(0) >> 1)
	if x > maxInt-width || y > maxInt-height {
		ip.err = fmt.Errorf("crop rectangle coordinates overflow int")
		return ip
	}
	cropRect := image.Rectangle{
		Min: image.Pt(x, y),
		Max: image.Pt(x+width, y+height),
	}

	if !cropRect.In(bounds) {
		ip.err = fmt.Errorf("crop rectangle %v is out of image bounds %v", cropRect, bounds)
		return ip
	}

	// Create a new RGBA image and draw the cropped portion onto it.
	croppedImg, err := newRGBAChecked(image.Rect(0, 0, width, height))
	if err != nil {
		ip.err = fmt.Errorf("invalid crop image size: %w", err)
		return ip
	}
	draw.Draw(croppedImg, croppedImg.Bounds(), ip.currentImage, cropRect.Min, draw.Src)

	ip.currentImage = croppedImg
	return ip
}

// Resize resizes a non-empty image to the specified width and height using Catmull-Rom interpolation.
// Catmull-Rom provides a good balance of quality and performance among standard library options
// (available in image/draw since Go 1.18).
// Returns the ImageProcessor for chaining. An error is set if dimensions are invalid, the source
// image is empty, or the destination buffer size cannot be represented safely.
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
	if ip.currentImage == nil {
		ip.err = fmt.Errorf("no image available to resize")
		return ip
	}

	originalBounds := ip.currentImage.Bounds()
	originalWidth, originalHeight, err := checkedImageDimensions(originalBounds)
	if err != nil {
		ip.err = fmt.Errorf("invalid source image bounds: %w", err)
		return ip
	}
	if originalWidth == 0 || originalHeight == 0 {
		ip.err = fmt.Errorf("cannot resize an image with empty bounds %v", originalBounds)
		return ip
	}

	dstRect := image.Rect(0, 0, width, height)
	newImg, err := newRGBAChecked(dstRect)
	if err != nil {
		ip.err = fmt.Errorf("invalid resize image size: %w", err)
		return ip
	}

	// Use Catmull-Rom interpolator from image/draw package (standard library)
	draw.CatmullRom.Scale(newImg, dstRect, ip.currentImage, originalBounds, draw.Src, nil)

	ip.currentImage = newImg
	return ip
}

// Grayscale converts the image to grayscale using optimized direct buffer access.
// For maximum performance on large images, consider using GrayscaleFast() instead.
// Empty images are left unchanged. Returns the ImageProcessor for chaining.
// This method is safe for concurrent use.
func (ip *ImageProcessor) Grayscale() *ImageProcessor {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	if ip.err != nil {
		return ip
	}
	return ip.grayscaleDirect()
}

// GrayscaleFast converts the image to grayscale using optimized parallel processing.
// This method is significantly faster than Grayscale() for large images.
// Empty images are left unchanged. Returns the ImageProcessor for chaining.
// This method is safe for concurrent use.
func (ip *ImageProcessor) GrayscaleFast() *ImageProcessor {
	ip.mu.Lock()
	defer ip.mu.Unlock()

	if ip.err != nil {
		return ip
	}
	if ip.currentImage == nil {
		ip.err = fmt.Errorf("no image available to convert to grayscale")
		return ip
	}

	bounds := ip.currentImage.Bounds()
	width, height, err := checkedImageDimensions(bounds)
	if err != nil {
		ip.err = fmt.Errorf("invalid source image bounds: %w", err)
		return ip
	}
	if width == 0 || height == 0 {
		return ip
	}
	bufferBytes, err := checkedRGBABytes(width, height)
	if err != nil {
		ip.err = fmt.Errorf("invalid grayscale image size: %w", err)
		return ip
	}

	if shouldUseParallelProcessing(bufferBytes/rgbaBytesPerPixel, ip.perfOpts) {
		return ip.grayscaleParallel()
	}

	// For small images, use direct buffer access but single-threaded
	return ip.grayscaleDirect()
}

func shouldUseParallelProcessing(pixelCount int, opts PerformanceOptions) bool {
	if !opts.EnableParallelProcessing {
		return false
	}
	threshold := opts.MinSizeForParallel
	if threshold < 0 {
		threshold = 0
	}
	return pixelCount >= threshold
}

func grayscaleWorkerCount(height, maxGoroutines int) int {
	if height <= 0 {
		return 0
	}
	if maxGoroutines <= 0 {
		maxGoroutines = runtime.NumCPU()
	}
	if maxGoroutines < 1 {
		maxGoroutines = 1
	}
	if maxGoroutines > height {
		maxGoroutines = height
	}
	return maxGoroutines
}

func (ip *ImageProcessor) grayscaleBuffers() (srcRGBA, dstRGBA *image.RGBA, width, height int, err error) {
	if ip.currentImage == nil {
		return nil, nil, 0, 0, fmt.Errorf("no image available to convert to grayscale")
	}

	bounds := ip.currentImage.Bounds()
	width, height, err = checkedImageDimensions(bounds)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("invalid source image bounds: %w", err)
	}
	if width == 0 || height == 0 {
		return nil, nil, width, height, nil
	}
	if _, err = checkedRGBABytes(width, height); err != nil {
		return nil, nil, 0, 0, fmt.Errorf("invalid grayscale image size: %w", err)
	}

	if rgba, ok := ip.currentImage.(*image.RGBA); ok {
		srcRGBA = rgba
	} else {
		srcRGBA, err = newRGBAChecked(bounds)
		if err != nil {
			return nil, nil, 0, 0, fmt.Errorf("invalid grayscale source buffer: %w", err)
		}
		draw.Draw(srcRGBA, bounds, ip.currentImage, bounds.Min, draw.Src)
	}

	dstRGBA, err = newRGBAChecked(bounds)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("invalid grayscale destination buffer: %w", err)
	}
	return srcRGBA, dstRGBA, width, height, nil
}

func grayscaleRGBArow(srcRGBA, dstRGBA *image.RGBA, row, width int) {
	srcRowStart := row * srcRGBA.Stride
	dstRowStart := row * dstRGBA.Stride
	for x := 0; x < width; x++ {
		srcIndex := srcRowStart + x*rgbaBytesPerPixel
		dstIndex := dstRowStart + x*rgbaBytesPerPixel

		r := srcRGBA.Pix[srcIndex]
		g := srcRGBA.Pix[srcIndex+1]
		b := srcRGBA.Pix[srcIndex+2]
		alpha := srcRGBA.Pix[srcIndex+3]
		gray := uint8(0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b))

		dstRGBA.Pix[dstIndex] = gray
		dstRGBA.Pix[dstIndex+1] = gray
		dstRGBA.Pix[dstIndex+2] = gray
		dstRGBA.Pix[dstIndex+3] = alpha
	}
}

// grayscaleParallel processes the image using multiple goroutines for better performance.
func (ip *ImageProcessor) grayscaleParallel() *ImageProcessor {
	srcRGBA, dstRGBA, width, height, err := ip.grayscaleBuffers()
	if err != nil {
		ip.err = err
		return ip
	}
	if width == 0 || height == 0 {
		return ip
	}

	numGoroutines := grayscaleWorkerCount(height, ip.perfOpts.MaxGoroutines)
	if numGoroutines == 0 {
		return ip
	}

	var wg sync.WaitGroup

	// Process image in horizontal strips
	rowsPerGoroutine := height / numGoroutines

	for i := range numGoroutines {
		startRow := i * rowsPerGoroutine
		endRow := startRow + rowsPerGoroutine
		if i == numGoroutines-1 {
			endRow = height
		}

		wg.Go(func() {
			for y := startRow; y < endRow; y++ {
				grayscaleRGBArow(srcRGBA, dstRGBA, y, width)
			}
		})
	}

	wg.Wait()
	ip.currentImage = dstRGBA
	return ip
}

// grayscaleDirect processes the image using direct buffer access in a single thread.
func (ip *ImageProcessor) grayscaleDirect() *ImageProcessor {
	srcRGBA, dstRGBA, width, height, err := ip.grayscaleBuffers()
	if err != nil {
		ip.err = err
		return ip
	}
	if width == 0 || height == 0 {
		return ip
	}

	for row := 0; row < height; row++ {
		grayscaleRGBArow(srcRGBA, dstRGBA, row, width)
	}

	ip.currentImage = dstRGBA
	return ip
}

// AddTextWatermark adds a text watermark to the image with anti-aliasing.
// This uses golang.org/x/image/font package for proper font rendering.
// Returns the ImageProcessor for chaining. An error is set if text is empty,
// the image bounds are empty, options are invalid, the font fails to load, or
// drawing fails. Font-source options are applied in order; the last one wins.
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
	if ip.currentImage == nil {
		ip.err = fmt.Errorf("no image available for watermark")
		return ip
	}

	bounds := ip.currentImage.Bounds()
	width, height, err := checkedImageDimensions(bounds)
	if err != nil {
		ip.err = fmt.Errorf("invalid source image bounds: %w", err)
		return ip
	}
	if width == 0 || height == 0 {
		ip.err = fmt.Errorf("cannot add a watermark to an image with empty bounds %v", bounds)
		return ip
	}
	if _, err := checkedRGBABytes(width, height); err != nil {
		ip.err = fmt.Errorf("invalid watermark image size: %w", err)
		return ip
	}

	cfg := defaultWatermarkConfig()
	cfg.Text = text

	for index, opt := range options {
		if opt == nil {
			ip.err = fmt.Errorf("watermark option %d cannot be nil", index)
			return ip
		}
		opt(cfg)
	}
	if err := validateWatermarkConfig(cfg); err != nil {
		ip.err = err
		return ip
	}

	fontBytes := cfg.FontBytes
	if cfg.fontSource == watermarkFontSourcePath {
		fontBytes, err = os.ReadFile(cfg.FontPath)
		if err != nil {
			ip.err = fmt.Errorf("failed to read watermark font %q: %w", cfg.FontPath, err)
			return ip
		}
	}

	fnt, err := opentype.Parse(fontBytes)
	if err != nil {
		if cfg.fontSource == watermarkFontSourcePath {
			ip.err = fmt.Errorf("failed to parse watermark font %q: %w", cfg.FontPath, err)
		} else {
			ip.err = fmt.Errorf("failed to parse font bytes for watermark: %w", err)
		}
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

	metrics := face.Metrics()
	if metrics.Height < 0 || metrics.Ascent < 0 || metrics.Descent < 0 {
		ip.err = fmt.Errorf("watermark font produced invalid metrics")
		return ip
	}

	// Measure the text at a zero-origin baseline. Drawing uses a translated
	// overlay below so image bounds do not leak into the font's fixed-point space.
	measureDrawer := font.Drawer{Face: face}
	textBounds, _ := measureDrawer.BoundString(cfg.Text)
	if textBounds.Max.X < textBounds.Min.X || textBounds.Max.Y < textBounds.Min.Y {
		ip.err = fmt.Errorf("watermark font produced invalid text bounds")
		return ip
	}
	textWidth := float64(textBounds.Max.X)/64 - float64(textBounds.Min.X)/64
	textHeight := float64(metrics.Height) / 64
	ascent := float64(metrics.Ascent) / 64
	descent := float64(metrics.Descent) / 64

	var x, y float64

	switch cfg.Position {
	case PositionTopLeft:
		x = cfg.OffsetX
		y = cfg.OffsetY + ascent
	case PositionTopRight:
		x = float64(width) - textWidth - cfg.OffsetX
		y = cfg.OffsetY + ascent
	case PositionBottomLeft:
		x = cfg.OffsetX
		y = float64(height) - cfg.OffsetY - descent
	case PositionBottomRight:
		x = float64(width) - textWidth - cfg.OffsetX
		y = float64(height) - cfg.OffsetY - descent
	case PositionCenter:
		x = (float64(width)-textWidth)/2 + cfg.OffsetX
		y = (float64(height)-textHeight)/2 + ascent + cfg.OffsetY
	default:
		// validateWatermarkConfig rejects unknown positions before this switch.
		ip.err = fmt.Errorf("unknown watermark position: %d", cfg.Position)
		return ip
	}

	textLeft := x + float64(textBounds.Min.X)/64
	textRight := x + float64(textBounds.Max.X)/64
	textTop := y + float64(textBounds.Min.Y)/64
	textBottom := y + float64(textBounds.Max.Y)/64
	if textRight <= 0 || textLeft >= float64(width) || textBottom <= 0 || textTop >= float64(height) {
		// Drawing outside the destination is clipped, so a fully clipped string
		// leaves the current image unchanged without converting an extreme offset.
		return ip
	}

	xPixel, err := watermarkPixelCoordinate(x)
	if err != nil {
		ip.err = err
		return ip
	}
	yPixel, err := watermarkPixelCoordinate(y)
	if err != nil {
		ip.err = err
		return ip
	}
	xDot, xShift := watermarkFixedCoordinate(xPixel)
	yDot, yShift := watermarkFixedCoordinate(yPixel)

	minX, ok := checkedSubtractInt(0, xShift)
	if !ok {
		ip.err = fmt.Errorf("watermark horizontal position overflows int")
		return ip
	}
	maxX, ok := checkedSubtractInt(width, xShift)
	if !ok {
		ip.err = fmt.Errorf("watermark horizontal position overflows int")
		return ip
	}
	minY, ok := checkedSubtractInt(0, yShift)
	if !ok {
		ip.err = fmt.Errorf("watermark vertical position overflows int")
		return ip
	}
	maxY, ok := checkedSubtractInt(height, yShift)
	if !ok {
		ip.err = fmt.Errorf("watermark vertical position overflows int")
		return ip
	}
	overlayBounds := image.Rectangle{
		Min: image.Pt(minX, minY),
		Max: image.Pt(maxX, maxY),
	}

	imgWithWatermark, err := newRGBAChecked(bounds)
	if err != nil {
		ip.err = fmt.Errorf("invalid watermark image size: %w", err)
		return ip
	}
	overlay, err := newRGBAChecked(overlayBounds)
	if err != nil {
		ip.err = fmt.Errorf("invalid watermark overlay size: %w", err)
		return ip
	}
	draw.Draw(imgWithWatermark, bounds, ip.currentImage, bounds.Min, draw.Src)

	dr := &font.Drawer{
		Dst:  overlay,
		Src:  image.NewUniform(cfg.Color),
		Face: face,
		Dot: fixed.Point26_6{
			X: xDot,
			Y: yDot,
		},
	}
	dr.DrawString(cfg.Text)
	draw.Draw(imgWithWatermark, bounds, overlay, overlayBounds.Min, draw.Over)

	ip.currentImage = imgWithWatermark
	return ip
}

const maxWatermarkFontSize = (float64(1<<31) - 1) / 64

func validateWatermarkConfig(cfg *watermarkConfig) error {
	if cfg.Text == "" {
		return fmt.Errorf("watermark text cannot be empty")
	}
	if math.IsNaN(cfg.FontSize) || math.IsInf(cfg.FontSize, 0) || cfg.FontSize <= 0 {
		return fmt.Errorf("watermark font size must be finite and positive")
	}
	if cfg.FontSize > maxWatermarkFontSize {
		return fmt.Errorf("watermark font size exceeds the renderer's fixed-point range")
	}
	if isNilColor(cfg.Color) {
		return fmt.Errorf("watermark color cannot be nil")
	}
	if cfg.Position < PositionTopLeft || cfg.Position > PositionCenter {
		return fmt.Errorf("unknown watermark position: %d", cfg.Position)
	}
	if math.IsNaN(cfg.OffsetX) || math.IsInf(cfg.OffsetX, 0) || math.IsNaN(cfg.OffsetY) || math.IsInf(cfg.OffsetY, 0) {
		return fmt.Errorf("watermark offsets must be finite")
	}
	switch cfg.fontSource {
	case watermarkFontSourceBytes:
		// Empty or invalid bytes are reported by opentype.Parse with context.
	case watermarkFontSourcePath:
		if cfg.FontPath == "" {
			return fmt.Errorf("watermark font path cannot be empty")
		}
	default:
		return fmt.Errorf("watermark font source is not configured")
	}
	return nil
}

func isNilColor(c color.Color) bool {
	if c == nil {
		return true
	}
	value := reflect.ValueOf(c)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return value.IsNil()
	default:
		return false
	}
}

func watermarkPixelCoordinate(value float64) (int, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("watermark position is not finite")
	}
	value = math.Trunc(value)
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	if value < float64(minInt) || value >= float64(maxInt) {
		return 0, fmt.Errorf("watermark position exceeds the image coordinate range")
	}
	return int(value), nil
}

func watermarkFixedCoordinate(pixel int) (fixed.Int26_6, int) {
	const minFixedPixel = -(1 << 25)
	const maxFixedPixel = (1 << 25) - 1
	if pixel < minFixedPixel || pixel > maxFixedPixel {
		return 0, pixel
	}
	return fixed.I(pixel), 0
}

func checkedSubtractInt(a, b int) (int, bool) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	if b > 0 && a < minInt+b {
		return 0, false
	}
	if b < 0 && a > maxInt+b {
		return 0, false
	}
	return a - b, true
}

// PerformanceOptions controls optimization settings for image processing.
type PerformanceOptions struct {
	// MaxGoroutines limits the number of parallel goroutines for heavy operations.
	// If less than or equal to 0, defaults to runtime.NumCPU().
	MaxGoroutines int
	// EnableParallelProcessing enables parallel processing for suitable operations.
	EnableParallelProcessing bool
	// MinSizeForParallel sets the minimum image size (width * height) before
	// parallel processing is used. Values less than or equal to 0 allow parallel
	// processing for any non-empty image when parallel processing is enabled.
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
