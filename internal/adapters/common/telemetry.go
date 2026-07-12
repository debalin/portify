package common

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	Tracer            = otel.Tracer("portify-backend")
	Meter             = otel.Meter("portify-backend")
	PrometheusHandler http.Handler

	// Metrics
	ConversionsTotal     metric.Int64Counter
	TracksProcessedTotal metric.Int64Counter
	APIRequestsTotal     metric.Int64Counter
	APILatency           metric.Float64Histogram
	APIRetriesTotal      metric.Int64Counter
)

// InitTelemetry initializes the OTel Tracing and Metrics pipelines.
// Returns a shutdown function to flush data on server exit.
func InitTelemetry(ctx context.Context) func(context.Context) {
	// Create resource with environment attribute
	env := os.Getenv("PORTIFY_ENV")
	if env == "" {
		env = "local"
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String("portify-backend"),
			attribute.String("environment", env),
		),
	)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	// 1. Initialize Prometheus Exporter (always active for local pull metrics)
	promExporter, err := otelprom.New()
	if err != nil {
		log.Fatalf("failed to create prometheus exporter: %v", err)
	}

	// The exporter automatically registers itself with default registry, so we serve the global handler.
	PrometheusHandler = promhttp.Handler()

	meterReaders := []sdkmetric.Reader{promExporter}

	// Check if OTLP push endpoint is configured (for Grafana Cloud)
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	var tp *sdktrace.TracerProvider
	var otlpMetricExporter *otlpmetrichttp.Exporter

	if otlpEndpoint != "" {
		log.Printf("🔌 Initializing OpenTelemetry OTLP push exporter to %s...", otlpEndpoint)

		// Setup OTLP Tracing Exporter
		traceExporter, err := otlptracehttp.New(ctx)
		if err != nil {
			log.Fatalf("failed to create trace exporter: %v", err)
		}

		bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(res),
			sdktrace.WithSpanProcessor(bsp),
		)
		otel.SetTracerProvider(tp)

		// Setup OTLP Metrics Exporter
		otlpMetricExporter, err = otlpmetrichttp.New(ctx)
		if err != nil {
			log.Fatalf("failed to create OTLP metric exporter: %v", err)
		}
		meterReaders = append(meterReaders, sdkmetric.NewPeriodicReader(otlpMetricExporter))
	} else {
		log.Println("ℹ️  OTEL_EXPORTER_OTLP_ENDPOINT not set. OTLP tracing and metrics pushing are disabled (metrics served via local /metrics only).")
	}

	// 2. Setup MeterProvider with all registered readers and views
	var sdkReaders []sdkmetric.Option
	for _, r := range meterReaders {
		sdkReaders = append(sdkReaders, sdkmetric.WithReader(r))
	}
	sdkReaders = append(sdkReaders, sdkmetric.WithResource(res))

	// Configure custom explicit bucket boundaries for latency histogram (in seconds)
	latencyView := sdkmetric.NewView(
		sdkmetric.Instrument{Name: "portify_api_latency_seconds"},
		sdkmetric.Stream{
			Aggregation: sdkmetric.AggregationExplicitBucketHistogram{
				Boundaries: []float64{0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 1.25, 1.5, 1.75, 2.0, 2.5, 3.0, 4.0, 5.0, 7.5, 10.0},
			},
		},
	)
	sdkReaders = append(sdkReaders, sdkmetric.WithView(latencyView))

	mp := sdkmetric.NewMeterProvider(sdkReaders...)
	otel.SetMeterProvider(mp)

	return func(shutdownCtx context.Context) {
		if tp != nil {
			if err := tp.Shutdown(shutdownCtx); err != nil {
				log.Printf("error shutting down tracer provider: %v", err)
			}
		}
		if err := mp.Shutdown(shutdownCtx); err != nil {
			log.Printf("error shutting down meter provider: %v", err)
		}
	}
}

func initMetrics() {
	var err error
	ConversionsTotal, err = Meter.Int64Counter("portify_conversions_total",
		metric.WithDescription("Total playlist conversions"),
	)
	if err != nil {
		log.Printf("failed to create ConversionsTotal: %v", err)
	}

	TracksProcessedTotal, err = Meter.Int64Counter("portify_tracks_processed_total",
		metric.WithDescription("Total tracks processed"),
	)
	if err != nil {
		log.Printf("failed to create TracksProcessedTotal: %v", err)
	}

	APIRequestsTotal, err = Meter.Int64Counter("portify_api_requests_total",
		metric.WithDescription("Total outbound API requests to music providers"),
	)
	if err != nil {
		log.Printf("failed to create APIRequestsTotal: %v", err)
	}

	APILatency, err = Meter.Float64Histogram("portify_api_latency_seconds",
		metric.WithDescription("Latency of outbound API requests to music providers"),
	)
	if err != nil {
		log.Printf("failed to create APILatency: %v", err)
	}

	APIRetriesTotal, err = Meter.Int64Counter("portify_api_retries_total",
		metric.WithDescription("Total retries executed by the client"),
	)
	if err != nil {
		log.Printf("failed to create APIRetriesTotal: %v", err)
	}
}

func init() {
	initMetrics()
}
