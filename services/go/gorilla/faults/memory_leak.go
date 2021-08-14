package faults

import (
	"log"
	"runtime"
	"strings"
	"time"
)

type MemoryLeak struct {
	Size     int `json:"size"`		// In Megabytes
	Duration int `json:"duration"`	// In Millisecond
}

func (m MemoryLeak) Run() error {

	// Creates a goroutine that writes a big string to memory
	go func() string {
		log.Printf("creating memory leak of %dMB for %d seconds", m.Size, m.Duration)
		leak := strings.Repeat("a", m.Size*1024*1024)

		// Wait for the Duration
		time.Sleep(time.Duration(m.Duration) * time.Millisecond)

		// Force garbage collector to clean up
		leak = ""
		time.Sleep(time.Second)
		runtime.GC()

		// This may take will second to take fully effect
		log.Println("Leak was closed")
		return leak
	}()
	return nil
}
