package scaler

import (
    "context"
    "log"
    "time"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"

    "github.com/sankhadeeproycbowdhury/go_monitor/internal/models"
)

// Scaler watches aggregated CPU metrics and patches Kubernetes replicas.
type Scaler struct {
    disabled    bool   // ← add this line
    client      kubernetes.Interface
    namespace   string
    deployment  string
    scaleUp     float64
    scaleDown   float64
    maxReplicas int32
    minReplicas int32
    events      chan<- models.ScaleEvent
    cooldown    time.Duration
    lastScaled  time.Time
}

func New(
    namespace, deployment string,
    scaleUp, scaleDown float64,
    maxR, minR int32,
    events chan<- models.ScaleEvent,
) (*Scaler, error) {
    client, err := buildClient()
    if err != nil {
        return nil, err
    }
    return &Scaler{
        client:      client,
        namespace:   namespace,
        deployment:  deployment,
        scaleUp:     scaleUp,
        scaleDown:   scaleDown,
        maxReplicas: maxR,
        minReplicas: minR,
        events:      events,
        cooldown:    2 * time.Minute,
    }, nil
}

// Evaluate checks cpu avg against thresholds and scales if needed.
func (s *Scaler) Evaluate(ctx context.Context, cpuAvg float64) {
    if s.disabled {         
        return
    }
    if time.Since(s.lastScaled) < s.cooldown {
        return
    }

    dep, err := s.client.AppsV1().Deployments(s.namespace).
        GetScale(ctx, s.deployment, metav1.GetOptions{})
    if err != nil {
        log.Printf("scaler: K8s unavailable, disabling scaler: %v", err)
        s.disabled = true    // ← add this line
        return
    }

    current := dep.Spec.Replicas
    desired := current

    switch {
    case cpuAvg > s.scaleUp && current < s.maxReplicas:
        desired = current + 1
    case cpuAvg < s.scaleDown && current > s.minReplicas:
        desired = current - 1
    default:
        return
    }

    dep.Spec.Replicas = desired
    if _, err := s.client.AppsV1().Deployments(s.namespace).
        UpdateScale(ctx, s.deployment, dep, metav1.UpdateOptions{}); err != nil {
        log.Printf("scaler: update scale error: %v", err)
        return
    }

    s.lastScaled = time.Now()
    reason := "cpu_high"
    if desired < current {
        reason = "cpu_low"
    }

    ev := models.ScaleEvent{
        Timestamp:   time.Now(),
        Namespace:   s.namespace,
        Deployment:  s.deployment,
        OldReplicas: current,
        NewReplicas: desired,
        Reason:      reason,
    }
    log.Printf("scaler: scaled %s %d→%d (%s)", s.deployment, current, desired, reason)

    select {
    case s.events <- ev:
    default:
    }
}

func buildClient() (kubernetes.Interface, error) {
    // Try in-cluster first (running inside a pod)
    cfg, err := rest.InClusterConfig()
    if err != nil {
        // Fall back to local kubeconfig (development)
        cfg, err = clientcmd.BuildConfigFromFlags("",
            clientcmd.RecommendedHomeFile)
        if err != nil {
            return nil, err
        }
    }
    return kubernetes.NewForConfig(cfg)
}