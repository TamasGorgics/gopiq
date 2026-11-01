package engine

import (
	"fmt"
	"image"

	"golang.org/x/image/draw"
)

// Resize resizes a non-empty image using Catmull-Rom interpolation.
func (p *Processor) Resize(width, height int) *Processor {
	p.mu.Lock()
	defer p.mu.Unlock()

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
