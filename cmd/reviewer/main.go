package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

import (
	"github.com/google/uuid" //добавил и использовал, чтобы появился go.sum
)

func main() {
	_ = uuid.New()

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %V", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello Caritas"))
	})

	port := os.Getenv("SERVER_PORT")

	fmt.Println("Listening on :8081")
	err = http.ListenAndServe(":"+port, mux)
	if err != nil {
		log.Fatalf("Error starting HTTP server: %V", err)
	}
	fmt.Println("Server has started")
}
