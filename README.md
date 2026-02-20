# GPS Coordinates Kafka Microservice

A distributed microservice system for real-time GPS coordinate streaming using Apache Kafka, WebSockets, and Go. This project demonstrates building a scalable event-streaming pipeline for vehicle tracking and fleet management.

## 📚 Overview

This microservice architecture streams GPS coordinates from simulated On-Board Units (OBUs) through a WebSocket server into Apache Kafka, where they can be consumed and processed in real-time.

```
GPS Producer → WebSocket → Data Receiver → Kafka Broker → Consumer
  (Simulated OBU)            (HTTP Server)    (Message Queue)  (Processor)
```

### Key Features

- **Real-time GPS Streaming**: Continuous flow of location data via WebSocket
- **Event-Driven Architecture**: Kafka for reliable, scalable message delivery
- **Observability**: Prometheus metrics for monitoring performance
- **Containerized Deployment**: Docker Compose for easy local development
- **Performance Instrumentation**: Built-in latency measurement middleware

## 🏗️ Architecture

### System Components

1. **Producer** (`producer/`)
   - Simulates GPS data from vehicle On-Board Units
   - Generates random coordinates every second
   - Sends data via WebSocket connection
   - Measures generation latency

2. **Data Receiver** (`data_receiver_kafka_producer/`)
   - HTTP/WebSocket server (port 30000)
   - Receives GPS coordinates from producers
   - Publishes messages to Kafka topic
   - Exposes Prometheus metrics endpoint

3. **Kafka Broker** (Docker container)
   - Bitnami Kafka 3.5 with KRaft mode (no Zookeeper)
   - Topic: `gpscoords`
   - Port: 9092

4. **Consumer** (`kafka_reader/`)
   - Reads GPS messages from Kafka
   - Deserializes JSON payloads
   - Processes coordinates (currently prints to stdout)

### Data Flow

```
┌─────────────┐     WebSocket (JSON)      ┌──────────────────┐
│   Producer  │ ───────────────────────> │  Data Receiver   │
│ (simulator) │  GPS Coordinates         │  (HTTP Server)   │
└─────────────┘                            └────────┬─────────┘
                                                     │
                                                     │ Kafka Produce
                                                     ▼
                                            ┌─────────────────┐
                                            │  Kafka Broker   │
                                            │  Topic: gpscoords│
                                            └────────┬─────────┘
                                                     │
                                                     │ Kafka Consume
                                                     ▼
                                            ┌─────────────────┐
                                            │    Consumer     │
                                            │   (Processor)   │
                                            └─────────────────┘
```

## 🗂️ Project Structure

```
kafka_microservice/
├── producer/                      # GPS data producer (simulator)
│   ├── main.go                    # WebSocket client sending GPS data
│   └── middleware.go              # Timing instrumentation
│
├── data_receiver_kafka_producer/  # WebSocket → Kafka bridge
│   ├── main.go                    # HTTP server with /ws and /metrics endpoints
│   ├── handlers/
│   │   ├── receive.go             # WebSocket upgrade and message handling
│   │   └── middlware_handler.go   # Kafka write timing middleware
│   └── kafka/
│       └── kafkaproducer.go       # Kafka producer wrapper
│
├── kafka_reader/                  # Kafka consumer
│   ├── main.go                    # Consumer entry point
│   └── kafka/
│       └── kafkaconsumer.go       # Kafka consumer wrapper with goroutine
│
├── types/                         # Shared data structures
│   └── types.go                   # SourceCoords struct
│
├── docker-compose.yml             # Kafka broker configuration
├── Makefile                       # Build and run commands
├── go.mod                         # Go module dependencies
└── README.md                      # This file
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.20+** - [Install Go](https://golang.org/doc/install)
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **librdkafka** (for Confluent Kafka Go client)

#### Install librdkafka

**macOS:**
```bash
brew install librdkafka
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get install -y librdkafka-dev
```

**Linux (Fedora/RHEL):**
```bash
sudo dnf install -y librdkafka-devel
```

### 1. Start Kafka Broker

```bash
docker-compose up -d
```

This starts a Kafka broker (KRaft mode) on `localhost:9092`.

**Verify Kafka is running:**
```bash
docker ps
# Should show gpscords_app-kafka-1 container running
```

### 2. Start Data Receiver (WebSocket → Kafka)

Terminal 1:
```bash
make receiver
# Or manually:
go run ./data_receiver_kafka_producer
```

The server starts on `localhost:30000` with endpoints:
- `ws://localhost:30000/ws` - WebSocket for GPS data
- `http://localhost:30000/metrics` - Prometheus metrics

### 3. Start Consumer (Kafka → Processor)

Terminal 2:
```bash
make calculator
# Or manually:
go run ./kafka_reader
```

Consumer subscribes to `gpscoords` topic and prints messages.

### 4. Start Producer (GPS Simulator)

Terminal 3:
```bash
make producer
# Or manually:
go run ./producer
```

Producer generates GPS coordinates every second and sends via WebSocket.

### Expected Output

**Producer (Terminal 3):**
```
Producer sending: OBUID=8674665223082153551, Lat=45.23, Lon=67.89
GPS data generation took: 123µs
```

**Data Receiver (Terminal 1):**
```
Kafka receiver: OBUID=8674665223082153551, Lat=45.23, Lon=67.89
Publishing to Kafka: {"obuid":8674665223082153551,"lat":45.23,"lon":67.89}
Writing to Kafka took: 2.5ms
✓ Successfully produced record to topic gpscoords partition [0] @ offset 42
```

**Consumer (Terminal 2):**
```
Kafka consumer received: OBUID=8674665223082153551, Lat=45.23, Lon=67.89
```

## 📊 Monitoring

### Prometheus Metrics

Access metrics at: `http://localhost:30000/metrics`

Key metrics available:
- HTTP request counts and durations
- Go runtime metrics (goroutines, memory, GC)
- Custom application metrics (can be extended)

### Kafka Metrics

Monitor Kafka performance:
```bash
# Check topic info
docker exec -it gpscords_app-kafka-1 kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --describe --topic gpscoords

# Check consumer group lag
docker exec -it gpscords_app-kafka-1 kafka-consumer-groups.sh \
  --bootstrap-server localhost:9092 \
  --describe --group gps
```

## 🔧 Configuration

### Kafka Configuration

**Broker** (in `docker-compose.yml`):
- Port: 9092
- Mode: KRaft (no Zookeeper required)
- Data directory: Persisted to `kafka_data` volume

**Producer** (`data_receiver_kafka_producer/kafka/kafkaproducer.go`):
```go
server = "gpscords_app-kafka-1:9092"  // Broker address
topic  = "gpscoords"                  // Topic name
```

**Consumer** (`kafka_reader/kafka/kafkaconsumer.go`):
```go
server       = "gpscords_app-kafka-1:9092"  // Broker address
topic        = "gpscoords"                  // Topic to consume
offset_reset = "earliest"                   // Start from beginning
group_id     = "gps"                        // Consumer group ID
```

### WebSocket Configuration

**Producer** (`producer/main.go`):
```go
wsEndpoint = "ws://localhost:30000/ws"
```

**Data Receiver** (`data_receiver_kafka_producer/main.go`):
```go
addr = "localhost:30000"  // Can override via -addr flag
```

### GPS Data Generation

**Producer** (`producer/main.go`):
```go
// Generates random GPS coordinates:
// - OBUID: Random int (for unique device ID)
// - Lat: 1.0 to 100.0
// - Lon: 1.0 to 100.0
// - Frequency: 1 message per second
```

## 🧪 Testing

### Manual Testing

**Test WebSocket endpoint directly:**
```bash
# Using websocat (install: brew install websocat)
websocat ws://localhost:30000/ws
# Then send JSON:
{"obuid": 123, "lat": 40.7128, "lon": -74.0060}
```

**Produce test message:**
```bash
docker exec -it gpscords_app-kafka-1 kafka-console-producer.sh \
  --bootstrap-server localhost:9092 \
  --topic gpscoords
# Enter JSON and press Enter:
{"obuid": 999, "lat": 51.5074, "lon": -0.1278}
```

**Consume messages:**
```bash
docker exec -it gpscords_app-kafka-1 kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic gpscoords \
  --from-beginning
```

### Performance Testing

**Test with multiple producers:**
```bash
# Terminal 1
go run ./producer

# Terminal 2
go run ./producer

# Terminal 3
go run ./producer
```

Monitor consumer lag and throughput.

## 📦 Dependencies

### Go Modules

```go
require (
    github.com/confluentinc/confluent-kafka-go/v2  // Kafka client
    github.com/gorilla/websocket                    // WebSocket support
    github.com/prometheus/client_golang             // Prometheus metrics
)
```

### External Services

- **Kafka**: Message broker (Bitnami Docker image)
- **librdkafka**: C library for Kafka (required by confluent-kafka-go)

## 🔍 Key Concepts

### Kafka Fundamentals

**Producer:**
- Asynchronous message delivery
- Automatic retries and batching
- Delivery reports via event channel

**Consumer:**
- Consumer group for load balancing
- Automatic partition assignment
- Offset management (earliest/latest)
- Poll-based message retrieval

**Topics:**
- Append-only log structure
- Partitioned for parallelism
- Configurable retention policy

### WebSocket vs HTTP

**Why WebSocket?**
- Bi-directional communication
- Low latency for real-time data
- Persistent connection (no overhead of repeated HTTP handshakes)
- Ideal for streaming GPS updates

### Middleware Pattern

**Timing Middleware:**
```go
func MiddlewareReceiver(h func() (int, float64, float64)) (int, float64, float64) {
    start := time.Now()
    defer func() {
        log.Printf("Took: %v", time.Since(start))
    }()
    return h()
}
```

Benefits:
- Separation of concerns (timing vs business logic)
- Reusable across functions
- Easy to add/remove instrumentation

## 🚀 Advanced Usage

### Scaling Consumers

Run multiple consumers in same group:
```bash
# Terminal 1
go run ./kafka_reader

# Terminal 2
go run ./kafka_reader

# Kafka automatically distributes partitions
```

### Custom Message Processing

Modify `kafkaconsumeLoop` in `kafka_reader/kafka/kafkaconsumer.go`:
```go
case *kafka.Message:
    var coord types.SourceCoords
    json.Unmarshal(e.Value, &coord)
    
    // Custom processing:
    storeInDatabase(coord)
    checkGeofence(coord)
    calculateSpeed(coord)
    sendAlert(coord)
```

### Production Considerations

**Security:**
- [ ] Add TLS for WebSocket connections
- [ ] Enable SASL authentication for Kafka
- [ ] Implement API key validation

**Reliability:**
- [ ] Add WebSocket reconnection logic
- [ ] Implement circuit breakers
- [ ] Add message schema validation
- [ ] Configure Kafka replication factor

**Performance:**
- [ ] Remove `Flush()` from every Kafka write (batch instead)
- [ ] Use buffered channels
- [ ] Add connection pooling
- [ ] Tune Kafka producer/consumer configs

**Observability:**
- [ ] Add distributed tracing (OpenTelemetry)
- [ ] Custom Prometheus metrics (message rate, errors)
- [ ] Structured logging (JSON format)
- [ ] Health check endpoints

## 🛠️ Troubleshooting

### Issue: Cannot connect to Kafka

**Symptoms:**
```
ERROR: Failed to create Kafka producer: Local: Broker transport failure
```

**Solutions:**
1. Verify Kafka container is running: `docker ps"`
2. Check Kafka logs: `docker logs gpscords_app-kafka-1`
3. Ensure correct broker address in config
4. Wait 30s for Kafka to fully initialize after `docker-compose up`

### Issue: librdkafka not found

**Symptoms:**
```
fatal error: 'librdkafka/rdkafka.h' file not found
```

**Solutions:**
1. Install librdkafka: `brew install librdkafka` (macOS)
2. Set CGO flags: `export CGO_CFLAGS="-I/usr/local/include"`
3. See Prerequisites section for platform-specific install

### Issue: WebSocket connection refused

**Symptoms:**
```
ERROR: Couldn't dial WebSocket server: connection refused
```

**Solutions:**
1. Start data receiver first: `make receiver`
2. Verify server is listening: `lsof -i :30000`
3. Check firewall settings

### Issue: Consumer not receiving messages

**Solutions:**
1. Check topic exists: `docker exec -it gpscords_app-kafka-1 kafka-topics.sh --bootstrap-server localhost:9092 --list`
2. Verify messages in topic: Use kafka-console-consumer (see Testing section)
3. Check consumer group offset: Consumer starts from "earliest" but only for first run
4. Reset consumer group: `docker exec -it gpscords_app-kafka-1 kafka-consumer-groups.sh --bootstrap-server localhost:9092 --group gps --reset-offsets --to-earliest --topic gpscoords --execute`

## 📝 TODO / Future Enhancements

- [ ] Add unit tests and integration tests
- [ ] Implement proper error handling and retry logic
- [ ] Add Graceful shutdown for all components
- [ ] Schema registry for message validation (Avro/Protobuf)
- [ ] Add authentication and authorization
- [ ] Implement geofencing and alerting
- [ ] Store coordinates in time-series database (InfluxDB, TimescaleDB)
- [ ] Add GraphQL API for querying historical data
- [ ] Create dashboard for real-time visualization
- [ ] Kubernetes deployment manifests
- [ ] CI/CD pipeline configuration

## 📖 References

- [Apache Kafka Documentation](https://kafka.apache.org/documentation/)
- [Confluent Kafka Go Client](https://docs.confluent.io/kafka-clients/go/current/overview.html)
- [Gorilla WebSocket](https://github.com/gorilla/websocket)
- [Prometheus Client Go](https://github.com/prometheus/client_golang)
- [Kafka KRaft Mode](https://kafka.apache.org/documentation/#kraft)

## 📄 License

This is an educational project demonstrating Kafka-based microservices architecture.

---

**Happy Streaming! 🚀📍**
