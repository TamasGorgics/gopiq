package engine

import (
	"testing"

	"golang.org/x/image/math/fixed"
)

func TestWatermarkFixedPointCoordinateTranslation(t *testing.T) {
	const fixedPixelMax = (1 << 25) - 1
	for _, pixel := range []int{fixedPixelMax + 1, -fixedPixelMax - 2} {
		dot, shift := watermarkFixedCoordinate(pixel)
		if dot != 0 || shift != pixel {
			t.Errorf("watermarkFixedCoordinate(%d) = (%d, %d), want (0, %d)", pixel, dot, shift, pixel)
		}
	}
	for _, pixel := range []int{0, fixedPixelMax, -1 << 25} {
		dot, shift := watermarkFixedCoordinate(pixel)
		if want := fixed.I(pixel); dot != want || shift != 0 {
			t.Errorf("watermarkFixedCoordinate(%d) = (%d, %d), want (%d, 0)", pixel, dot, shift, want)
		}
	}
}
