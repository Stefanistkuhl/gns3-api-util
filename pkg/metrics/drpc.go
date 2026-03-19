package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type DRPCMetrics struct {
	RequestsTotal  *prometheus.CounterVec
	RequestLatency *prometheus.HistogramVec
	ErrorsTotal    *prometheus.CounterVec
}

func NewDRPCMetrics() *DRPCMetrics {
	return &DRPCMetrics{
		RequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "drpc_requests_total",
				Help: "Total number of DRPC requests.",
			},
			[]string{"method", "status"},
		),
		RequestLatency: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "drpc_request_duration_seconds",
				Help:    "DRPC request latency in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "status"},
		),
		ErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "drpc_errors_total",
				Help: "Total number of DRPC errors.",
			},
			[]string{"method", "error_type"},
		),
	}
}

func (m *DRPCMetrics) Observe(method string, statusCode int, started time.Time) {
	if m == nil {
		return
	}

	status := strconv.Itoa(statusCode)
	m.RequestsTotal.WithLabelValues(method, status).Inc()
	m.RequestLatency.WithLabelValues(method, status).
		Observe(time.Since(started).Seconds())
}

func (m *DRPCMetrics) ObserveError(method, errorType string) {
	if m == nil {
		return
	}
	m.ErrorsTotal.WithLabelValues(method, errorType).Inc()
}
