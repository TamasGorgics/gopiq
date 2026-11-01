package engine

import (
	"fmt"
	"image"
	_ "image/gif" // Register GIF decoding with image.Decode.
	"image/jpeg"
	"image/png"
	"io"
)

// decodeImage decodes an image from an io.Reader.
func decodeImage(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	return img, nil
}

// encodeImage encodes an image to an io.Writer in the specified format.
func encodeImage(w io.Writer, img image.Image, format Format) error {
	switch format {
	case FormatJPEG:
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 90})
	case FormatPNG:
		return png.Encode(w, img)
	case FormatGIF:
		return fmt.Errorf("GIF encoding is unsupported because palette quantization is not implemented")
	default:
		return fmt.Errorf("unsupported image format for encoding: %s", format.String())
	}
}
