package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"
)

type monitoring struct {
	Version         string
	Compiler        string
	OperatingSystem string
	NumGoroutines   int
	NumCPUs         int
	MaxCPUs         int
	MemStats		memStats
	Uptime          string
}

type memStats struct {
	Alloc uint64
	TotalAlloc uint64
	HeapSize uint64
	HeapUsed uint64
	HeapUnused uint64
	Lookups uint64
	Mallocs uint64
	Frees uint64
	TotalGCTime string
}

func Monitoring(startTime time.Time) http.Handler {
	timeStarted := startTime
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		stats := runtime.MemStats{}
		runtime.ReadMemStats(&stats)
		totalGCTime, _ := time.ParseDuration(fmt.Sprintf("%vns", stats.PauseTotalNs))
		memStats := memStats{
			stats.Alloc,
			stats.TotalAlloc,
			stats.HeapSys,
			stats.HeapInuse,
			stats.HeapIdle,
			stats.Lookups,
			stats.Mallocs,
			stats.Frees,
			totalGCTime.String(),
		}
		monitoring := monitoring{
			runtime.Version(),
			runtime.Compiler,
			runtime.GOARCH + " " + runtime.GOOS,
			runtime.NumGoroutine(),
			runtime.NumCPU(),
			runtime.GOMAXPROCS(-1),
			memStats,
			time.Since(timeStarted).String(),
		}

		data, err := json.MarshalIndent(monitoring, "", "    ")
		if err != nil {
			log.Println("Couldn't create process info:", err)
		}
		fmt.Fprintf(w, string(data[:]))
	})
}
