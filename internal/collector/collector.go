package collector

import (
    "context"
    "log"
    "runtime"
    "time"

    "github.com/shirou/gopsutil/v3/cpu"
    "github.com/shirou/gopsutil/v3/disk"
    "github.com/shirou/gopsutil/v3/host"
    "github.com/shirou/gopsutil/v3/mem"
    "github.com/shirou/gopsutil/v3/net"

    "github.com/sankhadeeproycbowdhury/go_monitor/internal/models"
)

// Collector polls system metrics at a fixed interval and emits snapshots.
type Collector struct {
    interval time.Duration
    out      chan<- models.MetricsSnapshot
}

func New(interval time.Duration, out chan<- models.MetricsSnapshot) *Collector {
    return &Collector{interval: interval, out: out}
}

// Run blocks until ctx is cancelled.
func (c *Collector) Run(ctx context.Context) {
    ticker := time.NewTicker(c.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            snap, err := c.collect()
            if err != nil {
                log.Printf("collector error: %v", err)
                continue
            }
            select {
            case c.out <- snap:
            default:
                // channel full — drop the snapshot rather than block
                log.Println("collector: channel full, dropping snapshot")
            }
        }
    }
}

func (c *Collector) collect() (models.MetricsSnapshot, error) {
    snap := models.MetricsSnapshot{Timestamp: time.Now()}

    // --- CPU ---
    perCore, err := cpu.Percent(0, true)
    if err != nil {
        return snap, err
    }
    avg, _ := cpu.Percent(0, false)
    info, _ := cpu.Info()
    cores, _ := cpu.Counts(false)
    threads, _ := cpu.Counts(true)

    snap.CPU = models.CPUMetrics{
        UsagePercent: perCore,
        AvgPercent:   safeAvg(avg),
        CoreCount:    cores,
        ThreadCount:  threads,
    }
    if len(info) > 0 {
        snap.CPU.CoreCount = cores
    }

    // --- Memory ---
    vmStat, err := mem.VirtualMemory()
    if err != nil {
        return snap, err
    }
    swapStat, _ := mem.SwapMemory()
    snap.Memory = models.MemoryMetrics{
        TotalBytes:   vmStat.Total,
        UsedBytes:    vmStat.Used,
        FreeBytes:    vmStat.Available,
        UsagePercent: vmStat.UsedPercent,
        SwapTotal:    swapStat.Total,
        SwapUsed:     swapStat.Used,
    }

    // --- Disk ---
    parts, _ := disk.Partitions(false)
    for _, p := range parts {
        usage, err := disk.Usage(p.Mountpoint)
        if err != nil {
            continue
        }
        snap.Disk = append(snap.Disk, models.DiskMetrics{
            Mountpoint:   p.Mountpoint,
            TotalBytes:   usage.Total,
            UsedBytes:    usage.Used,
            FreeBytes:    usage.Free,
            UsagePercent: usage.UsedPercent,
        })
    }

    // --- Network ---
    ioCounters, _ := net.IOCounters(true)
    for _, ioc := range ioCounters {
        snap.Network = append(snap.Network, models.NetMetrics{
            Interface:   ioc.Name,
            BytesSent:   ioc.BytesSent,
            BytesRecv:   ioc.BytesRecv,
            PacketsSent: ioc.PacketsSent,
            PacketsRecv: ioc.PacketsRecv,
        })
    }

    // --- Host ---
    hostInfo, _ := host.Info()
    snap.Host = models.HostMetrics{
        Hostname:      hostInfo.Hostname,
        OS:            hostInfo.OS,
        Platform:      hostInfo.Platform,
        KernelVersion: hostInfo.KernelVersion,
        UptimeSeconds: hostInfo.Uptime,
        Goroutines:    runtime.NumGoroutine(),
    }

    return snap, nil
}

func safeAvg(vals []float64) float64 {
    if len(vals) == 0 {
        return 0
    }
    return vals[0]
}