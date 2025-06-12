# Processing Operations

These methods are chainable and perform image manipulation.

- `Resize(width, height int)` - Resize using Catmull-Rom interpolation
- `Crop(x, y, width, height int)` - Crop to specified rectangle
- `Grayscale()` - Convert to grayscale
- `GrayscaleFast()` - Convert to grayscale using parallel processing for a significant speed boost.
- `AddTextWatermark(text, ...options)` - Add text watermark 