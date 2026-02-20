// Package types defines common data structures used across the GPS microservices.
// These types are shared between the producer, data receiver, and consumer components.
package types

// SourceCoords represents GPS coordinates from an On-Board Unit (OBU) device.
//
// This struct is:
//   - Serialized to JSON when sent over WebSocket
//   - Published to Kafka topic as JSON
//   - Deserialized by consumer for processing
//
// Fields:
//   - OBUID: Unique identifier for the vehicle's On-Board Unit
//   - Lat: Latitude coordinate (decimal degrees, -90 to +90)
//   - Lon: Longitude coordinate (decimal degrees, -180 to +180)
//
// JSON struct tags:
//   - Enable automatic marshaling/unmarshaling
//   - Use lowercase field names in JSON for consistency
//   - Required for compatibility between producer and consumer
//
// Example JSON:
//   {"obuid": 42, "lat": 40.7128, "lon": -74.0060}
//
// Use cases:
//   - Vehicle tracking and fleet management
//   - Route optimization and geofencing
//   - Real-time location monitoring
//   - Distance and speed calculations
type SourceCoords struct {
	OBUID int     `json:"obuid"` // On-Board Unit ID (vehicle identifier)
	Lat   float64 `json:"lat"`   // Latitude in decimal degrees
	Lon   float64 `json:"lon"`   // Longitude in decimal degrees
}
