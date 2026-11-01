package gopiq

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"testing"
)

func m7PatternImage(bounds image.Rectangle) *image.RGBA {
	img := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			relX := x - bounds.Min.X
			relY := y - bounds.Min.Y
			img.SetRGBA(x, y, color.RGBA{
				R: uint8((relX*37 + relY*11) % 256),
				G: uint8((relX*13 + relY*41) % 256),
				B: uint8((relX*7 + relY*23) % 256),
				A: uint8(64 + (relX*17+relY*29)%192),
			})
		}
	}
	return img
}

func TestM7ResizeFitAndFill(t *testing.T) {
	tests := []struct {
		name                string
		width, height       int
		targetWidth         int
		targetHeight        int
		fitWidth, fitHeight int
	}{
		{name: "landscape", width: 4, height: 3, targetWidth: 100, targetHeight: 100, fitWidth: 100, fitHeight: 75},
		{name: "portrait", width: 3, height: 4, targetWidth: 100, targetHeight: 100, fitWidth: 75, fitHeight: 100},
		{name: "rounding", width: 3, height: 2, targetWidth: 10, targetHeight: 10, fitWidth: 10, fitHeight: 7},
		{name: "square upscale", width: 1, height: 1, targetWidth: 5, targetHeight: 5, fitWidth: 5, fitHeight: 5},
		{name: "minimum dimension", width: 1, height: 100, targetWidth: 10, targetHeight: 10, fitWidth: 1, fitHeight: 10},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := m7PatternImage(image.Rect(11, -7, 11+tc.width, -7+tc.height))
			before := captureM1Pixels(source)

			fit := New(source).ResizeFit(tc.targetWidth, tc.targetHeight)
			if err := fit.Err(); err != nil {
				t.Fatalf("ResizeFit failed: %v", err)
			}
			fitImage := fit.testImage()
			if got, want := fitImage.Bounds(), image.Rect(0, 0, tc.fitWidth, tc.fitHeight); got != want {
				t.Fatalf("ResizeFit bounds = %v, want %v", got, want)
			}
			if !sameM1Pixels(before, captureM1Pixels(source)) {
				t.Fatal("ResizeFit modified its source image")
			}

			fill := New(source).ResizeFill(tc.targetWidth, tc.targetHeight)
			if err := fill.Err(); err != nil {
				t.Fatalf("ResizeFill failed: %v", err)
			}
			fillImage := fill.testImage()
			if got, want := fillImage.Bounds(), image.Rect(0, 0, tc.targetWidth, tc.targetHeight); got != want {
				t.Fatalf("ResizeFill bounds = %v, want %v", got, want)
			}
			if !sameM1Pixels(before, captureM1Pixels(source)) {
				t.Fatal("ResizeFill modified its source image")
			}
		})
	}
}

func TestM7ResizeFillCenterCrop(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 5, 1))
	colors := []color.RGBA{
		{R: 255, A: 255},
		{G: 255, A: 255},
		{B: 255, A: 255},
		{R: 255, G: 255, A: 255},
		{R: 255, B: 255, A: 255},
	}
	for x, c := range colors {
		source.SetRGBA(x, 0, c)
	}

	result := New(source).ResizeFill(2, 1)
	if err := result.Err(); err != nil {
		t.Fatalf("ResizeFill failed: %v", err)
	}
	output := result.testImage()
	if !sameM1Color(source.At(1, 0), output.At(0, 0)) || !sameM1Color(source.At(2, 0), output.At(1, 0)) {
		t.Fatal("ResizeFill did not use the expected center crop")
	}
}

func TestM7ResizeFitAndFillErrors(t *testing.T) {
	if err := New(m7PatternImage(image.Rect(0, 0, 4, 3))).ResizeFit(0, 10).Err(); err == nil {
		t.Fatal("ResizeFit accepted an invalid target width")
	}
	if err := New(m7PatternImage(image.Rect(0, 0, 4, 3))).ResizeFill(10, -1).Err(); err == nil {
		t.Fatal("ResizeFill accepted an invalid target height")
	}
	empty := image.NewRGBA(image.Rect(2, 3, 2, 3))
	if err := New(empty).ResizeFit(10, 10).Err(); err == nil {
		t.Fatal("ResizeFit accepted an empty source image")
	}
	if err := New(empty).ResizeFill(10, 10).Err(); err == nil {
		t.Fatal("ResizeFill accepted an empty source image")
	}
}

func TestM7FlipsAndRotations(t *testing.T) {
	bounds := image.Rect(10, 20, 13, 22)
	tests := []struct {
		name      string
		width     int
		height    int
		transform func(*ImageProcessor) *ImageProcessor
		sourceFor func(image.Rectangle, int, int) (int, int)
	}{
		{
			name: "flip horizontal", width: 3, height: 2,
			transform: func(p *ImageProcessor) *ImageProcessor { return p.FlipHorizontal() },
			sourceFor: func(bounds image.Rectangle, x, y int) (int, int) {
				return bounds.Max.X - 1 - x, bounds.Min.Y + y
			},
		},
		{
			name: "flip vertical", width: 3, height: 2,
			transform: func(p *ImageProcessor) *ImageProcessor { return p.FlipVertical() },
			sourceFor: func(bounds image.Rectangle, x, y int) (int, int) {
				return bounds.Min.X + x, bounds.Max.Y - 1 - y
			},
		},
		{
			name: "rotate 90", width: 2, height: 3,
			transform: func(p *ImageProcessor) *ImageProcessor { return p.Rotate90() },
			sourceFor: func(bounds image.Rectangle, x, y int) (int, int) {
				return bounds.Min.X + y, bounds.Max.Y - 1 - x
			},
		},
		{
			name: "rotate 180", width: 3, height: 2,
			transform: func(p *ImageProcessor) *ImageProcessor { return p.Rotate180() },
			sourceFor: func(bounds image.Rectangle, x, y int) (int, int) {
				return bounds.Max.X - 1 - x, bounds.Max.Y - 1 - y
			},
		},
		{
			name: "rotate 270", width: 2, height: 3,
			transform: func(p *ImageProcessor) *ImageProcessor { return p.Rotate270() },
			sourceFor: func(bounds image.Rectangle, x, y int) (int, int) {
				return bounds.Max.X - 1 - y, bounds.Min.Y + x
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			source := m7PatternImage(bounds)
			before := captureM1Pixels(source)
			result := tc.transform(New(source))
			if err := result.Err(); err != nil {
				t.Fatalf("%s failed: %v", tc.name, err)
			}
			output := result.testImage()
			if got, want := output.Bounds(), image.Rect(0, 0, tc.width, tc.height); got != want {
				t.Fatalf("%s bounds = %v, want %v", tc.name, got, want)
			}
			for y := 0; y < tc.height; y++ {
				for x := 0; x < tc.width; x++ {
					sourceX, sourceY := tc.sourceFor(bounds, x, y)
					if !sameM1Color(source.At(sourceX, sourceY), output.At(x, y)) {
						t.Fatalf("%s pixel (%d,%d) does not match source (%d,%d)", tc.name, x, y, sourceX, sourceY)
					}
				}
			}
			if !sameM1Pixels(before, captureM1Pixels(source)) {
				t.Fatalf("%s modified its source image", tc.name)
			}
		})
	}
}

func TestM7EmptyTransformIsNoOp(t *testing.T) {
	source := image.NewRGBA(image.Rect(2, 3, 2, 3))
	for _, result := range []*ImageProcessor{
		New(source).FlipHorizontal(),
		New(source).FlipVertical(),
		New(source).Rotate90(),
		New(source).Rotate180(),
		New(source).Rotate270(),
	} {
		if err := result.Err(); err != nil {
			t.Fatalf("empty transform failed: %v", err)
		}
		if result.testImage() != source {
			t.Fatal("empty transform replaced the source image")
		}
	}
}

func TestM7GIFOutputPolicy(t *testing.T) {
	source := m7PatternImage(image.Rect(0, 0, 4, 3))
	processor := New(source)
	data, err := processor.ToBytes(FormatGIF)
	if err != nil {
		t.Fatalf("ToBytes(FormatGIF) failed: %v", err)
	}

	decoded, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode GIF output: %v", err)
	}
	if format != "gif" {
		t.Fatalf("GIF output format = %q, want gif", format)
	}
	if got, want := decoded.Bounds().Size(), image.Pt(4, 3); got != want {
		t.Fatalf("GIF bounds = %v, want %v", got, want)
	}
	paletted, ok := decoded.(*image.Paletted)
	if !ok {
		t.Fatalf("GIF output type = %T, want *image.Paletted", decoded)
	}
	if len(paletted.Palette) > 256 {
		t.Fatalf("GIF palette has %d colors, want at most 256", len(paletted.Palette))
	}

	frames, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode GIF frames: %v", err)
	}
	if len(frames.Image) != 1 {
		t.Fatalf("GIF frame count = %d, want 1", len(frames.Image))
	}

	var streamed bytes.Buffer
	if err := processor.ToWriter(&streamed, FormatGIF); err != nil {
		t.Fatalf("ToWriter(FormatGIF) failed: %v", err)
	}
	if _, format, err := image.Decode(&streamed); err != nil || format != "gif" {
		t.Fatalf("streamed GIF = format %q, error %v", format, err)
	}
}

func TestM7GIFAlphaIsFlattenedOverWhite(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 2, 1))
	source.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	source.SetRGBA(1, 0, color.RGBA{A: 0})
	data, err := New(source).ToBytes(FormatGIF)
	if err != nil {
		t.Fatalf("ToBytes(FormatGIF) failed: %v", err)
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("decode GIF: %v", err)
	}
	r, g, b, a := decoded.At(1, 0).RGBA()
	if r>>8 < 200 || g>>8 < 200 || b>>8 < 200 || a>>8 < 200 {
		t.Fatalf("transparent GIF pixel was not flattened to white: %d,%d,%d,%d", r>>8, g>>8, b>>8, a>>8)
	}
}

func TestM7GIFDimensionLimit(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 65_536, 1))
	if _, err := New(source).ToBytes(FormatGIF); err == nil {
		t.Fatal("ToBytes(FormatGIF) accepted an oversized image")
	}
}
