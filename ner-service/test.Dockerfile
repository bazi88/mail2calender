FROM python:3.9-slim

WORKDIR /app

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy protos and test client
COPY protos/ ./protos/
COPY test_client.py .

# Generate protobuf files
RUN cd /app/protos && \
    python -m grpc_tools.protoc \
    -I. \
    --python_out=. \
    --grpc_python_out=. \
    ner.proto && \
    sed -i 's/import ner_pb2/from . import ner_pb2/' ner_pb2_grpc.py

# Create __init__.py
RUN touch protos/__init__.py

CMD ["python", "test_client.py"] 