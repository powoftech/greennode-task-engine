package processor

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"greennode/worker-node/internal/db"

	amqp "github.com/rabbitmq/amqp091-go"
)

// JobMessage matches the JSON payload sent by the Java API Gateway
type JobMessage struct {
	JobID      string `json:"jobId"`
	TaskType   string `json:"taskType"`
	Complexity int    `json:"complexity"`
}

type Processor struct {
	dbConn  *db.Database
	rmqConn *amqp.Connection
	channel *amqp.Channel
}

// New initializes the RabbitMQ connection
func New(database *db.Database) (*Processor, error) {
	host := getEnv("MQ_HOST", "localhost")
	port := getEnv("MQ_PORT", "5672")
	user := getEnv("MQ_USER", "green_user")
	pass := getEnv("MQ_PASSWORD", "green_password")

	url := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, pass, host, port)
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	concurrencyStr := getEnv("WORKER_CONCURRENCY", "10")
	concurrency, err := strconv.Atoi(concurrencyStr)
	if err != nil || concurrency < 1 {
		concurrency = 10
	}

	err = ch.Qos(
		concurrency, // prefetch count
		0,           // prefetch size
		false,       // global
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	log.Println("Connected to RabbitMQ successfully")
	return &Processor{dbConn: database, rmqConn: conn, channel: ch}, nil
}

// Start consuming messages
func (p *Processor) Start() error {
	msgs, err := p.channel.Consume(
		"worker.jobs.queue", // queue
		"",                  // consumer
		false,               // auto-ack (MUST BE FALSE for reliability)
		false,               // exclusive
		false,               // no-local
		false,               // no-wait
		nil,                 // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	log.Println("Worker is waiting for messages...")

	concurrencyStr := getEnv("WORKER_CONCURRENCY", "10")
	concurrency, err := strconv.Atoi(concurrencyStr)
	if err != nil || concurrency < 1 {
		concurrency = 10
	}

	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for d := range msgs {
				p.processMessage(d)
			}
		}(i)
	}

	wg.Wait()

	return nil
}

func (p *Processor) processMessage(d amqp.Delivery) {
	var job JobMessage
	if err := json.Unmarshal(d.Body, &job); err != nil {
		log.Printf("Error decoding JSON: %v. Payload: %s", err, string(d.Body))
		d.Nack(false, false) // Reject and don't requeue (send to DLX)
		return
	}

	log.Printf("Received Job [%s] Type: %s", job.JobID, job.TaskType)

	// 1. Enforce Idempotency: Try to claim the job
	claimed, err := p.dbConn.ClaimJob(job.JobID)
	if err != nil {
		log.Printf("Database error claiming job [%s]: %v", job.JobID, err)
		d.Nack(false, true) // Requeue to try again later
		return
	}
	if !claimed {
		log.Printf("Job [%s] already processed or claimed by another worker. Acknowledging safely.", job.JobID)
		d.Ack(false)
		return
	}

	// 2. Simulate AI Workload
	log.Printf("Processing Job [%s] with complexity %d...", job.JobID, job.Complexity)
	time.Sleep(time.Duration(job.Complexity) * time.Second)

	// 3. Mark as Completed
	mockResult := fmt.Sprintf(`{"status": "success", "processed_in_seconds": %d}`, job.Complexity)
	if err := p.dbConn.CompleteJob(job.JobID, mockResult); err != nil {
		log.Printf("Failed to update database for job [%s]: %v", job.JobID, err)
		d.Nack(false, true) // Requeue so it can be retried
		return
	}

	// 4. Send ACK to RabbitMQ
	log.Printf("Job [%s] completed successfully.", job.JobID)
	d.Ack(false)
}

// Close cleans up connections
func (p *Processor) Close() {
	p.channel.Close()
	p.rmqConn.Close()
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
