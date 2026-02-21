package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync/atomic"
)

// OTEL logging is intentionally disabled for now (console-only logging).
// Uncomment these imports and the OTEL setup code below when re-enabling.
// "go.opentelemetry.io/contrib/bridges/otelslog"
// "go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
// sdklog "go.opentelemetry.io/otel/sdk/log"
// "go.opentelemetry.io/otel/sdk/resource"
// "go.opentelemetry.io/otel/semconv/v1.26.0"

// Type alias for slog.Level for easier usage
type Level = slog.Level

const (
	LevelTrace   = slog.Level(-8)
	LevelDebug   = slog.LevelDebug // -4
	LevelInfo    = slog.LevelInfo  // 0
	LevelWarning = slog.LevelWarn  // 4
	LevelError   = slog.LevelError // 8
	LevelFatal   = slog.Level(12)  // 12

	// Aliases for backward compatibility
	TraceLevel   = LevelTrace
	DebugLevel   = LevelDebug
	InfoLevel    = LevelInfo
	WarningLevel = LevelWarning
	ErrorLevel   = LevelError
	FatalLevel   = LevelFatal
)

var (
	Logger       *slog.Logger
	programLevel = new(slog.LevelVar)
	shutdownFunc func(context.Context) error // Shutdown function for OTEL (nil if not using OTEL)
)

// Error counters for metrics endpoint (incremented regardless of sampling)
var (
	TotalErrors      atomic.Int64
	TotalWarnings    atomic.Int64
	Total5xxErrors   atomic.Int64
	Total4xxErrors   atomic.Int64
	Total400Errors   atomic.Int64
	Total404Errors   atomic.Int64
	Total429Errors   atomic.Int64
	SlowRequests     atomic.Int64
	ConnPoolWarnings atomic.Int64
)

func init() {
	// Default to DEBUG to ensure logs are visible
	programLevel.Set(slog.LevelDebug)

	// Get log level from environment variable (default: DEBUG)
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "DEBUG"
	}

	level, err := ParseLevel(levelStr)
	if err != nil {
		level = slog.LevelDebug
		fmt.Fprintf(os.Stderr, "⚠️  Invalid LOG_LEVEL '%s', using DEBUG\n", levelStr)
	}
	programLevel.Set(level)
	fmt.Fprintf(os.Stderr, "📊 Log level set to: %s\n", level)

	// OpenTelemetry logging is intentionally disabled for now.
	// If re-enabled, restore the OTEL setup branch below.
	setupJSONLogging()
	fmt.Fprintf(os.Stderr, "📝 JSON logging enabled\n")
	fmt.Fprintf(os.Stderr, "   OTEL logging is currently disabled; enable in code when ready.\n")
}

// setupJSONLogging configures standard JSON logging to stdout
func setupJSONLogging() {
	opts := &slog.HandlerOptions{
		Level: programLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// setupOTELLogging configures OpenTelemetry logging
// setupOTELLogging configures OpenTelemetry logging (currently disabled).
// Uncomment when OTEL collector logging is re-enabled.
// func setupOTELLogging(ctx context.Context, serviceName string) (func(context.Context) error, error) {
// 	fmt.Fprintf(os.Stderr, "  Setting up OTEL exporter...\n")
//
// 	// Resource = service identity
// 	res, err := resource.New(ctx,
// 		resource.WithAttributes(
// 			semconv.ServiceName(serviceName),
// 		),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create resource: %w", err)
// 	}
// 	fmt.Fprintf(os.Stderr, "  ✓ Created OTEL resource\n")
//
// 	// OTLP log exporter (gRPC)
// 	// Get endpoint from env var or use default
// 	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
// 	if endpoint == "" {
// 		endpoint = "localhost:4317"
// 	}
// 	fmt.Fprintf(os.Stderr, "  Connecting to OTLP collector at %s...\n", endpoint)
//
// 	exporter, err := otlploggrpc.New(ctx,
// 		otlploggrpc.WithEndpoint(endpoint),
// 	)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create OTLP exporter (check if collector is reachable): %w", err)
// 	}
// 	fmt.Fprintf(os.Stderr, "  ✓ Connected to OTLP collector at %s\n", endpoint)
//
// 	// Log processor (batching is recommended)
// 	processor := sdklog.NewBatchProcessor(exporter)
//
// 	// LoggerProvider
// 	loggerProvider := sdklog.NewLoggerProvider(
// 		sdklog.WithResource(res),
// 		sdklog.WithProcessor(processor),
// 	)
//
// 	// Bridge slog → OTel
// 	otelHandler := otelslog.NewHandler(
// 		serviceName,
// 		otelslog.WithLoggerProvider(loggerProvider),
// 	)
//
// 	// Wrap with level filtering
// 	handler := &levelHandler{
// 		level:   programLevel,
// 		handler: otelHandler,
// 	}
//
// 	// Create and set logger
// 	Logger = slog.New(handler)
// 	slog.SetDefault(Logger)
//
// 	// Return shutdown hook
// 	return loggerProvider.Shutdown, nil
// }

// levelHandler wraps a handler to filter by level
type levelHandler struct {
	level   slog.Leveler
	handler slog.Handler
}

func (h *levelHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *levelHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.handler.Handle(ctx, r)
}

func (h *levelHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelHandler{level: h.level, handler: h.handler.WithAttrs(attrs)}
}

func (h *levelHandler) WithGroup(name string) slog.Handler {
	return &levelHandler{level: h.level, handler: h.handler.WithGroup(name)}
}

// Shutdown gracefully shuts down the logger (only needed when using OTEL)
// Call this during application shutdown
func Shutdown(ctx context.Context) error {
	if shutdownFunc != nil {
		return shutdownFunc(ctx)
	}
	return nil
}

// SetLevel sets the minimum log level for the logger
func SetLevel(level slog.Level) {
	programLevel.Set(level)
}

// GetLevel returns the current minimum log level
func GetLevel() slog.Level {
	return programLevel.Level()
}

// ParseLevel converts a string level name to slog.Level
func ParseLevel(levelStr string) (slog.Level, error) {
	switch strings.ToUpper(levelStr) {
	case "TRACE":
		return LevelTrace, nil
	case "DEBUG":
		return LevelDebug, nil
	case "INFO":
		return LevelInfo, nil
	case "WARN", "WARNING":
		return LevelWarning, nil
	case "ERROR":
		return LevelError, nil
	case "FATAL":
		return LevelFatal, nil
	default:
		return LevelInfo, fmt.Errorf("unknown log level: %s (defaulting to INFO)", levelStr)
	}
}

// SetLevelFromEnv sets the log level from an environment variable
// If the environment variable is not set or invalid, defaultLevel is used
func SetLevelFromEnv(envVarName string, defaultLevel slog.Level) {
	levelStr := os.Getenv(envVarName)
	if levelStr == "" {
		programLevel.Set(defaultLevel)
		return
	}

	level, err := ParseLevel(levelStr)
	if err != nil {
		// Use default level if parsing fails
		programLevel.Set(defaultLevel)
		return
	}
	programLevel.Set(level)
}

// ============================================================================
// Logging Functions
// ============================================================================

// Trace logs a trace-level message
func Trace(msg string, args ...any) {
	slog.Log(context.Background(), LevelTrace, msg, args...)
}

// Debug logs a debug-level message
func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

// Info logs an info-level message
func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

// Warn logs a warning-level message
func Warn(msg string, args ...any) {
	TotalWarnings.Add(1)
	Logger.Warn(msg, args...)
}

// Error logs an error-level message
func Error(msg string, args ...any) {
	TotalErrors.Add(1)
	Logger.Error(msg, args...)
}

// Fatal logs a fatal-level message and exits
func Fatal(msg string, args ...any) {
	slog.Log(context.Background(), LevelFatal, msg, args...)
	// Shutdown OTEL if enabled before exiting
	if shutdownFunc != nil {
		_ = shutdownFunc(context.Background())
	}
	os.Exit(1)
}

// ============================================================================
// HTTP-Specific Logging Helpers
// ============================================================================

// ErrorHttp5xx logs HTTP 5xx errors and increments counters
// Counters are always incremented regardless of sampling
func ErrorHttp5xx() {
	Total5xxErrors.Add(1)
	TotalErrors.Add(1)
}

// WarnHttp4xx logs HTTP 4xx warnings and increments counters
// Counters are always incremented regardless of sampling
func WarnHttp4xx(status int) {
	Total4xxErrors.Add(1)
	TotalWarnings.Add(1)

	// Track specific common 4xx codes
	switch status {
	case 400:
		Total400Errors.Add(1)
	case 404:
		Total404Errors.Add(1)
	case 429:
		Total429Errors.Add(1)
	}
}

// WarnSlowRequest logs slow request warnings and increments counter
// Counter is always incremented regardless of sampling
func WarnSlowRequest() {
	SlowRequests.Add(1)
	TotalWarnings.Add(1)
}

// WarnConnPool logs connection pool warnings and increments counter
// Counter is always incremented regardless of sampling
func WarnConnPool() {
	ConnPoolWarnings.Add(1)
	TotalWarnings.Add(1)
}
