# Streaming and Decode Limits

`FromReader` and `ToWriter` let applications decode and encode without building an intermediate byte slice. The existing `FromBytes` and `ToBytes` APIs remain available.

## Reading from a reader

```go
file, err := os.Open("input.jpg")
if err != nil {
    return err
}
defer file.Close()

processor := gopiq.FromReader(file)
if err := processor.Err(); err != nil {
    return err
}
```

`FromReader` consumes the reader but does not close it. The caller retains ownership.

`FromReader` applies `DefaultDecodeLimits`:

- 64 MiB of encoded input
- 16,384 pixels in either dimension
- 100 million total pixels

Use `FromReaderWithLimits` to tune those limits:

```go
limits := gopiq.DecodeLimits{
    MaxEncodedBytes: 16 << 20,
    MaxWidth:        8192,
    MaxHeight:       8192,
    MaxPixels:       25_000_000,
}

processor := gopiq.FromReaderWithLimits(file, limits)
```

Zero or negative fields disable that individual limit. `DecodeLimits{}` therefore means unlimited.

## Legacy byte input

`FromBytes` retains its previous unlimited behavior:

```go
processor := gopiq.FromBytes(data)
```

For untrusted byte input, choose explicit limits:

```go
processor := gopiq.FromBytesWithLimits(data, gopiq.DefaultDecodeLimits())
```

## Decode-limit errors

Limit failures wrap `ErrDecodeLimit`:

```go
if errors.Is(processor.Err(), gopiq.ErrDecodeLimit) {
    // The input exceeded a configured size or dimension limit.
}
```

Malformed input produces a normal decode error, not `ErrDecodeLimit`.

Encoded-size checks are applied while reading. Dimension checks use the decoded image configuration before the full pixel decode where the image format allows it.

## Writing to a writer

```go
output, err := os.Create("output.png")
if err != nil {
    return err
}
defer output.Close()

if err := processor.ToWriter(output, gopiq.FormatPNG); err != nil {
    return err
}
```

`ToWriter` writes directly to the supplied writer and does not close it. It does not buffer the complete encoded image. A write failure may leave partially written output; use `ToBytes` when an in-memory result is required.

`ToWriter` supports JPEG, PNG, and GIF output. GIF output uses the Go standard-library Plan 9 256-color palette with Floyd–Steinberg dithering, flattens alpha over white, and writes a single frame.

## Ownership and resources

- `FromReader` consumes but does not close its reader.
- `ToWriter` writes to but does not close its writer.
- Decode limits are per construction call; there is no global limit state.
- `FromBytes` and `FromReader` never mutate caller-owned input.
- The library allocates decoded pixel storage; decode limits are the mechanism for bounding it.
