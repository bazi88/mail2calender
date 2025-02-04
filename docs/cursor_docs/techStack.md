# Technology Stack

## Backend (Go)
- Go version: 1.21
- Dependencies:
  ```
  - gin-gonic/gin: Web framework
  - ent: ORM & database toolkit
  - grpc-go: gRPC framework
  - zap: Logging
  - viper: Configuration
  - cobra: CLI framework
  ```

## NER Service (Python)
- Python version: 3.11
- Dependencies:
  ```
  - spacy: NLP framework
  - fastapi: API framework
  - pytorch: ML framework
  - transformers: Hugging Face models
  ```

## Infrastructure
1. Databases
   - PostgreSQL 15
   - Redis 7.0

2. Message Queue
   - Redis Streams
   - Bull Queue

3. Monitoring
   - Prometheus
   - Grafana
   - Jaeger

4. Deployment
   - Docker
   - Docker Compose
   - Nginx

## Development Tools
1. Code Quality
   - golangci-lint
   - pre-commit hooks
   - gofmt

2. Testing
   - go test
   - pytest
   - k6 (load testing)

3. Documentation
   - Swagger/OpenAPI
   - godoc
   - sphinx

## Third Party Services
1. Email Providers
   - Gmail API
   - Microsoft Graph
   - SMTP/IMAP

2. Calendar Services
   - Google Calendar
   - Microsoft Calendar
   - CalDAV

3. Authentication
   - JWT
   - OAuth2
   - API Keys

## Development Environment
1. Required Tools
   - Go 1.21+
   - Python 3.11+
   - Docker & Docker Compose
   - Make

2. Environment Variables
   - Development: .env.dev
   - Testing: .env.test
   - Production: .env.prod

3. Local Setup
   ```bash
   make setup      # Install dependencies
   make dev        # Start development
   make test       # Run tests
   make build      # Build containers
   ``` 