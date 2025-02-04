FROM python:3.9-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    git \
    && rm -rf /var/lib/apt/lists/*

# Set environment variables
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PIP_NO_CACHE_DIR=1 \
    PIP_DISABLE_PIP_VERSION_CHECK=1

# Copy requirements first (for better caching)
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy source code
COPY . .

# Generate protobuf files
RUN chmod +x scripts/generate_protos.sh
RUN ./scripts/generate_protos.sh

# Create non-root user
RUN useradd -m -u 1000 appuser && \
    chown -R appuser:appuser /app
USER appuser

# Set default configuration
ENV GRPC_PORT=50051 \
    METRICS_PORT=8000 \
    LOG_LEVEL=INFO \
    DEVICE=cpu \
    OTEL_OTLP_ENDPOINT=otel-collector:4317 \
    OTEL_RESOURCE_ATTRIBUTES=service.name=ner-service \
    OTEL_SERVICE_NAME=ner-service \
    REDIS_POOL_SIZE=10 \
    RATE_LIMIT=100 \
    CACHE_TTL=3600 \
    LOCAL_CACHE_SIZE=1000 \
    CACHE_VERSION=v1.0 \
    ELASTICSEARCH_HOST=elasticsearch:9200 \
    BATCH_SIZE=16

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "\
    import grpc; \
    from grpc_health.v1 import health_pb2, health_pb2_grpc; \
    channel = grpc.insecure_channel('localhost:${GRPC_PORT}'); \
    stub = health_pb2_grpc.HealthStub(channel); \
    response = stub.Check(health_pb2.HealthCheckRequest()); \
    assert response.status == health_pb2.HealthCheckResponse.SERVING"

# Run the service
CMD ["python", "main.py"]

# Document ports
EXPOSE ${GRPC_PORT}
EXPOSE ${METRICS_PORT}

# Labels
LABEL maintainer="NER Service Team" \
    version="1.0.0" \
    description="NER service with time extraction capabilities" \
    org.opencontainers.image.source="https://github.com/yourusername/mail2calendar" \
    org.opencontainers.image.licenses="Apache-2.0" \
    com.docker.compose.project="mail2calendar"

# Set stop signal
STOPSIGNAL SIGTERM