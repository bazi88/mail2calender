# Technology Stack

## Go Environment
- Go version: 1.21+
- Go modules for dependency management
- Air for hot reload development
- Task runner for development workflows

## Core Dependencies

### Backend Framework
- net/http for HTTP server
- gorilla/mux for routing
- ent for ORM and database operations
- golang-migrate for database migrations

### Authentication & Security
- golang-jwt/jwt for JWT handling
- oauth2 for OAuth2 integration
- bcrypt for password hashing
- secure sessions management

### Infrastructure
- Docker for containerization
- Docker Compose for local development
- Kubernetes (optional) for production deployment
- PostgreSQL for primary database
- Redis for caching and session storage

### Monitoring & Observability
- OpenTelemetry for distributed tracing
- Prometheus for metrics collection
- Grafana for visualization
- Loki for log aggregation

### NER Service Stack
- Python 3.9+
- spaCy for NLP processing
- gRPC for service communication
- Protocol Buffers for data serialization

## Development Environment Setup

### Prerequisites
1. Go 1.21 or higher
2. Docker and Docker Compose
3. Task runner
4. Python 3.9+ (for NER service)
5. PostgreSQL client
6. Redis tools

### Installation Steps
```bash
# Install Go
wget https://golang.org/dl/go1.21.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Install Task
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b ~/.local/bin

# Install Docker
sudo apt-get update
sudo apt-get install docker-ce docker-ce-cli containerd.io

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.x.x/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Setup Python environment (for NER service)
python3 -m venv venv
source venv/bin/activate
pip install -r ner-service/requirements.txt
```

### Development Tools
1. VSCode Extensions
   - Go
   - Docker
   - Python
   - Protocol Buffers
   - YAML
   - SQL Tools

2. Database Tools
   - pgAdmin or DBeaver
   - Redis Commander

### Configuration Files
1. .env
   ```
   # Server
   PORT=8080
   ENV=development
   
   # Database
   DB_HOST=localhost
   DB_PORT=5432
   DB_NAME=mail2calendar
   DB_USER=postgres
   DB_PASSWORD=postgres
   
   # Redis
   REDIS_HOST=localhost
   REDIS_PORT=6379
   
   # NER Service
   NER_SERVICE_HOST=localhost
   NER_SERVICE_PORT=50051
   ```

2. docker-compose.yml
   - PostgreSQL service
   - Redis service
   - NER service
   - Monitoring stack

### Development Workflow
1. Start infrastructure:
   ```bash
   task infra:up
   ```

2. Run migrations:
   ```bash
   task db:migrate
   ```

3. Start development server:
   ```bash
   task dev
   ```

4. Run tests:
   ```bash
   task test
   ```

## Third-party Services

### Email Providers
- SMTP servers
- Email parsing libraries
- Attachment handling

### Calendar Services
- Google Calendar API
- Microsoft Graph API
- CalDAV providers

### Authentication Providers
- OAuth2 providers
- JWT authentication
- Session management

## Infrastructure Components

### Database
- PostgreSQL for persistent storage
- Redis for caching and sessions
- Database migrations
- Data backup and recovery

### Monitoring
- Metrics collection
- Log aggregation
- Distributed tracing
- Alert management

### Security
- TLS/SSL certificates
- API authentication
- Rate limiting
- Data encryption