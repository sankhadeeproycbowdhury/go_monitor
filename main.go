package main

import (
    "context"
    "log"
    "net/http"
    "os/signal"
    "syscall"

    "github.com/sankhadeeproycbowdhury/go_monitor/config"
    "github.com/sankhadeeproycbowdhury/go_monitor/internal/api"
    "github.com/sankhadeeproycbowdhury/go_monitor/internal/collector"
    "github.com/sankhadeeproycbowdhury/go_monitor/internal/hub"
    "github.com/sankhadeeproycbowdhury/go_monitor/internal/models"
    "github.com/sankhadeeproycbowdhury/go_monitor/internal/scaler"
)

func main() {
    cfg := config.Load()

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    // Channels
    metricsCh := make(chan models.MetricsSnapshot, 10)
    scaleEventsCh := make(chan models.ScaleEvent, 10)

    // Hub
    wsHub := hub.New()

    // Collector — polls system metrics
    coll := collector.New(cfg.MetricsInterval, metricsCh)
    go coll.Run(ctx)

    // Scaler — watches metrics and patches K8s
    sc, err := scaler.New(
        cfg.K8sNamespace, cfg.K8sDeployment,
        cfg.ScaleUpThreshold, cfg.ScaleDownThreshold,
        cfg.MaxReplicas, cfg.MinReplicas,
        scaleEventsCh,
    )
    if err != nil {
        log.Printf("scaler init failed (K8s unavailable?): %v", err)
        sc = nil
    }

    // Fan-out goroutine: reads from channels, broadcasts to WS clients
    go func() {
        for {
            select {
            case snap := <-metricsCh:
                wsHub.BroadcastMetrics(snap)
                if sc != nil {
                    sc.Evaluate(ctx, snap.CPU.AvgPercent)
                }
            case ev := <-scaleEventsCh:
                wsHub.BroadcastScaleEvent(ev)
            case <-ctx.Done():
                return
            }
        }
    }()

    // HTTP server
    wsHandler := func(w http.ResponseWriter, r *http.Request) {
        hub.ServeWS(wsHub, w, r)
    }
    router := api.NewRouter(wsHub, wsHandler)

    srv := &http.Server{
        Addr:    ":" + cfg.Port,
        Handler: router,
    }

    go func() {
        log.Printf("server listening on :%s", cfg.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("server error: %v", err)
        }
    }()

    <-ctx.Done()
    log.Println("shutting down...")
    srv.Shutdown(context.Background())
}