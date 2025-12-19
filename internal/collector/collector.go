package collector

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/yourname/metric-agent/pkg/model"
)

type Sampler interface {
	Sample(ctx context.Context) (model.Metrics, error)
}

type Collector struct {
	sampler  Sampler
	interval time.Duration
	mu       sync.RWMutex
	last     model.Metrics
	quit     chan struct{}
	wg       sync.WaitGroup

	// prometheus metrics
	cpuGauge  prometheus.Gauge
	memGauge  prometheus.Gauge
	diskGauge prometheus.Gauge
	netIn     prometheus.Gauge
	netOut    prometheus.Gauge
}

func NewCollector(sampler Sampler, interval time.Duration, registry *prometheus.Registry) *Collector {
	c := &Collector{
		sampler:  sampler,
		interval: interval,
		quit:     make(chan struct{}),
		cpuGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "metric_agent_cpu_percent",
			Help: "CPU percent",
		}),
		memGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "metric_agent_mem_percent",
			Help: "Memory percent",
		}),
		diskGauge: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "metric_agent_disk_percent",
			Help: "Disk usage percent",
		}),
		netIn: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "metric_agent_net_bytes_in",
			Help: "Network bytes received (since boot)",
		}),
		netOut: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "metric_agent_net_bytes_out",
			Help: "Network bytes sent (since boot)",
		}),
	}

	if registry != nil {
		registry.MustRegister(c.cpuGauge, c.memGauge, c.diskGauge, c.netIn, c.netOut)
	} else {
		// fallback to default registry
		prometheus.MustRegister(c.cpuGauge, c.memGauge, c.diskGauge, c.netIn, c.netOut)
	}
	return c
}

// Start begins periodic sampling in a goroutine
func (c *Collector) Start(ctx context.Context) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()
		// perform initial sample immediately
		c.doSample(ctx)
		for {
			select {
			case <-ticker.C:
				c.doSample(ctx)
			case <-c.quit:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (c *Collector) Stop() {
	close(c.quit)
	c.wg.Wait()
}

// Last returns the latest snapshot
func (c *Collector) Last() model.Metrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.last
}

func (c *Collector) doSample(ctx context.Context) {
	m, err := c.sampler.Sample(ctx)
	if err != nil {
		// log short, but avoid adding a logger dependency in this starter
		return
	}
	c.mu.Lock()
	c.last = m
	c.mu.Unlock()

	// update prometheus
	c.cpuGauge.Set(m.CPUPercent)
	c.memGauge.Set(m.MemPercent)
	c.diskGauge.Set(m.DiskPercent)
	c.netIn.Set(float64(m.NetBytesIn))
	c.netOut.Set(float64(m.NetBytesOut))
}
