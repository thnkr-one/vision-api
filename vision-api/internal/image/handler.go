package image

import (
	"context"
	"io"
	"time"
)

// Format represents an image format
type Format string

const (
	JPEG Format = "jpeg"
	PNG  Format = "png"
	GIF  Format = "gif"
	BMP  Format = "bmp"
	WEBP Format = "webp"
)

// Dimensions represents image dimensions
type Dimensions struct {
	Width  int
	Height int
}

// Metadata contains image metadata
type Metadata struct {
	Format     Format
	Size       int64
	Dimensions Dimensions
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Extra      map[string]interface{}
}

// ProcessOptions contains options for image processing
type ProcessOptions struct {
	MaxSize        int64
	MaxDimensions  Dimensions
	Quality        int
	PreserveFormat bool
	Format         Format
}

// ImageHandler defines the interface for image processing operations
type ImageHandler interface {
	// Process handles a single image
	Process(ctx context.Context, input io.Reader, opts ProcessOptions) (io.Reader, error)

	// GetMetadata extracts metadata from an image
	GetMetadata(ctx context.Context, input io.Reader) (*Metadata, error)

	// ValidateFormat checks if an image format is supported
	ValidateFormat(format Format) bool

	// ValidateDimensions checks if image dimensions are within limits
	ValidateDimensions(dimensions Dimensions) bool
}

// ResizeHandler defines the interface for image resizing operations
type ResizeHandler interface {
	// Resize resizes an image according to the specified dimensions
	Resize(ctx context.Context, input io.Reader, dimensions Dimensions) (io.Reader, error)

	// FitToSize fits an image within maximum dimensions while preserving aspect ratio
	FitToSize(ctx context.Context, input io.Reader, maxDimensions Dimensions) (io.Reader, error)

	// GetResizedDimensions calculates new dimensions while preserving aspect ratio
	GetResizedDimensions(current, max Dimensions) Dimensions
}

// ValidationHandler defines the interface for image validation operations
type ValidationHandler interface {
	// ValidateImage performs comprehensive image validation
	ValidateImage(ctx context.Context, input io.Reader) error

	// ValidateSize checks if image size is within limits
	ValidateSize(size int64) error

	// GetSupportedFormats returns a list of supported formats
	GetSupportedFormats() []Format
}

// CompressHandler defines the interface for image compression operations
type CompressHandler interface {
	// Compress compresses an image with the specified quality
	Compress(ctx context.Context, input io.Reader, quality int) (io.Reader, error)

	// GetOptimalQuality determines the optimal quality setting for target size
	GetOptimalQuality(currentSize, targetSize int64) int
}

// Handler combines all image handling interfaces
type Handler interface {
	ImageHandler
	ResizeHandler
	ValidationHandler
	CompressHandler
}

// Option represents a functional option for configuring handlers
type Option func(h *handlerConfig)

// handlerConfig contains common configuration for handlers
type handlerConfig struct {
	MaxImageSize    int64
	MaxDimensions   Dimensions
	DefaultQuality  int
	SupportedTypes  []Format
	PreserveFormat  bool
}

// NewHandlerConfig creates a new handler configuration with defaults
func NewHandlerConfig() *handlerConfig {
	return &handlerConfig{
		MaxImageSize:   40 * 1024 * 1024, // 40MB
		MaxDimensions:  Dimensions{Width: 4096, Height: 4096},
		DefaultQuality: 85,
		SupportedTypes: []Format{JPEG, PNG, GIF, BMP},
		PreserveFormat: true,
	}
}

// WithMaxImageSize sets the maximum image size
func WithMaxImageSize(size int64) Option {
	return func(c *handlerConfig) {
		c.MaxImageSize = size
	}
}

// WithMaxDimensions sets the maximum image dimensions
func WithMaxDimensions(width, height int) Option {
	return func(c *handlerConfig) {
		c.MaxDimensions = Dimensions{Width: width, Height: height}
	}
}

// WithDefaultQuality sets the default compression quality
func WithDefaultQuality(quality int) Option {
	return func(c *handlerConfig) {
		if quality > 0 && quality <= 100 {
			c.DefaultQuality = quality
		}
	}
}

// WithSupportedTypes sets the supported image formats
func WithSupportedTypes(types []Format) Option {
	return func(c *handlerConfig) {
		c.SupportedTypes = types
	}
}

// WithPreserveFormat sets whether to preserve original format
func WithPreserveFormat(preserve bool) Option {
	return func(c *handlerConfig) {
		c.PreserveFormat = preserve
	}
}