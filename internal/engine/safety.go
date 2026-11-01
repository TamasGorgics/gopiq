package engine

import (
	"fmt"
	"image"
)

const rgbaBytesPerPixel = 4

// checkedImageDimensions returns the extent of bounds without allowing an
// invalid or overflowing rectangle to wrap around.
func checkedImageDimensions(bounds image.Rectangle) (width, height int, err error) {
	width, err = checkedExtent(bounds.Min.X, bounds.Max.X)
	if err != nil {
		return 0, 0, err
	}
	height, err = checkedExtent(bounds.Min.Y, bounds.Max.Y)
	if err != nil {
		return 0, 0, err
	}
	return width, height, nil
}

func checkedExtent(minimum, maximum int) (int, error) {
	if maximum < minimum {
		return 0, fmt.Errorf("invalid image bounds: maximum %d is less than minimum %d", maximum, minimum)
	}

	maxInt := int(^uint(0) >> 1)
	if minimum < 0 && maximum > maxInt+minimum {
		return 0, fmt.Errorf("image bounds extent overflows int")
	}
	return maximum - minimum, nil
}

// checkedRGBABytes verifies that an RGBA image with these dimensions can be
// represented by a Go slice and returns its pixel-buffer length.
func checkedRGBABytes(width, height int) (int, error) {
	if width < 0 || height < 0 {
		return 0, fmt.Errorf("image dimensions must not be negative")
	}
	if width == 0 || height == 0 {
		return 0, nil
	}

	maxInt := int(^uint(0) >> 1)
	if width > maxInt/rgbaBytesPerPixel {
		return 0, fmt.Errorf("RGBA image buffer size overflows int")
	}
	rowBytes := width * rgbaBytesPerPixel
	if height > maxInt/rowBytes {
		return 0, fmt.Errorf("RGBA image buffer size overflows int")
	}
	return rowBytes * height, nil
}

func newRGBAChecked(bounds image.Rectangle) (*image.RGBA, error) {
	width, height, err := checkedImageDimensions(bounds)
	if err != nil {
		return nil, err
	}
	if _, err := checkedRGBABytes(width, height); err != nil {
		return nil, err
	}
	return image.NewRGBA(bounds), nil
}
