package main

import (
	"github.com/google/uuid" //добавил и использовал, чтобы появился go.sum
)

func main() {
	_ = uuid.New()
}
