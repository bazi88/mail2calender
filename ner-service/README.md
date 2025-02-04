# Enhanced NER Service with Time Processing

This service provides Named Entity Recognition capabilities with special handling for time expressions.

## Features

- Base NER capabilities (Person, Organization, Location)
- Enhanced time entity extraction
- Time normalization to ISO format
- Support for Vietnamese time expressions
- Batch processing support
- gRPC API with health checks
- Prometheus metrics

## Core Production Features

### 1. Rate Limiting

```python
# Implementation via Redis
RATE_LIMIT_KEY = "ner_service:{ip}"
RATE_LIMIT = 100  # requests per minute
RATE_WINDOW = 60  # seconds

# Usage in gRPC interceptor
async def rate_limit_interceptor(ip):
    current = await redis.incr(RATE_LIMIT_KEY.format(ip=ip))
    if current > RATE_LIMIT:
        raise grpc.StatusCode.RESOURCE_EXHAUSTED
```

### 2. Redis Caching

```python
# Cache configuration
CACHE_TTL = 3600  # 1 hour
CACHE_KEY = "ner_result:{text_hash}"

# Example caching strategy
async def get_cached_result(text):
    text_hash = hashlib.md5(text.encode()).hexdigest()
    cache_key = CACHE_KEY.format(text_hash=text_hash)
    return await redis.get(cache_key)
```

### 3. Recurring Events Processing

```python
# Event pattern recognition
recurring_patterns = {
    "daily": r"mỗi ngày|hàng ngày",
    "weekly": r"mỗi tuần|hàng tuần",
    "monthly": r"mỗi tháng|hàng tháng",
    "yearly": r"mỗi năm|hàng năm"
}

# RRULE generation
from datetime import datetime
from dateutil.rrule import DAILY, WEEKLY, MONTHLY, YEARLY

def generate_rrule(pattern, start_date):
    freq_mapping = {
        "daily": DAILY,
        "weekly": WEEKLY,
        "monthly": MONTHLY,
        "yearly": YEARLY
    }
    return f"FREQ={freq_mapping[pattern]};DTSTART={start_date}"
```

## Production Setup

### Rate Limiting Configuration

1. Configure Redis connection:

```yaml
# docker-compose.yml
redis:
  image: redis:7-alpine
  command: redis-server --maxmemory 100mb --maxmemory-policy allkeys-lru
  environment:
    - REDIS_PASSWORD=${REDIS_PASSWORD}
```

2. Set rate limits:

```env
RATE_LIMIT_PER_MINUTE=100
RATE_LIMIT_BURST=150
```

### Caching Strategy

1. Cache levels:

   - L1: In-memory LRU cache (fast, limited size)
   - L2: Redis cache (larger, distributed)

2. Cache invalidation:
   - Time-based (TTL)
   - Version-based (model updates)

```python
# Cache configuration example
cache_config = {
    "l1_size": 1000,        # Number of items in memory
    "l1_ttl": 300,          # 5 minutes
    "l2_ttl": 3600,         # 1 hour
    "version_key": "v1.0"   # Cache version
}
```

### Recurring Events

1. Pattern recognition:

   - Recurring keywords
   - Frequency patterns
   - Time ranges

2. RRULE generation:
   - ISO 8601 recurring rules
   - Calendar integration support

```python
# Example recurring event
{
    "text": "Họp hàng tuần vào thứ 2 lúc 9h",
    "pattern": "weekly",
    "rrule": "FREQ=WEEKLY;BYDAY=MO;DTSTART=20250203T090000Z",
    "confidence": 0.95
}
```

## Performance Optimizations

### 1. Rate Limiting

- Token bucket algorithm
- IP-based and API key-based limits
- Distributed rate limiting via Redis

### 2. Caching

- Two-level caching strategy
- Cache warming for common requests
- Intelligent cache invalidation

### 3. Recurring Events

- Pattern matching optimization
- Efficient RRULE generation
- Calendar sync optimization

## Monitoring & Alerts

1. Rate Limiting Metrics:

   - Request rate per client
   - Rejection rate
   - Burst patterns

2. Cache Performance:

   - Hit/miss ratios
   - Cache size
   - Eviction rates

3. Event Processing:
   - Pattern recognition accuracy
   - Processing time
   - Error rates

## Future Improvements

1. **Rate Limiting**:

   - Dynamic rate adjustment
   - Client-specific limits
   - Rate limit prediction

2. **Caching**:

   - Predictive caching
   - Cache prewarming
   - Partial result caching

3. **Recurring Events**:
   - Complex pattern recognition
   - Multi-language support
   - Calendar optimization

These improvements make the NER service production-ready with proper rate limiting, efficient caching, and comprehensive recurring event support.
