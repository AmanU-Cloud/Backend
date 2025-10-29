package main

import (
	"log"
	"net/http"

	"github.com/Backend/internal/metrics"
)

func main() {
	metrics.InitMetrics()

	log.Println("Server run on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server run error: %v", err)
	}
}
