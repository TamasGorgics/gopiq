package gopiq

import (
	"fmt"
	"image"
	_ "image/gif" // Register GIF format for decoding
	"image/jpeg"
	_ "image/jpeg" // Register JPEG format for decoding
	"image/png"
	_ "image/png" // Register PNG format for decoding
	"io"
	"strings"
)

// ImageFormat represents supported image output formats.
type ImageFormat int

const (
	FormatUnknown ImageFormat = iota
	FormatJPEG
	FormatPNG
	FormatGIF // Can decode, but encoding to Paletted/GIF requires more work than current scope.
)

// String returns the string representation of the ImageFormat.
func (f ImageFormat) String() string {
	switch f {
	case FormatJPEG:
		return "jpeg"
	case FormatPNG:
		return "png"
	case FormatGIF:
		return "gif"
	default:
		return "unknown"
	}
}

// FormatFromString converts a string to an ImageFormat.
func FormatFromString(s string) ImageFormat {
	switch strings.ToLower(s) {
	case "jpeg", "jpg":
		return FormatJPEG
	case "png":
		return FormatPNG
	case "gif":
		return FormatGIF
	default:
		return FormatUnknown
	}
}

// decodeImage decodes an image from an io.Reader.
func decodeImage(r io.Reader) (image.Image, error) {
	img, _, err := image.Decode(r)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	return img, nil
}

// encodeImage encodes an image to an io.Writer in the specified format.
func encodeImage(w io.Writer, img image.Image, format ImageFormat) error {
	switch format {
	case FormatJPEG:
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 90}) // Default JPEG quality 90
	case FormatPNG:
		return png.Encode(w, img)
	case FormatGIF:
		// GIF encoding requires image.Paletted. Converting an arbitrary image.Image
		// to image.Paletted (e.g., quantizing colors) requires external libraries
		// beyond golang.org/x, or a complex manual implementation of color quantization.
		return fmt.Errorf("GIF encoding is not directly supported without 3rd-party color quantization")
	default:
		return fmt.Errorf("unsupported image format for encoding: %s", format.String())
	}
}

// newRGBA creates a new RGBA image with the given bounds.
func newRGBA(bounds image.Rectangle) *image.RGBA {
	return image.NewRGBA(bounds)
}
