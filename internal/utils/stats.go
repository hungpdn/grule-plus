package utils

import (
	"runtime"
	"time"
)

// Stats holds the runtime statistics of the application.
type Stats struct {
	NumCPU          int
	NumGoroutine    int
	MemAlloc        uint64
	MemTotalAlloc   uint64
	MemSys          uint64
	MemHeapSys      uint64
	MemHeapIde      uint64
	MemHeapReleased uint64
	MemNumGC        uint32
	AverageGCPause  float64
}

// GetStats retrieves the current runtime statistics and returns them as a Stats struct.
func GetStats() Stats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	stats := Stats{
		NumCPU:          runtime.NumCPU(),
		NumGoroutine:    runtime.NumGoroutine(),
		MemAlloc:        ByteToMb(memStats.Alloc),
		MemTotalAlloc:   ByteToMb(memStats.TotalAlloc),
		MemSys:          ByteToMb(memStats.Sys),
		MemHeapSys:      ByteToMb(memStats.HeapSys),
		MemHeapIde:      ByteToMb(memStats.HeapIdle),
		MemHeapReleased: ByteToMb(memStats.HeapReleased),
		MemNumGC:        memStats.NumGC,
	}
	if memStats.NumGC > 0 {
		avgPause := memStats.PauseTotalNs / uint64(memStats.NumGC)
		stats.AverageGCPause = float64(avgPause) / float64(time.Millisecond)
	}

	return stats
}

// ByteToMb converts bytes to megabytes.
func ByteToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
