package dataset

import (
	"fmt"
	"time"
)

// Format represents the output format for the dataset
type Format string

const (
	// FormatJSON outputs the dataset as a single JSON file
	FormatJSON Format = "json"
	// FormatJSONL outputs the dataset as a JSONL file (one JSON per line)
	FormatJSONL Format = "jsonl"
	// FormatCSV outputs the dataset as a CSV file
	FormatCSV Format = "csv"
)

// ProcessingStatus represents the status of record processing
type ProcessingStatus string

const (
	// StatusSuccess indicates successful processing
	StatusSuccess ProcessingStatus = "success"
	// StatusFailed indicates processing failure
	StatusFailed ProcessingStatus = "failed"
	// StatusSkipped indicates processing was skipped
	StatusSkipped ProcessingStatus = "skipped"
	// StatusPending indicates processing hasn't started
	StatusPending ProcessingStatus = "pending"
)

// BatchInfo contains information about a processing batch
type BatchInfo struct {
	ID        string    `json:"id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Size      int       `json:"size"`
	Status    string    `json:"status"`
}

// ExportOptions contains options for dataset export
type ExportOptions struct {
	// IncludeMetadata determines if metadata should be included
	IncludeMetadata bool `json:"include_metadata"`
	// IncludeFailures determines if failed records should be included
	IncludeFailures bool `json:"include_failures"`
	// MinConfidence is the minimum confidence threshold for inclusion
	MinConfidence float64 `json:"min_confidence"`
	// MaxRecords is the maximum number of records to include
	MaxRecords int `json:"max_records"`
	// SortBy determines the sorting field
	SortBy string `json:"sort_by"`
	// SortOrder determines the sorting order (asc/desc)
	SortOrder string `json:"sort_order"`
}

// ValidationError represents a dataset validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   any    `json:"value,omitempty"`
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("validation error for field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
	}
	return fmt.Sprintf("validation error for field '%s': %s", e.Field, e.Message)
}

// DatasetMetadata contains metadata about the dataset
type DatasetMetadata struct {
	GeneratedAt      time.Time         `json:"generated_at"`
	RecordCount      int               `json:"record_count"`
	Format           Format            `json:"format"`
	ProcessingStats  ProcessingStats   `json:"processing_stats"`
	ValidationErrors []ValidationError `json:"validation_errors,omitempty"`
}

// ProcessingStats contains statistics about dataset processing
type ProcessingStats struct {
	TotalDuration         time.Duration `json:"total_duration"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	SuccessRate           float64       `json:"success_rate"`
	ErrorRate             float64       `json:"error_rate"`
	TotalBatches          int           `json:"total_batches"`
	FailedBatches         int           `json:"failed_batches"`
}

// SortOptions defines sorting options for dataset records
type SortOptions struct {
	Field     string `json:"field"`     // Field to sort by
	Direction string `json:"direction"` // "asc" or "desc"
}

// FilterOptions defines filtering options for dataset records
type FilterOptions struct {
	MinConfidence   float64            `json:"min_confidence"`
	MaxConfidence   float64            `json:"max_confidence"`
	StartDate       time.Time          `json:"start_date"`
	EndDate         time.Time          `json:"end_date"`
	IncludeStatuses []ProcessingStatus `json:"include_statuses"`
	ExcludeStatuses []ProcessingStatus `json:"exclude_statuses"`
	RequiredLabels  []string           `json:"required_labels"`
	ExcludeLabels   []string           `json:"exclude_labels"`
}

// ValidateFilterOptions validates the filter options
func (f *FilterOptions) Validate() error {
	if f.MinConfidence < 0 || f.MinConfidence > 1 {
		return &ValidationError{
			Field:   "min_confidence",
			Message: "must be between 0 and 1",
			Value:   f.MinConfidence,
		}
	}

	if f.MaxConfidence < 0 || f.MaxConfidence > 1 {
		return &ValidationError{
			Field:   "max_confidence",
			Message: "must be between 0 and 1",
			Value:   f.MaxConfidence,
		}
	}

	if !f.StartDate.IsZero() && !f.EndDate.IsZero() && f.EndDate.Before(f.StartDate) {
		return &ValidationError{
			Field:   "date_range",
			Message: "end date must be after start date",
		}
	}

	return nil
}

// ValidateSortOptions validates the sort options
func (s *SortOptions) Validate() error {
	validFields := map[string]bool{
		"confidence":   true,
		"processed_at": true,
		"status":       true,
		"id":           true,
	}

	if !validFields[s.Field] {
		return &ValidationError{
			Field:   "sort_field",
			Message: "invalid sort field",
			Value:   s.Field,
		}
	}

	if s.Direction != "asc" && s.Direction != "desc" {
		return &ValidationError{
			Field:   "sort_direction",
			Message: "must be 'asc' or 'desc'",
			Value:   s.Direction,
		}
	}

	return nil
}
