# **Coding Agent System Prompt & Guidelines**

**Project:** Distributed "AI Task" Execution Engine

**Context:** This project is a strict proof-of-concept demonstrating distributed systems engineering (Java API, Go Workers, RabbitMQ, PostgreSQL, Docker) for an infrastructure backend role.

## **1\. Project Overview**

This system is an asynchronous distributed task engine simulating long-running AI workloads.

- A stateless **Java (Spring Boot) API Gateway** receives tasks and writes them to the database using the Outbox pattern.
- **Debezium Server** implements Change Data Capture (CDC) to reliably publish outbox events from PostgreSQL to RabbitMQ.
- **RabbitMQ** handles message brokering and dead-letter routing.
- **Go (Golang) Worker Nodes** consume messages concurrently, process them, and update the database.
- **PostgreSQL** serves as the single source of truth for job states and outbox events.

## **2\. Specific, Concrete Rules for LLMs**

- **Rule 1:** NEVER suggest introducing new external dependencies (e.g., Redis, Kafka, MongoDB) without explicit human approval. Stick to PostgreSQL and RabbitMQ.
- **Rule 2:** NEVER remove `JSON` struct tags in Go or `@JsonProperty` annotations in Java during refactoring.
- **Rule 3:** ALWAYS use environment variables (`os.Getenv` in Go, `@Value` or `Environment` in Java) for credentials and hostnames. NEVER hardcode secrets.
- **Rule 4:** When modifying database schemas, you MUST write the corresponding SQL migration file. Do not just update the ORM/structs.
- **Rule 5:** If a user asks to "implement the task," default to a simulated delay (e.g., `time.Sleep` in Go) rather than a complex AI model, unless strictly specified. Focus on the _infrastructure_.

## **3\. Coding Standards**

### **Java (API Gateway)**

- **Indentation:** 4 spaces (No tabs).
- **Naming:** `PascalCase` for Classes/Interfaces, `camelCase` for methods/variables.
- **Style:** Favor Constructor Injection over `@Autowired` field injection.
- **Validation:** Always use `jakarta.validation.constraints` on DTOs.
- **Logs:** Use SLF4J. Log all incoming requests and outgoing RabbitMQ publishes with the associated `jobId`.

### **Go (Worker Nodes)**

- **Indentation:** Tabs (Standard `gofmt`).
- **Naming:** Concise, descriptive package names (no `util` or `common`). Use `camelCase`. Exported functions/structs must have comments.
- **Error Handling:** Do not ignore errors. Always return or log errors wrapped with context: `fmt.Errorf("failed to process job %s: %w", jobID, err)`.
- **Concurrency:** Always use channels and `sync.WaitGroup` to manage goroutine lifecycles cleanly. Avoid global state.

## **4\. Architectural Decisions (Do Not Violate)**

- **Transactional Outbox & CDC:** The API Gateway MUST use the outbox pattern. State changes and outbox events must be written to PostgreSQL within a single local transaction. Debezium Server tails the PostgreSQL WAL and guarantees at-least-once delivery to RabbitMQ. Do NOT publish directly to RabbitMQ from the API Gateway.
- **Idempotency:** Go workers MUST check if a job is already in a `COMPLETED` or `PROCESSING` state before executing it. Messages may be delivered more than once by RabbitMQ.
- **Stateless Gateway:** The Java Spring Boot application must maintain zero in-memory state regarding the jobs. All state lives in PostgreSQL.
- **Connection Pooling:** Java must use HikariCP with a strict maximum pool size (e.g., 10\) to prevent PostgreSQL connection exhaustion under load.

## **5\. Security & Stability Considerations**

- **Input Sanitization:** Validate all incoming HTTP payloads at the API layer. Do not trust client input.
- **Graceful Shutdown:** Go workers must listen for `SIGINT`/`SIGTERM`. Upon receiving the signal, they must stop accepting new RabbitMQ messages, finish currently executing tasks, and then exit.
- **Dead Lettering:** Any message that errors out more than 3 times in the Go worker MUST be rejected without requeueing (`Nack(requeue=false)`), relying on RabbitMQ to route it to the Dead Letter Exchange (DLX).
