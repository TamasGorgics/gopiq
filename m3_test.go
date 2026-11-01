package gopiq

import (
	"image"
	"image/draw"
	"sync"
	"testing"
)

func TestM3OperationsDoNotMutateSource(t *testing.T) {
	sourceTypes := []struct {
		name   string
		create func(image.Rectangle) draw.Image
	}{
		{name: "rgba", create: func(bounds image.Rectangle) draw.Image { return image.NewRGBA(bounds) }},
		{name: "nrgba", create: func(bounds image.Rectangle) draw.Image { return image.NewNRGBA(bounds) }},
	}
	operations := []struct {
		name  string
		apply func(*ImageProcessor)
	}{
		{name: "crop", apply: func(processor *ImageProcessor) {
			processor.Crop(11, -7, 20, 15)
		}},
		{name: "resize", apply: func(processor *ImageProcessor) {
			processor.Resize(30, 20)
		}},
		{name: "grayscale", apply: func(processor *ImageProcessor) {
			processor.Grayscale()
		}},
		{name: "grayscale fast", apply: func(processor *ImageProcessor) {
			processor.GrayscaleFast()
		}},
		{name: "watermark", apply: func(processor *ImageProcessor) {
			processor.AddTextWatermark("M3", WithFontSize(12))
		}},
	}

	for _, sourceType := range sourceTypes {
		for _, operation := range operations {
			t.Run(sourceType.name+"/"+operation.name, func(t *testing.T) {
				source := sourceType.create(image.Rect(11, -7, 71, 43))
				fillM1Image(source)
				before := captureM1Pixels(source)

				processor := NewWithPerformanceOptions(source, PerformanceOptions{
					MaxGoroutines:            2,
					EnableParallelProcessing: true,
					MinSizeForParallel:       0,
				})
				operation.apply(processor)
				if err := processor.Err(); err != nil {
					t.Fatalf("%s failed: %v", operation.name, err)
				}
				if !sameM1Pixels(before, captureM1Pixels(source)) {
					t.Fatalf("%s modified its source image", operation.name)
				}
				if processor.testImage() == image.Image(source) {
					t.Fatalf("%s reused its source image as the destination", operation.name)
				}
			})
		}
	}
}

func TestM3CloneBranchesShareReadOnlySource(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 100, 80))
	fillM1Image(source)
	before := captureM1Pixels(source)

	original := New(source)
	cloned := original.Clone()

	var wg sync.WaitGroup
	wg.Go(func() {
		original.Resize(50, 40).Grayscale()
	})
	wg.Go(func() {
		cloned.Crop(10, 10, 30, 20).AddTextWatermark("clone", WithFontSize(10))
	})
	wg.Wait()

	if err := original.Err(); err != nil {
		t.Fatalf("original branch failed: %v", err)
	}
	if err := cloned.Err(); err != nil {
		t.Fatalf("clone branch failed: %v", err)
	}
	if !sameM1Pixels(before, captureM1Pixels(source)) {
		t.Fatal("concurrent clone branches modified their shared source image")
	}
	if original.testImage() == cloned.testImage() {
		t.Fatal("clone branches should hold independent current images")
	}
	if original.testImage() == image.Image(source) || cloned.testImage() == image.Image(source) {
		t.Fatal("clone branch operation reused the shared source as its destination")
	}
}
