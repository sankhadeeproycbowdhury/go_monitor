package models

import (
	"time"
	"github.com/sankhadeeproycbowdhury/go_monitor/internal/types"
)


// WSMessage is the envelope wrapping every outbound WebSocket payload.
type WSMessage struct {
    Type    types.MessageType `json:"type"`
    Payload interface{}       `json:"payload"`
}

// ScaleEvent is broadcast when Kubernetes scales a deployment.
type ScaleEvent struct {
    Timestamp    time.Time `json:"timestamp"`
    Namespace    string    `json:"namespace"`
    Deployment   string    `json:"deployment"`
    OldReplicas  int32     `json:"old_replicas"`
    NewReplicas  int32     `json:"new_replicas"`
    Reason       string    `json:"reason"` // "cpu_high", "cpu_low", "manual"
}

// ServerInfo is returned by the REST /info endpoint.
type ServerInfo struct {
    Hostname      string `json:"hostname"`
    OS            string `json:"os"`
    Platform      string `json:"platform"`
    Architecture  string `json:"architecture"`
    CPUCores      int    `json:"cpu_cores"`
    CPUThreads    int    `json:"cpu_threads"`
    CPUModel      string `json:"cpu_model"`
    TotalMemoryGB float64 `json:"total_memory_gb"`
    KernelVersion string `json:"kernel_version"`
    GoVersion     string `json:"go_version"`
}