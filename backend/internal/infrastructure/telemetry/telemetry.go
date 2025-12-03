package telemetry

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Config holds the configuration for OpenTelemetry
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string
}

// TracerProvider wraps the OpenTelemetry tracer provider
type TracerProvider struct {
	provider *sdktrace.TracerProvider
}

// InitTracerProvider initializes the OpenTelemetry tracer provider
func InitTracerProvider(ctx context.Context, config Config) (*TracerProvider, error) {
	// Create resource with service information
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", config.ServiceName),
			attribute.String("service.version", config.ServiceVersion),
			attribute.String("deployment.environment", config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Set up OTLP exporter
	conn, err := grpc.NewClient(
		config.OTLPEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	traceExporter, err := otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			otlptracegrpc.WithGRPCConn(conn),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Set global propagator to tracecontext (W3C Trace Context)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &TracerProvider{
		provider: tp,
	}, nil
}

func (tp *TracerProvider) TraceProvider() *sdktrace.TracerProvider {
	return tp.provider
}

// Shutdown gracefully shuts down the tracer provider
func (tp *TracerProvider) Shutdown(ctx context.Context) error {
	if tp.provider == nil {
		return nil
	}

	// Create a timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return tp.provider.Shutdown(shutdownCtx)
}

// Tracer returns a tracer for the given name
func (tp *TracerProvider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return tp.provider.Tracer(name, opts...)
}

// startSpan starts a new span with the given name and returns the span and a context containing the span
func startSpan(ctx context.Context, tracerName, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracer := otel.Tracer(tracerName)
	return tracer.Start(ctx, spanName, opts...)
}

// addSpanAttributes adds attributes to the span in the current context
func addSpanAttributes(span trace.Span, attributes ...attribute.KeyValue) {
	span.SetAttributes(attributes...)
}

// RecordError records an error in the span and sets the span status to error
func RecordError(span trace.Span, err error, attributes ...attribute.KeyValue) {
	span.SetStatus(codes.Error, err.Error())
	span.RecordError(err, trace.WithAttributes(attributes...))
}

// wrapWithSpan wraps a function with a span
func wrapWithSpan(ctx context.Context, tracerName, spanName string, fn func(context.Context) error, attributes ...attribute.KeyValue) error {
	ctx, span := startSpan(ctx, tracerName, spanName)
	defer span.End()

	if len(attributes) > 0 {
		addSpanAttributes(span, attributes...)
	}

	err := fn(ctx)
	if err != nil {
		RecordError(span, err)
	}

	return err
}

// ServiceTracer creates a helper struct for a specific service to simplify tracing
type ServiceTracer struct {
	serviceName string
}

// NewServiceTracer creates a new ServiceTracer for the given service name
func NewServiceTracer(serviceName string) *ServiceTracer {
	return &ServiceTracer{
		serviceName: serviceName,
	}
}

// StartSpan starts a new span for this service
func (st *ServiceTracer) StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return startSpan(ctx, st.serviceName, spanName, opts...)
}

// WrapWithSpan wraps a function with a span for this service
func (st *ServiceTracer) WrapWithSpan(ctx context.Context, spanName string, fn func(context.Context) error, attributes ...attribute.KeyValue) error {
	return wrapWithSpan(ctx, st.serviceName, spanName, fn, attributes...)
}
