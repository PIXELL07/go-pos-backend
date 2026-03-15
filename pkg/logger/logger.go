// we r gonna use zap coz:
// Uber's Zap is the preferred logging library in Go for high-performance applications, offering structured, leveled logging with exceptionally low overhead.
// Developers use it because it avoids memory allocation and reflection, making it 4–10 times faster than other loggers.
// It is ideal for microservices needing fast JSON logging, structured context, and log sampling

package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger
var Sugar *zap.SugaredLogger

// Init initialises the global logger based on environment.
func Init(env string) {
	var cfg zap.Config

	if env == "production" {
		cfg = zap.NewProductionConfig()
		cfg.EncoderConfig.TimeKey = "timestamp"
		cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		}
	}

	var err error
	log, err = cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		// Fallback to basic logger
		log, _ = zap.NewDevelopment()
	}
	Sugar = log.Sugar()
}

// Get returns the underlying zap.Logger (use for middleware).
func Get() *zap.Logger {
	if log == nil {
		Init("development")
	}
	return log
}

// levelled helpers

func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
}

func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

// Sugared helpers (printf-style)

func Infof(format string, args ...interface{}) {
	if Sugar == nil {
		Init("development")
	}
	Sugar.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	Sugar.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	Sugar.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	Sugar.Fatalf(format, args...)
}

func Debugf(format string, args ...interface{}) {
	Sugar.Debugf(format, args...)
}

// Request logger for Gin

// FYI:
// returns a Gin-compatible middleware that logs each request.
func GinZapLogger() func(method, path string, status, latencyMs int, clientIP, userAgent string) {
	return func(method, path string, status, latencyMs int, clientIP, userAgent string) {
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Int("latency_ms", latencyMs),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
		}

		switch {
		case status >= 500:
			Get().Error("Server error", fields...)
		case status >= 400:
			Get().Warn("Client error", fields...)
		default:
			Get().Info("Request", fields...)
		}
	}
}

func Sync() { // Sync flushes any buffered log entries.
	if log != nil { // Call before application exits.
		_ = log.Sync()
	}
}

// returns a child logger with a request-id field.
func WithRequestID(requestID string) *zap.Logger {
	return Get().With(zap.String("request_id", requestID))
}

// returns a child logger with a user-id field.
func WithUserID(userID string) *zap.Logger {
	return Get().With(zap.String("user_id", userID))
}

// returns a sensible default when APP_ENV is not set.
func Env() string {
	if e := os.Getenv("APP_ENV"); e != "" {
		return e
	}
	return "development"
}
