package vision

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
	"time"
)

// Client handles communication with the Google Cloud Vision API
type Client struct {
	mu          sync.Mutex
	options     *Options
	rateLimiter *RateLimiter
}

// Label represents an image label from the Vision API
type Label struct {
	Description string  `json:"description"`
	Score       float64 `json:"score"`
	Topicality  float64 `json:"topicality,omitempty"`
}

// Response represents the Vision API response
type Response struct {
	Labels []Label `json:"labelAnnotations"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// RateLimiter handles API rate limiting
type RateLimiter struct {
	mu        sync.Mutex
	requests  []time.Time
	rateLimit int
	window    time.Duration
}

// NewClient creates a new Vision API client
func NewClient(opts ...OptionFunc) (*Client, error) {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	if err := validateOptions(options); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	return &Client{
		options: options,
		rateLimiter: &RateLimiter{
			rateLimit: options.RateLimit,
			window:    time.Minute,
			requests:  make([]time.Time, 0, options.RateLimit),
		},
	}, nil
}

// DetectLabels detects labels in the given image
func (c *Client) DetectLabels(ctx context.Context, imagePath string) ([]Label, error) {
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit wait: %w", err)
	}

	var response Response
	for attempt := 0; attempt <= c.options.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			output, err := c.executeCommand(ctx, imagePath)
			if err == nil {
				if err := json.Unmarshal(output, &response); err != nil {
					return nil, fmt.Errorf("failed to parse API response: %w", err)
				}

				if response.Error != nil {
					return nil, fmt.Errorf("API error: %s", response.Error.Message)
				}

				return response.Labels, nil
			}

			if attempt == c.options.MaxRetries {
				return nil, fmt.Errorf("max retries exceeded: %w", err)
			}

			// Calculate backoff delay
			delay := c.options.InitialBackoff * (1 << uint(attempt))
			if delay > c.options.MaxBackoff {
				delay = c.options.MaxBackoff
			}

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
	}

	return nil, fmt.Errorf("failed to detect labels")
}

// executeCommand executes the gcloud command
func (c *Client) executeCommand(ctx context.Context, imagePath string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gcloud", "ml", "vision", "detect-labels", imagePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command execution failed: %w: %s", err, string(output))
	}
	return output, nil
}

// Wait implements rate limiting
func (r *RateLimiter) Wait(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	// Remove expired timestamps
	i := 0
	for ; i < len(r.requests) && r.requests[i].Before(cutoff); i++ {
	}
	if i > 0 {
		r.requests = r.requests[i:]
	}

	// Check if we need to wait
	if len(r.requests) >= r.rateLimit {
		waitTime := r.requests[0].Add(r.window).Sub(now)
		if waitTime > 0 {
			r.mu.Unlock()
			select {
			case <-ctx.Done():
				r.mu.Lock()
				return ctx.Err()
			case <-time.After(waitTime):
				r.mu.Lock()
			}
		}
	}

	// Record request
	r.requests = append(r.requests, now)
	return nil
}

// GetCurrentRate returns the current request rate
func (r *RateLimiter) GetCurrentRate() int {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	count := 0
	for _, t := range r.requests {
		if t.After(cutoff) {
			count++
		}
	}

	return count
}

// ResetRateLimit resets the rate limiter
func (r *RateLimiter) ResetRateLimit() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = r.requests[:0]
}
