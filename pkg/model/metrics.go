package model

import "time"

// Metrics is a snapshot of collected metrics.
type Metrics struct {
	Timestamp   time.Time `json:"timestamp"`
	CPUPercent  float64   `json:"cpu_percent"`
	MemPercent  float64   `json:"mem_percent"`
	DiskPercent float64   `json:"disk_percent"`
	NetBytesIn  uint64    `json:"net_bytes_in"`
	NetBytesOut uint64    `json:"net_bytes_out"`
}
