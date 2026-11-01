package engine

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
)

// RunBatch invokes operation for count indexes using at most maxWorkers
// concurrent goroutines.
func RunBatch(ctx context.Context, count, maxWorkers int, operation func(int)) error {
	if ctx == nil {
		return fmt.Errorf("context cannot be nil")
	}
	if operation == nil {
		return fmt.Errorf("batch operation cannot be nil")
	}
	if count <= 0 {
		return nil
	}
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}
	if maxWorkers < 1 {
		maxWorkers = 1
	}
	if maxWorkers > count {
		maxWorkers = count
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	var next atomic.Int64
	var wg sync.WaitGroup
	for range maxWorkers {
		wg.Go(func() {
			runBatchWorker(ctx, count, &next, operation)
		})
	}

	wg.Wait()
	return ctx.Err()
}

func runBatchWorker(ctx context.Context, count int, next *atomic.Int64, operation func(int)) {
	for {
		if ctx.Err() != nil {
			return
		}
		index := int(next.Add(1)) - 1
		if index >= count {
			return
		}
		operation(index)
	}
}
