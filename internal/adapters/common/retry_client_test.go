package common

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryRoundTripper_TransientRecovery(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		curr := atomic.AddInt32(&attempts, 1)
		if curr <= 2 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	rt := &RetryRoundTripper{
		Base:       http.DefaultTransport,
		Classifier: StandardClassifier,
		MaxRetries: 3,
		MinBackoff: 10 * time.Millisecond,
		MaxBackoff: 50 * time.Millisecond,
		ProviderID: "test-provider",
	}

	client := &http.Client{Transport: rt}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", resp.StatusCode)
	}

	finalAttempts := atomic.LoadInt32(&attempts)
	if finalAttempts != 3 {
		t.Errorf("Expected 3 total requests (2 failures + 1 success), got: %d", finalAttempts)
	}
}

func TestRetryRoundTripper_TransientExhaustion(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	rt := &RetryRoundTripper{
		Base:       http.DefaultTransport,
		Classifier: StandardClassifier,
		MaxRetries: 2,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
		ProviderID: "test-provider",
	}

	client := &http.Client{Transport: rt}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Expected no transport error, got: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected final status 503, got: %d", resp.StatusCode)
	}

	finalAttempts := atomic.LoadInt32(&attempts)
	if finalAttempts != 3 { // Attempt 0 + 2 retries = 3 total attempts
		t.Errorf("Expected 3 total attempts (1 initial + 2 retries), got: %d", finalAttempts)
	}
}

func TestRetryRoundTripper_FatalFailFast(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	rt := &RetryRoundTripper{
		Base:       http.DefaultTransport,
		Classifier: StandardClassifier,
		MaxRetries: 3,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
		ProviderID: "test-provider",
	}

	client := &http.Client{Transport: rt}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Expected no transport error, got: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got: %d", resp.StatusCode)
	}

	finalAttempts := atomic.LoadInt32(&attempts)
	if finalAttempts != 1 {
		t.Errorf("Expected exactly 1 attempt (aborted immediately on 401), got: %d", finalAttempts)
	}
}

func TestRetryRoundTripper_RequestBodyRecovery(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		curr := atomic.AddInt32(&attempts, 1)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if string(body) != "hello payload" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if curr == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	rt := &RetryRoundTripper{
		Base:       http.DefaultTransport,
		Classifier: StandardClassifier,
		MaxRetries: 2,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
		ProviderID: "test-provider",
	}

	client := &http.Client{Transport: rt}

	payload := bytes.NewReader([]byte("hello payload"))
	req, _ := http.NewRequestWithContext(context.Background(), "POST", server.URL, payload)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Expected no transport error, got: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got: %d", resp.StatusCode)
	}

	finalAttempts := atomic.LoadInt32(&attempts)
	if finalAttempts != 2 {
		t.Errorf("Expected 2 attempts, got: %d", finalAttempts)
	}
}

func TestRetryRoundTripper_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	rt := &RetryRoundTripper{
		Base:       http.DefaultTransport,
		Classifier: StandardClassifier,
		MaxRetries: 3,
		MinBackoff: 50 * time.Millisecond,
		MaxBackoff: 100 * time.Millisecond,
		ProviderID: "test-provider",
	}

	client := &http.Client{Transport: rt}

	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)

	// Cancel context after a very brief duration, during first backoff sleep
	go func() {
		time.Sleep(15 * time.Millisecond)
		cancel()
	}()

	_, err := client.Do(req)
	if err == nil {
		t.Fatal("Expected context cancellation error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected error to wrap context.Canceled, got: %v", err)
	}
}

func TestRetryRoundTripper_HookNotification(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	rt := &RetryRoundTripper{
		Base:       http.DefaultTransport,
		Classifier: StandardClassifier,
		MaxRetries: 2,
		MinBackoff: 1 * time.Millisecond,
		MaxBackoff: 5 * time.Millisecond,
		ProviderID: "test-provider",
	}

	client := &http.Client{Transport: rt}

	var hookCalled int32
	hook := func(event RetryEvent) {
		atomic.AddInt32(&hookCalled, 1)
		if event.ProviderID != "test-provider" {
			t.Errorf("Expected ProviderID 'test-provider', got: %s", event.ProviderID)
		}
		if event.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("Expected status 503, got: %d", event.StatusCode)
		}
	}

	ctx := WithRetryHook(context.Background(), hook)
	req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	defer resp.Body.Close()

	finalHookCalled := atomic.LoadInt32(&hookCalled)
	if finalHookCalled != 2 { // Hook should run on each retry attempt (attempt 1 and 2)
		t.Errorf("Expected hook to be called 2 times, got: %d", finalHookCalled)
	}
}

type badReader struct{}

func (badReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}
func (badReader) Close() error {
	return nil
}

func TestRetryRoundTripper_BodyReadError(t *testing.T) {
	rt := &RetryRoundTripper{
		Base:       http.DefaultTransport,
		Classifier: StandardClassifier,
	}
	req, _ := http.NewRequest("POST", "http://example.com", badReader{})
	_, err := rt.RoundTrip(req)
	if err == nil || err.Error() != "read error" {
		t.Errorf("expected read error, got: %v", err)
	}
}

type mockRoundTripper struct {
	roundTripFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func TestRetryRoundTripper_GetBodyError(t *testing.T) {
	rt := &RetryRoundTripper{
		Base: &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				return nil, errors.New("transient network error")
			},
		},
		Classifier: func(resp *http.Response, err error) (bool, time.Duration) {
			return true, 0
		},
		MaxRetries: 1,
		MinBackoff: 1 * time.Millisecond,
	}
	req, _ := http.NewRequest("POST", "http://example.com", bytes.NewBufferString("hello"))
	req.GetBody = func() (io.ReadCloser, error) {
		return nil, errors.New("getbody error")
	}
	_, err := rt.RoundTrip(req)
	if err == nil || err.Error() != "getbody error" {
		t.Errorf("expected getbody error, got: %v", err)
	}
}

func TestRetryRoundTripper_RetryAfterInvalid(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Header:     make(http.Header),
	}
	resp.Header.Set("Retry-After", "not-a-number")
	retryable, backoff := StandardClassifier(resp, nil)
	if !retryable {
		t.Error("expected retryable = true")
	}
	if backoff != 0 {
		t.Errorf("expected backoff = 0, got %v", backoff)
	}
}

func TestRetryRoundTripper_GetBackoffOverflow(t *testing.T) {
	rt := &RetryRoundTripper{
		Classifier: func(resp *http.Response, err error) (bool, time.Duration) {
			return false, 0
		},
	}
	// Normal cap at maxBackoff (not overflowing int64), subject to full jitter [0, maxBackoff]
	duration := rt.getBackoff(5, nil, nil, 10*time.Second, 100*time.Second)
	if duration <= 0 || duration > 100*time.Second {
		t.Errorf("expected backoff between 0 and maxBackoff, got %v", duration)
	}

	// Overflow test, should bypass jitter and return minBackoff
	durationOverflow := rt.getBackoff(100, nil, nil, 10*time.Second, 10000*time.Hour)
	if durationOverflow != 10*time.Second {
		t.Errorf("expected overflow protection to fallback to minBackoff, got %v", durationOverflow)
	}
}

func TestRetryRoundTripper_CancellationWithLastResp(t *testing.T) {
	var bodyClosed int32
	mockRespBody := &mockReadCloser{
		Reader: bytes.NewReader([]byte("transient failure details")),
		closeFunc: func() error {
			atomic.StoreInt32(&bodyClosed, 1)
			return nil
		},
	}
	rt := &RetryRoundTripper{
		Base: &mockRoundTripper{
			roundTripFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusServiceUnavailable,
					Body:       mockRespBody,
				}, nil
			},
		},
		Classifier: func(resp *http.Response, err error) (bool, time.Duration) {
			return true, 0
		},
		MaxRetries: 2,
		MinBackoff: 10 * time.Millisecond,
		MaxBackoff: 20 * time.Millisecond,
	}

	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "GET", "http://example.com", nil)

	hook := func(event RetryEvent) {
		cancel()
	}
	ctx = WithRetryHook(ctx, hook)
	req = req.WithContext(ctx)

	_, err := rt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected context canceled error")
	}

	if atomic.LoadInt32(&bodyClosed) != 1 {
		t.Error("expected last response body to be closed upon context cancellation")
	}
}

type mockReadCloser struct {
	io.Reader
	closeFunc func() error
}

func (m *mockReadCloser) Close() error {
	return m.closeFunc()
}

func TestStandardClassifier_OtherStatusCodes(t *testing.T) {
	cases := []struct {
		statusCode int
		expected   bool
	}{
		{http.StatusBadGateway, true},
		{http.StatusGatewayTimeout, true},
		{http.StatusInternalServerError, false},
		{http.StatusOK, false},
	}
	for _, tc := range cases {
		resp := &http.Response{StatusCode: tc.statusCode}
		retryable, _ := StandardClassifier(resp, nil)
		if retryable != tc.expected {
			t.Errorf("status %d: expected retryable = %t, got %t", tc.statusCode, tc.expected, retryable)
		}
	}
}
