package common

import (
	"bytes"
	"context"
	"io"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type retryHookKey struct{}

// RetryEvent holds details about an in-progress retry attempt.
type RetryEvent struct {
	ProviderID string
	Attempt    int
	Delay      time.Duration
	StatusCode int
	Err        error
}

// OnRetryHook is a callback function for retries.
type OnRetryHook func(event RetryEvent)

// WithRetryHook attaches an OnRetryHook callback to the context.
func WithRetryHook(ctx context.Context, hook OnRetryHook) context.Context {
	return context.WithValue(ctx, retryHookKey{}, hook)
}

// GetRetryHook retrieves the OnRetryHook from the context if present.
func GetRetryHook(ctx context.Context) OnRetryHook {
	if hook, ok := ctx.Value(retryHookKey{}).(OnRetryHook); ok {
		return hook
	}
	return nil
}

// ErrorClassifier classifies an HTTP response/error to decide whether it's retriable, and how long to wait.
type ErrorClassifier func(resp *http.Response, err error) (retryable bool, backoff time.Duration)

// StandardClassifier parses Retry-After headers (in seconds) or applies standard checks.
func StandardClassifier(resp *http.Response, err error) (bool, time.Duration) {
	if err != nil {
		return true, 0
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfterStr := resp.Header.Get("Retry-After")
		if retryAfterStr != "" {
			if seconds, err := strconv.Atoi(retryAfterStr); err == nil {
				return true, time.Duration(seconds) * time.Second
			}
		}
		return true, 0
	}
	if resp.StatusCode == http.StatusBadGateway ||
		resp.StatusCode == http.StatusServiceUnavailable ||
		resp.StatusCode == http.StatusGatewayTimeout {
		return true, 0
	}
	return false, 0
}

// RetryRoundTripper wraps an existing http.RoundTripper to handle retries and backoffs.
type RetryRoundTripper struct {
	Base       http.RoundTripper
	Classifier ErrorClassifier
	MaxRetries int
	MinBackoff time.Duration
	MaxBackoff time.Duration
	ProviderID string
}

func (r *RetryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// If there is a body, cache it in memory so we can rewind it for retries.
	if req.Body != nil && req.GetBody == nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		req.Body.Close()
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(bodyBytes)), nil
		}
		req.Body, _ = req.GetBody()
	}

	var lastResp *http.Response
	var lastErr error

	maxRetries := r.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	minBackoff := r.MinBackoff
	if minBackoff <= 0 {
		minBackoff = 250 * time.Millisecond
	}
	maxBackoff := r.MaxBackoff
	if maxBackoff <= 0 {
		maxBackoff = 5 * time.Second
	}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Rewind request body for retry
			if req.GetBody != nil {
				newBody, err := req.GetBody()
				if err != nil {
					return nil, err
				}
				req.Body = newBody
			}

			backoff := r.getBackoff(attempt-1, lastResp, lastErr, minBackoff, maxBackoff)

			if hook := GetRetryHook(req.Context()); hook != nil {
				statusCode := 0
				if lastResp != nil {
					statusCode = lastResp.StatusCode
				}
				hook(RetryEvent{
					ProviderID: r.ProviderID,
					Attempt:    attempt,
					Delay:      backoff,
					StatusCode: statusCode,
					Err:        lastErr,
				})
			}

			select {
			case <-req.Context().Done():
				if lastResp != nil {
					lastResp.Body.Close()
				}
				return nil, req.Context().Err()
			case <-time.After(backoff):
			}
		}

		resp, err := r.Base.RoundTrip(req)

		retryable, _ := r.Classifier(resp, err)
		if !retryable || attempt == maxRetries {
			return resp, err
		}

		// Close response body to prevent leaks before retrying.
		if resp != nil {
			io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
			resp.Body.Close()
		}

		lastResp = resp
		lastErr = err
	}

	return lastResp, lastErr
}

func (r *RetryRoundTripper) getBackoff(
	attempt int,
	resp *http.Response,
	err error,
	minBackoff, maxBackoff time.Duration,
) time.Duration {
	if _, customBackoff := r.Classifier(resp, err); customBackoff > 0 {
		return customBackoff
	}

	temp := float64(minBackoff) * math.Pow(2, float64(attempt))
	backoff := time.Duration(temp)
	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	if backoff <= 0 {
		return minBackoff
	}

	jitter := rand.Int63n(int64(backoff))
	return time.Duration(jitter)
}
