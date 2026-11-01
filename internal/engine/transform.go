package engine

import (
	"fmt"
	"image"
)

// FlipHorizontal mirrors the image left-to-right.
func (p *Processor) FlipHorizontal() *Processor {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.transformLocked(func(bounds image.Rectangle, x, y int) (int, int) {
		return bounds.Max.X - 1 - x, bounds.Min.Y + y
	}, false)
}

// FlipVertical mirrors the image top-to-bottom.
func (p *Processor) FlipVertical() *Processor {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.transformLocked(func(bounds image.Rectangle, x, y int) (int, int) {
		return bounds.Min.X + x, bounds.Max.Y - 1 - y
	}, false)
}

// Rotate90 rotates the image clockwise by 90 degrees.
func (p *Processor) Rotate90() *Processor {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.transformLocked(func(bounds image.Rectangle, x, y int) (int, int) {
		return bounds.Min.X + y, bounds.Max.Y - 1 - x
	}, true)
}

// Rotate180 rotates the image by 180 degrees.
func (p *Processor) Rotate180() *Processor {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.transformLocked(func(bounds image.Rectangle, x, y int) (int, int) {
		return bounds.Max.X - 1 - x, bounds.Max.Y - 1 - y
	}, false)
}

// Rotate270 rotates the image clockwise by 270 degrees.
func (p *Processor) Rotate270() *Processor {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.transformLocked(func(bounds image.Rectangle, x, y int) (int, int) {
		return bounds.Max.X - 1 - y, bounds.Min.Y + x
	}, true)
}

func (p *Processor) transformLocked(source func(image.Rectangle, int, int) (int, int), swapDimensions bool) *Processor {
	if p.err != nil {
		return p
	}
	if p.currentImage == nil {
		p.err = fmt.Errorf("no image available to transform")
		return p
	}

	bounds := p.currentImage.Bounds()
	width, height, err := checkedImageDimensions(bounds)
	if err != nil {
		p.err = fmt.Errorf("invalid source image bounds: %w", err)
		return p
	}
	if width == 0 || height == 0 {
		return p
	}

	dstWidth, dstHeight := width, height
	if swapDimensions {
		dstWidth, dstHeight = height, width
	}
	dst, err := newRGBAChecked(image.Rect(0, 0, dstWidth, dstHeight))
	if err != nil {
		p.err = fmt.Errorf("invalid transform image size: %w", err)
		return p
	}

	for y := 0; y < dstHeight; y++ {
		for x := 0; x < dstWidth; x++ {
			sourceX, sourceY := source(bounds, x, y)
			dst.Set(x, y, p.currentImage.At(sourceX, sourceY))
		}
	}
	p.currentImage = dst
	return p
}
