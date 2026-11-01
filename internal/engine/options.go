package engine

import (
	"image/color"
	"runtime"

	"golang.org/x/image/font/gofont/goregular"
)

// Options controls optimization settings for image processing.
type Options struct {
	MaxGoroutines            int
	EnableParallelProcessing bool
	MinSizeForParallel       int
}

// DefaultOptions returns optimized defaults for most use cases.
func DefaultOptions() Options {
	return Options{
		MaxGoroutines:            runtime.NumCPU(),
		EnableParallelProcessing: true,
		MinSizeForParallel:       10000, // 100x100 pixels
	}
}

// Format identifies a recognized image format. GIF is decode-only.
type Format int

// Recognized image formats.
const (
	FormatUnknown Format = iota
	FormatJPEG
	FormatPNG
	FormatGIF
)

// String returns the string representation of the format.
func (f Format) String() string {
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

// Position defines a watermark anchor.
type Position int

// Supported watermark anchor positions.
const (
	PositionTopLeft Position = iota
	PositionTopRight
	PositionBottomLeft
	PositionBottomRight
	PositionCenter
)

// FontSource identifies where watermark font data comes from.
type FontSource uint8

// Supported watermark font sources.
const (
	FontSourceUnset FontSource = iota
	FontSourceBytes
	FontSourcePath
)

// WatermarkConfig holds a fully resolved watermark configuration.
type WatermarkConfig struct {
	Text       string
	FontPath   string
	FontBytes  []byte
	FontSource FontSource
	FontSize   float64
	Color      color.Color
	Position   Position
	OffsetX    float64
	OffsetY    float64
}

// DefaultWatermarkConfig provides sane watermark defaults.
func DefaultWatermarkConfig() WatermarkConfig {
	return WatermarkConfig{
		FontSize:   24,
		Color:      color.RGBA{255, 255, 255, 128}, // White with 50% opacity
		Position:   PositionBottomRight,
		OffsetX:    10,
		OffsetY:    10,
		FontBytes:  goregular.TTF,
		FontSource: FontSourceBytes,
	}
}
