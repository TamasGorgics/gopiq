package engine

import (
	"runtime"
	"testing"
)

func TestPerformanceThresholdAndWorkerLimits(t *testing.T) {
	defaultOptions := DefaultOptions()
	if shouldUseParallelProcessing(9_999, defaultOptions) {
		t.Fatal("default options selected parallel processing below their threshold")
	}
	if !shouldUseParallelProcessing(10_000, defaultOptions) {
		t.Fatal("default options did not select parallel processing at their threshold")
	}
	if !shouldUseParallelProcessing(1, Options{EnableParallelProcessing: true, MinSizeForParallel: 0}) {
		t.Fatal("zero threshold should allow parallel processing for a non-empty image")
	}
	if !shouldUseParallelProcessing(1, Options{EnableParallelProcessing: true, MinSizeForParallel: -1}) {
		t.Fatal("negative threshold should allow parallel processing for a non-empty image")
	}
	if shouldUseParallelProcessing(100, Options{EnableParallelProcessing: false, MinSizeForParallel: 0}) {
		t.Fatal("disabled parallel processing selected the parallel path")
	}

	if got := grayscaleWorkerCount(0, 4); got != 0 {
		t.Fatalf("worker count for empty image = %d, want 0", got)
	}
	if got := grayscaleWorkerCount(1, 8); got != 1 {
		t.Fatalf("worker count for one row = %d, want 1", got)
	}
	if got := grayscaleWorkerCount(8, 3); got != 3 {
		t.Fatalf("worker count with explicit limit = %d, want 3", got)
	}
	defaultWorkers := runtime.NumCPU()
	if defaultWorkers > 8 {
		defaultWorkers = 8
	}
	if got := grayscaleWorkerCount(8, 0); got != defaultWorkers {
		t.Fatalf("worker count for non-positive limit = %d, want %d", got, defaultWorkers)
	}
}
