package processor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/your-username/vision-api/internal/utils"
)

// VisionProcessor handles image processing with Vision API integration
type VisionProcessor struct {
	options     *Options
	tracker     ProgressTracker
	tempManager *utils.TempFileManager
	mu          sync.RWMutex
}

// NewProcessor creates a new vision processor with the given options
func NewProcessor(opts ...OptionFunc) (*VisionProcessor, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	if err := options.validate(); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	tempManager, err := utils.NewTempFileManager(options.TempDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp manager: %w", err)
	}

	return &VisionProcessor{
		options:     options,
		tempManager: tempManager,
	}, nil
}

// Process implements the ImageProcessor interface
func (p *VisionProcessor) Process(ctx context.Context, input ProcessInput) (ProcessOutput, error) {
	startTime := time.Now()

	// Validate input
	if err := p.validateInput(input); err != nil {
		return ProcessOutput{}, err
	}

	// Process image
	output, err := p.processImage(ctx, input)

	// Record metrics
	p.recordMetrics(time.Since(startTime), err == nil)

	return output, err
}

// ProcessBatch implements batch processing
func (p *VisionProcessor) ProcessBatch(ctx context.Context, inputs []ProcessInput) ([]ProcessOutput, error) {
	if len(inputs) == 0 {
		return nil, nil
	}

	// Create buffered channels for processing
	jobs := make(chan ProcessInput, len(inputs))
	results := make(chan ProcessOutput, len(inputs))
	errors := make(chan error, 1)

	// Start worker pool
	var wg sync.WaitGroup
	for i := 0; i < p.options.PoolSize; i++ {
		wg.Add(1)
		go p.worker(ctx, &wg, jobs, results)
	}

	// Feed jobs to workers
	go func() {
		defer close(jobs)
		for _, input := range inputs {
			select {
			case <-ctx.Done():
				return
			case jobs <- input:
			}
		}
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Gather all results
	outputs := make([]ProcessOutput, 0, len(inputs))
	for result := range results {
		outputs = append(outputs, result)
		if p.tracker != nil {
			p.tracker.Update(int64(len(outputs)), int64(len(inputs)))
		}
	}

	// Clean up temp files if configured
	if p.options.DeleteTempFiles {
		if err := p.tempManager.Cleanup(); err != nil {
			return outputs, fmt.Errorf("cleanup error: %w", err)
		}
	}

	select {
	case err := <-errors:
		return outputs, err
	default:
		return outputs, nil
	}
}

// worker processes jobs from the jobs channel
func (p *VisionProcessor) worker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan ProcessInput, results chan<- ProcessOutput) {
	defer wg.Done()

	for job := range jobs {
		select {
		case <-ctx.Done():
			return
		default:
			output, err := p.Process(ctx, job)
			if err != nil {
				output.Error = err
			}
			results <- output
		}
	}
}

// processImage handles the core image processing logic
func (p *VisionProcessor) processImage(ctx context.Context, input ProcessInput) (ProcessOutput, error) {
	// Prepare image
	processedImage, err := p.prepareImage(ctx, input)
	if err != nil {
		return ProcessOutput{}, fmt.Errorf("image preparation failed: %w", err)
	}

	// Detect labels
	labels, err := p.detectLabels(ctx, processedImage)
	if err != nil {
		return ProcessOutput{}, fmt.Errorf("label detection failed: %w", err)
	}

	// Create output
	output := ProcessOutput{
		Filename: input.Filename,
		Labels:   labels,
		Metadata: map[string]interface{}{
			"processedAt": time.Now(),
			"size":        processedImage.Size,
			"format":      processedImage.Format,
		},
	}

	// Save results if output directory is configured
	if p.options.OutputDir != "" {
		if err := p.saveResults(output); err != nil {
			return output, fmt.Errorf("failed to save results: %w", err)
		}
	}

	return output, nil
}

// prepareImage prepares an image for processing
func (p *VisionProcessor) prepareImage(ctx context.Context, input ProcessInput) (*utils.FileInfo, error) {
	// Create temp file for processing
	tempFile, err := p.tempManager.CreateTemp(fmt.Sprintf("vision-%s-", input.Filename))
	if err != nil {
		return nil, err
	}
	defer tempFile.Close()

	// Process image using handler
	if err := p.options.ImageHandler.Process(ctx, input.Reader, tempFile); err != nil {
		return nil, err
	}

	// Get file info
	return utils.GetFileInfo(tempFile.Name())
}

// detectLabels detects labels in an image
func (p *VisionProcessor) detectLabels(ctx context.Context, fileInfo *utils.FileInfo) ([]Label, error) {
	labels, err := p.options.VisionClient.DetectLabels(ctx, fileInfo.Path)
	if err != nil {
		return nil, fmt.Errorf("vision API error: %w", err)
	}

	return labels, nil
}

// saveResults saves processing results
func (p *VisionProcessor) saveResults(output ProcessOutput) error {
	outputPath := filepath.Join(p.options.OutputDir, output.Filename+".json")
	return utils.SaveJSON(outputPath, output)
}

// validateInput validates the process input
func (p *VisionProcessor) validateInput(input ProcessInput) error {
	if input.Reader == nil {
		return fmt.Errorf("input reader is required")
	}

	if input.Filename == "" {
		return fmt.Errorf("filename is required")
	}

	ext := filepath.Ext(input.Filename)
	if ext == "" {
		return fmt.Errorf("filename must have an extension")
	}

	format := ext[1:] // Remove dot
	valid := false
	for _, allowedFormat := range p.options.AllowedFormats {
		if format == allowedFormat {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}

// recordMetrics records processing metrics
func (p *VisionProcessor) recordMetrics(duration time.Duration, success bool) {
	// Implement metrics recording if needed
}

// SetProgressTracker sets the progress tracking mechanism
func (p *VisionProcessor) SetProgressTracker(tracker ProgressTracker) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tracker = tracker
}

// Cleanup performs cleanup operations
func (p *VisionProcessor) Cleanup() error {
	return p.tempManager.Cleanup()
}
