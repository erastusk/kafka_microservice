package main

import (
	"log"
	"time"
)

func MiddlewareReceiver(h func() (int, float64, float64)) (int, float64, float64) {
	start := time.Now()
	defer func() {
		log.Println("Took", time.Since(start))
	}()
	return h()
}
