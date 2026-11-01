package gopiq

import (
	"strings"

	"github.com/TamasGorgics/gopiq/internal/engine"
)

// ImageFormat represents a recognized image format. JPEG, PNG, and GIF are
// supported; GIF output uses the documented lossy palette policy.
type ImageFormat int

// Supported image formats.
const (
	FormatUnknown ImageFormat = iota
	FormatJPEG
	FormatPNG
	FormatGIF // GIF output uses a lossy standard-library palette policy.
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
