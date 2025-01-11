package dataset

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Generator handles dataset generation from processing results
type Generator struct {
	options *Options
	mu      sync.RWMutex
}

// Record represents a single dataset record
type Record struct {
	ID           string                 `json:"id"`
	ImagePath    string                 `json:"image_path"`
	Labels       []string               `json:"labels"`
	Confidence   float64                `json:"confidence"`
	ProcessedAt  time.Time              `json:"processed_at"`
	Status       string                 `json:"status"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// Stats contains dataset generation statistics
type Stats struct {
	TotalRecords      int           `json:"total_records"`
	SuccessfulCount   int           `json:"successful_count"`
	FailedCount       int           `json:"failed_count"`
	SkippedCount      int           `json:"skipped_count"`
	ProcessingTime    time.Duration `json:"processing_time"`
	AverageLabels     float64       `json:"average_labels"`
	UniqueLabels      int           `json:"unique_labels"`
	AverageConfidence float64       `json:"average_confidence"`
}

// NewGenerator creates a new dataset generator with the given options
func NewGenerator(opts ...OptionFunc) (*Generator, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	if err := validateOptions(options); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	return &Generator{
		options: options,
	}, nil
}

// GenerateDataset generates a dataset from the processing results
func (g *Generator) GenerateDataset(ctx context.Context, records []Record) error {
	if err := g.validateOutputDir(); err != nil {
		return err
	}

	switch g.options.Format {
	case FormatJSON:
		return g.generateJSON(ctx, records)
	case FormatCSV:
		return g.generateCSV(ctx, records)
	case FormatJSONL:
		return g.generateJSONL(ctx, records)
	default:
		return fmt.Errorf("unsupported format: %s", g.options.Format)
	}
}

// generateJSON generates a JSON dataset file
func (g *Generator) generateJSON(ctx context.Context, records []Record) error {
	outputPath := filepath.Join(g.options.OutputDir, "dataset.json")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	if g.options.PrettyPrint {
		encoder.SetIndent("", "  ")
	}

	dataset := struct {
		Records []Record `json:"records"`
		Stats   Stats    `json:"stats"`
	}{
		Records: records,
		Stats:   g.calculateStats(records),
	}

	if err := encoder.Encode(dataset); err != nil {
		return fmt.Errorf("failed to encode dataset: %w", err)
	}

	return nil
}

// generateCSV generates a CSV dataset file
func (g *Generator) generateCSV(ctx context.Context, records []Record) error {
	outputPath := filepath.Join(g.options.OutputDir, "dataset.csv")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"id", "image_path", "labels", "confidence", "processed_at", "status", "error_message"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write records
	for _, record := range records {
		labelsJSON, err := json.Marshal(record.Labels)
		if err != nil {
			return fmt.Errorf("failed to marshal labels: %w", err)
		}

		row := []string{
			record.ID,
			record.ImagePath,
			string(labelsJSON),
			fmt.Sprintf("%.4f", record.Confidence),
			record.ProcessedAt.Format(time.RFC3339),
			record.Status,
			record.ErrorMessage,
		}

		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

// generateJSONL generates a JSONL dataset file
func (g *Generator) generateJSONL(ctx context.Context, records []Record) error {
	outputPath := filepath.Join(g.options.OutputDir, "dataset.jsonl")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, record := range records {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := encoder.Encode(record); err != nil {
				return fmt.Errorf("failed to encode record: %w", err)
			}
		}
	}

	return nil
}

// calculateStats calculates dataset statistics
func (g *Generator) calculateStats(records []Record) Stats {
	var stats Stats
	stats.TotalRecords = len(records)
	uniqueLabels := make(map[string]struct{})
	totalLabels := 0
	totalConfidence := 0.0

	for _, record := range records {
		switch record.Status {
		case "success":
			stats.SuccessfulCount++
		case "failed":
			stats.FailedCount++
		case "skipped":
			stats.SkippedCount++
		}

		totalLabels += len(record.Labels)
		totalConfidence += record.Confidence

		for _, label := range record.Labels {
			uniqueLabels[label] = struct{}{}
		}
	}

	if stats.SuccessfulCount > 0 {
		stats.AverageLabels = float64(totalLabels) / float64(stats.SuccessfulCount)
		stats.AverageConfidence = totalConfidence / float64(stats.SuccessfulCount)
	}
	stats.UniqueLabels = len(uniqueLabels)

	return stats
}

// validateOutputDir ensures the output directory exists and is writable
func (g *Generator) validateOutputDir() error {
	if g.options.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}

	if err := os.MkdirAll(g.options.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Check if directory is writable
	testFile := filepath.Join(g.options.OutputDir, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("output directory is not writable: %w", err)
	}
	f.Close()
	os.Remove(testFile)

	return nil
}
