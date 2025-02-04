#!/bin/bash

# Exit on error
set -e

# Create directories if they don't exist
mkdir -p protos

# Generate Python code
python -m grpc_tools.protoc \
    --proto_path=protos \
    --python_out=protos \
    --grpc_python_out=protos \
    protos/ner.proto

# Create __init__.py if it doesn't exist
touch protos/__init__.py

echo "Proto files generated successfully!"