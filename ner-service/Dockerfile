FROM python:3.9-slim

WORKDIR /app

# Install system dependencies
RUN apt-get update && apt-get install -y \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Copy requirements and install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Create Python package structure
RUN mkdir -p protos server model utils config
COPY protos/ ./protos/
COPY server/ ./server/
COPY model/ ./model/
COPY utils/ ./utils/
COPY config/ ./config/
COPY main.py .

# Create __init__.py files
RUN touch protos/__init__.py server/__init__.py model/__init__.py utils/__init__.py config/__init__.py

# Generate gRPC code
RUN cd /app && \
    python -m grpc_tools.protoc \
    -I. \
    --python_out=. \
    --grpc_python_out=. \
    protos/ner.proto

# Create directory for model cache
RUN mkdir -p /root/.cache/huggingface

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD python -c "\
    import grpc;\
    from grpc_health.v1 import health_pb2, health_pb2_grpc;\
    channel = grpc.insecure_channel('localhost:${GRPC_PORT:-50051}');\
    stub = health_pb2_grpc.HealthStub(channel);\
    response = stub.Check(health_pb2.HealthCheckRequest());\
    exit(0 if response.status == health_pb2.HealthCheckResponse.SERVING else 1)"

# Expose ports
EXPOSE ${GRPC_PORT:-50051}
EXPOSE ${METRICS_PORT:-8000}

# Set Python path
ENV PYTHONPATH=/app

# Run service
CMD ["python", "main.py"]