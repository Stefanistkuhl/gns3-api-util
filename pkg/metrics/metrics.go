package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type Config struct {
	Enabled bool
}

type Manager struct {
	Enabled  bool
	Registry *prometheus.Registry
}

func New(cfg Config) *Manager {
	m := &Manager{
		Enabled: cfg.Enabled,
	}

	if !cfg.Enabled {
		return m
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	m.Registry = reg
	return m
}

func (m *Manager) Register(cs ...prometheus.Collector) {
	if !m.Enabled || m.Registry == nil {
		return
	}
	m.Registry.MustRegister(cs...)
}
