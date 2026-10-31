# Watermark Options

Use these functional options to customize text watermarks.

- `WithFontPath(path string)` - Load a `.ttf` or `.otf` font file.
- `WithFontBytes(data []byte)` - Use font data directly.
- `WithFontSize(size float64)` - Set a finite, positive font size within the renderer's fixed-point range.
- `WithColor(color color.Color)` - Set a non-nil text color.
- `WithPosition(pos WatermarkPosition)` - Select one of the five named anchors.
- `WithOffset(x, y float64)` - Set finite pixel offsets. Positive values inset corner anchors; at the center they move right/down. Negative values are allowed and can move text outside the image.

The last supplied font-source option (`WithFontPath` or `WithFontBytes`) wins. If neither is supplied, gopiq uses its embedded Go font. A missing or invalid font, invalid option, empty text, or empty destination image sets the processor's chain error.

Named positions are calculated relative to the image bounds, including non-zero origins. Text that extends beyond the image is clipped rather than moved back into view. Watermarking draws onto a copy and leaves the source image unchanged.
