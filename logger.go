package chix

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZapLogger is a structured logger backed by uber/zap.
type ZapLogger struct {
	Logger *zap.Logger
}

// NewZapLogger is a middleware for go.uber.org/zap to log requests.
func NewZapLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&ZapLogger{logger})
}

// NewLogEntry creates a new ZapLogEntry for the request.
func (l *ZapLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	var logFields []zapcore.Field

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields = append(logFields, zap.String("request_id", reqID))
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	logFields = append(logFields, zap.String("http_proto", r.Proto))
	logFields = append(logFields, zap.String("http_schema", scheme))
	logFields = append(logFields, zap.String("http_method", r.Method))
	logFields = append(logFields, zap.String("uri", fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)))
	logFields = append(logFields, zap.String("remote_addr", r.RemoteAddr))

	return &ZapLogEntry{
		Logger: l.Logger.With(logFields...),
	}
}

// ZapLogEntry records the final log when a request completes.
type ZapLogEntry struct {
	Logger *zap.Logger
}

// Write ...
func (e *ZapLogEntry) Write(status, bytes int, elapsed time.Duration) {
	e.Logger = e.Logger.With(
		zap.Int("resp_status", status),
		zap.Int("resp_bytes_length", bytes),
		zap.Duration("resp_elasped", elapsed),
	)

	e.Logger.Info("request complete")
}

// Panic ...
func (e *ZapLogEntry) Panic(v interface{}, stack []byte) {
	e.Logger = e.Logger.With(
		zap.String("stack", string(stack)),
		zap.String("panic", fmt.Sprintf("%+v", v)),
	)
}
