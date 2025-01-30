#!/bin/bash

# Màu sắc cho output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# URL cơ sở
BASE_URL="http://localhost:8080"

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
test_api "/api/v1/ner/extract" "POST" '{"text":"Google is a technology company headquartered in Mountain View, California. Microsoft is located in Redmond, Washington. Amazon has its headquarters in Seattle."}' 200

# Test 5: Rate limit test
echo "Test 5: Rate limit test (gửi nhiều request liên tiếp)"
for i in {1..5}; do
    test_api "/api/v1/ner/extract" "POST" '{"text":"Quick test for rate limit"}' 200
    sleep 1
done 