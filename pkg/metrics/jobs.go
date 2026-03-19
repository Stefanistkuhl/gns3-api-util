package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type JobMetrics struct {
	RunsTotal      *prometheus.CounterVec
	RunDuration    *prometheus.HistogramVec
	ItemsProcessed *prometheus.CounterVec
	ErrorsTotal    *prometheus.CounterVec
}

func NewJobMetrics() *JobMetrics {
	return &JobMetrics{
		RunsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "job_runs_total",
				Help: "Total number of background job runs.",
			},
			[]string{"job", "result"},
		),
		RunDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "job_duration_seconds",
				Help:    "Background job duration in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"job", "result"},
		),
		ItemsProcessed: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "job_items_processed_total",
				Help: "Total items processed by background jobs.",
			},
			[]string{"job"},
		),
		ErrorsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "job_errors_total",
				Help: "Total job errors.",
			},
			[]string{"job", "error_type"},
		),
	}
}

func (m *JobMetrics) Observe(job string, started time.Time, ok bool) {
	if m == nil {
		return
	}
	result := "ok"
	if !ok {
		result = "error"
	}
	m.RunsTotal.WithLabelValues(job, result).Inc()
	m.RunDuration.WithLabelValues(job, result).
		Observe(time.Since(started).Seconds())
}
