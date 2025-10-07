# Toll Calculator - Event-Driven Microservices System

A production-grade demonstration of a microservices architecture implementing a toll calculation system using event-driven design patterns, Apache Kafka for message streaming, dual transport layers (HTTP and gRPC), and comprehensive monitoring with Prometheus and Grafana.

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Microservices](#microservices)
- [Technology Stack](#technology-stack)
- [Data Flow](#data-flow)
- [Transport Layers](#transport-layers)
- [Monitoring and Observability](#monitoring-and-observability)
- [Project Structure](#project-structure)
- [Setup, Configuration & Development](#setup-configuration--development)
- [API Documentation](#api-documentation)
- [Testing](#testing)

## Overview

The Toll Calculator system is a production-grade demonstration of modern microservices architecture that simulates a comprehensive highway toll collection system. In this system, On-Board Units (OBUs) represent physical devices installed in vehicles that continuously broadcast their geographical coordinates as they travel along highways. These location data points flow through a sophisticated distributed system that calculates distances traveled, aggregates journey metrics per vehicle, and ultimately generates accurate toll invoices based on the total distance covered.

This project serves as a practical reference implementation for developers looking to understand and implement event-driven microservices architectures in real-world scenarios. Rather than presenting isolated concepts, it demonstrates how multiple architectural patterns and technologies work together cohesively to solve complex distributed systems challenges. The system handles real-time data ingestion, asynchronous message processing, cross-service communication using multiple protocols, and comprehensive observability—all while maintaining loose coupling between services and ensuring fault tolerance.

### Key Architectural Patterns and Technologies Demonstrated

**Event-Driven Architecture with Apache Kafka**: The system leverages Apache Kafka as its central nervous system, enabling asynchronous, decoupled communication between services. When OBU data arrives, it's published to Kafka topics where downstream consumers can process it independently without blocking the data ingestion pipeline. This approach ensures high throughput, fault tolerance through message persistence, and the ability to replay events when needed. The publish-subscribe pattern allows multiple consumers to process the same events for different purposes without impacting each other.

**Microservices Design Patterns**: Each service is designed with clear boundaries and single responsibilities, following domain-driven design principles. The implementation showcases several critical patterns including the middleware/decorator pattern for cross-cutting concerns (logging, metrics), service interface abstraction for testability and flexibility, and the repository pattern for data access. These patterns ensure that each service can be developed, tested, deployed, and scaled independently while maintaining clean separation of concerns.

**Dual Transport Protocol Strategy**: The system intentionally implements both HTTP REST and gRPC transport layers to demonstrate when and why you might choose one over the other. HTTP REST endpoints provide human-readable, easily debuggable interfaces ideal for external clients and administrative operations. Meanwhile, gRPC is employed for high-frequency internal service-to-service communication where binary serialization, type safety, and performance are critical. This dual approach shows how different protocols can coexist within a single system, each serving its optimal use case.

**Real-Time Data Processing Pipeline**: The WebSocket implementation enables bidirectional, persistent connections between OBU simulators and the data receiver, allowing continuous streaming of location data with minimal latency. This real-time ingestion is crucial for time-sensitive applications where stale data could lead to inaccurate calculations. The system demonstrates how to manage WebSocket connection lifecycles, handle reconnections gracefully, and integrate streaming data with message queue architectures.

**Comprehensive Observability Stack**: Production systems require deep visibility into their behavior, and this project implements a complete observability solution. Structured logging with Logrus provides contextual, searchable log entries tagged with request IDs for distributed tracing. Prometheus metrics collection exposes detailed performance indicators including request rates, error rates, and latency distributions across percentiles. Grafana provides visualization and dashboarding capabilities, transforming raw metrics into actionable insights. Together, these tools enable operators to monitor system health, diagnose issues, and optimize performance.

**Containerized Infrastructure**: Docker Compose orchestrates the entire ecosystem, including Kafka brokers, Prometheus, and Grafana, allowing developers to spin up the complete infrastructure with a single command. This containerization approach ensures consistency across development, testing, and production environments while simplifying dependency management and deployment workflows.

**Resilience and Operational Excellence**: The system implements graceful shutdown mechanisms ensuring that services can terminate cleanly without losing in-flight requests or corrupting data. Context-based lifecycle management propagates cancellation signals through the service hierarchy, allowing coordinated cleanup. Request tracing with unique IDs enables end-to-end tracking of individual requests as they flow through multiple services, which is essential for debugging distributed systems. Error handling patterns demonstrate how to fail fast when appropriate while maintaining system stability.

## Architecture

The system implements a layered microservices architecture designed around event-driven communication patterns and clear separation of concerns. Each component operates independently with well-defined interfaces, allowing for independent scaling, deployment, and maintenance. The architecture prioritizes asynchronous communication where possible to maximize throughput and resilience, while using synchronous protocols for operations requiring immediate responses.

### Component Interaction Model

```
┌─────────────┐
│     OBU     │  (Simulates vehicle location data)
│  Simulator  │
└──────┬──────┘
       │ WebSocket (Persistent bidirectional connection)
       ▼
┌─────────────────┐
│  Data Receiver  │  (Receives OBU data via WebSocket)
└────────┬────────┘
         │ Produces to Kafka (Async, fire-and-forget)
         ▼
    ┌────────┐
    │ Kafka  │  (Message broker - obudata topic)
    └────┬───┘
         │ Consumes from Kafka (Pull-based, at-least-once delivery)
         ▼
┌───────────────────────┐
│ Distance Calculator   │  (Calculates distance from coordinates)
└──────────┬────────────┘
           │ gRPC (Binary protocol for high-throughput internal calls)
           ▼
    ┌──────────────┐
    │  Aggregator  │  (Aggregates distances, calculates invoices)
    └──────┬───────┘
           │ HTTP REST (Human-readable API for external access)
           ▼
    ┌─────────────┐
    │   Gateway   │  (API Gateway for external clients)
    └─────────────┘
           │
           ▼
    ┌──────────────────┐
    │ Prometheus       │  (Metrics collection via pull-based scraping)
    │ +                │
    │ Grafana          │  (Metrics visualization and alerting)
    └──────────────────┘
```

### Architectural Flow Explanation

**Data Ingestion Layer (OBU Simulator → Data Receiver)**: The OBU simulator establishes persistent WebSocket connections to the data receiver, mimicking real-world IoT devices that maintain long-lived connections for continuous data streaming. This connection model is chosen over HTTP polling because it eliminates the overhead of repeated connection establishment and allows the server to push acknowledgments or commands back to the devices if needed. The data receiver is designed to handle thousands of concurrent WebSocket connections, each streaming location updates at regular intervals.

**Message Broker Layer (Kafka)**: Apache Kafka serves as the system's event backbone, decoupling the data ingestion pipeline from downstream processing. When the data receiver publishes OBU events to Kafka, it doesn't need to wait for processing to complete or even know which services will consume the events. Kafka persists these events on disk, providing durability guarantees and enabling multiple consumers to process the same event stream independently. This architecture allows the distance calculator to process events at its own pace without backpressure affecting the data receiver's ability to accept new connections. If the distance calculator crashes or needs maintenance, events accumulate in Kafka and are processed once the service recovers, ensuring no data loss.

**Processing Layer (Distance Calculator)**: The distance calculator consumes events from Kafka using a consumer group, which enables horizontal scaling—multiple instances can share the processing load with Kafka automatically balancing partitions across consumers. After calculating the distance between coordinate pairs, the service needs to reliably transmit this data to the aggregator. Since this is internal service-to-service communication with high message volume, gRPC is selected for its efficiency, binary serialization, and built-in support for strongly-typed contracts defined in Protocol Buffers.

**Business Logic Layer (Aggregator)**: The aggregator is the system's core business logic service, maintaining state about accumulated distances per vehicle and exposing both gRPC and HTTP interfaces. The gRPC interface receives high-frequency distance updates from the calculator, while the HTTP interface provides human-readable REST APIs for querying invoice data. This service implements the repository pattern, abstracting storage behind an interface that currently uses in-memory storage but could be swapped for PostgreSQL, Redis, or another persistent store without changing business logic. All operations are instrumented with Prometheus metrics, providing visibility into aggregation rates, storage operations, and invoice calculations.

**API Gateway Layer**: The gateway acts as a reverse proxy and single entry point for external clients, abstracting the internal service topology. Clients interact with a stable gateway API without needing to know which internal service handles their request or how services are deployed. This layer is where you would typically implement cross-cutting concerns like authentication, rate limiting, request transformation, and API versioning in a production system.

**Observability Layer (Prometheus + Grafana)**: Prometheus scrapes metrics endpoints exposed by services at regular intervals, collecting time-series data about request rates, error rates, latencies, and custom business metrics. Unlike push-based monitoring where services send metrics to a collector, Prometheus's pull model gives the monitoring system control over scrape frequency and makes it easier to detect when services become unavailable. Grafana queries Prometheus to visualize these metrics in dashboards, enabling operators to identify performance bottlenecks, capacity constraints, and anomalous behavior patterns.

## Microservices

### 1. OBU Simulator (`obu`)

**Purpose and Context**: The OBU (On-Board Unit) Simulator serves as the data source for the entire system, mimicking physical GPS devices that would be installed in vehicles traveling on toll highways. In a production environment, these would be actual hardware devices transmitting real coordinates as vehicles move. This simulator generates realistic movement patterns by creating coordinate pairs that represent a vehicle's journey, with each data point containing both the vehicle's current position and its previous position, allowing downstream services to calculate the distance traveled between these two points.

**Implementation Location**: `obu/main.go`

**Core Functionality and Design Decisions**:

The simulator is designed to generate continuous streams of location data for multiple vehicles simultaneously, testing the system's ability to handle concurrent data streams from numerous sources. It implements the following workflow:

**Data Generation**: For each of the 20 simulated OBUs, the simulator generates pseudo-random latitude and longitude coordinates within realistic ranges. The coordinate generation algorithm creates small incremental changes between successive location updates, simulating actual vehicle movement along a highway rather than random teleportation across the map. Each OBU maintains state between iterations, storing its previous coordinates so that each transmission includes both where the vehicle was and where it is now—a critical requirement for the distance calculator to perform its calculations.

**WebSocket Communication**: Rather than using HTTP requests for each data point (which would be inefficient for high-frequency updates), the simulator establishes persistent WebSocket connections to the data receiver. WebSocket was chosen because it provides a full-duplex communication channel over a single TCP connection, eliminating the overhead of establishing new connections for each data transmission. This is particularly important when simulating real-world IoT devices that would maintain persistent connections to minimize battery drain and reduce network overhead. The simulator implements connection resilience, though in the current simplified version, connection management is handled through the gorilla/websocket library's built-in capabilities.

**Request Tracing**: Every OBU data packet is tagged with a UUID-based request ID using the Google UUID library. This request ID flows through the entire system—from the simulator through Kafka, to the distance calculator, and finally to the aggregator—enabling distributed tracing. When debugging issues or analyzing system behavior, operators can grep logs across all services for a specific request ID to see the complete journey of that particular data point through the microservices pipeline. This is a fundamental practice in distributed systems observability.

**Concurrent Operation**: The simulator spawns a goroutine for each OBU, allowing all 20 units to generate and transmit data concurrently. This concurrency model tests the data receiver's ability to handle multiple simultaneous WebSocket connections and ensures that slow processing of one OBU's data doesn't block others. Go's lightweight goroutines make this approach practical and efficient, as each goroutine consumes minimal resources compared to operating system threads.

**Configuration Parameters**:

- **WebSocket Endpoint**: `ws://127.0.0.1:30000/ws` - Connects to the local data receiver instance
- **Transmission Interval**: 1 second - Balances between realistic high-frequency updates and system load
- **Number of Simulated OBUs**: 20 - Configurable to test different load scenarios
- **Coordinate Ranges**: Latitude and longitude values within realistic bounds for highway travel

**Data Structure**:

```go
type OBUData struct {
    OBUID     int     // Unique identifier for this OBU (vehicle)
    CurrLat   float64 // Current latitude coordinate
    CurrLong  float64 // Current longitude coordinate
    PrevLat   float64 // Previous latitude (for distance calculation)
    PrevLong  float64 // Previous longitude (for distance calculation)
    RequestID string  // UUID for distributed tracing across services
}
```

This structure is serialized to JSON before transmission, making it human-readable for debugging while remaining compact enough for frequent transmission.

### 2. Data Receiver (`data_receiver`)

**Purpose and Context**: The Data Receiver serves as the system's ingestion gateway, accepting real-time location data from potentially thousands of OBU devices via WebSocket connections and reliably publishing this data to Kafka for asynchronous downstream processing. This service operates at the system's edge, handling the critical responsibility of ensuring no incoming data is lost while maintaining high throughput and connection stability. It acts as a boundary between the external world (IoT devices) and the internal event-driven architecture.

**Implementation Location**: `data_receiver/main.go`, `data_receiver/producer.go`

**Core Functionality and Architectural Patterns**:

The data receiver implements a sophisticated connection management system designed for reliability and observability:

**WebSocket Server Implementation**: The service runs an HTTP server that upgrades incoming connections to WebSocket protocol using the gorilla/websocket library. Unlike traditional request-response HTTP servers where each request is handled independently and quickly closed, this server maintains long-lived connections for each OBU device. The implementation includes CORS support, allowing browser-based clients to connect during development and testing. Each incoming WebSocket connection spawns dedicated goroutines to handle reading and writing independently, preventing slow consumers from blocking the entire service.

**Connection Lifecycle Management**: The service implements a hierarchical context structure that elegantly handles cancellation propagation throughout the system. At the root level, a signal.NotifyContext listens for OS signals (SIGINT/SIGTERM), which are sent when operators attempt to shut down the service. This root context cascades cancellation to the HTTP server context, which in turn cancels all active WebSocket connection contexts. This pattern ensures that when the service needs to shut down, all resources are cleaned up in an orderly fashion—existing messages in flight are completed, WebSocket connections are closed gracefully with proper close frames, and Kafka producers flush their buffers before terminating.

**Context Hierarchy**:

```
signal.NotifyContext (main) ← Triggered by OS signals
    └── ServerHTTP context ← HTTP server lifecycle
            └── handleWS context (per connection) ← Individual WebSocket connection
                    └── wsReceiveLoop context ← Message receiving loop
```

This hierarchy ensures that cancellation propagates correctly through all layers, preventing goroutine leaks and ensuring clean shutdowns.

**Kafka Producer Integration**: After receiving OBU data through WebSocket connections, the service must reliably publish this data to Kafka. The Kafka producer is configured with specific delivery guarantees—using the confluent-kafka-go library, it implements asynchronous production with delivery callbacks. When a message is successfully written to Kafka, the producer receives an acknowledgment event; if the write fails, error events are logged with sufficient context to diagnose issues. The producer is configured with a timeout (10 seconds) to prevent indefinite blocking if Kafka becomes unavailable. This design allows the service to continue accepting WebSocket data even if Kafka experiences brief outages, though extended Kafka unavailability would eventually cause backpressure.

**Data Validation and Error Handling**: When receiving JSON-encoded OBU data over WebSocket, the service validates that the data can be successfully unmarshaled into the expected OBUData structure. Malformed data is rejected with appropriate error responses logged, but the WebSocket connection remains open to accept subsequent valid messages. This forgiveness is important in IoT scenarios where occasional transmission errors shouldn't terminate the entire connection.

**Middleware Pattern for Cross-Cutting Concerns**: The service employs the decorator pattern to wrap the Kafka producer with logging middleware. Rather than scattering logging code throughout the business logic, the LogMiddleware decorator intercepts all produce operations, logs relevant details (including request IDs for tracing), measures latency, and then delegates to the actual producer. This pattern keeps the core producer logic clean and makes it trivial to add additional middleware for metrics collection, authentication checks, or rate limiting without modifying existing code.

**Configuration Parameters**:

- **HTTP Listen Address**: `:30000` - The port on which the WebSocket server listens for incoming connections
- **Kafka Broker**: `localhost:9092` - Connection string for the Kafka cluster
- **Kafka Topic**: `obudata` - The topic where all OBU location events are published
- **Max Kafka Timeout**: 10 seconds - Maximum time to wait for Kafka acknowledgments before timing out
- **Graceful Shutdown Timeout**: Configurable duration allowing in-flight messages to complete before forced termination

**Transport Layer**: HTTP/WebSocket - Chosen for its ability to maintain persistent bidirectional connections with minimal overhead

**Middleware Stack**:

- `LogMiddleware`: Intercepts all Kafka produce operations to log event details, timing information, and request IDs for distributed tracing

**Error Handling Strategy**: The service distinguishes between recoverable and fatal errors. WebSocket read errors on individual connections (client disconnects, network issues) are logged but don't crash the service—only that specific connection is closed. Kafka production errors are logged with full context but also don't terminate the service, allowing it to continue processing other messages. Only catastrophic errors like failure to bind to the listen port cause the service to exit.

### 3. Distance Calculator (`distance_calculator`)

**Purpose and Context**: The Distance Calculator is a dedicated processing service that consumes OBU location events from Kafka, performs computational transformation (calculating the distance traveled between two coordinate pairs), and forwards the results to the Aggregator service. This service exemplifies the single-responsibility principle—it does one thing well and is designed to be horizontally scalable to handle increasing message volumes. By separating distance calculation into its own service, the architecture allows this computationally intensive operation to scale independently from data ingestion and aggregation.

**Implementation Location**: `distance_calculator/main.go`, `distance_calculator/service.go`, `distance_calculator/consumer.go`

**Core Functionality and Design Patterns**:

**Kafka Consumer Implementation**: The service operates as a Kafka consumer group member, meaning multiple instances of this service can run concurrently with Kafka automatically load-balancing message processing across them. When the service starts, it subscribes to the `obudata` topic and begins polling for messages. The consumer is configured with specific parameters that balance performance with reliability:

- **Group ID** (`myGroup`): All instances with the same group ID form a consumer group, with Kafka ensuring each message is delivered to only one member
- **Auto-commit**: The consumer automatically commits offset positions after successfully processing messages, allowing it to resume from the last committed position after restarts
- **Session timeout and heartbeats**: These parameters help Kafka detect failed consumers and rebalance partitions to healthy instances
- **Poll interval**: Configured to prevent the consumer from being considered dead during long processing operations

**Distance Calculation Algorithm**: The service implements Euclidean distance calculation using the formula `√((x2-x1)² + (y2-y1)²)`. While this is a simplified distance calculation (true geographic distance would use the Haversine formula to account for Earth's curvature), it serves the demonstration purpose and performs efficiently. The calculation takes the current and previous latitude/longitude pairs from each OBU message and computes the straight-line distance between them. This value represents the distance the vehicle traveled since its last position update.

**Service Layer Architecture**: The calculator implements a clean service-oriented architecture with explicit interfaces, separating concerns into distinct layers:

```go
type CalculatorServicer interface {
    CalculateDistance(OBUData) (float64, error)
}
```

This interface abstraction provides several benefits. First, it enables easy unit testing through mock implementations—tests can verify behavior without requiring Kafka or gRPC infrastructure. Second, it allows runtime behavior modification through middleware wrapping. Third, it documents the contract that any calculator implementation must fulfill, making the codebase more maintainable and understandable.

**Middleware Pattern for Cross-Cutting Concerns**: Just like the data receiver, this service uses the decorator pattern to wrap the core calculator service with middleware for logging. The LogMiddleware intercepts each CalculateDistance call, logs the input data and request ID, performs the calculation by delegating to the wrapped service, logs the result and latency, then returns the calculated distance. This separation keeps business logic clean while providing comprehensive observability. Additional middleware could be added for metrics collection, circuit breaking, or result validation without modifying the core calculation logic.

**gRPC Client Communication**: After calculating distance, the service must reliably send this information to the Aggregator. It uses a gRPC client configured to connect to the Aggregator's gRPC endpoint. gRPC was chosen for this internal service-to-service communication because:

- **Performance**: Binary Protocol Buffer serialization is more efficient than JSON, reducing CPU overhead and network bandwidth
- **Type safety**: The .proto contract ensures both services agree on the message structure at compile time
- **Streaming support**: While not currently used, gRPC supports bidirectional streaming for future enhancements
- **Built-in features**: Deadline propagation, automatic retries, and connection multiplexing come standard

When sending data via gRPC, the service constructs an AggregateRequest containing the OBU ID, calculated distance value, timestamp, and the original request ID for tracing continuity. Errors during gRPC transmission are logged with full context but don't crash the service, allowing it to continue processing subsequent Kafka messages.

**Request Tracing Continuity**: The service preserves the request ID from the original OBU data packet through its entire processing pipeline. When consuming from Kafka, the request ID is extracted from the message. During distance calculation, it's logged alongside the result. When sending to the Aggregator via gRPC, it's included in the request. This end-to-end tracing enables operators to follow a single vehicle's location update from initial transmission through final invoice calculation.

**Configuration Parameters**:

- **Kafka Topic**: `obudata` - Source topic containing OBU location events
- **Kafka Brokers**: `localhost:9092` - Kafka cluster connection string
- **Kafka Group ID**: `myGroup` - Consumer group identifier enabling load distribution
- **Session Timeout**: 6000ms - Maximum time Kafka waits for heartbeats before considering consumer dead
- **Heartbeat Interval**: 2000ms - Frequency of heartbeat messages to Kafka coordinator
- **Max Poll Interval**: 300000ms - Maximum time between poll() calls before consumer is considered failed
- **Auto-commit Interval**: 1000ms - Frequency of automatic offset commits
- **Aggregator HTTP Endpoint**: `http://127.0.0.1:3000` - Backup HTTP communication channel
- **Aggregator gRPC Endpoint**: `127.0.0.1:3001` - Primary high-performance communication channel

**Service Interface Contract**:

```go
type CalculatorServicer interface {
    CalculateDistance(OBUData) (float64, error)
}
```

**Middleware Stack**:

- `LogMiddleware`: Wraps the calculator service to provide comprehensive logging of all distance calculation operations, including input parameters, calculated results, latency measurements, and request IDs for distributed tracing

**Error Handling and Resilience**: The service handles various error scenarios gracefully. Kafka consumption errors are logged but don't terminate the service unless they're persistent connection failures. Calculation errors (though unlikely with valid numeric inputs) are logged with full context. gRPC communication failures are logged and the service continues processing subsequent messages, though in a production system you'd want to implement retry logic with exponential backoff or dead letter queue patterns for failed aggregations.

### 4. Aggregator (`aggregator`)

**Purpose and Context**: The Aggregator represents the system's core business logic service, functioning as the stateful component responsible for maintaining cumulative distance records per vehicle and computing toll invoices based on configured pricing rules. Unlike the previous services which are primarily concerned with data transport and transformation, the Aggregator implements domain logic—understanding concepts like invoices, pricing, and customer accounts (represented by OBU IDs). This service demonstrates how to build a microservice that exposes multiple transport protocols, implements comprehensive observability, and uses pluggable storage abstractions for flexibility.

**Implementation Location**: `aggregator/main.go`, `aggregator/service.go`, `aggregator/store.go`, `aggregator/http.go`, `aggregator/grpc.go`, `aggregator/middleware.go`

**Core Functionality and Architectural Patterns**:

**Dual Transport Layer Architecture**: The Aggregator uniquely implements both HTTP and gRPC servers running concurrently on different ports, serving different use cases:

- **gRPC Interface (Port 3001)**: Optimized for high-frequency internal service-to-service communication, specifically receiving distance aggregation requests from the Distance Calculator. This interface uses Protocol Buffer serialization for efficiency and supports the `Aggregate(AggregateRequest)` RPC method. The binary protocol reduces serialization overhead and network bandwidth compared to JSON, making it ideal for the potentially millions of aggregation operations occurring as vehicles continuously transmit location data.

- **HTTP REST Interface (Port 3000)**: Provides human-readable, easily accessible APIs for querying invoice data and exposing metrics. This interface serves three primary endpoints:
  - `POST /aggregate`: Accepts distance aggregation via HTTP/JSON (alternative to gRPC for clients without protobuf support)
  - `GET /invoice?id={obuID}`: Retrieves computed invoice for a specific OBU, returning total distance and calculated amount
  - `GET /metrics`: Exposes Prometheus metrics in text format for scraping by monitoring systems

This dual-protocol approach demonstrates that different APIs can coexist within a single service, each optimized for its specific use case rather than forcing all interactions through a single protocol.

**Business Logic Implementation**: The Aggregator implements two core business operations:

**AggregateDistance Operation**: When distance data arrives (via either gRPC or HTTP), the service must atomically update the cumulative distance for the specified OBU. This operation retrieves the current total from storage, adds the new distance value, and stores the updated total. The implementation uses the repository pattern (Storer interface) to abstract storage operations, allowing the underlying storage mechanism to change without affecting business logic. Currently using in-memory storage for simplicity, the same interface could back a PostgreSQL database, Redis cache, or distributed key-value store by simply implementing the Storer interface for that technology.

**CalculateInvoice Operation**: Invoice calculation retrieves the total accumulated distance for an OBU and applies pricing logic (currently a simple multiplication: `distance × basePrice`). The base price of 3.7 per distance unit is configurable, allowing different pricing tiers or dynamic pricing strategies to be implemented. The invoice computation includes validation—if no data exists for the requested OBU ID, an appropriate error is returned rather than returning a zero invoice, ensuring clients can distinguish between "vehicle hasn't traveled" and "vehicle ID unknown."

**Service Layer Architecture with Interfaces**: The service is built around clean interface abstractions that enable testability, flexibility, and middleware composition:

```go
type Aggregator interface {
    AggregateDistance(context.Context, Distance) error
    CalculateInvoice(context.Context, int) (*Invoice, error)
}
```

This interface defines the contract that any aggregator implementation must satisfy, completely decoupling the business logic from transport concerns. The HTTP and gRPC handlers don't implement business logic—they simply unmarshal requests and delegate to the Aggregator interface, then marshal responses. This separation means the same business logic serves both protocols without duplication.

**Pluggable Storage Abstraction**: Storage operations are abstracted behind the Storer interface:

```go
type Storer interface {
    Insert(context.Context, Distance) error
    Get(context.Context, int) (float64, error)
}
```

The current in-memory implementation uses a sync.RWMutex-protected map for thread-safe concurrent access. This demonstrates proper concurrent programming in Go—readers can proceed in parallel using RLock(), while writes require exclusive access with Lock(). For production use, implementing this interface for PostgreSQL would involve SQL queries, for Redis would involve cache operations, each with appropriate error handling and connection management, but none of these changes would affect the business logic layer.

**Middleware Chain Pattern**: The service employs a sophisticated middleware chain that wraps the core business logic with cross-cutting concerns:

1. **MetricsMiddleware**: Outermost layer that instruments all operations with Prometheus metrics, tracking request counts, error counts, and latency distributions. This middleware intercepts both AggregateDistance and CalculateInvoice calls, measures execution time, increments counters, and records histograms before delegating to the wrapped service.

2. **LogMiddleware**: Inner layer providing structured logging with request IDs, parameters, results, and timing information. By positioning logging inside metrics collection, we ensure that logged timings match what metrics record.

This composable middleware approach adheres to the decorator pattern—each middleware wraps the service interface and adds behavior, then delegates to the next layer. Additional middleware for authentication, rate limiting, or caching could be inserted into this chain without modifying existing code.

**Comprehensive Observability**: The Aggregator implements production-grade observability:

**Prometheus Metrics**: The service exposes detailed metrics including:

- HTTP request counters per endpoint (aggregate, invoice)
- HTTP request latency histograms with configurable buckets (0.1s, 0.5s, 1s) enabling percentile calculations
- Service-level request and error counters
- Service-level latency distributions

These metrics enable operators to answer questions like "What's the 95th percentile latency for invoice calculations?" or "How many aggregation errors occurred in the last hour?"

**Structured Logging**: Using Logrus, every operation is logged with contextual fields including request IDs for distributed tracing, OBU IDs, distance values, computed amounts, and execution latencies. These structured logs are machine-parseable, enabling automated log analysis and correlation.

**Configuration Parameters**:

- **HTTP Port**: `:3000` - Configured via `AGG_HTTP_PORT` environment variable
- **gRPC Port**: `:3001` - Configured via `AGG_GRPC_PORT` environment variable
- **Store Type**: `memory` - Configured via `AGG_STORE_TYPE` (extensible to support multiple storage backends)
- **Base Toll Price**: `3.7` - Price per distance unit for invoice calculations
- **Metric Buckets**: `[0.1, 0.5, 1.0]` - Histogram buckets in seconds for latency measurement

**Service Interface Contract**:

```go
type Aggregator interface {
    AggregateDistance(context.Context, Distance) error
    CalculateInvoice(context.Context, int) (*Invoice, error)
}
```

**Storage Interface Contract**:

```go
type Storer interface {
    Insert(context.Context, Distance) error
    Get(context.Context, int) (float64, error)
}
```

**Middleware Composition**:

1. `MetricsMiddleware`: Outermost layer - Prometheus instrumentation for request counting, error tracking, and latency measurement
2. `LogMiddleware`: Inner layer - Structured logging with contextual information and request tracing

**API Endpoints**:

**HTTP REST**:

- `POST /aggregate`: Accept distance aggregation requests in JSON format (alternative to gRPC)
- `GET /invoice?id={obuID}`: Query invoice for specific OBU, returns JSON with total distance and computed amount
- `GET /metrics`: Prometheus metrics endpoint in text exposition format

**gRPC**:

- `Aggregate(AggregateRequest) returns (None)`: High-performance distance aggregation endpoint for internal services

**Concurrency and Thread Safety**: The in-memory storage implementation uses proper synchronization primitives (sync.RWMutex) to ensure thread-safe concurrent access. Multiple goroutines can safely read accumulated distances simultaneously, while write operations acquire exclusive locks. This demonstrates fundamental concurrent programming patterns essential for building reliable services in Go.

### 5. Gateway (`gateway`)

**Purpose and Context**: The Gateway service functions as the system's public-facing API layer, implementing the API Gateway pattern to provide a single, stable entry point for external clients while abstracting the complexity of the internal microservices architecture. This service acts as a reverse proxy, routing client requests to appropriate backend services, and serves as the ideal location for implementing cross-cutting concerns that apply to all external traffic such as authentication, rate limiting, request/response transformation, and API versioning. In production systems, the gateway becomes the system's public contract, allowing internal services to evolve, scale, and redeploy without impacting external clients.

**Implementation Location**: `gateway/main.go`, `gateway/handler/invoice.go`, `gateway/utils/http.go`, `gateway/config/constants.go`

**Core Functionality and Design Patterns**:

**API Gateway Pattern Implementation**: The gateway implements the API Gateway pattern, which provides several critical benefits for microservices architectures:

**Service Topology Abstraction**: External clients interact with a single gateway endpoint without needing knowledge of how many backend services exist, where they're deployed, or which service handles which operation. When a client requests an invoice via `GET /invoice?id={obuID}`, they don't need to know that this requires communication with the Aggregator service, or that the Aggregator runs on port 3000. If the Aggregator were split into multiple services or migrated to a different infrastructure, the gateway's public API could remain unchanged while only the internal routing logic is updated.

**Simplified Client Development**: Instead of requiring clients to understand multiple service endpoints, authentication schemes, and communication protocols, they interact with a single consistent HTTP REST API. This dramatically simplifies client development for web applications, mobile apps, and third-party integrations. The gateway translates external requests into whatever internal protocol is most efficient—currently HTTP to HTTP, but it could translate REST requests into gRPC calls to backend services if desired.

**Request Proxying and Transformation**: The current implementation demonstrates basic request proxying—when the gateway receives an invoice request, it constructs an HTTP request to the Aggregator service, forwards relevant parameters, awaits the response, and returns it to the client. The gateway handles error scenarios, translating internal service errors into appropriate HTTP status codes and client-friendly error messages. In more advanced implementations, the gateway could perform request transformation (modifying headers, enriching requests with additional context), response aggregation (combining data from multiple services into a single response), or protocol translation (REST to gRPC).

**Configuration and Service Discovery**: The gateway is configured via environment variables and command-line flags, demonstrating externalized configuration patterns essential for cloud-native applications:

- The aggregator endpoint is read from the `AGG_SERVICE_ENDPOINT` environment variable, allowing the same gateway binary to connect to different backends in development, staging, and production environments without code changes
- The listen port is configurable via the `-httpListenAddr` flag, enabling flexible deployment scenarios
- This approach follows the twelve-factor app methodology, making the service cloud-ready

In production systems, this configuration would typically integrate with service discovery mechanisms (Consul, Kubernetes DNS, etcd) to dynamically locate backend services rather than using static endpoints.

**Graceful Shutdown Implementation**: Like other HTTP services in the system, the gateway implements graceful shutdown to ensure reliability during deployments and maintenance:

When the gateway receives a termination signal (SIGINT/SIGTERM), it stops accepting new connections while allowing in-flight requests to complete within a configurable timeout period. This prevents client errors during deployments—existing requests finish successfully while new traffic is routed to other gateway instances (in a multi-instance deployment). This pattern is critical for achieving zero-downtime deployments in production environments.

**Handler-Based Routing Architecture**: The gateway uses a modular handler architecture where each route is implemented by a dedicated handler function. Currently, the invoice handler (`gateway/handler/invoice.go`) encapsulates all logic for retrieving invoices, including parameter validation, backend communication, error handling, and response formatting. This organization makes the codebase maintainable and testable—each handler can be unit tested independently, and new routes can be added without modifying existing handlers.

**HTTP Client Configuration**: The gateway maintains an HTTP client configured for reliable communication with backend services. Best practices for production would include:

- Connection pooling for efficiency (reusing TCP connections)
- Timeout configuration to prevent hanging requests
- Retry logic with exponential backoff for transient failures
- Circuit breaker patterns to prevent cascading failures when backends are unhealthy

**Configuration Parameters**:

- **Listen Port**: `:8000` - Default HTTP server port, configurable via `-httpListenAddr` command-line flag
- **Aggregator Service Endpoint**: Read from `AGG_SERVICE_ENDPOINT` environment variable (e.g., `http://localhost:3000`)
- **Graceful Shutdown Timeout**: Configurable duration for completing in-flight requests during shutdown

**HTTP API Endpoints**:

**GET /invoice?id={obuID}**: Retrieve toll invoice for a specific OBU

- **Request**: Query parameter `id` specifying the OBU identifier
- **Response**: JSON object containing OBU ID, total distance traveled, and calculated toll amount
- **Error Handling**: Returns appropriate HTTP status codes (400 for invalid requests, 500 for backend failures) with JSON error messages
- **Backend**: Proxies to Aggregator service `GET /invoice` endpoint

**Future Enhancement Opportunities**: The gateway architecture is designed for extension. In a production system, you would typically add:

- **Authentication and Authorization**: JWT validation, OAuth 2.0 integration, API key management
- **Rate Limiting**: Per-client request throttling to prevent abuse and ensure fair resource allocation
- **Request/Response Logging**: Comprehensive audit trails of all external API access
- **Metrics and Monitoring**: Request rates, latency, error rates per endpoint and per client
- **API Versioning**: Supporting multiple API versions concurrently (e.g., `/v1/invoice`, `/v2/invoice`)
- **Response Caching**: Cache frequently-accessed invoice data to reduce backend load
- **Request Validation**: Schema validation for all incoming requests
- **CORS Configuration**: Enabling browser-based clients from specific origins
- **TLS/HTTPS**: Encrypting all external traffic

## Technology Stack

### Core Technologies

- **Go 1.23**: Primary programming language
- **Apache Kafka**: Event streaming platform for asynchronous communication
- **Protocol Buffers (protobuf)**: Binary serialization for gRPC
- **gRPC**: High-performance RPC framework
- **WebSocket**: Real-time bidirectional communication

### Libraries and Frameworks

- **confluent-kafka-go v2.11.0**: Kafka client library
- **gorilla/websocket v1.5.3**: WebSocket implementation
- **google.golang.org/grpc v1.64.1**: gRPC framework
- **google.golang.org/protobuf v1.36.6**: Protocol Buffers
- **sirupsen/logrus v1.9.3**: Structured logging
- **prometheus/client_golang v1.23.0**: Prometheus metrics
- **joho/godotenv v1.5.1**: Environment variable management
- **google/uuid v1.6.0**: UUID generation

### Infrastructure

- **Docker**: Containerization
- **Docker Compose**: Multi-container orchestration
- **Prometheus**: Metrics collection and storage
- **Grafana**: Metrics visualization and dashboards

## Data Flow

This section provides a detailed walkthrough of how data moves through the system, from the initial generation of vehicle location data through final invoice calculation and retrieval. Understanding these flows is crucial for debugging issues, optimizing performance, and extending the system with new capabilities.

### 1. Data Generation and Transmission Flow

**Source**: OBU Simulator → **Destination**: Data Receiver → **Protocol**: WebSocket → **Frequency**: 1 second intervals

The data journey begins with the OBU Simulator, which mimics physical GPS devices installed in vehicles. Each of the 20 simulated OBUs operates independently in its own goroutine, executing the following sequence every second:

**Coordinate Generation**: The simulator generates new latitude and longitude coordinates for each OBU. Rather than producing completely random coordinates, the algorithm creates incremental changes from the previous position, simulating realistic vehicle movement along a continuous route. This ensures that calculated distances make sense and don't show vehicles teleporting across the map.

**Data Packet Assembly**: For each OBU, a data packet is constructed containing six critical pieces of information: the OBU's unique identifier (an integer from 0-19), the current latitude and longitude (representing where the vehicle is now), the previous latitude and longitude (where it was during the last transmission), and a freshly generated UUID serving as the request ID. This request ID becomes the thread that ties together all subsequent processing of this particular data point across multiple services.

**WebSocket Transmission**: The assembled data is serialized to JSON format and transmitted over the persistent WebSocket connection to the data receiver. Using WebSocket rather than making individual HTTP requests for each data point eliminates the overhead of TCP handshake, TLS negotiation, and HTTP header parsing that would occur with repeated HTTP calls. The connection remains open between transmissions, with the simulator simply writing JSON payloads to the socket at one-second intervals.

**Connection Management**: If the WebSocket connection fails (network issue, data receiver restart), the current implementation logs the error. In a production system, you would implement reconnection logic with exponential backoff to automatically re-establish the connection without losing data.

### 2. Data Ingestion and Event Publishing Flow

**Source**: Data Receiver → **Destination**: Kafka → **Protocol**: Kafka Binary Protocol → **Delivery Guarantee**: At-least-once

When OBU data arrives at the Data Receiver via WebSocket, it enters the ingestion pipeline:

**WebSocket Message Reception**: The data receiver's WebSocket handler is continuously reading from each active connection. When a message arrives, it's received as a byte slice containing JSON-encoded data. The handler's goroutine blocks on read operations until data arrives, making this an efficient event-driven model that doesn't waste CPU on polling.

**JSON Deserialization and Validation**: The received JSON is unmarshaled into an OBUData struct. This deserialization process validates the JSON structure—ensuring all required fields are present and properly typed. If deserialization fails (malformed JSON, missing fields, wrong types), an error is logged but the WebSocket connection remains open to accept subsequent valid messages. This resilience prevents a single malformed packet from severing the entire data stream from that vehicle.

**Kafka Message Production**: Once validated, the OBUData is serialized (typically back to JSON or another format) and published to the Kafka topic `obudata`. The data receiver's Kafka producer is configured for asynchronous production with delivery callbacks:

- The producer writes the message to its internal buffer and immediately returns control to the application
- Kafka's producer library batches multiple messages together for efficient transmission
- Once Kafka brokers acknowledge receipt, a delivery callback is triggered
- If production fails (Kafka unavailable, network partition), error callbacks provide notification

This asynchronous model allows the data receiver to continue accepting WebSocket data even during brief Kafka outages, though sustained Kafka unavailability would eventually fill producer buffers and cause backpressure.

**Logging and Observability**: The middleware layer wraps the Kafka producer, logging each production event with the request ID, OBU ID, and timing information. This creates an audit trail showing exactly when each piece of data entered the Kafka event stream.

### 3. Event Consumption and Distance Calculation Flow

**Source**: Kafka → **Service**: Distance Calculator → **Destination**: Aggregator → **Protocol**: gRPC

The Distance Calculator operates as an event-driven consumer, continuously polling Kafka for new messages:

**Kafka Consumer Poll Loop**: The service runs a polling loop that repeatedly calls the Kafka consumer's Poll() method. This method blocks until messages are available or a timeout occurs. When messages arrive, they're returned in batches for efficient processing. The consumer group mechanism ensures that if multiple distance calculator instances are running, Kafka automatically distributes partitions across them for parallel processing.

**Message Deserialization**: Each consumed Kafka message contains the serialized OBUData. The calculator deserializes this data back into structured form, extracting the current coordinates, previous coordinates, OBU ID, and crucially, the request ID for tracing continuity.

**Distance Calculation**: With both coordinate pairs extracted, the calculator applies the Euclidean distance formula: `√((currLat - prevLat)² + (currLong - prevLong)²)`. This computation produces a float64 value representing the straight-line distance between the two points. While simplified compared to true geographic distance calculations (Haversine formula), this demonstrates the transformation logic pattern—consuming events, applying computational transformation, producing derived events.

**Middleware Interception**: Before and after the calculation, middleware intercepts the operation for logging. The LogMiddleware records the input coordinates, calculated distance, request ID, and computation latency. This provides visibility into calculation accuracy and performance.

**gRPC Transmission to Aggregator**: The calculated distance must now be forwarded to the Aggregator for accumulation. The calculator constructs a gRPC AggregateRequest message using the Protocol Buffers format, containing the OBU ID, distance value, current Unix timestamp, and request ID. This message is sent via a gRPC client to the Aggregator's gRPC endpoint (port 3001). gRPC's binary serialization makes this transmission efficient despite occurring for every distance calculation—potentially thousands per second across all vehicles.

**Offset Commit**: After successfully processing a batch of messages and sending their results to the Aggregator, the Kafka consumer commits the offset positions back to Kafka. This commit tells Kafka "I've successfully processed messages up to offset X." If the distance calculator crashes and restarts, it resumes from the last committed offset, ensuring at-least-once processing (messages may be reprocessed but none are skipped).

### 4. Distance Aggregation and State Management Flow

**Service**: Aggregator → **Storage**: In-Memory Store → **Side Effect**: Prometheus Metrics

When the Aggregator receives distance data via its gRPC endpoint, it updates cumulative state:

**gRPC Request Reception**: The Aggregator's gRPC server receives the AggregateRequest over the HTTP/2 connection from the Distance Calculator. The gRPC library automatically deserializes the Protocol Buffer message into the AggregateRequest struct, making the OBU ID, distance value, timestamp, and request ID available to the handler.

**Middleware Chain Execution**: Before business logic executes, the request passes through the middleware chain. The MetricsMiddleware starts a timer and increments request counters. The LogMiddleware records the incoming request details with the request ID for tracing.

**Business Logic - Distance Accumulation**: The core Aggregator service receives the Distance data and invokes the storage layer's Insert operation. The storage interface is currently backed by an in-memory map, but the abstraction allows for swapping in database implementations. The Insert operation:

1. Acquires a write lock on the storage (ensuring thread safety)
2. Retrieves the current accumulated distance for this OBU ID (or 0 if this is the first record)
3. Adds the new distance to the accumulated total
4. Stores the updated total back in the map
5. Releases the lock

This read-modify-write operation is atomic from the perspective of concurrent goroutines—multiple distance updates for different OBUs can proceed in parallel, while updates for the same OBU are serialized to prevent data races.

**Metrics Recording**: As the middleware unwinds after successful processing, it records latency in Prometheus histograms and increments success counters. These metrics are accumulated in memory and exposed via the `/metrics` endpoint for Prometheus to scrape.

**Response**: The gRPC handler returns an empty response (the None message type) indicating successful processing. From the Distance Calculator's perspective, the aggregation is complete and it can proceed to process the next Kafka message.

### 5. Invoice Calculation and Retrieval Flow

**Client** → **Gateway** → **Aggregator** → **Protocol**: HTTP REST

When a client needs to retrieve an invoice (to bill a customer for their highway usage), the following query flow executes:

**Client Request Initiation**: An external client (web application, mobile app, billing system) sends an HTTP GET request to the Gateway at `http://gateway:8000/invoice?id={obuID}`, where `{obuID}` is the unique identifier for the vehicle whose invoice is being requested.

**Gateway Request Proxying**: The Gateway receives this request and performs several operations:

1. **Parameter Extraction**: Extracts the `id` query parameter from the URL
2. **Validation**: Validates that the parameter exists and is properly formatted
3. **Backend Request Construction**: Constructs a new HTTP request to the Aggregator service at `http://aggregator:3000/invoice?id={obuID}`
4. **Request Forwarding**: Issues the HTTP request to the Aggregator, awaiting the response
5. **Response Proxying**: Receives the Aggregator's response and forwards it back to the client, translating any error status codes as needed

This proxying layer means clients never directly communicate with the Aggregator—they only know about the Gateway endpoint, allowing backend services to be relocated, scaled, or reimplemented without impacting clients.

**Aggregator Invoice Calculation**: When the Aggregator receives the invoice request at its HTTP endpoint:

1. **Request Handling**: The HTTP handler extracts the OBU ID from query parameters and validates its format
2. **Storage Retrieval**: The service calls the storage layer's Get() operation to retrieve the accumulated distance for this OBU. The storage acquires a read lock (allowing concurrent invoice queries), looks up the distance in the map, and returns it
3. **Business Logic - Price Calculation**: The total distance is multiplied by the configured base price (3.7 per distance unit) to compute the total toll amount
4. **Invoice Assembly**: An Invoice struct is populated with the OBU ID, total distance, and calculated amount
5. **Error Handling**: If no data exists for the requested OBU ID, an error is returned with an appropriate message and HTTP 500 status code, distinguishing "no data" from "vehicle hasn't traveled" (which would show distance of 0)
6. **Response Serialization**: The Invoice is serialized to JSON and returned with HTTP 200 status

**Client Receipt**: The client receives the JSON invoice containing structured data about the vehicle's total distance and amount owed, which can be displayed in a user interface or processed by a billing system.

## Transport Layers

The system intentionally implements multiple transport protocols, each chosen for specific use cases based on their characteristics and trade-offs. This polyglot approach demonstrates that modern microservices architectures don't need to standardize on a single protocol—instead, you should select the protocol that best fits each communication pattern.

### HTTP/REST and WebSocket

**Services Using HTTP**: Data Receiver (WebSocket variant), Aggregator (REST API), Gateway (REST API)

HTTP serves as the system's primary human-facing protocol, chosen specifically for scenarios where human readability, broad compatibility, and ease of debugging outweigh raw performance concerns.

**Use Cases and Design Rationale**:

**WebSocket for Real-Time Streaming (Data Receiver)**: The Data Receiver uses WebSocket, an upgraded HTTP connection that maintains a persistent bidirectional channel between client and server. This protocol choice is crucial for the OBU simulator use case:

- **Efficiency**: Establishing a TCP connection, performing TLS handshake, and exchanging HTTP headers is expensive. For devices sending location updates every second, repeating this overhead for each data point would waste bandwidth and battery power. WebSocket establishes the connection once, then streams data frames with minimal per-message overhead.
- **Push Capability**: While not used in the current implementation, WebSocket's bidirectional nature allows the server to push commands or acknowledgments back to OBU devices without the devices polling.
- **Connection State**: The persistent connection provides natural session management—the server knows when a device connects and disconnects, enabling connection monitoring and alerting.
- **Compatibility**: WebSocket starts as an HTTP request with an Upgrade header, making it compatible with existing HTTP infrastructure (load balancers, proxies, firewalls) while providing the benefits of a persistent connection.

**RESTful HTTP for Query and External APIs (Aggregator, Gateway)**: The Aggregator and Gateway expose HTTP REST APIs for querying invoice data and (in the Aggregator's case) accepting distance aggregation requests:

- **Human-Readable**: JSON-over-HTTP is easily inspectable with tools like curl, Postman, or browser developer tools. When debugging why an invoice calculation is incorrect, you can manually query the API and examine the response without special tools.
- **Tooling Ecosystem**: HTTP enjoys universal tooling support—every programming language has HTTP client libraries, every monitoring system can perform HTTP health checks, every load balancer understands HTTP routing.
- **Caching**: HTTP's built-in caching semantics (Cache-Control headers, ETags) enable optimization for frequently-accessed invoice data without application changes. A caching proxy or CDN could be placed in front of the Gateway.
- **Stateless**: REST's stateless nature simplifies horizontal scaling—any Gateway or Aggregator instance can handle any request without session affinity concerns.
- **Documentation**: OpenAPI/Swagger specifications can document REST APIs for external consumers, generating interactive documentation and client SDKs.

**Performance Characteristics**: HTTP/REST's text-based JSON serialization is less efficient than binary protocols. A typical JSON invoice response might be 100-150 bytes, while a binary equivalent could be 30-40 bytes. For low-frequency operations like invoice queries, this overhead is negligible. However, for high-frequency operations (like distance aggregation), the cumulative cost becomes significant—which is why the system uses gRPC for that communication path.

**Active Endpoints**:

- **Data Receiver**: `ws://localhost:30000/ws` - WebSocket endpoint for OBU data streaming
- **Aggregator HTTP**: `http://localhost:3000/aggregate` - REST endpoint for distance aggregation (alternative to gRPC)
- **Aggregator HTTP**: `http://localhost:3000/invoice` - REST endpoint for invoice queries
- **Aggregator HTTP**: `http://localhost:3000/metrics` - Prometheus metrics in text format
- **Gateway**: `http://localhost:8000/invoice` - Public-facing invoice query API

### gRPC with Protocol Buffers

**Services Using gRPC**: Distance Calculator → Aggregator communication

gRPC represents the system's high-performance internal communication protocol, specifically chosen for the Distance Calculator to Aggregator communication path where efficiency and type safety are paramount.

**Use Cases and Design Rationale**:

**High-Frequency Internal Service Communication**: The Distance Calculator sends aggregation requests to the Aggregator for every distance calculation performed—potentially thousands of messages per second as 20 vehicles transmit location updates each second. This high-frequency communication makes gRPC's efficiency critical:

**Binary Serialization (Protocol Buffers)**: Unlike JSON which requires parsing text and allocating strings, Protocol Buffers serializes data to a compact binary format. An AggregateRequest containing an OBU ID, distance value, timestamp, and request ID serializes to approximately 25-30 bytes in protobuf compared to 120-150 bytes in JSON. This 4-5x reduction in payload size translates directly to reduced network bandwidth, lower serialization CPU overhead, and decreased latency.

**HTTP/2 Foundation**: gRPC runs over HTTP/2, which provides several benefits:

- **Multiplexing**: Multiple RPC calls share a single TCP connection without head-of-line blocking, reducing connection overhead
- **Header Compression**: HPACK compression reduces repetitive header transmission
- **Binary Framing**: More efficient than HTTP/1.1's text-based protocol
- **Flow Control**: Built-in backpressure mechanisms prevent overwhelming receivers

**Type Safety and Contract-First Development**: The Protocol Buffers schema in `types/ptypes.proto` serves as an explicit contract between services:

```protobuf
service Aggregator {
    rpc Aggregate(AggregateRequest) returns (None);
}

message AggregateRequest {
    int64 ObuID = 1;
    double Value = 2;
    int64 Unix = 3;
    string RequestID = 4;
}
```

This schema is language-agnostic and machine-readable. The protoc compiler generates strongly-typed Go code (`types/ptypes.pb.go` for message serialization, `types/ptypes_grpc.pb.go` for service stubs) ensuring that:

- Field name typos cause compile-time errors rather than runtime failures
- Type mismatches (sending a string where an int64 is expected) are caught during compilation
- Both client and server agree on the exact message structure and field semantics
- Schema evolution is managed with versioning and compatibility rules

**Performance Benchmarking**: In typical benchmarks, gRPC with protobuf shows 5-10x better throughput than JSON-over-HTTP for small messages and 2-3x lower latency, primarily due to reduced serialization overhead and HTTP/2's efficiency.

**Code Generation and Development Experience**: The protobuf compiler generates idiomatic Go code including:

- Struct definitions with proper Go types (int64, float64, string)
- Marshal/Unmarshal methods for serialization
- Client stubs that look like normal Go function calls
- Server interfaces that services implement

This generated code eliminates boilerplate and ensures consistency between client and server.

**Future Capabilities**: While currently using simple unary RPC (request-response), gRPC supports:

- **Server Streaming**: Aggregator could stream accumulated distance updates to monitoring dashboards
- **Client Streaming**: Distance Calculator could batch multiple calculations into a single stream
- **Bidirectional Streaming**: Real-time aggregation with continuous feedback
- **Deadline Propagation**: Request timeouts automatically flow through service calls
- **Automatic Retries**: Configurable retry policies for transient failures

**Trade-offs and When Not to Use gRPC**: Despite its performance advantages, gRPC has limitations that explain why it's not used everywhere:

- **Browser Support**: Browsers don't natively support HTTP/2 client-side features required for gRPC, making it unsuitable for direct browser-to-server communication (though gRPC-Web exists as a proxy-based solution)
- **Human Readability**: Binary protobuf payloads can't be easily inspected with standard tools—you need protobuf-aware decoders
- **Tooling**: While improving, gRPC tooling (debugging, testing, monitoring) is less mature than HTTP/REST's decades of development
- **Firewall/Proxy Compatibility**: Some legacy network infrastructure doesn't properly handle HTTP/2

**Proto Definition Location**: `types/ptypes.proto`

**Generated Code Artifacts**:

- `types/ptypes.pb.go` - Protocol Buffer message serialization code
- `types/ptypes_grpc.pb.go` - gRPC client stubs and server interfaces

### Protocol Selection Guidelines

The system demonstrates a pragmatic approach to protocol selection:

**Choose HTTP/REST when**:

- External clients need to access the API (browsers, mobile apps, third-party integrations)
- Human readability and debuggability are priorities
- Operations are low-frequency (queries, administrative operations)
- Caching would benefit the use case
- Universal tooling compatibility is required

**Choose gRPC when**:

- Internal service-to-service communication with high message volume
- Performance and efficiency are critical (low latency, high throughput requirements)
- Strong typing and contract enforcement are important
- Future need for streaming operations is anticipated
- Both ends of the communication are under your control (not external clients)

**Choose WebSocket when**:

- Bidirectional real-time communication is needed
- Persistent connections with minimal per-message overhead
- Server needs to push data to clients without polling
- Compatible with HTTP infrastructure while providing streaming benefits

## Monitoring and Observability

Production microservices require comprehensive observability to understand system behavior, diagnose issues, and optimize performance. This system implements a complete observability stack combining structured logging for detailed event tracking, Prometheus metrics for quantitative performance measurement, and Grafana for visualization and alerting. Together, these tools enable operators to answer critical questions like "Is the system healthy?", "Where are the bottlenecks?", and "What caused this incident?"

### Structured Logging with Logrus

**Framework and Philosophy**: The system uses Logrus, a popular structured logging library for Go that outputs logs in machine-parseable formats (JSON) rather than unstructured text. Structured logging is crucial for microservices because it enables automated log analysis, correlation across services, and efficient searching in log aggregation systems like ELK stack or Splunk.

**Log Levels and Their Usage**:

The system employs two primary log levels:

- **Info**: Normal operational events that provide visibility into system behavior (distance calculations, invoice generations, successful aggregations). These logs create an audit trail of system operations.
- **Error**: Exceptional conditions that indicate problems requiring attention (Kafka connection failures, gRPC communication errors, invalid requests). Error logs include stack traces and contextual information for debugging.

**Contextual Log Fields**: Each log entry includes structured fields providing context:

**Request ID (Distributed Tracing)**: Every log entry related to processing a specific OBU data packet includes the original request ID (UUID). This enables distributed tracing—when debugging why a particular vehicle's invoice is incorrect, you can grep all service logs for that request ID and see the complete journey of that data point through the system. For example, you might discover that the distance calculator received coordinates but the Aggregator never received the calculated distance, pinpointing a gRPC communication failure.

**Latency Measurements**: Operations log their execution time using high-precision timers (`time.Since(start)`). This creates a historical record of performance that can reveal degradation trends. If invoice calculations that normally take 200μs suddenly take 50ms, the logs provide evidence of a performance regression.

**Domain-Specific Fields**: Logs include relevant business data—OBU IDs, calculated distances, invoice amounts, coordinate values. This contextual information means you don't need to cross-reference multiple data sources to understand what a log entry is describing.

**Example Log Entry Anatomy**:

```json
{
  "level": "info",
  "msg": "Aggregate distance",
  "took": "500μs",
  "distance": { "OBUID": 12345, "Value": 10.5 },
  "requestId": "550e8400-e29b-41d4-a716-446655440000",
  "timestamp": "2025-01-15T10:23:45.123Z"
}
```

This entry tells you: what operation occurred (distance aggregation), how long it took (500 microseconds), which vehicle it concerned (OBU 12345), how far they traveled (10.5 distance units), and the request ID for correlation with other logs.

**Middleware Integration**: Logging is implemented as middleware that wraps service operations, ensuring consistent log formatting and automatic inclusion of timing information without scattering logging logic throughout business code. This separation of concerns keeps business logic clean and testable.

### Prometheus Metrics Collection

**Architecture and Collection Model**: Prometheus operates on a pull-based model where it periodically scrapes HTTP endpoints exposed by services to collect metrics. This differs from push-based systems where services send metrics to a collector. The pull model provides several advantages: the monitoring system controls scrape frequency, dead services are easily detected (scrape failures), and services don't need to know where to send metrics.

**Metrics Endpoint**: The Aggregator service exposes metrics at `http://localhost:3000/metrics` in Prometheus text exposition format. When Prometheus scrapes this endpoint, it receives a text document containing all metric values with their current readings.

**Metric Types and Use Cases**:

The system implements three primary Prometheus metric types:

**Counters (Monotonically Increasing Values)**: Counters only increase over time and reset when the service restarts. They're ideal for counting events:

1. **http_request_counter_aggregate**: Total number of HTTP POST requests to the `/aggregate` endpoint. By querying the rate of change (`rate(http_request_counter_aggregate[1m])`), you determine requests per second.

2. **http_request_counter_invoice**: Total invoice query requests via HTTP. Tracking this helps understand system load patterns and identify usage spikes.

3. **aggregator_request_counter**: Total business-level aggregation operations (combining both HTTP and gRPC aggregations). This metric answers "How much distance data are we processing?"

4. **calculator_request_counter**: Total distance calculations performed. Comparing this to the aggregator counter reveals if calculations are being lost between services.

5. **aggregator_error_counter** / **calculator_error_counter**: Total errors encountered. A sudden increase indicates a problem requiring investigation.

**Histograms (Distribution of Values)**: Histograms measure distributions by categorizing observations into configurable buckets. They're essential for understanding latency characteristics:

1. **http_request_latency_aggregate**: Measures how long HTTP aggregation requests take, with observations bucketed into 0.1s, 0.5s, and 1s ranges. This enables calculating percentiles—you can query "What's the 95th percentile latency?" to understand typical worst-case performance.

2. **http_request_latency_invoice**: Invoice query latency distribution. If most queries complete in <100ms but some take >1s, the histogram reveals this bimodal distribution that averages would hide.

3. **aggregator_request_latency** / **calculator_request_latency**: Service-level operation latencies independent of transport protocol. These measure the pure business logic performance.

**Histogram Buckets Explained**: The configured buckets [0.1s, 0.5s, 1s] create four ranges: ≤0.1s, 0.1-0.5s, 0.5-1.0s, and >1.0s. Prometheus counts how many observations fall into each bucket, enabling percentile calculations. For example, if 95% of requests fall in the ≤0.1s bucket, you know the 95th percentile latency is under 100ms.

**Prometheus Configuration**: The `prometheus.yml` file configures scraping behavior:

```yaml
scrape_configs:
  - job_name: "aggregator"
    scrape_interval: 15s
    scrape_timeout: 10s
    metrics_path: /metrics
    static_configs:
      - targets: ["host.docker.internal:3000"]
```

This configuration instructs Prometheus to:

- Scrape the Aggregator service every 15 seconds (balancing freshness with overhead)
- Timeout scrapes after 10 seconds to detect hung services
- Use the `/metrics` endpoint path
- Connect to `host.docker.internal:3000` (special Docker DNS name allowing containers to reach the host machine)

**Useful Prometheus Queries**:

Understanding how to query Prometheus transforms raw metrics into actionable insights:

**Request Rate**: `rate(aggregator_request_counter[1m])` calculates the per-second rate of aggregation requests over the last minute, smoothing out momentary spikes.

**Error Rate**: `rate(aggregator_error_counter[1m])` shows errors per second. Alerting on this metric enables proactive issue detection.

**Success Rate**: `rate(aggregator_request_counter[1m]) - rate(aggregator_error_counter[1m])` computes successful requests per second.

**Error Percentage**: `rate(aggregator_error_counter[1m]) / rate(aggregator_request_counter[1m]) * 100` calculates what percentage of requests fail.

**Average Latency**: `rate(aggregator_request_latency_sum[1m]) / rate(aggregator_request_latency_count[1m])` computes the average latency over the last minute.

**95th Percentile Latency**: `histogram_quantile(0.95, rate(aggregator_request_latency_bucket[1m]))` calculates the latency below which 95% of requests fall—a critical SLA metric.

**Throughput Comparison**: `rate(calculator_request_counter[1m]) - rate(aggregator_request_counter[1m])` reveals if the calculator is processing more distance calculations than the aggregator is receiving, indicating data loss.

### Grafana Visualization and Dashboarding

**Purpose and Integration**: Grafana provides a graphical interface for visualizing Prometheus metrics, creating dashboards that display system health at a glance, and configuring alerts that notify operators of anomalous conditions. While Prometheus excels at storing and querying time-series data, Grafana makes that data accessible and actionable through intuitive visualizations.

**Access and Authentication**: Grafana runs on `http://localhost:4000` with default credentials (username: `admin`, password: `admin`). In production, you would change these credentials, enable HTTPS, and integrate with corporate authentication systems (LDAP, OAuth, SAML).

**Pre-Configured Data Source**: The system includes provisioned configuration (`grafana/provisioning/datasources/prometheus.yml`) that automatically configures Grafana to query Prometheus at `http://prometheus:9090`. This provisioning eliminates manual setup—Grafana starts with the data source already configured and ready to use.

**Data Source Configuration Details**:

```yaml
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
```

This declares Prometheus as the default data source, reachable at the Docker Compose service name `prometheus` on port 9090, with Grafana proxying requests (server-side) rather than browsers querying Prometheus directly (client-side).

**Dashboard Creation Workflow**:

1. **Navigate to Dashboards → New Dashboard**
2. **Add Visualization Panel**: Each panel displays a specific metric or set of related metrics
3. **Configure Query**: Write PromQL queries to fetch desired metrics
4. **Select Visualization Type**: Line graphs for time series, gauges for current values, heat maps for latency distributions, bar charts for comparisons
5. **Apply Display Options**: Titles, units, legends, thresholds, colors
6. **Arrange Panels**: Organize multiple panels into a comprehensive dashboard layout

**Recommended Dashboard Panels**:

**Request Rate Overview Panel**: Line graph showing `rate(aggregator_request_counter[1m])` and `rate(calculator_request_counter[1m])` on the same chart. This visualization immediately reveals throughput and whether both services are processing similar volumes (if not, investigation is warranted).

**Error Rate Monitoring Panel**: Line graph of `rate(aggregator_error_counter[1m])` and `rate(calculator_error_counter[1m])`. A flat line near zero indicates healthy operation; spikes indicate problems.

**Latency Percentiles Panel**: Multiple lines showing `histogram_quantile(0.50, rate(...))`, `histogram_quantile(0.95, rate(...))`, and `histogram_quantile(0.99, rate(...))` for the same metric. This shows how latency varies from typical (p50) to worst-case (p99), revealing whether slow requests are outliers or common.

**Latency Heatmap Panel**: Heatmap visualization of `http_request_latency_aggregate_bucket` showing latency distribution over time. Color intensity indicates frequency, making patterns like "latency spikes every hour" or "bimodal distribution" visually obvious.

**Invoice Operations Panel**: Counter showing total invoice calculations with `http_request_counter_invoice`. This business metric answers "How many invoices have we generated?"

**System Health Summary Panel**: Stat panels (single large numbers) showing current request rate, error rate percentage, and average latency. These give at-a-glance health status.

**Alerting Capabilities**: Grafana can evaluate queries periodically and send alerts via email, Slack, PagerDuty, or other channels when thresholds are exceeded. Example alerts:

- Error rate exceeds 5% for 5 consecutive minutes
- P95 latency exceeds 1 second for 10 consecutive minutes
- Request rate drops by 90% (indicating possible outage)

**Metric Retention**: Prometheus is configured to retain metrics for 200 hours (approximately 8 days), providing a week of historical data for trend analysis and incident investigation. Production systems typically retain 30-90 days or use long-term storage solutions.

## Project Structure

```
toll-calculator/
├── aggregator/                    # Aggregator microservice
│   ├── client/                    # Client implementations for other services
│   │   ├── client.go              # Client interface
│   │   ├── grpc_client.go         # gRPC client implementation
│   │   └── http_client.go         # HTTP client implementation
│   ├── grpc.go                    # gRPC server implementation
│   ├── http.go                    # HTTP handlers and metrics
│   ├── main.go                    # Service entry point
│   ├── middleware.go              # Logging and metrics middleware
│   ├── service.go                 # Core business logic
│   └── store.go                   # Storage interface and in-memory implementation
│
├── data_receiver/                 # Data receiver microservice
│   ├── main.go                    # WebSocket server and entry point
│   ├── middleware.go              # Logging middleware
│   └── producer.go                # Kafka producer implementation
│
├── distance_calculator/           # Distance calculator microservice
│   ├── consumer.go                # Kafka consumer implementation
│   ├── main.go                    # Service entry point
│   ├── middleware.go              # Logging middleware
│   └── service.go                 # Distance calculation logic
│
├── gateway/                       # API Gateway
│   ├── config/
│   │   └── constants.go           # Configuration constants
│   ├── handler/
│   │   └── invoice.go             # Invoice handler
│   ├── utils/
│   │   └── http.go                # HTTP utilities
│   └── main.go                    # Gateway entry point
│
├── obu/                           # OBU simulator
│   └── main.go                    # Coordinate generation and WebSocket client
│
├── types/                         # Shared types and Protocol Buffers
│   ├── ptypes.proto               # Protocol Buffer definitions
│   ├── ptypes.pb.go               # Generated protobuf code
│   ├── ptypes_grpc.pb.go          # Generated gRPC code
│   └── types.go                   # Go type definitions
│
├── grafana/                       # Grafana configuration
│   └── provisioning/
│       └── datasources/
│           └── prometheus.yml     # Prometheus data source config
│
├── bin/                           # Compiled binaries (generated)
│
├── .env                           # Environment variables
├── docker-compose.yml             # Docker Compose configuration
├── prometheus.yml                 # Prometheus scrape configuration
├── Makefile                       # Build and run commands
├── go.mod                         # Go module definition
├── go.sum                         # Go module checksums
└── README.md                      # This file
```

## Setup, Configuration & Development

This section provides comprehensive guidance for setting up your development environment, configuring the system, building and running all services, and working with the codebase. Following these instructions will have you running the complete toll calculator system within minutes.

### Prerequisites and System Requirements

Before beginning, ensure your development machine meets these requirements and has the necessary software installed:

**Required Software and Tools**:

**Go 1.23 or Higher**: The entire system is written in Go, requiring Go 1.23 for compatibility with dependencies and language features. Verify your installation with `go version`. If Go is not installed or is an older version, download it from https://go.dev/dl/.

**Docker and Docker Compose**: The infrastructure components (Kafka, Prometheus, Grafana) run in Docker containers orchestrated by Docker Compose. This approach ensures consistent infrastructure across development, testing, and production environments without requiring manual installation of Kafka brokers or metrics systems. Verify with `docker --version` and `docker compose version`. Install Docker Desktop for macOS/Windows or Docker Engine for Linux if needed.

**Protocol Buffer Compiler (Development Only)**: If you plan to modify the gRPC service definitions in `types/ptypes.proto`, you need the protoc compiler to regenerate Go code from the proto schema. Most users won't need this unless changing service contracts.

Installation varies by platform:

```bash
# macOS
brew install protobuf

# Ubuntu/Debian
sudo apt-get install protobuf-compiler

# Verify installation
protoc --version
```

**Go Protocol Buffer Plugins (Development Only)**: These plugins enable protoc to generate Go-specific code. Install them globally:

```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Ensure `$GOPATH/bin` is in your PATH so protoc can find these plugins.

**System Resource Requirements**:

- **Memory**: Minimum 4GB RAM, 8GB recommended. Kafka and Prometheus are memory-intensive when processing high message volumes.
- **Disk Space**: Minimum 2GB free for Docker images, compiled binaries, and Kafka/Prometheus data storage.
- **Operating System**: macOS, Linux, or Windows with WSL2. Native Windows is not recommended due to path and shell script compatibility issues.

### Installation and Initial Setup

**Step 1: Clone the Repository**

Clone the project and navigate into the directory:

```bash
git clone <repository-url>
cd toll-calculator
```

**Step 2: Install Go Dependencies**

Download all required Go modules specified in `go.mod`:

```bash
go mod download
```

This fetches dependencies including Kafka clients, gRPC libraries, Logrus, Prometheus client, and WebSocket libraries. Dependencies are cached in `$GOPATH/pkg/mod` for reuse across projects.

**Step 3: Configure Environment Variables**

The system uses a `.env` file for configuration, following twelve-factor app principles for externalizing configuration. Create or verify the `.env` file in the project root:

```bash
cat .env
```

Default configuration (create this file if it doesn't exist):

```env
# Aggregator HTTP server port - REST API and metrics endpoint
AGG_HTTP_PORT=:3000

# Aggregator gRPC server port - high-performance distance aggregation
AGG_GRPC_PORT=:3001

# Storage backend type - currently only 'memory' implemented
# Future: 'postgres', 'redis', 'dynamodb'
AGG_STORE_TYPE=memory

# Aggregator service endpoint for Gateway to connect to
AGG_SERVICE_ENDPOINT=http://localhost:3000
```

These environment variables are loaded by services at startup using the godotenv library. You can override them by setting environment variables before running services, which takes precedence over `.env` values.

**Step 4: Start Infrastructure Services**

The Docker Compose configuration defines three infrastructure services:

```bash
docker compose up -d
```

The `-d` flag runs containers in detached mode (background). This command starts:

- **Kafka Broker**: Message queue for event-driven architecture
- **Prometheus**: Metrics collection and storage
- **Grafana**: Metrics visualization and dashboarding

**Verify Infrastructure Health**:

```bash
docker compose ps
```

Expected output showing all services healthy:

```
NAME         IMAGE                       STATUS
broker       apache/kafka:latest         Up (healthy)
grafana      grafana/grafana:latest      Up
prometheus   prom/prometheus:latest      Up
```

If Kafka shows "Up (health: starting)", wait 30-60 seconds for the health check to pass. Kafka requires initialization time to create internal topics and become ready.

**Inspect Logs if Issues Occur**:

```bash
# View all infrastructure logs
docker compose logs

# View specific service logs
docker compose logs broker
docker compose logs prometheus
```

**Step 5: Build the Microservices**

Compile all Go services into executable binaries in the `bin/` directory:

```bash
# Build all services individually
go build -o bin/obu obu/main.go
go build -o bin/receiver ./data_receiver
go build -o bin/calculator ./distance_calculator
go build -o bin/agg ./aggregator
go build -o bin/gateway gateway/main.go
```

Alternatively, use the provided Makefile which handles building and running in one command (see Development section below).

The compiled binaries are platform-specific executables. On macOS, you may need to code-sign them to prevent security warnings:

```bash
codesign -s - bin/receiver
codesign -s - bin/calculator
codesign -s - bin/agg
codesign -s - bin/gateway
codesign -s - bin/obu
```

### Running the System

**Service Startup Order and Dependencies**:

Services have dependencies that dictate startup order. Starting them out of order causes connection failures as services attempt to connect to dependencies that aren't yet running.

**Dependency Graph**:

- Infrastructure (Kafka, Prometheus, Grafana) - no dependencies, start first
- Aggregator - depends on infrastructure being ready
- Distance Calculator - depends on Kafka and Aggregator
- Data Receiver - depends on Kafka
- Gateway - depends on Aggregator
- OBU Simulator - depends on Data Receiver

**Recommended Startup Sequence**:

**1. Start Infrastructure** (if not already running from Installation):

```bash
docker compose up -d
```

Wait for Kafka health check using `docker compose ps`. The broker must show "Up (healthy)" before proceeding.

**2. Start the Aggregator**:

The Aggregator must start first among application services because the Distance Calculator sends calculated distances to it.

```bash
make agg
```

Or manually:

```bash
go build -o bin/agg ./aggregator && ./bin/agg
```

Expected console output:

```
INFO[0000] Starting distance aggregator gRPC Transport Layer on port :3001
INFO[0000] Starting distance aggregator HTTP Transport Layer on port :3000
```

The Aggregator is now accepting gRPC aggregation requests on port 3001, HTTP REST requests on port 3000, and exposing metrics at http://localhost:3000/metrics.

**3. Start the Distance Calculator**:

The calculator consumes from Kafka and sends to the Aggregator, so both must be running.

```bash
make calculator
```

Or manually:

```bash
go build -o bin/calculator ./distance_calculator && ./bin/calculator
```

Expected output:

```
INFO[0000] Distance Calculator service
INFO[0000] kafka transport started
```

The calculator is now polling Kafka for OBU data events.

**4. Start the Data Receiver**:

The receiver accepts WebSocket connections and produces to Kafka.

```bash
make receiver
```

Or manually:

```bash
go build -o bin/receiver ./data_receiver && ./bin/receiver
```

Expected output:

```
INFO[0000] Starting data receiver on :30000
```

The WebSocket server is now listening on ws://localhost:30000/ws for OBU connections.

**5. Start the API Gateway** (Optional):

The gateway provides a public API for external clients. It's optional because you can also query the Aggregator directly.

```bash
make gateway GATEWAY_PORT=:8080
```

Or manually:

```bash
go build -o bin/gateway gateway/main.go && ./bin/gateway -httpListenAddr=:8080
```

Expected output:

```
INFO[0000] Starting api gateway on port :8080
```

**6. Start the OBU Simulator**:

Finally, start the simulator to generate vehicle location data.

```bash
make obu
```

Or manually:

```bash
go build -o bin/obu obu/main.go && ./bin/obu
```

The simulator immediately begins transmitting data. You should see corresponding logs in the Data Receiver and Distance Calculator terminals indicating data flow through the system.

**Running Multiple Services Simultaneously**:

For development and testing, run each service in a separate terminal window/tab, or use terminal multiplexers:

**Using Separate Terminals**:

```bash
# Terminal 1
make agg

# Terminal 2
make calculator

# Terminal 3
make receiver

# Terminal 4
make gateway GATEWAY_PORT=:8080

# Terminal 5
make obu
```

**Using tmux** (recommended for development):

```bash
# Create a new tmux session
tmux new -s toll-calc

# Split into panes (Ctrl+B then %)
# In each pane, run a different service
```

**Stopping Services Cleanly**:

**Stop Go Services**: Press `Ctrl+C` in each terminal. This sends SIGINT, triggering graceful shutdown logic that completes in-flight requests and closes connections cleanly.

**Stop Infrastructure**:

```bash
docker compose down
```

**Stop Infrastructure and Delete Data**:

```bash
docker compose down -v
```

The `-v` flag removes volumes, deleting Kafka messages, Prometheus metrics, and Grafana configurations—useful for starting completely fresh.

### Configuration Details

**Environment Variable Reference**:

All configuration is externalized via environment variables loaded from the `.env` file:

- **AGG_HTTP_PORT** (default: `:3000`): Aggregator's HTTP server port for REST API and metrics
- **AGG_GRPC_PORT** (default: `:3001`): Aggregator's gRPC server port for high-performance aggregation
- **AGG_STORE_TYPE** (default: `memory`): Storage backend selection. Currently only `memory` is implemented, but the interface supports adding `postgres`, `redis`, etc.
- **AGG_SERVICE_ENDPOINT** (default: `http://localhost:3000`): Full URL for Gateway to reach Aggregator

**Service Port Allocation**:

| Service         | Port  | Protocol  | Access URL              | Purpose               |
| --------------- | ----- | --------- | ----------------------- | --------------------- |
| Kafka Broker    | 9092  | Kafka     | localhost:9092          | Message streaming     |
| Data Receiver   | 30000 | WebSocket | ws://localhost:30000/ws | OBU data ingestion    |
| Aggregator HTTP | 3000  | HTTP      | http://localhost:3000   | REST API & metrics    |
| Aggregator gRPC | 3001  | gRPC      | localhost:3001          | Distance aggregation  |
| Gateway         | 8000  | HTTP      | http://localhost:8000   | Public API            |
| Prometheus      | 9090  | HTTP      | http://localhost:9090   | Metrics database      |
| Grafana         | 4000  | HTTP      | http://localhost:4000   | Metrics visualization |

**Kafka Configuration**:

The system uses a single Kafka topic for OBU data:

**Topic**: `obudata` - Contains all vehicle location events from the Data Receiver

**Producer Configuration** (Data Receiver):

- Bootstrap servers: `localhost:9092`
- Delivery timeout: 10 seconds
- Asynchronous production with delivery callbacks
- No explicit acknowledgment requirements (fire-and-forget for demonstration)

**Consumer Configuration** (Distance Calculator):

- Bootstrap servers: `localhost:9092`
- Group ID: `myGroup` - enables load balancing across multiple calculator instances
- Auto-offset reset: `earliest` - process all messages from beginning on first start
- Session timeout: 6000ms - maximum time between heartbeats before considered dead
- Heartbeat interval: 2000ms - frequency of heartbeat messages
- Max poll interval: 300000ms - maximum time between poll() calls
- Auto-commit enabled: offsets committed automatically every 1000ms

These settings balance reliability (not losing messages) with performance (efficient batch processing).

### Development Workflows

**Using the Makefile**:

The Makefile provides convenient targets for building and running services:

```bash
# Build and run individual services
make obu           # Start OBU simulator
make receiver      # Start data receiver
make calculator    # Start distance calculator
make agg           # Start aggregator
make gateway       # Start API gateway with default port

# Build gateway with custom port
make gateway GATEWAY_PORT=:9000

# Regenerate Protocol Buffer code after modifying types/ptypes.proto
make proto
```

**Regenerating Protocol Buffers**:

If you modify `types/ptypes.proto`, regenerate Go code:

```bash
make proto
```

This executes:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       types/ptypes.proto
```

Which generates:

- `types/ptypes.pb.go` - Message serialization code
- `types/ptypes_grpc.pb.go` - gRPC service stubs and interfaces

**Code Organization and Development Patterns**:

**Middleware Pattern**: Services use middleware for cross-cutting concerns. When adding new middleware:

1. Implement as a function that takes and returns the service interface
2. Apply in service initialization before starting servers
3. Chain multiple middleware in desired order (outer to inner execution)

**Interface-Based Design**: Business logic is defined by interfaces, with implementations injected. When adding features:

1. Define the interface contract (method signatures)
2. Implement the interface in a dedicated file
3. Write unit tests against the interface using mocks
4. Wire up the implementation in `main.go`

**Graceful Shutdown**: All HTTP servers implement graceful shutdown. When modifying services:

1. Create a context from `signal.NotifyContext` for SIGINT/SIGTERM
2. Pass context to background goroutines
3. In shutdown, stop accepting new connections and wait for in-flight requests
4. Set reasonable shutdown timeout (5-30 seconds)

**Testing the System**:

Verify end-to-end functionality:

```bash
# Wait a few seconds after starting the simulator
sleep 5

# Query an invoice for OBU ID 5 via Gateway
curl http://localhost:8080/invoice?id=5

# Expected response
{
  "obuID": 5,
  "totalDistance": 23.456,
  "totalAmount": 86.7872
}

# Query directly via Aggregator
curl http://localhost:3000/invoice?id=5

# Test aggregation endpoint directly
curl -X POST http://localhost:3000/aggregate \
  -H "Content-Type: application/json" \
  -d '{"obuID":999,"value":42.5,"unix":1234567890,"requestId":"test-123"}'

# Verify metrics exposure
curl http://localhost:3000/metrics | grep aggregator
```

## API Documentation

### Aggregator Service

#### POST /aggregate

Aggregate distance data for an OBU.

**Request**:

```bash
curl -X POST http://localhost:3000/aggregate \
  -H "Content-Type: application/json" \
  -d '{
    "obuID": 12345,
    "value": 10.5,
    "unix": 1691234567890,
    "requestId": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

**Response**: `200 OK`

```json
{}
```

#### GET /invoice

Retrieve invoice for a specific OBU.

**Request**:

```bash
curl http://localhost:3000/invoice?id=12345
```

**Response**: `200 OK`

```json
{
  "obuID": 12345,
  "totalDistance": 150.75,
  "totalAmount": 557.775
}
```

**Error Responses**:

- `400 Bad Request`: Missing or invalid `id` parameter

  ```json
  {
    "error": "missing or incorrect 'obuid' query parameter"
  }
  ```

- `500 Internal Server Error`: OBU not found
  ```json
  {
    "error": "could not find data for the obuid 12345"
  }
  ```

#### GET /metrics

Prometheus metrics endpoint.

**Request**:

```bash
curl http://localhost:3000/metrics
```

**Response**: Prometheus text format metrics

### Gateway Service

#### GET /invoice

Retrieve invoice through the API Gateway.

**Request**:

```bash
curl http://localhost:8080/invoice?id=12345
```

**Response**: Same as Aggregator `/invoice` endpoint

### gRPC Service (Aggregator)

**Service**: `Aggregator`
**Method**: `Aggregate`
**Endpoint**: `localhost:3001`

**Request** (AggregateRequest):

```protobuf
{
  "ObuID": 12345,
  "Value": 10.5,
  "Unix": 1691234567890,
  "RequestID": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Response**: Empty (None)

**Example using grpcurl**:

```bash
grpcurl -plaintext -d '{
  "ObuID": 12345,
  "Value": 10.5,
  "Unix": 1691234567890,
  "RequestID": "test-request"
}' localhost:3001 Aggregator/Aggregate
```

## Configuration

### Environment Variables

Configure services via `.env` file:

```env
# Aggregator Configuration
AGG_HTTP_PORT=:3000              # Aggregator HTTP server port
AGG_GRPC_PORT=:3001              # Aggregator gRPC server port
AGG_STORE_TYPE=memory            # Storage type (currently only 'memory' supported)
AGG_SERVICE_ENDPOINT=http://localhost:3000  # Aggregator endpoint for Gateway
```

### Service Ports

| Service         | Port  | Protocol  | Access                  |
| --------------- | ----- | --------- | ----------------------- |
| Kafka Broker    | 9092  | TCP       | localhost:9092          |
| Data Receiver   | 30000 | WebSocket | ws://localhost:30000/ws |
| Aggregator HTTP | 3000  | HTTP      | http://localhost:3000   |
| Aggregator gRPC | 3001  | gRPC      | localhost:3001          |
| Gateway         | 8000  | HTTP      | http://localhost:8000   |
| Prometheus      | 9090  | HTTP      | http://localhost:9090   |
| Grafana         | 4000  | HTTP      | http://localhost:4000   |

### Kafka Configuration

**Topic**: `obudata`

**Producer Configuration** (Data Receiver):

```go
bootstrap.servers: localhost:9092
```

**Consumer Configuration** (Distance Calculator):

```go
bootstrap.servers:       localhost:9092
group.id:                myGroup
auto.offset.reset:       earliest
session.timeout.ms:      6000
heartbeat.interval.ms:   2000
max.poll.interval.ms:    300000
enable.auto.commit:      true
auto.commit.interval.ms: 1000
```

### Prometheus Configuration

**Scrape Interval**: 15 seconds
**Scrape Timeout**: 10 seconds
**Target**: `host.docker.internal:3000` (Aggregator metrics)

Edit `prometheus.yml` to add more targets:

```yaml
scrape_configs:
  - job_name: "aggregator"
    scrape_interval: 15s
    scrape_timeout: 10s
    metrics_path: /metrics
    static_configs:
      - targets: ["host.docker.internal:3000"]
```

### Grafana Configuration

**Admin Password**: `admin` (change in production)
**HTTP Port**: 4000

Auto-provisioned data source: Prometheus at `http://prometheus:9090`

## Development

### Building Individual Services

Use the Makefile for convenient builds:

```bash
# Build and run OBU simulator
make obu

# Build and run Data Receiver
make receiver

# Build and run Distance Calculator
make calculator

# Build and run Aggregator
make agg

# Build and run Gateway with custom port
make gateway GATEWAY_PORT=:9000
```

### Regenerating Protocol Buffers

After modifying `types/ptypes.proto`:

```bash
make proto
```

This runs:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       types/ptypes.proto
```

### Code Signing (macOS)

The Makefile includes code signing for certain binaries on macOS:

```bash
codesign -s - bin/receiver
codesign -s - bin/calculator
```

This prevents macOS security warnings.

### Project Patterns

#### Middleware Pattern

Services use middleware for cross-cutting concerns:

```go
// Service with middleware
svc = NewInvoiceAggregator(store)
svc = NewMetricsMiddleware(svc)
svc = NewLogMiddleware(svc)
```

#### Middleware Chain

Multiple middlewares can be chained:

```go
func Chain(a Aggregator, mws ...func(Aggregator) Aggregator) Aggregator {
    for i := len(mws) - 1; i >= 0; i-- {
        a = mws[i](a)
    }
    return a
}

svc = Chain(svc,
    func(s Aggregator) Aggregator { return NewMetricsMiddleware(s) },
    func(s Aggregator) Aggregator { return NewLogMiddleware(s) },
)
```

#### Interface-Based Design

Services are defined via interfaces for testability and flexibility:

```go
type Aggregator interface {
    AggregateDistance(context.Context, Distance) error
    CalculateInvoice(context.Context, int) (*Invoice, error)
}

type Storer interface {
    Insert(context.Context, Distance) error
    Get(context.Context, int) (float64, error)
}
```

#### Graceful Shutdown

All HTTP services implement graceful shutdown:

```go
func gracefulShutdown(ctx context.Context, timeout time.Duration, srv *http.Server) {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    wg := &sync.WaitGroup{}
    wg.Add(1)

    go func() {
        defer wg.Done()
        if err := srv.Shutdown(ctx); err != nil {
            logrus.Errorf("Failed to gracefully shutdown server %v", err)
        }
    }()
    wg.Wait()
    logrus.Info("Graceful shutdown complete")
}
```

## Monitoring the Services

### Viewing Logs

Services use structured logging. Monitor logs in real-time:

```bash
# View specific service logs (if running via Docker)
docker compose logs -f broker

# View application logs in terminal where service is running
# Example aggregator log:
INFO[0001] Aggregate distance    distance={...} requestId=... took=500µs
INFO[0002] Calculated Invoice    OBUID=12345 totalDist=150.75 totalAmount=557.775
```

### Prometheus Queries

Access Prometheus UI at `http://localhost:9090`

**Useful Queries**:

1. **Request Rate**:

   ```promql
   rate(aggregator_request_counter[1m])
   ```

2. **Error Rate**:

   ```promql
   rate(aggregator_error_counter[1m])
   ```

3. **Average Latency**:

   ```promql
   rate(aggregator_request_latency_sum[1m]) / rate(aggregator_request_latency_count[1m])
   ```

4. **95th Percentile Latency**:

   ```promql
   histogram_quantile(0.95, rate(aggregator_request_latency_bucket[1m]))
   ```

5. **HTTP Request Count by Endpoint**:
   ```promql
   http_request_counter_aggregate
   http_request_counter_invoice
   ```

### Grafana Dashboards

1. Access Grafana at `http://localhost:4000`
2. Login with `admin` / `admin`
3. Navigate to **Dashboards** → **New** → **Import**
4. Create custom dashboards using Prometheus metrics

**Recommended Dashboard Panels**:

- **Request Rate**: Line graph showing requests per second
- **Error Rate**: Line graph showing errors per second
- **Latency Heatmap**: Distribution of request latencies
- **Invoice Calculations**: Counter of total invoices generated
- **Distance Aggregations**: Counter of distance aggregation operations

### Kafka Monitoring

**View Topics**:

```bash
docker exec -it broker kafka-topics --bootstrap-server localhost:9092 --list
```

**View Consumer Groups**:

```bash
docker exec -it broker kafka-consumer-groups --bootstrap-server localhost:9092 --list
```

**Describe Consumer Group**:

```bash
docker exec -it broker kafka-consumer-groups --bootstrap-server localhost:9092 --group myGroup --describe
```

**Consume Messages (for debugging)**:

```bash
docker exec -it broker kafka-console-consumer --bootstrap-server localhost:9092 --topic obudata --from-beginning
```

### Health Checks

**Kafka Health**:

```bash
docker compose ps broker
# Should show "Up (healthy)"
```

**Aggregator HTTP**:

```bash
curl -I http://localhost:3000/metrics
# Should return 200 OK
```

**Data Receiver WebSocket**:

```bash
# Use wscat or similar WebSocket client
wscat -c ws://localhost:30000/ws
```

## Testing

Comprehensive testing ensures the system behaves correctly under various conditions, from individual component verification to full end-to-end flow validation and performance characterization under load. This section outlines testing strategies at multiple levels, from manual smoke tests to automated integration tests and load testing for capacity planning.

### End-to-End System Validation

**Complete Flow Verification**: This test validates that data flows correctly through the entire system—from OBU simulation through distance calculation and aggregation to final invoice retrieval. This is the primary smoke test to run after making changes or deploying updates.

**Prerequisites**: All infrastructure services must be running (`docker compose up -d`) and healthy.

**Test Procedure**:

```bash
# Terminal 1: Start Aggregator
make agg

# Terminal 2: Start Distance Calculator
make calculator

# Terminal 3: Start Data Receiver
make receiver

# Terminal 4: Start OBU Simulator
make obu
```

**Validation Steps**:

After starting all services, wait 5-10 seconds for data to flow through the system, then query invoices:

```bash
# Query invoice for OBU ID 5 (or any ID from 0-19)
curl http://localhost:3000/invoice?id=5

# Expected response structure
{
  "obuID": 5,
  "totalDistance": 42.856,  # Actual value varies based on simulated movement
  "totalAmount": 158.5672   # totalDistance × 3.7
}
```

**Interpreting Results**:

- **Success**: JSON response with non-zero totalDistance and corresponding totalAmount
- **Failure (404/500)**: Indicates data isn't flowing through the system—check service logs
- **Zero Distance**: Simulator may not have started or data hasn't propagated yet—wait longer and retry

**Observing Data Flow**: Watch service logs to see data moving through the pipeline:

```bash
# Data Receiver logs show WebSocket connections and Kafka production
# Distance Calculator logs show Kafka consumption and distance calculations
# Aggregator logs show aggregation operations and metric recording
```

### Component-Level Testing

Testing individual components in isolation verifies their behavior without requiring the entire system to be running. This is useful for development, debugging specific services, and validating API contracts.

**Testing Data Receiver (WebSocket Endpoint)**:

The Data Receiver accepts WebSocket connections and publishes received data to Kafka. Test it by manually sending OBU data:

```bash
# Install wscat if needed: npm install -g wscat
wscat -c ws://localhost:30000/ws

# Once connected, send JSON payload
> {"obuID":99999,"currLat":10.5,"currLong":20.3,"prevLat":10.0,"prevLong":20.0,"requestId":"test-123"}
```

**Expected Behavior**: The Data Receiver logs should show the received message and Kafka production. You won't receive a WebSocket response (fire-and-forget pattern), but checking Kafka confirms the message was published:

```bash
# Verify message reached Kafka
docker exec -it broker kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic obudata --from-beginning --max-messages 1
```

**Testing Aggregator HTTP Endpoint**:

The Aggregator exposes both HTTP and gRPC interfaces. Test the HTTP interface for manual distance aggregation and invoice retrieval:

```bash
# Manually aggregate distance for a test OBU
curl -X POST http://localhost:3000/aggregate \
  -H "Content-Type: application/json" \
  -d '{"obuID":99999,"value":15.5,"unix":1691234567890,"requestId":"test-manual-123"}'

# Expected response: 200 OK with empty JSON body
{}

# Retrieve invoice for the test OBU
curl http://localhost:3000/invoice?id=99999

# Expected response
{
  "obuID": 99999,
  "totalDistance": 15.5,
  "totalAmount": 57.35  # 15.5 × 3.7
}
```

**Testing Scenarios**:

- **Multiple Aggregations**: Send multiple aggregate requests with the same OBU ID and verify totalDistance accumulates
- **Different OBUs**: Send aggregations for different OBU IDs and verify isolation (distances don't mix)
- **Error Handling**: Query non-existent OBU IDs and verify appropriate error responses

**Testing Aggregator gRPC Endpoint**:

The gRPC interface is used by the Distance Calculator for high-performance aggregation. Test it using grpcurl, a command-line tool for gRPC:

```bash
# Install grpcurl if needed
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Send gRPC aggregation request
grpcurl -plaintext -d '{"ObuID":99999,"Value":25.5,"Unix":1691234567890,"RequestID":"grpc-test-456"}' \
  localhost:3001 Aggregator/Aggregate

# Expected response: Empty (None message type)
{}
```

**Verifying gRPC Aggregation**: After sending via gRPC, query via HTTP to confirm the aggregation was recorded:

```bash
curl http://localhost:3000/invoice?id=99999

# totalDistance should now include the gRPC-sent value (15.5 + 25.5 = 41.0)
{
  "obuID": 99999,
  "totalDistance": 41.0,
  "totalAmount": 151.7
}
```

**Testing Gateway Proxying**:

The Gateway should proxy requests to the Aggregator transparently:

```bash
# Query via Gateway (runs on port 8000)
curl http://localhost:8000/invoice?id=99999

# Should return identical response to querying Aggregator directly
# This validates that the Gateway correctly proxies and doesn't modify responses
```

### Metrics and Observability Verification

Validating that metrics are being collected and exposed correctly is essential for production operation:

**Testing Prometheus Metrics Exposure**:

```bash
# Fetch metrics from Aggregator
curl -s http://localhost:3000/metrics

# Expected output includes Prometheus text format metrics
# HELP aggregator_request_counter Total number of aggregation requests
# TYPE aggregator_request_counter counter
aggregator_request_counter 42

# Filter for specific metrics
curl -s http://localhost:3000/metrics | grep aggregator_request_counter

# Verify histogram buckets for latency
curl -s http://localhost:3000/metrics | grep aggregator_request_latency_bucket
```

**Testing Prometheus Scraping**:

Verify Prometheus is successfully scraping the Aggregator:

```bash
# Open Prometheus UI
open http://localhost:9090

# Navigate to Status → Targets
# The "aggregator" target should show State: UP
```

**Testing Grafana Integration**:

```bash
# Access Grafana
open http://localhost:4000

# Login: admin / admin
# Navigate to Explore
# Select Prometheus data source
# Query: rate(aggregator_request_counter[1m])
# Should show graph of request rate
```

**Unit Testing Guidance**: While this project focuses on demonstrating architecture, production systems should include unit tests for all business logic:

- **Service Layer Tests**: Mock storage interfaces and verify business logic
- **Storage Tests**: Test in-memory store operations with concurrent access
- **Middleware Tests**: Verify logging and metrics middleware behavior
- **Handler Tests**: Test HTTP/gRPC handlers with mock services

Example unit test structure:

```go
func TestInvoiceCalculation(t *testing.T) {
    // Create mock storage
    mockStore := &MockStorer{
        distances: map[int]float64{123: 100.0},
    }

    // Create service with mock
    svc := NewInvoiceAggregator(mockStore)

    // Test invoice calculation
    invoice, err := svc.CalculateInvoice(context.Background(), 123)

    // Assert expectations
    assert.NoError(t, err)
    assert.Equal(t, 100.0, invoice.TotalDistance)
    assert.Equal(t, 370.0, invoice.TotalAmount)
}
```
