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
	Version         string   `json:"version"`
	Compiler        string   `json:"compiler"`
	OperatingSystem string   `json:"operatingSystem"`
	NumGoroutines   int      `json:"numGoroutines"`
	NumCPUs         int      `json:"numCPUs"`
	MaxCPUs         int      `json:"maxCPUs"`
	MemStats        memStats `json:"memStats"`
	Uptime          string   `json:"uptime"`
}

type memStats struct {
	Alloc       uint64 `json:"alloc"`
	TotalAlloc  uint64 `json:"totalAlloc"`
	HeapSize    uint64 `json:"heapSize"`
	HeapUsed    uint64 `json:"heapUsed"`
	HeapUnused  uint64 `json:"heapUnused"`
	Lookups     uint64 `json:"lookups"`
	Mallocs     uint64 `json:"mallocs"`
	Frees       uint64 `json:"frees"`
	TotalGCTime string `json:"totalGCTime"`
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
		_, _ = fmt.Fprintf(w, string(data[:]))
	})
}

func Ping() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, _ = fmt.Fprintf(w, "pong")
	})
}