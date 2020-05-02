package chix

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewZapLogger(logger *zap.Logger) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&ZapLogger{logger: logger})
}

type ZapLogger struct {
	logger *zap.Logger
}

func (l *ZapLogger) NewLogEntry(r *http.Request) middleware.LogEntry {
	var logFields []zapcore.Field

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		logFields = append(logFields, zap.String("req_id", reqID))
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	logFields = append(logFields,
		zap.String("http_scheme", scheme),
		zap.String("http_proto", r.Proto),
		zap.String("http_method", r.Method),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.String("uri", fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)),
	)

	logger := l.logger.With(logFields...)
	logger.Info("request started")

	return &ZapLoggerEntry{logger: logger}
}

type ZapLoggerEntry struct {
	logger *zap.Logger
}

func (e *ZapLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	e.logger = e.logger.With(
		zap.Int("resp_status", status),
		zap.Int("resp_bytes_length", bytes),
		zap.Duration("resp_elapsed", elapsed),
	)

	e.logger.Info("request complete")
}

func (e *ZapLoggerEntry) Panic(v interface{}, stack []byte) {
	e.logger = e.logger.With(
		zap.ByteString("stack", stack),
		zap.Any("panic", v),
	)
}
