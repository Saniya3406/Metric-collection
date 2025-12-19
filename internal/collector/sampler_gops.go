package collector

import (
	"context"
	"errors"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	"github.com/yourname/metric-agent/pkg/model"
)

type GopsSampler struct{}

// Sample implements Sampler by calling gopsutil functions.
// It returns aggregated values for a snapshot.
func (s *GopsSampler) Sample(ctx context.Context) (model.Metrics, error) {
	// CPU percentage: use a short 200ms sampling window for an instantaneous-ish value
	cpuPercents, err := cpu.PercentWithContext(ctx, 200*time.Millisecond, false)
	if err != nil {
		return model.Metrics{}, err
	}
	if len(cpuPercents) == 0 {
		return model.Metrics{}, errors.New("cpu.Percent returned empty")
	}

	vm, err := mem.VirtualMemoryWithContext(ctx)
	if err != nil {
		return model.Metrics{}, err
	}

	du, err := disk.UsageWithContext(ctx, "/")
	if err != nil {
		return model.Metrics{}, err
	}

	// network counters (per interface) -> sum
	ioCounters, err := net.IOCountersWithContext(ctx, false)
	if err != nil {
		return model.Metrics{}, err
	}
	var in, out uint64
	if len(ioCounters) > 0 {
		in = ioCounters[0].BytesRecv
		out = ioCounters[0].BytesSent
		// If you want to sum across interfaces:
		// for _, v := range ioCounters { in += v.BytesRecv; out += v.BytesSent }
	}

	return model.Metrics{
		Timestamp:   time.Now(),
		CPUPercent:  cpuPercents[0],
		MemPercent:  vm.UsedPercent,
		DiskPercent: du.UsedPercent,
		NetBytesIn:  in,
		NetBytesOut: out,
	}, nil
}
