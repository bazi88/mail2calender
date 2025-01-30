# System Architecture

## Project Structure

### Core Components
1. `/cmd`
   - Main application entry points
   - `/app` - Main service executable
   - `/migrate` - Database migration tool
   - `/seed` - Database seeder
   - `/route` - HTTP route definitions

2. `/internal`
   - `/delivery` - HTTP/gRPC handlers
   - `/domain` - Business logic and interfaces
   - `/infrastructure` - External service implementations
   - `/middleware` - HTTP middleware
   - `/server` - Server setup and configuration
   - `/cache` - Caching implementations
   - `/config` - Internal configuration
   - `/utility` - Shared utilities

3. `/config`
   - Configuration management
   - Environment variable handling
   - Service configuration (API, cache, database, etc.)

4. `/database`
   - Migration scripts
   - Database seeding
   - Schema definitions

## Service Architecture

### Mail Processing Flow
1. Email Reception
   - SMTP server integration
   - Email parsing and validation
   - Content extraction

2. NER Processing
   - gRPC communication with NER service
   - Entity extraction and classification
   - Event detail parsing

3. Calendar Integration
   - OAuth2 authentication
   - Calendar API interaction
   - Event creation and management

## Database Schema

### Core Tables
1. Users
   ```sql
   CREATE TABLE users (
     id UUID PRIMARY KEY,
     email VARCHAR NOT NULL UNIQUE,
     created_at TIMESTAMP,
     updated_at TIMESTAMP
   )
   ```

2. Sessions
   ```sql
   CREATE TABLE sessions (
     id UUID PRIMARY KEY,
     user_id UUID REFERENCES users(id),
     expires_at TIMESTAMP,
     created_at TIMESTAMP
   )
   ```

3. Email Configurations
   ```sql
   CREATE TABLE email_configs (
     id UUID PRIMARY KEY,
     user_id UUID REFERENCES users(id),
     provider VARCHAR,
     credentials JSONB,
     created_at TIMESTAMP
   )
   ```

4. Calendar Configurations
   ```sql
   CREATE TABLE calendar_configs (
     id UUID PRIMARY KEY,
     user_id UUID REFERENCES users(id),
     provider VARCHAR,
     credentials JSONB,
     created_at TIMESTAMP
   )
   ```

## External Services Integration

### NER Service
- Python-based NLP processing
- gRPC API for entity extraction
- Containerized deployment
- Model versioning and updates

### Email Services
- SMTP server configuration
- Email parsing and processing
- Attachment handling
- Error recovery

### Calendar Services
- OAuth2 authentication flow
- API rate limiting
- Error handling
- Conflict resolution

### Monitoring and Telemetry
- OpenTelemetry integration
- Prometheus metrics
- Grafana dashboards
- Loki log aggregation

## Security Architecture
1. Authentication
   - JWT-based API authentication
   - OAuth2 for external services
   - Session management

2. Authorization
   - Role-based access control
   - Resource ownership validation
   - API endpoint protection

3. Data Protection
   - Encryption at rest
   - Secure credential storage
   - Personal data handling (GDPR)