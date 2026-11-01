package gopiq_test

import (
	"image"
	"testing"

	"github.com/TamasGorgics/gopiq"
)

func TestPublicFacadeSmoke(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 64, 48))
	processor := gopiq.New(source).Resize(32, 24).Grayscale()
	if err := processor.Err(); err != nil {
		t.Fatalf("public chain failed: %v", err)
	}

	cloned := processor.Clone()
	first, err := processor.Image()
	if err != nil {
		t.Fatalf("Image failed: %v", err)
	}
	second, err := cloned.Image()
	if err != nil {
		t.Fatalf("cloned Image failed: %v", err)
	}
	if first != second {
		t.Fatal("Clone should share the current image reference")
	}

	data, err := processor.ToBytes(gopiq.FormatPNG)
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}
	decoded := gopiq.FromBytes(data)
	if err := decoded.Err(); err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}
	decodedImage, err := decoded.Image()
	if err != nil {
		t.Fatalf("decoded Image failed: %v", err)
	}
	if got, want := decodedImage.Bounds().Size(), image.Pt(32, 24); got != want {
		t.Fatalf("decoded bounds = %v, want %v", got, want)
	}
}
