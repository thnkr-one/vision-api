package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/thnkr-one/vision-api/config"
	"github.com/your-username/vision-api/internal/image"
	"github.com/your-username/vision-api/internal/processor"
	"github.com/your-username/vision-api/internal/progress"
	"github.com/your-username/vision-api/pkg/dataset"
	"github.com/your-username/vision-api/pkg/vision"
)

var (
	configFile  string
	imageDir    string
	outputDir   string
	concurrency int
	debug       bool
)

func init() {
	flag.StringVar(&configFile, "config", "config.yaml", "Path to configuration file")
	flag.StringVar(&imageDir, "input", "", "Directory containing images to process")
	flag.StringVar(&outputDir, "output", "", "Directory for processed outputs")
	flag.IntVar(&concurrency, "concurrency", 0, "Number of concurrent processors")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Override config with command line flags if provided
	if imageDir != "" {
		cfg.Storage.InputDir = imageDir
	}
	if outputDir != "" {
		cfg.Storage.OutputDir = outputDir
	}
	if concurrency > 0 {
		cfg.Vision.PoolSize = concurrency
	}

	// Validate directories
	if err := validateDirectories(cfg); err != nil {
		return err
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-shutdown
		log.Println("Shutting down gracefully...")
		cancel()
	}()

	// Initialize components
	visionClient, err := initializeVisionClient(cfg)
	if err != nil {
		return fmt.Errorf("initializing vision client: %w", err)
	}

	imageHandler, err := initializeImageHandler(cfg)
	if err != nil {
		return fmt.Errorf("initializing image handler: %w", err)
	}

	processor, err := initializeProcessor(cfg, visionClient, imageHandler)
	if err != nil {
		return fmt.Errorf("initializing processor: %w", err)
	}

	// Find images to process
	images, err := findImages(cfg.Storage.InputDir)
	if err != nil {
		return fmt.Errorf("finding images: %w", err)
	}

	if len(images) == 0 {
		log.Println("No images found to process")
		return nil
	}

	// Initialize progress tracker
	tracker := progress.NewTracker(int64(len(images)), os.Stdout)
	processor.SetProgressTracker(tracker)
	tracker.Start()
	defer tracker.Finish()

	// Process images
	log.Printf("Processing %d images...", len(images))
	startTime := time.Now()

	results, err := processor.ProcessBatch(ctx, createProcessInputs(images))
	if err != nil {
		return fmt.Errorf("processing images: %w", err)
	}

	// Generate dataset
	if err := generateDataset(cfg, results); err != nil {
		return fmt.Errorf("generating dataset: %w", err)
	}

	duration := time.Since(startTime)
	log.Printf("Processing completed in %v", duration)

	return nil
}

func validateDirectories(cfg *config.Config) error {
	if cfg.Storage.InputDir == "" {
		return fmt.Errorf("input directory is required")
	}
	if cfg.Storage.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}

	for _, dir := range []string{cfg.Storage.InputDir, cfg.Storage.OutputDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	return nil
}

func initializeVisionClient(cfg *config.Config) (*vision.Client, error) {
	return vision.NewClient(
		vision.WithRateLimit(cfg.Vision.RateLimit),
		vision.WithMaxRetries(cfg.Vision.MaxRetries),
		vision.WithTimeout(time.Duration(cfg.Vision.TimeoutSeconds)*time.Second),
		vision.WithMaxConcurrent(cfg.Vision.PoolSize),
		vision.WithDebug(debug),
	)
}

func initializeImageHandler(cfg *config.Config) (image.Handler, error) {
	return image.NewHandler(
		image.WithMaxImageSize(int64(cfg.Image.MaxSizeMB)*1024*1024),
		image.WithMaxDimensions(cfg.Image.MaxWidth, cfg.Image.MaxHeight),
		image.WithDefaultQuality(cfg.Image.Quality),
	)
}

func initializeProcessor(cfg *config.Config, client *vision.Client, handler image.Handler) (processor.ImageProcessor, error) {
	return processor.NewProcessor(
		processor.WithPoolSize(cfg.Vision.PoolSize),
		processor.WithBatchSize(cfg.Vision.BatchSize),
		processor.WithImageHandler(handler),
		processor.WithVisionClient(client),
	)
}

func findImages(dir string) ([]string, error) {
	var images []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isImageFile(path) {
			images = append(images, path)
		}
		return nil
	})
	return images, err
}

func isImageFile(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return true
	default:
		return false
	}
}

func createProcessInputs(images []string) []processor.ProcessInput {
	inputs := make([]processor.ProcessInput, len(images))
	for i, path := range images {
		inputs[i] = processor.ProcessInput{
			Filename: filepath.Base(path),
			Metadata: map[string]interface{}{
				"path": path,
			},
		}
	}
	return inputs
}

func generateDataset(cfg *config.Config, results []processor.ProcessOutput) error {
	generator, err := dataset.NewGenerator(
		dataset.WithOutputDir(cfg.Storage.OutputDir),
		dataset.WithFormat(dataset.FormatJSONL),
	)
	if err != nil {
		return err
	}

	records := make([]dataset.Record, len(results))
	for i, result := range results {
		records[i] = dataset.Record{
			ID:        result.Filename,
			ImagePath: result.Metadata["path"].(string),
			Labels:    extractLabels(result.Labels),
			Status:    string(getStatus(result.Error)),
		}
		if result.Error != nil {
			records[i].ErrorMessage = result.Error.Error()
		}
	}

	return generator.GenerateDataset(context.Background(), records)
}

func extractLabels(labels []vision.Label) []string {
	result := make([]string, len(labels))
	for i, label := range labels {
		result[i] = label.Description
	}
	return result
}

func getStatus(err error) dataset.ProcessingStatus {
	if err == nil {
		return dataset.StatusSuccess
	}
	return dataset.StatusFailed
}
