package gopiq

import (
	"strings"

	"github.com/TamasGorgics/gopiq/internal/engine"
)

// ImageFormat represents a recognized image format. GIF is decode-only;
// JPEG and PNG are supported for both decoding and encoding.
type ImageFormat int

const (
	FormatUnknown ImageFormat = iota
	FormatJPEG
	FormatPNG
	FormatGIF // Can be decoded, but GIF encoding is not supported.
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

func toEngineFormat(format ImageFormat) engine.Format {
	switch format {
	case FormatJPEG:
		return engine.FormatJPEG
	case FormatPNG:
		return engine.FormatPNG
	case FormatGIF:
		return engine.FormatGIF
	default:
		return engine.FormatUnknown
	}
}
