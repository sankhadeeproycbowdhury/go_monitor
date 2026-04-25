package api

import (
	"encoding/json"
	"net/http"
	"runtime"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/sankhadeeproycbowdhury/go_monitor/internal/hub"
	"github.com/sankhadeeproycbowdhury/go_monitor/internal/models"
)

type Handler struct {
	hub *hub.Hub
}

func NewHandler(h *hub.Hub) *Handler {
	return &Handler{hub: h}
}

// GET /api/info — static server information
func (h *Handler) Info(w http.ResponseWriter, r *http.Request) {
	hostInfo, _ := host.Info()
	cpuInfo, _ := cpu.Info()
	cores, _ := cpu.Counts(false)
	threads, _ := cpu.Counts(true)
	vmStat, _ := mem.VirtualMemory()

	model := ""
	if len(cpuInfo) > 0 {
		model = cpuInfo[0].ModelName
	}

	info := models.ServerInfo{
		Hostname:      hostInfo.Hostname,
		OS:            hostInfo.OS,
		Platform:      hostInfo.Platform,
		Architecture:  runtime.GOARCH,
		CPUCores:      cores,
		CPUThreads:    threads,
		CPUModel:      model,
		TotalMemoryGB: float64(vmStat.Total) / (1 << 30),
		KernelVersion: hostInfo.KernelVersion,
		GoVersion:     runtime.Version(),
	}

	writeJSON(w, http.StatusOK, info)
}

// GET /api/health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"clients": h.hub.Count(),
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
