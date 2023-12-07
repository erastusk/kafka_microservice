package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/erastusk/gpscords/data_receiver_kafka_producer/handlers"
)

var addr = flag.String("addr", "localhost:30000", "http service address")

func main() {
	http.HandleFunc("/ws", handlers.ReceiveWs)
	http.Handle("/metrics", promhttp.Handler())
	log.Println("starting server")
	log.Fatal(http.ListenAndServe(*addr, nil))
}
