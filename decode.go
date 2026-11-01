package gopiq

import (
	"io"

	"github.com/TamasGorgics/gopiq/internal/engine"
)

const (
	defaultMaxEncodedBytes int64 = 64 << 20
	defaultMaxWidth              = 16_384
	defaultMaxHeight             = 16_384
	defaultMaxPixels       int64 = 100_000_000
)

// DecodeLimits controls encoded-size and decoded-dimension limits.
// Zero or negative fields mean no limit for that field.
type DecodeLimits struct {
	// MaxEncodedBytes limits encoded input bytes. Zero or negative means no limit.
	MaxEncodedBytes int64
	// MaxWidth limits decoded image width. Zero or negative means no limit.
	MaxWidth int
	// MaxHeight limits decoded image height. Zero or negative means no limit.
	MaxHeight int
	// MaxPixels limits decoded width * height. Zero or negative means no limit.
	MaxPixels int64
}

// DefaultDecodeLimits returns the recommended limits used by FromReader.
//
// The defaults are 64 MiB of encoded input, 16,384 pixels in either
// dimension, and 100 million total pixels. They are intended for untrusted
// input; use explicit DecodeLimits to tune them.
func DefaultDecodeLimits() DecodeLimits {
	return DecodeLimits{
		MaxEncodedBytes: defaultMaxEncodedBytes,
		MaxWidth:        defaultMaxWidth,
		MaxHeight:       defaultMaxHeight,
		MaxPixels:       defaultMaxPixels,
	}
}

// ErrDecodeLimit identifies an image rejected by configured decode limits.
// Use errors.Is to detect it.
var ErrDecodeLimit = engine.ErrDecodeLimit

// FromReader decodes an image from a reader using DefaultDecodeLimits.
//
// The reader is consumed but not closed; the caller retains ownership. Use
// FromReaderWithLimits to customize or disable limits.
func FromReader(r io.Reader) *ImageProcessor {
	return FromReaderWithLimits(r, DefaultDecodeLimits())
}

// FromReaderWithLimits decodes an image from a reader using the supplied limits.
//
// Zero or negative limit fields disable that individual limit; DecodeLimits{}
// therefore means unlimited. The reader is consumed but not closed.
func FromReaderWithLimits(r io.Reader, limits DecodeLimits) *ImageProcessor {
	return &ImageProcessor{
		core: engine.FromReader(r, toEngineDecodeLimits(limits), engine.DefaultOptions()),
	}
}

// FromBytesWithLimits decodes an image from a byte slice using the supplied limits.
//
// Zero or negative limit fields disable that individual limit. Use FromBytes
// for the legacy unlimited behavior.
func FromBytesWithLimits(data []byte, limits DecodeLimits) *ImageProcessor {
	return &ImageProcessor{
		core: engine.FromBytesWithLimits(data, toEngineDecodeLimits(limits), engine.DefaultOptions()),
	}
}

func toEngineDecodeLimits(limits DecodeLimits) engine.DecodeLimits {
	return engine.DecodeLimits{
		MaxEncodedBytes: limits.MaxEncodedBytes,
		MaxWidth:        limits.MaxWidth,
		MaxHeight:       limits.MaxHeight,
		MaxPixels:       limits.MaxPixels,
	}
}
