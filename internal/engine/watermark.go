package engine

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"reflect"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// AddTextWatermark adds a text watermark. optionErr carries an option-level
// failure discovered by the public facade so the first chain error is preserved.
func (p *Processor) AddTextWatermark(text string, cfg WatermarkConfig, optionErr error) *Processor {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.err != nil {
		return p
	}
	if text == "" {
		p.err = fmt.Errorf("watermark text cannot be empty")
		return p
	}
	if p.currentImage == nil {
		p.err = fmt.Errorf("no image available for watermark")
		return p
	}

	bounds := p.currentImage.Bounds()
	width, height, err := checkedImageDimensions(bounds)
	if err != nil {
		p.err = fmt.Errorf("invalid source image bounds: %w", err)
		return p
	}
	if width == 0 || height == 0 {
		p.err = fmt.Errorf("cannot add a watermark to an image with empty bounds %v", bounds)
		return p
	}
	if _, err := checkedRGBABytes(width, height); err != nil {
		p.err = fmt.Errorf("invalid watermark image size: %w", err)
		return p
	}
	if optionErr != nil {
		p.err = optionErr
		return p
	}

	cfg.Text = text
	if err := validateWatermarkConfig(cfg); err != nil {
		p.err = err
		return p
	}

	fontBytes := cfg.FontBytes
	if cfg.FontSource == FontSourcePath {
		fontBytes, err = os.ReadFile(cfg.FontPath)
		if err != nil {
			p.err = fmt.Errorf("failed to read watermark font %q: %w", cfg.FontPath, err)
			return p
		}
	}

	fnt, err := opentype.Parse(fontBytes)
	if err != nil {
		if cfg.FontSource == FontSourcePath {
			p.err = fmt.Errorf("failed to parse watermark font %q: %w", cfg.FontPath, err)
		} else {
			p.err = fmt.Errorf("failed to parse font bytes for watermark: %w", err)
		}
		return p
	}

	face, err := opentype.NewFace(fnt, &opentype.FaceOptions{
		Size:    cfg.FontSize,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		p.err = fmt.Errorf("failed to create font face for watermark: %w", err)
		return p
	}
	defer face.Close()

	metrics := face.Metrics()
	if metrics.Height < 0 || metrics.Ascent < 0 || metrics.Descent < 0 {
		p.err = fmt.Errorf("watermark font produced invalid metrics")
		return p
	}

	measureDrawer := font.Drawer{Face: face}
	textBounds, _ := measureDrawer.BoundString(cfg.Text)
	if textBounds.Max.X < textBounds.Min.X || textBounds.Max.Y < textBounds.Min.Y {
		p.err = fmt.Errorf("watermark font produced invalid text bounds")
		return p
	}
	textWidth := float64(textBounds.Max.X)/64 - float64(textBounds.Min.X)/64
	textHeight := float64(metrics.Height) / 64
	ascent := float64(metrics.Ascent) / 64
	descent := float64(metrics.Descent) / 64

	var x, y float64
	switch cfg.Position {
	case PositionTopLeft:
		x = cfg.OffsetX
		y = cfg.OffsetY + ascent
	case PositionTopRight:
		x = float64(width) - textWidth - cfg.OffsetX
		y = cfg.OffsetY + ascent
	case PositionBottomLeft:
		x = cfg.OffsetX
		y = float64(height) - cfg.OffsetY - descent
	case PositionBottomRight:
		x = float64(width) - textWidth - cfg.OffsetX
		y = float64(height) - cfg.OffsetY - descent
	case PositionCenter:
		x = (float64(width)-textWidth)/2 + cfg.OffsetX
		y = (float64(height)-textHeight)/2 + ascent + cfg.OffsetY
	default:
		p.err = fmt.Errorf("unknown watermark position: %d", cfg.Position)
		return p
	}

	textLeft := x + float64(textBounds.Min.X)/64
	textRight := x + float64(textBounds.Max.X)/64
	textTop := y + float64(textBounds.Min.Y)/64
	textBottom := y + float64(textBounds.Max.Y)/64
	if textRight <= 0 || textLeft >= float64(width) || textBottom <= 0 || textTop >= float64(height) {
		return p
	}

	xPixel, err := watermarkPixelCoordinate(x)
	if err != nil {
		p.err = err
		return p
	}
	yPixel, err := watermarkPixelCoordinate(y)
	if err != nil {
		p.err = err
		return p
	}
	xDot, xShift := watermarkFixedCoordinate(xPixel)
	yDot, yShift := watermarkFixedCoordinate(yPixel)

	minX, ok := checkedSubtractInt(0, xShift)
	if !ok {
		p.err = fmt.Errorf("watermark horizontal position overflows int")
		return p
	}
	maxX, ok := checkedSubtractInt(width, xShift)
	if !ok {
		p.err = fmt.Errorf("watermark horizontal position overflows int")
		return p
	}
	minY, ok := checkedSubtractInt(0, yShift)
	if !ok {
		p.err = fmt.Errorf("watermark vertical position overflows int")
		return p
	}
	maxY, ok := checkedSubtractInt(height, yShift)
	if !ok {
		p.err = fmt.Errorf("watermark vertical position overflows int")
		return p
	}
	overlayBounds := image.Rectangle{
		Min: image.Pt(minX, minY),
		Max: image.Pt(maxX, maxY),
	}

	imgWithWatermark, err := newRGBAChecked(bounds)
	if err != nil {
		p.err = fmt.Errorf("invalid watermark image size: %w", err)
		return p
	}
	overlay, err := newRGBAChecked(overlayBounds)
	if err != nil {
		p.err = fmt.Errorf("invalid watermark overlay size: %w", err)
		return p
	}
	draw.Draw(imgWithWatermark, bounds, p.currentImage, bounds.Min, draw.Src)

	drawer := &font.Drawer{
		Dst:  overlay,
		Src:  image.NewUniform(cfg.Color),
		Face: face,
		Dot: fixed.Point26_6{
			X: xDot,
			Y: yDot,
		},
	}
	drawer.DrawString(cfg.Text)
	draw.Draw(imgWithWatermark, bounds, overlay, overlayBounds.Min, draw.Over)

	p.currentImage = imgWithWatermark
	return p
}

const maxWatermarkFontSize = (float64(1<<31) - 1) / 64

func validateWatermarkConfig(cfg WatermarkConfig) error {
	if cfg.Text == "" {
		return fmt.Errorf("watermark text cannot be empty")
	}
	if math.IsNaN(cfg.FontSize) || math.IsInf(cfg.FontSize, 0) || cfg.FontSize <= 0 {
		return fmt.Errorf("watermark font size must be finite and positive")
	}
	if cfg.FontSize > maxWatermarkFontSize {
		return fmt.Errorf("watermark font size exceeds the renderer's fixed-point range")
	}
	if isNilColor(cfg.Color) {
		return fmt.Errorf("watermark color cannot be nil")
	}
	if cfg.Position < PositionTopLeft || cfg.Position > PositionCenter {
		return fmt.Errorf("unknown watermark position: %d", cfg.Position)
	}
	if math.IsNaN(cfg.OffsetX) || math.IsInf(cfg.OffsetX, 0) || math.IsNaN(cfg.OffsetY) || math.IsInf(cfg.OffsetY, 0) {
		return fmt.Errorf("watermark offsets must be finite")
	}
	switch cfg.FontSource {
	case FontSourceBytes:
	case FontSourcePath:
		if cfg.FontPath == "" {
			return fmt.Errorf("watermark font path cannot be empty")
		}
	default:
		return fmt.Errorf("watermark font source is not configured")
	}
	return nil
}

func isNilColor(c color.Color) bool {
	if c == nil {
		return true
	}
	value := reflect.ValueOf(c)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return value.IsNil()
	default:
		return false
	}
}

func watermarkPixelCoordinate(value float64) (int, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0, fmt.Errorf("watermark position is not finite")
	}
	value = math.Trunc(value)
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	if value < float64(minInt) || value >= float64(maxInt) {
		return 0, fmt.Errorf("watermark position exceeds the image coordinate range")
	}
	return int(value), nil
}

func watermarkFixedCoordinate(pixel int) (fixed.Int26_6, int) {
	const minFixedPixel = -(1 << 25)
	const maxFixedPixel = (1 << 25) - 1
	if pixel < minFixedPixel || pixel > maxFixedPixel {
		return 0, pixel
	}
	return fixed.I(pixel), 0
}

func checkedSubtractInt(a, b int) (int, bool) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	if b > 0 && a < minInt+b {
		return 0, false
	}
	if b < 0 && a > maxInt+b {
		return 0, false
	}
	return a - b, true
}
