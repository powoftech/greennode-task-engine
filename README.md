# Distributed Task Execution Engine

## 🚀 Overview

A cloud-native, distributed task execution system designed for high availability, fault tolerance, and scalability. Built as a proof-of-concept for modern backend engineering, this orchestrator seamlessly manages asynchronous job scheduling, robust message delivery via the Transactional Outbox pattern, and high-performance task processing.

**This project was specifically designed to demonstrate the critical engineering competencies outlined for cloud platform services: correctness, performance, maintainability, and systematic problem-solving.**

## 💡 System Architecture & Design Patterns

The architecture reflects best practices utilized in large-scale distributed platforms:

- **Polyglot Microservices:**
  - **API Gateway (Java / Spring Boot):** Manages RESTful ingress, handles job requests, and securely persists data.
  - **Worker Node (Go):** A high-throughput, low-latency executor for processing background tasks.
- **Transactional Outbox Pattern:** Guarantees _at-least-once_ message delivery without distributed transactions. Job creations are committed to a PostgreSQL outbox table, captured by **Debezium** via Change Data Capture (CDC), and streamed seamlessly into **RabbitMQ**.
- **Infrastructure as Code (IaC):** Utilizes **Terraform** to ensure cloud deployment reproducibility and maintain infra parity across environments.
- **Distributed Tracing & Observability:** Integrated with **Jaeger** and **OpenTelemetry**. Every request is tagged with a trace context propagated from the API Gateway to the Go Worker Node, ensuring comprehensive observability and systematic debugging capabilities.

## 🛠️ Technology Stack

- **Languages:** Java 25 (Spring Boot), Go 1.26
- **Database & CDC:** PostgreSQL, Debezium
- **Message Broker:** RabbitMQ
- **Observability:** Jaeger, OpenTelemetry
- **Infrastructure & DevOps:** Docker, Docker Compose, Terraform, k6 (Load Testing)

## 📊 Performance & Load Testing

Systematic load testing has been conducted using **k6** to validate performance, reliability, and scaling boundaries:

- **P95 Latency:** HTTP request duration maintained extremely low latency ($P95 = 5.44$ms) under load.
- **Throughput & Success Rate:** Successfully executed ~10,500 checks with a **100% success rate** (0 dropped requests, 0 status errors) across 1,000 concurrent Virtual Users (VUs).
- **End-to-End Reliability:** Analyzed and verified through RabbitMQ metrics and Jaeger trace spans, proving the exact delivery flow and measuring true End-to-End processing times up to ~36 seconds for intensive distributed batch jobs.

![k6](docs/images/k6-load-test-result.png)
![RabbitMQ Management UI](docs/images/rabbitmq-load-test-result.png)
![Jaeger UI](docs/images/jaeger-load-test-result.png)

_(Insert images from `docs/images/` here showcasing the k6 load test logs, RabbitMQ queue metrics, and Jaeger UI trace graphs)_

## 🚀 How to Run (Local Development)

The entire infrastructure and microservices stack is containerized for immediate execution.

```bash
# Start the supporting infrastructure (PostgreSQL, RabbitMQ, Jaeger, Debezium)
docker compose up -d

# The API Gateway is available at http://localhost:8080
# Jaeger UI is available at http://localhost:16686
# RabbitMQ Management is at http://localhost:15672
```

## 📈 Learning Goals & Future Improvements

- Implementing **Kubernetes** manifests for dynamic scaling of worker nodes based on queue depth.
- Integrating **Prometheus and Grafana** for extended time-series SLA monitoring.
- Adding OAuth2/JWT authentication at the API Gateway layer.
