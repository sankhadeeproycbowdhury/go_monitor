package models

import "time"

// MetricsSnapshot is sent to all WebSocket clients every tick.
type MetricsSnapshot struct {
    Timestamp   time.Time        `json:"timestamp"`
    CPU         CPUMetrics       `json:"cpu"`
    Memory      MemoryMetrics    `json:"memory"`
    Disk        []DiskMetrics    `json:"disk"`
    Network     []NetMetrics     `json:"network"`
    Host        HostMetrics      `json:"host"`
}

type CPUMetrics struct {
    UsagePercent   []float64 `json:"usage_percent"`   // per-core
    AvgPercent     float64   `json:"avg_percent"`
    CoreCount      int       `json:"core_count"`
    ThreadCount    int       `json:"thread_count"`
}

type MemoryMetrics struct {
    TotalBytes     uint64  `json:"total_bytes"`
    UsedBytes      uint64  `json:"used_bytes"`
    FreeBytes      uint64  `json:"free_bytes"`
    UsagePercent   float64 `json:"usage_percent"`
    SwapTotal      uint64  `json:"swap_total"`
    SwapUsed       uint64  `json:"swap_used"`
}

type DiskMetrics struct {
    Mountpoint     string  `json:"mountpoint"`
    TotalBytes     uint64  `json:"total_bytes"`
    UsedBytes      uint64  `json:"used_bytes"`
    FreeBytes      uint64  `json:"free_bytes"`
    UsagePercent   float64 `json:"usage_percent"`
}

type NetMetrics struct {
    Interface      string `json:"interface"`
    BytesSent      uint64 `json:"bytes_sent"`
    BytesRecv      uint64 `json:"bytes_recv"`
    PacketsSent    uint64 `json:"packets_sent"`
    PacketsRecv    uint64 `json:"packets_recv"`
}

type HostMetrics struct {
    Hostname       string `json:"hostname"`
    OS             string `json:"os"`
    Platform       string `json:"platform"`
    KernelVersion  string `json:"kernel_version"`
    UptimeSeconds  uint64 `json:"uptime_seconds"`
    Goroutines     int    `json:"goroutines"`
}