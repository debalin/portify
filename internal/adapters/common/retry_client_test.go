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
