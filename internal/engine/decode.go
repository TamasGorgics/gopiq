package engine

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"math"
)

// ErrDecodeLimit identifies an image rejected by configured decode limits.
var ErrDecodeLimit = errors.New("image exceeds configured decode limits")

// DecodeLimits controls encoded-size and decoded-dimension limits.
// Zero or negative fields mean no limit for that field.
type DecodeLimits struct {
	MaxEncodedBytes int64
	MaxWidth        int
	MaxHeight       int
	MaxPixels       int64
}

func (l DecodeLimits) unlimited() bool {
	return l.MaxEncodedBytes <= 0 &&
		l.MaxWidth <= 0 &&
		l.MaxHeight <= 0 &&
		l.MaxPixels <= 0
}

func decodeImageWithLimits(r io.Reader, limits DecodeLimits) (image.Image, error) {
	if limits.unlimited() {
		return decodeImage(r)
	}

	limited := r
	var limitedReader *io.LimitedReader
	if limits.MaxEncodedBytes > 0 && limits.MaxEncodedBytes < math.MaxInt64 {
		limitedReader = &io.LimitedReader{R: r, N: limits.MaxEncodedBytes + 1}
		limited = limitedReader
	}

	var prefix bytes.Buffer
	config, _, err := image.DecodeConfig(io.TeeReader(limited, &prefix))
	if err != nil {
		if limitExceeded(limitedReader) {
			return nil, encodedLimitError(limits.MaxEncodedBytes)
		}
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	if limitExceeded(limitedReader) {
		return nil, encodedLimitError(limits.MaxEncodedBytes)
	}
	if err := validateDecodeConfig(config, limits); err != nil {
		return nil, err
	}

	img, _, err := image.Decode(io.MultiReader(bytes.NewReader(prefix.Bytes()), limited))
	if err != nil {
		if limitExceeded(limitedReader) {
			return nil, encodedLimitError(limits.MaxEncodedBytes)
		}
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	if limitExceeded(limitedReader) {
		return nil, encodedLimitError(limits.MaxEncodedBytes)
	}
	return img, nil
}

func validateDecodeConfig(config image.Config, limits DecodeLimits) error {
	if config.Width <= 0 || config.Height <= 0 {
		return fmt.Errorf("invalid image dimensions %dx%d", config.Width, config.Height)
	}
	if limits.MaxWidth > 0 && config.Width > limits.MaxWidth {
		return limitError("image width %d exceeds limit %d", config.Width, limits.MaxWidth)
	}
	if limits.MaxHeight > 0 && config.Height > limits.MaxHeight {
		return limitError("image height %d exceeds limit %d", config.Height, limits.MaxHeight)
	}
	if limits.MaxPixels > 0 {
		if int64(config.Width) > math.MaxInt64/int64(config.Height) {
			return limitError("image pixel count overflows int64")
		}
		pixels := int64(config.Width) * int64(config.Height)
		if pixels > limits.MaxPixels {
			return limitError("image pixel count %d exceeds limit %d", pixels, limits.MaxPixels)
		}
	}
	return nil
}

func limitExceeded(r *io.LimitedReader) bool {
	return r != nil && r.N == 0
}

func encodedLimitError(limit int64) error {
	return limitError("encoded input exceeds limit of %d bytes", limit)
}

func limitError(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrDecodeLimit, fmt.Sprintf(format, args...))
}
