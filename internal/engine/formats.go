package engine

import (
	"fmt"
	"image"
	"image/color"
	"image/gif" // Register GIF decoding and provide GIF encoding.
	"image/jpeg"
	"image/png"
	"io"

	"golang.org/x/image/draw"
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
		return encodeGIF(w, img)
	default:
		return fmt.Errorf("unsupported image format for encoding: %s", format.String())
	}
}

func encodeGIF(w io.Writer, img image.Image) error {
	bounds := img.Bounds()
	if bounds.Dx() > 65535 || bounds.Dy() > 65535 {
		return fmt.Errorf("GIF dimensions exceed 65535 pixels")
	}

	opaque := image.NewRGBA(bounds)
	draw.Draw(opaque, bounds, image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(opaque, bounds, img, bounds.Min, draw.Over)
	return gif.Encode(w, opaque, nil)
}
