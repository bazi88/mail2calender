FROM python:3.9-slim

WORKDIR /app

# Cài đặt các dependencies hệ thống
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    pkg-config \
    git \
    && rm -rf /var/lib/apt/lists/*

# Cài đặt các dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy proto files
COPY protos/ ./protos/

# Copy test client
COPY test_client.py .

# Generate proto files
RUN cd /app/protos && \
    python -m grpc_tools.protoc \
    -I. \
    --python_out=. \
    --grpc_python_out=. \
    ner.proto && \
    sed -i 's/import ner_pb2/from . import ner_pb2/' ner_pb2_grpc.py

# Tạo __init__.py để import được
RUN touch protos/__init__.py

CMD ["python", "test_client.py"] 