package chix

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	dflBuckets = []float64{10, 25, 50}
)

const (
	reqsName    = "http_requests_total"
	latencyName = "http_request_duration_milliseconds"
)

// PrometheusMiddleware ...
type PrometheusMiddleware struct {
	reqs    *prometheus.CounterVec
	latency *prometheus.HistogramVec
}

// NewPrometheusMetrics ...
func NewPrometheusMetrics(name string, buckets ...float64) func(next http.Handler) http.Handler {
	var m PrometheusMiddleware
	m.reqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        reqsName,
			Help:        "How many HTTP requests processed, partitioned by status code, method and HTTP path.",
			ConstLabels: prometheus.Labels{"service": name},
		},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(m.reqs)

	if len(buckets) == 0 {
		buckets = dflBuckets
	}
	m.latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:        latencyName,
		Help:        "How long it took to process the request, partitioned by status code, method and HTTP path.",
		ConstLabels: prometheus.Labels{"service": name},
		Buckets:     buckets,
	},
		[]string{"code", "method", "path"},
	)
	prometheus.MustRegister(m.latency)

	return prometheusMetrics(&m)
}

func prometheusMetrics(m *PrometheusMiddleware) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			start := time.Now()
			defer func(start time.Time) {
				m.reqs.WithLabelValues(http.StatusText(ww.Status()), r.Method, r.URL.Path).Inc()
				m.latency.WithLabelValues(http.StatusText(ww.Status()), r.Method, r.URL.Path).Observe(float64(time.Since(start).Nanoseconds() / 1000000))
			}(start)

			next.ServeHTTP(ww, r)
		}
		return http.HandlerFunc(fn)
	}
}
