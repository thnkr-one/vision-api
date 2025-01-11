package processor

import (
	"fmt"
	"time"

	"../../pkg/vision"
	"../image"
)

// ProcessorOptions contains configuration for the image processor
type ProcessorOptions struct {
	// PoolSize is the number of concurrent processors
	PoolSize int

	// BatchSize is the number of images to process in a batch
	BatchSize int

	// RetryAttempts is the maximum number of retry attempts
	RetryAttempts int

	// RetryDelay is the initial delay between retries
	RetryDelay time.Duration

	// MaxRetryDelay is the maximum delay between retries
	MaxRetryDelay time.Duration

	// ImageHandler handles image processing operations
	ImageHandler image.Handler

	// VisionClient is the client for the Vision API
	VisionClient *vision.Client

	// MaxFileSize is the maximum file size in bytes
	MaxFileSize int64

	// OutputDir is the directory for processed outputs
	OutputDir string

	// TempDir is the directory for temporary files
	TempDir string

	// DeleteTempFiles determines if temporary files should be deleted
	DeleteTempFiles bool

	// AllowedFormats is a list of allowed image formats
	AllowedFormats []string
}

// OptionFunc is a function that configures Options
type OptionFunc func(*Options)

// defaultOptions returns the default processor options
func defaultOptions() *Options {
	return &Options{
		PoolSize:        4,
		BatchSize:       100,
		RetryAttempts:   3,
		RetryDelay:      time.Second,
		MaxRetryDelay:   time.Second * 30,
		MaxFileSize:     40 * 1024 * 1024, // 40MB
		DeleteTempFiles: true,
		AllowedFormats:  []string{"jpg", "jpeg", "png", "gif", "bmp"},
	}
}

// WithPoolSize sets the number of concurrent processors
func WithPoolSize(size int) OptionFunc {
	return func(o *Options) {
		if size > 0 {
			o.PoolSize = size
		}
	}
}

// WithBatchSize sets the batch size
func WithBatchSize(size int) OptionFunc {
	return func(o *Options) {
		if size > 0 {
			o.BatchSize = size
		}
	}
}

// WithRetryAttempts sets the maximum retry attempts
func WithRetryAttempts(attempts int) OptionFunc {
	return func(o *Options) {
		if attempts >= 0 {
			o.RetryAttempts = attempts
		}
	}
}

// WithRetryDelay sets the retry delay
func WithRetryDelay(delay time.Duration) OptionFunc {
	return func(o *Options) {
		if delay > 0 {
			o.RetryDelay = delay
		}
	}
}

// WithMaxRetryDelay sets the maximum retry delay
func WithMaxRetryDelay(delay time.Duration) OptionFunc {
	return func(o *Options) {
		if delay > 0 {
			o.MaxRetryDelay = delay
		}
	}
}

// WithImageHandler sets the image handler
func WithImageHandler(handler image.Handler) OptionFunc {
	return func(o *Options) {
		o.ImageHandler = handler
	}
}

// WithVisionClient sets the vision client
func WithVisionClient(client *vision.Client) OptionFunc {
	return func(o *Options) {
		o.VisionClient = client
	}
}

// WithMaxFileSize sets the maximum file size
func WithMaxFileSize(size int64) OptionFunc {
	return func(o *Options) {
		if size > 0 {
			o.MaxFileSize = size
		}
	}
}

// WithOutputDir sets the output directory
func WithOutputDir(dir string) OptionFunc {
	return func(o *Options) {
		o.OutputDir = dir
	}
}

// WithTempDir sets the temporary directory
func WithTempDir(dir string) OptionFunc {
	return func(o *Options) {
		o.TempDir = dir
	}
}

// WithDeleteTempFiles sets whether to delete temporary files
func WithDeleteTempFiles(delete bool) OptionFunc {
	return func(o *Options) {
		o.DeleteTempFiles = delete
	}
}

// WithAllowedFormats sets the allowed image formats
func WithAllowedFormats(formats []string) OptionFunc {
	return func(o *Options) {
		if len(formats) > 0 {
			o.AllowedFormats = formats
		}
	}
}

// validate checks if the options are valid
func (o *ProcessorOptions) validate() error {
	if o.PoolSize < 1 {
		return fmt.Errorf("pool size must be at least 1")
	}

	if o.BatchSize < 1 {
		return fmt.Errorf("batch size must be at least 1")
	}

	if o.RetryAttempts < 0 {
		return fmt.Errorf("retry attempts cannot be negative")
	}

	if o.RetryDelay < 0 {
		return fmt.Errorf("retry delay cannot be negative")
	}

	if o.MaxRetryDelay < o.RetryDelay {
		return fmt.Errorf("maximum retry delay must be greater than or equal to retry delay")
	}

	if o.ImageHandler == nil {
		return fmt.Errorf("image handler is required")
	}

	if o.VisionClient == nil {
		return fmt.Errorf("vision client is required")
	}

	if o.MaxFileSize < 1 {
		return fmt.Errorf("max file size must be at least 1 byte")
	}

	if o.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}

	if len(o.AllowedFormats) == 0 {
		return fmt.Errorf("at least one allowed format is required")
	}

	return nil
}
