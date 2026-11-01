// Package gopiq provides a fluent, thread-safe image processing API.
//
// The exported types and functions in this package are the supported public
// API. Implementation details live under internal/engine and are not part of
// the public contract.
package gopiq

import (
	"fmt"
	"image"
	"io"
	"sync"

	"github.com/TamasGorgics/gopiq/internal/engine"
)

// ImageProcessor holds the current state of the image being processed
// and any error that occurred during a chainable operation.
//
// Methods are individually safe for concurrent use, but a multi-call chain is
// not atomic: calls from other goroutines may interleave between chained
// methods. Images passed to constructors and returned by Image are read-only
// while the processor or its clones may be used concurrently.
//
// The zero value is an empty processor. Prefer New, NewWithPerformanceOptions,
// or FromBytes to initialize one with an image.
type ImageProcessor struct {
	once sync.Once
	core *engine.Processor
}

func (p *ImageProcessor) processor() *engine.Processor {
	if p == nil {
		return nil
	}
	p.once.Do(func() {
		if p.core == nil {
			p.core = engine.NewUninitialized(engine.DefaultOptions())
		}
	})
	return p.core
}

// New creates a new ImageProcessor from an existing image.Image. The processor
// retains the supplied image rather than copying it; callers must treat it as
// read-only while the processor or its clones are in use.
// Returns an error if the provided image is nil.
func New(img image.Image) *ImageProcessor {
	return &ImageProcessor{
		core: engine.New(img, engine.DefaultOptions()),
	}
}

// NewWithPerformanceOptions creates a new ImageProcessor with custom performance settings.
// The processor retains the supplied image rather than copying it; callers must treat it as
// read-only while the processor or its clones are in use.
func NewWithPerformanceOptions(img image.Image, opts PerformanceOptions) *ImageProcessor {
	return &ImageProcessor{
		core: engine.New(img, toEngineOptions(opts)),
	}
}

// SetPerformanceOptions updates the performance settings for this processor.
// This method is safe for concurrent use.
func (p *ImageProcessor) SetPerformanceOptions(opts PerformanceOptions) *ImageProcessor {
	if core := p.processor(); core != nil {
		core.SetOptions(toEngineOptions(opts))
	}
	return p
}

// FromBytes creates a new ImageProcessor by decoding an image from a byte slice.
// It supports JPEG, PNG, and GIF formats. Returns an error if decoding fails.
// FromBytes retains legacy unlimited decoding; use FromBytesWithLimits for
// untrusted or size-constrained input.
func FromBytes(data []byte) *ImageProcessor {
	return &ImageProcessor{
		core: engine.FromBytes(data, engine.DefaultOptions()),
	}
}

// ToBytes converts the current processed image to a byte slice in the specified format.
// It supports FormatJPEG, FormatPNG, and FormatGIF. GIF encoding is lossy and
// uses the documented standard-library palette policy. Returns an error if encoding
// fails or if a previous error in the chain exists.
// This method is safe for concurrent use.
func (p *ImageProcessor) ToBytes(format ImageFormat) ([]byte, error) {
	core := p.processor()
	if core == nil {
		return nil, fmt.Errorf("no image available to convert to bytes")
	}
	return core.ToBytes(toEngineFormat(format))
}

// ToWriter encodes the current image directly to w in the specified format.
// It supports FormatJPEG, FormatPNG, and FormatGIF. GIF encoding is lossy and
// uses the documented standard-library palette policy. The writer is not closed.
// A write failure may leave partially written output; use ToBytes if an
// in-memory result is required.
// This method is safe for concurrent use.
func (p *ImageProcessor) ToWriter(w io.Writer, format ImageFormat) error {
	core := p.processor()
	if core == nil {
		return fmt.Errorf("no image available to encode")
	}
	return core.ToWriter(w, toEngineFormat(format))
}

// Image returns the current image.Image and any error encountered in the processing chain.
// The returned image must be treated as read-only. It may share storage with the image
// passed to New and with images held by clones; callers must not mutate it while the
// processor or its clones are in use. Mutating an image through a concrete-type
// assertion is outside the thread-safety guarantee.
// This method is safe for concurrent use.
//
//nolint:nilnil // The zero-value processor intentionally has no image and no error.
func (p *ImageProcessor) Image() (image.Image, error) {
	core := p.processor()
	if core == nil {
		return nil, nil
	}
	return core.Image()
}

// Err returns the first error encountered in the processing chain.
// This method is safe for concurrent use.
func (p *ImageProcessor) Err() error {
	core := p.processor()
	if core == nil {
		return nil
	}
	return core.Err()
}

// Clone creates a shallow copy of the ImageProcessor. It copies the current
// error and performance options but shares the current image reference, and
// therefore its pixel storage, with the original.
//
// The returned processor has independent processor state and can be used from
// another goroutine when the shared image is treated as read-only. Processing
// methods do not modify their input; for non-empty images they create a new
// destination. To obtain independent pixels, copy the image before constructing
// a processor; Clone does not deep-copy image data.
func (p *ImageProcessor) Clone() *ImageProcessor {
	if p == nil {
		return &ImageProcessor{}
	}
	return &ImageProcessor{
		core: p.processor().Clone(),
	}
}

// Crop crops the image to the specified rectangle defined by x, y, width, and height.
// Coordinates are in the source image's coordinate space; the result starts at (0, 0).
// Returns the ImageProcessor for chaining. An error is set if the crop area is out of bounds
// or dimensions are invalid or cannot be represented safely.
// This method is safe for concurrent use.
func (p *ImageProcessor) Crop(x, y, width, height int) *ImageProcessor {
	if core := p.processor(); core != nil {
		core.Crop(x, y, width, height)
	}
	return p
}

// Resize resizes a non-empty image to the specified width and height using Catmull-Rom interpolation.
// Catmull-Rom provides a good balance of quality and performance among standard library options
// (available in image/draw since Go 1.18).
// Returns the ImageProcessor for chaining. An error is set if dimensions are invalid, the source
// image is empty, or the destination buffer size cannot be represented safely.
// This method is safe for concurrent use.
func (p *ImageProcessor) Resize(width, height int) *ImageProcessor {
	if core := p.processor(); core != nil {
		core.Resize(width, height)
	}
	return p
}

// ResizeFit scales a non-empty image to fit inside the requested dimensions
// while preserving its aspect ratio. The result may be smaller than the target
// box in one dimension; output bounds start at (0, 0). Empty images and invalid
// dimensions set a chain error. Upscaling is allowed.
// This method is safe for concurrent use.
func (p *ImageProcessor) ResizeFit(width, height int) *ImageProcessor {
	if core := p.processor(); core != nil {
		core.ResizeFit(width, height)
	}
	return p
}

// ResizeFill scales a non-empty image to cover the requested dimensions while
// preserving its aspect ratio, then center-crops to exactly width×height.
// Output bounds start at (0, 0). Empty images and invalid dimensions set a chain
// error. Upscaling is allowed.
// This method is safe for concurrent use.
func (p *ImageProcessor) ResizeFill(width, height int) *ImageProcessor {
	if core := p.processor(); core != nil {
		core.ResizeFill(width, height)
	}
	return p
}

// FlipHorizontal mirrors the image left-to-right. Output starts at (0, 0) for
// non-empty images; empty images are left unchanged.
// This method is safe for concurrent use.
func (p *ImageProcessor) FlipHorizontal() *ImageProcessor {
	if core := p.processor(); core != nil {
		core.FlipHorizontal()
	}
	return p
}

// FlipVertical mirrors the image top-to-bottom. Output starts at (0, 0) for
// non-empty images; empty images are left unchanged.
// This method is safe for concurrent use.
func (p *ImageProcessor) FlipVertical() *ImageProcessor {
	if core := p.processor(); core != nil {
		core.FlipVertical()
	}
	return p
}

// Rotate90 rotates the image clockwise by 90 degrees. Width and height are
// swapped and output starts at (0, 0) for non-empty images; empty images are
// left unchanged.
// This method is safe for concurrent use.
func (p *ImageProcessor) Rotate90() *ImageProcessor {
	if core := p.processor(); core != nil {
		core.Rotate90()
	}
	return p
}

// Rotate180 rotates the image by 180 degrees. Output starts at (0, 0) for
// non-empty images; empty images are left unchanged.
// This method is safe for concurrent use.
func (p *ImageProcessor) Rotate180() *ImageProcessor {
	if core := p.processor(); core != nil {
		core.Rotate180()
	}
	return p
}

// Rotate270 rotates the image clockwise by 270 degrees. Width and height are
// swapped and output starts at (0, 0) for non-empty images; empty images are
// left unchanged.
// This method is safe for concurrent use.
func (p *ImageProcessor) Rotate270() *ImageProcessor {
	if core := p.processor(); core != nil {
		core.Rotate270()
	}
	return p
}

// Grayscale converts the image to grayscale using optimized direct buffer access.
// For maximum performance on large images, consider using GrayscaleFast() instead.
// Empty images are left unchanged. Returns the ImageProcessor for chaining.
// This method is safe for concurrent use.
func (p *ImageProcessor) Grayscale() *ImageProcessor {
	if core := p.processor(); core != nil {
		core.Grayscale()
	}
	return p
}

// GrayscaleFast converts the image to grayscale using optimized parallel processing.
// This method is significantly faster than Grayscale() for large images.
// Empty images are left unchanged. Returns the ImageProcessor for chaining.
// This method is safe for concurrent use.
func (p *ImageProcessor) GrayscaleFast() *ImageProcessor {
	if core := p.processor(); core != nil {
		core.GrayscaleFast()
	}
	return p
}

// AddTextWatermark adds a text watermark to the image with anti-aliasing.
// This uses golang.org/x/image/font package for proper font rendering.
// Returns the ImageProcessor for chaining. An error is set if text is empty,
// the image bounds are empty, options are invalid, the font fails to load, or
// drawing fails. Font-source options are applied in order; the last one wins.
// This method is safe for concurrent use.
func (p *ImageProcessor) AddTextWatermark(text string, options ...WatermarkOption) *ImageProcessor {
	core := p.processor()
	if core == nil {
		return p
	}

	cfg := defaultWatermarkConfig()
	var optionErr error
	for index, opt := range options {
		if opt == nil {
			optionErr = fmt.Errorf("watermark option %d cannot be nil", index)
			break
		}
		opt(cfg)
	}

	core.AddTextWatermark(text, toEngineWatermarkConfig(cfg), optionErr)
	return p
}
