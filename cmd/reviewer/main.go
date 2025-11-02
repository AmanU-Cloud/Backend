package main

import (
	"log"
	"net/http"

	"github.com/Backend/reviewer/internal/config"
	"github.com/Backend/reviewer/internal/handler"
	"github.com/google/uuid" //добавил и использовал, чтобы появился go.sum
)

func main() {
	_ = uuid.New()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	h := handler.CORS(handler.CORSConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		AllowCredentials: cfg.CORS.AllowCredentials,
		MaxAgeSeconds:    cfg.CORS.MaxAgeSeconds,
	})(mux)

	_ = http.ListenAndServe(cfg.HTTP.Address, h)
}
