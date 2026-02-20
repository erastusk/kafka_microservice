// Package main provides middleware functions for the GPS producer.
package main

import (
	"log"
	"time"
)

// MiddlewareReceiver wraps a function call with timing instrumentation.
//
// This middleware:
//   - Records start time
//   - Executes the provided function
//   - Calculates and logs execution duration
//   - Returns the function's results unchanged
//
// Parameters:
//   - h: Function that returns (OBUID int, latitude float64, longitude float64)
//
// Returns:
//   - The three values returned by the wrapped function
//
// Use case: Measuring latency of GPS data generation operations.
func MiddlewareReceiver(h func() (int, float64, float64)) (int, float64, float64) {
	// Record start time for latency measurement
	start := time.Now()

	// Defer logging to ensure it runs after function completes
	defer func() {
		log.Printf("GPS data generation took: %v", time.Since(start))
	}()

	// Execute wrapped function and return its results
	return h()
}
