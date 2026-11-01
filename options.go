package gopiq

import (
	"image/color"

	"github.com/TamasGorgics/gopiq/internal/engine"
)

// WatermarkPosition defines common positions for the watermark.
type WatermarkPosition int

// Supported watermark anchor positions.
const (
	PositionTopLeft WatermarkPosition = iota
	PositionTopRight
	PositionBottomLeft
	PositionBottomRight
	PositionCenter
)

// watermarkConfig holds configuration for adding text watermark.
type watermarkConfig struct {
	Text       string
	FontPath   string
	FontBytes  []byte
	fontSource watermarkFontSource
	FontSize   float64
	Color      color.Color
	Position   WatermarkPosition
	OffsetX    float64
	OffsetY    float64
}

type watermarkFontSource uint8

const (
	watermarkFontSourceUnset watermarkFontSource = iota
	watermarkFontSourceBytes
	watermarkFontSourcePath
)

func defaultWatermarkConfig() *watermarkConfig {
	return &watermarkConfig{
		FontSize: 24,
		Color:    color.RGBA{255, 255, 255, 128},
		Position: PositionBottomRight,
		OffsetX:  10,
		OffsetY:  10,
	}
}

// WatermarkOption is a functional option for configuring the watermark.
type WatermarkOption func(*watermarkConfig)

// WithFontPath specifies a font file for the watermark. If both font-source
// options are supplied, the last one takes precedence.
func WithFontPath(path string) WatermarkOption {
	return func(wc *watermarkConfig) {
		wc.FontPath = path
		wc.fontSource = watermarkFontSourcePath
	}
}

// WithFontBytes specifies font data directly (e.g., from an embedded font).
// This is generally preferred for self-contained libraries. If both font-source
// options are supplied, the last one takes precedence.
func WithFontBytes(data []byte) WatermarkOption {
	return func(wc *watermarkConfig) {
		wc.FontBytes = data
		wc.fontSource = watermarkFontSourceBytes
	}
}

// WithFontSize sets the font size for the watermark. The size must be finite,
// positive, and within the renderer's fixed-point range.
func WithFontSize(size float64) WatermarkOption {
	return func(wc *watermarkConfig) { wc.FontSize = size }
}

// WithColor sets the color for the watermark.
func WithColor(c color.Color) WatermarkOption {
	return func(wc *watermarkConfig) { wc.Color = c }
}

// WithPosition sets the position of the watermark.
func WithPosition(pos WatermarkPosition) WatermarkOption {
	return func(wc *watermarkConfig) { wc.Position = pos }
}

// WithOffset sets an additional offset (in pixels) from the chosen position.
// Positive offsets move inward from corner anchors; for the center position,
// positive X moves right and positive Y moves down. Negative offsets can move
// text partly outside the image.
func WithOffset(x, y float64) WatermarkOption {
	return func(wc *watermarkConfig) { wc.OffsetX = x; wc.OffsetY = y }
}

func toEngineWatermarkConfig(cfg *watermarkConfig) engine.WatermarkConfig {
	out := engine.DefaultWatermarkConfig()
	switch cfg.fontSource {
	case watermarkFontSourceUnset:
		// Keep the engine's embedded default font.
	case watermarkFontSourceBytes:
		out.FontBytes = cfg.FontBytes
		out.FontSource = engine.FontSourceBytes
	case watermarkFontSourcePath:
		out.FontPath = cfg.FontPath
		out.FontSource = engine.FontSourcePath
	}
	out.FontSize = cfg.FontSize
	out.Color = cfg.Color
	out.Position = engine.Position(cfg.Position)
	out.OffsetX = cfg.OffsetX
	out.OffsetY = cfg.OffsetY
	return out
}

// PerformanceOptions controls optimization settings for image processing.
type PerformanceOptions struct {
	// MaxGoroutines limits the number of parallel goroutines for heavy operations.
	// If less than or equal to 0, defaults to runtime.NumCPU().
	MaxGoroutines int
	// EnableParallelProcessing enables parallel processing for suitable operations.
	EnableParallelProcessing bool
	// MinSizeForParallel sets the minimum image size (width * height) before
	// parallel processing is used. Values less than or equal to 0 allow parallel
	// processing for any non-empty image when parallel processing is enabled.
	MinSizeForParallel int
}

// DefaultPerformanceOptions returns optimized defaults for most use cases.
func DefaultPerformanceOptions() PerformanceOptions {
	return fromEngineOptions(engine.DefaultOptions())
}

func toEngineOptions(opts PerformanceOptions) engine.Options {
	return engine.Options{
		MaxGoroutines:            opts.MaxGoroutines,
		EnableParallelProcessing: opts.EnableParallelProcessing,
		MinSizeForParallel:       opts.MinSizeForParallel,
	}
}

func fromEngineOptions(opts engine.Options) PerformanceOptions {
	return PerformanceOptions{
		MaxGoroutines:            opts.MaxGoroutines,
		EnableParallelProcessing: opts.EnableParallelProcessing,
		MinSizeForParallel:       opts.MinSizeForParallel,
	}
}
