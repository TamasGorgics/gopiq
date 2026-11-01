package engine

import (
	"errors"
	"image"
	"testing"
)

func TestValidateDecodeConfigLimits(t *testing.T) {
	config := image.Config{Width: 101, Height: 81}

	tests := []struct {
		name   string
		config image.Config
		limits DecodeLimits
		want   bool
	}{
		{name: "within limits", config: config, limits: DecodeLimits{MaxWidth: 101, MaxHeight: 81, MaxPixels: 8_181}},
		{name: "width", config: config, limits: DecodeLimits{MaxWidth: 100}, want: true},
		{name: "height", config: config, limits: DecodeLimits{MaxHeight: 80}, want: true},
		{name: "pixels", config: config, limits: DecodeLimits{MaxPixels: 8_180}, want: true},
		{
			name:   "pixel overflow",
			config: image.Config{Width: int(^uint(0) >> 1), Height: 2},
			limits: DecodeLimits{MaxPixels: 1},
			want:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDecodeConfig(tc.config, tc.limits)
			if tc.want != errors.Is(err, ErrDecodeLimit) {
				t.Fatalf("validateDecodeConfig error = %v, want limit = %t", err, tc.want)
			}
		})
	}

	if err := validateDecodeConfig(image.Config{Width: 0, Height: 1}, DecodeLimits{}); err == nil {
		t.Fatal("validateDecodeConfig accepted empty dimensions")
	}
}
