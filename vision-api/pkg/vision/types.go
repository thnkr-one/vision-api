package vision

import "time"

// APIVersion represents the Vision API version
type APIVersion string

const (
	// V1 is the stable Vision API version
	V1 APIVersion = "v1"
)

// FeatureType represents different Vision API features
type FeatureType string

const (
	// LabelDetection detects broad sets of categories within an image
	LabelDetection FeatureType = "LABEL_DETECTION"
	// ObjectLocalization detects and extracts multiple objects in an image
	ObjectLocalization FeatureType = "OBJECT_LOCALIZATION"
	// ImageProperties computes general attributes of the image
	ImageProperties FeatureType = "IMAGE_PROPERTIES"
)

// RequestStatus represents the status of an API request
type RequestStatus string

const (
	// StatusPending indicates request is waiting to be processed
	StatusPending RequestStatus = "PENDING"
	// StatusInProgress indicates request is being processed
	StatusInProgress RequestStatus = "IN_PROGRESS"
	// StatusCompleted indicates request completed successfully
	StatusCompleted RequestStatus = "COMPLETED"
	// StatusFailed indicates request failed
	StatusFailed RequestStatus = "FAILED"
)

// ErrorCode represents specific Vision API error codes
type ErrorCode int

const (
	// ErrorCodeUnknown indicates an unknown error occurred
	ErrorCodeUnknown ErrorCode = iota
	// ErrorCodeInvalidInput indicates invalid input parameters
	ErrorCodeInvalidInput
	// ErrorCodeRateLimitExceeded indicates rate limit was exceeded
	ErrorCodeRateLimitExceeded
	// ErrorCodePermissionDenied indicates lack of required permissions
	ErrorCodePermissionDenied
	// ErrorCodeTimeout indicates request timed out
	ErrorCodeTimeout
)

// APIError represents a structured Vision API error
type APIError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// RequestMetadata contains metadata about the API request
type RequestMetadata struct {
	RequestID  string        `json:"request_id"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	Duration   time.Duration `json:"duration"`
	Status     RequestStatus `json:"status"`
	RetryCount int           `json:"retry_count"`
	StatusCode int           `json:"status_code"`
	BytesSent  int64         `json:"bytes_sent"`
	BytesRecv  int64         `json:"bytes_recv"`
}

// Vertex represents a vertex in the image
type Vertex struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// BoundingPoly represents a bounding polygon for detected objects
type BoundingPoly struct {
	NormalizedVertices []Vertex `json:"normalized_vertices"`
}

// ObjectAnnotation represents detected object details
type ObjectAnnotation struct {
	Name        string       `json:"name"`
	Score       float64      `json:"score"`
	BoundingBox BoundingPoly `json:"bounding_poly"`
}

// ImageContext represents context information about the image
type ImageContext struct {
	LanguageHints []string            `json:"language_hints,omitempty"`
	Location      *LocationInfo       `json:"location,omitempty"`
	CropHints     *CropHintsParams    `json:"crop_hints_params,omitempty"`
	WebDetection  *WebDetectionParams `json:"web_detection_params,omitempty"`
}

// LocationInfo contains geographical information
type LocationInfo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// CropHintsParams contains parameters for crop hints
type CropHintsParams struct {
	AspectRatios []float64 `json:"aspect_ratios"`
}

// WebDetectionParams contains parameters for web detection
type WebDetectionParams struct {
	IncludeGeoResults bool `json:"include_geo_results"`
}

// AnnotateRequest represents a request to annotate an image
type AnnotateRequest struct {
	Image     []byte        `json:"-"`
	Features  []FeatureType `json:"features"`
	ImagePath string        `json:"-"`
	Context   *ImageContext `json:"image_context,omitempty"`
}

// AnnotateResponse represents the response from image annotation
type AnnotateResponse struct {
	Labels   []Label            `json:"label_annotations,omitempty"`
	Objects  []ObjectAnnotation `json:"object_annotations,omitempty"`
	Error    *APIError          `json:"error,omitempty"`
	Metadata RequestMetadata    `json:"metadata"`
}
