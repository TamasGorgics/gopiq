# API Reference

This section covers the core methods for creating and managing `ImageProcessor` instances.

- `New(img image.Image) *ImageProcessor` - Create processor from image
- `FromBytes(data []byte) *ImageProcessor` - Decode JPEG, PNG, or GIF bytes into a processor
- `Clone() *ImageProcessor` - Create independent copy
- `Image() (image.Image, error)` - Get current image
- `ToBytes(format ImageFormat) ([]byte, error)` - Export to bytes
- `Err() error` - Get any error from the processing chain

`ToBytes` supports JPEG and PNG output. GIF is decode-only; attempting to encode GIF returns an error.
