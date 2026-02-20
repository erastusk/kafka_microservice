// Package handlers provides HTTP and WebSocket handlers for receiving GPS coordinate data.
// This acts as the entry point for GPS data flowing into the Kafka pipeline.
package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/erastusk/gpscords/data_receiver_kafka_producer/kafka"
	"github.com/erastusk/gpscords/types"
)

// upgrader configures WebSocket connection upgrade parameters.
// Defines buffer sizes for reading from and writing to WebSocket connections.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1028, // Buffer for incoming messages (1KB)
	WriteBufferSize: 1028, // Buffer for outgoing messages (1KB)
}

// ReceiveWs handles HTTP requests and upgrades them to WebSocket connections.
//
// Flow:
//   1. HTTP client requests upgrade to WebSocket
//   2. Upgrade HTTP connection to WebSocket protocol
//   3. Start reading messages from the WebSocket connection
//
// This handler is registered at the /ws endpoint and serves as the entry point
// for GPS data producers to send coordinates.
//
// Parameters:
//   - w: HTTP ResponseWriter for sending responses
//   - r: HTTP Request containing upgrade headers
//
// Error handling:
//   - Connection upgrade errors are logged
//   - Connection is passed to ReadMessageLoop for message processing
func ReceiveWs(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ERROR: WebSocket upgrade failed: %v", err)
		return
	}
	
	log.Println("WebSocket connection established, starting message loop...")
	
	// Start reading messages from the WebSocket connection
	ReadMessageLoop(c)
}

// ReadMessageLoop continuously reads GPS coordinate messages from a WebSocket connection
// and publishes them to Kafka.
//
// Flow:
//   1. Create Kafka producer instance
//   2. Loop indefinitely:
//      a. Read JSON message from WebSocket
//      b. Parse into SourceCoords struct
//      c. Marshal back to JSON for Kafka
//      d. Send to Kafka via middleware (with timing)
//   3. Handle errors and unexpected connection closures
//
// Parameters:
//   - c: Active WebSocket connection
//
// The connection is closed automatically when:
//   - Client disconnects
//   - Read error occurs
//   - Unexpected closure is detected
//
// Error handling:
//   - JSON unmarshal errors are logged but don't break the loop
//   - Unexpected closures log error and terminate loop
//   - Connection is always closed via defer
func ReadMessageLoop(c *websocket.Conn) {
	// Ensure WebSocket connection is closed when function exits
	defer c.Close()
	
	// Initialize Kafka producer for publishing GPS coordinates
	k, err := kafka.NewKafkaProducer()
	if err != nil {
		log.Printf("ERROR: Failed to create Kafka producer: %v", err)
		return
	}
	
	var recv types.SourceCoords
	
	// Continuously read and process messages
	for {
		// Read GPS coordinates from WebSocket as JSON
		err := c.ReadJSON(&recv)
		if err != nil {
			// Check if it's an unexpected closure
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseAbnormalClosure,
				websocket.CloseGoingAway) {
				log.Printf("ERROR: Unexpected WebSocket closure: %v", err)
			} else {
				log.Printf("INFO: WebSocket closed normally or read error: %v", err)
			}
			break
		}
		
		// Log received GPS coordinates
		log.Printf("Kafka receiver: OBUID=%d, Lat=%.2f, Lon=%.2f", recv.OBUID, recv.Lat, recv.Lon)
		
		// Marshal GPS coordinates to JSON for Kafka
		resp, err := json.Marshal(recv)
		if err != nil {
			log.Printf("ERROR: Failed to marshal GPS data: %v", err)
			continue
		}
		
		log.Printf("Publishing to Kafka: %s", string(resp))
		
		// Send to Kafka via timing middleware
		MiddlewareRead(resp, k)
	}
	
	log.Println("Message loop terminated, connection closed")
}
