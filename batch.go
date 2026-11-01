package gopiq

import (
	"context"
	"fmt"
	"image"

	"github.com/TamasGorgics/gopiq/internal/engine"
)

// BatchOptions controls bounded batch execution.
type BatchOptions struct {
	// MaxWorkers limits concurrent operations. Values less than or equal to
	// zero use runtime.NumCPU(); the count is capped by the number of inputs.
	MaxWorkers int
}

// ProcessBatch processes images concurrently with bounded workers and returns
// results in input order.
//
// Each input is passed to a separate ImageProcessor and then to operation.
// Operation calls may run concurrently and must be safe for shared state. Input
// images are read-only. Per-image errors remain available through each result's
// Err method; they do not stop unrelated items.
//
// Cancellation stops scheduling new items, waits for in-flight operations to
// finish, and returns partial results with ctx.Err(). Unstarted result slots
// remain nil. An empty input returns an empty result slice and nil error.
func ProcessBatch(ctx context.Context, images []image.Image, operation func(*ImageProcessor) *ImageProcessor, options BatchOptions) ([]*ImageProcessor, error) {
	if ctx == nil {
		return nil, fmt.Errorf("context cannot be nil")
	}
	if operation == nil {
		return nil, fmt.Errorf("batch operation cannot be nil")
	}

	results := make([]*ImageProcessor, len(images))
	if len(images) == 0 {
		return results, nil
	}

	err := engine.RunBatch(ctx, len(images), options.MaxWorkers, func(index int) {
		results[index] = operation(New(images[index]))
	})
	return results, err
}
