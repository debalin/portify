package common

import (
	"context"
	"os"
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func TestTelemetry_Disabled(t *testing.T) {
	// Ensure OTEL_EXPORTER_OTLP_ENDPOINT is empty
	origEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	defer os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", origEndpoint)

	ctx := context.Background()
	shutdown := InitTelemetry(ctx)
	defer shutdown(ctx)

	if PrometheusHandler == nil {
		t.Error("Expected PrometheusHandler to be initialized even if OTLP is disabled")
	}

	// Verify that recording metrics does not panic
	ConversionsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("source", "spotify"),
		attribute.String("destination", "youtube"),
		attribute.String("status", "success"),
	))

	TracksProcessedTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", "youtube"),
		attribute.String("status", "success"),
	))

	APIRequestsTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", "youtube"),
		attribute.Int("status_code", 200),
	))

	APILatency.Record(ctx, 0.25, metric.WithAttributes(
		attribute.String("provider", "youtube"),
	))

	APIRetriesTotal.Add(ctx, 1, metric.WithAttributes(
		attribute.String("provider", "youtube"),
		attribute.Int("attempt", 1),
		attribute.Int("status_code", 429),
	))
}

func TestTelemetry_Enabled(t *testing.T) {
	// Mock endpoint setup to trigger OTLP initialization path
	origEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	origHeaders := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")
	os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "Authorization=Basic dGVzdDpwYXNz")
	defer func() {
		os.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", origEndpoint)
		os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", origHeaders)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	shutdown := InitTelemetry(ctx)
	shutdown(ctx)
}
