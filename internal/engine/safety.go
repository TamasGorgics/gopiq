package engine

import (
	"fmt"
	"image"
	"sync"
)

const rgbaBytesPerPixel = 4

// rgbaPool is a sync.Pool for reusing RGBA image buffers to reduce allocations.
var rgbaPool = sync.Pool{
	New: func() any {
		return image.NewRGBA(image.Rect(0, 0, 100, 100))
	},
}

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

func checkedExtent(min, max int) (int, error) {
	if max < min {
		return 0, fmt.Errorf("invalid image bounds: maximum %d is less than minimum %d", max, min)
	}

	maxInt := int(^uint(0) >> 1)
	if min < 0 && max > maxInt+min {
		return 0, fmt.Errorf("image bounds extent overflows int")
	}
	return max - min, nil
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

// getPooledRGBA returns an RGBA image from the pool, resized to the given bounds.
func getPooledRGBA(bounds image.Rectangle) *image.RGBA {
	width, height, err := checkedImageDimensions(bounds)
	if err != nil {
		return nil
	}
	pixelsNeeded, err := checkedRGBABytes(width, height)
	if err != nil {
		return nil
	}

	img := rgbaPool.Get().(*image.RGBA)
	if img.Bounds().Dx() < width || img.Bounds().Dy() < height {
		fresh, allocErr := newRGBAChecked(bounds)
		if allocErr != nil {
			rgbaPool.Put(img)
			return nil
		}
		img = fresh
	} else {
		img.Rect = bounds
		if len(img.Pix) < pixelsNeeded {
			img.Pix = make([]uint8, pixelsNeeded)
		} else {
			for i := 0; i < pixelsNeeded; i++ {
				img.Pix[i] = 0
			}
		}
		img.Stride = 4 * width
	}

	return img
}

// returnPooledRGBA returns an RGBA image to the pool for reuse.
func returnPooledRGBA(img *image.RGBA) {
	if img == nil {
		return
	}
	width, height, err := checkedImageDimensions(img.Bounds())
	if err != nil {
		return
	}
	bufferBytes, err := checkedRGBABytes(width, height)
	if err != nil {
		return
	}

	if bufferBytes/rgbaBytesPerPixel <= 2000*2000 {
		rgbaPool.Put(img)
	}
}
