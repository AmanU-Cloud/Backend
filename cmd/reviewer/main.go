package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type conf struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	}
}
import (
	"github.com/google/uuid" //добавил и использовал, чтобы появился go.sum
)

func main() {
	_ = uuid.New()

func (c *conf) getConf() *conf {
	yamlFile, err := os.ReadFile("cfg/config.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	return c
}

func main() {

	var config conf
	config.getConf()

	mux := http.NewServeMux()

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("hello Caritas"))
		if err != nil {
			return
		}
	})
	port := config.Server.Port
	host := config.Server.Host

	fmt.Printf("Listening on %v : %v", host, port)
	err := http.ListenAndServe(host+":"+port, mux)
	if err != nil {
		log.Fatalf("Error starting HTTP server: %V", err)
	}
	fmt.Println("Server has started")
}
