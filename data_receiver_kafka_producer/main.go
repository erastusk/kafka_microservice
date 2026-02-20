// Package main implements an HTTP/WebSocket server that receives GPS coordinates
// and publishes them to a Kafka topic. This acts as a bridge between WebSocket
// clients and Kafka.
//
// Architecture:
//
//	WebSocket Client → HTTP Server (this) → Kafka Producer → Kafka Broker
//
// The server exposes two endpoints:
//   - /ws: WebSocket endpoint for receiving GPS coordinates
//   - /metrics: Prometheus metrics endpoint for monitoring
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/erastusk/gpscords/data_receiver_kafka_producer/handlers"
)

// addr defines the HTTP server address and port.
// Can be overridden via command-line flag: -addr=host:port
var addr = flag.String("addr", "localhost:30000", "http service address")

// main initializes and starts the HTTP server with WebSocket support.
//
// Endpoints:
//   - GET/POST /ws: Upgrade HTTP connection to WebSocket for GPS data streaming
//   - GET /metrics: Prometheus metrics (request counts, latencies, etc.)
//
// The server runs indefinitely until terminated or an error occurs.
func main() {
	// Parse command-line flags
	flag.Parse()

	// Register WebSocket handler for GPS coordinate reception
	http.HandleFunc("/ws", handlers.ReceiveWs)

	// Register Prometheus metrics handler for monitoring
	http.Handle("/metrics", promhttp.Handler())

	log.Printf("Starting GPS receiver server on %s", *addr)
	log.Println("Endpoints:")
	log.Println("  - ws://localhost:30000/ws (WebSocket for GPS data)")
	log.Println("  - http://localhost:30000/metrics (Prometheus metrics)")

	// Start HTTP server (blocks until error or termination)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
