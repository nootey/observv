package telemetry

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/log"
	lognoop "go.opentelemetry.io/otel/log/noop"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	"go.uber.org/zap"
)

type Config struct {
	ServiceName string
	Version     string
	Environment string
	Endpoint    string
	Disabled    bool
}

type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider
}

func New(ctx context.Context, cfg Config, logger *zap.Logger) (*Provider, error) {
	if cfg.Disabled {
		logger.Info("telemetry disabled")
		return &Provider{}, nil
	}

	// Stdout only: through the log bridge, export errors would feed the exporter.
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		logger.Warn("telemetry error", zap.Error(err))
	}))

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.Version),
			semconv.DeploymentEnvironmentNameKey.String(cfg.Environment),
		),
		resource.WithTelemetrySDK(),
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("resource: %w", err)
	}

	p := &Provider{}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpointURL(cfg.Endpoint))
	if err != nil {
		return nil, fmt.Errorf("trace exporter: %w", err)
	}
	p.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)

	metricExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithEndpointURL(cfg.Endpoint))
	if err != nil {
		return nil, errors.Join(fmt.Errorf("metric exporter: %w", err), p.Shutdown(ctx))
	}
	p.meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)

	logExporter, err := otlploggrpc.New(ctx, otlploggrpc.WithEndpointURL(cfg.Endpoint))
	if err != nil {
		return nil, errors.Join(fmt.Errorf("log exporter: %w", err), p.Shutdown(ctx))
	}
	p.loggerProvider = sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)),
		sdklog.WithResource(res),
	)

	otel.SetTracerProvider(p.tracerProvider)
	otel.SetMeterProvider(p.meterProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if err := runtime.Start(runtime.WithMeterProvider(p.meterProvider)); err != nil {
		return nil, errors.Join(fmt.Errorf("runtime metrics: %w", err), p.Shutdown(ctx))
	}

	logger.Info("telemetry initialized",
		zap.String("endpoint", cfg.Endpoint),
		zap.String("service_name", cfg.ServiceName),
		zap.String("service_version", cfg.Version),
	)

	return p, nil
}

func (p *Provider) LoggerProvider() log.LoggerProvider {
	if p.loggerProvider == nil {
		return lognoop.NewLoggerProvider()
	}
	return p.loggerProvider
}

func (p *Provider) Shutdown(ctx context.Context) error {
	var errs []error
	if p.tracerProvider != nil {
		if err := p.tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("tracer: %w", err))
		}
	}
	if p.meterProvider != nil {
		if err := p.meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("meter: %w", err))
		}
	}
	if p.loggerProvider != nil {
		if err := p.loggerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, fmt.Errorf("logger: %w", err))
		}
	}
	return errors.Join(errs...)
}
