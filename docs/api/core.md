# API Reference

This section covers the core methods for creating and managing `ImageProcessor` instances.

- `New(img image.Image) *ImageProcessor` - Create processor from image
- `FromBytes(data []byte) *ImageProcessor` - Create processor from image bytes
- `Clone() *ImageProcessor` - Create independent copy
- `Image() (image.Image, error)` - Get current image
- `ToBytes(format ImageFormat) ([]byte, error)` - Export to bytes
- `Err() error` - Get any error from the processing chain 