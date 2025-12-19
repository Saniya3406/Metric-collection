package collector

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/yourname/metric-agent/pkg/model"
)

type mockSampler struct {
	m model.Metrics
}

func (m *mockSampler) Sample(ctx context.Context) (model.Metrics, error) {
	return m.m, nil
}

func TestCollectorStartsAndSamples(t *testing.T) {
	now := time.Now()
	ms := &mockSampler{
		m: model.Metrics{
			Timestamp:   now,
			CPUPercent:  12.3,
			MemPercent:  45.6,
			DiskPercent: 78.9,
			NetBytesIn:  1000,
			NetBytesOut: 2000,
		},
	}

	reg := prometheus.NewRegistry()
	c := NewCollector(ms, 10*time.Millisecond, reg)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	c.Start(ctx)
	// allow at least one sample to happen
	time.Sleep(20 * time.Millisecond)

	last := c.Last()
	if last.CPUPercent != 12.3 {
		t.Fatalf("expected cpu 12.3 got %v", last.CPUPercent)
	}
	if last.NetBytesIn != 1000 {
		t.Fatalf("expected net in 1000 got %v", last.NetBytesIn)
	}

	c.Stop()
}
