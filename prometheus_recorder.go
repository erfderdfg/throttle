package limiter

import "github.com/prometheus/client_golang/prometheus"

// PrometheusRecorder implements MetricsRecorder using Prometheus counters and histograms.
type PrometheusRecorder struct {
	callsTotal  *prometheus.CounterVec
	errorsTotal *prometheus.CounterVec
	latency     *prometheus.HistogramVec
}
