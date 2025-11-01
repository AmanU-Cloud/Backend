package main

import (
	"net/http"

	"github.com/Backend/reviewer/internal/handler"
	"github.com/google/uuid" //добавил и использовал, чтобы появился go.sum
)

func main() {
	_ = uuid.New()
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	h := handler.CORS()(mux)

	http.ListenAndServe(":8080", h)
}
