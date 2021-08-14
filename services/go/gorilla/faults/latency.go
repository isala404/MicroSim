package faults

import (
	"time"
)

type Latency struct {
	Delay int `json:"delay"` // In Millisecond
}

func (l Latency) Run() error {
	time.Sleep(time.Duration(l.Delay) * time.Millisecond)
	return nil
}
