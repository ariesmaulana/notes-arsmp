package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsRegistry = prometheus.NewRegistry()

	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests by method and status code.",
		},
		[]string{"method", "code"},
	)
)

func init() {
	// WrapRegistererWith adds service="notes-arsmp" to every metric emitted,
	// so the label is present regardless of what Alloy's relabeling does.
	wrapped := prometheus.WrapRegistererWith(prometheus.Labels{"service": "notes-arsmp"}, metricsRegistry)
	// Register default Go collector - provides memory, GC, CPU, goroutine metrics
	wrapped.MustRegister(prometheus.NewGoCollector())
	// Register process collector for CPU, memory, file descriptors
	wrapped.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	// Register HTTP request counter
	wrapped.MustRegister(httpRequestsTotal)
}

// MetricsHandler returns an http.Handler that exposes Prometheus metrics.
func MetricsHandler() http.Handler {
	return promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{})
}

// statusRecorder wraps http.ResponseWriter to capture the status code written.
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.statusCode = code
	sr.ResponseWriter.WriteHeader(code)
}

// HTTPMetricsMiddleware records http_requests_total for every request,
// labelled by HTTP method and response status code.
// The /metrics endpoint itself is excluded to avoid noise.
func HTTPMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}
		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)
		httpRequestsTotal.WithLabelValues(r.Method, strconv.Itoa(rec.statusCode)).Inc()
	})
}

// StartMetricsServer starts a dedicated metrics server on the given address.
// Use this if you want metrics on a separate port (e.g., :8081).
func StartMetricsServer(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", MetricsHandler())
	log.Printf("[metrics] server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("[metrics] server error: %v", err)
	}
}
