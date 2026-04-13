package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsHandler returns an http.Handler that exposes Prometheus metrics.
// The default Go collector provides:
// - go_memstats_*: Memory statistics (heap, stack, GC)
// - go_goroutines: Number of goroutines
// - go_gc_duration_seconds: GC pause durations
// - go_cpu_classes_*: CPU usage breakdown (Go 1.21+)
func MetricsHandler() http.Handler {
	registry := prometheus.NewRegistry()
	// Register default Go collector - provides memory, GC, CPU, goroutine metrics
	registry.MustRegister(prometheus.NewGoCollector())
	// Register process collector for CPU, memory, file descriptors- requires procfs
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
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