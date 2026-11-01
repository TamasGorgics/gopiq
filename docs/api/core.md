# API Reference

This section covers the core methods for creating and managing `ImageProcessor` instances.

- `New(img image.Image) *ImageProcessor` - Create processor from image; the image is retained, not copied
- `FromBytes(data []byte) *ImageProcessor` - Decode JPEG, PNG, or GIF bytes using legacy unlimited decoding
- `FromBytesWithLimits(data []byte, limits DecodeLimits) *ImageProcessor` - Decode bytes with explicit limits
- `FromReader(r io.Reader) *ImageProcessor` - Decode from a reader using `DefaultDecodeLimits`
- `FromReaderWithLimits(r io.Reader, limits DecodeLimits) *ImageProcessor` - Decode from a reader with explicit limits
- `Clone() *ImageProcessor` - Create a shallow processor copy; error and performance state are copied, image storage is shared
- `Image() (image.Image, error)` - Get the current image for reading
- `ToBytes(format ImageFormat) ([]byte, error)` - Export to an in-memory byte slice
- `ToWriter(w io.Writer, format ImageFormat) error` - Encode directly to a writer
- `Err() error` - Get any error from the processing chain

Methods are individually synchronized, but a multi-call chain is not atomic. Images passed to `New` and images returned by `Image` must be treated as read-only while the processor or its clones may be used concurrently. `Clone` does not deep-copy pixels.

`ToBytes` and `ToWriter` support JPEG and PNG output. GIF is decode-only; attempting to encode GIF returns an error.

See [Streaming and Decode Limits](../streaming.md) for `DecodeLimits`, `DefaultDecodeLimits`, `ErrDecodeLimit`, ownership, and writer error behavior.
