package main

import (
	"encoding/json"
	"fmt"
	"github.com/Backend/reviewer/internal/memecached"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"gopkg.in/yaml.v3"
)

type conf struct {
	Server struct {
		Port string `yaml:"port"`
		Host string `yaml:"host"`
	}
	Memcached struct {
		Servers string `yaml:"servers"`
		Ttl     int64  `yaml:"ttl"`
	}
}

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

	//эндпоинт для memcache GET localhost:8080/bubble?n=100000.
	//вызывать повторно, ttl - 1 минута
	//работает только через докер
	mux.HandleFunc("/bubble", bubbleHandler)

	cachePort := config.Memcached.Servers
	cacheTTL := config.Memcached.Ttl
	fmt.Println("config", config)
	memecached.Init(cachePort, time.Duration(cacheTTL)*time.Second)

	port := config.Server.Port
	host := config.Server.Host

	fmt.Printf("Listening on %v : %v", host, port)
	err := http.ListenAndServe(host+":"+port, mux)
	if err != nil {
		log.Fatalf("Error starting HTTP server: %V", err)
	}
}

func bubbleSort(arr []int) {
	n := len(arr)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-1-i; j++ {
			if arr[j] > arr[j+1] {
				arr[j], arr[j+1] = arr[j+1], arr[j]
			}
		}
	}
}

func bubbleHandler(w http.ResponseWriter, r *http.Request) {
	str := r.URL.Query().Get("n")
	if str == "" {
		http.Error(w, "missing query param 'n'", http.StatusBadRequest)
		return
	}

	n, err := strconv.Atoi(str)
	if err != nil {
		http.Error(w, "invalid 'n': must be integer", http.StatusBadRequest)
		return
	}

	cacheKey := "bubble:" + str

	// Пытаемся получить из кеша
	if item, err := memecached.Client().Get(cacheKey); err == nil {
		fmt.Println("cache has data")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		w.Write(item.Value)
		return
	}

	// Генерируем массив
	arr := make([]int, n)
	for i, value := range rand.Perm(n) {
		arr[i] = value
	}
	fmt.Println(arr)

	bubbleSort(arr)

	jsonData, err := json.Marshal(arr)
	if err != nil {
		http.Error(w, "failed to marshal result", http.StatusInternalServerError)
		return
	}

	fmt.Println("cache has No data")
	err = memecached.Client().Set(&memcache.Item{
		Key:        cacheKey,
		Value:      jsonData,
		Expiration: int32(memecached.DefaultTTL().Seconds()),
	})
	if err != nil {
		http.Error(w, "failed to set cache", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	w.Write(jsonData)
}
