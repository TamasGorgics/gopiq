package gopiq

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"image/png"
	"testing"
)

func fuzzPNGSeed() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 4, 3))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func fuzzJPEGSeed() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 4, 3))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		panic(err)
	}
	return buf.Bytes()
}

func fuzzGIFSeed() []byte {
	const onePixelGIF = "R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"
	data, err := base64.StdEncoding.DecodeString(onePixelGIF)
	if err != nil {
		panic(err)
	}
	return data
}

func FuzzFromBytesWithLimits(f *testing.F) {
	f.Add(fuzzPNGSeed(), int64(0), 0, 0, int64(0))
	f.Add(fuzzJPEGSeed(), int64(1<<20), 1024, 1024, int64(1<<20))
	f.Add(fuzzGIFSeed(), int64(0), 16, 16, int64(256))
	f.Add([]byte("not an image"), int64(10), 32, 32, int64(1024))

	f.Fuzz(func(t *testing.T, data []byte, maxEncoded int64, maxWidth, maxHeight int, maxPixels int64) {
		t.Helper()
		if len(data) > 1<<20 {
			t.Skip()
		}
		processor := FromBytesWithLimits(data, DecodeLimits{
			MaxEncodedBytes: maxEncoded,
			MaxWidth:        maxWidth,
			MaxHeight:       maxHeight,
			MaxPixels:       maxPixels,
		})
		_ = processor.Err()
		if img, err := processor.Image(); err == nil && img != nil {
			_ = img.Bounds()
		}
	})
}

func FuzzFromReaderWithLimits(f *testing.F) {
	f.Add(fuzzPNGSeed(), int64(0), 0, 0, int64(0))
	f.Add(fuzzJPEGSeed(), int64(1<<20), 1024, 1024, int64(1<<20))
	f.Add(fuzzGIFSeed(), int64(0), 16, 16, int64(256))
	f.Add([]byte("not an image"), int64(10), 32, 32, int64(1024))

	f.Fuzz(func(t *testing.T, data []byte, maxEncoded int64, maxWidth, maxHeight int, maxPixels int64) {
		t.Helper()
		if len(data) > 1<<20 {
			t.Skip()
		}
		processor := FromReaderWithLimits(bytes.NewReader(data), DecodeLimits{
			MaxEncodedBytes: maxEncoded,
			MaxWidth:        maxWidth,
			MaxHeight:       maxHeight,
			MaxPixels:       maxPixels,
		})
		_ = processor.Err()
		if img, err := processor.Image(); err == nil && img != nil {
			_ = img.Bounds()
		}
	})
}
