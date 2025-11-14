# Toll Calculator - Event-Driven Microservices System

A production-grade demonstration of **_event-driven microservices architecture_** implementing a distributed toll calculation system for highway vehicle tracking. The system processes real-time location data from **_On-Board Units (OBUs)_** through a sophisticated pipeline that calculates travel distances and generates toll invoices.

**Key Architectural Highlights:**

- **Event-Driven Communication**: **_Apache Kafka_** serves as the central message broker, enabling asynchronous, decoupled service communication with guaranteed message persistence and replay capabilities
- **Multiple Transport Layers**: Dual protocol implementation featuring **_HTTP REST APIs_** for external client access and human-readable interfaces, alongside **_gRPC_** for high-throughput, type-safe internal service-to-service communication
- **Real-Time Data Streaming**: **_WebSocket connections_** provide persistent, bidirectional channels for continuous OBU location data ingestion with minimal latency
- **Microservices Pattern**: Four specialized services (**_Data Receiver_**, **_Distance Calculator_**, **_Aggregator_**, **_Gateway_**) each with single responsibility, independent scalability, and clean separation of concerns
- **Comprehensive Observability**: Full monitoring stack with **_Prometheus_** metrics collection, **_Grafana_** dashboards, structured logging with **_Logrus_**, and distributed request tracing
- **Containerized Infrastructure**: **_Docker Compose_** orchestration for the complete ecosystem including Kafka, Zookeeper, Prometheus, and Grafana

This project serves as a practical reference implementation for developers building production-ready distributed systems, demonstrating how to handle real-time data processing, implement multiple communication protocols, ensure fault tolerance, and maintain operational visibility across microservices.
