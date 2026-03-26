package com.greennode.api_gateway.service;

import com.greennode.api_gateway.dto.JobMessage;
import com.greennode.api_gateway.dto.JobRequest;
import com.greennode.api_gateway.entity.Job;
import com.greennode.api_gateway.entity.JobStatus;
import com.greennode.api_gateway.entity.OutboxEvent;
import com.greennode.api_gateway.repository.JobRepository;
import com.greennode.api_gateway.repository.OutboxEventRepository;
import io.micrometer.tracing.Tracer;
import java.util.Optional;
import java.util.UUID;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import tools.jackson.core.JacksonException;
import tools.jackson.databind.json.JsonMapper;

/**
 * Orchestrates the core business logic. Uses the Transactional Outbox Pattern with Debezium CDC to
 * ensure atomicity between database writes and message publishing.
 */
@Service
public class JobService {

    private static final Logger log = LoggerFactory.getLogger(JobService.class);

    private final JobRepository jobRepository;
    private final OutboxEventRepository outboxEventRepository;
    private final JsonMapper jsonMapper;
    private final Tracer tracer;

    // Constructor Injection (Preferred over @Autowired)
    public JobService(
            JobRepository jobRepository,
            OutboxEventRepository outboxEventRepository,
            JsonMapper jsonMapper,
            Tracer tracer) {
        this.jobRepository = jobRepository;
        this.outboxEventRepository = outboxEventRepository;
        this.jsonMapper = jsonMapper;
        this.tracer = tracer;
    }

    /**
     * Submits a new job using the Transactional Outbox Pattern with Debezium CDC. @Transactional
     * ensures that both the Job and OutboxEvent are written atomically. Debezium captures the
     * outbox insert via CDC and publishes to RabbitMQ, preventing race conditions.
     */
    @Transactional
    public Job submitJob(JobRequest request) {
        UUID jobId = UUID.randomUUID();

        String currentTraceId =
                tracer.currentSpan() != null
                        ? tracer.currentSpan().context().traceId()
                        : "no-trace-id";

        Job job = new Job(jobId, request.getTaskType(), JobStatus.PENDING);
        job = jobRepository.save(job);
        log.info(
                "Persisted new job [{}] to database with status PENDING. TraceID: {}",
                jobId,
                currentTraceId);

        try {
            JobMessage message =
                    new JobMessage(
                            jobId, request.getTaskType(), request.getComplexity(), currentTraceId);
            String messageJson = jsonMapper.writeValueAsString(message);

            OutboxEvent event =
                    new OutboxEvent(
                            jobId.toString(), // aggregate_id
                            "JOB", // aggregate_type
                            "JOB_CREATED", // type
                            messageJson // payload
                            );
            outboxEventRepository.save(event);
            log.info("Persisted outbox event [{}] for job [{}]", event.getId(), jobId);
        } catch (JacksonException e) {
            log.error("Failed to serialize JobMessage for job [{}]: {}", jobId, e.getMessage());
            throw new RuntimeException("Failed to serialize job message", e);
        }

        return job;
    }

    public Optional<Job> getJobStatus(UUID jobId) {
        return jobRepository.findById(jobId);
    }
}
