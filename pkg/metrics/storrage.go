package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type StorageMetrics struct {
	FileWritesTotal   *prometheus.CounterVec
	FileWriteBytes    *prometheus.CounterVec
	FileWriteDuration *prometheus.HistogramVec
	FileReadDuration  *prometheus.HistogramVec
	ActiveUploads     *prometheus.GaugeVec
}

func NewStorageMetrics() *StorageMetrics {
	return &StorageMetrics{
		FileWritesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "file_writes_total",
				Help: "Total number of file writes.",
			},
			[]string{"bucket", "result"},
		),
		FileWriteBytes: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "file_write_bytes_total",
				Help: "Total bytes written to files.",
			},
			[]string{"bucket"},
		),
		FileWriteDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "file_write_duration_seconds",
				Help:    "Duration of file writes.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"bucket", "result"},
		),
		FileReadDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "file_read_duration_seconds",
				Help:    "Duration of file reads.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"bucket", "result"},
		),
		ActiveUploads: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "active_uploads",
				Help: "Number of currently active uploads.",
			},
			[]string{"bucket"},
		),
	}
}

func (m *StorageMetrics) ObserveWrite(bucket string, bytes int64, started time.Time, ok bool) {
	if m == nil {
		return
	}
	result := "ok"
	if !ok {
		result = "error"
	}
	m.FileWritesTotal.WithLabelValues(bucket, result).Inc()
	m.FileWriteBytes.WithLabelValues(bucket).Add(float64(bytes))
	m.FileWriteDuration.WithLabelValues(bucket, result).
		Observe(time.Since(started).Seconds())
}

func (m *StorageMetrics) ObserveRead(bucket string, started time.Time, ok bool) {
	if m == nil {
		return
	}
	result := "ok"
	if !ok {
		result = "error"
	}
	m.FileReadDuration.WithLabelValues(bucket, result).
		Observe(time.Since(started).Seconds())
}
