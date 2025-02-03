#!/bin/bash

# Generate Python code from proto file
python -m grpc_tools.protoc \
    -I./protos \
    --python_out=. \
    --grpc_python_out=. \
    ./protos/ner/ner.proto

# Fix imports in generated files
sed -i 's/import ner_pb2 as ner__pb2/import ner_pb2 as ner__pb2/' ner_pb2_grpc.py 
