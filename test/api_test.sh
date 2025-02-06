#!/bin/bash

# Màu sắc cho output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# URL cơ sở
BASE_URL="http://localhost:8080"

# Function kiểm tra service
confirm_service_running() {
    local response=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL")
    if [ "$response" -ne 200 ]; then
        echo -e "${RED}Service is not running or returned status $response. Aborting tests.${NC}"
        exit 1
    fi
    echo -e "${GREEN}Service is running. Proceeding with tests.${NC}"
}

# Function test API
test_api() {
    local endpoint=$1
    local method=$2
    local data=$3
    local expected_status=$4
    
    echo "Testing $method $endpoint"
    echo "Request data: $data"
    
    response=$(curl -s -w "\n%{http_code}" -X $method \
        -H "Content-Type: application/json" \
        -d "$data" \
        "$BASE_URL$endpoint")
    
    status_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | sed '$d')
    
    echo "Response: $response_body"
    echo "Status code: $status_code"
    
    if [ "$status_code" -eq "$expected_status" ]; then
        echo -e "${GREEN}Test passed${NC}"
    else
        echo -e "${RED}Test failed. Expected status $expected_status but got $status_code${NC}"
    fi
    echo "----------------------------------------"
}

# Kiểm tra service trước khi chạy các test
confirm_service_running

# Test 1: Extract entities với text hợp lệ
echo "Test 1: Extract entities với text hợp lệ"
test_api "/api/v1/ner/extract" "POST" '{"text":"Apple Inc. is located in Cupertino, California."}' 200

# Test 2: Extract entities với text rỗng
echo "Test 2: Extract entities với text rỗng"
test_api "/api/v1/ner/extract" "POST" '{"text":""}' 400

# Test 3: Extract entities với request không hợp lệ
echo "Test 3: Extract entities với request không hợp lệ"
test_api "/api/v1/ner/extract" "POST" '{"invalid_field":"test"}' 400

# Test 4: Extract entities với text dài
echo "Test 4: Extract entities với text dài"
test_api "/api/v1/ner/extract" "POST" '{"text":"Google is a technology company headquartered in Mountain View, California. Apple has its headquarters in Cupertino. Amazon has its headquarters in Seattle."}' 200

# Test 5: Rate limit test
echo "Test 5: Rate limit test (gửi nhiều request liên tiếp)"
for i in {1..5}; do
    test_api "/api/v1/ner/extract" "POST" '{"text":"Quick test for rate limit"}' 200
    sleep 1
done

# Kiểm tra tiêu đề bảo mật
curl -I http://localhost:8080/api/endpoint | grep "X-Content-Type-Options"
# Thêm các kiểm tra khác cho các tiêu đề bảo mật
```

name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
          cache: true
          
      - name: Lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest
          args: --timeout=5m
          
      - name: Test
        run: go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
        
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.txt

  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Run Snyk
        uses: snyk/actions/golang@master
        env:
          SNYK_TOKEN: ${{ secrets.SNYK_TOKEN }}

      - name: Run Gosec
        uses: securego/gosec@master
        with:
          args: ./...

name: CD

on:
  push:
    tags:
      - 'v*'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2

      - name: Login to Docker Hub
        uses: docker/login-action@v2
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}

      - name: Build and push API
        uses: docker/build-push-action@v4
        with:
          context: .
          file: ./docker/Dockerfile.api
          push: true
          tags: ${{ secrets.DOCKERHUB_USERNAME }}/mail2calendar-api:${{ github.ref_name }}
          cache-from: type=registry,ref=${{ secrets.DOCKERHUB_USERNAME }}/mail2calendar-api:buildcache
          cache-to: type=registry,ref=${{ secrets.DOCKERHUB_USERNAME }}/mail2calendar-api:buildcache,mode=max

      - name: Build and push NER service  
        uses: docker/build-push-action@v4
        with:
          context: ./ner-service
          file: ./docker/Dockerfile.ner
          push: true
          tags: ${{ secrets.DOCKERHUB_USERNAME }}/mail2calendar-ner:${{ github.ref_name }}
          cache-from: type=registry,ref=${{ secrets.DOCKERHUB_USERNAME }}/mail2calendar-ner:buildcache
          cache-to: type=registry,ref=${{ secrets.DOCKERHUB_USERNAME }}/mail2calendar-ner:buildcache,mode=max