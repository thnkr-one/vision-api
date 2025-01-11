package processor

import (
	"fmt"
	"time"
)

// Options contains configuration for the Vision API client
type Options struct {
	// RateLimit specifies requests per minute limit
	RateLimit int

	// MaxRetries specifies maximum number of retry attempts
	MaxRetries int

	// InitialBackoff specifies the initial backoff duration
	InitialBackoff time.Duration

	// MaxBackoff specifies the maximum backoff duration
	MaxBackoff time.Duration

	// Timeout specifies the timeout for individual requests
	Timeout time.Duration

	// MaxConcurrent specifies maximum concurrent requests
	MaxConcurrent int

	// Debug enables debug logging
	Debug bool

	// ProjectID specifies the Google Cloud project ID
	ProjectID string

	// Credentials specifies the path to service account credentials
	Credentials string
}

// OptionFunc is a function that configures Options
type OptionFunc func(*Options)

// defaultOptions returns the default configuration options
func defaultOptions() *Options {
	return &Options{
		RateLimit:      1800,        // 1800 requests per minute (Google's default)
		MaxRetries:     3,           // 3 retry attempts
		InitialBackoff: time.Second, // Start with 1 second backoff
		MaxBackoff:     time.Second * 32,
		Timeout:        time.Second * 30,
		MaxConcurrent:  10,
		Debug:          false,
	}
}

// WithRateLimit sets the rate limit
func WithRateLimit(limit int) OptionFunc {
	return func(o *Options) {
		o.RateLimit = limit
	}
}

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(retries int) OptionFunc {
	return func(o *Options) {
		o.MaxRetries = retries
	}
}

// WithInitialBackoff sets the initial backoff duration
func WithInitialBackoff(backoff time.Duration) OptionFunc {
	return func(o *Options) {
		o.InitialBackoff = backoff
	}
}

// WithMaxBackoff sets the maximum backoff duration
func WithMaxBackoff(backoff time.Duration) OptionFunc {
	return func(o *Options) {
		o.MaxBackoff = backoff
	}
}

// WithTimeout sets the request timeout
func WithTimeout(timeout time.Duration) OptionFunc {
	return func(o *Options) {
		o.Timeout = timeout
	}
}

// WithMaxConcurrent sets the maximum concurrent requests
func WithMaxConcurrent(max int) OptionFunc {
	return func(o *Options) {
		o.MaxConcurrent = max
	}
}

// WithDebug enables or disables debug logging
func WithDebug(debug bool) OptionFunc {
	return func(o *Options) {
		o.Debug = debug
	}
}

// WithProjectID sets the Google Cloud project ID
func WithProjectID(projectID string) OptionFunc {
	return func(o *Options) {
		o.ProjectID = projectID
	}
}

// WithCredentials sets the path to service account credentials
func WithCredentials(path string) OptionFunc {
	return func(o *Options) {
		o.Credentials = path
	}
}

// validateOptions validates the configuration options
func validateOptions(o *Options) error {
	if o.RateLimit < 1 {
		return fmt.Errorf("rate limit must be at least 1")
	}

	if o.MaxRetries < 0 {
		return fmt.Errorf("max retries cannot be negative")
	}

	if o.InitialBackoff < 0 {
		return fmt.Errorf("initial backoff cannot be negative")
	}

	if o.MaxBackoff < o.InitialBackoff {
		return fmt.Errorf("max backoff must be greater than or equal to initial backoff")
	}

	if o.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	if o.MaxConcurrent < 1 {
		return fmt.Errorf("max concurrent must be at least 1")
	}

	return nil
}
