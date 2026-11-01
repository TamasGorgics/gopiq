package engine

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"sync"
)

// Processor holds the current image, chain error, and performance options.
// Its methods are individually synchronized.
type Processor struct {
	mu           sync.RWMutex
	currentImage image.Image
	err          error
	options      Options
}

// New creates a Processor from an existing image. The image is retained, not
// copied, and callers must treat it as read-only while the processor is in use.
func New(img image.Image, opts Options) *Processor {
	if img == nil {
		return &Processor{err: fmt.Errorf("initial image cannot be nil")}
	}
	return &Processor{
		currentImage: img,
		options:      opts,
	}
}

// NewUninitialized creates an empty Processor. It is used for the zero value of
// the public facade so that operations behave like an empty processor.
func NewUninitialized(opts Options) *Processor {
	return &Processor{options: opts}
}

// FromBytes decodes an image from a byte slice without decode limits.
func FromBytes(data []byte, opts Options) *Processor {
	return FromBytesWithLimits(data, DecodeLimits{}, opts)
}

// FromBytesWithLimits decodes an image from a byte slice with decode limits.
func FromBytesWithLimits(data []byte, limits DecodeLimits, opts Options) *Processor {
	if len(data) == 0 {
		return &Processor{err: fmt.Errorf("input byte slice is empty")}
	}
	if limits.MaxEncodedBytes > 0 && int64(len(data)) > limits.MaxEncodedBytes {
		return &Processor{err: encodedLimitError(limits.MaxEncodedBytes)}
	}
	return FromReader(bytes.NewReader(data), limits, opts)
}

// FromReader decodes an image from a reader with decode limits.
// The reader is consumed but not closed.
func FromReader(r io.Reader, limits DecodeLimits, opts Options) *Processor {
	if r == nil {
		return &Processor{err: fmt.Errorf("reader cannot be nil")}
	}
	img, err := decodeImageWithLimits(r, limits)
	if err != nil {
		return &Processor{err: err}
	}
	return &Processor{
		currentImage: img,
		options:      opts,
	}
}

// SetOptions updates the processor's performance settings.
func (p *Processor) SetOptions(opts Options) *Processor {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.options = opts
	return p
}

// Options returns the processor's current performance settings.
func (p *Processor) Options() Options {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.options
}

// ToBytes encodes the current image in the requested format.
func (p *Processor) ToBytes(format Format) ([]byte, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.err != nil {
		return nil, p.err
	}
	if p.currentImage == nil {
		return nil, fmt.Errorf("no image available to convert to bytes")
	}

	var buf bytes.Buffer
	if err := encodeImage(&buf, p.currentImage, format); err != nil {
		return nil, fmt.Errorf("failed to encode image to bytes: %w", err)
	}
	return buf.Bytes(), nil
}

// ToWriter encodes the current image directly to a writer.
func (p *Processor) ToWriter(w io.Writer, format Format) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.err != nil {
		return p.err
	}
	if p.currentImage == nil {
		return fmt.Errorf("no image available to encode")
	}
	if w == nil {
		return fmt.Errorf("writer cannot be nil")
	}
	if err := encodeImage(w, p.currentImage, format); err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}
	return nil
}

// Image returns the current image and chain error.
func (p *Processor) Image() (image.Image, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentImage, p.err
}

// Err returns the first error in the chain.
func (p *Processor) Err() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.err
}

// Clone creates a shallow copy sharing the current image reference.
func (p *Processor) Clone() *Processor {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return &Processor{
		currentImage: p.currentImage,
		err:          p.err,
		options:      p.options,
	}
}
