package engine

import (
	"bytes"
	"image"
	"image/color"
	"testing"
)

func TestDecodeAndEncodeImage(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 8, 6))
	source.Set(0, 0, color.RGBA{R: 255, A: 255})

	for _, format := range []Format{FormatJPEG, FormatPNG} {
		var buf bytes.Buffer
		if err := encodeImage(&buf, source, format); err != nil {
			t.Fatalf("encodeImage(%s) failed: %v", format, err)
		}
		decoded, err := decodeImage(&buf)
		if err != nil {
			t.Fatalf("decodeImage(%s) failed: %v", format, err)
		}
		if got := decoded.Bounds().Size(); got != image.Pt(8, 6) {
			t.Fatalf("decoded %s size = %v, want 8x6", format, got)
		}
	}

	var buf bytes.Buffer
	if err := encodeImage(&buf, source, FormatGIF); err == nil {
		t.Fatal("encodeImage accepted unsupported GIF output")
	}
	if _, err := decodeImage(bytes.NewReader([]byte("not an image"))); err == nil {
		t.Fatal("decodeImage accepted invalid data")
	}
}
