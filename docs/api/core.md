# API Reference

This section covers the core methods for creating and managing `ImageProcessor` instances.

- `New(img image.Image) *ImageProcessor` - Create processor from image; the image is retained, not copied
- `FromBytes(data []byte) *ImageProcessor` - Decode JPEG, PNG, or GIF bytes into a processor
- `Clone() *ImageProcessor` - Create a shallow processor copy; error and performance state are copied, image storage is shared
- `Image() (image.Image, error)` - Get the current image for reading
- `ToBytes(format ImageFormat) ([]byte, error)` - Export to bytes
- `Err() error` - Get any error from the processing chain

Methods are individually synchronized, but a multi-call chain is not atomic. Images passed to `New` and images returned by `Image` must be treated as read-only while the processor or its clones may be used concurrently. `Clone` does not deep-copy pixels.

`ToBytes` supports JPEG and PNG output. GIF is decode-only; attempting to encode GIF returns an error.
