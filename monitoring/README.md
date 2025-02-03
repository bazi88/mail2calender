# Monitoring & Observability

## Kiến Trúc Monitoring

```
                   ┌─────────────┐
                   │  Grafana    │
                   │ Dashboards  │
                   └──────┬──────┘
                          │
          ┌───────────────┴───────────────┐
          │                               │
   ┌──────┴──────┐               ┌───────┴──────┐
   │ Prometheus  │               │    Jaeger    │
   │  Metrics    │               │   Tracing    │
   └──────┬──────┘               └──────┬───────┘
          │                             │
   ┌──────┴──────┐               ┌──────┴───────┐
   │ OpenTelemetry│               │OpenTelemetry │
   │  Collector   │◄─────────────►│   Agent     │
   └──────┬──────┘               └──────┬───────┘
          │                             │
    ┌─────┴─────┐                ┌─────┴─────┐
    │   API     │                │    NER    │
    │ Service   │                │  Service  │
    └───────────┘                └───────────┘
```

## Metrics Collection

### System Metrics
- CPU Usage
- Memory Usage
- Disk I/O
- Network Traffic

### Application Metrics
1. API Service
   - Request Rate
   - Response Time
   - Error Rate
   - Active Users

2. NER Service
   - Processing Time
   - Model Accuracy
   - Queue Length
   - Memory Usage

3. Worker
   - Job Processing Rate
   - Queue Length
   - Processing Time
   - Error Rate

## Alerting Rules

### Critical Alerts
- Service Down
- High Error Rate (>5%)
- High Latency (>2s)
- Memory Usage (>90%)

### Warning Alerts
- Increased Error Rate (>1%)
- Increased Latency (>1s)
- Queue Buildup
- Disk Space (>80%)

## Dashboards

### Overview
- System Health
- Service Status
- Error Rates
- Performance Metrics

### Service Specific
1. API Dashboard
   - Endpoint Performance
   - Authentication Stats
   - Rate Limiting
   - Error Distribution

2. NER Dashboard
   - Model Performance
   - Processing Queue
   - Resource Usage
   - Accuracy Metrics

3. Worker Dashboard
   - Job Statistics
   - Queue Metrics
   - Processing Time
   - Error Analysis

## Log Management

### Log Levels
- ERROR: Lỗi nghiêm trọng
- WARN: Cảnh báo
- INFO: Thông tin quan trọng
- DEBUG: Chi tiết debug

### Log Aggregation
- Elasticsearch
- Kibana Visualization
- Log Retention
- Search & Analysis 