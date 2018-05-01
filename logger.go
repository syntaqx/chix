package chix

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewZapLogger ...
func NewZapLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&ZapLogger{logger})
}

// ZapLogger ...
type ZapLogger struct {
	Logger *zap.Logger
}

// NewLogEntry ...
func (l *ZapLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	entry := &ZapLoggerEntry{Logger: l.Logger}

	var logFields []zapcore.Field

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields = append(logFields, zap.String("req_id", reqID))
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	logFields = append(logFields, zap.String("http_schema", scheme))
	logFields = append(logFields, zap.String("http_proto", r.Proto))
	logFields = append(logFields, zap.String("http_method", r.Method))
	logFields = append(logFields, zap.String("remote_addr", r.RemoteAddr))
	logFields = append(logFields, zap.String("user_agent", r.UserAgent()))
	logFields = append(logFields, zap.String("uri", fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)))

	entry.Logger = entry.Logger.With(logFields...)

	return entry
}

// ZapLoggerEntry ...
type ZapLoggerEntry struct {
	Logger *zap.Logger
}

// Write ...
func (e *ZapLoggerEntry) Write(status, bytes int, elapsed time.Duration) {
	e.Logger = e.Logger.With(
		zap.Int("resp_status", status),
		zap.Int("resp_bytes_length", bytes),
		zap.Duration("resp_elasped", elapsed),
	)

	e.Logger.Info("request complete")
}

func (e *ZapLoggerEntry) Panic(v interface{}, stack []byte) {
	e.Logger = e.Logger.With(
		zap.String("stack", string(stack)),
		zap.String("panic", fmt.Sprintf("%+v", v)),
	)
}
