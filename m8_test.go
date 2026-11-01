package gopiq

import (
	"context"
	"errors"
	"image"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func m8Images(count int) []image.Image {
	images := make([]image.Image, count)
	for i := range images {
		images[i] = image.NewRGBA(image.Rect(i, 0, i+4, 3))
	}
	return images
}

func TestM8ProcessBatchPreservesOrder(t *testing.T) {
	images := m8Images(8)
	results, err := ProcessBatch(context.Background(), images, func(processor *ImageProcessor) *ImageProcessor {
		return processor
	}, BatchOptions{MaxWorkers: 4})
	if err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}
	for index, result := range results {
		if result == nil {
			t.Fatalf("result %d is nil", index)
		}
		img, err := result.Image()
		if err != nil {
			t.Fatalf("result %d Image: %v", index, err)
		}
		if img != images[index] {
			t.Fatalf("result %d is out of order", index)
		}
	}
}

func TestM8ProcessBatchBoundsWorkers(t *testing.T) {
	images := m8Images(8)
	started := make(chan struct{}, len(images))
	release := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })

	done := make(chan error, 1)
	go func() {
		_, err := ProcessBatch(context.Background(), images, func(processor *ImageProcessor) *ImageProcessor {
			started <- struct{}{}
			<-release
			return processor
		}, BatchOptions{MaxWorkers: 2})
		done <- err
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("expected two workers to start")
		}
	}
	select {
	case <-started:
		t.Fatal("more workers started than MaxWorkers")
	case <-time.After(50 * time.Millisecond):
	}

	releaseOnce.Do(func() { close(release) })
	if err := <-done; err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}
}

func TestM8ProcessBatchCancellation(t *testing.T) {
	images := m8Images(4)

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	var called atomic.Int32
	results, err := ProcessBatch(canceled, images, func(processor *ImageProcessor) *ImageProcessor {
		called.Add(1)
		return processor
	}, BatchOptions{MaxWorkers: 2})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("pre-canceled error = %v, want context.Canceled", err)
	}
	if called.Load() != 0 {
		t.Fatalf("pre-canceled batch called operation %d times", called.Load())
	}
	if len(results) != len(images) {
		t.Fatalf("pre-canceled results length = %d, want %d", len(results), len(images))
	}

	inFlight, cancelInFlight := context.WithCancel(context.Background())
	started := make(chan struct{}, 1)
	release := make(chan struct{})
	type outcome struct {
		results []*ImageProcessor
		err     error
	}
	done := make(chan outcome, 1)
	go func() {
		results, err := ProcessBatch(inFlight, images, func(processor *ImageProcessor) *ImageProcessor {
			started <- struct{}{}
			<-release
			return processor
		}, BatchOptions{MaxWorkers: 1})
		done <- outcome{results: results, err: err}
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("batch did not start")
	}
	cancelInFlight()
	close(release)
	result := <-done
	if !errors.Is(result.err, context.Canceled) {
		t.Fatalf("in-flight cancellation error = %v, want context.Canceled", result.err)
	}
	if len(result.results) != len(images) {
		t.Fatalf("in-flight results length = %d, want %d", len(result.results), len(images))
	}
	if result.results[0] == nil {
		t.Fatal("in-flight first item was not completed")
	}
	if result.results[1] != nil {
		t.Fatal("in-flight batch started another item after cancellation")
	}
}

func TestM8ProcessBatchPerItemErrors(t *testing.T) {
	images := m8Images(2)
	results, err := ProcessBatch(context.Background(), images, func(processor *ImageProcessor) *ImageProcessor {
		if processor.testImage() == images[0] {
			processor.Resize(0, 0)
		}
		return processor
	}, BatchOptions{MaxWorkers: 1})
	if err != nil {
		t.Fatalf("ProcessBatch failed: %v", err)
	}
	if results[0] == nil || results[0].Err() == nil {
		t.Fatal("first item did not retain its processing error")
	}
	if results[1] == nil || results[1].Err() != nil {
		t.Fatalf("second item error = %v, want nil", results[1].Err())
	}
}

func TestM8ProcessBatchValidationAndEmpty(t *testing.T) {
	var nilContext context.Context
	if _, err := ProcessBatch(nilContext, m8Images(1), func(processor *ImageProcessor) *ImageProcessor { return processor }, BatchOptions{}); err == nil {
		t.Fatal("ProcessBatch accepted a nil context")
	}
	if _, err := ProcessBatch(context.Background(), m8Images(1), nil, BatchOptions{}); err == nil {
		t.Fatal("ProcessBatch accepted a nil operation")
	}

	results, err := ProcessBatch(context.Background(), m8Images(2), func(processor *ImageProcessor) *ImageProcessor {
		return processor
	}, BatchOptions{})
	if err != nil {
		t.Fatalf("default-worker ProcessBatch failed: %v", err)
	}
	if len(results) != 2 || results[0] == nil || results[1] == nil {
		t.Fatalf("default-worker results = %#v", results)
	}

	results, err = ProcessBatch(context.Background(), nil, func(processor *ImageProcessor) *ImageProcessor { return processor }, BatchOptions{MaxWorkers: 2})
	if err != nil {
		t.Fatalf("empty ProcessBatch failed: %v", err)
	}
	if results == nil || len(results) != 0 {
		t.Fatalf("empty ProcessBatch results = %#v, want non-nil empty slice", results)
	}
}
