package processor

import (
	"context"
	"io"
)

// ImageProcessor defines the core interface for image processing operations
type ImageProcessor interface {
	// Process handles a single image processing request
	Process(ctx context.Context, input ProcessInput) (ProcessOutput, error)

	// ProcessBatch handles multiple image processing requests
	ProcessBatch(ctx context.Context, inputs []ProcessInput) ([]ProcessOutput, error)

	// AddHandler adds a processing handler to the pipeline
	AddHandler(handler Handler)

	// SetProgressTracker sets the progress tracking mechanism
	SetProgressTracker(tracker ProgressTracker)
}

// Handler defines the interface for individual processing steps
type Handler interface {
	// Handle processes a single input and returns the result
	Handle(ctx context.Context, input []byte) ([]byte, error)

	// GetName returns the handler's name for logging and metrics
	GetName() string
}

// ProgressTracker defines the interface for tracking processing progress
type ProgressTracker interface {
	// Update updates the current progress
	Update(current, total int64)

	// Finish marks the processing as complete
	Finish()

	// Error reports an error in processing
	Error(err error)
}

// ProcessInput represents the input for image processing
type ProcessInput struct {
	// Reader provides the image data
	Reader io.Reader

	// Filename is the original filename
	Filename string

	// Metadata contains additional processing instructions
	Metadata map[string]interface{}
}

// ProcessOutput represents the result of image processing
type ProcessOutput struct {
	// Data contains the processed image data
	Data []byte

	// Filename is the output filename
	Filename string

	// Labels contains vision API labels
	Labels []Label

	// Error contains any processing error
	Error error

	// Metadata contains additional output information
	Metadata map[string]interface{}
}

// Label represents a vision API label
type Label struct {
	Description string  `json:"description"`
	Score       float64 `json:"score"`
}

// Options contains configuration for the processor
type Options struct {
	// MaxRetries specifies the maximum number of retries for failed operations
	MaxRetries int

	// BatchSize specifies the size of processing batches
	BatchSize int

	// Concurrent specifies whether to process images concurrently
	Concurrent bool

	// MaxConcurrent specifies the maximum number of concurrent operations
	MaxConcurrent int

	// ErrorHandler handles errors during processing
	ErrorHandler func(error)
}

// DefaultOptions returns the default processor options
func DefaultOptions() Options {
	return Options{
		MaxRetries:    3,
		BatchSize:     100,
		Concurrent:    true,
		MaxConcurrent: 8,
		ErrorHandler:  func(err error) {},
	}
}
