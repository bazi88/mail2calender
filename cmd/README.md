# /cmd

Thư mục này chứa các entry point chính của ứng dụng.

## Cấu Trúc

```
/api
  main.go          # HTTP/gRPC API server
  config.go        # Server configuration
  routes.go        # API routes

/worker  
  main.go          # Background job worker
  processor.go     # Job processing logic
  scheduler.go     # Task scheduling

/migration
  main.go          # Database migration tool
  versions.go      # Migration versions
```

## Các Thành Phần

### API Server
- HTTP và gRPC endpoints
- Authentication & authorization
- Request validation
- Response handling

### Background Worker
- Email processing
- Calendar sync
- Scheduled tasks
- Error recovery

### Migration Tool
- Database schema updates
- Data migration
- Rollback handling

## Sử Dụng

```bash
# Start API server
go run cmd/api/main.go

# Run worker
go run cmd/worker/main.go

# Execute migrations
go run cmd/migration/main.go up
``` 