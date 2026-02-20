// Package main implements a GPS data producer that simulates On-Board Unit (OBU)
// devices sending GPS coordinates via WebSocket.
//
// Architecture:
//
//	GPS Producer (this) → WebSocket → Data Receiver Server → Kafka
//
// This producer:
//   - Generates random GPS coordinates (simulates OBU devices)
//   - Sends coordinates every second via WebSocket
//   - Measures and logs latency for each send operation
//   - Runs indefinitely until terminated
package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"

	"github.com/erastusk/gpscords/types"
)

// retOBUdata generates random GPS coordinates simulating an On-Board Unit.
//
// Returns:
//   - OBUID: Random integer for device identification
//   - Latitude: Random float between 1.0 and 100.0
//   - Longitude: Random float between 1.0 and 100.0
//
// Note: In production, this would read from actual GPS hardware or a GPS API.
func retOBUdata() (int, float64, float64) {
	return rand.Int(), rand.Float64()*99 + 1, rand.Float64()*99 + 1
}

// wsEndpoint defines the WebSocket server address to connect to.
// Must match the data receiver server's /ws endpoint.
var wsEndpoint = "ws://localhost:30000/ws"

// main establishes a WebSocket connection and continuously sends GPS coordinates.
//
// Flow:
//  1. Connect to WebSocket server
//  2. Generate random GPS coordinates
//  3. Wrap coordinates in timing middleware
//  4. Send JSON-encoded data via WebSocket
//  5. Wait 1 second and repeat
//
// Error handling:
//   - Connection errors are logged and program exits
//   - Send errors are logged but program continues
func main() {
	// Seed random number generator for varied GPS coordinates
	rand.Seed(time.Now().UnixNano())

	log.Printf("Connecting to WebSocket server at %s...", wsEndpoint)

	// Establish WebSocket connection to data receiver
	conn, _, err := websocket.DefaultDialer.Dial(wsEndpoint, nil)
	if err != nil {
		log.Fatalf("ERROR: Couldn't dial WebSocket server: %v", err)
	}
	defer conn.Close()

	log.Println("WebSocket connection established, starting to send GPS data...")

	// Continuously generate and send GPS coordinates
	for {
		// Generate GPS data with timing middleware
		a, b, c := MiddlewareReceiver(retOBUdata)

		// Create GPS coordinate struct
		t := types.SourceCoords{
			OBUID: a,
			Lat:   b,
			Lon:   c,
		}

		// Wait 1 second between sends to simulate real-time GPS updates
		time.Sleep(time.Second)

		// Log the data being sent
		fmt.Printf("Producer sending: OBUID=%d, Lat=%.2f, Lon=%.2f\n", t.OBUID, t.Lat, t.Lon)

		// Send JSON-encoded GPS data over WebSocket
		err = conn.WriteJSON(t)
		if err != nil {
			log.Printf("ERROR: Unable to write message: %v", err)
			// In production, consider reconnection logic here
		}
	}
}
