package engine

import (
	"fmt"
	"image"

	"golang.org/x/image/draw"
)

// Crop crops the image to the specified source-coordinate rectangle.
func (p *Processor) Crop(x, y, width, height int) *Processor {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.cropLocked(x, y, width, height)
}

func (p *Processor) cropLocked(x, y, width, height int) *Processor {
	if p.err != nil {
		return p
	}
	if width <= 0 || height <= 0 {
		p.err = fmt.Errorf("crop dimensions must be positive (width: %d, height: %d)", width, height)
		return p
	}
	if p.currentImage == nil {
		p.err = fmt.Errorf("no image available to crop")
		return p
	}

	bounds := p.currentImage.Bounds()
	if _, _, err := checkedImageDimensions(bounds); err != nil {
		p.err = fmt.Errorf("invalid source image bounds: %w", err)
		return p
	}

	maxInt := int(^uint(0) >> 1)
	if x > maxInt-width || y > maxInt-height {
		p.err = fmt.Errorf("crop rectangle coordinates overflow int")
		return p
	}
	cropRect := image.Rectangle{
		Min: image.Pt(x, y),
		Max: image.Pt(x+width, y+height),
	}

	if !cropRect.In(bounds) {
		p.err = fmt.Errorf("crop rectangle %v is out of image bounds %v", cropRect, bounds)
		return p
	}

	croppedImg, err := newRGBAChecked(image.Rect(0, 0, width, height))
	if err != nil {
		p.err = fmt.Errorf("invalid crop image size: %w", err)
		return p
	}
	draw.Draw(croppedImg, croppedImg.Bounds(), p.currentImage, cropRect.Min, draw.Src)

	p.currentImage = croppedImg
	return p
}
