package limiter

import "github.com/prometheus/client_golang/prometheus"

// PrometheusRecorder implements MetricsRecorder using Prometheus counters and
// histograms. Inject it via WithRecorder to expose per-namespace allow/deny
// counters and p99 admission latency on a /metrics endpoint.
type PrometheusRecorder struct {
	callsTotal  *prometheus.CounterVec
	errorsTotal *prometheus.CounterVec
	latency     *prometheus.HistogramVec
}

// NewPrometheusRecorder creates a PrometheusRecorder and registers its metrics
// with reg. Pass prometheus.DefaultRegisterer for the default global registry.
func NewPrometheusRecorder(reg prometheus.Registerer) *PrometheusRecorder {
	r := &PrometheusRecorder{
		callsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "ratelimit_calls_total",
			Help: "Total rate-limit decisions, partitioned by namespace and status (allowed/denied).",
		}, []string{"namespace", "status"}),

		errorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "ratelimit_errors_total",
			Help: "Total rate-limit errors, partitioned by namespace and error type.",
		}, []string{"namespace", "type"}),

		latency: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "ratelimit_latency_seconds",
			Help:    "Admission-check latency in seconds, partitioned by namespace and status.",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05},
		}, []string{"namespace", "status"}),
	}
	reg.MustRegister(r.callsTotal, r.errorsTotal, r.latency)
	return r
}
