# System Architecture

## Cấu Trúc Dự Án
```
/cmd
  /api         # HTTP/gRPC API server
  /worker      # Background job worker
  /migration   # Database migration tool

/internal
  /domain      # Business logic & interfaces
  /repository  # Data access layer
  /service     # Application services
  /transport   # HTTP/gRPC handlers

/config        # Configuration management
  /dev
  /prod
  /test

/database      # Database migrations & seeds
  /migrations
  /seeds

/ent          # Entity framework
  /schema     # Database schema
  /generate   # Generated code

/ner-service   # Named Entity Recognition
  /model      # ML models
  /api        # REST API

/test         # Testing utilities
  /mock      # Mock implementations
  /fixtures  # Test data
```

## Kiến Trúc Dịch Vụ
1. Core Services
   - API Server (Go)
   - Background Worker (Go) 
   - NER Service (Python)

2. Infrastructure
   - PostgreSQL (Primary DB)
   - Redis (Cache & Queue)
   - Nginx (Reverse Proxy)
   - OpenTelemetry (Monitoring)

## Schema Cơ Sở Dữ Liệu
1. Users & Authentication
   - users
   - sessions
   - api_keys

2. Email Processing
   - email_accounts
   - email_messages
   - email_attachments

3. Calendar Integration  
   - calendar_accounts
   - events
   - participants

4. System Management
   - jobs
   - audit_logs
   - settings

## Tích Hợp Dịch Vụ
1. Authentication
   - JWT based auth
   - OAuth2 providers
   - API key management

2. External APIs
   - Email providers
   - Calendar services
   - NER endpoints

3. Background Processing
   - Job queues
   - Scheduled tasks
   - Error handling

4. Monitoring
   - Metrics collection
   - Distributed tracing
   - Error reporting 