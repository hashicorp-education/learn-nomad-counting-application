package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

var (
	counter int
	mu      sync.Mutex
)

type CountResponse struct {
	Count        int    `json:"count"`
	Timestamp    string `json:"timestamp"`
	Architecture string `json:"architecture"`
	Platform     string `json:"platform"`
	GoVersion    string `json:"go_version"`
}

type HealthResponse struct {
	Status       string `json:"status"`
	Timestamp    string `json:"timestamp"`
	Architecture string `json:"architecture"`
	Platform     string `json:"platform"`
	GoVersion    string `json:"go_version"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9001"
	}

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/", countHandler)

	log.Printf("Starting counting service on port %s", port)
	log.Printf("Architecture: %s", runtime.GOARCH)
	log.Printf("Platform: %s", runtime.GOOS)
	log.Printf("Go version: %s", runtime.Version())

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	response := HealthResponse{
		Status:       "healthy",
		Timestamp:    time.Now().Format(time.RFC3339),
		Architecture: runtime.GOARCH,
		Platform:     runtime.GOOS,
		GoVersion:    runtime.Version(),
	}

	json.NewEncoder(w).Encode(response)
}

func countHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	mu.Lock()
	counter++
	currentCount := counter
	mu.Unlock()

	response := CountResponse{
		Count:        currentCount,
		Timestamp:    time.Now().Format(time.RFC3339),
		Architecture: runtime.GOARCH,
		Platform:     runtime.GOOS,
		GoVersion:    runtime.Version(),
	}

	log.Printf("Count request: %d", currentCount)
	json.NewEncoder(w).Encode(response)
}

// Made with Bob
