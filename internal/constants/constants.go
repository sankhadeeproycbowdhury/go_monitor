package constants

import (
	"time"

	"github.com/sankhadeeproycbowdhury/go_monitor/internal/types"
)

const (
	MsgMetrics    types.MessageType = "metrics"
	MsgScaleEvent types.MessageType = "scale_event"
	MsgError      types.MessageType = "error"
)

const (
	WriteWait      = 10 * time.Second
	PongWait       = 60 * time.Second
	PingPeriod     = (PongWait * 9) / 10
	MaxMessageSize = 512
)
