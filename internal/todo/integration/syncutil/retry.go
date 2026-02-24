package syncutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"
)

// RetryConfig holds retry settings for rate limit handling.
type RetryConfig struct {
	MaxRetries     int
	BaseRetryDelay time.Duration
	MaxRetryDelay  time.Duration
}

// DefaultRetryConfig returns standard retry settings.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     5,
		BaseRetryDelay: 1 * time.Second,
		MaxRetryDelay:  30 * time.Second,
	}
}

// RequestHooks provides client-specific behavior injected into the shared retry loop.
type RequestHooks struct {
	// SetAuth sets authentication and other required headers on the request before each attempt.
	SetAuth func(req *http.Request)

	// HandleRateLimit inspects an error response and returns a retryable error if it's a rate limit.
	// Return nil to fall through to transient/fatal error handling.
	HandleRateLimit func(resp *http.Response, body []byte) error

	// HandleAPIError inspects an error response and returns a formatted error.
	// Return nil to use the default "HTTP <status>: <body>" format.
	HandleAPIError func(statusCode int, body []byte) error
}

// DoWithRetry executes an HTTP request with retry logic for transient and rate-limit errors.
// It buffers the request body so it can be replayed on retries, applies exponential backoff
// with jitter, and delegates client-specific auth/rate-limit/error handling to hooks.
func DoWithRetry(httpClient *http.Client, req *http.Request, cfg RetryConfig, hooks RequestHooks, result any) error {
	// Buffer the request body so it can be replayed on retries
	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("reading request body: %w", err)
		}
		_ = req.Body.Close()
	}

	// Reset the body after reading so the first attempt has a valid body
	if bodyBytes != nil {
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			// Calculate delay with exponential backoff and jitter
			delay := min(cfg.BaseRetryDelay*time.Duration(1<<(attempt-1)), cfg.MaxRetryDelay)
			// Add jitter (0-25% of delay)
			jitter := time.Duration(rand.Int64N(int64(delay / 4)))
			delay += jitter

			select {
			case <-req.Context().Done():
				return req.Context().Err()
			case <-time.After(delay):
			}

			// Reset the body for retry
			if bodyBytes != nil {
				req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			}
		}

		// Apply auth headers
		if hooks.SetAuth != nil {
			hooks.SetAuth(req)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			// Check for transient network errors (stream errors, connection resets, etc.)
			if IsTransientNetworkError(err) {
				lastErr = fmt.Errorf("transient error: %s", err.Error())
				continue // Retry
			}
			return fmt.Errorf("executing request: %w", err)
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return fmt.Errorf("reading response: %w", err)
		}

		if resp.StatusCode >= httpErrorThreshold {
			// Check for rate limit errors via client-specific hook
			if hooks.HandleRateLimit != nil {
				if rateLimitErr := hooks.HandleRateLimit(resp, body); rateLimitErr != nil {
					lastErr = rateLimitErr
					continue // Retry
				}
			}

			// Check for transient HTTP errors (5xx, CloudFront errors, etc.)
			if IsTransientHTTPError(resp.StatusCode, body) {
				lastErr = fmt.Errorf("transient error: HTTP %d", resp.StatusCode)
				continue // Retry
			}

			// Try client-specific error parsing
			if hooks.HandleAPIError != nil {
				if apiErr := hooks.HandleAPIError(resp.StatusCode, body); apiErr != nil {
					return apiErr
				}
			}

			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}

		if result != nil && len(body) > 0 {
			if err := json.Unmarshal(body, result); err != nil {
				return fmt.Errorf("decoding response: %w", err)
			}
		}

		return nil
	}

	// All retries exhausted
	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// Transient network error substrings used for retry detection.
var transientNetworkPatterns = []string{
	"stream error",
	"INTERNAL_ERROR",
	"connection reset",
	"connection refused",
	"EOF",
	"timeout",
	"Timeout",
}

// Transient HTTP body patterns that indicate infrastructure/CDN errors.
var transientBodyPatterns = []string{
	"CloudFront",
	"cloudfront",
	"try again",
	"Try again",
}

// HTTP status code boundaries for retry classification.
const (
	httpErrorThreshold    = 400
	httpServerErrorMin    = 500
	httpServerErrorMax    = 600
)

// IsTransientNetworkError checks if an error is a transient network error that should be retried.
func IsTransientNetworkError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	for _, pattern := range transientNetworkPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	return false
}

// IsTransientHTTPError checks if an HTTP error is transient and should be retried.
// Covers 5xx server errors, CDN/infrastructure errors (CloudFront), and "try again" messages.
func IsTransientHTTPError(statusCode int, body []byte) bool {
	// 5xx server errors are always transient
	if statusCode >= httpServerErrorMin && statusCode < httpServerErrorMax {
		return true
	}
	// Some 4xx errors from CDN/infrastructure are transient
	if statusCode == http.StatusBadRequest || statusCode == http.StatusBadGateway ||
		statusCode == http.StatusServiceUnavailable || statusCode == http.StatusGatewayTimeout {
		bodyStr := string(body)
		for _, pattern := range transientBodyPatterns {
			if strings.Contains(bodyStr, pattern) {
				return true
			}
		}
	}
	return false
}
