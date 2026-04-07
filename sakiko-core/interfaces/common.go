package interfaces

import (
	"encoding/json"
	"time"
)

type Mode string

const (
	ModeSerial   Mode = "serial"
	ModeParallel Mode = "parallel"
)

type RuntimeStatus struct {
	Running     bool `json:"running"`
	RunningTask int  `json:"runningTask"`
	TotalTask   int  `json:"totalTask"`
}

type RuntimeStatusResponse struct {
	Status RuntimeStatus `json:"status"`
}

type JSONValue = json.RawMessage

type Clock interface {
	Now() time.Time
}
