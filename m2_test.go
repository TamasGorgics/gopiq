package gopiq

import (
	"encoding/base64"
	"image"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

func TestM2WatermarkFontPathAndSourcePrecedence(t *testing.T) {
	fontPath := filepath.Join(t.TempDir(), "custom.ttf")
	if err := os.WriteFile(fontPath, goregular.TTF, 0o600); err != nil {
		t.Fatalf("write test font: %v", err)
	}
	invalidFontPath := filepath.Join(t.TempDir(), "invalid.ttf")
	if err := os.WriteFile(invalidFontPath, []byte{1, 2, 3}, 0o600); err != nil {
		t.Fatalf("write invalid test font: %v", err)
	}

	newProcessor := func() *ImageProcessor {
		return New(image.NewRGBA(image.Rect(0, 0, 160, 90)))
	}

	tests := []struct {
		name    string
		options []WatermarkOption
		wantErr bool
	}{
		{name: "default embedded font"},
		{name: "font from path", options: []WatermarkOption{WithFontPath(fontPath)}},
		{name: "missing path", options: []WatermarkOption{WithFontPath(filepath.Join(t.TempDir(), "missing.ttf"))}, wantErr: true},
		{name: "invalid font file", options: []WatermarkOption{WithFontPath(invalidFontPath)}, wantErr: true},
		{name: "bytes override path", options: []WatermarkOption{WithFontPath(fontPath), WithFontBytes([]byte{1, 2, 3})}, wantErr: true},
		{name: "path overrides bytes", options: []WatermarkOption{WithFontBytes([]byte{1, 2, 3}), WithFontPath(fontPath)}},
		{name: "bytes override missing path", options: []WatermarkOption{WithFontPath(filepath.Join(t.TempDir(), "missing.ttf")), WithFontBytes(goregular.TTF)}},
		{name: "empty bytes remain explicit", options: []WatermarkOption{WithFontBytes(nil)}, wantErr: true},
		{name: "empty path", options: []WatermarkOption{WithFontPath("")}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			processor := newProcessor().AddTextWatermark("font source", tc.options...)
			if (processor.Err() != nil) != tc.wantErr {
				t.Fatalf("AddTextWatermark error = %v, wantErr %t", processor.Err(), tc.wantErr)
			}
		})
	}
}

func TestM2WatermarkRejectsInvalidOptions(t *testing.T) {
	var typedNilColor *m2NilColor
	tests := []struct {
		name   string
		option WatermarkOption
	}{
		{name: "nil option", option: nil},
		{name: "nil color", option: WithColor(nil)},
		{name: "typed nil color", option: WithColor(typedNilColor)},
		{name: "zero font size", option: WithFontSize(0)},
		{name: "negative font size", option: WithFontSize(-1)},
		{name: "NaN font size", option: WithFontSize(math.NaN())},
		{name: "infinite font size", option: WithFontSize(math.Inf(1))},
		{name: "unrepresentable font size", option: WithFontSize(math.MaxFloat64)},
		{name: "NaN horizontal offset", option: WithOffset(math.NaN(), 0)},
		{name: "infinite vertical offset", option: WithOffset(0, math.Inf(-1))},
		{name: "unknown position", option: WithPosition(WatermarkPosition(99))},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			processor := New(image.NewRGBA(image.Rect(0, 0, 120, 80))).AddTextWatermark("validate", tc.option)
			if processor.Err() == nil {
				t.Fatal("invalid watermark option did not set a processor error")
			}
		})
	}
}

func TestM2WatermarkPositionsAreRelativeToBounds(t *testing.T) {
	const width, height = 150, 100
	zeroOrigin := image.NewRGBA(image.Rect(0, 0, width, height))
	translated := image.NewRGBA(image.Rect(37, -29, 37+width, -29+height))
	fillM1Image(zeroOrigin)
	fillM1Image(translated)
	zeroOriginBefore := captureM1Pixels(zeroOrigin)
	translatedBefore := captureM1Pixels(translated)

	positions := []struct {
		name     string
		position WatermarkPosition
	}{
		{name: "top left", position: PositionTopLeft},
		{name: "top right", position: PositionTopRight},
		{name: "bottom left", position: PositionBottomLeft},
		{name: "bottom right", position: PositionBottomRight},
		{name: "center", position: PositionCenter},
	}
	for _, tc := range positions {
		t.Run(tc.name, func(t *testing.T) {
			options := []WatermarkOption{
				WithPosition(tc.position),
				WithFontSize(18),
				WithColor(color.RGBA{R: 255, G: 30, B: 90, A: 255}),
				WithOffset(7, 5),
			}
			zeroProcessor := New(zeroOrigin).AddTextWatermark("M2", options...)
			translatedProcessor := New(translated).AddTextWatermark("M2", options...)
			if err := zeroProcessor.Err(); err != nil {
				t.Fatalf("zero-origin watermark failed: %v", err)
			}
			if err := translatedProcessor.Err(); err != nil {
				t.Fatalf("translated watermark failed: %v", err)
			}
			if got, want := translatedProcessor.currentImage.Bounds(), translated.Bounds(); got != want {
				t.Fatalf("watermarked bounds = %v, want %v", got, want)
			}

			changed := false
			var changedBounds image.Rectangle
			for y := 0; y < height; y++ {
				for x := 0; x < width; x++ {
					zeroColor := zeroProcessor.currentImage.At(x, y)
					translatedColor := translatedProcessor.currentImage.At(x+translated.Bounds().Min.X, y+translated.Bounds().Min.Y)
					if !sameM1Color(zeroColor, translatedColor) {
						t.Fatalf("watermark output differs at relative pixel (%d,%d)", x, y)
					}
					if !sameM1Color(zeroOrigin.At(x, y), zeroColor) {
						changed = true
						if changedBounds.Empty() {
							changedBounds = image.Rect(x, y, x+1, y+1)
						} else {
							if x < changedBounds.Min.X {
								changedBounds.Min.X = x
							}
							if y < changedBounds.Min.Y {
								changedBounds.Min.Y = y
							}
							if x+1 > changedBounds.Max.X {
								changedBounds.Max.X = x + 1
							}
							if y+1 > changedBounds.Max.Y {
								changedBounds.Max.Y = y + 1
							}
						}
					}
				}
			}
			if !changed {
				t.Fatal("watermark produced no visible pixels")
			}
			switch tc.position {
			case PositionTopLeft:
				if changedBounds.Min.X >= width/2 || changedBounds.Min.Y >= height/2 {
					t.Fatalf("top-left watermark landed in the wrong region: %v", changedBounds)
				}
			case PositionTopRight:
				if changedBounds.Max.X <= width/2 || changedBounds.Min.Y >= height/2 {
					t.Fatalf("top-right watermark landed in the wrong region: %v", changedBounds)
				}
			case PositionBottomLeft:
				if changedBounds.Min.X >= width/2 || changedBounds.Max.Y <= height/2 {
					t.Fatalf("bottom-left watermark landed in the wrong region: %v", changedBounds)
				}
			case PositionBottomRight:
				if changedBounds.Max.X <= width/2 || changedBounds.Max.Y <= height/2 {
					t.Fatalf("bottom-right watermark landed in the wrong region: %v", changedBounds)
				}
			case PositionCenter:
				centerX := (changedBounds.Min.X + changedBounds.Max.X) / 2
				centerY := (changedBounds.Min.Y + changedBounds.Max.Y) / 2
				if abs(centerX-(width/2+7)) > 12 || abs(centerY-(height/2+5)) > 12 {
					t.Fatalf("center watermark offset was not applied: bounds %v", changedBounds)
				}
			}
		})
	}

	if !sameM1Pixels(translatedBefore, captureM1Pixels(translated)) {
		t.Fatal("watermarking modified the source image")
	}
	if !sameM1Pixels(zeroOriginBefore, captureM1Pixels(zeroOrigin)) {
		t.Fatal("watermarking modified the zero-origin source image")
	}
}

func TestM2WatermarkNegativeAndFullyClippedOffsets(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 100, 60))
	fillM1Image(source)

	partiallyClipped := New(source).AddTextWatermark(
		"clip",
		WithPosition(PositionTopLeft),
		WithOffset(-5, -5),
	)
	if err := partiallyClipped.Err(); err != nil {
		t.Fatalf("finite negative offsets should be valid: %v", err)
	}
	if partiallyClipped.currentImage == source || sameM1Pixels(captureM1Pixels(source), captureM1Pixels(partiallyClipped.currentImage)) {
		t.Fatal("partially clipped negative offsets did not draw a visible watermark")
	}

	fullyClipped := New(source).AddTextWatermark(
		"off canvas",
		WithPosition(PositionTopLeft),
		WithOffset(math.MaxFloat64, math.MaxFloat64),
	)
	if err := fullyClipped.Err(); err != nil {
		t.Fatalf("finite offsets outside the image should be clipped: %v", err)
	}
	if fullyClipped.currentImage != source {
		t.Fatal("fully clipped watermark should leave the image unchanged")
	}

	fullyClippedLeft := New(source).AddTextWatermark(
		"off canvas",
		WithPosition(PositionTopLeft),
		WithOffset(-1000, -1000),
	)
	if err := fullyClippedLeft.Err(); err != nil {
		t.Fatalf("finite negative offsets outside the image should be clipped: %v", err)
	}
	if fullyClippedLeft.currentImage != source {
		t.Fatal("fully clipped negative offsets should not reposition text")
	}
}

func TestM2WatermarkFixedPointCoordinateTranslation(t *testing.T) {
	const fixedPixelMax = (1 << 25) - 1
	for _, pixel := range []int{fixedPixelMax + 1, -fixedPixelMax - 2} {
		dot, shift := watermarkFixedCoordinate(pixel)
		if dot != 0 || shift != pixel {
			t.Errorf("watermarkFixedCoordinate(%d) = (%d, %d), want (0, %d)", pixel, dot, shift, pixel)
		}
	}
	for _, pixel := range []int{0, fixedPixelMax, -1 << 25} {
		dot, shift := watermarkFixedCoordinate(pixel)
		if want := fixed.I(pixel); dot != want || shift != 0 {
			t.Errorf("watermarkFixedCoordinate(%d) = (%d, %d), want (%d, 0)", pixel, dot, shift, want)
		}
	}
}

func TestM2GIFDecodingAndFormatIdentifiers(t *testing.T) {
	// This fixture is intentionally not generated with image/gif: the test must
	// verify that gopiq itself registers the GIF decoder.
	const onePixelGIF = "R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
	data, err := base64.StdEncoding.DecodeString(onePixelGIF)
	if err != nil {
		t.Fatalf("decode GIF fixture: %v", err)
	}

	processor := FromBytes(data)
	if err := processor.Err(); err != nil {
		t.Fatalf("FromBytes did not decode GIF: %v", err)
	}
	img, err := processor.Image()
	if err != nil {
		t.Fatalf("get decoded GIF image: %v", err)
	}
	if got, want := img.Bounds().Size(), image.Pt(1, 1); got != want {
		t.Fatalf("decoded GIF dimensions = %v, want %v", got, want)
	}
	if got := FormatFromString("GIF"); got != FormatGIF {
		t.Fatalf("FormatFromString(GIF) = %v, want FormatGIF", got)
	}
	if got := FormatGIF.String(); got != "gif" {
		t.Fatalf("FormatGIF.String() = %q, want gif", got)
	}
	for _, tc := range []struct {
		input string
		want  ImageFormat
	}{
		{input: "jpeg", want: FormatJPEG},
		{input: "JPG", want: FormatJPEG},
		{input: "png", want: FormatPNG},
		{input: "gif", want: FormatGIF},
		{input: "webp", want: FormatUnknown},
	} {
		if got := FormatFromString(tc.input); got != tc.want {
			t.Errorf("FormatFromString(%q) = %v, want %v", tc.input, got, tc.want)
		}
	}
	for _, tc := range []struct {
		format ImageFormat
		want   string
	}{
		{format: FormatJPEG, want: "jpeg"},
		{format: FormatPNG, want: "png"},
		{format: FormatGIF, want: "gif"},
		{format: FormatUnknown, want: "unknown"},
	} {
		if got := tc.format.String(); got != tc.want {
			t.Errorf("ImageFormat(%d).String() = %q, want %q", tc.format, got, tc.want)
		}
	}
	if _, err := processor.ToBytes(FormatGIF); err == nil {
		t.Fatal("GIF encoding unexpectedly succeeded")
	} else if !strings.Contains(err.Error(), "GIF encoding is unsupported") {
		t.Fatalf("GIF encode error is not clear: %v", err)
	}
}

type m2NilColor struct{}

func (*m2NilColor) RGBA() (uint32, uint32, uint32, uint32) { return 0, 0, 0, 0 }
