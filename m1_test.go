package gopiq

import (
	"image"
	"image/color"
	"image/draw"
	"runtime"
	"sync"
	"testing"
)

func TestM1CropUsesSourceCoordinatesAndChecksOverflow(t *testing.T) {
	source := image.NewRGBA(image.Rect(10, 20, 16, 26))
	fillM1Image(source)
	before := captureM1Pixels(source)

	processor := New(source).Crop(11, 21, 3, 2)
	if err := processor.Err(); err != nil {
		t.Fatalf("Crop on non-zero-origin image failed: %v", err)
	}
	if got, want := processor.currentImage.Bounds(), image.Rect(0, 0, 3, 2); got != want {
		t.Fatalf("cropped bounds = %v, want %v", got, want)
	}
	if !sameM1Color(source.At(11, 21), processor.currentImage.At(0, 0)) {
		t.Fatal("cropped image did not preserve the source pixel at the crop origin")
	}
	if !sameM1Pixels(before, captureM1Pixels(source)) {
		t.Fatal("Crop modified its source image")
	}

	maxInt := int(^uint(0) >> 1)
	for _, tc := range []struct {
		name string
		x, y int
		w, h int
	}{
		{name: "x overflow", x: maxInt, y: 0, w: 1, h: 1},
		{name: "y overflow", x: 0, y: maxInt, w: 1, h: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := New(createTestImage(2, 2)).Crop(tc.x, tc.y, tc.w, tc.h).Err(); err == nil {
				t.Fatal("Crop accepted overflowing rectangle coordinates")
			}
		})
	}

	chain := New(createTestImage(2, 2)).Crop(maxInt, 0, 1, 1)
	firstErr := chain.Err()
	chain.Resize(1, 1).Grayscale()
	if chain.Err() != firstErr {
		t.Fatalf("later operations replaced the first error: got %v, want %v", chain.Err(), firstErr)
	}
}

func TestM1ImageOperationsRejectUnrepresentableBuffers(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	width := maxInt/rgbaBytesPerPixel + 1
	source := m1BoundsImage{bounds: image.Rect(0, 0, width, 1)}

	if err := New(source).Crop(0, 0, width, 1).Err(); err == nil {
		t.Fatal("Crop accepted a destination whose RGBA buffer size overflows int")
	}
	if err := New(source).Grayscale().Err(); err == nil {
		t.Fatal("Grayscale accepted a destination whose RGBA buffer size overflows int")
	}
	if err := NewWithPerformanceOptions(source, PerformanceOptions{
		EnableParallelProcessing: true,
		MinSizeForParallel:       0,
	}).GrayscaleFast().Err(); err == nil {
		t.Fatal("GrayscaleFast accepted a destination whose RGBA buffer size overflows int")
	}
	if err := New(source).AddTextWatermark("too large").Err(); err == nil {
		t.Fatal("AddTextWatermark accepted a destination whose RGBA buffer size overflows int")
	}
}

func TestM1ResizeChecksEmptyAndUnrepresentableSizes(t *testing.T) {
	empty := image.NewRGBA(image.Rect(4, 7, 4, 7))
	if err := New(empty).Resize(2, 2).Err(); err == nil {
		t.Fatal("Resize accepted an empty source image")
	}

	maxInt := int(^uint(0) >> 1)
	for _, tc := range []struct {
		name          string
		width, height int
	}{
		{name: "row overflow", width: maxInt, height: 1},
		{name: "buffer overflow", width: 1, height: maxInt},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := New(image.NewRGBA(image.Rect(0, 0, 1, 1))).Resize(tc.width, tc.height).Err(); err == nil {
				t.Fatal("Resize accepted an unrepresentable RGBA buffer size")
			}
		})
	}
}

func TestM1ResizeSupportsNonZeroBoundsAndPreservesSource(t *testing.T) {
	source := image.NewRGBA(image.Rect(-5, 9, 2, 15))
	fillM1Image(source)
	before := captureM1Pixels(source)

	processor := New(source).Resize(9, 4)
	if err := processor.Err(); err != nil {
		t.Fatalf("Resize on non-zero-origin image failed: %v", err)
	}
	if got, want := processor.currentImage.Bounds(), image.Rect(0, 0, 9, 4); got != want {
		t.Fatalf("resized bounds = %v, want %v", got, want)
	}
	if !sameM1Pixels(before, captureM1Pixels(source)) {
		t.Fatal("Resize modified its source image")
	}
}

func TestM1GrayscaleSupportsNonZeroBoundsAndSubimages(t *testing.T) {
	nonZeroOrigin := image.NewRGBA(image.Rect(12, -7, 20, 0))
	fillM1Image(nonZeroOrigin)

	backing := image.NewRGBA(image.Rect(0, 0, 23, 17))
	fillM1Image(backing)
	subImage := backing.SubImage(image.Rect(4, 3, 13, 10)).(*image.RGBA)

	nonRGBA := image.NewNRGBA(image.Rect(-8, 11, 0, 18))
	fillM1Image(nonRGBA)
	oneRow := image.NewRGBA(image.Rect(5, -1, 9, 0))
	fillM1Image(oneRow)

	for _, tc := range []struct {
		name string
		image.Image
	}{
		{name: "non-zero-origin RGBA", Image: nonZeroOrigin},
		{name: "RGBA subimage with padded stride", Image: subImage},
		{name: "non-RGBA image", Image: nonRGBA},
		{name: "single-row image", Image: oneRow},
	} {
		t.Run(tc.name, func(t *testing.T) {
			checkM1GrayscalePaths(t, tc.Image)
		})
	}
}

func TestM1GrayscaleEmptyImagesAreNoOps(t *testing.T) {
	options := PerformanceOptions{
		EnableParallelProcessing: true,
		MinSizeForParallel:       0,
	}
	for _, bounds := range []image.Rectangle{
		image.Rect(10, 20, 10, 20),
		image.Rect(10, 20, 10, 25),
		image.Rect(10, 20, 15, 20),
	} {
		t.Run(bounds.String(), func(t *testing.T) {
			source := image.NewRGBA(bounds)
			for name, processor := range map[string]*ImageProcessor{
				"sequential": New(source).Grayscale(),
				"fast":       NewWithPerformanceOptions(source, options).GrayscaleFast(),
			} {
				if err := processor.Err(); err != nil {
					t.Errorf("%s grayscale returned an error: %v", name, err)
					continue
				}
				if processor.currentImage != source {
					t.Errorf("%s grayscale replaced an empty image instead of leaving it unchanged", name)
				}
				if got := processor.currentImage.Bounds(); got != bounds {
					t.Errorf("%s grayscale bounds = %v, want %v", name, got, bounds)
				}
			}
		})
	}
}

func TestM1EmptyImageOperations(t *testing.T) {
	source := image.NewRGBA(image.Rect(2, 3, 2, 3))
	if err := New(source).Crop(2, 3, 1, 1).Err(); err == nil {
		t.Fatal("Crop accepted an empty source image")
	}
	if err := New(source).Resize(1, 1).Err(); err == nil {
		t.Fatal("Resize accepted an empty source image")
	}
	if err := New(source).AddTextWatermark("empty image").Err(); err == nil {
		t.Fatal("AddTextWatermark accepted an empty destination image")
	}
}

func TestM1InvalidImageBoundsReturnErrors(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	source := m1BoundsImage{
		bounds: image.Rectangle{
			Min: image.Pt(minInt, 0),
			Max: image.Pt(maxInt, 1),
		},
	}

	if err := New(source).Grayscale().Err(); err == nil {
		t.Fatal("Grayscale accepted source bounds whose extent overflows int")
	}
	if err := New(source).Resize(1, 1).Err(); err == nil {
		t.Fatal("Resize accepted source bounds whose extent overflows int")
	}
	if err := New(source).Crop(0, 0, 1, 1).Err(); err == nil {
		t.Fatal("Crop accepted source bounds whose extent overflows int")
	}
}

func TestM1PerformanceOptionsCanChangeConcurrentlyWithGrayscale(t *testing.T) {
	processor := New(createTestImage(40, 35))
	options := []PerformanceOptions{
		{MaxGoroutines: 0, EnableParallelProcessing: true, MinSizeForParallel: 0},
		{MaxGoroutines: 2, EnableParallelProcessing: true, MinSizeForParallel: -1},
		{EnableParallelProcessing: false, MinSizeForParallel: 0},
	}

	const goroutines = 6
	const iterations = 12
	start := make(chan struct{})
	var wg sync.WaitGroup
	for index := range goroutines {
		wg.Go(func() {
			<-start
			for j := 0; j < iterations; j++ {
				if index%2 == 0 {
					processor.SetPerformanceOptions(options[(index+j)%len(options)])
					continue
				}
				processor.GrayscaleFast()
			}
		})
	}
	close(start)
	wg.Wait()

	if err := processor.Err(); err != nil {
		t.Fatalf("concurrent performance-option changes caused grayscale to fail: %v", err)
	}
}

func TestM1PerformanceThresholdAndWorkerLimits(t *testing.T) {
	defaultOptions := DefaultPerformanceOptions()
	if shouldUseParallelProcessing(9_999, defaultOptions) {
		t.Fatal("default options selected parallel processing below their threshold")
	}
	if !shouldUseParallelProcessing(10_000, defaultOptions) {
		t.Fatal("default options did not select parallel processing at their threshold")
	}
	if !shouldUseParallelProcessing(1, PerformanceOptions{EnableParallelProcessing: true, MinSizeForParallel: 0}) {
		t.Fatal("zero threshold should allow parallel processing for a non-empty image")
	}
	if !shouldUseParallelProcessing(1, PerformanceOptions{EnableParallelProcessing: true, MinSizeForParallel: -1}) {
		t.Fatal("negative threshold should allow parallel processing for a non-empty image")
	}
	if shouldUseParallelProcessing(100, PerformanceOptions{EnableParallelProcessing: false, MinSizeForParallel: 0}) {
		t.Fatal("disabled parallel processing selected the parallel path")
	}

	if got := grayscaleWorkerCount(0, 4); got != 0 {
		t.Fatalf("worker count for empty image = %d, want 0", got)
	}
	if got := grayscaleWorkerCount(1, 8); got != 1 {
		t.Fatalf("worker count for one row = %d, want 1", got)
	}
	if got := grayscaleWorkerCount(8, 3); got != 3 {
		t.Fatalf("worker count with explicit limit = %d, want 3", got)
	}
	defaultWorkers := runtime.NumCPU()
	if defaultWorkers > 8 {
		defaultWorkers = 8
	}
	if got := grayscaleWorkerCount(8, 0); got != defaultWorkers {
		t.Fatalf("worker count for non-positive limit = %d, want %d", got, defaultWorkers)
	}
}

func checkM1GrayscalePaths(t *testing.T, source image.Image) {
	t.Helper()
	before := captureM1Pixels(source)
	standard := New(source).Grayscale()
	fast := NewWithPerformanceOptions(source, PerformanceOptions{
		MaxGoroutines:            -1,
		EnableParallelProcessing: true,
		MinSizeForParallel:       0,
	}).GrayscaleFast()
	if err := standard.Err(); err != nil {
		t.Fatalf("Grayscale failed: %v", err)
	}
	if err := fast.Err(); err != nil {
		t.Fatalf("GrayscaleFast failed: %v", err)
	}
	if got, want := standard.currentImage.Bounds(), source.Bounds(); got != want {
		t.Fatalf("Grayscale bounds = %v, want source bounds %v", got, want)
	}
	if got, want := fast.currentImage.Bounds(), source.Bounds(); got != want {
		t.Fatalf("GrayscaleFast bounds = %v, want source bounds %v", got, want)
	}

	bounds := source.Bounds()
	normalized := image.NewRGBA(bounds)
	draw.Draw(normalized, bounds, source, bounds.Min, draw.Src)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			standardColor := standard.currentImage.At(x, y)
			fastColor := fast.currentImage.At(x, y)
			if !sameM1Color(standardColor, fastColor) {
				t.Fatalf("grayscale paths differ at (%d,%d)", x, y)
			}
			r, g, b, alpha := standardColor.RGBA()
			if r != g || g != b {
				t.Fatalf("output at (%d,%d) is not grayscale", x, y)
			}

			offset := normalized.PixOffset(x, y)
			wantGray := uint8(
				0.2126*float64(normalized.Pix[offset]) +
					0.7152*float64(normalized.Pix[offset+1]) +
					0.0722*float64(normalized.Pix[offset+2]),
			)
			wantChannel := uint32(wantGray) * 0x101
			if r != wantChannel {
				t.Fatalf("grayscale at (%d,%d) = %d, want BT.709 value %d", x, y, r>>8, wantGray)
			}
			wantAlpha := uint32(normalized.Pix[offset+3]) * 0x101
			if alpha != wantAlpha {
				t.Fatalf("alpha changed at (%d,%d): got %d, want %d", x, y, alpha, wantAlpha)
			}
		}
	}
	if !sameM1Pixels(before, captureM1Pixels(source)) {
		t.Fatal("grayscale modified its source image")
	}
}

func fillM1Image(img draw.Image) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			seed := x - bounds.Min.X + 3*(y-bounds.Min.Y)
			alpha := uint8(64 + seed%192)
			img.Set(x, y, color.NRGBA{
				R: uint8((seed*37 + 17) % 256),
				G: uint8((seed*19 + 91) % 256),
				B: uint8((seed*53 + 3) % 256),
				A: alpha,
			})
		}
	}
}

func captureM1Pixels(img image.Image) [][4]uint32 {
	bounds := img.Bounds()
	pixels := make([][4]uint32, 0, bounds.Dx()*bounds.Dy())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, alpha := img.At(x, y).RGBA()
			pixels = append(pixels, [4]uint32{r, g, b, alpha})
		}
	}
	return pixels
}

func sameM1Pixels(a, b [][4]uint32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func sameM1Color(a, b color.Color) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

type m1BoundsImage struct {
	bounds image.Rectangle
}

func (img m1BoundsImage) ColorModel() color.Model { return color.RGBAModel }
func (img m1BoundsImage) Bounds() image.Rectangle { return img.bounds }
func (img m1BoundsImage) At(_, _ int) color.Color { return color.RGBA{A: 255} }
