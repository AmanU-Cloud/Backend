package main

import (
	"log"
	"net/http"

	"github.com/Caritas-Team/reviewer/internal/metrics"
	"github.com/google/uuid" //добавил и использовал, чтобы появился go.sum
)

func main() {
	_ = uuid.New()
	metrics.InitMetrics()

	log.Println("Server run on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server run error: %v", err)
	}
}
