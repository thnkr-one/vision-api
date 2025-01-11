package image

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"  // Register GIF format
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"io"
	"math"

	"github.com/disintegration/imaging"
)

// Resizer implements the ResizeHandler interface
type Resizer struct {
	config *handlerConfig
}

// NewResizer creates a new image resizer with the given options
func NewResizer(opts ...Option) *Resizer {
	config := NewHandlerConfig()
	for _, opt := range opts {
		opt(config)
	}
	return &Resizer{
		config: config,
	}
}

// Resize implements ResizeHandler.Resize
func (r *Resizer) Resize(ctx context.Context, input io.Reader, dimensions Dimensions) (io.Reader, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return r.resize(input, dimensions)
	}
}

// FitToSize implements ResizeHandler.FitToSize
func (r *Resizer) FitToSize(ctx context.Context, input io.Reader, maxDimensions Dimensions) (io.Reader, error) {
	// First get the image dimensions
	img, format, err := image.Decode(input)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	currentDims := Dimensions{
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
	}

	// Calculate new dimensions
	newDims := r.GetResizedDimensions(currentDims, maxDimensions)

	// If no resize needed, return original
	if newDims == currentDims {
		return input, nil
	}

	// Perform resize
	resized := imaging.Resize(img, newDims.Width, newDims.Height, imaging.Lanczos)

	// Encode the result
	var buf bytes.Buffer
	if err := r.encodeImage(resized, format, &buf); err != nil {
		return nil, fmt.Errorf("failed to encode resized image: %w", err)
	}

	return &buf, nil
}

// GetResizedDimensions implements ResizeHandler.GetResizedDimensions
func (r *Resizer) GetResizedDimensions(current, max Dimensions) Dimensions {
	// If current dimensions are smaller, no resize needed
	if current.Width <= max.Width && current.Height <= max.Height {
		return current
	}

	// Calculate aspect ratios
	currentRatio := float64(current.Width) / float64(current.Height)
	maxRatio := float64(max.Width) / float64(max.Height)

	var newWidth, newHeight int

	if currentRatio > maxRatio {
		// Width is the limiting factor
		newWidth = max.Width
		newHeight = int(math.Round(float64(max.Width) / currentRatio))
	} else {
		// Height is the limiting factor
		newHeight = max.Height
		newWidth = int(math.Round(float64(max.Height) * currentRatio))
	}

	return Dimensions{
		Width:  newWidth,
		Height: newHeight,
	}
}

// resize performs the actual image resizing
func (r *Resizer) resize(input io.Reader, dimensions Dimensions) (io.Reader, error) {
	// Decode image
	img, format, err := image.Decode(input)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Perform resize using Lanczos resampling
	resized := imaging.Resize(img, dimensions.Width, dimensions.Height, imaging.Lanczos)

	// Encode the resized image
	var buf bytes.Buffer
	if err := r.encodeImage(resized, format, &buf); err != nil {
		return nil, fmt.Errorf("failed to encode resized image: %w", err)
	}

	return &buf, nil
}

// encodeImage encodes the image in the appropriate format
func (r *Resizer) encodeImage(img image.Image, format string, w io.Writer) error {
	switch format {
	case "jpeg", "jpg":
		return imaging.Encode(w, img, imaging.JPEG, imaging.Quality(r.config.DefaultQuality))
	case "png":
		return imaging.Encode(w, img, imaging.PNG)
	case "gif":
		return imaging.Encode(w, img, imaging.GIF)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// ValidateFormat checks if the format is supported
func (r *Resizer) ValidateFormat(format Format) bool {
	for _, supported := range r.config.SupportedTypes {
		if format == supported {
			return true
		}
	}
	return false
}

// ValidateDimensions checks if dimensions are within limits
func (r *Resizer) ValidateDimensions(dimensions Dimensions) bool {
	return dimensions.Width <= r.config.MaxDimensions.Width &&
		dimensions.Height <= r.config.MaxDimensions.Height &&
		dimensions.Width > 0 &&
		dimensions.Height > 0
}