package telemetry

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

const (
	defaultServiceName = "ragtime-backend"
	defaultEndpoint    = "localhost:4317"
)

// InitMetrics configures the global OTEL meter provider with an OTLP gRPC exporter.
func InitMetrics(ctx context.Context) (func(context.Context) error, error) {
	serviceName := strings.TrimSpace(os.Getenv("OTEL_SERVICE_NAME"))
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	exporter, err := buildOTLPMetricExporter(ctx)
	if err != nil {
		return nil, err
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	provider := metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(exporter, metric.WithInterval(10*time.Second)),
		),
		metric.WithResource(res),
	)

	otel.SetMeterProvider(provider)

	return provider.Shutdown, nil
}

func buildOTLPMetricExporter(ctx context.Context) (*otlpmetricgrpc.Exporter, error) {
	endpoint := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if endpoint == "" {
		endpoint = defaultEndpoint
	}

	opts := []otlpmetricgrpc.Option{}

	insecure := true
	if raw := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_INSECURE")); raw != "" {
		insecure = strings.EqualFold(raw, "true")
	}

	if strings.Contains(endpoint, "://") {
		parsed, err := url.Parse(endpoint)
		if err != nil {
			return nil, fmt.Errorf("parse OTEL_EXPORTER_OTLP_ENDPOINT: %w", err)
		}
		if parsed.Host == "" {
			return nil, fmt.Errorf("OTEL_EXPORTER_OTLP_ENDPOINT missing host: %q", endpoint)
		}
		opts = append(opts, otlpmetricgrpc.WithEndpoint(parsed.Host))
		if parsed.Scheme == "https" {
			insecure = false
		}
	} else {
		opts = append(opts, otlpmetricgrpc.WithEndpoint(endpoint))
	}

	if insecure {
		opts = append(opts, otlpmetricgrpc.WithInsecure())
	}

	exporter, err := otlpmetricgrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create OTLP metric exporter: %w", err)
	}
	return exporter, nil
}
