package gopiq

import (
	"bytes"
	"errors"
	"image"
	"io"
	"testing"
	"testing/iotest"
)

type m6ErrorWriter struct{}

func (m6ErrorWriter) Write([]byte) (int, error) {
	return 0, errors.New("forced write failure")
}

func TestM6FromReaderAndToWriter(t *testing.T) {
	source := createTestImage(32, 24)
	data, err := imageToPNGBytes(source)
	if err != nil {
		t.Fatalf("encode test image: %v", err)
	}

	processor := FromReader(iotest.OneByteReader(bytes.NewReader(data)))
	if err := processor.Err(); err != nil {
		t.Fatalf("FromReader failed: %v", err)
	}
	decoded, err := processor.Image()
	if err != nil {
		t.Fatalf("Image failed: %v", err)
	}
	if got, want := decoded.Bounds().Size(), image.Pt(32, 24); got != want {
		t.Fatalf("decoded size = %v, want %v", got, want)
	}

	var streamed bytes.Buffer
	if err := processor.ToWriter(&streamed, FormatPNG); err != nil {
		t.Fatalf("ToWriter failed: %v", err)
	}
	inMemory, err := processor.ToBytes(FormatPNG)
	if err != nil {
		t.Fatalf("ToBytes failed: %v", err)
	}
	if !bytes.Equal(streamed.Bytes(), inMemory) {
		t.Fatal("ToWriter and ToBytes produced different PNG output")
	}
}

func TestM6DecodeLimits(t *testing.T) {
	source := image.NewRGBA(image.Rect(0, 0, 100, 80))
	fillM1Image(source)
	data, err := imageToPNGBytes(source)
	if err != nil {
		t.Fatalf("encode test image: %v", err)
	}

	tests := []struct {
		name   string
		limits DecodeLimits
	}{
		{name: "encoded bytes", limits: DecodeLimits{MaxEncodedBytes: int64(len(data) - 1)}},
		{name: "width", limits: DecodeLimits{MaxWidth: 99}},
		{name: "height", limits: DecodeLimits{MaxHeight: 79}},
		{name: "pixels", limits: DecodeLimits{MaxPixels: 7_999}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			processor := FromReaderWithLimits(bytes.NewReader(data), tc.limits)
			if !errors.Is(processor.Err(), ErrDecodeLimit) {
				t.Fatalf("FromReaderWithLimits error = %v, want ErrDecodeLimit", processor.Err())
			}

			byteProcessor := FromBytesWithLimits(data, tc.limits)
			if !errors.Is(byteProcessor.Err(), ErrDecodeLimit) {
				t.Fatalf("FromBytesWithLimits error = %v, want ErrDecodeLimit", byteProcessor.Err())
			}
		})
	}

	if processor := FromReaderWithLimits(bytes.NewReader(data), DecodeLimits{}); processor.Err() != nil {
		t.Fatalf("zero limits should be unlimited: %v", processor.Err())
	}
	if processor := FromBytesWithLimits(data, DecodeLimits{}); processor.Err() != nil {
		t.Fatalf("zero byte limits should be unlimited: %v", processor.Err())
	}
}

func TestM6DecodeErrorsAndOwnership(t *testing.T) {
	if processor := FromReader(nil); processor.Err() == nil {
		t.Fatal("FromReader accepted a nil reader")
	}
	if processor := FromBytesWithLimits(nil, DefaultDecodeLimits()); processor.Err() == nil {
		t.Fatal("FromBytesWithLimits accepted empty data")
	}

	malformed := FromReaderWithLimits(bytes.NewReader([]byte("not an image")), DefaultDecodeLimits())
	if malformed.Err() == nil {
		t.Fatal("malformed input did not produce an error")
	}
	if errors.Is(malformed.Err(), ErrDecodeLimit) {
		t.Fatal("malformed input was reported as a decode-limit error")
	}

	limits := DefaultDecodeLimits()
	if limits.MaxEncodedBytes <= 0 || limits.MaxWidth <= 0 || limits.MaxHeight <= 0 || limits.MaxPixels <= 0 {
		t.Fatalf("DefaultDecodeLimits returned a non-positive limit: %+v", limits)
	}
}

func TestM6ToWriterErrors(t *testing.T) {
	processor := New(createTestImage(8, 6))
	if err := processor.ToWriter(nil, FormatPNG); err == nil {
		t.Fatal("ToWriter accepted a nil writer")
	}
	if err := processor.ToWriter(m6ErrorWriter{}, FormatPNG); err == nil {
		t.Fatal("ToWriter ignored a writer error")
	}
	if err := processor.ToWriter(io.Discard, FormatGIF); err == nil {
		t.Fatal("ToWriter accepted unsupported GIF output")
	}

	broken := New(nil)
	if err := broken.ToWriter(io.Discard, FormatPNG); err == nil {
		t.Fatal("ToWriter on an errored processor succeeded")
	}
}
