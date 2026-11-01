package gopiq

import "image"

const testRGBABytesPerPixel = 4

func (p *ImageProcessor) testImage() image.Image {
	if p == nil {
		return nil
	}
	img, _ := p.Image()
	return img
}

func (p *ImageProcessor) testOptions() PerformanceOptions {
	if p == nil {
		return PerformanceOptions{}
	}
	return fromEngineOptions(p.processor().Options())
}
