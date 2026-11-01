# Processing Operations

These methods are chainable and perform image manipulation.

- `Resize(width, height int)` - Resize a non-empty image to exact dimensions using Catmull-Rom interpolation.
- `ResizeFit(width, height int)` - Preserve aspect ratio and fit within the target box; upscaling is allowed.
- `ResizeFill(width, height int)` - Preserve aspect ratio, cover the target box, and center-crop to exact dimensions.
- `Crop(x, y, width, height int)` - Crop using source-image coordinates; the result has zero-origin bounds.
- `FlipHorizontal()` - Mirror left-to-right.
- `FlipVertical()` - Mirror top-to-bottom.
- `Rotate90()` - Rotate clockwise 90 degrees.
- `Rotate180()` - Rotate 180 degrees.
- `Rotate270()` - Rotate clockwise 270 degrees.
- `Grayscale()` - Convert to grayscale; empty images are left unchanged.
- `GrayscaleFast()` - Convert to grayscale using parallel processing when enabled and above the configured threshold; empty images are left unchanged.
- `AddTextWatermark(text, ...options)` - Add text watermark; an empty destination image returns a chain error.

Resize operations reject non-positive target dimensions and empty sources. `ResizeFit` rounds to the nearest fitting dimensions and clamps to at least one pixel; `ResizeFill` center-crops with floor division for odd differences. Flips and rotations produce zero-origin output for non-empty images, preserve alpha, and leave empty images unchanged.

All operations treat their source image as read-only. When they succeed on non-empty input, they create a new destination image. Do not mutate an image passed to `New` or returned by `Image` while the processor or its clones may be used concurrently.
