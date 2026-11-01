package engine

import (
	"fmt"
	"image"
	"runtime"
	"sync"

	"golang.org/x/image/draw"
)

// Grayscale converts the image to grayscale using direct buffer access.
func (p *Processor) Grayscale() *Processor {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.err != nil {
		return p
	}
	return p.grayscaleDirect()
}

// GrayscaleFast converts the image to grayscale using parallel processing when
// enabled and above the configured threshold.
func (p *Processor) GrayscaleFast() *Processor {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.err != nil {
		return p
	}
	if p.currentImage == nil {
		p.err = fmt.Errorf("no image available to convert to grayscale")
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
	bufferBytes, err := checkedRGBABytes(width, height)
	if err != nil {
		p.err = fmt.Errorf("invalid grayscale image size: %w", err)
		return p
	}

	if shouldUseParallelProcessing(bufferBytes/rgbaBytesPerPixel, p.options) {
		return p.grayscaleParallel()
	}
	return p.grayscaleDirect()
}

func shouldUseParallelProcessing(pixelCount int, opts Options) bool {
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

func (p *Processor) grayscaleBuffers() (srcRGBA, dstRGBA *image.RGBA, width, height int, err error) {
	if p.currentImage == nil {
		return nil, nil, 0, 0, fmt.Errorf("no image available to convert to grayscale")
	}

	bounds := p.currentImage.Bounds()
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

	if rgba, ok := p.currentImage.(*image.RGBA); ok {
		srcRGBA = rgba
	} else {
		srcRGBA, err = newRGBAChecked(bounds)
		if err != nil {
			return nil, nil, 0, 0, fmt.Errorf("invalid grayscale source buffer: %w", err)
		}
		draw.Draw(srcRGBA, bounds, p.currentImage, bounds.Min, draw.Src)
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

func (p *Processor) grayscaleParallel() *Processor {
	srcRGBA, dstRGBA, width, height, err := p.grayscaleBuffers()
	if err != nil {
		p.err = err
		return p
	}
	if width == 0 || height == 0 {
		return p
	}

	numGoroutines := grayscaleWorkerCount(height, p.options.MaxGoroutines)
	if numGoroutines == 0 {
		return p
	}

	var wg sync.WaitGroup
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
	p.currentImage = dstRGBA
	return p
}

func (p *Processor) grayscaleDirect() *Processor {
	srcRGBA, dstRGBA, width, height, err := p.grayscaleBuffers()
	if err != nil {
		p.err = err
		return p
	}
	if width == 0 || height == 0 {
		return p
	}

	for row := 0; row < height; row++ {
		grayscaleRGBArow(srcRGBA, dstRGBA, row, width)
	}

	p.currentImage = dstRGBA
	return p
}
