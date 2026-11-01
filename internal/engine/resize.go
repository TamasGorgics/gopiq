package engine

import (
	"fmt"
	"image"
	"math"

	"golang.org/x/image/draw"
)

// Resize resizes a non-empty image using Catmull-Rom interpolation.
func (p *Processor) Resize(width, height int) *Processor {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.resizeLocked(width, height)
}

// ResizeFit scales a non-empty image to fit inside the requested dimensions.
func (p *Processor) ResizeFit(width, height int) *Processor {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.err != nil {
		return p
	}
	if width <= 0 || height <= 0 {
		p.err = fmt.Errorf("resize dimensions must be positive (width: %d, height: %d)", width, height)
		return p
	}
	if err := p.validateResizeSource(); err != nil {
		p.err = err
		return p
	}

	sourceWidth, sourceHeight, err := checkedImageDimensions(p.currentImage.Bounds())
	if err != nil {
		p.err = fmt.Errorf("invalid source image bounds: %w", err)
		return p
	}
	scale := math.Min(float64(width)/float64(sourceWidth), float64(height)/float64(sourceHeight))
	targetWidth := clampDimension(int(math.Round(float64(sourceWidth)*scale)), width)
	targetHeight := clampDimension(int(math.Round(float64(sourceHeight)*scale)), height)
	return p.resizeLocked(targetWidth, targetHeight)
}

// ResizeFill scales a non-empty image to cover the requested dimensions and
// center-crops the result.
func (p *Processor) ResizeFill(width, height int) *Processor {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.err != nil {
		return p
	}
	if width <= 0 || height <= 0 {
		p.err = fmt.Errorf("resize dimensions must be positive (width: %d, height: %d)", width, height)
		return p
	}
	if err := p.validateResizeSource(); err != nil {
		p.err = err
		return p
	}

	sourceWidth, sourceHeight, err := checkedImageDimensions(p.currentImage.Bounds())
	if err != nil {
		p.err = fmt.Errorf("invalid source image bounds: %w", err)
		return p
	}
	scale := math.Max(float64(width)/float64(sourceWidth), float64(height)/float64(sourceHeight))
	scaledWidth := int(math.Ceil(float64(sourceWidth) * scale))
	scaledHeight := int(math.Ceil(float64(sourceHeight) * scale))
	if scaledWidth < width {
		scaledWidth = width
	}
	if scaledHeight < height {
		scaledHeight = height
	}

	p.resizeLocked(scaledWidth, scaledHeight)
	if p.err != nil {
		return p
	}
	return p.cropLocked((scaledWidth-width)/2, (scaledHeight-height)/2, width, height)
}

func (p *Processor) resizeLocked(width, height int) *Processor {
	if p.err != nil {
		return p
	}
	if width <= 0 || height <= 0 {
		p.err = fmt.Errorf("resize dimensions must be positive (width: %d, height: %d)", width, height)
		return p
	}
	if p.currentImage == nil {
		p.err = fmt.Errorf("no image available to resize")
		return p
	}

	originalBounds := p.currentImage.Bounds()
	originalWidth, originalHeight, err := checkedImageDimensions(originalBounds)
	if err != nil {
		p.err = fmt.Errorf("invalid source image bounds: %w", err)
		return p
	}
	if originalWidth == 0 || originalHeight == 0 {
		p.err = fmt.Errorf("cannot resize an image with empty bounds %v", originalBounds)
		return p
	}

	dstRect := image.Rect(0, 0, width, height)
	newImg, err := newRGBAChecked(dstRect)
	if err != nil {
		p.err = fmt.Errorf("invalid resize image size: %w", err)
		return p
	}

	draw.CatmullRom.Scale(newImg, dstRect, p.currentImage, originalBounds, draw.Src, nil)

	p.currentImage = newImg
	return p
}

func (p *Processor) validateResizeSource() error {
	if p.currentImage == nil {
		return fmt.Errorf("no image available to resize")
	}
	bounds := p.currentImage.Bounds()
	width, height, err := checkedImageDimensions(bounds)
	if err != nil {
		return fmt.Errorf("invalid source image bounds: %w", err)
	}
	if width == 0 || height == 0 {
		return fmt.Errorf("cannot resize an image with empty bounds %v", bounds)
	}
	return nil
}

func clampDimension(value, maximum int) int {
	if value < 1 {
		return 1
	}
	if value > maximum {
		return maximum
	}
	return value
}
