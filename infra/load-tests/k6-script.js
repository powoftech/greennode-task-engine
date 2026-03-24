// @ts-check
import http from "k6/http";
import { check, sleep } from "k6";
import { Trend } from "k6/metrics";

// Custom metric to track how long the entire asynchronous process takes
const e2eProcessingTime = new Trend("e2e_processing_time");

// 1. Configure the Load Test Stages
export const options = {
  stages: [
    { duration: "10s", target: 50 }, // Ramp up to 50 virtual users over 10 seconds
    { duration: "30s", target: 50 }, // Hold at 50 virtual users for 30 seconds
    { duration: "10s", target: 0 }, // Ramp down to 0 users
  ],
  thresholds: {
    // Gateway ingestion speed: 95% of POST requests within 200ms
    http_req_duration: ["p(95)<200"],
    http_req_failed: ["rate<0.01"],
    // System-wide E2E processing: 1000ms task + <2000ms infrastructure overhead
    e2e_processing_time: ["p(95)<3000"],
  },
};

// 2. The simulated user behavior
export default function () {
  // The target URL (Use an environment variable so we can change it to the Cloud IP later)
  const baseUrl = __ENV.API_URL || "http://localhost:8080";
  const url = `${baseUrl}/api/v1/jobs`;

  // Fixed complexity to 1 to precisely measure infrastructure overhead
  const randomComplexity = 1;

  const payload = JSON.stringify({
    taskType: "matrix_multiplication",
    complexity: randomComplexity,
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  // 3. Send the POST request
  const response = http.post(url, payload, params);

  // 4. Verify the response is 202 Accepted
  const isSuccessful = check(response, {
    "is status 202": (r) => r.status === 202,
  });

  if (isSuccessful) {
    const body = JSON.parse(response.body?.toString() || "{}");
    const jobId = body.id;

    // 5. Polling Loop: Wait for the Go worker to finish processing
    let status = "PENDING";
    let attempts = 0;
    const startTime = Date.now();

    // Poll every 2 seconds, up to 30 times
    while (status !== "COMPLETED" && attempts < 30) {
      sleep(2);
      attempts++;

      const getResponse = http.get(`${url}/${jobId}`);
      if (getResponse.status === 200) {
        const getBody = JSON.parse(getResponse.body?.toString() || "{}");
        status = getBody.status;
      }
    }

    // If it successfully processed, record the total E2E time
    if (status === "COMPLETED") {
      const endTime = Date.now();
      e2eProcessingTime.add(endTime - startTime);
    }
  }

  // 6. Short sleep to simulate real user pacing
  sleep(0.1);
}
