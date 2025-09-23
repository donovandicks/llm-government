package internal

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	otelpyroscope "github.com/grafana/otel-profiling-go"
	"github.com/grafana/pyroscope-go"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

var (
	serviceName = os.Getenv("OTEL_SERVICE_NAME")
	logger      = otelslog.NewLogger(serviceName)

	Meter  = otel.Meter("github.com/donovandicks/rgo")
	Tracer = otel.Tracer(serviceName)
)

// SetupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up profiler
	profiler, err := startProfiler()
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, func(c context.Context) error { return profiler.Stop() })

	// Set up trace provider.
	tracerProvider, err := newTracerProvider(ctx)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(otelpyroscope.NewTracerProvider(tracerProvider))

	// Set up meter provider.
	meterProvider, err := newMeterProvider(ctx)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	loggerProvider, err := newLoggerProvider(ctx)
	if err != nil {
		handleErr(err)
		return shutdown, err
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))
	if err != nil {
		logger.ErrorContext(ctx, "otel runtime instrumentation failed", "error", err)
	}
	slog.SetDefault(logger)
	slog.SetLogLoggerLevel(getLogLevel())
	return shutdown, err
}

func getLogLevel() slog.Level {
	envLevel := os.Getenv("LOG_LEVEL")
	if envLevel == "" {
		envLevel = "DEBUG"
	}
	switch envLevel {
	case "ERROR":
		return slog.LevelError
	case "WARN":
		fallthrough
	case "WARNING":
		return slog.LevelWarn
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		fallthrough
	default:
		return slog.LevelInfo
	}
}

func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func newTracerProvider(ctx context.Context) (*trace.TracerProvider, error) {
	traceExporter, err := otlptrace.New(ctx, otlptracehttp.NewClient())
	if err != nil {
		return nil, err
	}

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
	)
	return tracerProvider, nil
}

func newMeterProvider(ctx context.Context) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(metricExporter),
		),
	)
	return meterProvider, nil
}

func newLoggerProvider(ctx context.Context) (*log.LoggerProvider, error) {
	logExporter, err := otlploghttp.New(
		ctx,
		otlploghttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(
			log.NewBatchProcessor(logExporter),
		),
	)
	return loggerProvider, nil
}

func startProfiler() (*pyroscope.Profiler, error) {
	addr := os.Getenv("OTEL_EXPORTER_PYROSCOPE_ENDPOINT")
	if addr == "" {
		return nil, errors.New("profiler missing server address")
	}

	slog.Debug("starting profiler", "serviceName", serviceName)
	return pyroscope.Start(pyroscope.Config{
		ApplicationName: serviceName,
		ServerAddress:   addr,
		Logger:          pyroscope.StandardLogger,
	})
}
