# Processing Operations

These methods are chainable and perform image manipulation.

- `Resize(width, height int)` - Resize a non-empty image using Catmull-Rom interpolation; invalid or unrepresentable sizes set a chain error.
- `Crop(x, y, width, height int)` - Crop using source-image coordinates; the result has zero-origin bounds.
- `Grayscale()` - Convert to grayscale; empty images are left unchanged.
- `GrayscaleFast()` - Convert to grayscale using parallel processing when enabled and above the configured threshold; empty images are left unchanged.
- `AddTextWatermark(text, ...options)` - Add text watermark; an empty destination image returns a chain error.

Crop and resize reject invalid dimensions. Cropping uses absolute coordinates from the source image bounds and returns a chain error for empty bounds. Resizing an image with empty bounds also returns a chain error.
