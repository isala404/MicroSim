package faults

import (
	"time"
)

type Latency struct {
	Delay int64 `json:"delay"`
}

func (l Latency) Run() error {
	time.Sleep(time.Duration(l.Delay) * time.Millisecond)
	return nil
}
