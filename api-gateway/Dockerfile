# Stage 1: Build the application
FROM maven:3.9-eclipse-temurin-25-alpine AS builder

WORKDIR /app

# Copy the pom.xml and source code
COPY pom.xml .
COPY src ./src

# Package the application (skip tests to speed up the docker build)
RUN mvn clean package -DskipTests

# Stage 2: Create the production runtime image
FROM eclipse-temurin:25-jre-alpine

WORKDIR /app

# Copy ONLY the compiled jar file from the builder stage
COPY --from=builder /app/target/*.jar app.jar

# Expose the API port
EXPOSE 8080

RUN addgroup -S gatewaygroup && adduser -S gatewayuser -G gatewaygroup
USER gatewayuser

# Run the application
ENTRYPOINT ["java", "-jar", "app.jar"]